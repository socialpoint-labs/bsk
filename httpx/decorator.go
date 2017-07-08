package httpx

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Decorator wraps/decorate a http.Handler with additional functionality.
type Decorator func(http.Handler) http.Handler

// CloseNotifierDecorator returns a decorator that cancels the context when the client
// connection closes unexpectedly.
//
// The http.CloseNotifier interface is implemented by http.ResponseWriters which
// allows detecting when the underlying connection has gone away.
//
// This mechanism can be used to cancel long operations on the server
// if the client has disconnected before the response is ready.
func CloseNotifierDecorator() Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Cancel the context if the client closes the connection
			if wcn, ok := w.(http.CloseNotifier); ok {
				ctx, cancel := context.WithCancel(r.Context())
				r = r.WithContext(ctx)

				// Canceling this context releases resources associated with it, so code should
				// call cancel as soon as the operations running in this Context complete.
				// As we have created a new context here, we are responsible for releasing it's resources
				// by calling the cancellation function.
				defer cancel()

				// CloseNotify returns a channel that receives at most a single value when
				// the client connection has gone away.

				// Runs a go-routine to receive this notification and cancel the request context
				go func() {
					select {
					case <-wcn.CloseNotify():
						cancel()
					case <-ctx.Done():
					}
				}()
			}

			h.ServeHTTP(w, r)
		})
	}
}

// AddHeaderDecorator returns a decorator that adds the given header to the HTTP response.
func AddHeaderDecorator(key, value string) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(key, value)
			h.ServeHTTP(w, r)
		})
	}
}

// SetHeaderDecorator returns a decorator that sets the given header to the HTTP response.
func SetHeaderDecorator(key, value string) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, value)
			h.ServeHTTP(w, r)
		})
	}
}

// CheckHeaderDecorator returns a decorator that checks if the given request header
// matches the given value, if the header does not exist or doesn't match then
// respond with the provided status code header and its value as content.
func CheckHeaderDecorator(headerName, headerValue string, statusCode int) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := r.Header.Get(headerName)
			if value != headerValue {
				w.WriteHeader(statusCode)
				// we don't care about the error if we can't write
				_, _ = w.Write([]byte(http.StatusText(statusCode)))
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// CheckSPAuthHeaderDecorator checks for the sign header with the provided secret,
// if doesn't match will return HTTP 401.
func CheckSPAuthHeaderDecorator(secret string) Decorator {
	return CheckHeaderDecorator("X-SP-Sign", secret, http.StatusUnauthorized)
}

// RootDecorator decorates a handler to distinguish root path from 404s
// ServeMux matches "/" for both, root path and all unmatched URLs
// How to bypass: https://golang.org/pkg/net/http/#example_ServeMux_Handle
func RootDecorator() Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// StripPrefixDecorator removes prefix from URL.
func StripPrefixDecorator(prefix string) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
				r.URL.Path = p
				h.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		})
	}
}

// EnableCORSDecorator adds required response headers to enable CORS and serves OPTIONS requests.
func EnableCORSDecorator() Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin,Accept,Content-Type,Authorization")

			// Stop here if its Preflighted OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// Condition represents a condition based on the http.Request and the current state of the http.ResponseWriter
type Condition func(w http.ResponseWriter, r *http.Request) bool

// IfDecorator is a special adapter that will skip to the 'then' handler if a condition
// applies at runtime, or pass the control to the adapted handler otherwise.
func IfDecorator(cond Condition, then http.Handler) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cond(w, r) {
				then.ServeHTTP(w, r)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

// TimeoutDecorator returns a adapter which adds a timeout to the context.
// Child handlers have the responsibility to obey the context deadline
func TimeoutDecorator(timeout time.Duration) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}

// LoggingDecorator returns an adapter that log requests to a given logger
func LoggingDecorator(logWriter io.Writer) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			resLogger := &responseLogger{w, 0, 0}
			h.ServeHTTP(resLogger, req)
			fmt.Fprintln(logWriter, formatLogLine(req, time.Now(), resLogger.Status(), resLogger.Size()))
		})
	}
}

func formatLogLine(req *http.Request, ts time.Time, status int, size int) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.URL.RequestURI()
	formattedTime := ts.Format("02/Jan/2006:15:04:05 -0700")

	return fmt.Sprintf("%s [%s] %s %s %s %d %d", host, formattedTime, req.Method, uri, req.Proto, status, size)
}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP
// status code and body size
type responseLogger struct {
	http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	size, err := l.ResponseWriter.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.ResponseWriter.WriteHeader(s)
	l.status = s
}

func (l responseLogger) Status() int {
	return l.status
}

func (l responseLogger) Size() int {
	return l.size
}
