package metrics_test

import (
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestStatsDEncoder(t *testing.T) {
	var tests = []struct {
		name  string
		op    metrics.Op
		value interface{}
		tags  metrics.Tags
		rate  float64
		out   string
	}{
		{"x", metrics.OpCounterAdd, 123, nil, 1, "x:123|c|@1.0000|#\n"},
		{"x", metrics.OpCounterAdd, time.Second, nil, 1, "x:1s|c|@1.0000|#\n"},
		{"x", metrics.OpCounterAdd, time.Millisecond, nil, 1, "x:1ms|c|@1.0000|#\n"},
		{"x", metrics.OpCounterAdd, time.Nanosecond * 5, nil, 1, "x:5ns|c|@1.0000|#\n"},
		{"x", metrics.OpCounterAdd, (time.Nanosecond * 5).Nanoseconds(), nil, 1, "x:5|c|@1.0000|#\n"},
		{"x", metrics.OpGaugeUpdate, 123, nil, 1, "x:123|g|@1.0000|#\n"},
		{"x", metrics.OpGaugeUpdate, 1.23, nil, 1, "x:1.23|g|@1.0000|#\n"},
		{"x", metrics.OpGaugeUpdate, -123, nil, 1, "x:-123|g|@1.0000|#\n"},
		{"x", metrics.OpGaugeUpdate, -123, nil, 0.250, "x:-123|g|@0.2500|#\n"},
		{"x", metrics.OpHistogramUpdate, 123, nil, 0.250, "x:123|h|@0.2500|#\n"},
		{"x", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "x:123.456|h|@0.2500|#\n"},
		{"abc_xyz.sp.com", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "abc_xyz.sp.com:123.456|h|@0.2500|#\n"},

		{"x", metrics.OpCounterAdd, 123, []metrics.Tag{{Key: "x", Value: "1"}}, 1, "x:123|c|@1.0000|#x:1\n"},
		{"x", metrics.OpCounterAdd, 123, []metrics.Tag{{Key: "x", Value: "1"}, {Key: "y", Value: 2}}, 1, "x:123|c|@1.0000|#x:1,y:2\n"},
		{"x", metrics.OpCounterAdd, 123, []metrics.Tag{{Key: "x", Value: "1"}, {Key: "y", Value: 2}, {Key: "z", Value: "value"}}, 1, "x:123|c|@1.0000|#x:1,y:2,z:value\n"},

		{"event title", metrics.OpEventSend, "event text", nil, 1, "_e{11,10}:event title|event text|#\n"},
		{"event title", metrics.OpEventSend, "event text", []metrics.Tag{{Key: "x", Value: "1"}, {Key: "y", Value: 2}, {Key: "z", Value: "value"}}, 1, "_e{11,10}:event title|event text|#x:1,y:2,z:value\n"},

		{"timer", metrics.OpTimerStop, 123.321, nil, 1, "timer:123.321|ms|@1.0000|#\n"},
	}

	for _, test := range tests {
		out, err := metrics.StatsDEncoder(test.name, test.op, test.value, test.tags, test.rate)
		assert.NoError(t, err)
		assert.Equal(t, test.out, out)
	}
}

func TestLibratoStatsDEncoder(t *testing.T) {
	var tests = []struct {
		name  string
		op    metrics.Op
		value interface{}
		tags  metrics.Tags
		rate  float64
		out   string
	}{
		{"x", metrics.OpCounterAdd, 123, nil, 1, "x:123|c|@1.0000\n"},
		{"x", metrics.OpCounterAdd, time.Second, nil, 1, "x:1s|c|@1.0000\n"},
		{"x", metrics.OpCounterAdd, time.Millisecond, nil, 1, "x:1ms|c|@1.0000\n"},
		{"x", metrics.OpCounterAdd, time.Nanosecond * 5, nil, 1, "x:5ns|c|@1.0000\n"},
		{"x", metrics.OpCounterAdd, (time.Nanosecond * 5).Nanoseconds(), nil, 1, "x:5|c|@1.0000\n"},
		{"x", metrics.OpGaugeUpdate, 123, nil, 1, "x:123|g|@1.0000\n"},
		{"x", metrics.OpGaugeUpdate, 1.23, nil, 1, "x:1.23|g|@1.0000\n"},
		{"x", metrics.OpGaugeUpdate, -123, nil, 1, "x:-123|g|@1.0000\n"},
		{"x", metrics.OpGaugeUpdate, -123, nil, 0.250, "x:-123|g|@0.2500\n"},
		{"x", metrics.OpHistogramUpdate, 123, nil, 0.250, "x:123|h|@0.2500\n"},
		{"x", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "x:123.456|h|@0.2500\n"},
		{"abc_xyz.sp.com", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "abc_xyz.sp.com:123.456|h|@0.2500\n"},

		{"event title", metrics.OpEventSend, "event text", nil, 1, "_e{11,10}:event title|event text\n"},
		{"event title", metrics.OpEventSend, "event text", []metrics.Tag{{Key: "x", Value: "1"}, {Key: "y", Value: 2}, {Key: "z", Value: "value"}}, 1, "_e{11,10}:event title|event text\n"},

		{"timer", metrics.OpTimerStop, 123.321, nil, 1, "timer:123.321|ms|@1.0000\n"},
	}

	for _, test := range tests {
		out, err := metrics.LibratoStatsDEncoder(test.name, test.op, test.value, test.tags, test.rate)
		assert.NoError(t, err)
		assert.Equal(t, test.out, out)
	}
}

func TestNamespacedEncoder(t *testing.T) {
	ne := metrics.NamespacedEncoder(metrics.StatsDEncoder, "test_namespace")
	out, err := ne("test", metrics.OpCounterAdd, 123, nil, 1)

	assert.NoError(t, err)
	assert.Equal(t, "test_namespace.test:123|c|@1.0000|#\n", out)
}
