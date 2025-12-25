package main

import (
	"flag"
	"fmt"
	"log"
	"slices"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"network_monitor/internal/network"
	"network_monitor/internal/utils"
)

func main() {
	supportedModes := []string{"single", "continuous", "traceroute"}
	mode := flag.String("mode", "single", fmt.Sprintf("One of %v", supportedModes))
	flag.Parse()
	if !slices.Contains(supportedModes, *mode) {
		log.Fatalf("Unsupported mode, needs one of %v", supportedModes)
	}

	switch *mode {
	case "single":
		singlePing()
	default:
		log.Fatalf("%v mode not supported yet", *mode)
	}
}

func singlePing() {
	opts := network.ICMPPingOpts{
		Ip: "8.8.8.8",
	}
	rm, peer, err := network.ICMPPing(opts)
	if err != nil {
		log.Fatalf("Ping failed: %v", err)
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		body, ok := rm.Body.(*icmp.Echo)
		if !ok {
			log.Fatalf("Unexpected ICMP reply data: %v", body.Data)
		}
		if len(body.Data) < 8 {
			log.Fatalf("Echo reply data too short, received length: %v", len(body.Data))
		}

		start := utils.BinaryToTime(body.Data[:8])
		now := time.Now()
		duration := now.Sub(start)
		log.Printf("Received echo reply from %v, duration: %v", peer, duration)
	case ipv4.ICMPTypeTimeExceeded:
		log.Printf("Time exceeded from %v", peer)
	case ipv4.ICMPTypeDestinationUnreachable:
		log.Printf("Destination unreachable")
	default:
		log.Fatalf("Unexpected ICMP type: %v", rm.Type)
	}
}
