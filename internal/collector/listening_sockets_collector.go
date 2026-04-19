package collector

import (
	"context"
	"pulsecat/internal/metrics"
)

// represents a single listening socket in internal format.
type ListeningSocket struct {
	Command  string
	Pid      uint32
	User     string
	Protocol string
	Port     uint32
	Address  string
}

// represents a collection of listening sockets in internal format.
type ListeningSockets struct {
	Sockets []*ListeningSocket
}

// aplaceholder collector that returns simulated listening sockets data.
type DummyListeningSocketsCollector struct {
	PeriodicCollector
}

// creates a new dummy listening sockets collector.
func NewDummyListeningSocketsCollector() *DummyListeningSocketsCollector {
	return &DummyListeningSocketsCollector{}
}

// returns the metric type for listening sockets.
func (c *DummyListeningSocketsCollector) Type() metrics.MetricType {
	return metrics.LISTENING_SOCKETS
}

// returns a human-readable name for this collector.
func (c *DummyListeningSocketsCollector) Name() string {
	return "dummy_listening_sockets"
}

// returns a simulated listening sockets snapshot.
// The data matches the logic in server.CollectStatistics.
func (c *DummyListeningSocketsCollector) Collect(ctx context.Context) (any, error) {
	return &ListeningSockets{
		Sockets: []*ListeningSocket{
			{
				Command:  "sshd",
				Pid:      1234,
				User:     "root",
				Protocol: "tcp",
				Port:     22,
				Address:  "0.0.0.0",
			},
		},
	}, nil
}
