// Package middleware implements middleware necessary to process requests.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	contentEncoding = "Content-Encoding"
	encodingGzip    = "gzip"
	acceptEncoding  = "Accept-Encoding"
)

var gzipPool = sync.Pool{New: func() interface{} {
	w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
	return w
}}

type compResponseWriter struct {
	http.ResponseWriter
	cw *gzip.Writer
}

func newCompReponseWriter(w http.ResponseWriter) *compResponseWriter {
	z := gzipPool.Get().(*gzip.Writer)
	z.Reset(w)

	return &compResponseWriter{
		ResponseWriter: w,
		cw:             z,
	}
}

func (w *compResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *compResponseWriter) Write(p []byte) (int, error) {
	return w.cw.Write(p)
}

func (w *compResponseWriter) Close() error {
	defer gzipPool.Put(w.cw)

	return w.cw.Close()
}

type decompReaderCloser struct {
	io.ReadCloser
	dr *gzip.Reader
}

func newDecompReaderCloser(r io.ReadCloser) (*decompReaderCloser, error) {
	dr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &decompReaderCloser{ReadCloser: r, dr: dr}, nil
}

func (r *decompReaderCloser) Read(p []byte) (int, error) {
	return r.dr.Read(p)
}

func (r *decompReaderCloser) Close() error {
	if err := r.ReadCloser.Close(); err != nil {
		return err
	}
	return r.dr.Close()
}

// WithCompressing adds support for request and response compression.
func WithCompressing(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decompress request if necessary
		encodings := r.Header.Values(contentEncoding)
		isCompressed := contains(encodings, encodingGzip)
		if isCompressed {
			decompBody, err := newDecompReaderCloser(r.Body)
			if err == nil {
				defer decompBody.Close()
				r.Body = decompBody
			}
		}

		// Compress response if possible
		encodings = r.Header.Values(acceptEncoding)
		canCompress := contains(encodings, encodingGzip)
		if canCompress {
			w.Header().Add(contentEncoding, encodingGzip)
			compW := newCompReponseWriter(w)
			defer compW.Close()
			w = compW
		}

		handler(w, r)
	}
}

func contains(ss []string, str string) bool {
	for _, s := range ss {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}
