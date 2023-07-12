package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
	"github.com/a-tho/monitor/internal/storage"
)

type Config struct {
	// Flags
	SrvAddr         string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	LogFormat       string `env:"LOG_FORMAT"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`

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
	cfg.initLogger()

	cfg.metrics = storage.New(cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
	defer cfg.metrics.Close()

	mux := server.NewServer(cfg.metrics)
	go func() {
		if err := http.ListenAndServe(cfg.SrvAddr, mux); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGQUIT)

	<-quit

	return nil
}

func (c *Config) parseConfig() error {
	flag.StringVar(&c.SrvAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.LogLevel, "log", "debug", "log level")
	flag.IntVar(&c.StoreInterval, "i", 300, "interval in seconds after which readings saved to disk")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "file where to save current values")
	flag.BoolVar(&c.Restore, "r", true, "whether or not to load previously saved values on server start")
	flag.Parse()

	if err := env.Parse(c); err != nil {
		return err
	}

	return nil
}

func (c Config) initLogger() {
	level := zerolog.ErrorLevel
	if newLevel, err := zerolog.ParseLevel(c.LogLevel); err == nil {
		level = newLevel
	}
	out := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}
	log.Logger = zerolog.New(out).Level(level).With().Timestamp().Stack().Caller().Logger()
}
