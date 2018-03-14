package metrics_test

import (
	"math"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestMetricsRecorderRegistry(t *testing.T) {
	a := assert.New(t)

	r := metrics.NewRecorder()

	c := r.Counter("counter")
	a.Equal(c, r.Get("counter"))

	g := r.Gauge("gauge")
	a.Equal(g, r.Get("gauge"))

	timer := r.Timer("timer")
	a.Equal(timer, r.Get("timer"))

	a.Nil(r.Get("does-not-exists"))
}

func TestMetricsRecorder(t *testing.T) {
	a := assert.New(t)

	moreTags := metrics.Tags{
		metrics.NewTag("moretag1", "value1"),
		metrics.NewTag("moretag2", "value2"),
	}

	lastTagKey := "lastTagKey"
	lastTagValue := "lastTagValue"
	lastTag := metrics.NewTag(lastTagKey, lastTagValue)

	for _, tags := range []metrics.Tags{
		{},
		{metrics.Tag{Key: "foo", Value: "bar"}},
		{metrics.Tag{Key: "foo", Value: "bar"}, metrics.Tag{Key: "foo2", Value: "bar2"}},
		{metrics.NewTag("foo", "bar"), metrics.NewTag("foo2", "bar2")},
		{metrics.NewTag("foo", "bar"), metrics.NewTag("foo2", "bar2")},
	} {
		r := metrics.NewRecorder()

		// test counter inc
		metricName := "counter"
		c := r.Counter(metricName, tags...).(*metrics.RecorderCounter)
		c.Inc()
		c.Inc()

		a.EqualValues(2, c.Value)

		// test counter tags
		a.Equal(c.Tags(), tags)
		c.WithTags(moreTags...) // another way to set tags

		c.Inc()
		a.EqualValues(3, c.Value)

		a.Equal(append(tags, moreTags...), c.Tags())

		// test counter add from inc
		c.Add(10)
		a.EqualValues(13, c.Value)

		// test counter add

		c = r.Counter(metricName, tags...).(*metrics.RecorderCounter)
		c.WithTags(tags...)
		c.Add(10)
		a.EqualValues(23, c.Value)

		// test gauge
		metricName = "gauge"
		g := r.Gauge(metricName, tags...).(*metrics.RecorderGauge)
		g.Update(math.Pi)
		a.Equal(math.Pi, g.Value)
		g.Update(math.E)
		a.EqualValues(math.E, g.Value)

		// test gauge tags
		a.Equal(g.Tags(), tags)
		g.WithTags(moreTags...) // another way to set tags
		g.Update(math.Ln2)
		a.EqualValues(math.Ln2, g.Value)
		a.EqualValues(g.Tags(), append(tags, moreTags...))
		g.WithTag(lastTagKey, lastTagValue) // and another way to add one tag
		a.Equal(g.Tags(), append(append(tags, moreTags...), lastTag))

		// test event
		metricName = "event"
		e := r.Event(metricName, tags...).(*metrics.RecorderEvent)
		e.Send()
		a.Equal("event|", e.Event)
		e.SendWithText("msg")
		a.Equal("event|msg", e.Event)

		// test event tags
		a.Equal(e.Tags(), tags)
		e.WithTags(moreTags...) // another way to set tags
		e.SendWithText("msg2")
		a.Equal("event|msg2", e.Event)
		a.Equal(append(tags, moreTags...), e.Tags())
		e.WithTag(lastTagKey, lastTagValue) // and another way to add one tag
		a.Equal(e.Tags(), append(append(tags, moreTags...), lastTag))

		// test Timer
		metricName = "timer"
		t := r.Timer(metricName, tags...)
		t.Start()
		t.Stop()

		// test timer tags
		a.Equal(t.Tags(), tags)
		t.WithTags(moreTags...) // another way to set tags
		a.Equal(t.Tags(), append(tags, moreTags...))
		t.WithTag(lastTagKey, lastTagValue) // and another way to add one tag
		a.Equal(t.Tags(), append(append(tags, moreTags...), lastTag))
	}
}

func TestRecorder_ConcurrentSafety(t *testing.T) {
	a := assert.New(t)
	r := metrics.NewRecorder()

	ch := make(chan bool)

	// Register several types of metrics
	r.Counter("counter")
	r.Gauge("gauge")
	r.Timer("timer")
	r.Event("event")

	thread := func() {
		c := r.Get("counter").(*metrics.RecorderCounter)
		c.Inc()

		g := r.Get("gauge").(*metrics.RecorderGauge)
		g.Update(123)

		timer := r.Get("timer").(*metrics.RecorderTimer)
		timer.Start()
		timer.Stop()

		e := r.Get("event").(*metrics.RecorderEvent)
		e.SendWithText("life")

		ch <- true
	}

	c := r.Get("counter").(*metrics.RecorderCounter)
	g := r.Get("gauge").(*metrics.RecorderGauge)
	timer := r.Get("timer").(*metrics.RecorderTimer)

	go thread()
	go thread()

	<-ch
	<-ch

	a.EqualValues(2, c.Value)
	a.EqualValues(123, g.Value)
	a.WithinDuration(timer.StartedTime, timer.StoppedTime, time.Duration(time.Millisecond))
}
