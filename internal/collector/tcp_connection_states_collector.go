package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents TCP connection states data in internal format.
type TcpConnectionStates struct {
	Established uint32
	Listen      uint32
}

// a placeholder collector that returns simulated TCP connection states data.
type DummyTcpConnectionStatesCollector struct{}

// creates a new dummy TCP connection states collector.
func NewDummyTcpConnectionStatesCollector() Collector {
	return &DummyTcpConnectionStatesCollector{}
}

// returns the metric type for TCP connection states.
func (c *DummyTcpConnectionStatesCollector) Type() metrics.MetricType {
	return metrics.TCP_CONNECTION_STATES
}

// returns a human-readable name for this collector.
func (c *DummyTcpConnectionStatesCollector) Name() string {
	return "dummy_tcp_connection_states"
}

// returns a simulated TCP connection states snapshot.
// The data matches the logic in server.CollectStatistics.
func (c *DummyTcpConnectionStatesCollector) Collect(ctx context.Context) (any, error) {
	now := time.Now()
	second := now.Second()
	return &TcpConnectionStates{
		Established: 10 + uint32(second%5),
		Listen:      5,
	}, nil
}
