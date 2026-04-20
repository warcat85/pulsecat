package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents TCP connection states data in internal format.
type TCPConnectionStates struct {
	Established uint32
	Listen      uint32
}

// a placeholder collector that returns simulated TCP connection states data.
type DummyTCPConnectionStatesCollector struct{}

// creates a new dummy TCP connection states collector.
func NewDummyTCPConnectionStatesCollector() Collector {
	return &DummyTCPConnectionStatesCollector{}
}

// returns the metric type for TCP connection states.
func (c *DummyTCPConnectionStatesCollector) Type() metrics.MetricType {
	return metrics.TCPConnectionStates
}

// returns a human-readable name for this collector.
func (c *DummyTCPConnectionStatesCollector) Name() string {
	return "dummy_tcp_connection_states"
}

// returns a simulated TCP connection states snapshot.
func (c *DummyTCPConnectionStatesCollector) Collect(_ context.Context) (metrics.Sample, error) {
	now := time.Now()
	second := now.Second()
	return &TCPConnectionStates{
		Established: 10 + uint32(second%5), //nolint:gosec // cannot be over 5
		Listen:      5,
	}, nil
}
