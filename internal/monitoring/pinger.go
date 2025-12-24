package monitoring

import (
	"log/slog"
	"net"
	"network-monitor/internal/config"
	"os"
	"sync"
	"time"

	"github.com/tatsushid/go-fastping"
)

type TimeoutHandler interface {
	OnTimeout()
}

type Pinger struct {
	p       *fastping.Pinger
	mu      sync.Mutex
	replied map[string]bool
	th      TimeoutHandler
}

func NewPinger(opts *config.Opts, metrics *config.Metrics) *Pinger {
	p := fastping.NewPinger()
	p.MaxRTT = time.Duration(opts.PingInterval) * time.Second

	pinger := Pinger{
		p:       p,
		replied: make(map[string]bool),
	}

	pinger.addIps(opts.PingIps)
	pinger.configure(opts, metrics)

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
	}
}

func (pinger *Pinger) configure(opts *config.Opts, metrics *config.Metrics) {
	pinger.p.OnRecv = func(ip *net.IPAddr, rtt time.Duration) {
		pinger.mu.Lock()
		defer pinger.mu.Unlock()

		// can receive 0s durations after a timeout, we should ignore them
		if rtt > 0 {
			slog.Debug("Response received", "ip", ip, "duration", rtt)
			metrics.DurationHist.WithLabelValues(ip.String()).Observe(rtt.Seconds())
			pinger.replied[ip.String()] = true
		}
	}

	pinger.p.OnIdle = func() {
		metrics.TotalPingsCounter.Inc()
		pinger.mu.Lock()
		defer pinger.mu.Unlock()

		for ip, replied := range pinger.replied {
			if !replied {
				slog.Info("Ping timed out", "ip", ip)

				metrics.TotalTimoutCounter.WithLabelValues(ip).Inc()
				// Record a metric value with 1ms above the ping interval so that we don't skew the hist
				metrics.DurationHist.WithLabelValues(ip).Observe(float64(opts.PingInterval) + 0.001)

				if pinger.th != nil {
					pinger.th.OnTimeout()
				}
			}

			// reset for next round
			pinger.replied[ip] = false
		}
	}
}

func (pinger *Pinger) Run() {
	pinger.p.RunLoop()
}
