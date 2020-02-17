package main

import (
	"context"
	"net"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/ropes/banken/cmd/banken/cmd"
	"github.com/ropes/banken/pkg/sniff"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
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

func init() {
	// Cobra configuration
	bpf = pflag.String(flagBPF, "tcp port 80 or port 443", "BPF configuration string")
	logLevel = pflag.String(flagLogLevel, "info", "verbosity of logging")
	logSink = pflag.String(flagLogSink, "", "logging destination, leave blank to disable")
	alertThreshold = pflag.Int("alert-threshold", 100, "alerting threshold of http requests per 2 minute span ")
	pflag.Parse()
	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		log.Fatal(err)
	}
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

func main() {
	logger := configuration()

	// Catch shutdown signals
	runCtx, can := context.WithCancel(context.Background())
	catchCancelSignal(can, unix.SIGINT, unix.SIGHUP, unix.SIGTERM, unix.SIGQUIT)
	defer can()

	// Initialize command
	// Detect interfaces
	ifaces, err := detectInterfaces()
	if err != nil {
		logger.Fatal(err)
	}

	// TODO: Initialize Traffic Monitor alerter
	// TODO: Initialize Route Monitor

	// Initialize sniffer
	bpfFlag := viper.GetString("bpf")
	packetStream := make(chan sniff.HTTPXPacket, 5)
	// Initialze stream consumers before reading packets
	const consumers = 5
	for i := 0; i < consumers; i++ {
		go func() {
			for p := range packetStream {
				// TODO: Increment traffic counter
				// TODO:
				log.Infof("PacketConsumer received: %v",
					cmd.HTTPURLSlug(p.Host, p.Path))
			}
		}()
	}

	for _, iface := range ifaces {
		go func(iface string) {
			ctxLogger := logger.WithFields(log.Fields{"iface": iface})
			sniff.InterfaceListener(runCtx, packetStream, iface, bpfFlag, 1600, ctxLogger.Logger)
		}(iface)
	}

	// Wait for stop signal
	<-runCtx.Done()
}
