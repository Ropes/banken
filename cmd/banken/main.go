package main

import (
	"context"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/ropes/banken/cmd/banken/cmd"
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

func main() {
	logger := configuration()

	// Catch shutdown signals
	runCtx, can := context.WithCancel(context.Background())
	catchCancelSignal(can, unix.SIGINT, unix.SIGHUP, unix.SIGTERM, unix.SIGQUIT)
	defer can()

	banken := cmd.NewBanken(runCtx, logger)

	ifaces, packets, err := banken.Init()
	if err != nil {
		can()
		logger.Fatal(err)
	}

	banken.Run(ifaces, packets)
}
