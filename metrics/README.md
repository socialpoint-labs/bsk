# package metrics

`package metrics` provides a set of utilities for service instrumentation.

It currently supports the following metric types:

- Counters
- Gauges
- Histograms
- Events

Metrics and instrumentation is a well-defined and mature topic, so `package metrics` ONLY provides a common and minimal
interface to forward metrics to 3rd party tools/aggregators/etc.

It also supports:

- namespacing the metric names using `WithNamespace`
- automatic Go VM stats using `WithGoStats`

## Usage

For detailed usage see [examples](example_test.go)

## Integrating with Datadog

Dogstatsd (Datadog agent) is a statsd backend server, so you can send custom metrics to the agent using UDP and the statsd 
protocol.

To integrate with Datadog agent, just provide an UDP network connection for the publisher `io.writer`. 

## Integration with DataDog in AWS Lambda functions

A publisher and an encoder are provided to update DataDog metrics from within the execution of AWS Lambda functions.

Only counters, gauges and histograms are supported by DataDog at the moment.

Find more info on [how the integration works](https://docs.datadoghq.com/integrations/amazon_lambda/) in the DataDog site.