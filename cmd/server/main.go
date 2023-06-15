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
	gauge   monitor.MetricRepo[monitor.Gauge]
	counter monitor.MetricRepo[monitor.Counter]
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

	cfg.gauge = storage.New[monitor.Gauge]()
	cfg.counter = storage.New[monitor.Counter]()

	mux := server.New(cfg.gauge, cfg.counter)
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
