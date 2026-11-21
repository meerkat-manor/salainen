package pathlock

import "sync"

type entry struct {
	mu sync.RWMutex
}

var locks sync.Map

func get(path string) *entry {
	value, _ := locks.LoadOrStore(path, &entry{})
	return value.(*entry)
}

func Lock(path string) {
	get(path).mu.Lock()
}

func Unlock(path string) {
	get(path).mu.Unlock()
}

func RLock(path string) {
	get(path).mu.RLock()
}

func RUnlock(path string) {
	get(path).mu.RUnlock()
}
