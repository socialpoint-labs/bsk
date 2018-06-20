package httpx_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"
	"io"
	"time"

	"github.com/socialpoint-labs/bsk/httpx"
	"github.com/stretchr/testify/assert"
)

type closeNotifyWriter struct {
	*httptest.ResponseRecorder
	closed bool
}

func (w *closeNotifyWriter) CloseNotify() <-chan bool {
	notify := make(chan bool, 1)
	if w.closed {
		// return an already "closed" notifier
		notify <- true
	}
	return notify
}

func TestCloseNotifierDecorator_Client_Close_Connections(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			w.Header().Set("status", "canceled")
		}
	})

	h := httpx.CloseNotifierDecorator()(handler)

	w := &closeNotifyWriter{ResponseRecorder: httptest.NewRecorder(), closed: true}

	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	assert.NoError(t, err)

	h.ServeHTTP(w, r)

	assert.Equal(t, w.Header().Get("status"), "canceled")
}

func TestCloseNotifierDecorator_Request_Ends_Without_Cancelation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx.Err() == context.Canceled {
			w.Header().Set("status", "canceled")
		}
	})

	h := httpx.CloseNotifierDecorator()(handler)

	w := &closeNotifyWriter{ResponseRecorder: httptest.NewRecorder(), closed: false}

	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	assert.NoError(t, err)

	h.ServeHTTP(w, r)

	assert.NotEqual(t, w.Header().Get("status"), "canceled")
}

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
	content, err := ioutil.ReadAll(w.Body)
	assert.NoError(err)
	assert.Equal(string(content), "Internal Server Error")

	// now try the same thing but with the header in the request
	r, err = http.NewRequest("GET", "http://foo.bar", nil)
	assert.NoError(err)
	r.Header.Set(header, value)

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(w.Code, http.StatusOK)
	content, err = ioutil.ReadAll(w.Body)
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
			fmt.Fprint(w, text)
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
