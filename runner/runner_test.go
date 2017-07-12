package runner_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/runner"
	"github.com/stretchr/testify/assert"
)

func wgAddAdapter(wg *sync.WaitGroup) runner.AdapterFunc {
	return func(runner runner.Runner) runner.Runner {
		wg.Add(1)
		return runner
	}
}

func wgDoneRunner(ctx context.Context, wg *sync.WaitGroup) runner.RunnerFunc {
	return func(ctx context.Context) {
		wg.Done()
	}
}

func TestARunnerRuns(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	runner.Empty().Run(ctx)
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

	runner.EmptyAdapter().Adapt(runner.Empty()).Run(ctx)
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
	mr := runner.Multi(
		adapter.Adapt(r),
		runner.MultiAdapter(
			runner.EmptyAdapter(),
			runner.EmptyAdapter(),
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
