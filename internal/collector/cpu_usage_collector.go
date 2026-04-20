package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents CPU usage data in internal format.
type CpuUsage struct {
	User   float64
	System float64
	Idle   float64
}

// a placeholder collector that returns simulated CPU usage data.
type DummyCpuUsageCollector struct{}

// creates a new dummy CPU usage collector.
func NewDummyCpuUsageCollector() Collector {
	return &DummyCpuUsageCollector{}
}

// returns the metric type for CPU usage.
func (c *DummyCpuUsageCollector) Type() metrics.MetricType {
	return metrics.CPU_USAGE
}

// returns a human-readable name for this collector.
func (c *DummyCpuUsageCollector) Name() string {
	return "dummy_cpu_usage"
}

// returns a simulated CPU usage snapshot.
// The data matches the logic in server.CollectStatistics.
func (c *DummyCpuUsageCollector) Collect(ctx context.Context) (any, error) {
	now := time.Now()
	second := now.Second()
	return &CpuUsage{
		User:   10.5 + float64(second%10),
		System: 5.2 + float64(second%5),
		Idle:   84.3 - float64(second%15),
	}, nil
}
