// Package cmd contains the core application logic and goroutine launch points.
package cmd

import (
	"context"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/ropes/banken/pkg/sniff"
	"github.com/ropes/banken/pkg/traffic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	flagLogLevel     = "log-level"
	flagLogSink      = "log-sink"
	flagDebugLogSink = "debug-log-sink"
	flagBPF          = "bpf"
)

var (
	bpf            *string
	logLevel       *string
	logSink        *string
	debugLogSink   *string
	alertThreshold *int
)

func configuration() *log.Logger {
	// Initialize Logging
	logLevelVal, err := log.ParseLevel(*logLevel)
	logger := log.New()
	if err != nil {
		logger.Fatalf("error parsing loglevel configuration: %v", err)
	}
	logger.SetLevel(logLevelVal)
	if *logSink != "" {
		var lf *os.File
		var err error
		if *logSink == "stderr" {
			lf = os.Stderr
		} else {
			lf, err = os.OpenFile(*logSink, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				logger.Fatalf("unable to open %q for logging", *logSink)
			}
		}
		defer lf.Close()
		log.SetOutput(lf)
	} else {
		logger.SetOutput(os.Stdout)
	}
	return logger
}

func catchCancelSignal(can context.CancelFunc, sig ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	go func() {
		<-c
		can()
	}()
}

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

func (b *Banken) Init() ([]string, chan sniff.HTTPXPacket, error) {
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
		for n := range notifications {
			logger.Infof("RequestRate Notification: %q", n.String())
		}
	}(b.ad, b.logger)

	// Initialize Route Counter
	b.rc = new(traffic.RequestCounter)
	rcTick := time.NewTicker(10 * time.Second)
	go func() {
		for range rcTick.C {
			m := b.rc.Export()
			b.logger.Infof("Top %d URLs", b.topN)
			reqs := topNRequests(m, b.topN)
			for _, v := range reqs {
				b.logger.Infof("%v %d", v.URL, v.C)
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

// Run initializes traffic capture monitors for each interface.
func (b *Banken) Run(ifaces []string, packetStream chan sniff.HTTPXPacket) {
	ctx := b.ctx
	bpfFilter := viper.GetString("bpf")
	for _, iface := range ifaces {
		go func(iface string) {
			ctxLogger := b.debugLogger.WithFields(log.Fields{"iface": iface})
			b.debugLogger.Infof("BPF: %q", b.bpf)
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
