// Package server implements a multiplexer and handlers necessary for
// processing incoming requests.
package server

import (
	"fmt"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/go-chi/chi/v5"
)

type server struct {
	gauge   monitor.MetricRepo[monitor.Gauge]
	counter monitor.MetricRepo[monitor.Counter]
}

// New creates a new multiplexer with configured handlers
func New(
	gauge monitor.MetricRepo[monitor.Gauge],
	counter monitor.MetricRepo[monitor.Counter],
) *chi.Mux {
	srv := server{gauge: gauge, counter: counter}
	mux := chi.NewRouter()

	path := fmt.Sprintf("/%s/{%s}/{%s}/{%s}", UpdPath, TypePath, NamePath, ValuePath)
	mux.Post(path, srv.UpdHandler)

	return mux
}

const (
	UpdPath = "update"

	GaugePath   = "gauge"
	CounterPath = "counter"

	TypePath  = "type"
	NamePath  = "name"
	ValuePath = "value"
)
