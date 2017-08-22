package contextx

import "context"

// A Runner allows to run a context-aware process to get advantage of
// context deadlines, cancelation signals, and other request-scoped values
// across API boundaries and between such processes.
type Runner interface {
	Run(context.Context)
}

// The RunnerFunc type is an adapter to allow the use of ordinary functions as Runners.
type RunnerFunc func(context.Context)

// Run calls the wrapped function with the provided context.
func (f RunnerFunc) Run(ctx context.Context) {
	f(ctx)
}

// NopRunner returns a no-op Runner that does nothing.
func NopRunner() RunnerFunc {
	return func(context.Context) {}
}

// MultiRunner returns a Runner that run multiple provided Runner in go-routines.
func MultiRunner(runners ...Runner) Runner {
	return RunnerFunc(func(ctx context.Context) {
		for _, runner := range runners {
			go runner.Run(ctx)
		}
	})
}

// An Adapter adds certain functionality to a Runner.
// Note: should be passed by value to enforce immutability, which have some
// nice properties like safety in concurrent programming.
type Adapter interface {
	Adapt(Runner) Runner
}

// AdapterFunc allows the use of ordinary functions as a Adapter.
type AdapterFunc func(Runner) Runner

// Adapt implements Adapter so the AdapterFunc complies with the interface.
func (a AdapterFunc) Adapt(runner Runner) Runner {
	return a(runner)
}

// MultiAdapter receives multiple Adapters and returns a new Adapter
// that will adapt the runner with all of them, from first to last.
func MultiAdapter(adapters ...Adapter) Adapter {
	return AdapterFunc(func(runner Runner) Runner {
		for _, adapter := range adapters {
			runner = adapter.Adapt(runner)
		}
		return runner
	})
}
