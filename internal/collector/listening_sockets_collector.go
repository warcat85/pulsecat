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

// a placeholder collector that returns simulated listening sockets data.
type DummyListeningSocketsCollector struct{}

// creates a new dummy listening sockets collector.
func NewDummyListeningSocketsCollector() Collector {
	return &DummyListeningSocketsCollector{}
}

// returns the metric type for listening sockets.
func (c *DummyListeningSocketsCollector) Type() metrics.MetricType {
	return metrics.ListeningSockets
}

// returns a human-readable name for this collector.
func (c *DummyListeningSocketsCollector) Name() string {
	return "dummy_listening_sockets"
}

// returns a simulated listening sockets snapshot.
func (c *DummyListeningSocketsCollector) Collect(_ context.Context) (metrics.Sample, error) {
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
