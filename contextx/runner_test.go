package contextx_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/contextx"
	"github.com/stretchr/testify/assert"
)

func wgAddAdapter(wg *sync.WaitGroup) contextx.AdapterFunc {
	return func(runner contextx.Runner) contextx.Runner {
		wg.Add(1)
		return runner
	}
}

func wgDoneRunner(ctx context.Context, wg *sync.WaitGroup) contextx.RunnerFunc {
	return func(ctx context.Context) {
		wg.Done()
	}
}

func TestARunnerRuns(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	contextx.NopRunner().Run(ctx)
	// nothing to assert here really

	wg := &sync.WaitGroup{}
	wg.Add(1)
	runner := wgDoneRunner(ctx, wg)
	runner.Run(ctx)
	assert.False(waitTimeout(wg, time.Second), "waitgroup timeout")
}

func TestRunnerAdaptation(t *testing.T) {
	assert := assert.New(t)
	ctx := context.TODO()

	noopAdapter().Adapt(contextx.NopRunner()).Run(ctx)
	// nothing to assert here really

	wg := &sync.WaitGroup{}

	adapter := wgAddAdapter(wg)
	runner := wgDoneRunner(ctx, wg)
	adapter.Adapt(runner).Run(ctx)
	assert.False(waitTimeout(wg, time.Second), "waitgroup timeout")
}

func TestMultiRunnerAndMultiAdapter(t *testing.T) {
	assert := assert.New(t)
	ctx := context.TODO()

	wg := &sync.WaitGroup{}

	r := wgDoneRunner(ctx, wg)
	adapter := wgAddAdapter(wg)
	mr := contextx.MultiRunner(
		adapter.Adapt(r),
		contextx.MultiAdapter(
			noopAdapter(),
			noopAdapter(),
			adapter,
		).Adapt(r),
	)

	mr.Run(ctx)
	assert.False(waitTimeout(wg, time.Second), "waitgroup timeout")
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// noopyAdapter returns a no-op adapter that  return the same provided Runner.
func noopAdapter() contextx.Adapter {
	return contextx.AdapterFunc(func(runner contextx.Runner) contextx.Runner {
		return runner
	})
}
