package monitor

// A Metric is used for a single metric type (e.g. gauge or counter) and keeps
// a value for each metric name.
type Metric[T float64 | int64] interface {
	Set(k string, v T)
	Add(k string, v T)
}
