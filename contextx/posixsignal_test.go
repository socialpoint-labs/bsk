package contextx

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextCancellationWhenSignalsAreNotified(t *testing.T) {
	a := assert.New(t)

	testCases := []struct {
		signal      syscall.Signal
		isCanceled  bool
		description string
	}{
		{syscall.SIGINT, true, "SIGINT should cancel the context"},
		{syscall.SIGHUP, false, "config signal handling does nothing, so context won't be canceled"},
		{syscall.SIGWINCH, false, "this signal is not even handled, same thing than sighup"},
	}

	for _, testCase := range testCases {
		finished := make(chan struct{})

		runner := func() RunnerFunc {
			return func(ctx context.Context) {
				<-ctx.Done()
				close(finished)
			}
		}

		ctx := context.Background()
		c := make(chan os.Signal)

		go signalsAdapter(c).Adapt(runner()).Run(ctx)

		// send the signal
		c <- testCase.signal

		// wait till runner has finished running
		select {
		case <-finished:
			a.Equal(testCase.isCanceled, true, testCase.description)
		case <-time.After(10 * time.Millisecond):
			a.Equal(testCase.isCanceled, false, testCase.description)
		}
	}
}
