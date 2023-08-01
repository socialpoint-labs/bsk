package httpx_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/httpx"
)

func TestAddHeaderDecorator(t *testing.T) {
	assert := assert.New(t)

	h := httpx.AddHeaderDecorator("key", "value1")(
		httpx.AddHeaderDecorator("key", "value2")(
			httpx.NoopHandler()))

	w := httptest.NewRecorder()
	r := &http.Request{}

	h.ServeHTTP(w, r)

	headers := w.Result().Header[http.CanonicalHeaderKey("key")]
	assert.Equal("value1", headers[0])
	assert.Equal("value2", headers[1])
}

func TestSetHeaderDecorator(t *testing.T) {
	assert := assert.New(t)

	h := httpx.SetHeaderDecorator("key", "value1")(
		httpx.SetHeaderDecorator("key", "value2")(
			httpx.NoopHandler()))

	w := httptest.NewRecorder()
	r := &http.Request{}

	h.ServeHTTP(w, r)

	assert.Equal("value2", w.Header().Get("key"))
}

func TestCheckHeaderDecorator(t *testing.T) {
	assert := assert.New(t)

	header := "foo"
	value := "bar"
	code := http.StatusInternalServerError

	w := httptest.NewRecorder()
	r := &http.Request{}

	handler := httpx.CheckHeaderDecorator(header, value, code)(httpx.NoopHandler())

	handler.ServeHTTP(w, r)
	assert.Equal(w.Code, code)
	content, err := io.ReadAll(w.Body)
	assert.NoError(err)
	assert.Equal(string(content), "Internal Server Error")

	// now try the same thing but with the header in the request
	r, err = http.NewRequest("GET", "http://foo.bar", nil)
	assert.NoError(err)
	r.Header.Set(header, value)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(w.Code, http.StatusOK)
	content, err = io.ReadAll(w.Body)
	assert.NoError(err)
	assert.Equal(string(content), "")
}

func TestRootDecorator(t *testing.T) {
	assert := assert.New(t)

	h := httpx.RootDecorator()(httpx.NoopHandler())

	// Test a request to the root path
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(err)

	h.ServeHTTP(w, req)
	assert.Equal(http.StatusOK, w.Code)

	// Test  a request to a random non-root path
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/whatever", nil)
	assert.NoError(err)

	h.ServeHTTP(w, req)
	assert.Equal(http.StatusNotFound, w.Code)

}

func TestEnableCORSDecorator(t *testing.T) {
	assert := assert.New(t)

	h := httpx.EnableCORSDecorator()(httpx.NoopHandler())

	w := httptest.NewRecorder()
	r := &http.Request{}
	r.Method = "OPTIONS"

	h.ServeHTTP(w, r)

	assert.Equal("*", w.Result().Header.Get("Access-Control-Allow-Origin"))
	assert.Equal("GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS", w.Result().Header.Get("Access-Control-Allow-Methods"))
	assert.Equal("Origin,Accept,Content-Type,Authorization", w.Result().Header.Get("Access-Control-Allow-Headers"))

	assert.Equal(http.StatusOK, w.Code)
}

func TestIfDecorator(t *testing.T) {
	trueHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cond", "true")
	})

	cond := func(w http.ResponseWriter, r *http.Request) bool {
		return r.URL.Path == "/true"
	}

	h := httpx.IfDecorator(cond, trueHandler)(httpx.NoopHandler())

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com/true", nil)
	h.ServeHTTP(w, r)

	assert.Equal(t, "true", w.Header().Get("cond"))

	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "http://example.com/false", nil)
	h.ServeHTTP(nil, r)

	assert.Equal(t, "", w.Header().Get("cond"))
}

func TestTimeoutDecorator(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Deadline(); ok {
			w.Header().Set("status", "deadline")
		}
	})

	h := httpx.TimeoutDecorator(time.Second)(handler)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	assert.NoError(t, err)

	h.ServeHTTP(w, r)
	assert.Equal(t, "deadline", w.Header().Get("status"))
}

type writerMock struct {
	io.Writer
	loggedBytes []byte
}

func TestLogging(t *testing.T) {
	writerMock := &writerMock{}
	text := "Test OK"

	router := httpx.NewRouter()
	router.Route(
		"/test",
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, text)
		}),
		httpx.LoggingDecorator(writerMock),
	)

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:8080/test", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, text, recorder.Body.String())

	assert.Contains(t, string(writerMock.loggedBytes), "GET /test HTTP/1.1 200")
}

func (w *writerMock) Write(p []byte) (n int, err error) {
	w.loggedBytes = append(w.loggedBytes, p...)

	return len(p), nil
}
