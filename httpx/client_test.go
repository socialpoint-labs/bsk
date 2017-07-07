package httpx_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/httpx"
	"github.com/stretchr/testify/assert"
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

			body, err := ioutil.ReadAll(resp.Body)
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

			body, err := ioutil.ReadAll(resp.Body)
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
