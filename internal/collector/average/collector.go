package average

import (
	"context"
	"log"
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
	"pulsecat/internal/storage"
)

// a collector that calculates the average of a metric.
type Collector struct {
	metricType metrics.MetricType
	storage    storage.Storage
	numSamples int
	calculator metrics.AverageCalculator
}

func NewCollector(
	metricType metrics.MetricType, storage storage.Storage,
	calculator metrics.AverageCalculator, numSamples int,
) collector.Collector {
	return &Collector{
		metricType: metricType,
		storage:    storage,
		numSamples: numSamples,
		calculator: calculator,
	}
}

func (c *Collector) Type() metrics.MetricType { return c.metricType }
func (c *Collector) Name() string             { return c.metricType.String() }
func (c *Collector) Collect(ctx context.Context) (metrics.Sample, error) {
	log.Printf("Average collector returning average of %d samples\n", c.numSamples)
	samples := c.storage.Last(ctx, c.numSamples)
	return c.calculator(samples), nil
}
