package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/storage"
	"github.com/stretchr/testify/assert"
)

func Test_server_updateHandler(t *testing.T) {
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
				path:   "/" + GaugePath + "/" + "Apple" + "/" + "3",
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				respBody:    errPostMethod + "\n",
				contentType: "text/plain; charset=utf-8",
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
				path:   "/" + "wrongtype" + "/" + "Apple" + "/" + "3",
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
				path:   "/" + CounterPath + "/" + "Apple" + "/" + "wrongvalue",
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
				path:   "/" + GaugePath + "/" + "Apple" + "/" + "3",
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
			s := server{gauge: storage.New[monitor.Gauge](), counter: storage.New[monitor.Counter]()}

			r := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			w := httptest.NewRecorder()

			s.updateHandler(w, r)

			// Validate response
			res := w.Result()
			defer assert.NoError(t, res.Body.Close())

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			respBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.respBody, string(respBody))

			// Validate server storage
			assert.JSONEq(t, tt.want.gauge, s.gauge.String())
			assert.JSONEq(t, tt.want.counter, s.counter.String())
		})
	}
}
