package metrics

import (
	"strconv"
	"time"
)

// Provides some syntax sugar for better readability and easy integration
const (
	FlushEvery3s  = time.Second * 3
	FlushEvery5s  = time.Second * 5
	FlushEvery10s = time.Second * 10
	FlushEvery15s = time.Second * 15
)

// Op represents a metric operation
type Op uint

// Represent the update operations for each metric type
const (
	OpCounterAdd = iota
	OpGaugeUpdate
	OpHistogramUpdate
	OpEventSend
	OpTimerStop
)

func (op Op) String() string {
	name := []string{"counter add", "gauge update", "histogram update", "event send", "timer stop"}
	i := uint8(op)
	switch {
	case i <= uint8(OpTimerStop):
		return name[i]
	default:
		return strconv.Itoa(int(i))
	}
}

// Tag is a key/value pair associated with an observation for a specific
// metric. Tags may be ignored by implementations.
type Tag struct {
	Key   string
	Value interface{}
}

// NewTag returns a tag with the provided key and value
func NewTag(key string, value interface{}) Tag {
	return Tag{Key: key, Value: value}
}

// Tags is a slice of tags
type Tags []Tag

// NotifyFunc is the interface for a function that allows to notify metrics changes
type NotifyFunc func(Op, string, interface{}, Tags)

// ErrorHandler is the interface for a function that can be used to handle errors occurring in
// go-routines running in background and out of the control of the user
type ErrorHandler func(e error)

// DiscardErrors is an ErrorHandler that just discard the errors.
var DiscardErrors = func(error) {}

// Metrics is the common interface for a central registry and factory of metrics
type Metrics interface {
	// Provide a counter with the given name and tags
	Counter(name string, tags ...Tag) Counter

	// Provide a gauge with the given name and tags
	Gauge(name string, tags ...Tag) Gauge

	// Provide an event with the given name and tags
	Event(name string, tags ...Tag) Event

	// Provide a timer with the given name and tags
	Timer(name string, tags ...Tag) Timer

	// Provide a histogram with the given name and tags
	Histogram(name string, tags ...Tag) Histogram
}

// Metric is the interface for the common methods that all the metrics have.
type Metric interface {
	Name() string
	Tags() Tags
}

// Counter is a monotonically-increasing, unsigned, 64-bit integer used to
// capture the number of times an event has occurred. By tracking the deltas
// between measurements of a counter over intervals of time, an aggregation
// layer can derive rates, acceleration, etc.
type Counter interface {
	Metric
	Inc()
	Add(delta uint64)
	WithTags(tags ...Tag) Counter
	WithTag(key string, value interface{}) Counter
}

// Gauge captures instantaneous measurements of a value.
type Gauge interface {
	Metric
	Update(value interface{})
	WithTags(tags ...Tag) Gauge
	WithTag(key string, value interface{}) Gauge
}

// Event sends a single event
type Event interface {
	Metric
	Send()
	SendWithText(text string)
	WithTags(tags ...Tag) Event
	WithTag(key string, value interface{}) Event
}

// Timer times a duration
type Timer interface {
	Metric
	Start()
	Stop()
	WithTags(tags ...Tag) Timer
	WithTag(key string, value interface{}) Timer
}

// Histogram hold series of unsigned 64-bit integer values that enable obtaining
// their statistical distribution
type Histogram interface {
	Metric
	AddValue(value uint64)
	WithTags(tags ...Tag) Histogram
	WithTag(key string, value interface{}) Histogram
}

// WithNamespace composes Metrics so when creating types of metrics will
// namespace their names.
func WithNamespace(m Metrics, namespace string) Metrics {
	return NewNamespaced(m, namespace)
}
