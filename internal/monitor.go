package monitor

type Gauge float64

type Counter int64

type Metrics struct {
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // parameter, taking a value of gauge or counter
	Delta *int64   `json:"delta,omitempty"` // metric value in case of a counter
	Value *float64 `json:"value,omitempty"` // metric value in case of a gauge
}

// A MetricRepo is used for a single metric type (e.g. gauge or counter) and
// stores a value for each metric name.
type MetricRepo interface {
	SetGauge(k string, v Gauge) MetricRepo
	GetGauge(k string) (v Gauge, ok bool)
	StringGauge() string
	GetAllGauge() map[string]Gauge

	AddCounter(k string, v Counter) MetricRepo
	GetCounter(k string) (v Counter, ok bool)
	StringCounter() string
	GetAllCounter() map[string]Counter
	// HTML() (*bytes.Buffer, error)
}

// An Observer is used to collect and transmit metrics.
type Observer interface {
	Observe() error
}
