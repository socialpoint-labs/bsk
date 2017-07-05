// Package httpc provides a set of features for easy extension of the default `net/http` client.
// The idea is pretty simple, start with an http.DefaultClient and decorate it with the
// extra functionality you need.
package httpc

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// A Client sends http.Requests and returns http.Responses or errors in
// case of failure.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// ClientFunc is a function type that implements the Client interface.
type ClientFunc func(*http.Request) (*http.Response, error)

// Do sends an HTTP request and returns an HTTP response or an error.
// When err is nil, resp always contains a non-nil resp.Body.
//
// Callers should close resp.Body when done reading from it. If
// resp.Body is not closed, the Client's underlying RoundTripper
// (typically Transport) may not be able to re-use a persistent TCP
// connection to the server for a subsequent "keep-alive" request.
//
// The request Body, if non-nil, will be closed by the underlying
// Transport, even on errors.
//
// Generally Get, Post, or PostForm will be used instead of Do.
//
// See http.Client in the standard library for more details.
func (f ClientFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

// A Decorator wraps a Client with extra behaviour.
type Decorator func(Client) Client

// Decorate decorates a Client with the given decorators in reverse order.
func Decorate(c Client, ds ...Decorator) Client {
	decorated := c
	for i := len(ds) - 1; i >= 0; i-- {
		decorated = ds[i](decorated)
	}
	return decorated
}

// FaultTolerance returns a Decorator that extends a Client with fault tolerance
// configured with the given attempts and backoff duration.
func FaultTolerance(attempts int, backoff time.Duration) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (res *http.Response, err error) {
			for i := 0; i <= attempts; i++ {
				if res, err = c.Do(r); err == nil {
					break
				}
				time.Sleep(backoff * time.Duration(i))
			}
			return res, err
		})
	}
}

// Header returns a Decorator that adds the given HTTP header to every request
// done by a Client.
func Header(name, value string) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Add(name, value)
			return c.Do(r)
		})
	}
}

// Logger returns a Decorator that logs HTTP requests to an io.Writer
func Logger(w io.Writer) Decorator {
	return Loggerf(w, func(r *http.Request) string {
		return fmt.Sprintf("%s %s\n", r.Method, r.URL.String())
	})
}

// Loggerf returns a Decorator that logs HTTP requests to an io.Writer.
// The output can be customized by passing a LoggerFormatter.
func Loggerf(w io.Writer, f LoggerFormatter) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			_, err := w.Write([]byte(f(r)))
			if err != nil {
				return nil, err
			}

			return c.Do(r)
		})
	}
}

// FakeResponse mocks a response content and status code
type FakeResponse struct {
	Content    string
	StatusCode int
}

// NewFake creates a NewFakeResponse, to avoid the "composite literal uses unkeyed fields" govet check.
func NewFake(content string, statusCode int) FakeResponse {
	return FakeResponse{content, statusCode}
}

// Fake returns a client Decorator making it respond with the provided status
// code, but with every call it will return a different content with the
// provided fakeResponses.
func Fake(fakeResponses ...FakeResponse) Decorator {
	return func(c Client) Client {
		mu := &sync.Mutex{}
		times := 0
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			mu.Lock()
			defer mu.Unlock()
			if len(fakeResponses) == 0 || times > (len(fakeResponses)-1) {
				panic("not enough fake responses")
			}
			resp := &http.Response{
				StatusCode: fakeResponses[times].StatusCode,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(fakeResponses[times].Content))),
			}
			times++
			return resp, nil
		})
	}
}

// Query returns a Decorator that adds the given HTTP query to every request
// done by a Client.
func Query(name string, value string) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			q := r.URL.Query()
			q.Add(name, value)
			r.URL.RawQuery = q.Encode()
			return c.Do(r)
		})
	}
}

// LoggerFormatter formats a request to be logged
type LoggerFormatter func(r *http.Request) string
