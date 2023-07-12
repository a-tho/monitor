package telemetry

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
)

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
	// Prepare request url
	url := fmt.Sprintf("http://%s/%s/", o.SrvAddr, server.UpdPath)
	// Prepare request body
	var buf bytes.Buffer
	compressBuf := gzip.NewWriter(&buf)
	enc := json.NewEncoder(compressBuf)
	if err := enc.Encode(metric); err != nil {
		return err
	}
	compressBuf.Close()

	// Prepare and send request
	client := resty.New()
	if _, err := client.R().SetBody(buf.Bytes()).SetHeader(contentEncoding, encodingGzip).
		SetHeader(contentType, typeApplicationJSON).Post(url); err != nil {
		return err
	}
	return nil
}
