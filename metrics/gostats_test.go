package metrics_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/metrics"
)

func TestGoStats(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := assert.New(t)
	numGoMetrics := reflect.TypeOf(metrics.GoMetrics{}).NumField()
	rec := make(recorder, numGoMetrics)

	duration := time.Millisecond * 100
	publisher := metrics.NewPublisher(rec, metrics.StatsDEncoder, duration*2, nil)
	go publisher.Run(ctx)

	runner := metrics.NewGoStatsRunner(publisher, duration, metrics.Tag{Key: "test", Value: "life"})
	go runner.Run(ctx)

	time.Sleep(duration)
	timeout := time.After(duration * 10)
	var encodedFlushedMetrics string
loop:
	for {
		select {
		case encodedFlushedMetrics = <-rec:
			break loop
		case <-timeout:
			a.Fail("timeout reached and the publisher didn't flush out the metrics")
			return
		}
	}

	flushedMetrics := strings.Split(encodedFlushedMetrics, "\n")
	// remove last empty element due to how Split works
	flushedMetrics = flushedMetrics[:len(flushedMetrics)-1]
	// -1 because go.gc.pause depends on GC usage and its not deterministic
	a.True(len(flushedMetrics) >= numGoMetrics-1)

	for _, flushedMetric := range flushedMetrics {
		a.Contains(flushedMetric, "go.")
		a.Contains(flushedMetric, "#test:life,vm:go")
	}
}
