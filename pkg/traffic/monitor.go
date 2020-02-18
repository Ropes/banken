package traffic

import (
	"sync"
	"time"

	"github.com/ropes/banken/pkg/traffic/internal/timeseries"
)

type clock struct {
	t time.Time
}

func newclock(t time.Time) *clock {
	return &clock{t: t}
}

func (c *clock) Time() time.Time {
	return c.t
}

type nowClock struct{}

func (c *nowClock) Time() time.Time {
	return time.Now()
}

// Monitor aggregates http request counts into a searchable data
// structure.
type Monitor struct {
	tsdb   *timeseries.TimeSeries
	incMux sync.Mutex
}

// NewMonitor initializes the data type with clock and NewFloat observable
// for storing time series data points.
//
// The internal timeseries operation requires that Increment calls timestamp
// not surpass the timeseries's clock time. Primarily a testing concern.
func NewMonitor(t time.Time) *Monitor {
	//c := newclock(t)
	c := &nowClock{}
	return &Monitor{
		tsdb: timeseries.NewTimeSeriesWithClock(timeseries.NewFloat, c),
	}
}

// Increment the count by i for given clock time.
func (tm *Monitor) Increment(i int, clock time.Time) {
	f := new(timeseries.Float)
	*f = timeseries.Float(i)

	tm.incMux.Lock()
	tm.tsdb.AddWithTime(f, clock)
	tm.incMux.Unlock()
}

// RangeSum aggregates the occurrences within the delta duration parameter.
func (tm *Monitor) RangeSum(start, finish time.Time) int {
	obs := tm.tsdb.Range(start, finish)
	f := obs.(*timeseries.Float)
	return int(*f)
}

// RecentSum aggregates the occurrences within the delta duration parameter.
func (tm *Monitor) RecentSum(delta time.Duration) int {
	obs := tm.tsdb.Recent(delta)
	f := obs.(*timeseries.Float)
	return int(*f)
}
