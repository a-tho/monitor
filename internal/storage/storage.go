// Package storage implements a trivial storage.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	monitor "github.com/a-tho/monitor/internal"
)

// MemStorage represents the storage.
type MemStorage struct {
	db          *sql.DB
	DataGauge   map[string]monitor.Gauge
	DataCounter map[string]monitor.Counter

	file *os.File
	m    sync.Mutex
	// Whether recording is synchronuous
	syncMode bool
}

// New returns an initialized storage.
func New(dsn string, fileStoragePath string, storeInterval int, restore bool) (*MemStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	storage := MemStorage{
		db:          db,
		DataGauge:   make(map[string]monitor.Gauge),
		DataCounter: make(map[string]monitor.Counter),
	}

	if fileStoragePath != "" {
		file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return &storage, nil
		}

		if restore {
			dec := json.NewDecoder(file)
			var storageIn MemStorage
			if err = dec.Decode(&storageIn); err == nil {
				storage.DataGauge = storageIn.DataGauge
				storage.DataCounter = storageIn.DataCounter
			}
		}

		syncMode := storeInterval == 0
		storage.file = file
		storage.syncMode = syncMode

		if !syncMode {
			go storage.backup(storeInterval)
		}
	}

	return &storage, nil
}

func (s *MemStorage) backup(storeInterval int) {
	// Write to the file every storeInterval seconds
	var ticker <-chan time.Time
	if storeInterval > 0 {
		t := time.NewTicker(time.Duration(storeInterval) * time.Second)
		defer t.Stop()
		ticker = t.C
	}

	// and close the file when SIGINT is passed
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGQUIT)

	for {
		select {
		case <-ticker:
			s.writeToFile()
		case <-quit:
			return
		}
	}
}

// SetGauge inserts or updates a gauge metric value v for the key k.
func (s *MemStorage) SetGauge(k string, v monitor.Gauge) monitor.MetricRepo {
	s.m.Lock()
	s.DataGauge[k] = v
	s.m.Unlock()

	if s.syncMode {
		s.writeToFile()
	}

	return s
}

// AddCounter adds a counter metric value v for the key k.
func (s *MemStorage) AddCounter(k string, v monitor.Counter) monitor.MetricRepo {
	s.m.Lock()
	s.DataCounter[k] += v
	s.m.Unlock()

	if s.syncMode {
		s.writeToFile()
	}

	return s
}

// GetGauge retrieves the gauge value for the key k.
func (s *MemStorage) GetGauge(k string) (v monitor.Gauge, ok bool) {
	s.m.Lock()
	v, ok = s.DataGauge[k]
	s.m.Unlock()

	return
}

// GetCounter retrieves the counter value for the key k.
func (s *MemStorage) GetCounter(k string) (v monitor.Counter, ok bool) {
	s.m.Lock()
	v, ok = s.DataCounter[k]
	s.m.Unlock()

	return
}

// StringGauge produces a JSON representation of gauge metrics kept in the
// storage
func (s *MemStorage) StringGauge() string {
	s.m.Lock()
	out, _ := json.Marshal(s.DataGauge)
	s.m.Unlock()

	return string(out)
}

// StringCounter produces a JSON representation of counter metrics kept in the
// storage
func (s *MemStorage) StringCounter() string {
	s.m.Lock()
	out, _ := json.Marshal(s.DataCounter)
	s.m.Unlock()

	return string(out)
}

// WriteAllGauge writes gauge metrics as HTML into specified writer.
func (s *MemStorage) WriteAllGauge(wr io.Writer) error {
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		return err
	}

	s.m.Lock()
	if err = tmpl.Execute(wr, s.DataGauge); err != nil {
		return err
	}
	s.m.Unlock()

	return nil
}

// WriteAllCounter writes counter metrics as HTML into specified writer.
func (s *MemStorage) WriteAllCounter(wr io.Writer) error {
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		return err
	}

	s.m.Lock()
	if err = tmpl.Execute(wr, s.DataCounter); err != nil {
		return err
	}
	s.m.Unlock()

	return nil
}

func (s *MemStorage) PingContext(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *MemStorage) Close() error {
	s.writeToFile()

	_ = s.db.Close() // ignore error for now

	if s.file == nil {
		return nil
	}
	return s.file.Close()
}

func (s *MemStorage) writeToFile() (err error) {
	if s.file == nil {
		return nil
	}

	s.m.Lock()
	s.file.Truncate(0)
	enc := json.NewEncoder(s.file)
	err = enc.Encode(s)
	s.m.Unlock()

	return err
}

const metricsTemplate = `
		{{range $key, $value := .}}
			<p>{{$key}}: {{$value}}</p>
		{{end}}`
