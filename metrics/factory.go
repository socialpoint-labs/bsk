package metrics

import (
	"net/url"
	"time"

	"github.com/socialpoint-labs/bsk/contextx"
)

// NewMetricsRunnerFromDSN creates a new metrics publisher and returns its Metrics
// and Runner from a DSN configuration. If the configuration is not valid it panics.
func NewMetricsRunnerFromDSN(DSN string) (Metrics, contextx.Runner) {
	// param validation
	URL, err := url.Parse(DSN)
	if err != nil {
		panic(err)
	}

	params := URL.Query()
	namespace := params.Get("namespace")
	if URL.Scheme == "datadog" && namespace == "" {
		panic("datadog metrics need a namespace")
	}

	// publisher is both Metrics and Runner
	var publisher *Publisher
	switch URL.Scheme {
	case "datadog":
		publisher = NewDataDog()
	case "stdout":
		publisher = NewStdout(100*time.Millisecond, DiscardErrors)
	case "discard":
		publisher = NewDiscardAll()
	default:
		panic("invalid metrics publisher type")
	}

	// init metrics
	var serviceTag Tag
	var m Metrics
	if namespace != "" {
		m = WithNamespace(publisher, namespace)
		serviceTag = NewTag("namespace", namespace)
	} else {
		m = publisher
	}

	// init runner
	var r contextx.Runner
	if params.Get("gostats") == "false" {
		r = publisher
	} else {
		r = contextx.MultiRunner(
			publisher,
			NewGoStatsRunner(publisher, FlushEvery15s, serviceTag),
		)
	}

	return m, r
}
