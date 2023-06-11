// Package server implements a multiplexer and handlers necessary for
// processing incoming requests.
package server

import (
	"net/http"

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

	StrippedUpdHandler := http.StripPrefix(UpdPath, http.HandlerFunc(srv.UpdHandler))
	mux.Post(UpdPath+"/", StrippedUpdHandler.ServeHTTP)

	return mux
}

const (
	UpdPath = "/update"

	GaugePath   = "gauge"
	CounterPath = "counter"
)
