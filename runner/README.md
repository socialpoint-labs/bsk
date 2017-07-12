# Package runner

## Description

`package runner` provides abstractions needed to run tasks. The goal is to
provide a common abstraction that will enable running and stopping tasks, and
also to compose and decorate its executions.

## Types

Everything is build around this simple interface:
```go
type Runner interface {
	Run(context.Context)
}
```
It receives a `context.Context` type, which is very useful when applying [Go
Concurrency Patterns](https://blog.golang.org/context) and implement concerns
like cancellation.

There is also available its functional countertype `RunnerFunc` in case no
state is needed, and `Multi` which runs multiple `Runner`s at the same time.

An `Adapter` augments the capabilities of a `Runner` to do more things when it
runs. There is also available its functional countertype `AdapterFunc` in case no
state is needed, and a `MultiAdapter` which adapts multiple `Runner`s at the
same time.

## TODO

Create runners and adapters for the basic cross-cutting concerns of an application:

- [x] OS signal handling (stop, reload, kill, config change)
- [ ] graceful shutdown handling
- [ ] panic handling
- [ ] logging
- [ ] instrumentation
- [ ] event dispatching
- [ ] compile information
- [ ] metrics
