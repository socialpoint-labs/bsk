package run_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/socialpoint-labs/bsk/run"
)

func TestWithDeadline(t *testing.T) {
	t.Parallel()

	t.Run("it returns when deadline is meet", func(t *testing.T) {
		delta := time.Millisecond * 300
		now := time.Now()
		deadline := now.Add(delta)

		var start, finish time.Time
		fn := func(ctx context.Context) error {
			start = time.Now()
			defer func() {
				finish = time.Now()
			}()

			<-ctx.Done()
			return nil
		}

		err := run.WithDeadline(deadline, fn)(context.Background())

		duration := finish.Sub(start)
		assert.True(t, duration >= delta)
		assert.NoError(t, err)
	})
}
