package config

import (
	"flag"
	"os"
	"time"

	monitor "github.com/a-tho/monitor/internal"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	Metrics     monitor.MetricRepo
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func (c *Config) ParseConfig() error {
	flag.StringVar(&c.SrvAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.LogLevel, "log", "debug", "log level")
	flag.IntVar(&c.StoreInterval, "i", 300, "interval in seconds after which readings saved to disk")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "file where to save current values")
	flag.BoolVar(&c.Restore, "r", true, "whether or not to load previously saved values on server start")
	flag.StringVar(&c.DatabaseDSN, "d", "postgres://postgres:123456@localhost:5432/database", "database dsn")
	flag.Parse()

	if err := env.Parse(c); err != nil {
		return err
	}

	return nil
}

func (c Config) InitLogger() {
	level := zerolog.ErrorLevel
	if newLevel, err := zerolog.ParseLevel(c.LogLevel); err == nil {
		level = newLevel
	}
	out := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMicro}
	log.Logger = zerolog.New(out).Level(level).With().Timestamp().Stack().Caller().Logger()
}
