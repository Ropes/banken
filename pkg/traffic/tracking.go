package traffic

import "sync"

type keyCounter struct {
	count uint64
	mux   sync.Mutex
}

func (k *keyCounter) Add(inc uint64) {
	k.mux.Lock()
	k.count += inc
	k.mux.Unlock()
}

func (k *keyCounter) Get() uint64 {
	return k.count
}

// RequestCounter provides safe concurrent counting
// of URLs requests made.
type RequestCounter struct {
	reqs sync.Map
}
