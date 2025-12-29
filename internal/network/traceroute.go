package network

import (
	"net"
)

type Hop struct {
	IP      net.Addr
	Domains []string
}

func Traceroute(ip *net.IPAddr) ([]Hop, error) {
	// hops := make([]Hop, 0)

	// for i := 1; i <= 30; i++ {
	// 	icmpPing, err := NewICMPPing()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer icmpPing.Close()

	// 	opts := ICMPPingOpts{
	// 		IP:  ip,
	// 		TTL: i,
	// 	}
	// 	rtn := make(chan ICMPPingResponse)
	// 	err = icmpPing.Ping(opts, rtn)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for res := range rtn {
	// 		switch res.Message.Type {
	// 		case ipv4.ICMPTypeEchoReply:
	// 			addr, _ := net.LookupAddr(res.Peer.String())

	// 			hops = append(hops, Hop{
	// 				IP:      res.Peer,
	// 				Domains: addr,
	// 			})

	// 			if res.Peer == ip {
	// 				break
	// 			}
	// 		case ipv4.ICMPTypeTimeExceeded:
	// 			// TODO: THIS IS THE TIMEOUT
	// 		default:
	// 			// TODO: NOT SURE HERE
	// 		}
	// 	}
	// }

	// return hops, nil
	return nil, nil
}
