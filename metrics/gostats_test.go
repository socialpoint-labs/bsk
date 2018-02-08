package metrics_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestGoStats(t *testing.T) {
	assert := assert.New(t)
	numGoMetrics := reflect.TypeOf(metrics.GoMetrics{}).NumField()
	rec := make(recorder, numGoMetrics)

	duration := time.Millisecond * 100
	publisher := metrics.NewPublisher(rec, metrics.StatsDEncoder, duration*2, nil)
	ctx := context.Background()
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
			assert.Fail("timeout reached and the publisher didn't flush out the metrics")
			return
		}
	}

	flushedMetrics := strings.Split(encodedFlushedMetrics, "\n")
	// remove last empty element due to how Split works
	flushedMetrics = flushedMetrics[:len(flushedMetrics)-1]
	// -1 because go.gc.pause depends on GC usage and its not deterministic
	assert.True(len(flushedMetrics) >= numGoMetrics-1)

	for _, flushedMetric := range flushedMetrics {
		assert.Contains(flushedMetric, "go.")
		assert.Contains(flushedMetric, "#test:life,vm:go")
	}
}
