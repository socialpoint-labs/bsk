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
	assert := assert.New(t)

	for _, testCase := range []struct {
		waitTime     time.Duration
		deltaTime    time.Duration
		tags         metrics.Tags
		expectedTags int
	}{
		{time.Millisecond * 50, 2 * time.Millisecond * 50, nil, 4},
		{time.Millisecond * 50, 2 * time.Millisecond * 50, []metrics.Tag{
			metrics.NewTag("test", "life"),
			metrics.NewTag("potato", "paco"),
		}, 6},
	} {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(testCase.waitTime)
		})

		recorder := metrics.NewRecorder()

		h := httpx.InstrumentDecorator(recorder)(handler)
		if testCase.tags != nil {
			h = httpx.InstrumentDecorator(recorder, testCase.tags...)(handler)
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest("", "", nil)
		assert.NoError(err)

		h.ServeHTTP(w, r)

		timer := recorder.Get("http.request_duration").(*metrics.RecorderTimer)
		assert.WithinDuration(timer.StartedTime(), timer.StoppedTime(), testCase.deltaTime)
		assert.Len(timer.Tags(), testCase.expectedTags)
	}
}
