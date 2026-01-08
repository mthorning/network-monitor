package monitoring

import (
	"log/slog"
	"net"
	"network_monitor/internal/config"
	"network_monitor/internal/network"
	"network_monitor/internal/utils"
	"os"
	"strings"
)

type Manager struct {
	pingLoop       *network.PingLoop
	opts           config.Opts
	timeoutTracker *timeoutTracker
	traceTracker   *utils.Tracker[[]network.Hop]
	traceCountdown int
}

func NewManager(opts config.Opts, metrics *config.Metrics) (*Manager, error) {
	pl, err := network.NewPingLoop(opts.PingInterval)
	if err != nil {
		return nil, err
	}
	m := Manager{
		pingLoop:       pl,
		opts:           opts,
		timeoutTracker: newTimeoutTracker(opts.PingIps),
		traceTracker:   utils.NewTracker[[]network.Hop](),
		traceCountdown: opts.TraceFrequency,
	}

	m.addIps(opts.PingIps)
	m.configure(metrics)

	return &m, nil
}

func (m *Manager) Run() {
	m.pingLoop.Run()
}

func (m *Manager) addIps(ips []string) {
	for _, ip := range ips {
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)
		if err != nil {
			slog.Error("Error resolving IP", "error", err.Error())
			os.Exit(1)
		}
		m.pingLoop.AddIpAddr(ra)
	}
}

func (m *Manager) configure(metrics *config.Metrics) {
	m.pingLoop.OnIntervalStart = func() {
		m.traceCountdown -= 1
		slog.Debug("Trace countdown", "num", m.traceCountdown)

		for _, ip := range m.opts.PingIps {
			metrics.TotalPingsCounter.WithLabelValues(ip).Inc()

			if m.traceCountdown == 0 {
				if hops, ok := m.runTrace(ip); ok {
					slog.Debug("Trace run", "ip", ip, "hops", hops)
					m.traceTracker.Set(ip, hops)
				}
			}
		}

		if m.traceCountdown <= 0 {
			m.traceCountdown = m.opts.TraceFrequency
		}
	}

	m.pingLoop.OnResponse = func(res *network.PingLoopResponse) {
		// can receive 0s durations after a timeout, we should ignore them
		if res.Duration > 0 {
			metrics.DurationHist.WithLabelValues(res.Peer.String()).Observe(res.Duration.Seconds())
			m.timeoutTracker.replyReceived(res.Peer.String())
		}
	}

	m.pingLoop.OnIntervalEnd = func(ospid, seq int) {
		timeouts := m.timeoutTracker.getTimeouts()
		slog.Debug("Interval ended", "timeouts", timeouts)

		for _, t := range timeouts {
			metrics.TotalTimoutCounter.WithLabelValues(t.ip).Inc()

			if t.count >= m.opts.TraceTimeoutThreshold {
				if hops, ok := m.runTrace(t.ip); ok {
					slog.Warn("Ping threshold crossed", "ip", t.ip, "good", m.traceTracker.Get(t.ip), "bad", hops, "ospid", ospid, "seq", seq)
					m.timeoutTracker.resetCount(t.ip)
				}
			}
		}
	}
}

func (m *Manager) runTrace(ip string) ([]network.Hop, bool) {
	if strings.HasPrefix(ip, "192.168") {
		return nil, false
	}
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		slog.Error("Error resolving IP for traceroute", "error", err.Error(), "ip", ip)
		return nil, false
	}

	hops, err := network.Traceroute(ra)
	if err != nil {
		slog.Error("Error from traceroute", "error", err.Error(), "ip", ip)
		return nil, false
	}

	return hops, true
}
