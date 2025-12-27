package network

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"network_monitor/internal/utils"
	"slices"
	"time"

	"golang.org/x/net/icmp"
)

type PingLoopResponse struct {
	Body     *icmp.Echo
	Peer     net.Addr
	Duration time.Duration
}

type PingLoop struct {
	interval      time.Duration
	pingIps       []*net.IPAddr
	OnResponse    func(*PingLoopResponse)
	OnIntervalEnd func()
	OnError       func(err error)
	resChan       chan PingLoopResponse
	errChan       chan error
	icmpPing      *ICMPPing
}

func NewPingLoop(interval uint) (*PingLoop, error) {
	icmpPing, err := NewICMPPing()
	if err != nil {
		return nil, err
	}

	p := PingLoop{
		icmpPing: icmpPing,
		interval: time.Duration(interval) * time.Second,
		pingIps:  make([]*net.IPAddr, 0),
		resChan:  make(chan PingLoopResponse),
		errChan:  make(chan error),
	}

	return &p, nil
}

func (p *PingLoop) AddIpAddr(ip *net.IPAddr) {
	if !slices.Contains(p.pingIps, ip) {
		p.pingIps = append(p.pingIps, ip)
	}
}

func (p *PingLoop) Run() error {
	if len(p.pingIps) == 0 {
		return errors.New("At least one IP to ping is required")
	}
	if p.OnResponse == nil {
		return errors.New("OnResponse not set")
	}
	if p.OnIntervalEnd == nil {
		return errors.New("OnIntervalEnd not set")
	}
	if p.OnError == nil {
		return errors.New("OnError not set")
	}

	go p.listenToResChan()
	go p.listenToErrChan()
	go p.startLoop()

	return nil
}

func (p *PingLoop) startLoop() {
	seq := 0
	for {
		for _, ip := range p.pingIps {
			seq += 1
			opts := ICMPPingOpts{
				IP:                   ip,
				Seq:                  seq,
				ReadDeadlineDuration: p.interval,
			}
			go p.makePing(opts)
		}

		time.Sleep(p.interval)
		p.OnIntervalEnd()
	}
}

func (p *PingLoop) listenToResChan() {
	for res := range p.resChan {
		p.OnResponse(&res)
	}
}

func (p *PingLoop) listenToErrChan() {
	for err := range p.errChan {
		p.OnError(err)
	}
}

func (p *PingLoop) makePing(opts ICMPPingOpts) {
	slog.Debug("Pinging", "ip", opts.IP, "seq", opts.Seq)
	res, err := p.icmpPing.Ping(opts)
	if err != nil {
		p.errChan <- err
		return
	}

	duration, err := getDuration(res.Body)
	if err != nil {
		p.errChan <- err
		return
	}

	response := PingLoopResponse{
		Body:     res.Body,
		Peer:     res.Peer,
		Duration: duration,
	}

	p.resChan <- response
}

func getDuration(body *icmp.Echo) (time.Duration, error) {
	if len(body.Data) < 8 {
		return 0, errors.New(fmt.Sprintf("Echo reply data too short, length: %d", len(body.Data)))
	}

	start := utils.BinaryToTime(body.Data[:8])
	now := time.Now()
	duration := now.Sub(start)
	return duration, nil
}
