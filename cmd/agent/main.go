package main

import (
	"time"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/telemetry"
)

var (
	// Flags
	srvAddr      string
	pollInterval time.Duration
	reportStep   int
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := parseFlags(); err != nil {
		return err
	}

	var obs monitor.Observer = telemetry.New(srvAddr, pollInterval, reportStep)
	obs.Observe()

	return nil
}
