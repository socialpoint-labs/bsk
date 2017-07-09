package server_test

import (
	"context"
	"testing"

	"math/rand"
	"time"

	"sync"

	"github.com/socialpoint-labs/bsk/server"
	"github.com/stretchr/testify/assert"
)

func TestRunDaemon(t *testing.T) {
	ch := make(chan struct{})

	runner := func() server.RunnerFunc {
		return func(ctx context.Context) {
			ch <- struct{}{}
		}
	}

	go server.RunDaemon(runner(), runner())

	<-ch
	<-ch
}

// We must assert a correct goroutine lifecycle (start and finish properly).
func TestDaemonRunsRunnersThatStartAndFinish(t *testing.T) {
	runner := func(wg *sync.WaitGroup) server.RunnerFunc {
		return func(ctx context.Context) {
			<-ctx.Done()
			wg.Done()
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	numRunners := rand.Intn(50)
	runners := []server.Runner{}
	wg := &sync.WaitGroup{}
	wg.Add(numRunners)

	for i := 0; i < numRunners; i++ {
		runners = append(runners, runner(wg))
	}

	go server.RunDaemonWithContext(ctx, runners...)

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	cancel()

	select {
	case <-done:
		assert.True(t, true)
	case <-time.After(time.Millisecond * 500):
		t.Fail()
	}
}
