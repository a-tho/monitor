package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
	"github.com/a-tho/monitor/internal/storage"
)

type Config struct {
	// Flags
	SrvAddr         string        `env:"ADDRESS"`
	LogLevel        string        `env:"LOG_LEVEL"`
	LogFormat       string        `env:"LOG_FORMAT"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	Restore         bool          `env:"RESTORE"`

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

	log.Info().Str("SrvAddr", cfg.SrvAddr).Msg("")
	log.Info().Dur("StoreInterval", cfg.StoreInterval).Msg("")
	log.Info().Str("FileStoragePath", cfg.FileStoragePath).Msg("")
	log.Info().Bool("Restore", cfg.Restore).Msg("")

	cfg.metrics = storage.New(cfg.FileStoragePath, cfg.StoreInterval == 0, cfg.Restore)
	defer cfg.metrics.Close()

	mux := server.NewServer(cfg.metrics, log)
	go func() {
		if err := http.ListenAndServe(cfg.SrvAddr, mux); err != nil {
			panic(err)
		}
	}()

	// Write to the file every StoreInterval seconds
	var ticker <-chan time.Time
	if cfg.StoreInterval > 0 {
		t := time.NewTicker(cfg.StoreInterval)
		defer t.Stop()
		ticker = t.C
	}

	// and close the file when SIGINT is passed
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	for {
		select {
		case <-ticker:
			cfg.metrics.WriteToFile()
		case <-quit:
			return cfg.metrics.WriteToFile()
		}
	}
}

func (c *Config) parseConfig() error {
	flag.StringVar(&c.SrvAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.LogLevel, "log", "debug", "log level")
	flag.DurationVar(&c.StoreInterval, "i", 300*time.Second, "interval in seconds after which readings saved to disk")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "file where to save current values")
	flag.BoolVar(&c.Restore, "r", true, "whether or not to load previously saved values on server start")
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
	}
	out := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}
	logCtx := zerolog.New(out).Level(level).With().Timestamp().Stack().Caller()
	return logCtx.Logger()
}
