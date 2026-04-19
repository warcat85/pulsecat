package collector

import (
	"context"
	"log"
	"pulsecat/internal/storage"
	"sync"
	"time"

	v1 "pulsecat/pkg/api/v1"
)

// Snapshot represents a single metric snapshot with its timestamp.
type Snapshot struct {
	Timestamp time.Time
	Data      any // concrete type depends on metric type (e.g., *v1.LoadAverage)
}

// Collector is the interface that all metric collectors must implement.
type Collector interface {
	// Type returns the metric type this collector produces.
	Type() v1.MetricType
	// Name returns a human-readable name for the collector.
	Name() string
	// Collect performs a single collection cycle and returns a snapshot.
	// The returned data must be of the appropriate protobuf message type.
	Collect(ctx context.Context) (any, error)
}

// Registry manages a set of collectors and their associated ring buffers.
type Registry struct {
	mu          sync.RWMutex
	collectors  map[v1.MetricType]Collector
	buffers     map[v1.MetricType]*storage.RingBuffer[Snapshot]
	interval    time.Duration
	bufferCap   int
	cancelFuncs map[v1.MetricType]context.CancelFunc
	wg          sync.WaitGroup
}

// NewRegistry creates a new collector registry with the given collection interval
// and buffer capacity (maximum number of snapshots to retain per collector).
func NewRegistry(interval time.Duration, bufferCap int) *Registry {
	return &Registry{
		collectors:  make(map[v1.MetricType]Collector),
		buffers:     make(map[v1.MetricType]*storage.RingBuffer[Snapshot]),
		interval:    interval,
		bufferCap:   bufferCap,
		cancelFuncs: make(map[v1.MetricType]context.CancelFunc),
	}
}

// Register adds a collector to the registry and creates a ring buffer for it.
// If a collector for the same metric type already exists, it is replaced.
func (r *Registry) Register(c Collector) {
	r.mu.Lock()
	defer r.mu.Unlock()

	typ := c.Type()
	r.collectors[typ] = c
	r.buffers[typ] = storage.NewRingBuffer[Snapshot](r.bufferCap)
	log.Printf("Registered collector %s (%v)", c.Name(), typ)
}

// Start begins periodic collection for all registered collectors.
// Each collector runs in its own goroutine, collecting at the configured interval.
// The provided context can be used to stop all collectors.
func (r *Registry) Start(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for typ, collector := range r.collectors {
		if _, already := r.cancelFuncs[typ]; already {
			continue
		}
		ctx, cancel := context.WithCancel(ctx)
		r.cancelFuncs[typ] = cancel
		r.wg.Add(1)
		go r.runCollector(ctx, collector)
	}
	log.Printf("Started %d collector(s) with interval %v", len(r.collectors), r.interval)
}

// Stop halts all collectors and waits for them to finish.
func (r *Registry) Stop() {
	r.mu.Lock()
	for _, cancel := range r.cancelFuncs {
		cancel()
	}
	r.mu.Unlock()

	r.wg.Wait()
	log.Printf("All collectors stopped")
}

// runCollector runs the collection loop for a single collector.
func (r *Registry) runCollector(ctx context.Context, c Collector) {
	defer r.wg.Done()

	typ := c.Type()
	buffer := r.buffers[typ]
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot, err := c.Collect(ctx)
			if err != nil {
				log.Printf("Collector %s failed: %v", c.Name(), err)
				continue
			}
			buffer.Push(Snapshot{
				Timestamp: time.Now(),
				Data:      snapshot,
			})
		}
	}
}

// GetBuffer returns the ring buffer for the given metric type.
// Returns nil if no collector is registered for that type.
func (r *Registry) GetBuffer(typ v1.MetricType) *storage.RingBuffer[Snapshot] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.buffers[typ]
}

// GetCollector returns the collector for the given metric type.
// Returns nil if not found.
func (r *Registry) GetCollector(typ v1.MetricType) Collector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.collectors[typ]
}

// ListCollectors returns a slice of all registered metric types.
func (r *Registry) ListCollectors() []v1.MetricType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]v1.MetricType, 0, len(r.collectors))
	for typ := range r.collectors {
		types = append(types, typ)
	}
	return types
}
