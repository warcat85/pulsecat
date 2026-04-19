package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents network statistics data in internal format.
type NetworkStats struct {
	TotalBytesReceived uint64
	TotalBytesSent     uint64
}

// aplaceholder collector that returns simulated network statistics.
type DummyNetworkStatsCollector struct {
	PeriodicCollector
}

// creates a new dummy network stats collector.
func NewDummyNetworkStatsCollector() *DummyNetworkStatsCollector {
	return &DummyNetworkStatsCollector{}
}

// returns the metric type for network stats.
func (c *DummyNetworkStatsCollector) Type() metrics.MetricType {
	return metrics.NETWORK_STATS
}

// returns a human-readable name for this collector.
func (c *DummyNetworkStatsCollector) Name() string {
	return "dummy_network_stats"
}

// returns a simulated network stats snapshot.
// The data matches the logic in server.CollectStatistics.
func (c *DummyNetworkStatsCollector) Collect(ctx context.Context) (any, error) {
	now := time.Now()
	second := now.Second()
	return &NetworkStats{
		TotalBytesReceived: 1000000 + uint64(second%1000)*1000,
		TotalBytesSent:     500000 + uint64(second%500)*1000,
	}, nil
}
