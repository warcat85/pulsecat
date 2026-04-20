package collector

import (
	"context"
	"pulsecat/internal/metrics"
)

type Meow struct {
	Message string
}

// MeowCollector is a collector that returns a Meow message.
type MeowCollector struct{}

// NewMeowCollector creates a new MeowCollector.
func NewMeowCollector() Collector {
	return &MeowCollector{}
}

// Type returns the metric type for Meow.
func (c *MeowCollector) Type() metrics.MetricType {
	return metrics.Meow
}

// Name returns the human-readable name.
func (c *MeowCollector) Name() string {
	return metrics.Meow.String()
}

// Collect returns a Meow message.
func (c *MeowCollector) Collect(_ context.Context) (metrics.Sample, error) {
	return &Meow{
		Message: "Meow!",
	}, nil
}
