# Package server

`package server` provides the types needed to run application servers.

# Design guide and goals

## context.Context

A core dependency of this package is the `context.Context` type, which is very
useful when applying [Go Concurrency Patterns](https://blog.golang.org/context)
and implement things like goroutine cancellation etc.

## Main types

#### Runner

A `Runner` is the basic behaviour of a Server, which will recive a
`context.Context` and run some logic.

There is also available its functional countertype `RunnerFunc` in case no
state is needed, and `MultiRunner` which runs multiple `Runner`s at the same
time.

#### Adapter

An `Adapter` augments the capabilities of a `Runner` to do more things when it
runs.

There is also available its functional countertype `AdapterFunc` in case no
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
