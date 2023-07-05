// Package storage implements a trivial storage.
package storage

import (
	"encoding/json"
	"os"

	monitor "github.com/a-tho/monitor/internal"
)

// MemStorage represents the storage.
type MemStorage struct {
	DataGauge   map[string]monitor.Gauge
	DataCounter map[string]monitor.Counter

	file *os.File
	// Whether recording is synchronuous
	syncMode bool
}

// New returns an initialized storage.
func New(fileStoragePath string, syncMode bool, restore bool) *MemStorage {
	storage := MemStorage{
		DataGauge:   make(map[string]monitor.Gauge),
		DataCounter: make(map[string]monitor.Counter),
	}

	if fileStoragePath != "" {
		file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return &storage
		}

		if restore {
			dec := json.NewDecoder(file)
			var storageIn MemStorage
			if err = dec.Decode(&storageIn); err == nil {
				storage = storageIn
			}
		}

		storage.file = file
		storage.syncMode = syncMode
	}

	return &storage
}

// SetGauge inserts or updates a gauge metric value v for the key k.
func (s *MemStorage) SetGauge(k string, v monitor.Gauge) monitor.MetricRepo {
	s.DataGauge[k] = v

	if s.syncMode {
		s.WriteToFile()
	}

	return s
}

// AddCounter add a counter metric value v for the key k.
func (s *MemStorage) AddCounter(k string, v monitor.Counter) monitor.MetricRepo {
	s.DataCounter[k] += v

	if s.syncMode {
		s.WriteToFile()
	}

	return s
}

// GetGauge retrieves the gauge value for the key k.
func (s *MemStorage) GetGauge(k string) (v monitor.Gauge, ok bool) {
	v, ok = s.DataGauge[k]
	return
}

// GetCounter retrieves the counter value for the key k.
func (s *MemStorage) GetCounter(k string) (v monitor.Counter, ok bool) {
	v, ok = s.DataCounter[k]
	return
}

// StringGauge produces a JSON representation of gauge metrics kept in the
// storage
func (s *MemStorage) StringGauge() string {
	out, _ := json.Marshal(s.DataGauge)
	return string(out)
}

// StringCounter produces a JSON representation of counter metrics kept in the
// storage
func (s *MemStorage) StringCounter() string {
	out, _ := json.Marshal(s.DataCounter)
	return string(out)
}

// StringCounter exposes the substorage with gauge metrics
func (s *MemStorage) GetAllGauge() map[string]monitor.Gauge {
	return s.DataGauge
}

// StringCounter exposes the substorage with counter metrics
func (s *MemStorage) GetAllCounter() map[string]monitor.Counter {
	return s.DataCounter
}

// // Marshal returns the JSON encoding of MemStorage.
// func (s MemStorage) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(s)
// }

// // Unmarshal parses the JSON-encoded data and stores the result
// // in MemStorage.
// func (s *MemStorage) UnmarshalJSON(data []byte) error {
// 	return json.Unmarshal(data, s)
// }

func (s *MemStorage) Close() error {
	if s.file == nil {
		return nil
	}
	return s.file.Close()
}

func (s *MemStorage) WriteToFile() error {
	if s.file == nil {
		return nil
	}
	s.file.Truncate(0)
	enc := json.NewEncoder(s.file)
	return enc.Encode(s)
}
