package average

import (
	"fmt"
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
	"pulsecat/internal/runner"
	"pulsecat/internal/storage"
	"time"
)

var calculators = metrics.MetricMap[metrics.AverageCalculator]{
	metrics.LoadAverage:         CalculateLoadAverage,
	metrics.CPUUsage:            CalculateCPUUsage,
	metrics.NetworkStats:        CalculateNetworkStats,
	metrics.TCPConnectionStates: CalculateTCPConnectionStates,
}

type CollectorBuilder struct {
	storages metrics.MetricMap[storage.Storage]
}

func NewCollectorBuilder(storages metrics.MetricMap[storage.Storage]) *CollectorBuilder {
	return &CollectorBuilder{
		storages: storages,
	}
}

func (b *CollectorBuilder) BuildCollector(
	metricType metrics.MetricType, numSamples int,
) (collector.Collector, error) {
	storage, ok := b.storages[metricType]
	if !ok {
		return nil, fmt.Errorf("unable to find storage for metric type %s", metricType)
	}
	calculator, ok := calculators[metricType]
	if !ok {
		return nil, metrics.ErrCollectorDisabled(metricType)
	}
	return NewCollector(metricType, storage, calculator, numSamples), nil
}

func (b *CollectorBuilder) BuildRunner(
	c collector.Collector, consumer runner.Consumer, frequency int,
) *runner.Runner {
	interval := time.Duration(frequency) * time.Second
	return runner.NewRunner(c, consumer, interval)
}
