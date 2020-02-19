package cmd

import (
	"context"
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ropes/banken/pkg/sniff"
	"github.com/ropes/banken/pkg/traffic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagLogLevel = "loglevel"
	flagLogSink  = "logsink"
	flagBPF      = "bpf"
)

var (
	bpf            *string
	logLevel       *string
	logSink        *string
	alertThreshold *int
)
var rootCmd = &cobra.Command{
	Use:   "banken",
	Short: "http traffic monitor for unix systems",
}

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

	status traffic.Notification
}

func NewBanken(ctx context.Context, logger *log.Logger) *Banken {
	return &Banken{
		ctx:         ctx,
		logger:      logger,
		debugLogger: logger,
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
	ad := traffic.NewAlertDetector(b.ctx, time.Now(), 5, notifications)
	go func(a *traffic.AlertDetector, logger *log.Logger) {
		for n := range notifications {
			logger.Infof("Alert Notification: %q", n.String())
		}
	}(ad, b.logger)

	// Initialize Route Counter
	rc := new(traffic.RequestCounter)
	rcTick := time.NewTicker(10 * time.Second)
	go func() {
		for range rcTick.C {
			m := rc.Export()
			b.debugLogger.Infof("Common URLs")
			for k, v := range m {
				b.logger.Infof("%v %d", k, v)
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
				ad.Increment(1, p.TS)

				// Record the URL's route to counter
				u := HTTPURLSlug(p.Host, p.Path)
				log.Tracef("PacketConsumer received: %v", u)
				rc.IncKey(u, uint64(1))
			}
		}()
	}

	return ifaces, packetStream, nil
}

func (b *Banken) Run(ifaces []string, packetStream chan sniff.HTTPXPacket) {
	ctx := b.ctx
	// Initialize sniffer for each interface
	bpfFilter := viper.GetString("bpf")
	for _, iface := range ifaces {
		go func(iface string) {
			ctxLogger := b.debugLogger.WithFields(log.Fields{"iface": iface})
			sniff.InterfaceListener(ctx, packetStream, iface, bpfFilter, 1600, ctxLogger.Logger)
		}(iface)
	}

	// Wait for stop signal
	<-ctx.Done()
}
