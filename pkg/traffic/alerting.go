package traffic

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

var _ (Notification) = (*Alert)(nil)
var _ (Notification) = (*NominalStatus)(nil)
var _ (Notification) = (*NilStatus)(nil)

// Notification of AlertDetector state back to caller.
type Notification interface {
	String() string
}

// Alert indicates that HTTP request rate surpassed the limit.
type Alert struct {
	hits int
	ts   time.Time
}

// Alert formats state of alert to caller.
func (a Alert) String() string {
	return fmt.Sprintf("High traffic generated an alert --- hits = %d, triggered at %s", a.hits, a.ts.Format(time.RFC3339))
}

// NominalStatus indicates normal HTTP request rate conditions.
type NominalStatus struct {
	ts time.Time
}

// String formats state information to watcher.
func (s NominalStatus) String() string {
	return fmt.Sprintf("Traffic within nominal parameters - time: %s", s.ts.Format(time.RFC3339))
}

// NilStatus informs caller that AlertDetector state has exited operation.
type NilStatus struct{}

func (e NilStatus) String() string {
	return fmt.Sprintf("state execution has ended: %v", time.Now())
}

// StateFunc provides clean transitions between
// code execution paths.
type StateFunc func(*AlertDetector) StateFunc

// AlertDetector provides notification when traffic
// breaks nominal throughput limits.
type AlertDetector struct {
	ctx           context.Context
	monitor       *Monitor
	upperLimit    int
	testSpan      time.Duration
	testTicker    *time.Ticker
	checkInterval time.Duration
	notify        chan Notification

	localInc *uint64
	flush    *time.Ticker

	startState StateFunc
	reqState   chan struct{}
	getState   chan Notification
}

// NewAlertDetector initializes alerting of events when
func NewAlertDetector(ctx context.Context, now time.Time, alertThreshold int, notification chan Notification) *AlertDetector {
	m := NewMonitor()
	zero := uint64(0)
	testTick := time.NewTicker(2 * time.Second)

	ad := &AlertDetector{
		ctx:        ctx,
		upperLimit: alertThreshold,
		testSpan:   2 * time.Minute,
		testTicker: testTick,
		monitor:    m,
		localInc:   &zero,
		flush:      time.NewTicker(2 * time.Second),

		notify:     notification,
		startState: Nominal,
		getState:   make(chan Notification, 1),
		reqState:   make(chan struct{}, 1),
	}
	go ad.flushIncrements()
	go ad.runState()
	return ad
}

// newTestAlertDetector used to configure state for testing.
func newTestAlertDetector(ctx context.Context, alertThreshold int, notification chan Notification, state StateFunc, timeSpan time.Duration) *AlertDetector {
	m := NewMonitor()
	zero := uint64(0)
	testTick := time.NewTicker(2 * time.Second)

	ad := &AlertDetector{
		ctx:        ctx,
		upperLimit: alertThreshold,
		testSpan:   timeSpan,
		testTicker: testTick,
		monitor:    m,
		localInc:   &zero,
		flush:      time.NewTicker(2 * time.Second),

		notify:     notification,
		startState: state,
		reqState:   make(chan struct{}, 1),
		getState:   make(chan Notification, 1),
	}
	go ad.flushIncrements()
	go ad.runState()
	return ad
}

// Increment is a public method to aggregate http req counts into
// concurrency safe variable before being flushed.
func (a *AlertDetector) Increment(inc int, now time.Time) {
	atomic.AddUint64(a.localInc, uint64(inc))
}

// GetState informs caller of AlertDetector's current operation state.
// Channels are used to request and return Alert state to protect
// external mutation of the state value itself.
func (a *AlertDetector) GetState() Notification {
	a.reqState <- struct{}{}
	status := <-a.getState
	return status
}

// GetSpanCount provides access to the occurrence count within a sepcified
// time interval[start, end].
func (a *AlertDetector) GetSpanCount(start, end time.Time) int {
	return a.monitor.RangeSum(start, end)
}

func (a *AlertDetector) flushIncrements() {
	for {
		select {
		case <-a.ctx.Done():
			// Context closed, exit incrementing
			return
		case now := <-a.flush.C:
			// Extract the current value, and zero the localInc variable.
			inc := atomic.SwapUint64(a.localInc, uint64(0))
			if inc > 0 {
				a.monitor.Increment(int(inc), now)
			}
		}
	}
}

// runState operates the alert state transition logic.
func (a *AlertDetector) runState() {
	state := a.startState
	for state != nil {
		state = state(a)
	}
}

// Nominal state tests the monitor time span's request count
// against the upperLimit alerting threshold.
// iff threshold is broken, switch to Alerting state and notify
// output.
func Nominal(a *AlertDetector) StateFunc {
	for {
		select {
		case <-a.ctx.Done():
			return nil
		case <-a.reqState:
			a.getState <- NominalStatus{ts: time.Now()}
		case now := <-a.testTicker.C:
			v := a.monitor.RecentSum(a.testSpan)
			if v > a.upperLimit { // Alerting threshold triggered
				a.notify <- Alert{ts: now, hits: v}
				return Alerted
			}
		}
	}
}

// Alerted state periodically notifies the output that number of
// requests still exceeds the allowed upperLimit.
// iff monitored timespan request count drops below the uppperLimit
// the state returns to Nominal and notifies output.
func Alerted(a *AlertDetector) StateFunc {
	for {
		select {
		case <-a.ctx.Done():
			return nil
		case <-a.reqState:
			v := a.monitor.RecentSum(a.testSpan)
			a.getState <- Alert{ts: time.Now(), hits: v}
		case now := <-a.testTicker.C:
			v := a.monitor.RecentSum(a.testSpan)
			if v < a.upperLimit {
				a.notify <- NominalStatus{ts: now}
				return Nominal
			}
		}
	}
}
