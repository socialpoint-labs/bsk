package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/httpx"
	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestInstrument_RequestsDuration(t *testing.T) {
	const waitTime = 50 * time.Millisecond
	const deltaTime = 2 * waitTime

	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		tags         metrics.Tags
		expectedTags int
	}{
		{
			expectedTags: 4,
		},
		{
			tags: []metrics.Tag{
				metrics.NewTag("test", "test-value"),
				metrics.NewTag("foo", "bar"),
			},
			expectedTags: 6,
		},
	} {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(waitTime)
			w.WriteHeader(http.StatusNoContent)
		})

		recorder := metrics.NewRecorder()

		h := httpx.InstrumentDecorator(recorder)(handler)
		if tc.tags != nil {
			h = httpx.InstrumentDecorator(recorder, tc.tags...)(handler)
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest("", "", nil)
		a.NoError(err)

		h.ServeHTTP(w, r)

		timer, _ := recorder.Get("http.request_duration").(*metrics.RecorderTimer)
		a.WithinDuration(timer.StartedTime(), timer.StoppedTime(), deltaTime)
		a.Len(timer.Tags(), tc.expectedTags)
		a.Equal(http.StatusNoContent, w.Code)
	}
}
