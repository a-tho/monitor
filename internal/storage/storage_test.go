package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	monitor "github.com/a-tho/monitor/internal"
)

func TestStorageSetGauge(t *testing.T) {
	type args struct {
		k string
		v monitor.Gauge
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "set metric",
			args: args{k: "Apple", v: monitor.Gauge(3)},
			want: `{"Apple": 3.0}`,
		},
		{
			name: "reset metric",
			args: args{k: "Apple", v: monitor.Gauge(2)},
			want: `{"Apple": 2.0}`,
		},
		{
			name: "set another metric",
			args: args{k: "Cherry", v: monitor.Gauge(79)},
			want: `{"Apple": 2.0, "Cherry": 79.0}`,
		},
	}

	s := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s.SetGauge(tt.args.k, tt.args.v)
			assert.JSONEq(t, tt.want, s.StringGauge())
		})
	}
}

func TestStorageAddCounter(t *testing.T) {
	type args struct {
		k string
		v monitor.Counter
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "set metric",
			args: args{k: "Mississippi", v: monitor.Counter(3)},
			want: `{"Mississippi": 3}`,
		},
		{
			name: "reset metric",
			args: args{k: "Mississippi", v: monitor.Counter(2)},
			want: `{"Mississippi": 5}`,
		},
		{
			name: "set another metric",
			args: args{k: "Nile", v: monitor.Counter(79)},
			want: `{"Nile": 79, "Mississippi": 5}`,
		},
	}

	s := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s.AddCounter(tt.args.k, tt.args.v)
			assert.JSONEq(t, tt.want, s.StringCounter())
		})
	}
}
