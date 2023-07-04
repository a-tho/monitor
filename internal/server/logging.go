package server

import (
	"net/http"
	"time"
)

type (
	respData struct {
		code int
		size int
	}

	logResponseWriter struct {
		http.ResponseWriter
		data *respData
	}
)

func (w *logResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.data.size += size
	return size, err
}

func (w *logResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.data.code = code
}

func (s server) WithLogging(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	wrapped := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		respData := respData{code: 200}

		lw := logResponseWriter{ResponseWriter: w, data: &respData}

		s.log.Info().Str("uri", r.RequestURI).Msg("")
		s.log.Info().Str("method", r.Method).Msg("")

		handler(&lw, r)

		duration := time.Since(start)

		s.log.Info().Dur("duration", duration).Msg("")
		s.log.Info().Int("code", respData.code).Msg("")
		s.log.Info().Int("size", respData.size).Msg("")
	}
	return wrapped
}
