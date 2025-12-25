package main

import (
	"log/slog"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"network_monitor/internal/network"
	"network_monitor/internal/utils"
)

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))

	singlePing()
}

func singlePing() {
	opts := network.ICMPPingOpts{
		Ip: "8.8.8.8",
	}
	rm, peer, err := network.ICMPPing(opts)
	if err != nil {
		slog.Error("Ping failed", "error", err)
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
		duration := now.Sub(start)
		slog.Debug("Received echo reply", "peer", peer, "duration", duration, "start", start, "now", now)
	case ipv4.ICMPTypeDestinationUnreachable:
		slog.Debug("Destination unreachable", "peer", peer)
	case ipv4.ICMPTypeTimeExceeded:
		slog.Debug("Time exceeded", "peer", peer)
	default:
		slog.Debug("Unexpected ICMP type", "type", rm.Type, "peer", peer)
	}
}
