package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	postMethodError = "use POST for saving metrics"
	metricPathError = "invalid metric path"
)

// updateHandler handles requests for adding metrics
func (s *server) updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, postMethodError, http.StatusMethodNotAllowed)
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
	case gaugePath:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.gauge.Set(name, v)
	case counterPath:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.counter.Add(name, v)
	default:
		http.Error(w, metricPathError, http.StatusBadRequest)
		return
	}
}

func splitMetricPath(path string) (typ, name, value string, err error) {
	if len(path) == 0 {
		err = fmt.Errorf(metricPathError)
		return
	}
	ss := strings.Split(path, "/")
	switch len(ss) {
	case 3:
		typ, name, value = ss[0], ss[1], ss[2]
	case 2:
		typ, name = ss[0], ss[1]
	case 1:
		typ = ss[0]
	default:
		err = fmt.Errorf(metricPathError)
	}
	return
}
