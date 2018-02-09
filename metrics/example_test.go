package metrics_test

import (
	"errors"
	"fmt"
	"net"

	"context"

	"github.com/socialpoint-labs/bsk/metrics"
)

func ExampleCounter() {
	discardAllMetrics := metrics.NewDiscardAll()
	go discardAllMetrics.Run(context.Background())

	counter := discardAllMetrics.Counter("test.counter")
	counter.Add(123)
	counter.Inc()

	// Output:
}

func ExampleGauge() {
	discardAllMetrics := metrics.NewDiscardAll()
	go discardAllMetrics.Run(context.Background())

	gauge := discardAllMetrics.Gauge("test.counter")
	gauge.Update(20)
	gauge.Update(10)

	// Output:
}

func ExampleEvent() {
	discardAllMetrics := metrics.NewDiscardAll()
	go discardAllMetrics.Run(context.Background())

	event := discardAllMetrics.Event("event title")
	event.Send()
	event.SendWithText("event text")

	// Output:
}

func Example_statsDBackend() {
	// because UDP is fire and forget this always work. If you create
	// and UDP server and then close it it will fail as expected.
	addr, err := net.ResolveUDPAddr("udp", "localhost:1234")
	if err != nil {
		panic(err)
	}

	client, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	publisher := metrics.NewPublisher(client, metrics.StatsDEncoder, metrics.FlushEvery5s, nil)
	go publisher.Run(context.Background())

	counter := publisher.Counter("test.counter")
	counter.Add(1)
	counter.Add(2)

	gauge := publisher.Gauge("test.gauge")
	gauge.Update(20)
	gauge.Update(10)

	event := publisher.Event("event title")
	event.Send()
	event.SendWithText("event text")

	// Output:
}

func ExampleErrorHandler() {
	errs := make(chan error)

	handler := func(err error) { errs <- err }

	publisher := metrics.NewPublisher(&FailingWriter{}, metrics.StatsDEncoder, metrics.FlushEvery5s, handler)
	go publisher.Run(context.Background())

	gauge := publisher.Gauge("test.counter")
	gauge.Update(20)

	publisher.Flush()
	fmt.Println(<-errs)

	// Output: error: don't care; I always fail
}

func ExampleWithNamespace() {
	discardAllMetrics := metrics.NewDiscardAll()
	go discardAllMetrics.Run(context.Background())
	// all metric names will be prefixed with "my_namespace."
	namespacedMetrics := metrics.WithNamespace(discardAllMetrics, "my_name")
	// will be named my_name.one
	namespacedMetrics.Counter("one").Inc()
	// you can even compose them further
	projectMetrics := metrics.WithNamespace(namespacedMetrics, "test_project")
	// will be named test_project.my_name.two
	projectMetrics.Counter("two").Inc()
	// Output:
}

func Example_withGoStats() {
	discardAllMetrics := metrics.NewDiscardAll()
	ctx := context.Background()
	go discardAllMetrics.Run(ctx)

	runner := metrics.NewGoStatsRunner(discardAllMetrics, metrics.FlushEvery15s)
	go runner.Run(ctx)

	// Output:
}

type FailingWriter struct {
}

func (w *FailingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("error: don't care; I always fail")
}
