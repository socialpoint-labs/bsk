package httpx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouterRootPathMatch(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	r := NewRouter()
	r.Route("/", StatusOKHandler)
	r.ServeHTTP(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestRouterNotFound(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest("GET", "/hello", nil)
	recorder := httptest.NewRecorder()

	r := NewRouter()
	r.ServeHTTP(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusNotFound, recorder.Code)
}

func TestRouterInvalidPathDoesNotMatch(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest("GET", "/hello", nil)
	recorder := httptest.NewRecorder()

	r := NewRouter()
	r.Route("/", StatusOKHandler)
	r.ServeHTTP(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusNotFound, recorder.Code)
}

func TestRouterOK(t *testing.T) {
	assert := assert.New(t)

	req, err := http.NewRequest("GET", "/hello", nil)
	recorder := httptest.NewRecorder()

	r := NewRouter()
	r.Route("/hello", StatusOKHandler)
	r.ServeHTTP(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestRouterAdapterAdaptsOneOfMultipleHandlers(t *testing.T) {
	assert := assert.New(t)

	key := "key"
	value := "value"

	r := NewRouter()
	r.Route("/hello", NoopHandler())
	r.Route("/custom", AddValueToContextDecorator(key, value)(NoopHandler()))

	req, err := http.NewRequest("GET", "/hello", nil)
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("", recorder.Header().Get("custom"))

	req, err = http.NewRequest("GET", "/custom", nil)
	recorder = httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal(value, recorder.Header().Get(key))

}

func TestRouterWithDecorators(t *testing.T) {
	assert := assert.New(t)

	decorators := []Decorator{
		CheckSPAuthHeaderDecorator("valid"),
		EnableCORSDecorator(),
	}

	routerHello := NewRouter()
	routerHello.Route("/foo", StatusOKHandler)

	routerAdmin := NewRouter()
	routerAdmin.Route("/hello", routerHello)
	routerAdmin.Route("/world", StatusOKHandler)

	routerDashboard := NewRouter()
	routerDashboard.Route("/", StatusOKHandler)

	router := NewRouter()
	router.Route("/", StatusOKHandler)
	router.Route("/_catalog", StatusOKHandler, decorators...)
	router.Route("/admin", routerAdmin, decorators...)
	router.Route("/dashboard", routerDashboard)

	for _, testcase := range []struct {
		uri            string
		expectedStatus int
		secret         string
	}{
		{"/", http.StatusOK, ""},
		{"/admin", http.StatusNotFound, ""},
		{"/_catalog", http.StatusUnauthorized, ""},
		{"/admin/hello/foo", http.StatusUnauthorized, ""},
		{"/admin/world", http.StatusUnauthorized, ""},
		{"/_catalog", http.StatusUnauthorized, "invalid"},
		{"/admin/hello/foo", http.StatusUnauthorized, "invalid"},
		{"/admin/world", http.StatusUnauthorized, "invalid"},
		{"/_catalog", http.StatusOK, "valid"},
		{"/admin/hello/foo", http.StatusOK, "valid"},
		{"/admin/world", http.StatusOK, "valid"},
		{"/dashboard", http.StatusOK, ""},
	} {
		req, err := http.NewRequest("GET", testcase.uri, nil)
		recorder := httptest.NewRecorder()

		if testcase.secret != "" {
			req.Header.Add("X-SP-Sign", testcase.secret)
		}

		router.ServeHTTP(recorder, req)

		assert.NoError(err)
		assert.Equal(testcase.expectedStatus, recorder.Code, fmt.Sprintf("testcase: (%v, %v, %v)", testcase.uri, testcase.expectedStatus, testcase.secret))
	}
}

func TestDecoratorsAreOnlyExecutedOnce(t *testing.T) {
	assert := assert.New(t)
	var count int

	worldRouter := NewRouter()
	worldRouter.Route("/", StatusOKHandler)

	helloRouter := NewRouter()
	helloRouter.Route("/world", worldRouter)

	baseRouter := NewRouter()
	baseRouter.Route("/hello", helloRouter, countDecorator(&count))

	req, err := http.NewRequest("GET", "/hello/world", nil)
	recorder := httptest.NewRecorder()

	baseRouter.ServeHTTP(recorder, req)

	assert.NoError(err)
	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal(1, count, "Decorators are being executed more than once per request")
}

func AddValueToContextDecorator(key, value string) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(key, value)
			h.ServeHTTP(w, r)
		})
	}
}

func countDecorator(count *int) Decorator {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*count++
			h.ServeHTTP(w, r)
		})
	}
}
