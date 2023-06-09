package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	monitor "github.com/a-tho/monitor/internal"
)

const (
	errPostMethod  = "use POST for saving metrics"
	errMetricPath  = "invalid metric path"
	errMetricValue = "invalid metric value"
)

// UpdateHandler handles requests for adding metrics
func (s *server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errPostMethod, http.StatusMethodNotAllowed)
		return
	}

	typ, name, value, err := splitMetricPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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

func splitMetricPath(path string) (typ, name, value string, err error) {
	if len(path) == 0 {
		err = errors.New(errMetricPath)
		return
	}
	ss := strings.Split(path, "/")
	switch len(ss) {
	case 4:
		typ, name, value = ss[1], ss[2], ss[3]
	case 3:
		typ, name = ss[1], ss[2]
	case 2:
		typ = ss[1]
	default:
		err = errors.New(errMetricPath)
	}
	return
}
