package metrics

import (
	"errors"
	"fmt"
)

// DataDogLambdaPublisher creates a publisher that sends metrics to the datadog-lambda-go library.
//
// All metrics reported will be submitted as distribution metrics (https://docs.datadoghq.com/metrics/distributions/).
//
// Submitting events is not supported.
//
// Callers will typically pass `ddlambda.Metric` in the `f` argument of the constructor. The func type is introduced to avoid adding a dependency with the DataDog library.
//
// The documentation and implementation of the DataDog library for Lambda can be found in https://github.com/DataDog/datadog-lambda-go
func DataDogLambdaPublisher(f DataDogLambdaFunc, eh ErrorHandler) Metrics {
	if eh == nil {
		eh = DiscardErrors
	}

	return &dataDogLambdaPublisher{f: f, eh: eh}
}

// DataDogLambdaFunc is the signature of the function to send metrics to the DataDog Lambda library.
type DataDogLambdaFunc = func(metric string, value float64, tags ...string)

type dataDogLambdaPublisher struct {
	f  DataDogLambdaFunc
	eh ErrorHandler
}

// Counter returns a new counter with the provided name and tags
func (p *dataDogLambdaPublisher) Counter(name string, tags ...Tag) Counter {
	return &publisherCounter{publisherMetric{name: name, tags: tags, nf: p.notify}}
}

// Gauge returns a new Gauge with the provided name and tags
func (p *dataDogLambdaPublisher) Gauge(name string, tags ...Tag) Gauge {
	return &publisherGauge{publisherMetric{name: name, tags: tags, nf: p.notify}}
}

// Event returns a new Event with the provided title and tags
// Sending events is not supported, a no-op implementation is provided for compatibility
func (p *dataDogLambdaPublisher) Event(title string, tags ...Tag) Event {
	return &publisherEvent{publisherMetric{name: title, tags: tags, nf: p.notify}}
}

// Timer returns a new Timer with the provided name and tags
func (p *dataDogLambdaPublisher) Timer(name string, tags ...Tag) Timer {
	return &timerEvent{publisherMetric: publisherMetric{name: name, tags: tags, nf: p.notify}}
}

// Histogram returns a new Histogram with the provided name and tags
func (p *dataDogLambdaPublisher) Histogram(name string, tags ...Tag) Histogram {
	return &publisherHistogram{publisherMetric{name: name, tags: tags, nf: p.notify}}
}

func (p *dataDogLambdaPublisher) notify(op Op, name string, value interface{}, tags Tags) {
	if op == OpEventSend {
		p.eh(errors.New("sending event is not supported in the DataDog Lambda Publisher"))
		return
	}

	v, err := valueAsFloat64(value)
	if err != nil {
		p.eh(fmt.Errorf("could not publish metric `%s`: %w", name, err))
		return
	}

	p.f(name, v, tagsToStrings(tags)...)
}

func valueAsFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	}

	return 0, fmt.Errorf("value `%v` cannot be casted to float64", value)
}

func tagsToStrings(tags Tags) []string {
	ss := make([]string, len(tags))

	for n, tag := range tags {
		ss[n] = fmt.Sprintf("%s:%s", tag.Key, tag.Value)
	}

	return ss
}
