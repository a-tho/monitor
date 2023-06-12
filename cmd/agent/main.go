package main

import (
	"errors"
	"flag"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/telemetry"
	"github.com/caarlos0/env"
)

type Config struct {
	SrvAddr string `env:"ADDRESS"`
	Poll    int    `env:"POLL_INTERVAL"`
	Report  int    `env:"REPORT_INTERVAL"`
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

	var obs monitor.Observer = telemetry.New(cfg.SrvAddr, cfg.Poll, cfg.Report/cfg.Poll)
	if err := obs.Observe(); err != nil {
		return err
	}

	return nil
}

func parseConfig(cfg *Config) error {
	flag.StringVar(&cfg.SrvAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.Poll, "p", 2, "rate of polling metrics in seconds")
	flag.IntVar(&cfg.Report, "r", 10, "rate of reporting metrics in seconds")
	flag.Parse()

	// Both poll/report intervals must be positive, report interval has to be
	// greater than and a multiple of poll interval
	if cfg.Poll <= 0 || cfg.Report <= 0 ||
		cfg.Report < cfg.Poll || cfg.Report%cfg.Poll != 0 {
		return errors.New("invalid p or r")
	}

	if err := env.Parse(cfg); err != nil {
		return err
	}

	return nil
}
