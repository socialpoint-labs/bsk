package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testdata = map[string]interface{}{"test": true}

func newTestRequest() *http.Request {
	r, err := http.NewRequest("GET", "Something", nil)
	if err != nil {
		panic("bad request: " + err.Error())
	}
	return r
}

func TestJSON(t *testing.T) {
	assert := assert.New(t)

	w := httptest.NewRecorder()
	r := newTestRequest()

	assert.Equal(JSONEncoder.ContentType(w, r), "application/json; charset=utf-8")
	assert.NoError(JSONEncoder.Encode(w, r, testdata))

	var data map[string]interface{}
	assert.NoError(json.Unmarshal(w.Body.Bytes(), &data))
	assert.Equal(data, testdata)
}

type testEncoder struct{}

func (testEncoder) Encode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return nil
}
func (testEncoder) ContentType(w http.ResponseWriter, r *http.Request) string {
	return "test/encoder"
}

func testRequestWithAccept(accept string) *http.Request {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic("bad request: " + err.Error())
	}
	r.Header.Set("Accept", accept)
	return r
}
