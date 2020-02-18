package traffic

import (
	"fmt"
	"sync"
	"testing"
)

func TestZeroValues(t *testing.T) {

	tests := []struct {
		i       int
		workers int
	}{
		{
			i:       100,
			workers: 5,
		},
		{
			i:       100,
			workers: 50,
		},
		{
			i:       10000,
			workers: 5,
		},
		{
			i:       10000,
			workers: 50,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("add-%d-with-%d", test.i, test.workers), func(t *testing.T) {
			kc := keyCounter{}

			work := make(chan uint64, 1)
			var wg sync.WaitGroup
			for i := 0; i < test.workers; i++ {
				wg.Add(1)
				go func(work chan uint64) {
					for x := range work {
						if x == uint64(0) { // exit if zero value
							break
						}
						kc.Add(x)
					}
					wg.Done()
				}(work)
			}
			for i := 0; i < test.i; i++ {
				work <- uint64(1)
				//fmt.Println("inserted:", i)
			}
			close(work)
			wg.Wait()

			if kc.Get() != uint64(test.i) {
				t.Errorf("Add() operation count had unexpected[%d] value: %d", test.i, kc.Get())
			}
			t.Logf("Expected %d, counted: %d", test.i, kc.Get())
		})
	}
}

func TestConcurrentKeyMap(t *testing.T) {
	keys := []string{"hihi", "inu", "おはよう", "felt"}
	tests := []struct {
		i       int
		workers int
	}{
		{
			i:       5,
			workers: 5,
		},
		{
			i:       50000,
			workers: 5,
		},
		{
			i:       5,
			workers: 50,
		},
		{
			i:       50000,
			workers: 500,
		},
	}
	type tup struct {
		key string
		cnt uint64
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rc := new(RequestCounter)
			work := make(chan tup, 1)
			var wg sync.WaitGroup
			for i := 0; i < test.workers; i++ {
				wg.Add(1)
				go func(work chan tup) {
					for x := range work {
						rc.IncKey(x.key, x.cnt)
					}
					wg.Done()
				}(work)
			}

			for i := 0; i < test.i; i++ {
				for _, k := range keys {
					work <- tup{
						cnt: uint64(1),
						key: k,
					}
				}
			}
			close(work)
			wg.Wait()

			output := rc.Export()
			for _, k := range keys {
				v, ok := output[k]
				if !ok {
					t.Errorf("key %q not found in exported map", k)
				} else {
					if int(v) != test.i {
						t.Errorf("key %q: value %d != %d", k, v, test.i)
					}
				}

			}
		})
	}

}
