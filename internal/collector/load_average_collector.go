package collector

import (
	"context"
	"pulsecat/internal/metrics"
	"time"
)

// represents load average data in internal format.
type LoadAverage struct {
	OneMin     float64
	FiveMin    float64
	FifteenMin float64
}

// a placeholder collector that returns simulated load average data.
type DummyLoadAverageCollector struct{}

// creates a new dummy load average collector.
func NewDummyLoadAverageCollector() Collector {
	return &DummyLoadAverageCollector{}
}

// returns the metric type for load average.
func (c *DummyLoadAverageCollector) Type() metrics.MetricType {
	return metrics.LoadAverage
}

// returns a human-readable name for this collector.
func (c *DummyLoadAverageCollector) Name() string {
	return "dummy_load_average"
}

// returns a simulated load average snapshot.
func (c *DummyLoadAverageCollector) Collect(_ context.Context) (metrics.Sample, error) {
	now := time.Now()
	second := now.Second()
	baseLoad := 0.1 + float64(second%30)*0.01
	return &LoadAverage{
		OneMin:     baseLoad,
		FiveMin:    baseLoad * 0.9,
		FifteenMin: baseLoad * 0.8,
	}, nil
}
