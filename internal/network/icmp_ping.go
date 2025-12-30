package network

import (
	"errors"
	"log/slog"
	"net"
	"network_monitor/internal/utils"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPPing struct {
	conn *icmp.PacketConn
}

type ICMPPingOpts struct {
	ID  int
	IP  *net.IPAddr
	TTL int
	Seq int
}

type ICMPPingResponse struct {
	Message *icmp.Message
	Peer    net.Addr
}

func NewICMPPing() (*ICMPPing, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	return &ICMPPing{
		conn: c,
	}, nil
}

func (p *ICMPPing) Ping(opts ICMPPingOpts) error {
	if err := checkOpts(&opts); err != nil {
		return err
	}

	pc := p.conn.IPv4PacketConn()
	if err := pc.SetTTL(opts.TTL); err != nil {
		return err
	}

	now := time.Now()
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:   opts.ID,
			Seq:  opts.Seq,
			Data: utils.TimeToBinary(now),
		},
	}

	mb, err := m.Marshal(nil)
	if err != nil {
		return err
	}

	if _, err := p.conn.WriteTo(mb, opts.IP); err != nil {
		return err
	}
	return nil
}

func (p *ICMPPing) Read(readDeadlineDuration time.Duration) (chan ICMPPingResponse, error) {
	rtn := make(chan ICMPPingResponse)
	rd := readDeadlineDuration
	if rd == 0 {
		rd = 3 * time.Second
	}

	if err := p.conn.SetReadDeadline(time.Now().Add(readDeadlineDuration)); err != nil {
		return nil, err
	}

	buf := make([]byte, 1500)

	go func() {
		for {
			n, peer, err := p.conn.ReadFrom(buf)
			if err != nil {
				break
			}

			msg, err := icmp.ParseMessage(1, buf[:n])
			if err != nil {
				slog.Warn("Unable to parse icmp message")
				continue
			}

			res := ICMPPingResponse{
				Message: msg,
				Peer:    peer,
			}

			rtn <- res
		}
		close(rtn)
	}()
	return rtn, nil
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

	return nil
}
