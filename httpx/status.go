package httpx

import "net/http"

// ResponseIsSuccessful returns whether the response is successful
func ResponseIsSuccessful(resp *http.Response) bool {
	return IsSuccessful(resp.StatusCode)
}

// ResponseIsRedirection returns whether the response is a redirect
func ResponseIsRedirection(resp *http.Response) bool {
	return IsRedirection(resp.StatusCode)
}

// ResponseIsClientError returns whether the response is a client error
func ResponseIsClientError(resp *http.Response) bool {
	return IsClientError(resp.StatusCode)
}

// ResponseIsServerError returns whether the response is a server error
func ResponseIsServerError(resp *http.Response) bool {
	return IsServerError(resp.StatusCode)
}

// StatusFunc is a function that recives a status code and returns a boolean.
type StatusFunc func(code int) bool

// IsSuccessful returns whether the response is successful
func IsSuccessful(code int) bool {
	return code >= 200 && code < 300
}

// IsRedirection returns whether the codeonse is a redirect
func IsRedirection(code int) bool {
	return code >= 300 && code < 400
}

// IsClientError returns whether the codeonse is a client error
func IsClientError(code int) bool {
	return code >= 400 && code < 500
}

// IsServerError returns whether the codeonse is a server error
func IsServerError(code int) bool {
	return code >= 500 && code < 600
}
