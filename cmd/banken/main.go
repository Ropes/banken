package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/ropes/banken/cmd/banken/cmd"
	"github.com/ropes/banken/pkg/view"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const (
	flagLogLevel     = "log-level"
	flagLogSink      = "log-sink"
	flagDebugLogSink = "debug-log-sink"
	flagBPF          = "bpf"
	flagTopReqs      = "top-n-reqs"
	flagAlertThresh  = "alert-threshold"
)

var (
	bpf            string
	logLevel       string
	logSink        string
	debugLogSink   string
	alertThreshold int
	topNReqs       int
)

func init() {
	// Cobra configuration
	rootCmd.PersistentFlags().StringVarP(&logLevel, flagLogLevel, "l", "info", "log verbosity level")
	rootCmd.PersistentFlags().StringVarP(&logSink, flagLogSink, "s", "/tmp/banken.log", "logging destination, leave blank to disable")
	rootCmd.PersistentFlags().StringVarP(&debugLogSink, flagDebugLogSink, "d", "stderr", "debug logging destination")

	monitor.PersistentFlags().StringVarP(&bpf, flagBPF, "b", "tcp port 80 or port 443", "BPF configuration string")
	monitor.PersistentFlags().IntVarP(&alertThreshold, flagAlertThresh, "a", 100, "alerting threshold of http requests per 2 minute span ")
	monitor.PersistentFlags().IntVarP(&topNReqs, flagTopReqs, "t", 10, "top number of URL:RequestCounts to display")
}

var rootCmd = &cobra.Command{
	Use:   "banken",
	Short: "番犬(watchdog) HTTP traffic monitor for unix systems",
	Long: `番犬(watchdog) HTTP traffic monitor for unix systems.

	Utilizes LibPCAP to read network traffic from local interfaces and parse HTTP requests.
	
	`,
}

var monitor = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor http traffic request destinations, counts, and notify when requests exceed alert threshold.",
	Long: `Utilizes LibPCAP to read network traffic from local interfaces and parse HTTP requests.

	 
	`,
	Run: func(cobraCmd *cobra.Command, args []string) {
		logger := configuration()

		// Catch shutdown signals
		runCtx, can := context.WithCancel(context.Background())
		defer can()
		catchCancelSignal(can, unix.SIGINT, unix.SIGHUP, unix.SIGTERM, unix.SIGQUIT)

		banken := cmd.NewBanken(runCtx, alertThreshold, topNReqs, bpf, logger)

		// Initialize View and Banken data models
		topN, reqCnts, alerts := view.Init(runCtx, topNReqs)
		ifaces, packets, err := banken.Init(topN, reqCnts, alerts)
		if err != nil {
			can()
			logger.Fatal(err)
		}

		go func() {
			view.Run(can, topN, reqCnts, alerts)
		}()
		banken.Run(ifaces, packets)
	},
}

func configuration() *log.Logger {
	// Initialize Logging
	logLevelVal, err := log.ParseLevel(logLevel)
	logger := log.New()
	debug := log.New()
	if err != nil {
		logger.Fatalf("error parsing loglevel configuration: %v", err)
	}
	logger.SetLevel(logLevelVal)
	if logSink != "" {
		var lf *os.File
		var err error
		if logSink == "stderr" {
			lf = os.Stderr
		} else {
			lf, err = os.OpenFile(logSink, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				logger.Fatalf("unable to open %q for logging", logSink)
			}
		}
		defer lf.Close()
		debug.SetOutput(lf)
	} else {
		logger.SetOutput(os.Stdout)
		debug.SetOutput(os.Stdout)
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
	rootCmd.AddCommand(monitor)
	rootCmd.Execute()
}
