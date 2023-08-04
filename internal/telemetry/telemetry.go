// Package telemetry implements a trivial way to collect and transmit metrics.
package telemetry

import (
	"context"
	"encoding/base64"
	"time"

	monitor "github.com/a-tho/monitor/internal"
)

const (
	contentEncoding     = "Content-Encoding"
	contentType         = "Content-Type"
	encodingGzip        = "gzip"
	typeApplicationJSON = "application/json"
	bodySignature       = "HashSHA256"
)

type Observer struct {
	SrvAddr        string
	pollInterval   time.Duration
	reportStep     int
	reportInterval time.Duration
	signKey        []byte

	// local storage for the polled metrics that have not been reported yet
	polled []MetricInstance
}

// A MetricInstance holds a set of metrics collected roughly at the same moment
// in time.
type MetricInstance struct {
	Gauges map[string]monitor.Gauge
}

func NewObserver(srvAddr string, pollInterval, reportStep int, signKeyStr string) *Observer {
	signKey, err := base64.StdEncoding.DecodeString(signKeyStr)
	if err != nil {
		signKey = []byte{}
	}

	obs := Observer{
		SrvAddr:        srvAddr,
		pollInterval:   time.Duration(pollInterval) * time.Second,
		reportStep:     reportStep,
		reportInterval: time.Duration(pollInterval*reportStep) * time.Second,
		polled:         make([]MetricInstance, reportStep),
		signKey:        signKey,
	}
	for i := range obs.polled {
		obs.polled[i].Gauges = make(map[string]monitor.Gauge)
	}
	return &obs
}

func (o *Observer) Observe(ctx context.Context) error {
	pollCount := 0
	for {
		o.poll(pollCount)

		pollCount++
		if pollCount%o.reportStep == 0 {
			_ = o.prepare(ctx) // don't exit if failed to send metrics
		}

		timer := time.NewTimer(o.pollInterval)
		select {
		case <-timer.C:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
