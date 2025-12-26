package network

import (
	"log/slog"
	"net"
	"network_monitor/internal/config"
	"time"

	"github.com/tatsushid/go-fastping"
)

type FastPing struct {
	p             *fastping.Pinger
	OnResponse    func(ip *net.IPAddr, rtt time.Duration)
	OnIntervalEnd func()
}

func NewFastPing(opts config.Opts) *FastPing {
	p := fastping.NewPinger()
	p.MaxRTT = time.Duration(opts.PingInterval) * time.Second

	fp := FastPing{
		p:             p,
		OnResponse:    nil,
		OnIntervalEnd: nil,
	}

	fp.p.OnRecv = func(ip *net.IPAddr, rtt time.Duration) {
		if fp.OnResponse != nil {
			fp.OnResponse(ip, rtt)
		}
	}

	fp.p.OnIdle = func() {
		if fp.OnIntervalEnd != nil {
			fp.OnIntervalEnd()
		}
	}

	return &fp
}

func (f *FastPing) AddIPAddr(ip *net.IPAddr) {
	f.p.AddIPAddr(ip)
}

func (f *FastPing) SetOnResponse(cb func(ip *net.IPAddr, rtt time.Duration)) {
	f.OnResponse = cb
}

func (f *FastPing) SetOnIntervalEnd(cb func()) {
	f.OnIntervalEnd = cb
}

func (f *FastPing) Run() {
	f.p.RunLoop()

	select {
	case <-f.p.Done():
		if err := f.p.Err(); err != nil {
			slog.Debug("Error from fastping", "error", err)
		}
	}
}
