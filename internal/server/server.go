// Package server implements a multiplexer and handlers necessary for
// processing incoming requests.
package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"

	monitor "github.com/a-tho/monitor/internal"
)

type server struct {
	gauge   monitor.MetricRepo[monitor.Gauge]
	counter monitor.MetricRepo[monitor.Counter]
}

// NewServer creates a new multiplexer with configured handlers
func NewServer(
	gauge monitor.MetricRepo[monitor.Gauge],
	counter monitor.MetricRepo[monitor.Counter],
) *chi.Mux {
	srv := server{gauge: gauge, counter: counter}
	mux := chi.NewRouter()

	mux.Get("/", srv.GetAllHandler)

	path := fmt.Sprintf("/%s/{%s}/{%s}/{%s}", UpdPath, TypePath, NamePath, ValuePath)
	mux.Post(path, srv.UpdHandler)

	path = fmt.Sprintf("/%s/{%s}/{%s}", ValuePath, TypePath, NamePath)
	mux.Get(path, srv.GetValHandler)

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
