package collector

import (
	"context"
	"log"
	"pulsecat/internal/metrics"
	"sync"
)

// manages a set of collectors and their associated ring buffers.
type Runner struct {
	mu         sync.RWMutex
	collectors metrics.MetricMap[Collector]
	wg         sync.WaitGroup
}

// creates a new collector runner
func NewRunner() *Runner {
	return &Runner{
		collectors: make(metrics.MetricMap[Collector]),
	}
}

// adds a collector to the registry.
// If a collector for the same metric type already exists, it is replaced.
func (r *Runner) Register(c Collector) {
	r.mu.Lock()
	defer r.mu.Unlock()

	typ := c.Type()
	r.collectors[typ] = c
	log.Printf("Registered collector %s (%v)", c.Name(), typ)
}

// begins periodic collection for all registered collectors.
// Each collector runs in its own goroutine, collecting at the configured interval.
// The provided context can be used to stop all collectors.
func (r *Runner) Start(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, collector := range r.collectors {
		r.wg.Add(1)
		go func() {
			collector.Start()
			r.wg.Done()
		}()
	}
	log.Printf("Started %d collector(s)", len(r.collectors))
}

// halts all collectors and waits for them to finish.
func (r *Runner) Stop() {
	r.mu.Lock()
	for _, collector := range r.collectors {
		collector.Stop()
	}
	r.mu.Unlock()

	r.wg.Wait()
	log.Printf("All collectors stopped")
}

// returns the collector for the given metric type.
// Returns nil if not found.
func (r *Runner) GetCollector(typ metrics.MetricType) Collector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collectors[typ]
}

// returns a slice of all registered metric types.
func (r *Runner) ListCollectors() []metrics.MetricType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]metrics.MetricType, 0, len(r.collectors))
	for typ := range r.collectors {
		types = append(types, typ)
	}
	return types
}
