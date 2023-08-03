// Package server implements a multiplexer and handlers necessary for
// processing incoming requests.
package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"

	monitor "github.com/a-tho/monitor/internal"
	mw "github.com/a-tho/monitor/internal/middleware"
)

type server struct {
	metrics monitor.MetricRepo
}

// NewServer creates a new multiplexer with configured handlers
func NewServer(
	metrics monitor.MetricRepo,
	key string,
) *chi.Mux {
	srv := server{metrics: metrics}
	mux := chi.NewRouter()

	mux.Get("/", mw.WithSigning(mw.WithLogging(mw.WithCompressing(srv.All)), key))

	path := fmt.Sprintf("/%s/{%s}/{%s}/{%s}", UpdPath, TypePath, NamePath, ValuePath)
	mux.Post(path, mw.WithLogging(srv.UpdateLegacy))

	path = fmt.Sprintf("/%s/", UpdPath)
	mux.Post(path, mw.WithSigning(mw.WithLogging(mw.WithCompressing(srv.Update)), key))

	path = fmt.Sprintf("/%s/", UpdsPath)
	mux.Post(path, mw.WithSigning(mw.WithLogging(mw.WithCompressing(srv.Updates)), key))

	path = fmt.Sprintf("/%s/{%s}/{%s}", ValuePath, TypePath, NamePath)
	mux.Get(path, mw.WithLogging(srv.ValueLegacy))

	path = fmt.Sprintf("/%s/", ValuePath)
	mux.Post(path, mw.WithSigning(mw.WithLogging(mw.WithCompressing(srv.Value)), key))

	path = "/ping"
	mux.Get(path, mw.WithSigning(mw.WithLogging((mw.WithCompressing(srv.Ping))), key))

	return mux
}

const (
	UpdPath  = "update"
	UpdsPath = "updates"

	GaugePath   = "gauge"
	CounterPath = "counter"

	TypePath  = "type"
	NamePath  = "name"
	ValuePath = "value"
)
