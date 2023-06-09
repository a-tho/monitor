package main

import (
	"net/http"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
	"github.com/a-tho/monitor/internal/storage"
)

var (
	gauge   monitor.MetricRepo[monitor.Gauge]
	counter monitor.MetricRepo[monitor.Counter]
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	gauge = storage.New[monitor.Gauge]()
	counter = storage.New[monitor.Counter]()

	mux := server.New(gauge, counter)
	return http.ListenAndServe("localhost:8080", mux)
}
