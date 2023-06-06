// Package server implements the multiplexer and the handlers necessary for
// processing incoming requests.
package server

import (
	"net/http"

	monitor "github.com/a-tho/monitor/internal"
)

type server struct {
	gauge   monitor.Metric[float64]
	counter monitor.Metric[int64]
}

// New creates a new multiplexer with configured handlers
func New(
	gauge monitor.Metric[float64],
	counter monitor.Metric[int64],
) *http.ServeMux {
	srv := server{gauge: gauge, counter: counter}
	mux := http.NewServeMux()
	mux.Handle(pathPrefix, http.StripPrefix(pathPrefix, http.HandlerFunc(srv.updateHandler)))
	return mux
}

const (
	pathPrefix  = "/update/"
	gaugePath   = "gauge"
	counterPath = "counter"
)
