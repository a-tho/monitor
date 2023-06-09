// Package server implements a multiplexer and handlers necessary for
// processing incoming requests.
package server

import (
	"net/http"

	monitor "github.com/a-tho/monitor/internal"
)

type server struct {
	gauge   monitor.MetricRepo[monitor.Gauge]
	counter monitor.MetricRepo[monitor.Counter]
}

// New creates a new multiplexer with configured handlers
func New(
	gauge monitor.MetricRepo[monitor.Gauge],
	counter monitor.MetricRepo[monitor.Counter],
) *http.ServeMux {
	srv := server{gauge: gauge, counter: counter}
	mux := http.NewServeMux()
	mux.Handle(PathPrefix+"/", http.StripPrefix(PathPrefix, http.HandlerFunc(srv.UpdateHandler)))
	return mux
}

const (
	PathPrefix  = "/update"
	GaugePath   = "gauge"
	CounterPath = "counter"
)
