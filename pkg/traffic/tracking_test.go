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
			fmt.Println("work submitted to pool")
			wg.Wait()

			if kc.Get() != uint64(test.i) {
				t.Errorf("Add() operation count had unexpected[%d] value: %d", test.i, kc.Get())
			}
			t.Logf("Expected %d, counted: %d", test.i, kc.Get())
		})
	}
}
