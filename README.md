[番犬](https://jisho.org/word/%E7%95%AA%E7%8A%AC)
------

'Guard dog' application for monitoring HTTP traffic on local interfaces. Fun project to dig back into concurrency for potentially high load data streams.

## Running


## Building

### Requirements

All building and testing was done on a Linux machine. However it should be able to build on any Unix OS assuming it has Libpcap header files. `ldd banken` displays link of `libpcap.so.0.8 => /usr/lib/x86_64-linux-gnu/libpcap.so.0.8`.

* [Go 1.12+](https://golang.org/doc/install) installation for compiling.
* Debian Apt Packages for reference: `make`, `libpcap-dev`, `libpcap0.8`, `libpcap0.8-dev`
* Root privileges on machine of execution to grant packet capture abilities on binary.

### Build with Make

* `make build` will compile the binary.
* `make grant-capture` will `setcap cap_net_raw,cap_net_admin=eip` on the binary.
* `make run` to execute Banken; listen to network traffic, reports statistics to terminal.

## Known Issues

* Scrolling Alert messages, doesn't always work.
    * Unsure as to the cause, with certain terminal sizes, scrolling of threshold alert notifications does not work. view.Run() invokes scroll function calls upon input, but for some reason it doesn't appear to function if the terminal width is smaller than ~100?
    * All threshold alerts are logged, so they are still recorded.
        * `cat banken.log | grep 'RequestRate Notification'`

## Functionality Criteria

* CLI tool monitoring HTTP traffic crossing local interfaces.
* Shell 10 second status update
    * Display traffic statistics
        * http vs https traffic
        * Total throughput
    * HTTP traffic highest domain/section hits: http://hihi.com/hihi: 22, http://neh.wtf/: 1
* Anomaly detection
    * Configured threshold for http requests for sliding window of the past 2 minutes.
    * When traffic for past 2 minutes has surpassed the configured threshold, signal an alert message.
        * Format: `High traffic generated an alert - hits = {value}, triggered at {time}`
        * Simple: print formatted text to shell
        * Bonus: Track and print the highest number of 
* Proper testing of components.
* Integration testing of application.

## Implementation Design

* Intercept all traffic from the local interfaces.
* Filter traffic down to only HTTP requests.
* Duplex HTTP requests to two consumers: Anomaly detector, and Route monitor.
* Anomaly Detector:
    * Input: Scalar count of HTTP reqs per 10s.
    * Read HttpReq's and construct time series counter to detect if threshold is breached.
    * Anomalyzer as a bonus to detect anomalies without relying on a static threshold setting.
    * Nominal vs Alerted state machine
        * Alerted state tracks the HTTP request/paths
* Route monitor
    * Retain key'd counts of `<host>/<slug>/*`
    * 

## Components

* Configuration 
    * Read config information from EnvVars or CLI Flags.
    * Intercept kill signals and dump stats on session at end of runtime.
* Debugging
    * Debug logging to StdErr
    * Metrics to track?

## Improvements to make
* More integration tests. 
    * `make go-test-banken` does execute a test against actual interfaces. The testing could be expanded though.
* Configurable Logging format.
* Monitor HTTPS data flows and record traffic bandwidth per source.
    * Record bytes traversed per source/dest.
* Smarter [anomaly detection](https://github.com/lytics/anomalyzer), which could take into acount average usage but still detect large spikes.
* TermUI resizes nicely after launch(Currently does not).
* Use [termui](https://github.com/gizak/termui) to add bar graphs of traffic volume.

## Acknowledgements

* Hoisted source code from package of the still experimental [golang.org/x/net/internal/timseries](https://pkg.go.dev/golang.org/x/net@v0.0.0-20200202094626-16171245cfb2/internal/timeseries?tab=doc) package to manage data. The package used by `x/net/trace` for compiling traffic statistics, and provides minute-hour granularity bucketing, and automatic downsampling.
* [TermUI](https://github.com/gizak/termui) was a fun library to learn.
* GoPacket now in the standard lib is very handy!

![Banken does not approve of so many websites still using sniffable HTTP...](https://i.ytimg.com/vi/j8ctVhScNW0/hqdefault.jpg)
[* source(IDK...)](https://www.youtube.com/watch?v=j8ctVhScNW0)