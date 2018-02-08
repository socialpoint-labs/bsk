package metrics

import (
	"context"
	"runtime"
	"time"

	"github.com/socialpoint-labs/bsk/contextx"
)

// use packages var to avoid allocating every tick
var memStats runtime.MemStats

// A GoStatsRunner is a contextx.Runner that captures go VM stats and
// publishes them to the Metrics dependency every tick.
type GoStatsRunner struct {
	metrics Metrics
	tick    time.Duration
	tags    Tags
}

// NewGoStatsRunner returns a new Runner with the provided metrics and tick.
func NewGoStatsRunner(metrics Metrics, tick time.Duration, t ...Tag) contextx.Runner {
	return &GoStatsRunner{
		metrics: metrics,
		tick:    tick,
		tags:    append(t, Tag{"vm", "go"}),
	}
}

type goMetrics struct {
	// memory
	memAlloc   Gauge
	memFrees   Gauge
	memLookups Gauge
	memMallocs Gauge
	// others
	numGoroutines Gauge
}

// Run captures new values from the Go VM and publishes them to the metrics.
// Be careful (but much less so) with this because debug.ReadGCStats calls
// the C function runtime·lock(runtime·mheap) which, while not a stop-the-world
// operation, isn't something you want to be doing all the time.
func (r *GoStatsRunner) Run(ctx context.Context) {
	goMetrics := &goMetrics{
		// memory stats
		memAlloc:   r.metrics.Gauge("go.mem.allocated_bytes", r.tags...),
		memFrees:   r.metrics.Gauge("go.mem.frees", r.tags...),
		memLookups: r.metrics.Gauge("go.mem.lookups", r.tags...),
		memMallocs: r.metrics.Gauge("go.mem.allocations", r.tags...),
		// others
		numGoroutines: r.metrics.Gauge("go.goroutines", r.tags...),
	}

	ticker := time.NewTicker(r.tick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			collect(goMetrics)
		case <-ctx.Done():
			return
		}
	}
}

func collect(metrics *goMetrics) {
	// memory metrics
	runtime.ReadMemStats(&memStats) // This takes 50-200us.
	metrics.memAlloc.Update(memStats.Alloc)
	metrics.memFrees.Update(memStats.Frees)
	metrics.memLookups.Update(memStats.Lookups)
	metrics.memMallocs.Update(memStats.Mallocs)

	// others
	metrics.numGoroutines.Update(runtime.NumGoroutine())
}
