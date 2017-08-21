package metrics_test

import (
	"fmt"
	"testing"

	"time"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestStatsDEncoder(t *testing.T) {
	assert := assert.New(t)

	for i, tc := range []struct {
		encoder metrics.Encoder
		name    string
		op      metrics.Op
		value   interface{}
		tags    metrics.Tags
		rate    float64
		out     string
	}{
		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, 123, nil, 1, "x:123|c|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, time.Second, nil, 1, "x:1s|c|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, time.Millisecond, nil, 1, "x:1ms|c|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, time.Nanosecond * 5, nil, 1, "x:5ns|c|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, (time.Nanosecond * 5).Nanoseconds(), nil, 1, "x:5|c|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpGaugeUpdate, 123, nil, 1, "x:123|g|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpGaugeUpdate, 1.23, nil, 1, "x:1.23|g|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpGaugeUpdate, -123, nil, 1, "x:-123|g|@1.0000|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpGaugeUpdate, -123, nil, 0.250, "x:-123|g|@0.2500|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpHistogramUpdate, 123, nil, 0.250, "x:123|h|@0.2500|#\n"},
		{metrics.StatsDEncoder, "x", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "x:123.456|h|@0.2500|#\n"},
		{metrics.StatsDEncoder, "abc_xyz.sp.com", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "abc_xyz.sp.com:123.456|h|@0.2500|#\n"},

		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, 123, []metrics.Tag{metrics.NewTag("x", "1")}, 1, "x:123|c|@1.0000|#x:1\n"},
		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, 123, []metrics.Tag{metrics.NewTag("x", "1"), metrics.NewTag("y", 2)}, 1, "x:123|c|@1.0000|#x:1,y:2\n"},
		{metrics.StatsDEncoder, "x", metrics.OpCounterAdd, 123, []metrics.Tag{metrics.NewTag("x", "1"), metrics.NewTag("y", 2), metrics.NewTag("z", "value")}, 1, "x:123|c|@1.0000|#x:1,y:2,z:value\n"},

		{metrics.StatsDEncoder, "event title", metrics.OpEventSend, "event text", nil, 1, "_e{11,10}:event title|event text|#\n"},
		{metrics.StatsDEncoder, "event title", metrics.OpEventSend, "event text", []metrics.Tag{metrics.NewTag("x", "1"), metrics.NewTag("y", 2), metrics.NewTag("z", "value")}, 1, "_e{11,10}:event title|event text|#x:1,y:2,z:value\n"},

		{metrics.StatsDEncoder, "timer", metrics.OpTimerStop, 123.321, nil, 1, "timer:123.321|ms|@1.0000|#\n"},

		{metrics.LibratoStatsDEncoder, "x", metrics.OpCounterAdd, 123, nil, 1, "x:123|c|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpCounterAdd, time.Second, nil, 1, "x:1s|c|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpCounterAdd, time.Millisecond, nil, 1, "x:1ms|c|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpCounterAdd, time.Nanosecond * 5, nil, 1, "x:5ns|c|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpCounterAdd, (time.Nanosecond * 5).Nanoseconds(), nil, 1, "x:5|c|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpGaugeUpdate, 123, nil, 1, "x:123|g|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpGaugeUpdate, 1.23, nil, 1, "x:1.23|g|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpGaugeUpdate, -123, nil, 1, "x:-123|g|@1.0000\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpGaugeUpdate, -123, nil, 0.250, "x:-123|g|@0.2500\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpHistogramUpdate, 123, nil, 0.250, "x:123|h|@0.2500\n"},
		{metrics.LibratoStatsDEncoder, "x", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "x:123.456|h|@0.2500\n"},
		{metrics.LibratoStatsDEncoder, "abc_xyz.sp.com", metrics.OpHistogramUpdate, 123.456, nil, 0.250, "abc_xyz.sp.com:123.456|h|@0.2500\n"},

		{metrics.LibratoStatsDEncoder, "event title", metrics.OpEventSend, "event text", nil, 1, "_e{11,10}:event title|event text\n"},
		{metrics.LibratoStatsDEncoder, "event title", metrics.OpEventSend, "event text", []metrics.Tag{metrics.NewTag("x", "1"), metrics.NewTag("y", 2), metrics.NewTag("z", "value")}, 1, "_e{11,10}:event title|event text\n"},

		{metrics.LibratoStatsDEncoder, "timer", metrics.OpTimerStop, 123.321, nil, 1, "timer:123.321|ms|@1.0000\n"},
	} {
		tc := tc
		t.Run(fmt.Sprintf("subtest %d", i), func(t *testing.T) {
			t.Parallel()
			out, err := tc.encoder(tc.name, tc.op, tc.value, tc.tags, tc.rate)
			assert.NoError(err)
			assert.Equal(tc.out, out)
		})
	}
}

func TestNamespacedEncoder(t *testing.T) {
	assert := assert.New(t)

	ne := metrics.NamespacedEncoder(metrics.StatsDEncoder, "test_namespace")
	out, err := ne("test", metrics.OpCounterAdd, 123, nil, 1)

	assert.NoError(err)
	assert.Equal("test_namespace.test:123|c|@1.0000|#\n", out)
}
