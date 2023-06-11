package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, srv *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, srv.URL+path, body)
	require.NoError(t, err)

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestServerUpdHandler(t *testing.T) {
	type request struct {
		method string
		path   string
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
				path:   "/",
			},
			want: want{
				code:        http.StatusNotFound,
				respBody:    "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
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
				contentType: "text/plain; charset=utf-8",
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
				contentType: "text/plain; charset=utf-8",
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
			gauge := storage.New[monitor.Gauge]()
			counter := storage.New[monitor.Counter]()
			srv := httptest.NewServer(New(gauge, counter))
			defer srv.Close()

			resp, respBody := testRequest(t, srv, tt.request.method, tt.request.path, nil)

			// Validate response
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			assert.Equal(t, tt.want.respBody, string(respBody))

			// Validate server storage
			assert.JSONEq(t, tt.want.gauge, gauge.String())
			assert.JSONEq(t, tt.want.counter, counter.String())
		})
	}
}
