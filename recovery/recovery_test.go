package recovery

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/logx"
)

func TestWithRequestResponseLogs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("panic handler serializes panic", func(t *testing.T) {
		reached := false
		defer func() {
			a.True(reached)
		}()

		w := bytes.NewBufferString("")
		logger := logx.New(logx.WriterOpt(w))
		spyExitFunc := func() {
			a.Contains(w.String(), "test panicking")
			reached = true
		}

		handler := Handler(logger, WithExitFunction(spyExitFunc))
		defer handler()

		panic("test panicking")
	})
}
