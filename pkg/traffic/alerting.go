package traffic

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// Notification of AlertDetector state back to caller.
// TODO: Convert to Interface
type Notification string

// Alert caller to request limit breach.
type Alert Notification

// NominalStatus returned to caller.
type NominalStatus Notification

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
	checkInterval time.Duration
	notify        chan Notification

	localInc *uint64
	flush    *time.Ticker

	startState StateFunc
}

// NewAlertDetector initializes alerting of events when
func NewAlertDetector(ctx context.Context, now time.Time, alertThreshold int, notification chan Notification) *AlertDetector {
	m := NewMonitor(now)
	zero := uint64(0)

	ad := &AlertDetector{
		ctx:           ctx,
		upperLimit:    alertThreshold,
		testSpan:      2 * time.Minute,
		checkInterval: 2 * time.Second,
		monitor:       m,
		localInc:      &zero,
		flush:         time.NewTicker(2 * time.Second),

		notify:     notification,
		startState: Nominal,
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

// Nominal state tests the monitor time span's request count
// against the upperLimit alerting threshold.
// iff threshold is broken, switch to Alerting state and notify
// output.
func Nominal(a *AlertDetector) StateFunc {
	chkTick := time.NewTicker(a.checkInterval)
	for {
		select {
		case <-a.ctx.Done():
			return nil
		case now := <-chkTick.C:
			v := a.monitor.RecentSum(a.testSpan)
			if v > a.upperLimit { // Alerting threshold triggered
				a.notify <- Notification(fmt.Sprintf("now: %v", now))
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
	chkTick := time.NewTicker(a.checkInterval)
	for {
		select {
		case <-a.ctx.Done():
			return nil
		case now := <-chkTick.C:
			v := a.monitor.RecentSum(a.testSpan)
			if v < a.upperLimit {
				a.notify <- Notification(fmt.Sprintf("now: %v", now))
				return Nominal
			}
		}
	}
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
	// TODO: shutdown logic
}
