package metrics

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"context"
)

const (
	bufferSize     = 1024
	datadogAddress = "127.0.0.1:8125"
	// this is datadog's agent default flush time, in case we lower it in the agent's conf change it here also
	datadogFlush = FlushEvery15s
)

// Publisher is a Metrics implementation that watches metrics changes and publish encoded
// metrics to and io.Writer. This allows to forward metrics to an UDP server using
// the StatsD protocol
type Publisher struct {
	writer        io.Writer
	encoder       Encoder
	errorHandler  ErrorHandler
	flushInterval time.Duration

	queue      chan string
	forceFlush chan struct{}
}

// NewPublisher creates a new metrics publisher
func NewPublisher(w io.Writer, e Encoder, flushInterval time.Duration, errorHandler ErrorHandler) *Publisher {
	if errorHandler == nil {
		errorHandler = DiscardErrors
	}

	return &Publisher{
		queue:      make(chan string),
		forceFlush: make(chan struct{}),

		writer:        w,
		encoder:       e,
		flushInterval: flushInterval,
		errorHandler:  errorHandler,
	}
}

// NewDiscardAll returns a concrete publisher instance that discards all
// metrics, useful to be used as a testing dummy/stub or when you don't care
// that all reported metrics get discarded
func NewDiscardAll() *Publisher {
	return NewPublisher(ioutil.Discard, StatsDEncoder, FlushEvery15s, DiscardErrors)
}

// NewStdout returns a publisher that sends the metrics to stdout.
func NewStdout(flushEvery time.Duration, errorHandler ErrorHandler) *Publisher {
	return NewPublisher(os.Stdout, StdoutEncoder, flushEvery, errorHandler)
}

// NewDataDog returns a publisher that sends the metrics to the datadog agent.
func NewDataDog() *Publisher {
	addr, err := net.ResolveUDPAddr("udp", datadogAddress)
	if err != nil {
		panic(fmt.Sprintf("cannot resolve UDP addr `%s`: `%s`", datadogAddress, err))
	}

	client, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(fmt.Sprintf("cannot create UDP client: `%s`", err.Error()))
	}

	return NewPublisher(client, StatsDEncoder, datadogFlush, nil)
}

// Counter returns a new counter with the provided name and tags
func (p *Publisher) Counter(name string, tags ...Tag) Counter {
	return &publisherCounter{publisherMetric{name: name, tags: tags, nf: p.notify}}
}

// Gauge returns a new Gauge with the provided name and tags
func (p *Publisher) Gauge(name string, tags ...Tag) Gauge {
	return &publisherGauge{publisherMetric{name: name, tags: tags, nf: p.notify}}
}

// Event returns a new Event with the provided title and tags
func (p *Publisher) Event(title string, tags ...Tag) Event {
	return &publisherEvent{publisherMetric{name: title, tags: tags, nf: p.notify}}
}

// Timer returns a new Timer with the provided title and tags
func (p *Publisher) Timer(title string, tags ...Tag) Timer {
	return &timerEvent{publisherMetric: publisherMetric{name: title, tags: tags, nf: p.notify}}
}

// Flush forces the flush of the publisher
func (p *Publisher) Flush() {
	p.forceFlush <- struct{}{}
}

func (p *Publisher) notify(op Op, name string, value interface{}, tags Tags) {
	code, err := p.encoder(name, op, value, tags, 1)
	if err != nil {
		p.errorHandler(err)
	}

	p.queue <- code
}

// Run makes the publisher a server.Runner
func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	buf := &bytes.Buffer{}
	defer p.flush(buf)

	for {
		select {
		case cmd := <-p.queue:
			// we don't care if errors, this is fire and forget
			_, _ = buf.WriteString(cmd)

			if buf.Len() >= bufferSize {
				p.flush(buf)
			}

		case <-ticker.C:
			p.flush(buf)

		case <-p.forceFlush:
			p.flush(buf)

		case <-ctx.Done():
			return
		}
	}
}

func (p *Publisher) flush(w io.WriterTo) {
	_, err := w.WriteTo(p.writer)
	if err != nil {
		p.errorHandler(err)
	}
}

// publisherMetric is the parent struct with the common fields
// fields and methods for the rest of metrics.
type publisherMetric struct {
	name string
	tags Tags
	nf   NotifyFunc
}

func (m publisherMetric) Name() string {
	return m.name
}

func (m publisherMetric) Tags() Tags {
	return m.tags
}

type publisherCounter struct {
	publisherMetric
}

func (c publisherCounter) Add(delta uint64) {
	c.nf(OpCounterAdd, c.name, delta, c.tags)
}

func (c publisherCounter) Inc() {
	c.Add(1)
}

func (c publisherCounter) WithTags(tags ...Tag) Counter {
	for _, tag := range tags {
		c.tags = append(c.tags, tag)
	}
	return c
}

func (c publisherCounter) WithTag(key string, value interface{}) Counter {
	c.tags = append(c.tags, NewTag(key, value))
	return c
}

type publisherGauge struct {
	publisherMetric
}

func (g publisherGauge) Update(value interface{}) {
	g.nf(OpGaugeUpdate, g.name, value, g.tags)
}

func (g publisherGauge) WithTags(tags ...Tag) Gauge {
	for _, tag := range tags {
		g.tags = append(g.tags, tag)
	}
	return g
}

func (g publisherGauge) WithTag(key string, value interface{}) Gauge {
	g.tags = append(g.tags, NewTag(key, value))
	return g
}

type publisherEvent struct {
	publisherMetric
}

func (e publisherEvent) Send() {
	e.nf(OpEventSend, e.name, "", e.tags)
}

func (e publisherEvent) SendWithText(text string) {
	e.nf(OpEventSend, e.name, text, e.tags)
}

func (e publisherEvent) WithTags(tags ...Tag) Event {
	for _, tag := range tags {
		e.tags = append(e.tags, tag)
	}
	return e
}

func (e publisherEvent) WithTag(key string, value interface{}) Event {
	e.tags = append(e.tags, NewTag(key, value))
	return e
}

type timerEvent struct {
	publisherMetric
	startedTime time.Time
}

func (e *timerEvent) Start() {
	e.startedTime = time.Now()
}

func (e *timerEvent) Stop() {
	if !e.startedTime.IsZero() {
		durationInMs := float64(time.Since(e.startedTime).Nanoseconds()) * 1e-6
		e.nf(OpTimerStop, e.name, durationInMs, e.tags)
		e.startedTime = time.Time{}
	}
}

func (e *timerEvent) WithTags(tags ...Tag) Timer {
	for _, tag := range tags {
		e.tags = append(e.tags, tag)
	}
	return e
}

func (e *timerEvent) WithTag(key string, value interface{}) Timer {
	e.tags = append(e.tags, NewTag(key, value))
	return e
}
