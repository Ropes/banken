[番犬](https://jisho.org/word/%E7%95%AA%E7%8A%AC)
------

'Guard dog' application for monitoring HTTP traffic on local interfaces. Fun project to dig back into concurrency for potentially high load data streams. Learned some new libraries, refreshed some Go concurrency patterns, and regained appreciation of the test -race detector! Certainly could have been implemented with simpler data structures, but adding the UI later was easier thanks to component composition.

![Running with shortened alert timespan](https://user-images.githubusercontent.com/489062/74881612-7d09a900-5322-11ea-9742-44e7fe98937d.png)

## Running

`banken monitor` to launch the service!

```
./banken monitor -h
Banken 番犬(watchdog) monitors HTTP network traffic from local interfaces and analyses request sources and throughput. 
	
	Terminal UI provides statistics on traffic counts over time, and top -t (default 10) URLs requested, to the first /section/. Alerts when the HTTP traffic rate surpasses the --alert-threshold per 2 minute timespan.

	HTTP request URL paths are truncated to their first section. eg: 'http://man7.org/linux/man-pages/man1/intro.1.html' is truncated and counted as 'http://man7.org/linux'. A URL to file on first path variable gets counted as a root request. eg: 'http://man7.org/style.css' will be counted to increment 'http://man7.org/'.

	If enabled by --log-sink and --log-level, logs are written periodically recording all of the information rendered in the terminal UI.

	Using Berkley Packet Filtering; by default only port 80 is monitored for HTTP packets. However that can be configured by supplying a different BPF via --bpf.

	Press 'q' to exit.

Usage:
  banken monitor [flags]

Flags:
  -a, --alert-threshold int   alerting threshold of http requests per 2 minute span  (default 10)
  -b, --bpf string            BPF configuration string (default "tcp port 80")
  -h, --help                  help for monitor
  -t, --top-n-reqs int        top number of URL:RequestCounts to display (default 10)

Global Flags:
  -l, --log-level string   log verbosity level (default "info")
  -s, --log-sink string    logging destination, leave blank to disable (default "/tmp/banken.log")
```


## Building

### Requirements

All building and testing was done on a Linux machine. However it should be able to build on any Unix OS assuming it has Libpcap header files. `ldd banken` displays link of `libpcap.so.0.8 => /usr/lib/x86_64-linux-gnu/libpcap.so.0.8`.

* [Go 1.12+](https://golang.org/doc/install) installation for compiling.
* Debian Apt Packages for reference: `make`, `libpcap-dev`, `libpcap0.8`, `libpcap0.8-dev`
* Root privileges on machine of execution to grant packet capture abilities on binary.

### Build with Make

* `make build` will compile the binary.
* `make grant-capture` will `sudo setcap cap_net_raw,cap_net_admin=eip` grans pcap network access to the binary so it doesn't have to be run as root.
* `make run` to execute Banken; listen to network traffic, reports statistics to terminal.
* `make banken` is the one-stop-shop to build all of the above and run!

## Known Issues

* Scrolling Alert messages; is fickle...
    * Unsure as to the cause, with certain terminal sizes, scrolling of threshold alert notifications does not work. view.Run() invokes scroll function calls upon input, but for some reason it doesn't appear to function if the terminal width is smaller than ~100 wide?
    * All threshold alerts are logged, so they are still recorded, but sometimes the terminal won't let user scroll over them.
        * `cat banken.log | grep 'RequestRate Notification'` to see alert detector notifications.
* Find something else? D: Submit an [issue](https://github.com/Ropes/banken/issues/new)!

## Implementation Design

* Intercept traffic constrained by BPF from the local interfaces.
* Filter traffic down to only HTTP requests.
* Duplex HTTP requests to two consumers: AlertDetector, and Route monitor.
* Alert Detector:
    * Input data into timeseries query structure(see Acknowledgements).
    * Query request count for the past 2 minutes; if above --alert-threshold; Alert UI. Conversely test 2 minute span, and notify UI when request count has dropped below threshold.
    * Nominal vs Alerted state machine
    * Also query TS counts for past 1m, 5m, 15m, 30m, 60m, 24hr request counts and update the UI.
* Route monitor
    * Retain key'd counts of `<host>/<slug>/*` & `<host>/*`
    * Read out top N and update UI.

## Potential Improvements to make
* More integration tests. 
    * `make go-test-banken` does execute a test against actual interfaces. The testing could be expanded though.
* Flag to configure Alert detection interval.(Shorten to less than 2 minutes)
* Configurable Logging format. JSON, syslog, etc
* Monitor HTTPS data flows and record traffic bandwidth per source.
    * Record bytes traversed per source/dest.
* Smarter [anomaly detection](https://github.com/lytics/anomalyzer), which could take into acount average usage but still detect large spikes.
* TermUI resizes nicely after launch(Currently does not).
*  [termui](https://github.com/gizak/termui) bar graphs of traffic volume.

## Acknowledgements

* Hoisted source code from package of the still experimental [golang.org/x/net/internal/timseries](https://pkg.go.dev/golang.org/x/net@v0.0.0-20200202094626-16171245cfb2/internal/timeseries?tab=doc) package to manage data. The package used by `x/net/trace` for compiling traffic statistics, and provides minute-hour granularity bucketing, and automatic downsampling.
* [TermUI](https://github.com/gizak/termui) was a fun library to learn.
* GoPacket now in the standard lib is very handy!

![Banken does not approve of so many websites still using sniffable HTTP...](https://i.ytimg.com/vi/j8ctVhScNW0/hqdefault.jpg)
[* source(IDK...)](https://www.youtube.com/watch?v=j8ctVhScNW0)