// Package server implements a multiplexer and handlers necessary for
// processing incoming requests.
package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"

	monitor "github.com/a-tho/monitor/internal"
)

type server struct {
	metrics monitor.MetricRepo
}

// NewServer creates a new multiplexer with configured handlers
func NewServer(
	metrics monitor.MetricRepo,
) *chi.Mux {
	srv := server{metrics: metrics}
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
