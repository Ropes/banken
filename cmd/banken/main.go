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
	flagLogLevel    = "log-level"
	flagLogSink     = "log-sink"
	flagBPF         = "bpf"
	flagTopReqs     = "top-n-reqs"
	flagAlertThresh = "alert-threshold"
)

var (
	bpf            string
	logLevel       string
	logSink        string
	alertThreshold int
	topNReqs       int
)

func init() {
	// Cobra configuration
	rootCmd.PersistentFlags().StringVarP(&logLevel, flagLogLevel, "l", "info", "log verbosity level")
	rootCmd.PersistentFlags().StringVarP(&logSink, flagLogSink, "s", "/tmp/banken.log", "logging destination, leave blank to disable")

	monitor.PersistentFlags().StringVarP(&bpf, flagBPF, "b", "tcp port 80", "BPF configuration string")
	monitor.PersistentFlags().IntVarP(&alertThreshold, flagAlertThresh, "a", 10, "alerting threshold of http requests per 2 minute span ")
	monitor.PersistentFlags().IntVarP(&topNReqs, flagTopReqs, "t", 10, "top number of URL:RequestCounts to display")
}

var rootCmd = &cobra.Command{
	Use:   "banken",
	Short: "Banken 番犬(watchdog) HTTP traffic monitor for unix systems",
	Long: `Banken 番犬(watchdog) is a network traffic diagnostic tool. 
	
	Monitors HTTP traffic for unix systems, utilizing libpcap to read network traffic from local interfaces and parse HTTP requests(currently).
	`,
}

var monitor = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor http traffic request destinations, counts, and notify when requests exceed alert threshold.",
	Long: `Banken 番犬(watchdog) monitors HTTP network traffic from local interfaces and analyses request sources and throughput. 
	
	Terminal UI provides statistics on traffic counts over time, and top -t (default 10) URLs requested, to the first /section/. Alerts when the HTTP traffic rate surpasses the --alert-threshold per 2 minute timespan.

	HTTP request URL paths are truncated to their first section. eg: 'http://man7.org/linux/man-pages/man1/intro.1.html' is truncated and counted as 'http://man7.org/linux'. A URL to file on first path variable gets counted as a root request. eg: 'http://man7.org/style.css' will be counted to increment 'http://man7.org/'.

	If enabled by --log-sink and --log-level, logs are written periodically recording all of the information rendered in the terminal UI. Set --log-sink to empty string, to flush logs into
	/dev/null.

	Using Berkley Packet Filtering; by default only port 80 is monitored for HTTP packets. However that can be configured by supplying a different BPF via --bpf.

	Press 'q' to exit.
	`,
	Run: func(cobraCmd *cobra.Command, args []string) {
		logger := logSetup()

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

func logSetup() *log.Logger {
	// Initialize Logging
	logLevelVal, err := log.ParseLevel(logLevel)
	logger := log.New()
	if err != nil {
		logger.Fatalf("error parsing loglevel configuration: %v", err)
	}
	logger.SetLevel(logLevelVal)
	if logSink != "" {
		var lf *os.File
		var err error
		// stdout is used by termui, so can't log there
		if logSink == "stderr" {
			// stderr can be redirected with 2>>/tmp/logfile
			lf = os.Stderr
		} else {
			lf, err = os.OpenFile(logSink, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				logger.Fatalf("unable to open %q for logging", logSink)
			}
		}
		logger.SetOutput(lf)
	} else {
		dn, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatalf("unable to open %q for logging", os.DevNull)
		}
		logger.SetOutput(dn)
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
