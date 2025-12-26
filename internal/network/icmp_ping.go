package network

import (
	"errors"
	"net"
	"network_monitor/internal/utils"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPPingOpts struct {
	IP                   string
	ReadDeadlineDuration time.Duration
	TTL                  int
}

func ICMPPing(opts ICMPPingOpts) (*icmp.Message, net.Addr, error) {
	if err := checkOpts(&opts); err != nil {
		return nil, nil, err
	}

	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, nil, err
	}
	defer c.Close()

	pc := c.IPv4PacketConn()
	if err := pc.SetTTL(opts.TTL); err != nil {
		return nil, nil, err
	}

	now := time.Now()
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:   os.Getpid(),
			Seq:  1,
			Data: utils.TimeToBinary(now),
		},
	}

	mb, err := m.Marshal(nil)
	if err != nil {
		return nil, nil, err
	}

	dest, err := net.ResolveIPAddr("ip4:icmp", opts.IP)
	if _, err := c.WriteTo(mb, dest); err != nil {
		return nil, nil, err
	}

	if err := c.SetReadDeadline(time.Now().Add(opts.ReadDeadlineDuration)); err != nil {
		return nil, nil, err
	}

	buf := make([]byte, 1500)
	n, peer, err := c.ReadFrom(buf)
	if err != nil {
		return nil, nil, err
	}

	rm, err := icmp.ParseMessage(1, buf[:n])
	if err != nil {
		return nil, nil, err
	}

	return rm, peer, nil
}

func checkOpts(opts *ICMPPingOpts) error {
	if opts.IP == "" {
		return errors.New("opts.Ip is required, no value set")
	}
	if opts.TTL == 0 {
		opts.TTL = 64
	}
	if opts.ReadDeadlineDuration == 0 {
		opts.ReadDeadlineDuration = 3 * time.Second
	}

	return nil
}
