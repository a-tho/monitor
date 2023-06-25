// Package server implements a multiplexer and handlers necessary for
// processing incoming requests.
package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	monitor "github.com/a-tho/monitor/internal"
)

type server struct {
	metrics monitor.MetricRepo

	log zerolog.Logger
}

// NewServer creates a new multiplexer with configured handlers
func NewServer(
	metrics monitor.MetricRepo,
	log zerolog.Logger,
) *chi.Mux {
	srv := server{metrics: metrics, log: log}
	mux := chi.NewRouter()

	mux.Get("/", srv.WithLogging(srv.GetAllHandler))

	path := fmt.Sprintf("/%s/{%s}/{%s}/{%s}", UpdPath, TypePath, NamePath, ValuePath)
	mux.Post(path, srv.WithLogging(srv.UpdHandler))

	path = fmt.Sprintf("/%s/{%s}/{%s}", ValuePath, TypePath, NamePath)
	mux.Get(path, srv.WithLogging(srv.GetValHandler))

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
