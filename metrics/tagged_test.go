package metrics_test

import (
	"testing"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewTaggedMetrics(t *testing.T) {
	recorder := metrics.NewRecorder()

	m := metrics.NewTaggedMetrics(
		recorder,
		metrics.NewTag("foo", "bar"),
	)

	m.Counter("test", metrics.NewTag("su", "pu")).Inc()

	tags := recorder.Get("test").Tags()

	assert.Len(t, tags, 2)
	assert.Equal(t, tags[0].Key, "foo")
	assert.Equal(t, tags[1].Key, "su")
}
