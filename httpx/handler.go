package httpx

import (
	"net/http"
)

// NoopHandler returns a ContextAwareHandler that does nothing when it receives a request
func NoopHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

// StatusHandler returns a ContextAwareHandler handler that replies with the give status code.
func StatusHandler(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}

// StatusOKHandler is a simple ContextAwareHandler that always reply with status 200,
// useful to be used for health checker handlers or tests.
var StatusOKHandler = StatusHandler(http.StatusOK)
