// Package telemetry implements a trivial way to collect and transmit metrics.
package telemetry

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
)

type Observer struct {
	SrvAddr        string
	pollInterval   time.Duration
	reportStep     int
	reportInterval time.Duration

	// local storage for the polled metrics that have not been reported yet
	polled []MetricInstance
}

// A MetricInstance holds a set of metrics collected roughly at the same moment
// in time.
type MetricInstance struct {
	Gauges map[string]monitor.Gauge
}

func NewObserver(srvAddr string, pollInterval, reportStep int) *Observer {
	obs := Observer{
		SrvAddr:        srvAddr,
		pollInterval:   time.Duration(pollInterval) * time.Second,
		reportStep:     reportStep,
		reportInterval: time.Duration(pollInterval*reportStep) * time.Second,
		polled:         make([]MetricInstance, reportStep),
	}
	for i := range obs.polled {
		obs.polled[i].Gauges = make(map[string]monitor.Gauge)
	}
	return &obs
}

func (o *Observer) Observe() error {
	pollCount := 0
	for {
		o.poll(pollCount)

		pollCount++
		if pollCount%o.reportStep == 0 {
			_ = o.report() // don't exit if failed to send metrics
		}

		time.Sleep(o.pollInterval)
	}
}

func (o *Observer) poll(pollCount int) {
	var (
		countSinceReport = pollCount % o.reportStep // the number of polls since the last report
		memStats         runtime.MemStats
	)
	runtime.ReadMemStats(&memStats)
	o.polled[countSinceReport].Gauges["Alloc"] = monitor.Gauge(memStats.Alloc)
	o.polled[countSinceReport].Gauges["BuckHashSys"] = monitor.Gauge(memStats.BuckHashSys)
	o.polled[countSinceReport].Gauges["Frees"] = monitor.Gauge(memStats.Frees)
	o.polled[countSinceReport].Gauges["GCCPUFraction"] = monitor.Gauge(memStats.GCCPUFraction)
	o.polled[countSinceReport].Gauges["GCSys"] = monitor.Gauge(memStats.GCSys)
	o.polled[countSinceReport].Gauges["HeapAlloc"] = monitor.Gauge(memStats.HeapAlloc)
	o.polled[countSinceReport].Gauges["HeapIdle"] = monitor.Gauge(memStats.HeapIdle)
	o.polled[countSinceReport].Gauges["HeapInuse"] = monitor.Gauge(memStats.HeapInuse)
	o.polled[countSinceReport].Gauges["HeapObjects"] = monitor.Gauge(memStats.HeapObjects)
	o.polled[countSinceReport].Gauges["HeapReleased"] = monitor.Gauge(memStats.HeapReleased)
	o.polled[countSinceReport].Gauges["HeapSys"] = monitor.Gauge(memStats.HeapSys)
	o.polled[countSinceReport].Gauges["LastGC"] = monitor.Gauge(memStats.LastGC)
	o.polled[countSinceReport].Gauges["Lookups"] = monitor.Gauge(memStats.Lookups)
	o.polled[countSinceReport].Gauges["MCacheInuse"] = monitor.Gauge(memStats.MCacheInuse)
	o.polled[countSinceReport].Gauges["MCacheSys"] = monitor.Gauge(memStats.MCacheSys)
	o.polled[countSinceReport].Gauges["MSpanInuse"] = monitor.Gauge(memStats.MSpanInuse)
	o.polled[countSinceReport].Gauges["MSpanSys"] = monitor.Gauge(memStats.MSpanSys)
	o.polled[countSinceReport].Gauges["Mallocs"] = monitor.Gauge(memStats.Mallocs)
	o.polled[countSinceReport].Gauges["NextGC"] = monitor.Gauge(memStats.NextGC)
	o.polled[countSinceReport].Gauges["NumForcedGC"] = monitor.Gauge(memStats.NumForcedGC)
	o.polled[countSinceReport].Gauges["NumGC"] = monitor.Gauge(memStats.NumGC)
	o.polled[countSinceReport].Gauges["OtherSys"] = monitor.Gauge(memStats.OtherSys)
	o.polled[countSinceReport].Gauges["PauseTotalNs"] = monitor.Gauge(memStats.PauseTotalNs)
	o.polled[countSinceReport].Gauges["StackInuse"] = monitor.Gauge(memStats.StackInuse)
	o.polled[countSinceReport].Gauges["StackSys"] = monitor.Gauge(memStats.StackSys)
	o.polled[countSinceReport].Gauges["Sys"] = monitor.Gauge(memStats.Sys)
	o.polled[countSinceReport].Gauges["TotalAlloc"] = monitor.Gauge(memStats.TotalAlloc)

	randomValue := rand.New(rand.NewSource(time.Now().Unix())).Float64()
	o.polled[countSinceReport].Gauges["RandomValue"] = monitor.Gauge(randomValue)
}

func (o *Observer) report() error {
	for _, instance := range o.polled {
		// Gauge metrics
		for key, val := range instance.Gauges {
			valFloat := float64(val)
			metric := monitor.Metrics{
				ID:    key,
				MType: server.GaugePath,
				Value: &valFloat,
			}
			if err := o.update(metric); err != nil {
				return err
			}
		}
	}

	// Counter metric
	delta := int64(o.reportStep)
	metric := monitor.Metrics{
		ID:    "PollCount",
		MType: server.CounterPath,
		Delta: &delta,
	}
	if err := o.update(metric); err != nil {
		return err
	}
	return nil
}

func (o *Observer) update(metric monitor.Metrics) error {
	// Prepare request arguments
	url := fmt.Sprintf("http://%s/%s/", o.SrvAddr, server.UpdPath)
	bodyBytes, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	// Prepare and send request
	client := resty.New()
	_, err = client.R().SetBody(bodyBytes).SetHeader(contentType, applicationJSON).Post(url)
	if err != nil {
		return err
	}
	return nil
}
