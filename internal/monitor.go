package monitor

type Gauge float64

type Counter int64

// A MetricRepo is used for a single metric type (e.g. gauge or counter) and
// stores a value for each metric name.
type MetricRepo[T Gauge | Counter] interface {
	Set(k string, v T)
	Add(k string, v T)
}

// An Observer is used to collect and transmit metrics.
type Observer interface {
	Observe()
}

// A MetricInstance holds a set of metrics collected roughly at the same moment
// in time.
type MetricInstance struct {
	Gauges    map[string]Gauge
	PollCount Counter
}
