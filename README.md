[番犬](https://jisho.org/word/%E7%95%AA%E7%8A%AC)
------

Guard dog application monitoring HTTP traffic on local interfaces.

## Functionality Criteria

* CLI tool monitoring HTTP traffic on local interfaces for indefinite amount of time.
* Sniff HTTP(S) activity
    * Track occurrences of HTTP address requests to the depth of first path slug.
* Shell 10 second status update
    * Display traffic statistics
        * http vs https traffic
        * Total throughput
    * HTTP traffic highest domain/slug hits: http://hihi.com/hihi: 22, http://neh.wtf/: 1
* Anomaly detection
    * Configured threshold for http requests for sliding window of past 2 minutes.
    * Detect when traffic for past 2 minutes has surpassed the configured threshold.
    * Alert the shell/UI 
        * Format: `High traffic generated an alert - hits = {value}, triggered at {time}`
        * Simple: print formatted text to shell
        * Bonus: Track and print the highest number of 
        * Bonus: update termui... list??
* Proper testing of components.
* Integration testing of application.
    * Ginkgo?

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
    * 2: termui with nice updates and recording of data.
    * 3: Mix of the two possible with anomaly recordings streaming up and lower section containing the stats box/window?

