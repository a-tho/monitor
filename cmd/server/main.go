package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
	"github.com/a-tho/monitor/internal/storage"
)

type Config struct {
	// Flags
	SrvAddr   string `env:"ADDRESS"`
	LogLevel  string `env:"LOG_LEVEL"`
	LogFormat string `env:"LOG_FORMAT"`

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
	if err := cfg.parseConfig(); err != nil {
		return err
	}
	log := cfg.initLogger()

	cfg.metrics = storage.New()

	mux := server.NewServer(cfg.metrics, log)
	return http.ListenAndServe(cfg.SrvAddr, mux)
}

func (c *Config) parseConfig() error {
	flag.StringVar(&c.SrvAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.LogLevel, "log", "debug", "log level")
	flag.Parse()

	if err := env.Parse(c); err != nil {
		return err
	}

	return nil
}

func (c Config) initLogger() zerolog.Logger {
	level := zerolog.ErrorLevel
	if newLevel, err := zerolog.ParseLevel(c.LogLevel); err == nil {
		level = newLevel
	} else {
		println("nope")
	}
	out := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}
	logCtx := zerolog.New(out).Level(level).With().Timestamp().Stack().Caller()
	return logCtx.Logger()
}
