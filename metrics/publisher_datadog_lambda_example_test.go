package metrics_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/metrics"
)

func ExampleDataDogLambdaPublisher() {
	f := func(metric string, value float64, tags ...string) {
		fmt.Printf("%s: %f, %v\n", metric, value, tags)
	}
	eh := func(e error) {
		fmt.Println(e.Error())
	}

	publisher := metrics.DataDogLambdaPublisher(f, eh)

	counter := publisher.Counter("commands_executed", metrics.Tag{Key: "host", Value: "life"}, metrics.Tag{Key: "project", Value: "bsk"})
	counter.Add(10)
	counter.WithTags(metrics.NewTag("cfoo", "cbar")).Inc()
	counter.Inc()

	gauge := publisher.Gauge("memory", metrics.Tag{Key: "host", Value: "life"}, metrics.Tag{Key: "project", Value: "bsk"})
	gauge.WithTags(metrics.NewTag("gfoo", "gbar")).Update(100.99)
	gauge.Update("invalid value")

	event := publisher.Event("events are not supported")
	event.Send()

	histogram := publisher.Histogram("ping", metrics.Tag{Key: "host", Value: "life"}, metrics.Tag{Key: "project", Value: "bsk"})
	histogram.WithTag("hfoo", "hbar")
	histogram.AddValue(100)
	histogram.AddValue(123)

	// Output:
	// commands_executed: 10.000000, [host:life project:bsk]
	// commands_executed: 1.000000, [host:life project:bsk cfoo:cbar]
	// commands_executed: 1.000000, [host:life project:bsk]
	// memory: 100.990000, [host:life project:bsk gfoo:gbar]
	// could not publish metric `memory`: value `invalid value` cannot be casted to float64
	// sending event is not supported in the DataDog Lambda Publisher
	// ping: 100.000000, [host:life project:bsk hfoo:hbar]
	// ping: 123.000000, [host:life project:bsk hfoo:hbar]
}

func TestDataDogLambdaPublisher_Timer(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	buff := bytes.Buffer{}

	f := func(metric string, value float64, tags ...string) {
		buff.WriteString(fmt.Sprintf("%s: %f, %v\n", metric, value, tags))
	}
	eh := func(e error) {
		t.Fatal("no error expected")
	}

	publisher := metrics.DataDogLambdaPublisher(f, eh)

	timer := publisher.Timer("request_duration", metrics.Tag{Key: "host", Value: "life"}, metrics.Tag{Key: "project", Value: "bsk"})
	timer.WithTag("tfoo", "tbar").Start()
	timer.Stop()

	out := buff.String()
	a.Contains(out, "request_duration: 0.0")
	a.Contains(out, "[host:life project:bsk tfoo:tbar]")
}
