package collector

import (
	"context"
	"log"
	"pulsecat/internal/metrics"
	v1 "pulsecat/pkg/api/v1"
)

type Collector interface {
	// the metric type this collector produces.
	Type() metrics.MetricType
	// human-readable name for the collector.
	Name() string
	// performs a single collection cycle and returns a snapshot.
	// The returned data must be of the appropriate type
	Collect(context.Context) (any, error)
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
func (c *DummyCollector) Collect(context.Context) (any, error) {
	log.Printf("Dummy collector %s collecting", c.Name())
	// Return an empty struct appropriate for the metric type.
	// This is just a placeholder; real collectors will return proper data.
	switch c.metricType {
	case metrics.LOAD_AVERAGE:
		return &v1.LoadAverage{}, nil
	case metrics.CPU_USAGE:
		return &v1.CpuUsage{}, nil
	case metrics.DISK_USAGE:
		return &v1.DiskUsages{}, nil
	case metrics.NETWORK_STATS:
		return &v1.NetworkStats{}, nil
	case metrics.TOP_TALKERS:
		return &v1.NetworkTalkers{}, nil
	case metrics.LISTENING_SOCKETS:
		return &v1.ListeningSockets{}, nil
	case metrics.TCP_CONNECTION_STATES:
		return &v1.TcpConnectionStates{}, nil
	case metrics.MEOW:
		return &v1.Meow{Message: "Meow!"}, nil
	default:
		return struct{}{}, nil
	}
}
