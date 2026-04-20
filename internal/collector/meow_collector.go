package collector

import (
	"context"
	"pulsecat/internal/metrics"
	v1 "pulsecat/pkg/api/v1"
)

// MeowCollector is a collector that returns a Meow message.
type MeowCollector struct{}

// NewMeowCollector creates a new MeowCollector.
func NewMeowCollector() *MeowCollector {
	return &MeowCollector{}
}

// Type returns the metric type for Meow.
func (c *MeowCollector) Type() metrics.MetricType {
	return metrics.MEOW
}

// Name returns the human-readable name.
func (c *MeowCollector) Name() string {
	return metrics.MEOW.String()
}

// Collect returns a Meow message.
func (c *MeowCollector) Collect(ctx context.Context) (any, error) {
	return &v1.Meow{
		Message: "Meow!",
	}, nil
}
