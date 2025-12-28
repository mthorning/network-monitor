package network

import (
	"errors"
	"fmt"
	"net"
	"network_monitor/internal/utils"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPPing struct {
	conn *icmp.PacketConn
}

type ICMPPingOpts struct {
	IP                   *net.IPAddr
	ReadDeadlineDuration time.Duration
	TTL                  int
	Seq                  int
}

type ICMPPingResponse struct {
	Body *icmp.Echo
	Peer net.Addr
}

var ErrTimeExceeded = errors.New("Time exceeded")

func NewICMPPing() (*ICMPPing, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	return &ICMPPing{
		conn: c,
	}, nil
}

func (p *ICMPPing) Ping(opts ICMPPingOpts) (*ICMPPingResponse, error) {
	if err := checkOpts(&opts); err != nil {
		return nil, err
	}

	pc := p.conn.IPv4PacketConn()
	if err := pc.SetTTL(opts.TTL); err != nil {
		return nil, err
	}

	now := time.Now()
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:   os.Getpid(),
			Seq:  opts.Seq,
			Data: utils.TimeToBinary(now),
		},
	}

	mb, err := m.Marshal(nil)
	if err != nil {
		return nil, err
	}

	if _, err := p.conn.WriteTo(mb, opts.IP); err != nil {
		return nil, err
	}

	if err := p.conn.SetReadDeadline(time.Now().Add(opts.ReadDeadlineDuration)); err != nil {
		return nil, err
	}

	buf := make([]byte, 1500)

	for {
		n, peer, err := p.conn.ReadFrom(buf)
		if err != nil {
			continue
		}

		msg, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			continue
		}

		switch msg.Type {
		case ipv4.ICMPTypeEchoReply:
			body := msg.Body.(*icmp.Echo)
			if body.ID == os.Getpid() && body.Seq == opts.Seq {
				r := ICMPPingResponse{
					Body: body,
					Peer: peer,
				}
				return &r, nil
			}
		case ipv4.ICMPTypeTimeExceeded:
			return nil, ErrTimeExceeded
		case ipv4.ICMPTypeDestinationUnreachable:
			return nil, errors.New("Destination unreachable")
		default:
			return nil, errors.New(fmt.Sprintf("Unhandled ICMP type: %s", msg.Type))
		}
	}
}

func (p *ICMPPing) Close() {
	defer p.conn.Close()
}

func checkOpts(opts *ICMPPingOpts) error {
	if opts.IP.String() == "" {
		return errors.New("opts.IP is required, no value set")
	}
	if opts.TTL == 0 {
		opts.TTL = 64
	}
	if opts.ReadDeadlineDuration == 0 {
		opts.ReadDeadlineDuration = 3 * time.Second
	}

	return nil
}
