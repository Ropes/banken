package traffic

import (
	"context"
	"time"
)

// StateFunc provides clean transitions between
// code execution paths.
type StateFunc func(*AlertDetector) StateFunc

// AlertDetector provides notification when traffic
// breaks nominal throughput limits.
type AlertDetector struct {
	ctx           context.Context
	monitor       *Monitor
	upperLimit    int
	checkInterval time.Duration
	notify        chan string

	startState StateFunc
}

// NewAlertDetector initializes alerting of events when
func NewAlertDetector(ctx context.Context, now time.Time, alertThreshold int, notification chan string) *AlertDetector {
	m := NewMonitor(now)

	ad := &AlertDetector{
		ctx:           ctx,
		monitor:       m,
		upperLimit:    alertThreshold,
		checkInterval: 2 * time.Second,
		notify:        notification,
		startState:    Nominal,
		// TODO: output
	}
	go ad.runState()
	return ad
}

// Nominal state tests the monitor time span's request count
// against the upperLimit alerting threshold.
// iff threshold is broken, switch to Alerting state and notify
// output.
func Nominal(a *AlertDetector) StateFunc {
	return Alerted
}

// Alerted state periodically notifies the output that number of
// requests still exceeds the allowed upperLimit.
// iff monitored timespan request count drops below the uppperLimit
// the state returns to Nominal and notifies output.
func Alerted(a *AlertDetector) StateFunc {
	return Nominal
}

// runState state transition logic.
func (a *AlertDetector) runState() {
	state := a.startState
	for state != nil {
		state = state(a)
	}
	// TODO: shutdown logic
}
