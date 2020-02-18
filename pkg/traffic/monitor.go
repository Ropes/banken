package traffic

import (
	"sync"
	"time"

	"github.com/ropes/banken/pkg/traffic/internal/timeseries"
)

type staticClock struct {
	t time.Time
}

func newStaticClock(t time.Time) *staticClock {
	return &staticClock{t: t}
}

func (c *staticClock) Time() time.Time {
	return c.t
}

type nowClock struct{}

func (c *nowClock) Time() time.Time {
	return time.Now()
}

// Monitor aggregates http request counts into a searchable data structure.
// timeseries.TimeSeries data structure is not a concurrency-safe package,
//so all calls to it are wrapped in a mutex.
type Monitor struct {
	tsdb  *timeseries.TimeSeries
	tsMux sync.RWMutex
}

// NewMonitor initializes the data type with clock and NewFloat observable
// for storing time series data points.
//
// The internal timeseries operation requires that Increment calls timestamp
// not surpass the timeseries's clock time. Primarily a testing concern.
func NewMonitor(t time.Time) *Monitor {
	c := &nowClock{}
	return &Monitor{
		tsdb: timeseries.NewTimeSeriesWithClock(timeseries.NewFloat, c),
	}
}

func newTestMonitor(t time.Time) *Monitor {
	c := newStaticClock(t)
	return &Monitor{
		tsdb: timeseries.NewTimeSeriesWithClock(timeseries.NewFloat, c),
	}
}

// Increment the count by i for given clock time.
func (tm *Monitor) Increment(i int, clock time.Time) {
	f := new(timeseries.Float)
	*f = timeseries.Float(i)

	tm.tsMux.Lock()
	tm.tsdb.AddWithTime(f, clock)
	tm.tsMux.Unlock()
}

// RangeSum aggregates the occurrences within the delta duration parameter.
func (tm *Monitor) RangeSum(start, finish time.Time) int {
	tm.tsMux.RLock()
	obs := tm.tsdb.Range(start, finish)
	tm.tsMux.RUnlock()
	f := obs.(*timeseries.Float)
	return int(*f)
}

// RecentSum aggregates the occurrences within the delta duration parameter.
func (tm *Monitor) RecentSum(delta time.Duration) int {
	tm.tsMux.RLock()
	obs := tm.tsdb.Recent(delta)
	tm.tsMux.RUnlock()
	f := obs.(*timeseries.Float)
	return int(*f)
}
