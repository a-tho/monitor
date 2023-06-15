package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	monitor "github.com/a-tho/monitor/internal"
)

const (
	errPostMethod  = "use POST for saving metrics"
	errMetricPath  = "invalid metric path"
	errMetricValue = "invalid metric value"
	errMetricHTML  = "failed to generate HTML page with metrics"

	pageHead = `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Metrics</title>
	</head>
	<body>`
	gaugeHeader   = `<h1>Gauge metrics</h1>`
	counterHeader = `<h1>Counter metrics</h1>`
	pageFooter    = `
	</body>
	</html>`
)

// UpdHandler handles requests for adding metrics
func (s *server) UpdHandler(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, TypePath)
	name := chi.URLParam(r, NamePath)
	value := chi.URLParam(r, ValuePath)
	if name == "" {
		http.NotFound(w, r)
		return
	}

	switch typ {
	case GaugePath:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, errMetricValue, http.StatusBadRequest)
			return
		}
		s.gauge.Set(name, monitor.Gauge(v))
	case CounterPath:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, errMetricValue, http.StatusBadRequest)
			return
		}
		s.counter.Add(name, monitor.Counter(v))
	default:
		http.Error(w, errMetricPath, http.StatusBadRequest)
	}
}

func (s *server) GetValHandler(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch typ {
	case GaugePath:
		value, ok := s.gauge.Get(name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		v := strconv.FormatFloat(float64(value), 'f', 2, 64)
		w.Write([]byte(v))
	case CounterPath:
		value, ok := s.counter.Get(name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		v := strconv.FormatInt(int64(value), 10)
		w.Write([]byte(v))
	default:
		http.Error(w, errMetricPath, http.StatusBadRequest)
	}
}

func (s *server) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	gaugeBuf, err := s.gauge.HTML()
	if err != nil {
		http.Error(w, errMetricHTML, http.StatusInternalServerError)
		return
	}
	counterBuf, err := s.counter.HTML()
	if err != nil {
		http.Error(w, errMetricHTML, http.StatusInternalServerError)
		return
	}

	w.Write([]byte(pageHead))
	w.Write([]byte(gaugeHeader))
	w.Write(gaugeBuf.Bytes())
	w.Write([]byte(counterHeader))
	w.Write(counterBuf.Bytes())
	w.Write([]byte(pageFooter))
}
