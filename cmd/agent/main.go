package main

import (
	"errors"
	"flag"
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
	if err := obs.Observe(); err != nil {
		return err
	}

	return nil
}

func parseFlags() error {
	flag.StringVar(&srvAddr, "a", "localhost:8080", "address and port to run server")
	pi := flag.Int("p", 2, "rate of polling metrics in seconds")
	ri := flag.Int("r", 10, "rate of reporting metrics in seconds")
	flag.Parse()

	// Both poll/report intervals must be positive, report interval has to be
	// greater than and a multiple of poll interval
	if *pi <= 0 || *ri <= 0 || *ri < *pi || *ri%*pi != 0 {
		return errors.New("invalid p or r")
	}
	pollInterval = time.Duration(*pi) * time.Second
	reportStep = *ri / *pi

	return nil
}
