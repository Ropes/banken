// Package cmd contains the core application logic and goroutine launch points.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/ropes/banken/pkg/sniff"
	"github.com/ropes/banken/pkg/traffic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Banken 番犬　App manages launching consumers of network traffic data
// as well as the data storage structures. After startup it operates updating
// the UI data structures from the models.
type Banken struct {
	ctx    context.Context
	logger *log.Logger

	at   int
	topN int
	bpf  string

	rc     *traffic.RequestCounter
	ad     *traffic.AlertDetector
	status traffic.Notification
}

// NewBanken initiates instance with at:AlertThreshold, topN: Top N(umber) of
// URLs to display, and bpf (filter) string to configure packet capture.
func NewBanken(ctx context.Context, at, topN int, bpf string, logger *log.Logger) *Banken {
	dl := log.New()
	dl.SetOutput(os.Stderr)

	return &Banken{
		ctx:    ctx,
		logger: logger,

		at:   at,
		topN: topN,
		bpf:  bpf,
	}
}

// Init launches all consumers of the collected packet data models, then logs
// and updates the UI with http traffic status.
func (b *Banken) Init(topN, reqCnts, alerts *widgets.List) ([]string, chan sniff.HTTPXPacket, error) {
	// Detect interfaces
	ifaces, err := sniff.DetectInterfaces()
	if err != nil {
		b.logger.Fatal(err)
		return nil, nil, err
	}

	// Initialize Traffic Monitor alerter
	notifications := make(chan traffic.Notification, 1)
	b.ad = traffic.NewAlertDetector(b.ctx, time.Now(), b.at, notifications)
	go func(a *traffic.AlertDetector, logger *log.Logger) {
		i := 0
		for n := range notifications {
			i++
			logger.Infof("RequestRate Notification: %s", n.String())
			if alerts != nil {
				alerts.Rows = append(alerts.Rows, fmt.Sprintf("[%d] %s", i, n.String()))
				ui.Render(alerts)
			}
		}
	}(b.ad, b.logger)

	// Initialize Route Counter
	b.rc = new(traffic.RequestCounter)
	rcTick := time.NewTicker(5 * time.Second)
	intervals := []struct {
		s string
		t time.Duration
	}{
		{
			s: "1m",
			t: 1 * time.Minute,
		},
		{
			s: "5m",
			t: 5 * time.Minute,
		},
		{
			s: "15m",
			t: 15 * time.Minute,
		},
		{
			s: "30m",
			t: 30 * time.Minute,
		},
		{
			s: "60m",
			t: 60 * time.Minute,
		},
		{
			s: "24hr",
			t: 24 * time.Hour,
		},
	}
	go func() {
		for range rcTick.C {
			m := b.rc.Export()
			f := log.Fields{}
			reqs := topNRequests(m, b.topN)
			top := make([]string, 0)
			for i, v := range reqs {
				s := fmt.Sprintf("%s -> %d", v.URL, v.C)
				f[fmt.Sprintf("%d", i+1)] = s
				top = append(top, fmt.Sprintf("[%d]: %s", i+1, s))
			}
			b.logger.WithFields(f).Infof("Top %d URLs", b.topN)

			counts := make([]string, 0)
			for _, i := range intervals {
				now := time.Now()
				c := b.ad.GetSpanCount(now.Add(-i.t), now)
				if c > 0 {
					counts = append(counts, fmt.Sprintf("%s: %d", i.s, c))
				}
			}

			if topN != nil && reqCnts != nil {
				if len(top) == 0 {
					top = []string{"waiting for http traffic..."}
				}
				topN.Rows = top
				ui.Render(topN)
				reqCnts.Rows = counts
				ui.Render(reqCnts)
			}
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

// Run initializes traffic capture for each interface, feeding data
// to analysis models.
func (b *Banken) Run(ifaces []string, packetStream chan sniff.HTTPXPacket) {
	ctx := b.ctx
	bpfFilter := viper.GetString("bpf")
	for _, iface := range ifaces {
		go func(iface string) {
			ctxLogger := b.logger.WithFields(log.Fields{"iface": iface})
			b.logger.Debugf("BPF: %q", b.bpf)
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
