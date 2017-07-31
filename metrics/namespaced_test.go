package metrics

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/stretchr/testify/assert"
)

func TestPublisherWithNamespacedMetrics(t *testing.T) {
	assert := assert.New(t)

	namespace := "my_namespace"
	encoder := func(name string, op Op, value interface{}, tags Tags, rate float64) (string, error) {
		assert.True(strings.HasPrefix(name, namespace+"."))
		return "", nil
	}
	publisher := NewPublisher(ioutil.Discard, encoder, time.Second, nil)
	go publisher.Run(context.Background())

	namespacedPublisher := WithNamespace(publisher, namespace)
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
	encoder := func(name string, op Op, value interface{}, tags Tags, rate float64) (string, error) {
		assert.True(strings.HasPrefix(name, namespaces))
		return "", nil
	}
	publisher := NewPublisher(ioutil.Discard, encoder, time.Second, nil)
	go publisher.Run(context.Background())

	namespacedPublisher := WithNamespace(publisher, namespace1)
	namespacedPublisher = WithNamespace(namespacedPublisher, namespace2)
	namespacedPublisher = WithNamespace(namespacedPublisher, namespace3)

	namespacedPublisher.Gauge("memory").Update(100)
}
