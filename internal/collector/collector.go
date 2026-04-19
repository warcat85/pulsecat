package collector

import (
	"context"
	"log"
	"pulsecat/internal/metrics"
	v1 "pulsecat/pkg/api/v1"
	"time"
)

type Collector interface {
	// the metric type this collector produces.
	Type() metrics.MetricType
	// human-readable name for the collector.
	Name() string
	// runs the collection loop
	Start()
	// stops the collection loop
	Stop()
	// performs a single collection cycle and returns a snapshot.
	// The returned data must be of the appropriate type
	Collect(context.Context) (any, error)
}

type PeriodicCollector struct {
	Collector
	context  context.Context
	cancel   context.CancelFunc
	interval time.Duration
	consumer Consumer
}

// runs the collection loop
func (c *PeriodicCollector) Start() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	context := c.context

	for {
		select {
		case <-context.Done():
			return
		case <-ticker.C:
			snapshot, err := c.Collect(context)
			if err != nil {
				log.Printf("Collector %s failed: %v", c.Name(), err)
				continue
			}
			err = c.consumer.Consume(context, metrics.Sample{
				Timestamp: time.Now(),
				Data:      snapshot,
			})
			if err != nil {
				log.Printf("Consumer failed: %v", c.Name(), err)
				continue
			}
		}
	}
}

func (c *PeriodicCollector) Stop() {
	c.cancel()
}

// a dummy collector that logs collection events.
// It returns empty data; used for testing the registry.
type DummyCollector struct {
	PeriodicCollector
	metricType metrics.MetricType
}

func NewDummyCollector(metricType metrics.MetricType) *DummyCollector {
	return &DummyCollector{
		metricType: metricType,
	}
}

func (c *DummyCollector) Type() metrics.MetricType { return c.metricType }
func (c *DummyCollector) Name() string             { return c.metricType.String() }
func (c *DummyCollector) Collect(context.Context) (any, error) {
	log.Printf("Placeholder collector %s collecting", c.Name())
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
	default:
		return struct{}{}, nil
	}
}
