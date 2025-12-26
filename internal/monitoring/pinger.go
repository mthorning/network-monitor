package monitoring

import (
	"log/slog"
	"net"
	"network_monitor/internal/config"
	"os"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/tatsushid/go-fastping"
)

type TimeoutHandler interface {
	OnTimeout()
}

type Tracker struct {
	mu      sync.Mutex
	tracked map[string]bool
}

func newTracker() Tracker {
	return Tracker{
		tracked: make(map[string]bool),
	}
}

type Pinger struct {
	p                *fastping.Pinger
	th               TimeoutHandler
	replied          Tracker
	previousTimeouts []string
}

func NewPinger(opts *config.Opts, metrics *config.Metrics) *Pinger {
	p := fastping.NewPinger()
	p.MaxRTT = time.Duration(opts.PingInterval) * time.Second

	pinger := Pinger{
		p:                p,
		replied:          newTracker(),
		previousTimeouts: make([]string, 0),
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
		pinger.replied.mu.Lock()
		defer pinger.replied.mu.Unlock()

		// can receive 0s durations after a timeout, we should ignore them
		if rtt > 0 {
			slog.Debug("Response received", "ip", ip, "duration", rtt)
			metrics.DurationHist.WithLabelValues(ip.String()).Observe(rtt.Seconds())
			pinger.replied.tracked[ip.String()] = true
		}
	}

	pinger.p.OnIdle = func() {
		metrics.TotalPingsCounter.Inc()
		pinger.replied.mu.Lock()
		defer pinger.replied.mu.Unlock()

		timeouts := make([]string, 0)
		for ip, replied := range pinger.replied.tracked {
			if !replied {
				timeouts = append(timeouts, ip)

				metrics.TotalTimoutCounter.WithLabelValues(ip).Inc()
				// Record a metric value with 1ms above the ping interval so that we don't skew the hist
				metrics.DurationHist.WithLabelValues(ip).Observe(float64(opts.PingInterval) + 0.001)

				if pinger.th != nil {
					pinger.th.OnTimeout()
				}
			}

			// reset for next round
			pinger.replied.tracked[ip] = false
		}

		sort.Slice(timeouts, func(i, j int) bool {
			return timeouts[i] > timeouts[j]
		})

		if !slices.Equal(pinger.previousTimeouts, timeouts) {
			slog.Info("Pings timed out", "ips", timeouts)
		}
		if len(pinger.previousTimeouts) > 0 && len(timeouts) == 0 {
			slog.Info("No more timeouts")
		}

		pinger.previousTimeouts = timeouts
	}
}

func (pinger *Pinger) Run() {
	pinger.p.RunLoop()
}
