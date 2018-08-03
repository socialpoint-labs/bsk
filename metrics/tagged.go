package metrics

// taggedMetrics is a metrics publisher decorator that add predefined tags to all metrics
type taggedMetrics struct {
	Metrics
	tags []Tag
}

// NewTaggedMetrics returns a new metrics publisher with predefined tags for all the metrics
func NewTaggedMetrics(m Metrics, tags ...Tag) Metrics {
	return &taggedMetrics{Metrics: m, tags: tags}
}

// Provide a Counter with the given name and tags
func (m *taggedMetrics) Counter(name string, tags ...Tag) Counter {
	return m.Metrics.Counter(name, m.tags...).WithTags(tags...)
}

// Provide a Gauge with the given name and tags
func (m *taggedMetrics) Gauge(name string, tags ...Tag) Gauge {
	return m.Metrics.Gauge(name, m.tags...).WithTags(tags...)
}

// Provide a Timer with the given name and tags
func (m *taggedMetrics) Timer(name string, tags ...Tag) Timer {
	return m.Metrics.Timer(name, m.tags...).WithTags(tags...)
}

// Provide an Event with the given name and tags
func (m *taggedMetrics) Event(name string, tags ...Tag) Event {
	return m.Metrics.Event(name, m.tags...).WithTags(tags...)
}

// Provide an Histogram with the given name and tags
func (m *taggedMetrics) Histogram(name string, tags ...Tag) Histogram {
	return m.Metrics.Histogram(name, m.tags...).WithTags(tags...)
}
