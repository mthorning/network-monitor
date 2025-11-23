package utils

import (
	"log/slog"
	"net"
	"network-monitor/internal/config"
	"os"
	"time"

	"github.com/tatsushid/go-fastping"
)

func newPinger(ip string, metrics *config.Metrics, logger *slog.Logger) *fastping.Pinger {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		logger.Error("Error resolving IP", "error", err.Error())
		os.Exit(1)
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		logger.Debug("Response received", "duration", rtt)
		metrics.DurationHist.WithLabelValues(ip).Observe(rtt.Seconds())
	}
	p.OnIdle = func() {}
	return p
}

func PingAndReport(ip string, opts *config.Opts, metrics *config.Metrics) {
	logger := slog.Default().With("ip", ip)
	p := newPinger(ip, metrics, logger)

	for {
		err := p.Run()
		if err != nil {
			logger.Error("Error running ping", "error", err.Error())
		}
		time.Sleep(time.Duration(opts.PingInterval) * time.Second)
	}
}
