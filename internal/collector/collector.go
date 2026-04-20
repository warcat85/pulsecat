package collector

import (
	"context"
	"log"
	"pulsecat/internal/metrics"
)

type Collector interface {
	// the metric type this collector produces.
	Type() metrics.MetricType
	// human-readable name for the collector.
	Name() string
	// performs a single collection cycle and returns a snapshot.
	// The returned data must be of the appropriate type
	Collect(context.Context) (metrics.Sample, error)
}

// a dummy collector that logs collection events.
// It returns empty data; used for testing the registry.
type DummyCollector struct {
	metricType metrics.MetricType
}

func NewDummyCollector(metricType metrics.MetricType) Collector {
	return &DummyCollector{
		metricType: metricType,
	}
}

func (c *DummyCollector) Type() metrics.MetricType { return c.metricType }
func (c *DummyCollector) Name() string             { return c.metricType.String() }
func (c *DummyCollector) Collect(context.Context) (metrics.Sample, error) {
	log.Printf("Dummy collector %s collecting", c.Name())
	// Return an empty struct appropriate for the metric type.
	// This is just a placeholder; real collectors will return proper data.
	switch c.metricType {
	case metrics.LoadAverage:
		return &LoadAverage{}, nil
	case metrics.CPUUsage:
		return &CPUUsage{}, nil
	case metrics.DiskUsage:
		return &DiskUsages{}, nil
	case metrics.NetworkStats:
		return &NetworkStats{}, nil
	case metrics.TopTalkers:
		return &NetworkTalkers{}, nil
	case metrics.ListeningSockets:
		return &ListeningSockets{}, nil
	case metrics.TCPConnectionStates:
		return &TCPConnectionStates{}, nil
	case metrics.Meow:
		return &Meow{Message: "Meow!"}, nil
	default:
		return struct{}{}, nil
	}
}
