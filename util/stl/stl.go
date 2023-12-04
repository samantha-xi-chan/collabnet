package stl

import (
	"sync"
)

type SafeMap struct {
	mu   sync.Mutex
	Data map[string]int64
}

func (sm *SafeMap) Set(key string, value int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Data[key] = value
}

func (sm *SafeMap) Get(key string) (int64, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	value, ok := sm.Data[key]
	return value, ok
}

func (sm *SafeMap) Iterate() map[string]int64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	result := make(map[string]int64)
	for key, value := range sm.Data {
		result[key] = value
	}

	return result
}

func (sm *SafeMap) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.Data, key)
}
