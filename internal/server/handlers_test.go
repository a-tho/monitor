package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/storage"
)

const (
	notFoundResponse = "404 page not found\n"

	textPlain = "text/plain; charset=utf-8"
)

type request struct {
	method  string
	path    string
	body    io.Reader
	headers map[string]string
}

type want struct {
	// Response-related fields
	code        int
	respBody    string
	contentType string

	//Storage-related fields
	gauge   string
	counter string
}

type state struct {
	metrics monitor.MetricRepo
}

func testRequest(t require.TestingT, srv *httptest.Server, method, path string, headers map[string]string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, srv.URL+path, body)
	require.NoError(t, err)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func compressJSONBody(t require.TestingT, body interface{}) io.Reader {
	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()

	err := json.NewEncoder(gzipWriter).Encode(body)
	require.NoError(t, err)

	err = gzipWriter.Close()
	require.NoError(t, err)

	return &buf
}

func TestServerUpdLegacyHandler(t *testing.T) {
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "invalid request method",
			request: request{
				method: http.MethodGet,
				path:   "/" + UpdPath + "/" + GaugePath + "/" + "Apple" + "/" + "3",
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				respBody:    "",
				contentType: "",
				gauge:       `{}`,
				counter:     `{}`,
			},
		},
		{
			name: "no metric name",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath,
			},
			want: want{
				code:        http.StatusNotFound,
				respBody:    notFoundResponse,
				contentType: textPlain,
				gauge:       `{}`,
				counter:     `{}`,
			},
		},
		{
			name: "wrong metric type",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath + "/" + "wrongtype" + "/" + "Apple" + "/" + "3",
			},
			want: want{
				code:        http.StatusBadRequest,
				respBody:    errMetricPath + "\n",
				contentType: textPlain,
				gauge:       `{}`,
				counter:     `{}`,
			},
		},
		{
			name: "wrong metric value for counter",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath + "/" + CounterPath + "/" + "Apple" + "/" + "wrongvalue",
			},
			want: want{
				code:        http.StatusBadRequest,
				respBody:    errMetricValue + "\n",
				contentType: textPlain,
				gauge:       `{}`,
				counter:     `{}`,
			},
		},
		{
			name: "valid gauge request",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath + "/" + GaugePath + "/" + "Apple" + "/" + "3",
			},
			want: want{
				code:        http.StatusOK,
				respBody:    "",
				contentType: "",
				gauge:       `{"Apple": 3}`,
				counter:     "{}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := storage.New(context.Background(), "", "", 5, false)
			if assert.NoError(t, err) {
				srv := httptest.NewServer(NewServer(metrics))
				defer srv.Close()

				resp, respBody := testRequest(t, srv, tt.request.method, tt.request.path, nil, nil)
				defer resp.Body.Close()

				// Validate response
				assert.Equal(t, tt.want.code, resp.StatusCode)
				assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.respBody, string(respBody))

				// Validate server storage
				gaugeJSON, err := metrics.StringGauge(context.TODO())
				assert.NoError(t, err)
				assert.JSONEq(t, tt.want.gauge, gaugeJSON)
				counterJSON, err := metrics.StringCounter(context.TODO())
				assert.NoError(t, err)
				assert.JSONEq(t, tt.want.counter, counterJSON)
			}
		})
	}
}

func TestServerUpdHandler(t *testing.T) {
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "invalid request method",
			request: request{
				method: http.MethodGet,
				path:   "/" + UpdPath + "/",
				body: strings.NewReader(
					`{"id":"Apple","type":"gauge","value":3}`,
				),
				headers: map[string]string{contentType: typeApplicationJSON},
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				respBody:    "",
				contentType: "",
				gauge:       `{}`,
				counter:     `{}`,
			},
		},
		{
			name: "wrong metric type",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath + "/",
				body: strings.NewReader(
					`{"id":"Apple","type":"wrongtype","value":3.0}`,
				),
				headers: map[string]string{contentType: typeApplicationJSON},
			},
			want: want{
				code:        http.StatusBadRequest,
				respBody:    errMetricType + "\n",
				contentType: textPlain,
				gauge:       `{}`,
				counter:     `{}`,
			},
		},
		{
			name: "no content type",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath + "/",
				body: strings.NewReader(
					`{"id":"Apple","type":"counter","value":3}`,
				),
			},
			want: want{
				code:        http.StatusNotFound,
				respBody:    notFoundResponse,
				contentType: textPlain,
				gauge:       "{}",
				counter:     "{}",
			},
		},
		{
			name: "wrong metric value for counter",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath + "/",
				body: strings.NewReader(
					`{"id":"Apple","type":"counter","value":"wrongvalue"}`,
				),
				headers: map[string]string{contentType: typeApplicationJSON},
			},
			want: want{
				code:        http.StatusBadRequest,
				respBody:    errMetricValue + "\n",
				contentType: textPlain,
				gauge:       `{}`,
				counter:     `{}`,
			},
		},
		{
			name: "valid gauge request",
			request: request{
				method: http.MethodPost,
				path:   "/" + UpdPath + "/",
				body: strings.NewReader(
					`{"id":"Apple","type":"gauge","value":3}`,
				),
				headers: map[string]string{contentType: typeApplicationJSON},
			},
			want: want{
				code:        http.StatusOK,
				respBody:    `{"id":"Apple","type":"gauge","value":3}` + "\n",
				contentType: typeApplicationJSON,
				gauge:       `{"Apple": 3}`,
				counter:     "{}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := storage.New(context.Background(), "", "", 5, false)
			if assert.NoError(t, err) {
				srv := httptest.NewServer(NewServer(metrics))
				defer srv.Close()

				resp, respBody := testRequest(
					t,
					srv,
					tt.request.method,
					tt.request.path,
					tt.request.headers,
					tt.request.body,
				)
				defer resp.Body.Close()

				// Validate response
				assert.Equal(t, tt.want.code, resp.StatusCode)
				assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.respBody, string(respBody))

				// Validate server storage
				gaugeJSON, err := metrics.StringGauge(context.TODO())
				assert.NoError(t, err)
				assert.JSONEq(t, tt.want.gauge, gaugeJSON)
				counterJSON, err := metrics.StringCounter(context.TODO())
				assert.NoError(t, err)
				assert.JSONEq(t, tt.want.counter, counterJSON)
			}
		})
	}
}

// func TestGetValHandler(t *testing.T) {
// 	// I don't know what the best practices for initializing exernal storage is
// 	// so I updated storage interface methods for modifying it: now they return
// 	// the storage, so that I can chain several storage modification operations
// 	// in one line (see the line for initializing the gauge storage below)

// 	tests := []struct {
// 		name    string
// 		request request
// 		want    want
// 		state   state
// 	}{
// 		{
// 			name: "no such metric name",
// 			request: request{
// 				method: http.MethodGet,
// 				path:   "/" + ValuePath + "/" + GaugePath + "/" + "Apple",
// 			},
// 			want: want{
// 				code:        http.StatusNotFound,
// 				respBody:    notFoundResponse,
// 				contentType: textPlain,
// 			},
// 			state: state{
// 				metrics: storage.New("", 5, false).SetGauge("Peach", monitor.Gauge(4.0)),
// 			},
// 		},
// 		{
// 			name: "metric value is present",
// 			request: request{
// 				method: http.MethodGet,
// 				path:   "/" + ValuePath + "/" + GaugePath + "/" + "Apple",
// 			},
// 			want: want{
// 				code:        http.StatusOK,
// 				respBody:    "20",
// 				contentType: textPlain,
// 			},
// 			state: state{
// 				metrics: storage.New("", 5, false).SetGauge("Apple", monitor.Gauge(20.0)),
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			srv := httptest.NewServer(NewServer(tt.state.metrics))
// 			defer srv.Close()

// 			resp, respBody := testRequest(t, srv, tt.request.method, tt.request.path, nil)
// 			defer resp.Body.Close()

// 			// Validate response
// 			assert.Equal(t, tt.want.code, resp.StatusCode)
// 			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
// 			assert.Equal(t, tt.want.respBody, respBody)
// 		})
// 	}
// }

func BenchmarkUpdateGauge(b *testing.B) {
	log.Logger = log.Logger.Level(zerolog.ErrorLevel)

	metrics, err := storage.New(context.Background(), "", "", 5, false)
	if assert.NoError(b, err) {
		// Init the server to test
		srv := httptest.NewServer(NewServer(metrics))
		defer srv.Close()

		method := http.MethodPost
		input := monitor.Metrics{MType: GaugePath}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()

			input.ID = strconv.Itoa(i)
			iFloat := float64(i)
			input.Value = &iFloat

			b.StartTimer()

			var body bytes.Buffer
			enc := json.NewEncoder(&body)
			enc.Encode(input)

			testRequest(b, srv, method,
				"/"+UpdPath,
				nil, &body)
		}
	}
}

func BenchmarkUpdateCounter(b *testing.B) {
	log.Logger = log.Logger.Level(zerolog.ErrorLevel)

	metrics, err := storage.New(context.Background(), "", "", 5, false)
	if assert.NoError(b, err) {
		// Init the server to test
		srv := httptest.NewServer(NewServer(metrics))
		defer srv.Close()

		b.ResetTimer()
		method := http.MethodPost
		for i := 0; i < b.N; i++ {
			b.StopTimer()

			iStr := strconv.Itoa(i)

			b.StartTimer()

			testRequest(b, srv, method,
				"/"+UpdPath+"/"+GaugePath+"/"+iStr+"/"+iStr,
				nil, nil)
		}
	}
}

func BenchmarkUpdatesGaugeAdd(b *testing.B) {
	log.Logger = log.Logger.Level(zerolog.ErrorLevel)

	metrics, err := storage.New(context.Background(), "", "", 5, false)
	if assert.NoError(b, err) {
		// Init the server to test
		srv := httptest.NewServer(NewServer(metrics))
		defer srv.Close()

		// Init empty inputs
		batchLen := 3000
		batchesCount := 3000
		inputs := make([][]monitor.Metrics, batchesCount)
		for i := range inputs {
			inputs[i] = make([]monitor.Metrics, batchLen)
		}

		// Init non-overlapping inputs for "Add"
		for i := range inputs {
			for j := range inputs[i] {
				// "02220333" for batchNum = 222, batchPos = 333
				inputs[i][j].ID = fmt.Sprintf(
					"%0"+strconv.Itoa(len(strconv.Itoa(batchesCount)))+"d"+
						"%0"+strconv.Itoa(len(strconv.Itoa(batchLen)))+"d",
					i, j,
				)
				inputs[i][j].MType = GaugePath
				value := float64(j)
				inputs[i][j].Value = &value
			}
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			var body bytes.Buffer
			enc := json.NewEncoder(&body)
			enc.Encode(inputs[i%batchesCount])
			b.StartTimer()

			testRequest(b, srv, http.MethodPost, "/"+UpdsPath, nil, &body)
		}
	}
}

func BenchmarkUpdatesGaugeUpdate(b *testing.B) {
	log.Logger = log.Logger.Level(zerolog.ErrorLevel)

	metrics, err := storage.New(context.Background(), "", "", 5, false)
	if assert.NoError(b, err) {
		// Init the server to test
		srv := httptest.NewServer(NewServer(metrics))
		defer srv.Close()

		// Init empty inputs
		batchLen := 3000
		batchesCount := 3000
		inputs := make([][]monitor.Metrics, batchesCount)
		for i := range inputs {
			inputs[i] = make([]monitor.Metrics, batchLen)
		}

		// Init overlapping inputs for "Update"
		for i := range inputs {
			for j := range inputs[i] {
				inputs[i][j].ID = strconv.Itoa(j)
				inputs[i][j].MType = GaugePath
				value := float64(j)
				inputs[i][j].Value = &value
			}
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			var body bytes.Buffer
			enc := json.NewEncoder(&body)
			enc.Encode(inputs[i%batchesCount])
			b.StartTimer()

			testRequest(b, srv, http.MethodPost, "/"+UpdsPath, nil, &body)
		}
	}
}

func BenchmarkAll(b *testing.B) {
	log.Logger = log.Logger.Level(zerolog.ErrorLevel)

	metrics, err := storage.New(context.Background(), "", "", 5, false)
	if assert.NoError(b, err) {
		// Init the server to test
		srv := httptest.NewServer(NewServer(metrics))
		defer srv.Close()

		// Set up the values in the storage
		valuesCount := 3000
		for i := 0; i < valuesCount; i++ {
			metrics.SetGauge(context.Background(), strconv.Itoa(i), monitor.Gauge(i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
			require.NoError(b, err)
			resp, err := srv.Client().Do(req)
			require.NoError(b, err)
			defer resp.Body.Close()
		}
	}
}
