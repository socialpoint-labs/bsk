# package throttler

`package throttler` provides utilities to limit the execution of a set of actions

A Typical scenario to use a throttler would be when we want to execute a large amount of actions against a remote 
service, but we don't want to cause a stampede.

The throttler is configured based on 2 parameters, an amount `n` of actions to be executed, and a time lapse `t`. The 
throttler guarantees that maximum `n` actions will be started every `t`.

## Usage

For detailed usage see [examples](example_test.go)