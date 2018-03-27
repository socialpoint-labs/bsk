package run_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRetry(t *testing.T) {
	t.Parallel()

	t.Run("it retries all the attempts", func(t *testing.T) {
		backoff := run.ConstantBackoff(3, 10*time.Millisecond)
		calls := make(chan struct{})

		classifier := func(err error) run.Result {
			return run.Retry
		}

		fn := func(ctx context.Context) error {
			calls <- struct{}{}
			return nil
		}

		go func() {
			err := run.WithRetry(backoff, classifier, fn)(context.Background())
			require.NoError(t, err)
		}()

		<-calls
		<-calls
		<-calls
	})

	t.Run("it cancels runs when the classifier cancels it", func(t *testing.T) {
		backoff := run.ConstantBackoff(3, 10*time.Millisecond)
		calls := make(chan struct{})

		classifier := func(err error) run.Result {
			return run.Cancel
		}

		fn := func(ctx context.Context) error {
			calls <- struct{}{}
			return nil
		}

		go func() {
			err := run.WithRetry(backoff, classifier, fn)(context.Background())
			require.NoError(t, err)
		}()

		<-calls
	})
}

func TestConstantBackoff(t *testing.T) {
	t.Parallel()

	backoff := run.ConstantBackoff(3, 10*time.Millisecond)
	assert.Equal(t, []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond}, backoff)
}

func TestExponentialBackoff(t *testing.T) {
	t.Parallel()

	backoff := run.ExponentialBackoff(3, 10*time.Millisecond)
	assert.Equal(t, []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond}, backoff)
}

func TestNotNilClassifier(t *testing.T) {
	assert.Equal(t, run.Succeed, run.NotNilClassifier()(nil))
	assert.Equal(t, run.Retry, run.NotNilClassifier()(errors.New("test error")))
}
