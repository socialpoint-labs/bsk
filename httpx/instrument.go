package httpx

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/socialpoint-labs/bsk/metrics"
)

// InstrumentDecorator returns an adapter that instrument requests with some metrics:
// - http.request_duration: requests duration
// - http.requests: number of requests
//
// Metrics are tagged with the HTTP method, requests path, response status code and response status class.
func InstrumentDecorator(met metrics.Metrics, t ...metrics.Tag) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timer := met.Timer("http.request_duration")
			timer.Start()

			delegate := &responseWriterDelegator{ResponseWriter: w}

			h.ServeHTTP(delegate, r)

			code := delegate.status

			tags := append(t,
				metrics.Tag{Key: "method", Value: strings.ToLower(r.Method)},
				metrics.Tag{Key: "path", Value: r.URL.EscapedPath()},
				metrics.Tag{Key: "code", Value: code},
				metrics.Tag{Key: "class", Value: httpStatusCodeClass(code)},
			)

			timer.WithTags(tags...).Stop()
		})
	}
}

// responseWriterDelegator is an implementation of a http.ResponseWriter that keeps track of the HTTP status code
// written during the request/response lifecycle
type responseWriterDelegator struct {
	http.ResponseWriter
	status int
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func httpStatusCodeClass(code int) string {
	return strconv.Itoa(code/100) + "xx"
}
