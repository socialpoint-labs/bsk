package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/socialpoint-labs/bsk/httpx"
	"github.com/stretchr/testify/assert"
)

func TestStatusHandler(t *testing.T) {
	assert := assert.New(t)

	for _, code := range []int{200, 250, 300, 330, 404, 500, 505} {

		h := httpx.StatusHandler(code)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, &http.Request{})
		assert.Equal(code, w.Code)
	}
}
