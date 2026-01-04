package network

import (
	"math/rand"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Hop struct {
	IP      net.Addr `json:"ip"`
	Domains []string `json:"domains,omitempty"`
}

func Traceroute(ip *net.IPAddr) ([]Hop, error) {
	hops := make([]Hop, 0)

loop:
	for i := 1; i <= 30; i++ {
		icmpPing, err := NewICMPPing()
		if err != nil {
			return nil, err
		}
		defer icmpPing.Close()

		rtn, err := icmpPing.Read(3 * time.Second)

		id := rand.Intn(0xffff)

		opts := ICMPPingOpts{
			IP:  ip,
			TTL: i,
			Seq: i,
			id:  id,
		}
		err = icmpPing.Ping(opts)
		if err != nil {
			return nil, err
		}

	read:
		for res := range rtn {
			switch res.Message.Type {
			case ipv4.ICMPTypeEchoReply:
				addr, _ := net.LookupAddr(res.Peer.String())

				hops = append(hops, Hop{
					IP:      res.Peer,
					Domains: addr,
				})

				body := res.Message.Body.(*icmp.Echo)
				if body.ID == id && body.Seq == i {
					break loop
				}
			case ipv4.ICMPTypeTimeExceeded:
				hops = append(hops, Hop{
					IP: res.Peer,
				})
				break read
			}
		}
	}

	return hops, nil
}
