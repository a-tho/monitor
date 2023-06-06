// Package storage implements a trivial storage.
package storage

// MemStorage represents the storage.
type MemStorage[T float64 | int64] struct {
	data map[string]T
}

// New returns an initialized storage.
func New[T float64 | int64]() *MemStorage[T] {
	var stor MemStorage[T]
	stor.data = make(map[string]T)
	return &stor
}

// Set inserts or updates a value v for the key k.
func (s *MemStorage[T]) Set(k string, v T) {
	s.data[k] = v
}

// Add adds v to the value for the key k.
func (s *MemStorage[T]) Add(k string, v T) {
	s.data[k] += v
}
