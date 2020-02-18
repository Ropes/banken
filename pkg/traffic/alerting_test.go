package traffic

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestAtomicUnderstanding(t *testing.T) {
	x := uint64(0)
	localInc := &x

	// Addition
	atomic.AddUint64(localInc, uint64(1))
	if int(atomic.LoadUint64(localInc)) != 1 {
		t.Errorf("Add* function should set original variable to 1")
	}

	// Read & Swap
	v := atomic.SwapUint64(localInc, uint64(0))
	if int(v) != 1 && int(atomic.LoadUint64(localInc)) != 0 {
		t.Errorf("Returned value should have been previous increment, and localInc should be zeroed?")
	}

}

func TestGetStateLogic(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	notify := make(chan Notification, 1)

	ad := newTestAlertDetector(ctx, start, 10, notify)

	switch reflect.ValueOf(ad.activeState).Pointer() {
	case reflect.ValueOf(Nominal).Pointer():
		t.Logf("Nominalstate detected")
	case reflect.ValueOf(Alerted).Pointer():
		v := ad.monitor.RecentSum(ad.testSpan)
		t.Errorf("alert detected %d", v)
	default:
		t.Errorf("ad state: %v %v %v", reflect.ValueOf(ad.activeState), reflect.ValueOf(Nominal),
			"neh")
	}
}

func TestBasicAlert(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	step := start.Add(1 * time.Minute)
	notify := make(chan Notification, 1)

	//ad := newTestAlertDetector(ctx, start.Add(2*time.Minute), 10, notify)
	ad := NewAlertDetector(ctx, start, 10, notify)
	ad.testSpan = 1 * time.Minute
	// Assert state is nominal
	state := ad.GetState()
	if reflect.TypeOf(state) != reflect.TypeOf(NominalStatus{}) {
		t.Errorf("types did not match: %v", reflect.TypeOf(state))
	}

	// Add data to monitor
	for i := 0; i < 50; i++ {
		t := step.Add(time.Duration(i) * time.Second)
		ad.Increment(1, t)
	}

	// Wait for notification that request limit was breached.
	expNotification := <-notify
	switch expNotification.(type) {
	case Alert:
		t.Logf("expected Alert returned: %v", expNotification)
		s := ad.GetState()
		if reflect.TypeOf(s) != reflect.TypeOf(Alert{}) {
			t.Errorf("types did not match: %v", reflect.TypeOf(s))
		}
	default:
		t.Errorf("notification should be an alert: %v", expNotification)
	}
}

func TestExitAlertStatus(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	notify := make(chan Notification, 1)

	//ad := newTestAlertDetector(ctx, start.Add(2*time.Minute), 10, notify)
	ad := NewAlertDetector(ctx, start, 10, notify)
	ad.testSpan = 1 * time.Minute
	ad.activeState = Alerted // Set test state to Alerted

	// Assert state is Alerted
	state := ad.GetState()
	if reflect.TypeOf(state) != reflect.TypeOf(Alert{}) {
		t.Errorf("types did not match: %v", reflect.TypeOf(state))
	}

	expNotification := <-notify
	switch expNotification.(type) {
	case NominalStatus:
		t.Logf("expected NominalStatus returned: %v", expNotification)
		s := ad.GetState()
		if reflect.TypeOf(s) != reflect.TypeOf(NominalStatus{}) {
			t.Errorf("types did not match: %v", reflect.TypeOf(s))
		}
	default:
		t.Errorf("notification should be a NominalStatus: %v", expNotification)
	}
}
