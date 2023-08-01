package httpx_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/httpx"
	"github.com/socialpoint-labs/bsk/metrics"
)

const errorMsg = "randomError"

type NoopClient struct{}

func (c *NoopClient) Do(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

type FailingClient struct {
	attempts int
}

func (c *FailingClient) Do(*http.Request) (*http.Response, error) {
	defer func() {
		c.attempts++
	}()

	// This makes the client fails in the first 2 attempts
	if c.attempts < 2 {
		return nil, errors.New(errorMsg)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

type MetricsInstrument struct {
	metrics.Metrics
	timerMock metrics.Timer
}

func (mi *MetricsInstrument) Timer(name string, tags ...metrics.Tag) metrics.Timer {
	return mi.timerMock
}

type TimerSpy struct {
	metrics.Timer
	tagMapMutex sync.RWMutex
	TagsMap     map[string]interface{}
	Started     bool
	Stopped     bool
}

func (tm *TimerSpy) WithTags(tags ...metrics.Tag) metrics.Timer {
	tm.tagMapMutex.Lock()
	defer tm.tagMapMutex.Unlock()
	for _, t := range tags {
		tm.TagsMap[t.Key] = t.Value
	}
	return tm
}

func (tm *TimerSpy) WithTag(key string, value interface{}) metrics.Timer {
	tm.tagMapMutex.Lock()
	defer tm.tagMapMutex.Unlock()
	tm.TagsMap[key] = value
	return tm
}

func (tm *TimerSpy) Tags() metrics.Tags {
	tm.tagMapMutex.RLock()
	defer tm.tagMapMutex.RUnlock()
	tags := metrics.Tags{}
	for key, value := range tm.TagsMap {
		tags = append(tags, metrics.Tag{Key: key, Value: value})
	}

	return tags
}

func (tm *TimerSpy) Start() {
	tm.Started = true
}
func (tm *TimerSpy) Stop() {
	tm.Stopped = true
}

func TestHeaderDecorator(t *testing.T) {
	assert := assert.New(t)

	client := httpx.DecorateClient(&NoopClient{}, httpx.Header("test", "123"))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("123", req.Header.Get("test"))
}

func TestFaultTolerance(t *testing.T) {
	assert := assert.New(t)

	client := httpx.DecorateClient(&FailingClient{}, httpx.FaultTolerance(5, time.Millisecond))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)

	assert.Equal(200, resp.StatusCode)
}

func TestInstrumentRequestDurationMetric(t *testing.T) {
	assert := assert.New(t)

	timer := &TimerSpy{TagsMap: make(map[string]interface{})}
	metricsInstrument := &MetricsInstrument{timerMock: timer}
	tag := "tag"
	value := "test"
	customTag := metrics.Tag{Key: tag, Value: value}
	client := httpx.DecorateClient(&NoopClient{}, httpx.InstrumentRequestDurationMetric(metricsInstrument, customTag))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)

	tags := timer.Tags()
	assert.Contains(tags, metrics.Tag{Key: "code", Value: http.StatusOK})
	assert.Contains(tags, metrics.Tag{Key: "method", Value: "get"})
	assert.Contains(tags, metrics.Tag{Key: tag, Value: value})
	assert.Equal(value, timer.TagsMap[tag])
	assert.True(timer.Started)
	assert.True(timer.Stopped)

	assert.Equal(200, resp.StatusCode)
}

func TestLogger(t *testing.T) {
	assert := assert.New(t)

	recorder := &bytes.Buffer{}

	client := httpx.DecorateClient(&NoopClient{}, httpx.Logger(recorder))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("GET http://example.com\n", recorder.String())
}

func TestLoggerf(t *testing.T) {
	assert := assert.New(t)

	recorder := &bytes.Buffer{}

	formatter := func(r *http.Request) string { return fmt.Sprintf("[%s][%s]", r.Method, r.URL.String()) }
	client := httpx.DecorateClient(&NoopClient{}, httpx.Loggerf(recorder, formatter))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("[GET][http://example.com]", recorder.String())
}

func TestFake(t *testing.T) {
	assert := assert.New(t)

	for _, fakes := range [][]httpx.FakeResponse{
		{httpx.NewFake("foo", http.StatusOK)},
		{httpx.NewFake("teapot", http.StatusTeapot)},
		// multiple/successive
		{httpx.NewFake("foo", http.StatusOK), httpx.NewFake("teapot", http.StatusTeapot)},
	} {
		client := httpx.DecorateClient(http.DefaultClient, httpx.Fake(fakes...))
		assert.NotNil(client)

		for _, fake := range fakes {
			resp, err := client.Do(&http.Request{})
			assert.NoError(err)
			assert.NotNil(resp)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(err)
			assert.Equal(fake.Content, string(body))
			assert.Equal(fake.StatusCode, resp.StatusCode)
		}
	}
}

func TestConcurrentFake(t *testing.T) {
	assert := assert.New(t)

	r := httpx.NewFake("teapot", http.StatusTeapot)
	client := httpx.DecorateClient(http.DefaultClient, httpx.Fake(r, r))
	assert.NotNil(client)

	wg := &sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			resp, err := client.Do(&http.Request{})
			assert.NoError(err)
			assert.NotNil(resp)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(err)
			assert.Equal("teapot", string(body))
			assert.Equal(http.StatusTeapot, resp.StatusCode)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestQueryDecorator(t *testing.T) {
	assert := assert.New(t)

	client := httpx.DecorateClient(&NoopClient{}, httpx.Query("test", "123"))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("123", req.URL.Query().Get("test"))
}
