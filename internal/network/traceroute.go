package network

import (
	"net"
)

type Hop struct {
	IP      net.Addr
	Domains []string
}

func Traceroute(ip *net.IPAddr) ([]Hop, error) {
	hops := make([]Hop, 0)

	for i := 1; i <= 30; i++ {
		icmpPing, err := NewICMPPing()
		if err != nil {
			return nil, err
		}
		defer icmpPing.Close()

		opts := ICMPPingOpts{
			IP:  ip,
			TTL: i,
		}
		res, err := icmpPing.Ping(opts)
		if err != nil {
			return nil, err
		}

		addr, _ := net.LookupAddr(res.Peer.String())

		hops = append(hops, Hop{
			IP:      res.Peer,
			Domains: addr,
		})

		if res.Peer == ip {
			break
		}
	}

	return hops, nil
}
