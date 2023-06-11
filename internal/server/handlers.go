package server

import (
	"net/http"
	"strconv"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/go-chi/chi/v5"
)

const (
	errPostMethod  = "use POST for saving metrics"
	errMetricPath  = "invalid metric path"
	errMetricValue = "invalid metric value"
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
		return
	}
}
