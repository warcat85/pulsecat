package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents CPU usage data in internal format.
type CPUUsage struct {
	User   float64
	System float64
	Idle   float64
}

// a placeholder collector that returns simulated CPU usage data.
type DummyCPUUsageCollector struct{}

// creates a new dummy CPU usage collector.
func NewDummyCPUUsageCollector() Collector {
	return &DummyCPUUsageCollector{}
}

// returns the metric type for CPU usage.
func (c *DummyCPUUsageCollector) Type() metrics.MetricType {
	return metrics.CPUUsage
}

// returns a human-readable name for this collector.
func (c *DummyCPUUsageCollector) Name() string {
	return "dummy_cpu_usage"
}

// returns a simulated CPU usage snapshot.
func (c *DummyCPUUsageCollector) Collect(_ context.Context) (metrics.Sample, error) {
	now := time.Now()
	second := now.Second()
	return &CPUUsage{
		User:   10.5 + float64(second%10),
		System: 5.2 + float64(second%5),
		Idle:   84.3 - float64(second%15),
	}, nil
}
