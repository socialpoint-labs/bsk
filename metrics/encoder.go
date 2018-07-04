package metrics

import (
	"fmt"
	"strings"
	"time"
)

// Encoder is the signature for encoders
type Encoder func(name string, op Op, value interface{}, tags Tags, rate float64) (string, error)

// StdoutEncoder is a simple encoder to be used to write to stdout.
func StdoutEncoder(name string, op Op, value interface{}, tags Tags, rate float64) (string, error) {
	return fmt.Sprintf("METRIC: %s | %d | %v | %v | %f\n", name, op, value, tags, rate), nil
}

// StatsDEncoder implements statsd protocol
func StatsDEncoder(name string, op Op, value interface{}, tags Tags, rate float64) (string, error) {
	ct := make([]string, len(tags))
	for k, t := range tags {
		ct[k] = fmt.Sprintf("%v:%v", t.Key, t.Value)
	}
	st := strings.Join(ct, ",")
	if st != "" {
		st = "|#" + st
	}

	switch op {
	case OpCounterAdd:
		return fmt.Sprintf("%s:%v|c|@%.4f%s\n", name, value, rate, st), nil
	case OpGaugeUpdate:
		return fmt.Sprintf("%s:%v|g|@%.4f%s\n", name, value, rate, st), nil
	case OpHistogramUpdate:
		return fmt.Sprintf("%s:%v|h|@%.4f%s\n", name, value, rate, st), nil
	case OpEventSend:
		title := fmt.Sprintf("%v", name)
		text := fmt.Sprintf("%v", value)
		return fmt.Sprintf("_e{%d,%d}:%s|%s%s\n", len(title), len(text), title, text, st), nil
	case OpTimerStop:
		return fmt.Sprintf("%s:%v|ms|@%.4f%s\n", name, value, rate, st), nil
	}

	return "", fmt.Errorf("statsd encoder: operation %v not supported", op)
}

// LibratoStatsDEncoder implements StatsD protocol but ignores tags to comply with Librato API
func LibratoStatsDEncoder(name string, op Op, value interface{}, _ Tags, rate float64) (string, error) {
	switch op {
	case OpCounterAdd:
		return fmt.Sprintf("%s:%v|c|@%.4f\n", name, value, rate), nil
	case OpGaugeUpdate:
		return fmt.Sprintf("%s:%v|g|@%.4f\n", name, value, rate), nil
	case OpHistogramUpdate:
		return fmt.Sprintf("%s:%v|h|@%.4f\n", name, value, rate), nil
	case OpEventSend:
		title := fmt.Sprintf("%v", name)
		text := fmt.Sprintf("%v", value)
		return fmt.Sprintf("_e{%d,%d}:%s|%s\n", len(title), len(text), title, text), nil
	case OpTimerStop:
		return fmt.Sprintf("%s:%v|ms|@%.4f\n", name, value, rate), nil
	}

	return "", fmt.Errorf("librato encoder: operation %v not supported", op)
}

// DataDogLambdaEncoder implements generating DataDog metrics from AWS Lambda functions.
// Supported metrics are: count, gauge, histogram.
// Events are NOT supported.
// See https://docs.datadoghq.com/integrations/amazon_lambda/
func DataDogLambdaEncoder(name string, op Op, value interface{}, tags Tags, rate float64) (string, error) {
	return formatDataDogLambdaMetric(name, op, value, tags, time.Now())
}

func formatDataDogLambdaMetric(name string, op Op, value interface{}, tags Tags, t time.Time) (string, error) {
	ct := make([]string, len(tags))
	for k, t := range tags {
		ct[k] = fmt.Sprintf("%v:%v", t.Key, t.Value)
	}
	st := strings.Join(ct, ",")

	switch op {
	case OpCounterAdd:
		return fmt.Sprintf("MONITORING|%d|%v|count|%s|#%s\n", t.Unix(), value, name, st), nil
	case OpGaugeUpdate:
		return fmt.Sprintf("MONITORING|%d|%v|gauge|%s|#%s\n", t.Unix(), value, name, st), nil
	case OpHistogramUpdate:
		return fmt.Sprintf("MONITORING|%d|%v|histogram|%s|#%s\n", t.Unix(), value, name, st), nil
	}

	return "", fmt.Errorf("datadog-lambda encoder: operation %q not supported", op)
}

// NamespacedEncoder creates a new encoder from a given encoder and namespace
func NamespacedEncoder(e Encoder, namespace string) Encoder {
	return func(name string, op Op, value interface{}, tags Tags, rate float64) (string, error) {
		return e(fmt.Sprintf("%s.%s", namespace, name), op, value, tags, rate)
	}
}
