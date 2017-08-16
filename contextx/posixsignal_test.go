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
	assert := assert.New(t)

	testCases := []struct {
		signal      syscall.Signal
		isCanceled  bool
		description string
	}{
		{syscall.SIGUSR1, true, "use SIGUSR1 instead of SIGKILL/SIGTERM otherwise the test process is killed"},
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
			assert.Equal(testCase.isCanceled, true, testCase.description)
		case <-time.After(10 * time.Millisecond):
			assert.Equal(testCase.isCanceled, false, testCase.description)
		}
	}
}
