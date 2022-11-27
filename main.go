package main

import (
	"fmt"
	"net"
	"time"

	flag "github.com/spf13/pflag"
	fastping "github.com/tatsushid/go-fastping"
)

type target struct {
	Addr   *net.IPAddr
	RTTs   [32]time.Duration
	CurRTT int
	Pong   bool
	Recv   int
	Lost   int
}

type targets map[string]*target

func main() {
	targets := make(targets)
	count := 0

	flag.IntVar(&count, "count", 0, "how many pings to send")
	flag.Parse()

	p := fastping.NewPinger()
	for _, addr := range flag.Args() {
		ip, err := net.ResolveIPAddr("ip4:icmp", addr)
		if err != nil {
			panic(err)
		}
		p.AddIPAddr(ip)
		targets[ip.String()] = &target{
			Addr: ip,
			Pong: false,
		}
	}

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		//fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		if t, ok := targets[addr.String()]; !ok {
			return
		} else {
			if t.CurRTT > len(t.RTTs)-1 {
				t.CurRTT = 0
			}
			t.RTTs[t.CurRTT] = rtt
			t.CurRTT++
			t.Pong = true
		}
	}
	p.OnIdle = func() {
		fmt.Println("IDLE")
		for _, t := range targets {
			sumRTT := int64(0)
			sumCount := int64(0)

			if t.Pong == false {
				t.RTTs[t.CurRTT] = time.Duration(0 * time.Second)
			}
			for _, rtt := range t.RTTs {
				sumRTT += rtt.Nanoseconds()
				if rtt.Nanoseconds() > 0 {
					sumCount++
				}
			}
			if sumCount > 0 {
				fmt.Printf("IP Addr: %s, RTT: %d\n", t.Addr.String(), sumRTT/sumCount)
			}
			t.Pong = false
		}
	}

	p.RunLoop()
	for {
	}
}
