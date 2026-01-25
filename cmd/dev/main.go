/* Probably best to ignore this, it's just a scratchpad */
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"slices"
	"strings"
	"time"

	"network_monitor/internal/network"
)

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))

	supportedModes := []string{"single", "continuous", "traceroute"}
	mode := flag.String("mode", "single", fmt.Sprintf("One of %v", supportedModes))
	flag.Parse()
	if !slices.Contains(supportedModes, *mode) {
		slog.Error("Unsupported mode", "supported", supportedModes)
		os.Exit(1)
	}

	switch *mode {
	case "single":
		slog.Debug("single mode")
		singlePing()
	case "continuous":
		slog.Debug("continuouse mode")
		continuousPing()
	case "traceroute":
		slog.Debug("traceroute mode")
		traceroute()
	default:
		slog.Error("Mode not supported yet", "mode", *mode)
		os.Exit(1)
	}
}

func singlePing() {
	dest, err := net.ResolveIPAddr("ip4:icmp", "8.8.8.8")
	if err != nil {
		slog.Error("Failed to resolve IP address")
		os.Exit(1)
	}
	icmpPing, err := network.NewICMPPing()
	if err != nil {
		slog.Error("ICMPPing creation failed", "error", err)
	}
	defer icmpPing.Close()

	rtn, err := icmpPing.Read(3 * time.Second)
	opts := network.ICMPPingOpts{
		IP: dest,
	}

	err = icmpPing.Ping(opts)
	if err != nil {
		slog.Error("Ping failed", "error", err)
	}

	for msg := range rtn {
		slog.Debug("Received message", "message", msg.Message.Type)
	}
}

func continuousPing() {
	pl, err := network.NewPingLoop(5)
	if err != nil {
		slog.Error("Error creating PingLoop", "error", err)
		os.Exit(1)
	}
	addIp(pl, "8.8.8.8")
	addIp(pl, "79.79.79.79")

	resChan, idleChan := make(chan *network.PingLoopResponse), make(chan bool)
	pl.OnResponse = func(res *network.PingLoopResponse) {
		resChan <- res
	}
	pl.OnIntervalEnd = func(_, _ int) {
		idleChan <- true
	}

	if err = pl.Run(); err != nil {
		slog.Error("Unable to Run", "error", err)
		os.Exit(1)
	}

	for {
		select {
		case res := <-resChan:
			slog.Debug("On response", "ip", res.Peer, "duration", res.Duration, "seq", res.Body.Seq)
		case <-idleChan:
			slog.Debug("Interval end")
		}
	}
}

func addIp(pl *network.PingLoop, ip string) {
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		slog.Error("Error resolving IP", "error", err.Error())
		os.Exit(1)
	}
	pl.AddIpAddr(ra)
}

func traceroute() {
	dest, err := net.ResolveIPAddr("ip4:icmp", "8.8.8.8")
	if err != nil {
		slog.Error("Failed to resolve IP address")
		os.Exit(1)
	}

	hops, err := network.Traceroute(dest)
	if err != nil {
		slog.Error("Error from traceroute", "error", err)
	}

	var hopStr strings.Builder
	for _, hop := range hops {
		ip := hop.IP.String()
		if ip == "" {
			continue
		}
		hopStr.WriteString(ip)
		if len(hop.Domains) > 0 {
			d := strings.Join(hop.Domains, ",")
			fmt.Fprintf(&hopStr, " (%s)", d)
		}
		hopStr.WriteByte('\n')
	}
	slog.Debug(hopStr.String())
}
