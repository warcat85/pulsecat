package runner

import (
	"context"
	"log"
	"pulsecat/internal/metrics"
	"sync"
)

// manages a set of collectors and their associated ring buffers.
type Manager struct {
	mu      sync.RWMutex
	runners metrics.MetricMap[*Runner]
	wg      sync.WaitGroup
}

// creates a new collector runner.
func NewManager() *Manager {
	return &Manager{
		runners: make(metrics.MetricMap[*Runner]),
	}
}

// adds a collector to the registry.
// If a collector for the same metric type already exists, it is replaced.
func (m *Manager) Register(r *Runner) {
	m.mu.Lock()
	defer m.mu.Unlock()

	kind := r.Collector().Type()
	m.runners[kind] = r
	log.Printf("Registered collector %s (%v)", r.Collector().Name(), kind)
}

// begins periodic collection for all registered collectors.
// Each collector runs in its own goroutine, collecting at the configured interval.
// The provided context can be used to stop all collectors.
func (m *Manager) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, runner := range m.runners {
		m.wg.Add(1)
		go func() {
			runner.Start(ctx)
			m.wg.Done()
		}()
	}
	log.Printf("Started %d collector(s)", len(m.runners))
}

// halts all collectors and waits for them to finish.
func (m *Manager) Stop() {
	m.mu.Lock()
	for _, runner := range m.runners {
		runner.Stop()
	}
	m.mu.Unlock()

	m.wg.Wait()
	log.Printf("All collectors stopped")
}

// returns the collector for the given metric type.
// Returns nil if not found.
func (m *Manager) GetRunner(kind metrics.MetricType) *Runner {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runners[kind]
}

// returns a slice of all registered metric types.
func (m *Manager) ListCollectors() []metrics.MetricType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	types := make([]metrics.MetricType, 0, len(m.runners))
	for kind := range m.runners {
		types = append(types, kind)
	}
	return types
}
