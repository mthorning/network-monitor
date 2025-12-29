package monitoring

import (
	"log/slog"
	"net"
	"network_monitor/internal/config"
	"network_monitor/internal/network"
	"os"
)

type Pinger struct {
	pl      *network.PingLoop
	opts    config.Opts
	tracker *PingTracker
}

func NewPinger(opts config.Opts, metrics *config.Metrics) (*Pinger, error) {
	pl, err := network.NewPingLoop(opts.PingInterval)
	if err != nil {
		return nil, err
	}
	pinger := Pinger{
		pl:      pl,
		opts:    opts,
		tracker: newPingTracker(),
	}

	pinger.addIps(opts.PingIps)
	pinger.configure(metrics)

	return &pinger, nil
}

func (pinger *Pinger) Run() {
	pinger.pl.Run()
}

func (pinger *Pinger) addIps(ips []string) {
	for _, ip := range ips {
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)
		if err != nil {
			slog.Error("Error resolving IP", "error", err.Error())
			os.Exit(1)
		}
		pinger.pl.AddIpAddr(ra)
		pinger.tracker.replies[ip] = false
	}
}

func (pinger *Pinger) configure(metrics *config.Metrics) {
	pinger.pl.OnResponse = func(res *network.PingLoopResponse) {
		// can receive 0s durations after a timeout, we should ignore them
		if res.Duration > 0 {
			metrics.DurationHist.WithLabelValues(res.Peer.String()).Observe(res.Duration.Seconds())
			pinger.tracker.replyReceived(res.Peer.String())
		}
	}

	pinger.pl.OnIntervalEnd = func() {
		for _, ip := range pinger.opts.PingIps {
			metrics.TotalPingsCounter.WithLabelValues(ip).Inc()
		}

		timeouts := pinger.tracker.getTimeouts()
		pinger.tracker.reset()
		slog.Debug("Interval ended", "timeouts", timeouts)

		for _, ip := range timeouts {
			metrics.TotalTimoutCounter.WithLabelValues(ip).Inc()
			// Record a metric value with 1ms above the ping interval so that we don't skew the hist
			metrics.DurationHist.WithLabelValues(ip).Observe(float64(pinger.opts.PingInterval) + 0.001)
		}
	}
}
