package monitoring

import (
	"log/slog"
	"net"
	"network_monitor/internal/config"
	"os"
	"time"
)

type PingLooper interface {
	AddIPAddr(ip *net.IPAddr)
	SetOnResponse(cb func(ip *net.IPAddr, rtt time.Duration))
	SetOnIntervalEnd(cb func())
	Run()
}

type Pinger struct {
	p       PingLooper
	opts    config.Opts
	tracker PingTracker
}

func NewPinger(p PingLooper, opts config.Opts, metrics *config.Metrics) *Pinger {
	pinger := Pinger{
		p:       p,
		opts:    opts,
		tracker: newPingTracker(),
	}

	pinger.addIps(opts.PingIps)
	pinger.configure(metrics)

	return &pinger
}

func (pinger *Pinger) addIps(ips []string) {
	for _, ip := range ips {
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)
		if err != nil {
			slog.Error("Error resolving IP", "error", err.Error())
			os.Exit(1)
		}
		pinger.p.AddIPAddr(ra)
		pinger.tracker.replies[ip] = false
	}
}

func (pinger *Pinger) configure(metrics *config.Metrics) {
	pinger.p.SetOnResponse(func(ip *net.IPAddr, rtt time.Duration) {
		// can receive 0s durations after a timeout, we should ignore them
		if rtt > 0 {
			slog.Debug("Response received", "ip", ip, "duration", rtt)
			metrics.DurationHist.WithLabelValues(ip.String()).Observe(rtt.Seconds())
			pinger.tracker.replyReceived(ip.String())
		}
	})

	pinger.p.SetOnIntervalEnd(func() {
		metrics.TotalPingsCounter.Inc()
		timeouts := pinger.tracker.getTimeouts()

		for _, ip := range timeouts {
			metrics.TotalTimoutCounter.WithLabelValues(ip).Inc()
			// Record a metric value with 1ms above the ping interval so that we don't skew the hist
			metrics.DurationHist.WithLabelValues(ip).Observe(float64(pinger.opts.PingInterval) + 0.001)
		}
	})
}

func (pinger *Pinger) Run() {
	pinger.p.Run()
}
