package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	monitor "github.com/a-tho/monitor/internal"
)

const (
	errPostMethod  = "use POST for saving metrics"
	errMetricPath  = "invalid metric path"
	errMetricType  = "invalid metric type"
	errMetricName  = "invalid metric name"
	errMetricValue = "invalid metric value"
	errMetricHTML  = "failed to generate HTML page with metrics"
	errDecompress  = "failed to decompress request body"

	// HTML
	metricsTemplate = `
		{{range $key, $value := .}}
			<p>{{$key}}: {{$value}}</p>
		{{end}}`
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

	contentType         = "Content-Type"
	contentEncoding     = "Content-Encoding"
	acceptEncoding      = "Accept-Encoding"
	typeApplicationJSON = "application/json"
	typeTextHTML        = "text/html"
	encodingGzip        = "gzip"
)

// UpdateLegacy handles requests for adding metrics
func (s *server) UpdateLegacy(w http.ResponseWriter, r *http.Request) {
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
		s.metrics.SetGauge(name, monitor.Gauge(v))
	case CounterPath:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, errMetricValue, http.StatusBadRequest)
			return
		}
		s.metrics.AddCounter(name, monitor.Counter(v))
	default:
		http.Error(w, errMetricPath, http.StatusBadRequest)
	}
}

func (s *server) Update(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(contentType) != typeApplicationJSON {
		http.NotFound(w, r)
		return
	}

	var input monitor.Metrics
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&input)
	if err != nil {
		http.Error(w, errMetricValue, http.StatusBadRequest)
		return
	}

	var respValue float64
	switch input.MType {
	case GaugePath:

		if input.Value == nil {
			http.Error(w, errMetricValue, http.StatusBadRequest)
			return
		}
		s.metrics.SetGauge(input.ID, monitor.Gauge(*input.Value))

		respValue = *input.Value

	case CounterPath:

		if input.Delta == nil {
			http.Error(w, errMetricValue, http.StatusBadRequest)
			return
		}
		s.metrics.AddCounter(input.ID, monitor.Counter(*input.Delta))

		input.Delta = nil
		counter, _ := s.metrics.GetCounter(input.ID)
		respValue = float64(counter)

	default:
		http.Error(w, errMetricType, http.StatusBadRequest)
		return
	}

	input.Value = &respValue
	w.Header().Set(contentType, typeApplicationJSON)
	enc := json.NewEncoder(w)
	enc.Encode(input)
}

func (s *server) ValueLegacy(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch typ {
	case GaugePath:
		value, ok := s.metrics.GetGauge(name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		v := strconv.FormatFloat(float64(value), 'f', -1, 64)
		w.Write([]byte(v))
	case CounterPath:
		value, ok := s.metrics.GetCounter(name)
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

func (s *server) Value(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(contentType) != typeApplicationJSON {
		http.NotFound(w, r)
		return
	}

	var input monitor.Metrics
	dec := json.NewDecoder(r.Body)
	dec.Decode(&input)

	if input.ID == "" {
		http.Error(w, errMetricName, http.StatusBadRequest)
		return
	}
	switch input.MType {
	case GaugePath:

		val, ok := s.metrics.GetGauge(input.ID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		valFloat := float64(val)
		input.Value = &valFloat

	case CounterPath:

		count, ok := s.metrics.GetCounter(input.ID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		countInt := int64(count)
		input.Delta = &countInt

	default:
		http.Error(w, errMetricType, http.StatusBadRequest)
		return
	}

	w.Header().Set(contentType, typeApplicationJSON)
	enc := json.NewEncoder(w)
	enc.Encode(input)
}

func (s *server) All(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("metrics").Parse(metricsTemplate)
	if err != nil {
		http.Error(w, errMetricHTML, http.StatusInternalServerError)
		return
	}

	var gaugeBuf bytes.Buffer
	if err = tmpl.Execute(&gaugeBuf, s.metrics.GetAllGauge()); err != nil {
		http.Error(w, errMetricHTML, http.StatusInternalServerError)
		return
	}

	var counterBuf bytes.Buffer
	if err = tmpl.Execute(&counterBuf, s.metrics.GetAllCounter()); err != nil {
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
