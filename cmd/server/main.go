package main

import (
	"flag"
	"net/http"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/server"
	"github.com/a-tho/monitor/internal/storage"
)

var (
	// Flags
	srvAddr string

	// Storage
	gauge   monitor.MetricRepo[monitor.Gauge]
	counter monitor.MetricRepo[monitor.Counter]
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	flag.StringVar(&srvAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	gauge = storage.New[monitor.Gauge]()
	counter = storage.New[monitor.Counter]()

	mux := server.New(gauge, counter)
	return http.ListenAndServe(srvAddr, mux)
}
