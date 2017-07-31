package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsRecorderRegistry(t *testing.T) {
	assert := assert.New(t)

	r := NewRecorder()

	c := r.Counter("counter")
	assert.Equal(c, r.Get("counter"))

	g := r.Gauge("gauge")
	assert.Equal(g, r.Get("gauge"))

	timer := r.Timer("timer")
	assert.Equal(timer, r.Get("timer"))

	assert.Nil(r.Get("does-not-exists"))
}

func TestMetricsRecorder(t *testing.T) {
	assert := assert.New(t)

	moreTags := Tags{
		NewTag("moretag1", "value1"),
		NewTag("moretag2", "value2"),
	}

	lastTagKey := "lastTagKey"
	lastTagValue := "lastTagValue"
	lastTag := NewTag(lastTagKey, lastTagValue)

	for _, tags := range []Tags{
		{},
		{Tag{Key: "foo", Value: "bar"}},
		{Tag{Key: "foo", Value: "bar"}, Tag{Key: "foo2", Value: "bar2"}},
		{NewTag("foo", "bar"), NewTag("foo2", "bar2")},
		{NewTag("foo", "bar"), NewTag("foo2", "bar2")},
	} {
		r := NewRecorder()

		// test counter inc
		metricName := "counter"
		c := r.Counter(metricName, tags...).(*RecorderCounter)
		c.Inc()
		c.Inc()

		assert.EqualValues(2, c.Value)

		// test counter tags
		assert.Equal(c.Tags(), tags)
		c.WithTags(moreTags...) // another way to set tags

		c.Inc()
		assert.EqualValues(3, c.Value)

		assert.Equal(append(tags, moreTags...), c.Tags())

		// test counter add from inc
		c.Add(10)
		assert.EqualValues(13, c.Value)

		// test counter add

		c = r.Counter(metricName, tags...).(*RecorderCounter)
		c.WithTags(tags...)
		c.Add(10)
		assert.EqualValues(23, c.Value)

		// test gauge
		metricName = "gauge"
		g := r.Gauge(metricName, tags...).(*RecorderGauge)
		g.Update(math.Pi)
		assert.Equal(math.Pi, g.Value)
		g.Update(math.E)
		assert.EqualValues(math.E, g.Value)

		// test gauge tags
		assert.Equal(g.Tags(), tags)
		g.WithTags(moreTags...) // another way to set tags
		g.Update(math.Ln2)
		assert.EqualValues(math.Ln2, g.Value)
		assert.EqualValues(g.Tags(), append(tags, moreTags...))
		g.WithTag(lastTagKey, lastTagValue) // and another way to add one tag
		assert.Equal(g.Tags(), append(append(tags, moreTags...), lastTag))

		// test event
		metricName = "event"
		e := r.Event(metricName, tags...).(*RecorderEvent)
		e.Send()
		assert.Equal("event|", e.Event)
		e.SendWithText("msg")
		assert.Equal("event|msg", e.Event)

		// test event tags
		assert.Equal(e.Tags(), tags)
		e.WithTags(moreTags...) // another way to set tags
		e.SendWithText("msg2")
		assert.Equal("event|msg2", e.Event)
		assert.Equal(append(tags, moreTags...), e.Tags())
		e.WithTag(lastTagKey, lastTagValue) // and another way to add one tag
		assert.Equal(e.Tags(), append(append(tags, moreTags...), lastTag))

		// test Timer
		metricName = "timer"
		t := r.Timer(metricName, tags...)
		t.Start()
		t.Stop()

		// test timer tags
		assert.Equal(t.Tags(), tags)
		t.WithTags(moreTags...) // another way to set tags
		assert.Equal(t.Tags(), append(tags, moreTags...))
		t.WithTag(lastTagKey, lastTagValue) // and another way to add one tag
		assert.Equal(t.Tags(), append(append(tags, moreTags...), lastTag))
	}
}

func TestRecorder_ConcurrentSafety(t *testing.T) {
	assert := assert.New(t)
	r := NewRecorder()

	ch := make(chan bool)

	// Register several types of metrics
	r.Counter("counter")
	r.Gauge("gauge")
	r.Timer("timer")
	r.Event("event")

	thread := func() {
		c := r.Get("counter").(*RecorderCounter)
		c.Inc()

		g := r.Get("gauge").(*RecorderGauge)
		g.Update(123)

		timer := r.Get("timer").(*RecorderTimer)
		timer.Start()
		timer.Stop()

		e := r.Get("event").(*RecorderEvent)
		e.SendWithText("life")

		ch <- true
	}

	c := r.Get("counter").(*RecorderCounter)
	g := r.Get("gauge").(*RecorderGauge)
	timer := r.Get("timer").(*RecorderTimer)

	go thread()
	go thread()

	<-ch
	<-ch

	assert.EqualValues(2, c.Value)
	assert.EqualValues(123, g.Value)
	assert.WithinDuration(timer.StartedTime, timer.StoppedTime, time.Duration(time.Millisecond))
}
