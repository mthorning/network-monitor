package utils

import "sync"

type Tracker[T any] struct {
	mu   sync.Mutex
	vals map[string]T
}

func NewTracker[T any]() *Tracker[T] {
	return &Tracker[T]{
		vals: make(map[string]T),
	}
}

func (t *Tracker[T]) Get(key string) T {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.vals[key]
}

func (t *Tracker[T]) GetAll() map[string]T {
	t.mu.Lock()
	defer t.mu.Unlock()
	copy := make(map[string]T)
	for k, v := range t.vals {
		copy[k] = v
	}
	return copy
}

func (t *Tracker[T]) Set(key string, val T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.vals[key] = val
}

func (t *Tracker[T]) SetAll(val T) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, _ := range t.vals {
		t.vals[k] = val
	}
}
