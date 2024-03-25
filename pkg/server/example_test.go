package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"

	"github.com/go-resty/resty/v2"

	monitor "github.com/a-tho/monitor/internal"
)

func Exampleserver_Update() {
	// Prepare request url
	url := "http://localhost:8080/update/"

	// Prepare request body
	var delta int64 = 3
	metric := monitor.Metrics{
		ID:    "PollCount",
		MType: CounterPath,
		Delta: &delta,
	}
	var buf bytes.Buffer
	compressBuf := gzip.NewWriter(&buf)
	enc := json.NewEncoder(compressBuf)
	_ = enc.Encode(metric)
	compressBuf.Close()

	// Prepare and send request
	ctx := context.Background()
	client := resty.New()
	_, _ = client.R().
		SetBody(buf.Bytes()).
		SetHeader(contentEncoding, encodingGzip).
		SetHeader(contentType, typeApplicationJSON).
		SetContext(ctx).
		Post(url)
}

func Exampleserver_All() {
	// Prepare request url
	url := "http://localhost:8080/"

	// Prepare and send request
	ctx := context.Background()
	client := resty.New()
	_, _ = client.R().
		SetHeader(contentType, typeApplicationJSON).
		SetContext(ctx).
		Get(url)
}
