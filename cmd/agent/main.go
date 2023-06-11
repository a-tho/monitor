package main

import (
	"time"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/telemetry"
)

func main() {
	srvAddr := "http://localhost:8080"
	pollInterval := 2 * time.Second
	reportStep := 5

	var obs monitor.Observer = telemetry.New(srvAddr, pollInterval, reportStep)
	obs.Observe()
}
