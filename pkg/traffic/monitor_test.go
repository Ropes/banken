package traffic

import (
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	const nominalInc = 2
	now := time.Now()
	tsNow := now.Add(20 * time.Minute)
	now = now.Add(5 * time.Minute)

	ts := NewMonitor(tsNow)
	ts.Increment(nominalInc, now)

	// Test adding older time series data, and querying the data by range.
	now = now.Add(1 * time.Minute)
	ts.Increment(nominalInc, now)
	sum := ts.RangeSum(now.Add(-2*time.Minute), now)
	if sum != 4 {
		t.Errorf("incorrect sum reported: %d", sum)
	}

	// Test summing values and excluding previous bucketed times.
	x := tsNow.Add(-3 * time.Minute)
	ts.Increment(nominalInc, x)
	sum = ts.RecentSum(4 * time.Minute)
	if sum != 2 {
		t.Errorf("incorrect sum reported for last 50min: %d", sum)
	}

}
