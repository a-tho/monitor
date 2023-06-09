// Package telemetry implements a trivial way to collect and transmit metrics.
package telemetry

import (
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
)

type Observer struct {
	pollInterval   time.Duration
	reportStep     int
	reportInterval time.Duration

	// local storage for the polled metrics that have not been reported yet
	polled []monitor.MetricInstance
}

func New(pollInterval time.Duration, reportStep int) *Observer {
	obs := Observer{
		pollInterval:   pollInterval,
		reportStep:     reportStep,
		reportInterval: pollInterval * time.Duration(reportStep),
		polled:         make([]monitor.MetricInstance, 0, reportStep),
	}
	for i := range obs.polled {
		obs.polled[i].Gauges = make(map[string]monitor.Gauge)
	}
	return &obs
}

func (o *Observer) Observe() {
	pollCount := 0
	for {
		// Poll and report metrics
		var (
			t        = pollCount % o.reportStep
			memStats runtime.MemStats
		)
		runtime.ReadMemStats(&memStats)
		o.polled[t].Gauges["Alloc"] = monitor.Gauge(memStats.Alloc)
		o.polled[t].Gauges["BuckHashSys"] = monitor.Gauge(memStats.BuckHashSys)
		o.polled[t].Gauges["Frees"] = monitor.Gauge(memStats.Frees)
		o.polled[t].Gauges["GCCPUFraction"] = monitor.Gauge(memStats.GCCPUFraction)
		o.polled[t].Gauges["GCSys"] = monitor.Gauge(memStats.GCSys)
		o.polled[t].Gauges["HeapAlloc"] = monitor.Gauge(memStats.HeapAlloc)
		o.polled[t].Gauges["HeapIdle"] = monitor.Gauge(memStats.HeapIdle)
		o.polled[t].Gauges["HeapInuse"] = monitor.Gauge(memStats.HeapInuse)
		o.polled[t].Gauges["HeapObjects"] = monitor.Gauge(memStats.HeapObjects)
		o.polled[t].Gauges["HeapReleased"] = monitor.Gauge(memStats.HeapReleased)
		o.polled[t].Gauges["HeapSys"] = monitor.Gauge(memStats.HeapSys)
		o.polled[t].Gauges["LastGC"] = monitor.Gauge(memStats.LastGC)
		o.polled[t].Gauges["Lookups"] = monitor.Gauge(memStats.Lookups)
		o.polled[t].Gauges["MCacheInuse"] = monitor.Gauge(memStats.MCacheInuse)
		o.polled[t].Gauges["MCacheSys"] = monitor.Gauge(memStats.MCacheSys)
		o.polled[t].Gauges["MSpanInuse"] = monitor.Gauge(memStats.MSpanInuse)
		o.polled[t].Gauges["MSpanSys"] = monitor.Gauge(memStats.MSpanSys)
		o.polled[t].Gauges["Mallocs"] = monitor.Gauge(memStats.Mallocs)
		o.polled[t].Gauges["NextGC"] = monitor.Gauge(memStats.NextGC)
		o.polled[t].Gauges["NumForcedGC"] = monitor.Gauge(memStats.NumForcedGC)
		o.polled[t].Gauges["NumGC"] = monitor.Gauge(memStats.NumGC)
		o.polled[t].Gauges["OtherSys"] = monitor.Gauge(memStats.OtherSys)
		o.polled[t].Gauges["PauseTotalNs"] = monitor.Gauge(memStats.PauseTotalNs)
		o.polled[t].Gauges["StackInuse"] = monitor.Gauge(memStats.StackInuse)
		o.polled[t].Gauges["StackSys"] = monitor.Gauge(memStats.StackSys)
		o.polled[t].Gauges["Sys"] = monitor.Gauge(memStats.Sys)
		o.polled[t].Gauges["TotalAlloc"] = monitor.Gauge(memStats.TotalAlloc)

		randomValue := rand.New(rand.NewSource(time.Now().Unix())).Float64()
		o.polled[t].Gauges["RandomValue"] = monitor.Gauge(randomValue)

		pollCount++
		if pollCount%o.reportStep == 0 {
			// Report to the server
			for _, instance := range o.polled {
				for key, value := range instance.Gauges {
					url := "http://localhost:8080" + server.PathPrefix + "/" +
						server.GaugePath + "/" + key + "/" + strconv.Itoa(int(value))
					resp, err := http.Post(url, "text-plain", nil)
					if err != nil {
						continue
					}
					resp.Body.Close()

				}

			}
		}

		time.Sleep(o.pollInterval)
	}
}
