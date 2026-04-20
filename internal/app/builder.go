package app

import (
	"pulsecat/internal/collector"
	"pulsecat/internal/config"
	"pulsecat/internal/metrics"
	"pulsecat/internal/runner"
	"pulsecat/internal/storage"
	"time"
)

type Builder struct {
	config    *config.Config
	interval  time.Duration
	bufferCap int
}

type CollectorBuilder func() collector.Collector

var collectorBuilders = metrics.MetricMap[CollectorBuilder]{
	metrics.LOAD_AVERAGE:          collector.NewDummyLoadAverageCollector,
	metrics.CPU_USAGE:             collector.NewDummyCpuUsageCollector,
	metrics.DISK_USAGE:            collector.NewDummyDiskUsageCollector,
	metrics.NETWORK_STATS:         collector.NewDummyNetworkStatsCollector,
	metrics.TOP_TALKERS:           collector.NewDummyTopTalkersCollector,
	metrics.LISTENING_SOCKETS:     collector.NewDummyListeningSocketsCollector,
	metrics.TCP_CONNECTION_STATES: collector.NewDummyTcpConnectionStatesCollector,
}

func NewBuilder(config *config.Config) *Builder {
	// Compute buffer capacity based on collection interval and buffer duration
	interval := time.Duration(config.CollectionInterval) * time.Second
	bufferCap := config.BufferDuration / config.CollectionInterval
	if bufferCap < 1 {
		panic("buffer capacity must be positive")
	}
	return &Builder{
		config:    config,
		interval:  interval,
		bufferCap: bufferCap,
	}
}

func (b *Builder) BuildComponents() (metrics.MetricMap[collector.Collector], *runner.Manager) {
	manager := runner.NewManager()
	collectorsMap := make(metrics.MetricMap[collector.Collector])
	for _, metricType := range b.metricTypes() {
		builder, ok := collectorBuilders[metricType]
		if !ok {
			panic("unsupported metric type")
		}
		collector := builder()
		collectorsMap[metricType] = collector
		manager.Register(b.BuildRunner(collector))
	}
	return collectorsMap, manager
}

func (b *Builder) metricTypes() []metrics.MetricType {
	types := make([]metrics.MetricType, 0, 7)
	if b.config.Monitors.LoadAverage {
		types = append(types, metrics.LOAD_AVERAGE)
	}
	if b.config.Monitors.CPUUsage {
		types = append(types, metrics.CPU_USAGE)
	}
	if b.config.Monitors.DiskUsage {
		types = append(types, metrics.DISK_USAGE)
	}
	if b.config.Monitors.NetworkStats {
		types = append(types, metrics.NETWORK_STATS)
	}
	if b.config.Monitors.TopTalkers {
		types = append(types, metrics.TOP_TALKERS)
	}
	if b.config.Monitors.ListeningSockets {
		types = append(types, metrics.LISTENING_SOCKETS)
	}
	if b.config.Monitors.TCPConnectionStates {
		types = append(types, metrics.TCP_CONNECTION_STATES)
	}
	return types
}

func (b *Builder) BuildDummyRunner(metricType metrics.MetricType) *runner.Runner {
	return runner.NewRunner(
		collector.NewDummyCollector(metricType),
		runner.NewDummyConsumer(),
		b.interval)
}

func (b *Builder) BuildManager() *runner.Manager {
	manager := runner.NewManager()
	// Register placeholder collectors for enabled monitors
	for _, metricType := range b.metricTypes() {
		c := collectorBuilders[metricType]()
		manager.Register(b.BuildRunner(c))
	}
	return manager
}

func (b *Builder) BuildRunner(c collector.Collector) *runner.Runner {
	return runner.NewRunner(c, b.BuildStorageConsumer(), b.interval)
}

/*
		case b.config.Monitors.LoadAverage {
	}
	if b.config.Monitors.LoadAverage {
		manager.Register(builder.BuildDummyRunner(metrics.LOAD_AVERAGE))
	}
	if config.Monitors.CPUUsage {
		manager.Register(builder.BuildDummyRunner(metrics.CPU_USAGE))
	}
	if config.Monitors.DiskUsage {
		manager.Register(builder.BuildDummyRunner(metrics.DISK_USAGE))
	}
	if config.Monitors.NetworkStats {
		manager.Register(builder.BuildDummyRunner(metrics.NETWORK_STATS))
	}
	if config.Monitors.TopTalkers {
		manager.Register(builder.BuildDummyRunner(metrics.TOP_TALKERS))
	}
	if config.Monitors.ListeningSockets {
		manager.Register(builder.BuildDummyRunner(metrics.LISTENING_SOCKETS))
	}
	if config.Monitors.TCPConnectionStates {
		manager.Register(builder.BuildDummyRunner(metrics.TCP_CONNECTION_STATES))
	}
	return collector.NewRunner(
		collector.NewCollector(metricType),
		b.interval,
		collector.NewConsumer(b.bufferCap))

	return collector.NewManager()
}*/

func (b *Builder) BuildStorageConsumer() runner.Consumer {
	return NewStorageConsumer(storage.NewBufferedStorage(b.bufferCap))
}

func (b *Builder) BuildDummyStorageConsumer() runner.Consumer {
	return NewStorageConsumer(storage.NewDummyStorage())
}
