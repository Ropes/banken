package traffic

import "sync"

type keyCounter struct {
	count uint64
	mux   sync.RWMutex
}

func (k *keyCounter) Add(inc uint64) {
	k.mux.Lock()
	k.count += inc
	k.mux.Unlock()
}

func (k *keyCounter) Get() uint64 {
	k.mux.RLock()
	defer k.mux.RUnlock()
	return k.count
}

// RequestCounter provides safe concurrent counting
// of URLs requests made.
type RequestCounter struct {
	reqs sync.Map
}

// IncKey safely increments a key's count. Adds a new keyCounter to the map,
// iff it does not exist.
func (r *RequestCounter) IncKey(key string, i uint64) {
	tmp := new(keyCounter)
	kc, _ := r.reqs.LoadOrStore(key, tmp)
	if k, ok := kc.(*keyCounter); ok {
		k.Add(i)
	}
}

// Export provides a standard map of collected key values.
func (r *RequestCounter) Export() map[string]uint64 {
	output := make(map[string]uint64)
	r.reqs.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if val, ok := value.(*keyCounter); ok {
				output[keyStr] = val.Get()
				return true
			}
		}
		return false
	})
	return output
}
