// Package storage implements a trivial storage.
package storage

import (
	"encoding/json"

	monitor "github.com/a-tho/monitor/internal"
)

// MemStorage represents the storage.
type MemStorage struct {
	dataGauge   map[string]monitor.Gauge
	dataCounter map[string]monitor.Counter
}

// New returns an initialized storage.
func New() *MemStorage {
	return &MemStorage{
		dataGauge:   make(map[string]monitor.Gauge),
		dataCounter: make(map[string]monitor.Counter),
	}
}

// SetGauge inserts or updates a gauge metric value v for the key k.
func (s *MemStorage) SetGauge(k string, v monitor.Gauge) monitor.MetricRepo {
	s.dataGauge[k] = v
	return s
}

// AddCounter add a counter metric value v for the key k.
func (s *MemStorage) AddCounter(k string, v monitor.Counter) monitor.MetricRepo {
	s.dataCounter[k] += v
	return s
}

// GetGauge retrieves the gauge value for the key k.
func (s *MemStorage) GetGauge(k string) (v monitor.Gauge, ok bool) {
	v, ok = s.dataGauge[k]
	return
}

// GetCounter retrieves the counter value for the key k.
func (s *MemStorage) GetCounter(k string) (v monitor.Counter, ok bool) {
	v, ok = s.dataCounter[k]
	return
}

// StringGauge produces a JSON representation of gauge metrics kept in the
// storage
func (s *MemStorage) StringGauge() string {
	out, _ := json.Marshal(s.dataGauge)
	return string(out)
}

// StringCounter produces a JSON representation of counter metrics kept in the
// storage
func (s *MemStorage) StringCounter() string {
	out, _ := json.Marshal(s.dataCounter)
	return string(out)
}

// StringCounter exposes the substorage with gauge metrics
func (s *MemStorage) GetAllGauge() map[string]monitor.Gauge {
	return s.dataGauge
}

// StringCounter exposes the substorage with counter metrics
func (s *MemStorage) GetAllCounter() map[string]monitor.Counter {
	return s.dataCounter
}
