package network

import (
	"net"
)

type Hop struct {
	Ip      string
	Domains []string
}

func Traceroute(ip string) ([]Hop, error) {
	hops := make([]Hop, 0)

	for i := 1; i <= 30; i++ {
		opts := ICMPPingOpts{
			IP:  ip,
			TTL: i,
		}
		_, peer, err := ICMPPing(opts)
		if err != nil {
			return nil, err
		}

		addr, _ := net.LookupAddr(peer.String())

		peerStr := peer.String()

		hops = append(hops, Hop{
			Ip:      peerStr,
			Domains: addr,
		})

		if peerStr == ip {
			break
		}
	}

	return hops, nil
}
