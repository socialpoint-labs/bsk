package metrics

import (
	"sync"
	"time"
)

// Recorder is a Metrics implementation that will hold the values written by
// its metrics types for testing purposes only.
type Recorder struct {
	registry map[string]Metric
	mu       sync.RWMutex // protects the registry
}

// NewRecorder returns an empty Recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		registry: make(map[string]Metric),
	}
}

// A RecorderMetric is the type that will implement
// the Counter, Gauge and Event metric types.
type RecorderMetric struct {
	name string
	tags Tags
}

// Name implements part of the Metric interface.
func (rm RecorderMetric) Name() string {
	return rm.name
}

// Tags implements part of the Metric interface.
func (rm RecorderMetric) Tags() Tags {
	return rm.tags
}

// A RecorderCounter is a RecorderMetric that implements Counter.
type RecorderCounter struct {
	RecorderMetric
	Value uint64
	mu    sync.Mutex // protects the whole struct
}

// Val returns the counter value in a thread-safe manner
func (c *RecorderCounter) Val() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Value
}

// Inc implements the Counter behaviour and stores the value in the Recorder.
func (c *RecorderCounter) Inc() {
	c.Add(1)
}

// Add implements the Counter behaviour and stores the value in the Recorder.
func (c *RecorderCounter) Add(delta uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Value += delta
}

// WithTags adds the passed tags to the Tags recorder map.
func (c *RecorderCounter) WithTags(tags ...Tag) Counter {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tags = append(c.tags, tags...)
	return c
}

// WithTag creates a new tag with the parameters and adds it to the Tags recorder map.
func (c *RecorderCounter) WithTag(key string, value interface{}) Counter {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tags = append(c.tags, NewTag(key, value))
	return c
}

// A RecorderGauge is a RecorderMetric that implements Gauge.
type RecorderGauge struct {
	RecorderMetric
	Value interface{}
	mu    sync.Mutex // protects the whole struct
}

// Update implements the Gauge behaviour and stores the value in the Recorder.
func (g *RecorderGauge) Update(value interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Value = value
}

// WithTags adds the passed tags to the Tags recorder map.
func (g *RecorderGauge) WithTags(tags ...Tag) Gauge {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.tags = append(g.tags, tags...)

	return g
}

// WithTag creates a new tag with the parameters and adds it to the Tags recorder map.
func (g *RecorderGauge) WithTag(key string, value interface{}) Gauge {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.tags = append(g.tags, NewTag(key, value))

	return g
}

// A RecorderEvent is a RecorderMetric that implements Event.
type RecorderEvent struct {
	RecorderMetric
	Event string

	mu sync.Mutex // protects the whole struct
}

// Send implements the Event behaviour.
func (e *RecorderEvent) Send() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Event = e.name + "|"
}

// SendWithText implements the Event behaviour and stores the
// event text in the Recorder.
func (e *RecorderEvent) SendWithText(text string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Event = e.name + "|" + text
}

// WithTags adds the passed tags to the Tags recorder map.
func (e *RecorderEvent) WithTags(tags ...Tag) Event {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tags = append(e.tags, tags...)
	return e
}

// WithTag creates a new tag with the parameters and adds it to the Tags recorder map.
func (e *RecorderEvent) WithTag(key string, value interface{}) Event {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tags = append(e.tags, NewTag(key, value))
	return e
}

// A RecorderTimer is a RecorderMetric that implements Timer.
type RecorderTimer struct {
	RecorderMetric
	StartedTime time.Time
	StoppedTime time.Time
	mu          sync.Mutex // protects the whole struct
}

// Start the timer.
func (t *RecorderTimer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.StartedTime = time.Now()
}

// Stop the timer.
func (t *RecorderTimer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.StoppedTime = time.Now()
}

// WithTags adds the passed tags to the Tags recorder map.
func (t *RecorderTimer) WithTags(tags ...Tag) Timer {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tag := range tags {
		t.tags = append(t.tags, tag)
	}
	return t
}

// WithTag creates a new tag with the parameters and adds it to the Tags recorder map.
func (t *RecorderTimer) WithTag(key string, value interface{}) Timer {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tags = append(t.tags, NewTag(key, value))
	return t
}

// Counter implements the Metrics behaviour to return a new Counter.
func (r *Recorder) Counter(name string, tags ...Tag) Counter {
	m := r.Get(name)
	if m == nil {
		m = &RecorderCounter{RecorderMetric: RecorderMetric{name, tags}}
		r.register(name, m)
	}

	return m.(Counter)
}

// Gauge implements the Metrics behaviour to return a new Gauge.
func (r *Recorder) Gauge(name string, tags ...Tag) Gauge {
	m := &RecorderGauge{RecorderMetric: RecorderMetric{name, tags}}
	r.register(name, m)
	return m
}

// Event implements the Metrics behaviour to return a new Event.
func (r *Recorder) Event(name string, tags ...Tag) Event {
	m := &RecorderEvent{RecorderMetric: RecorderMetric{name, tags}}
	r.register(name, m)
	return m
}

// Timer implements the Metrics behaviour to return a new Timer.
func (r *Recorder) Timer(name string, tags ...Tag) Timer {
	m := &RecorderTimer{RecorderMetric: RecorderMetric{name, tags}}
	r.register(name, m)
	return m
}

// Get returns the metric instance registered with the given name
func (r *Recorder) Get(name string) Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.registry[name]
}

func (r *Recorder) register(name string, metric Metric) {
	r.mu.Lock()
	r.registry[name] = metric
	r.mu.Unlock()
}

// HasTag return whether the given metric is tagged with the given key/value pair.
func HasTag(m Metric, key string, value interface{}) bool {
	for _, t := range m.Tags() {
		if t.Key == key && t.Value == value {
			return true
		}
	}

	return false
}
