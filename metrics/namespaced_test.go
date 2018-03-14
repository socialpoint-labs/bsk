package metrics_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/stretchr/testify/assert"
)

func TestPublisherWithNamespacedMetrics(t *testing.T) {
	assert := assert.New(t)

	namespace := "my_namespace"
	encoder := func(name string, op metrics.Op, value interface{}, tags metrics.Tags, rate float64) (string, error) {
		assert.True(strings.HasPrefix(name, namespace+"."))
		return "", nil
	}
	publisher := metrics.NewPublisher(ioutil.Discard, encoder, time.Second, nil)
	go publisher.Run(context.Background())

	namespacedPublisher := metrics.WithNamespace(publisher, namespace)
	namespacedPublisher.Counter("commands_executed").Inc()
	namespacedPublisher.Gauge("memory").Update(100)

	publisher.Flush()
}

func TestPublisherWithMultipleNamespaces(t *testing.T) {
	assert := assert.New(t)

	namespace1 := "namespace1"
	namespace2 := "namespace2"
	namespace3 := "namespace3"

	// namespaces are applied from last to first
	namespaces := fmt.Sprintf("%s.%s.%s", namespace1, namespace2, namespace3)
	encoder := func(name string, op metrics.Op, value interface{}, tags metrics.Tags, rate float64) (string, error) {
		assert.True(strings.HasPrefix(name, namespaces))
		return "", nil
	}
	publisher := metrics.NewPublisher(ioutil.Discard, encoder, time.Second, nil)
	go publisher.Run(context.Background())

	namespacedPublisher := metrics.WithNamespace(publisher, namespace1)
	namespacedPublisher = metrics.WithNamespace(namespacedPublisher, namespace2)
	namespacedPublisher = metrics.WithNamespace(namespacedPublisher, namespace3)

	namespacedPublisher.Gauge("memory").Update(100)
}
