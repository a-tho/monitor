// Package storage implements a trivial storage.
package storage

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	monitor "github.com/a-tho/monitor/internal"
)

// MemStorage represents the storage.
type MemStorage struct {
	// Database
	db                *sqlx.DB
	stmtSetGauge      *sqlx.Stmt
	stmtAddCounter    *sqlx.Stmt
	stmtGetGauge      *sqlx.Stmt
	stmtGetCounter    *sqlx.Stmt
	stmtStringGauge   *sqlx.Stmt
	stmtStringCounter *sqlx.Stmt
	stmtAllGauge      *sqlx.Stmt
	stmtAllCounter    *sqlx.Stmt

	// Memory
	DataGauge   map[string]monitor.Gauge
	DataCounter map[string]monitor.Counter
	file        *os.File
	m           sync.Mutex
	syncMode    bool // Whether recording is synchronuous
}

// New returns an initialized storage.
func New(dsn string, fileStoragePath string, storeInterval int, restore bool) (*MemStorage, error) {
	if dsn != "" {
		// DB may be available
		if storage, err := NewDBStorage(context.TODO(), dsn); err == nil {
			return storage, nil
		}
		// DB not available, revert to memory storage
	}

	// No available database, store in memory
	return NewMemStorage(fileStoragePath, storeInterval, restore)
}

func NewDBStorage(ctx context.Context, dsn string) (*MemStorage, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `
	CREATE TABLE gauge (
		"name" VARCHAR(50) PRIMARY KEY,
		"value" NUMERIC
	);`)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.ExecContext(ctx, `
	CREATE TABLE counter (
		"name" VARCHAR(50) PRIMARY KEY,
		"value" DOUBLE PRECISION
	);`)
	if err != nil {
		db.Close()
		return nil, err
	}

	stmtSetGauge, err := db.Preparex(`
	INSERT INTO gauge (name, value)
	VALUES
		($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET value = EXCLUDED.value;`)
	if err != nil {
		return nil, err
	}

	stmtAddCounter, err := db.Preparex(`
	INSERT INTO counter (name, value)
	VALUES
		($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET value = counter.value + EXCLUDED.value;`)
	if err != nil {
		return nil, err
	}

	stmtGetGauge, err := db.Preparex(`
	SELECT value FROM gauge WHERE name = $1`)
	if err != nil {
		return nil, err
	}

	stmtGetCounter, err := db.Preparex(`
	SELECT value FROM counter WHERE name = $1`)
	if err != nil {
		return nil, err
	}

	// https://alphahydrae.com/2021/02/how-to-export-postgresql-data-to-a-json-file/
	stmtStringGauge, err := db.Preparex(`
	SELECT json_agg(row_to_json(gauge)) FROM gauge;`)
	if err != nil {
		return nil, err
	}

	stmtStringCounter, err := db.Preparex(`
	SELECT json_agg(row_to_json(counter)) FROM counter;`)
	if err != nil {
		return nil, err
	}

	stmtAllGauge, err := db.Preparex(`
	SELECT name, value FROM gauge`)
	if err != nil {
		return nil, err
	}

	stmtAllCounter, err := db.Preparex(`
	SELECT name, value FROM counter`)
	if err != nil {
		return nil, err
	}

	storage := MemStorage{
		db:                db,
		stmtSetGauge:      stmtSetGauge,
		stmtAddCounter:    stmtAddCounter,
		stmtGetGauge:      stmtGetGauge,
		stmtGetCounter:    stmtGetCounter,
		stmtStringGauge:   stmtStringGauge,
		stmtStringCounter: stmtStringCounter,
		stmtAllGauge:      stmtAllGauge,
		stmtAllCounter:    stmtAllCounter,
	}
	return &storage, nil
}

func NewMemStorage(fileStoragePath string, storeInterval int, restore bool) (*MemStorage, error) {
	storage := MemStorage{
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
			go storage.memBackup(storeInterval)
		}
	}

	return &storage, nil
}

func (s *MemStorage) memBackup(storeInterval int) {
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
func (s *MemStorage) SetGauge(ctx context.Context, k string, v monitor.Gauge) (monitor.MetricRepo, error) {
	if s.db != nil {
		_, err := s.stmtSetGauge.ExecContext(ctx, k, v)
		return s, err
	}

	// No DB, use memory
	s.m.Lock()
	s.DataGauge[k] = v
	s.m.Unlock()

	if s.syncMode {
		s.writeToFile()
	}

	return s, nil
}

// AddCounter adds a counter metric value v for the key k.
func (s *MemStorage) AddCounter(ctx context.Context, k string, v monitor.Counter) (monitor.MetricRepo, error) {
	if s.db != nil {
		_, err := s.stmtAddCounter.ExecContext(ctx, k, v)
		return s, err
	}

	s.m.Lock()
	s.DataCounter[k] += v
	s.m.Unlock()

	if s.syncMode {
		s.writeToFile()
	}

	return s, nil
}

// GetGauge retrieves the gauge value for the key k.
func (s *MemStorage) GetGauge(ctx context.Context, k string) (v monitor.Gauge, ok bool) {
	if s.db != nil {
		row := s.stmtGetGauge.QueryRow(k)
		if err := row.Scan(&v); err != nil {
			return v, false
		}
		return v, true
	}

	s.m.Lock()
	v, ok = s.DataGauge[k]
	s.m.Unlock()

	return
}

// GetCounter retrieves the counter value for the key k.
func (s *MemStorage) GetCounter(ctx context.Context, k string) (v monitor.Counter, ok bool) {
	if s.db != nil {
		row := s.stmtGetCounter.QueryRowContext(ctx, k)
		if err := row.Scan(&v); err != nil {
			return v, false
		}
		return v, true
	}

	s.m.Lock()
	v, ok = s.DataCounter[k]
	s.m.Unlock()

	return
}

// StringGauge produces a JSON representation of gauge metrics kept in the
// storage
func (s *MemStorage) StringGauge(ctx context.Context) (string, error) {
	if s.db != nil {
		row := s.stmtStringCounter.QueryRowContext(ctx)
		var enc string
		err := row.Scan(&enc)
		return enc, err
	}

	s.m.Lock()
	out, err := json.Marshal(s.DataGauge)
	s.m.Unlock()

	return string(out), err
}

// StringCounter produces a JSON representation of counter metrics kept in the
// storage
func (s *MemStorage) StringCounter(ctx context.Context) (string, error) {
	if s.db != nil {
		row := s.stmtStringCounter.QueryRowContext(ctx)
		var enc string
		err := row.Scan(&enc)
		return enc, err
	}

	s.m.Lock()
	out, err := json.Marshal(s.DataCounter)
	s.m.Unlock()

	return string(out), err
}

// WriteAllGauge writes gauge metrics as HTML into specified writer.
func (s *MemStorage) WriteAllGauge(ctx context.Context, wr io.Writer) error {
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		return err
	}

	if s.db != nil {
		var (
			key   string
			value monitor.Gauge
		)
		dataGauge := make(map[string]monitor.Gauge)

		rows, err := s.stmtAllGauge.QueryContext(ctx)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			if err = rows.Scan(&key, &value); err != nil {
				return err
			}
			dataGauge[key] = value
		}
		if rows.Err() != nil {
			return err
		}

		err = tmpl.Execute(wr, dataGauge)

		return err
	}

	s.m.Lock()
	err = tmpl.Execute(wr, s.DataGauge)
	s.m.Unlock()

	return err
}

// WriteAllCounter writes counter metrics as HTML into specified writer.
func (s *MemStorage) WriteAllCounter(ctx context.Context, wr io.Writer) error {
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		return err
	}

	if s.db != nil {
		var (
			key   string
			value monitor.Gauge
		)
		dataGauge := make(map[string]monitor.Gauge)
		rows, err := s.stmtAllCounter.QueryContext(ctx)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			if err = rows.Scan(&key, &value); err != nil {
				return err
			}
			dataGauge[key] = value
		}
		if rows.Err() != nil {
			return err
		}

		err = tmpl.Execute(wr, dataGauge)

		return err
	}

	s.m.Lock()
	err = tmpl.Execute(wr, s.DataCounter)
	s.m.Unlock()

	return err
}

func (s *MemStorage) PingContext(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *MemStorage) Close() error {
	if s.db != nil {
		s.stmtSetGauge.Close()
		s.stmtAddCounter.Close()
		s.stmtGetGauge.Close()
		s.stmtGetCounter.Close()
		s.stmtStringGauge.Close()
		s.stmtStringCounter.Close()
		return s.db.Close()
	}

	s.writeToFile()

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
