package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDataDogLambdaMetric(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	now := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	var tests = []struct {
		name  string
		op    Op
		value interface{}
		tags  Tags
		rate  float64
		out   string
	}{
		{"x", OpCounterAdd, 123, nil, 1, "MONITORING|0|123|count|x|#\n"},
		{"x", OpCounterAdd, time.Second, nil, 1, "MONITORING|0|1s|count|x|#\n"},
		{"x", OpCounterAdd, time.Millisecond, nil, 1, "MONITORING|0|1ms|count|x|#\n"},
		{"x", OpCounterAdd, time.Nanosecond * 5, nil, 1, "MONITORING|0|5ns|count|x|#\n"},
		{"x", OpCounterAdd, (time.Nanosecond * 5).Nanoseconds(), nil, 1, "MONITORING|0|5|count|x|#\n"},
		{"x", OpGaugeUpdate, 123, nil, 1, "MONITORING|0|123|gauge|x|#\n"},
		{"x", OpGaugeUpdate, 1.23, nil, 1, "MONITORING|0|1.23|gauge|x|#\n"},
		{"x", OpGaugeUpdate, -123, nil, 1, "MONITORING|0|-123|gauge|x|#\n"},
		{"x", OpGaugeUpdate, -123, nil, 0.250, "MONITORING|0|-123|gauge|x|#\n"},
		{"x", OpHistogramUpdate, 123, nil, 0.250, "MONITORING|0|123|histogram|x|#\n"},
		{"x", OpHistogramUpdate, 123.456, nil, 0.250, "MONITORING|0|123.456|histogram|x|#\n"},
		{"abc_xyz.sp.com", OpHistogramUpdate, 123.456, nil, 0.250, "MONITORING|0|123.456|histogram|abc_xyz.sp.com|#\n"},

		{"x", OpCounterAdd, 123, []Tag{{Key: "x", Value: "1"}}, 1, "MONITORING|0|123|count|x|#x:1\n"},
		{"x", OpCounterAdd, 123, []Tag{{Key: "x", Value: "1"}, {Key: "y", Value: 2}}, 1, "MONITORING|0|123|count|x|#x:1,y:2\n"},
		{"x", OpCounterAdd, 123, []Tag{{Key: "x", Value: "1"}, {Key: "y", Value: 2}, {Key: "z", Value: "value"}}, 1, "MONITORING|0|123|count|x|#x:1,y:2,z:value\n"},
	}

	for _, test := range tests {
		out, err := formatDataDogLambdaMetric(test.name, test.op, test.value, test.tags, now)
		a.NoError(err)
		a.Equal(test.out, out)
	}

	_, err := formatDataDogLambdaMetric("some.metric", OpEventSend, "event message", Tags{}, time.Now())
	a.Equal(errors.New(`datadog-lambda encoder: operation "event send" not supported`), err)

	_, err = formatDataDogLambdaMetric("some.metric", OpTimerStop, now, Tags{}, time.Now())
	a.Equal(errors.New(`datadog-lambda encoder: operation "timer stop" not supported`), err)
}
