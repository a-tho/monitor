// Package storage implements a trivial storage.
package storage

import (
	monitor "github.com/a-tho/monitor/internal"
)

// MemStorage represents the storage.
type MemStorage[T monitor.Gauge | monitor.Counter] struct {
	data map[string]T
}

// New returns an initialized storage.
func New[T monitor.Gauge | monitor.Counter]() *MemStorage[T] {
	return &MemStorage[T]{
		data: make(map[string]T),
	}
}

// Set inserts or updates a value v for the key k.
func (s *MemStorage[T]) Set(k string, v T) {
	s.data[k] = v
}

// Add adds v to the value for the key k.
func (s *MemStorage[T]) Add(k string, v T) {
	s.data[k] += v
}
