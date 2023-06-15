package main

import (
	"flag"
	"net/http"

	"github.com/caarlos0/env"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
	"github.com/a-tho/monitor/internal/storage"
)

type Config struct {
	// Flags
	SrvAddr string `env:"ADDRESS"`

	// Storage
	metrics monitor.MetricRepo
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var cfg Config
	if err := parseConfig(&cfg); err != nil {
		return err
	}

	cfg.metrics = storage.New()

	mux := server.NewServer(cfg.metrics)
	return http.ListenAndServe(cfg.SrvAddr, mux)
}

func parseConfig(cfg *Config) error {
	flag.StringVar(&cfg.SrvAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return err
	}

	return nil
}
