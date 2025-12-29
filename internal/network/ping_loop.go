package network

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"network_monitor/internal/utils"
	"slices"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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
	resChan       chan PingLoopResponse
	icmpPing      *ICMPPing
	ospid         int
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
		ospid:    rand.Intn(0xffff),
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

	go p.listenToResChan()
	go p.startLoop()

	return nil
}

func (p *PingLoop) startLoop() {
	seq := 0
	for {
		var wg sync.WaitGroup
		wg.Add(1)
		rtn := make(chan ICMPPingResponse)

		// Read should time out and call Done, allowing the loop
		// to continue
		p.icmpPing.Read(rtn, &wg, p.interval)

		go p.listenForMessage(seq, seq+len(p.pingIps), rtn)

		for _, ip := range p.pingIps {
			seq += 1
			opts := ICMPPingOpts{
				ID:  p.ospid,
				IP:  ip,
				Seq: seq,
			}
			slog.Debug("Pinging", "ip", opts.IP, "seq", opts.Seq)

			err := p.icmpPing.Ping(opts)
			if err != nil {
				slog.Error("Failed to ping", "error", err, "ip", opts.IP)
				continue
			}
		}

		wg.Wait()
		close(rtn)
		p.OnIntervalEnd()
	}
}

func (p *PingLoop) listenToResChan() {
	for res := range p.resChan {
		p.OnResponse(&res)
	}
}

func (p *PingLoop) listenForMessage(min int, max int, rtn chan ICMPPingResponse) {
	for res := range rtn {
		if res.Message.Type == ipv4.ICMPTypeEchoReply {
			body := res.Message.Body.(*icmp.Echo)
			if body.ID == p.ospid && body.Seq > min && body.Seq <= max {
				duration, err := getDuration(body)
				if err != nil {
					slog.Warn("Unable to get duration", "error", err)
				}
				slog.Debug("Received response", "ip", res.Peer, "duration", duration, "seq", body.Seq)

				response := PingLoopResponse{
					Body:     body,
					Peer:     res.Peer,
					Duration: duration,
				}

				p.resChan <- response
			} else {
				slog.Debug("Received different ID/Seq ICMP message", "ospid", p.ospid, "id", body.ID, "seq", body.Seq)
			}
		} else {
			slog.Debug("Received different type ICMP message", "type", res.Message.Type)
		}
	}
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
