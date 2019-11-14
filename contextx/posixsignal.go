package contextx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// PosixSignalsAdapter returns an adapter that augments the passed runner
// context with cancellation when some OS signals occurs.
//
// It creates the os.Signal channel and call the private signalsAdapter to offer
// a simpler developer experience.
func PosixSignalsAdapter() AdapterFunc {
	return signalsAdapter(make(chan os.Signal))
}

// Note: We didn't figure out a way or running deterministic test while dealing
// with Posix signals. The tests were failing randomly and taking random times
// to propagate the signals to the process. For this reason we provide this
// private constructor, where a os.Signal channel is passed, to be used for tests.
func signalsAdapter(c chan os.Signal) AdapterFunc {
	return func(runner Runner) Runner {
		return RunnerFunc(func(ctx context.Context) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			// Run the adapter runner in a go-routine so that the current go-routine
			// takes care of signal handling
			go runner.Run(ctx)

			// Bear in mind that SIGKILL and SIGSTOP cannot be trapped, see
			// https://goo.gl/5gvRrN. Also know that pressing <ctrl-c> from CLI
			// will make the process receive a SIGINT.
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

			for {
				select {
				case signal := <-c:
					switch signal {
					case syscall.SIGINT, syscall.SIGTERM:
						cancel()

						// TODO: How do we wait until everybody cancelled ?

						return

					case syscall.SIGHUP:
						// TODO: reload, but how do we know if everybody has cancelled and returned
					}

				case <-ctx.Done():
					return
				}
			}
		})
	}
}
