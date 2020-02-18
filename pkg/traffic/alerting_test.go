package traffic

import (
	"context"
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

func TestBasicAlert(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	step := start.Add(1 * time.Minute)
	notify := make(chan Notification, 1)

	ad := NewAlertDetector(ctx, start, 10, notify)

	for i := 0; i < 50; i++ {
		t := step.Add(time.Duration(i) * time.Second)
		ad.Increment(1, t)
	}

	expNotification := <-notify
	switch expNotification.(type) {
	case Alert:
		t.Logf("expected Alert returned: %v", expNotification)
	default:
		t.Errorf("notification should be an alert: %v", expNotification)
	}

	ad.testSpan = 1 * time.Minute
	expNotification = <-notify
	switch expNotification.(type) {
	case NominalStatus:
		t.Logf("expected NominalStatus returned: %v", expNotification)
	default:
		t.Errorf("notification should be a NominalStatus: %v", expNotification)
	}
}
