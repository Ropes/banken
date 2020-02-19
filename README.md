[番犬](https://jisho.org/word/%E7%95%AA%E7%8A%AC)
------

Guard dog application for monitoring HTTP traffic on local interfaces.




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
* UI: ???
    * 1: Normal shell output with specified information dumps.

## Improvements to make
* More integration tests. 
    * `make go-test-banken` does execute a test against actual interfaces.
* Monitor HTTPS data flows and record traffic bandwidth per source.
    * Record bytes traversed per source/dest.
* Smarter [anomaly detection](https://github.com/lytics/anomalyzer), which could take into acount average usage but still detect large spikes.
* TermUI resizes nicely after launch(Currently does not).
* Use [termui](https://github.com/gizak/termui) for a clean updating interface.
    * Mix of the two possible with anomaly recordings streaming up and lower section containing the stats box/window?

## Acknowledgements

* Hoisted source code from package of the still experimental [golang.org/x/net/internal/timseries](https://pkg.go.dev/golang.org/x/net@v0.0.0-20200202094626-16171245cfb2/internal/timeseries?tab=doc) package to manage data. The package used by `x/net/trace` for compiling traffic statistics, and provides minute-hour granularity bucketing, and automatic downsampling.
* [TermUI](https://github.com/gizak/termui)
