package main

import (
	"time"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/telemetry"
)

func main() {
	pollInterval := 2 * time.Second
	reportStep := 5

	var obs monitor.Observer = telemetry.New(pollInterval, reportStep)
	obs.Observe()
}
