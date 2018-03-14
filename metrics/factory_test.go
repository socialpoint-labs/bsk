package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewPublisherWithDSN(t *testing.T) {
	assert := assert.New(t)

	var cancelFuncs []context.CancelFunc
	for _, testCase := range []struct {
		DSN     string
		isValid bool
	}{
		{"http://%41:8080/", false},
		{"random://", false},
		{"discard://", true},
		{"datadog://", false},
		{"datadog://?namespace=my_namespace", true},
		{"datadog://?namespace=my_namespace&gostats=false", true},
	} {
		if testCase.isValid {
			publisher, runner := metrics.NewMetricsRunnerFromDSN(testCase.DSN)
			assert.NotNil(publisher)
			assert.NotNil(runner)
			ctx, cancel := context.WithCancel(context.Background())
			cancelFuncs = append(cancelFuncs, cancel)
			go runner.Run(ctx)
		} else {
			assert.Panics(func() { metrics.NewMetricsRunnerFromDSN(testCase.DSN) })
		}
	}

	time.Sleep(100 * time.Millisecond) // let them run some
	for _, cancel := range cancelFuncs {
		cancel()
	}
}
