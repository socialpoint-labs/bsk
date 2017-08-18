package contextx

import "context"

// Runner is the basic behaviour of a Server, which will run some logic. It
// receives a Context to allow the implementation of several patterns like
// deadlines and cancellation signals, parameter bags, dependency injection
// container, etc.
type Runner interface {
	Run(context.Context)
}

// RunnerFunc is the functional countertype of an Runner, it allows the use of
// ordinary functions as a Runner.
type RunnerFunc func(context.Context)

// Run bust be implemented by RunnerFunc to implements Runner, so when someone
// calls the Run function over an element of this type it will execute itself.
func (f RunnerFunc) Run(ctx context.Context) {
	f(ctx)
}

// Empty is a Runner that does nothing for testing purposes only.
func Empty() RunnerFunc {
	return func(context.Context) {
		// don't do anything
	}
}

// Multi receives multiple Runners and returns a new RunnerFunc that runs them
// in go-routines.
func Multi(runners ...Runner) RunnerFunc {
	return func(ctx context.Context) {
		for _, runner := range runners {
			go runner.Run(ctx)
		}
	}
}

// An Adapter adds certain functionality to a Runner.
// Note: should be passed by value to enforce immutability, which have some
// nice properties like safety in concurrent programming.
type Adapter interface {
	Adapt(Runner) Runner
}

// AdapterFunc is the functional countertype of an Adapter, it allows the use of
// ordinary functions as a Adapter.
type AdapterFunc func(Runner) Runner

// Adapt implements Adapter so the AdapterFunc complies with the interface.
func (a AdapterFunc) Adapt(runner Runner) Runner {
	return a(runner)
}

// EmptyAdapter is an adapter func that does nothing for testing purposes only.
func EmptyAdapter() AdapterFunc {
	return func(runner Runner) Runner {
		// don't do anything
		return runner
	}
}

// MultiAdapter receives multiple Adapters and returns a new AdapterFunc
// that will adapt the runner with all of them, from first to last.
func MultiAdapter(adapters ...Adapter) AdapterFunc {
	return func(runner Runner) Runner {
		for _, adapter := range adapters {
			runner = adapter.Adapt(runner)
		}
		return runner
	}
}
