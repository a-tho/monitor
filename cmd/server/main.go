package main

import (
	"net/http"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
	"github.com/a-tho/monitor/internal/storage"
)

var (
	gauge   monitor.Metric[float64]
	counter monitor.Metric[int64]
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	gauge = storage.New[float64]()
	counter = storage.New[int64]()

	mux := server.New(gauge, counter)
	return http.ListenAndServe("localhost:8080", mux)
}
