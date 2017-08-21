package metrics

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/stretchr/testify/assert"
)

func TestGoStats(t *testing.T) {
	assert := assert.New(t)
	numGoMetrics := reflect.TypeOf(goMetrics{}).NumField()
	rec := make(recorder, numGoMetrics)

	duration := time.Millisecond * 100
	publisher := NewPublisher(rec, StatsDEncoder, duration*2, nil)
	ctx := context.Background()
	go publisher.Run(ctx)

	runner := NewGoStatsRunner(publisher, duration, Tag{"test", "life"})
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

type recorder chan string

func (r recorder) Write(b []byte) (n int, err error) {
	r <- string(b)
	return len(b), nil
}
