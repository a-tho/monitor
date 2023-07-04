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

	mux.Get("/", srv.WithLogging(srv.All))

	path := fmt.Sprintf("/%s/{%s}/{%s}/{%s}", UpdPath, TypePath, NamePath, ValuePath)
	mux.Post(path, srv.WithLogging(srv.UpdateLegacy))

	path = fmt.Sprintf("/%s/", UpdPath)
	mux.Post(path, srv.WithLogging(srv.Update))

	path = fmt.Sprintf("/%s/{%s}/{%s}", ValuePath, TypePath, NamePath)
	mux.Get(path, srv.WithLogging(srv.ValueLegacy))

	path = fmt.Sprintf("/%s/", ValuePath)
	mux.Post(path, srv.WithLogging(srv.Value))

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
