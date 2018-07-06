package metrics_test

import (
	"context"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewPublisherWithDSN(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

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
		{"datadog-lambda://", false},
		{"datadog-lambda://?namespace=my_namespace", true},
		{"datadog-lambda://?namespace=my_namespace&gostats=false", true},
	} {
		if testCase.isValid {
			publisher, runner := metrics.NewMetricsRunnerFromDSN(testCase.DSN)
			a.NotNil(publisher)
			a.NotNil(runner)
			ctx, cancel := context.WithCancel(context.Background())
			cancelFuncs = append(cancelFuncs, cancel)
			go runner.Run(ctx)
		} else {
			a.Panics(func() { metrics.NewMetricsRunnerFromDSN(testCase.DSN) })
		}
	}

	time.Sleep(100 * time.Millisecond) // let them run some
	for _, cancel := range cancelFuncs {
		cancel()
	}
}
