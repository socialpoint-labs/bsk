package run_test

import (
	"context"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/run"
)

func TestSleepWithContext(t *testing.T) {
	t.Parallel()

	t.Run("it sleeps", func(t *testing.T) {
		ctx := context.Background()
		run.Sleep(ctx, 1*time.Millisecond)
	})

	t.Run("it can be canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			run.Sleep(ctx, 10*time.Millisecond)
		}()

		cancel()
	})
}
