package utils

import (
	"log/slog"
	"net"
	"network-monitor/internal/config"
	"os"
	"sync"
	"time"

	"github.com/tatsushid/go-fastping"
)

func addIps(ips []string, p *fastping.Pinger) {
	for _, ip := range ips {
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)
		if err != nil {
			slog.Error("Error resolving IP", "error", err.Error())
			os.Exit(1)
		}
		p.AddIPAddr(ra)
	}
}

type ReplyTracker struct {
	mu      sync.Mutex
	replied map[string]bool
}

func PingAndReport(ips []string, opts *config.Opts, metrics *config.Metrics) {
	p := fastping.NewPinger()
	p.MaxRTT = time.Duration(opts.PingInterval) * time.Second

	addIps(ips, p)

	replyTracker := &ReplyTracker{
		replied: make(map[string]bool),
	}

	p.OnRecv = func(ip *net.IPAddr, rtt time.Duration) {
		replyTracker.mu.Lock()
		defer replyTracker.mu.Unlock()

		// can receive 0s durations after a timeout, we should ignore them
		if rtt > 0 {
			slog.Debug("Response received", "ip", ip, "duration", rtt)
			metrics.DurationHist.WithLabelValues(ip.String()).Observe(rtt.Seconds())
			replyTracker.replied[ip.String()] = true
		}
	}

	p.OnIdle = func() {
		metrics.TotalPingsCounter.Inc()
		replyTracker.mu.Lock()
		defer replyTracker.mu.Unlock()

		for ip, replied := range replyTracker.replied {
			if !replied {
				slog.Info("Ping timed out", "ip", ip)

				metrics.TotalTimoutCounter.WithLabelValues(ip).Inc()
				// Record a metric value with 1ms above the ping interval so that we don't skew the hist
				metrics.DurationHist.WithLabelValues(ip).Observe(float64(opts.PingInterval) + 0.001)
			}

			// reset for next round
			replyTracker.replied[ip] = false
		}
	}

	p.RunLoop()
}
