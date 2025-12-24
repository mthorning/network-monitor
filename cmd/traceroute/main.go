package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"network-monitor/internal/utils"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type PingProtocol string

const (
	ProtocolICMP PingProtocol = "icmp"
	ProtocolUDP  PingProtocol = "udp"
	ProtocolTCP  PingProtocol = "tcp"
)

type Opts struct {
	Protocol             PingProtocol
	Ip                   string
	ReadDeadlineDuration time.Duration
	TTL                  int
}

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))

	opts := Opts{
		Ip:                   "8.8.8.8",
		Protocol:             ProtocolICMP,
		ReadDeadlineDuration: 3 * time.Second,
		TTL:                  64,
	}

	if err := Run(&opts); err != nil {
		slog.Error("An Error occurred", "error", err)
	}
}

func Run(o *Opts) error {
	switch o.Protocol {
	case ProtocolICMP:
		return icmpPing(o)
	default:
		return errors.New(fmt.Sprintf("Unsupported protocol: %v", o.Protocol))
	}
}

func icmpPing(opts *Opts) error {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return err
	}
	defer c.Close()

	pc := c.IPv4PacketConn()
	if err := pc.SetTTL(opts.TTL); err != nil {
		return err
	}

	now := time.Now()
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:   os.Getpid(),
			Seq:  1,
			Data: utils.TimeToBinary(&now),
		},
	}

	mb, err := m.Marshal(nil)
	if err != nil {
		return err
	}

	dest, err := net.ResolveIPAddr("ip4:icmp", opts.Ip)
	if _, err := c.WriteTo(mb, dest); err != nil {
		return err
	}

	if err := c.SetReadDeadline(time.Now().Add(opts.ReadDeadlineDuration)); err != nil {
		return err
	}

	buf := make([]byte, 1500)
	n, peer, err := c.ReadFrom(buf)
	if err != nil {
		return err
	}

	rm, err := icmp.ParseMessage(1, buf[:n])
	if err != nil {
		return err
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		body, ok := rm.Body.(*icmp.Echo)
		if !ok {
			slog.Error("Unexpected ICMP reply data", "data", body.Data)
		}
		if len(body.Data) < 8 {
			slog.Error("echo reply data too short", "data", body.Data, "length", len(body.Data))
		}

		start := utils.BinaryToTime(body.Data[:8])
		now := time.Now()
		duration := now.Sub(*start)
		slog.Debug("Received echo reply", "peer", peer, "bytes", n, "duration", duration, "start", *start, "now", now)
	case ipv4.ICMPTypeDestinationUnreachable:
		slog.Debug("Destination unreachable", "peer", peer)
	case ipv4.ICMPTypeTimeExceeded:
		slog.Debug("Time exceeded", "peer", peer)
	default:
		slog.Debug("Unexpected ICMP type", "type", rm.Type, "peer", peer)
	}

	return nil
}
