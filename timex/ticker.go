package timex

import (
	"context"
	"time"

	"github.com/socialpoint-labs/bsk/runner"
)

// RunInterval runs the provided function at intervals specified by the interval argument.
func RunInterval(ctx context.Context, interval time.Duration, f func()) {
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for {
		f()

		select {
		case <-ticker.C:
			// continue
		case <-ctx.Done():
			return
		}
	}
}

// IntervalRunner returns a runner.Runner that runs the function RunInterval with the provided context
func IntervalRunner(interval time.Duration, f func()) runner.Runner {
	return runner.RunnerFunc(func(ctx context.Context) {
		RunInterval(ctx, interval, f)
	})
}
