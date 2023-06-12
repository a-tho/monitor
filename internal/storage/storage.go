// Package storage implements a trivial storage.
package storage

import (
	"bytes"
	"encoding/json"
	"html/template"

	monitor "github.com/a-tho/monitor/internal"
)

const (
	htmlMetricsTemplate = `
    {{range $key, $value := .}}
        <p>{{$key}}: {{$value}}</p>
    {{end}}`
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
func (s *MemStorage[T]) Set(k string, v T) monitor.MetricRepo[T] {
	s.data[k] = v
	return s
}

// Add adds v to the value for the key k.
func (s *MemStorage[T]) Add(k string, v T) monitor.MetricRepo[T] {
	s.data[k] += v
	return s
}

// Get retrieves the value for the key k.
func (s *MemStorage[T]) Get(k string) (v T, ok bool) {
	v, ok = s.data[k]
	return
}

func (s *MemStorage[T]) String() string {
	out, _ := json.Marshal(s.data)
	return string(out)
}

func (s MemStorage[T]) HTML() (*bytes.Buffer, error) {
	tmpl, err := template.New("metrics").Parse(htmlMetricsTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, s.data); err != nil {
		return nil, err
	}

	return &buf, nil
}
