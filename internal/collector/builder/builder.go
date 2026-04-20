package builder

import (
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
	"pulsecat/internal/runner"
	"time"
)

var collectorBuilders = metrics.MetricMap[func() collector.Collector]{
	metrics.LoadAverage:         collector.NewDummyLoadAverageCollector,
	metrics.CPUUsage:            collector.NewDummyCPUUsageCollector,
	metrics.DiskUsage:           collector.NewDummyDiskUsageCollector,
	metrics.NetworkStats:        collector.NewDummyNetworkStatsCollector,
	metrics.TopTalkers:          collector.NewDummyTopTalkersCollector,
	metrics.ListeningSockets:    collector.NewDummyListeningSocketsCollector,
	metrics.TCPConnectionStates: collector.NewDummyTCPConnectionStatesCollector,
	metrics.Meow:                collector.NewMeowCollector,
}

type CollectorBuilder interface {
	BuildCollector(metricType metrics.MetricType, numSamples int) (collector.Collector, error)
	BuildRunner(c collector.Collector, consumer runner.Consumer, frequency int) *runner.Runner
}
type DataCollectorBuilder struct{}

func NewDataCollectorBuilder() CollectorBuilder {
	return &DataCollectorBuilder{}
}

func (b *DataCollectorBuilder) BuildCollector(
	metricType metrics.MetricType, _ int,
) (collector.Collector, error) {
	builder, ok := collectorBuilders[metricType]
	if !ok {
		return nil, metrics.ErrCollectorDisabled(metricType)
	}
	return builder(), nil
}

func (b *DataCollectorBuilder) BuildRunner(
	c collector.Collector, consumer runner.Consumer, frequency int,
) *runner.Runner {
	interval := time.Duration(frequency) * time.Second
	return runner.NewRunner(c, consumer, interval)
}
