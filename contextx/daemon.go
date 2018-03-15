package contextx

import (
	"context"
)

// RunDaemon is a helper function that runs the provided runners,
// also adding the most common features most Unix  server daemons have
func RunDaemon(runners ...Runner) {
	// Run each runner in it's own go-routine
	runner := MultiRunner(runners...)

	PosixSignalsAdapter()(runner).Run(context.Background())
}

// RunDaemonWithContext is a helper function that runs the provided runners with a given context,
// also adding the most common features most Unix  server daemons have
func RunDaemonWithContext(c context.Context, runners ...Runner) {
	// Run each runner in it's own go-routine
	runner := MultiRunner(runners...)

	PosixSignalsAdapter()(runner).Run(c)
}
