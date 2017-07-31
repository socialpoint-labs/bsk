package metrics

import "fmt"

const namespaceSeparator = "."

type namespaced struct {
	namespace string
	adapted   Metrics
}

// NewNamespaced returns the metrics with the metrics names with a namespace or prefix.
func NewNamespaced(m Metrics, namespace string) Metrics {
	return &namespaced{namespace, m}
}

func (n *namespaced) Counter(name string, tags ...Tag) Counter {
	return n.adapted.Counter(n.prefix(name), tags...)
}

func (n *namespaced) Gauge(name string, tags ...Tag) Gauge {
	return n.adapted.Gauge(n.prefix(name), tags...)
}

func (n *namespaced) Event(name string, tags ...Tag) Event {
	return n.adapted.Event(n.prefix(name), tags...)
}

func (n *namespaced) Timer(name string, tags ...Tag) Timer {
	return n.adapted.Timer(n.prefix(name), tags...)
}

func (n *namespaced) prefix(name string) string {
	return fmt.Sprintf("%s%s%s", n.namespace, namespaceSeparator, name)
}
