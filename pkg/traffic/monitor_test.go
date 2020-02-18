package traffic

import (
	"fmt"
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	const nominalInc = 2
	now := time.Now()
	start := now.Add(-20 * time.Minute)

	//ts := newTestMonitor(tsNow)
	ts := NewMonitor()
	ts.Increment(nominalInc, now)

	// Test adding older time series data, and querying the data by range.
	startInc := start.Add(1 * time.Minute)
	ts.Increment(nominalInc, startInc)
	sum := ts.RangeSum(start.Add(-2*time.Minute), now)
	if sum != 4 {
		t.Errorf("incorrect sum reported: %d", sum)
	}

	// Test summing values and excluding previous bucketed times.
	x := now.Add(-3 * time.Minute)
	ts.Increment(nominalInc, x)
	sum = ts.RecentSum(4 * time.Minute)
	if sum != 4 {
		t.Errorf("incorrect sum reported for last 50min: %d", sum)
	}

	// Test full range
	sum = ts.RangeSum(start, now)
	if sum != 6 {
		t.Errorf("incorrect sum reported: %d", sum)
	}
}

func TestHeavyLoad(t *testing.T) {
	tests := []struct {
		timeSpan time.Duration
		inc      int
	}{
		{
			timeSpan: 30 * time.Minute,
			inc:      5,
		},
		{
			timeSpan: 30 * time.Minute,
			inc:      500,
		},
		{
			timeSpan: 120 * time.Minute,
			inc:      1000,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			now := time.Now()
			start := now.Add(-test.timeSpan)
			t.Logf("span: %v - %v", start, now)
			ts := NewMonitor()
			split := test.timeSpan.Seconds() / float64(test.inc)
			t.Logf("time split base: %v", time.Second*time.Duration(int(split)))

			for i := 0; i < test.inc; i++ {
				var timeInc time.Time
				if i == 0 {
					timeInc = start.Add(time.Second * 3)
				} else {
					timeInc = start.Add(time.Second * time.Duration(i*int(split)))
				}
				ts.Increment(1, timeInc)
			}
			minutes := test.timeSpan.Minutes() / float64(10)
			startSpan := time.Minute * time.Duration(minutes)
			start = start.Add(-startSpan)

			rangeSum := ts.RangeSum(start, time.Now())
			if rangeSum != test.inc {
				t.Errorf("RangeSum value %d, conflict with incremented %d", rangeSum, test.inc)
			}

			recentSum := ts.RecentSum(now.Sub(start))
			if recentSum != test.inc {
				t.Errorf("RecentSum %d did not return expected value: %d", recentSum, test.inc)
			}
		})
	}
}

func TestHeavyDenseIncrementing(t *testing.T) {
	tests := []struct {
		timeSpan time.Duration
		inc      int
	}{
		{
			timeSpan: 30 * time.Minute,
			inc:      5,
		},
		{
			timeSpan: 30 * time.Minute,
			inc:      500,
		},
		{
			timeSpan: 120 * time.Minute,
			inc:      1000,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			now := time.Now()
			start := now.Add(-test.timeSpan)
			t.Logf("span: %v - %v", start, now)
			ts := NewMonitor()
			expInc := int(test.timeSpan.Seconds()) * test.inc
			t.Logf("timespan: %v, incrementing: %d", test.timeSpan, expInc)

			for i := 0; i < int(test.timeSpan.Seconds()); i++ {
				var timeInc time.Time
				if i == 0 {
					timeInc = start.Add(time.Second * 1)
				} else {
					timeInc = start.Add(time.Second * time.Duration(i))
				}
				for j := 0; j < test.inc; j++ {
					ts.Increment(1, timeInc)
				}
			}
			minutes := test.timeSpan.Minutes() / float64(10)
			startSpan := time.Minute * time.Duration(minutes)
			start = start.Add(-startSpan)

			rangeSum := ts.RangeSum(start, now)
			if rangeSum != expInc {
				t.Errorf("RangeSum value %d, conflict with incremented %d", rangeSum, test.inc)
			}

			recentSum := ts.RecentSum(now.Sub(start))
			if recentSum != expInc {
				t.Errorf("RecentSum %d did not return expected value: %d", recentSum, test.inc)
			}
		})
	}
}

func TestHeavyDenseIncrementingNormalInit(t *testing.T) {
	tests := []struct {
		timeSpan time.Duration
		inc      int
	}{
		{
			timeSpan: 30 * time.Minute,
			inc:      5,
		},
		{
			timeSpan: 30 * time.Minute,
			inc:      500,
		},
		{
			timeSpan: 120 * time.Minute,
			inc:      1000,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			startTime := time.Now().Add(-test.timeSpan)
			tsNow := time.Now()
			t.Logf("span: %v - %v", startTime, tsNow)
			ts := NewMonitor()
			expInc := int(test.timeSpan.Seconds()) * test.inc
			t.Logf("timespan: %v, incrementing: %d", test.timeSpan, expInc)

			for i := 0; i < int(test.timeSpan.Seconds()); i++ {
				var timeInc time.Time
				if i == 0 {
					timeInc = startTime.Add(time.Second * 1)
				} else {
					timeInc = startTime.Add(time.Second * time.Duration(i))
				}
				for j := 0; j < test.inc; j++ {
					ts.Increment(1, timeInc)
				}
			}
			minutes := test.timeSpan.Minutes() / float64(10)
			startSpan := time.Minute * time.Duration(minutes)

			rangeSum := ts.RangeSum(startTime.Add(-startSpan), tsNow)
			if rangeSum != expInc {
				t.Errorf("RangeSum value %d, conflict with incremented %d", rangeSum, test.inc)
			}

			recentSum := ts.RecentSum(tsNow.Sub(startTime.Add(-startSpan)))
			if recentSum != expInc {
				t.Errorf("RecentSum %d did not return expected value: %d", recentSum, test.inc)
			}
		})
	}
}
