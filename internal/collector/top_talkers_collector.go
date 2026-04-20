package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents a network talker identified by protocol and port.
type ProtocolTalker struct {
	Protocol string
	Port     uint32
}

// represents a single network talker in internal format.
type NetworkTalker struct {
	Protocol       *ProtocolTalker
	BytesPerSecond uint64
	Percentage     float64
}

// represents a collection of network talkers in internal format.
type NetworkTalkers struct {
	Talkers []*NetworkTalker
}

// a placeholder collector that returns simulated top talkers data.
type DummyTopTalkersCollector struct{}

// creates a new dummy top talkers collector.
func NewDummyTopTalkersCollector() Collector {
	return &DummyTopTalkersCollector{}
}

// returns the metric type for top talkers.
func (c *DummyTopTalkersCollector) Type() metrics.MetricType {
	return metrics.TOP_TALKERS
}

// returns a human-readable name for this collector.
func (c *DummyTopTalkersCollector) Name() string {
	return "dummy_top_talkers"
}

// returns a simulated top talkers snapshot.
// The data matches the logic in server.CollectStatistics.
func (c *DummyTopTalkersCollector) Collect(ctx context.Context) (any, error) {
	now := time.Now()
	second := now.Second()
	return &NetworkTalkers{
		Talkers: []*NetworkTalker{
			{
				Protocol: &ProtocolTalker{
					Protocol: "TCP",
					Port:     80,
				},
				BytesPerSecond: 100000 + uint64(second%10000),
				Percentage:     80.0 + float64(second%5),
			},
		},
	}, nil
}
