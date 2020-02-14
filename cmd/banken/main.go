package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	logLevel       string
	logOutput      string
	alertThreshold int
)

func main() {
	// Cobra configuration
	pflag.String("bpf", "tcp port 80 or port 443", "BPF configuration string")
	pflag.String("loglevel", "info", "verbosity of logging")
	pflag.String("logoutput", "", "logging destination, leave blank to disable")
	pflag.Int("alert-threshold", 100, "alerting threshold of http requests per 2 minute span ")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// Initialize Logging
	logLevel = viper.GetString("loglevel")
	logLevelVal, err := log.ParseLevel(logLevel)
	logger := log.New()
	if err != nil {
		logger.Fatal("error parsing loglevel configuration")
	}
	logger.SetLevel(logLevelVal)

	// TODO: Initizlize Traffic Monitor alerter
	// TODO: Initialize Route Monitor

	// TODO: Initialize sniffer
	// TODO: Duplex data streams
}
