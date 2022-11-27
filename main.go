package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	flag "github.com/spf13/pflag"
	fastping "github.com/tatsushid/go-fastping"
)

type target struct {
	Addr   *net.IPAddr
	Index  int
	RTTs   [32]time.Duration
	CurRTT int
	Pong   bool
	Recv   int
	Lost   int
}

func (t *target) ResetPong() {
	t.Pong = false
}

func (t *target) AddRTT(rtt time.Duration) {
	if rtt == 0 {
		t.Lost++
		t.Pong = false
	} else {
		t.Recv++
		t.Pong = true
	}
	if t.CurRTT > len(t.RTTs)-1 {
		t.CurRTT = 0
	}
	t.RTTs[t.CurRTT] = rtt
	t.CurRTT++
}

func (t *target) AvgRTT() int64 {
	sumRTT := int64(0)
	sumCount := int64(0)

	for _, rtt := range t.RTTs {
		sumRTT += rtt.Nanoseconds()
		if rtt.Nanoseconds() > 0 {
			sumCount++
		}
	}
	if sumCount <= 0 {
		return 0
	}
	return sumRTT / sumCount
}

func (t *target) GetRTTs() []float64 {
	rtts := make([]float64, 0)

	for _, rtt := range t.RTTs {
		rtts = append(rtts, float64(rtt.Nanoseconds())/1000000.0)
	}
	return rtts
}

func (t *target) Status() string {
	if t.Pong == true {
		return "up"
	}
	return "down"
}

func (t *target) Loss() string {
	total := t.Recv + t.Lost
	return fmt.Sprintf("%03.2f%%", (float64(t.Lost)/float64(total))*100.0)
}

type targets map[string]*target

type pong struct {
	Addr *net.IPAddr
	RTT  time.Duration
}

func main() {
	cpong := make(chan pong)
	cidle := make(chan bool)
	targets := make(targets)
	count := 0

	csig := make(chan os.Signal, 1)
	signal.Notify(csig, os.Interrupt)
	signal.Notify(csig, syscall.SIGTERM)

	flag.IntVar(&count, "count", 0, "how many pings to send")
	flag.Parse()

	p := fastping.NewPinger()
	for idx, addr := range flag.Args() {
		ip, err := net.ResolveIPAddr("ip4:icmp", addr)
		if err != nil {
			panic(err)
		}
		p.AddIPAddr(ip)
		targets[ip.String()] = &target{
			Addr:  ip,
			Pong:  false,
			Index: idx,
		}
	}

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		cpong <- pong{
			Addr: addr,
			RTT:  rtt,
		}
	}
	p.OnIdle = func() {
		cidle <- true
	}

	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	sx, sy := ui.TerminalDimensions()

	tb := widgets.NewTable()
	tb.TextStyle = ui.NewStyle(ui.ColorWhite)
	tb.SetRect(0, 0, sx, sy)
	tb.TextAlignment = ui.AlignCenter
	tb.FillRow = true
	tb.RowStyles[0] = ui.NewStyle(ui.ColorBlack, ui.ColorWhite, ui.ModifierBold, ui.AlignLeft)

	uiEvents := ui.PollEvents()

	p.RunLoop()
	for {
		select {
		case pong := <-cpong:
			target, ok := targets[pong.Addr.String()]
			if !ok {
				continue
			}
			target.AddRTT(pong.RTT)
		case <-cidle:
			tb.Rows = make([][]string, len(targets)+1)
			tb.Rows[0] = []string{"host", "rtt", "loss", "state"}

			for _, t := range targets {
				if t.Pong == false {
					t.AddRTT(time.Duration(0 * time.Second))
				}
				status := t.Status()
				if status == "up" {
					tb.RowStyles[t.Index+1] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack)
				} else {
					tb.RowStyles[t.Index+1] = ui.NewStyle(ui.ColorWhite, ui.ColorRed)
				}
				avgrtt := t.AvgRTT()
				tb.Rows[t.Index+1] = []string{
					t.Addr.String(),
					fmt.Sprintf("%03.3fms", float64(avgrtt)/1000000),
					t.Loss(),
					status,
				}
				t.Pong = false
			}
			ui.Render(tb)
		case <-csig:
			return
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		}
	}
}
