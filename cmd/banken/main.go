package main

import (
	"context"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

var (
	logLevel       string
	logOutput      string
	alertThreshold int
)

var rootCmd = &cobra.Command{
	Use:   "banken",
	Short: "http traffic monitor for unix systems",
}

func configuration() {
	// Cobra configuration
	pflag.String("bpf", "tcp port 80 or port 443", "BPF configuration string")
	pflag.String("loglevel", "info", "verbosity of logging")
	pflag.String("logoutput", "", "logging destination, leave blank to disable")
	pflag.Int("alert-threshold", 100, "alerting threshold of http requests per 2 minute span ")
	pflag.Parse()
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Initialize Logging
	logLevel = viper.GetString("loglevel")
	logOutput = viper.GetString("logoutput")
	logLevelVal, err := log.ParseLevel(logLevel)
	logger := log.New()
	if err != nil {
		logger.Fatal("error parsing loglevel configuration")
	}
	logger.SetLevel(logLevelVal)
	if logOutput != "" {
		var lf *os.File
		var err error
		if logOutput == "stderr" {
			lf = os.Stderr
		} else {
			lf, err = os.OpenFile(logOutput, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				logger.Fatalf("unable to open %q for logging", logOutput)
			}
		}
		defer lf.Close()
		log.SetOutput(lf)
	}
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
	configuration()

	// Catch shutdown signals
	runCtx, can := context.WithCancel(context.Background())
	catchCancelSignal(can, unix.SIGINT, unix.SIGHUP, unix.SIGTERM, unix.SIGQUIT)
	defer can()

	// Initialize command

	// TODO: Initizlize Traffic Monitor alerter
	// TODO: Initialize Route Monitor

	// TODO: Initialize sniffer
	// TODO: Duplex data streams
	// TODO: Wait
	<-runCtx.Done()
}
