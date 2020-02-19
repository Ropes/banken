// Package cmd contains the core application logic and goroutine launch points.
package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/ropes/banken/pkg/sniff"
	"github.com/ropes/banken/pkg/traffic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func detectInterfaces() ([]string, error) {
	output := make([]string, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		output = append(output, i.Name)
	}
	return output, nil
}

// Banken 番犬　App
type Banken struct {
	ctx         context.Context
	logger      *log.Logger
	debugLogger *log.Logger

	at   int
	topN int
	bpf  string

	rc     *traffic.RequestCounter
	ad     *traffic.AlertDetector
	status traffic.Notification
}

func NewBanken(ctx context.Context, at, topN int, bpf string, logger *log.Logger) *Banken {
	dl := log.New()
	dl.SetOutput(os.Stderr)

	return &Banken{
		ctx:         ctx,
		logger:      logger,
		debugLogger: dl,

		at:   at,
		topN: topN,
		bpf:  bpf,
	}
}

func (b *Banken) Init(topN, reqCnts, alerts *widgets.List) ([]string, chan sniff.HTTPXPacket, error) {
	// Detect interfaces
	ifaces, err := detectInterfaces()
	if err != nil {
		b.logger.Fatal(err)
		return nil, nil, err
	}

	// Initialize Traffic Monitor alerter
	notifications := make(chan traffic.Notification, 1)
	b.ad = traffic.NewAlertDetector(b.ctx, time.Now(), 5, notifications)
	go func(a *traffic.AlertDetector, logger *log.Logger) {
		i := 0
		for n := range notifications {
			i++
			logger.Infof("RequestRate Notification: %q", n.String())
			alerts.Rows = append(alerts.Rows, fmt.Sprintf("[%d] %s", i, n.String()))
			ui.Render(alerts)
		}
	}(b.ad, b.logger)

	// Initialize Route Counter
	b.rc = new(traffic.RequestCounter)
	rcTick := time.NewTicker(5 * time.Second)
	go func() {
		for range rcTick.C {
			m := b.rc.Export()
			f := log.Fields{}
			reqs := topNRequests(m, b.topN)
			top := make([]string, b.topN)
			for i, v := range reqs {
				s := fmt.Sprintf("%s -> %d", v.URL, v.C)
				f[fmt.Sprintf("%d", i+1)] = s
				top = append(top, fmt.Sprintf("[%d]: %s", i+1, s))
			}
			topN.Rows = top
			b.logger.WithFields(f).Infof("Top %d URLs", b.topN)

			ui.Render(topN)
			ui.Render(reqCnts)
		}
	}()

	// Initialize traffic sniffer consumers
	const consumers = 5
	packetStream := make(chan sniff.HTTPXPacket, consumers)
	// Initialze stream consumers before reading packets
	for i := 0; i < consumers; i++ {
		go func() {
			for p := range packetStream {
				// Increment traffic counter
				b.ad.Increment(1, p.TS)

				// Record the URL's route to counter
				u := HTTPURLSlug(p.Host, p.Path)
				log.Tracef("PacketConsumer received: %v", u)
				b.rc.IncKey(u, uint64(1))
			}
		}()
	}

	return ifaces, packetStream, nil
}

// Run initializes traffic capture monitors for each interface.
func (b *Banken) Run(ifaces []string, packetStream chan sniff.HTTPXPacket) {
	ctx := b.ctx
	bpfFilter := viper.GetString("bpf")
	for _, iface := range ifaces {
		go func(iface string) {
			ctxLogger := b.debugLogger.WithFields(log.Fields{"iface": iface})
			b.debugLogger.Debugf("BPF: %q", b.bpf)
			sniff.InterfaceListener(ctx, packetStream, iface, bpfFilter, 1600, ctxLogger.Logger)
		}(iface)
	}

	// Wait for stop signal
	<-ctx.Done()
}

func (b *Banken) getAlertState() traffic.Notification {
	return b.ad.GetState()
}

func (b *Banken) tsReqSpanCount(start, end time.Time) int {
	return b.ad.GetSpanCount(start, end)
}

func (b *Banken) countMap() map[string]uint64 {
	return b.rc.Export()
}
