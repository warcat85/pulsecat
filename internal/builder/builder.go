package builder

import (
	"pulsecat/internal/collector"
	cbuilder "pulsecat/internal/collector/builder"
	"pulsecat/internal/config"
	"pulsecat/internal/metrics"
	"pulsecat/internal/runner"
	"pulsecat/internal/storage"
	"time"
)

type Builder struct {
	config           *config.Config
	interval         time.Duration
	bufferCap        int
	collectorBuilder cbuilder.CollectorBuilder
}

func NewBuilder(config *config.Config) *Builder {
	// Compute buffer capacity based on collection interval and buffer duration
	interval := time.Duration(config.CollectionInterval) * time.Second
	bufferCap := config.BufferDuration / config.CollectionInterval
	if bufferCap < 1 {
		panic("buffer capacity must be positive")
	}
	return &Builder{
		config:           config,
		interval:         interval,
		bufferCap:        bufferCap,
		collectorBuilder: cbuilder.NewDataCollectorBuilder(),
	}
}

func (b *Builder) BuildComponents() (metrics.MetricMap[storage.Storage], *runner.Manager) {
	manager := runner.NewManager()
	storages := make(metrics.MetricMap[storage.Storage])
	for _, metricType := range b.metricTypes() {
		collector, err := b.collectorBuilder.BuildCollector(metricType, b.bufferCap)
		if err != nil {
			panic(err)
		}
		storage := storage.NewBufferedStorage(b.bufferCap)
		storages[metricType] = storage
		manager.Register(b.BuildRunner(collector, storage))
	}
	return storages, manager
}

func (b *Builder) metricTypes() []metrics.MetricType {
	types := make([]metrics.MetricType, 0, 7)
	if b.config.Monitors.LoadAverage {
		types = append(types, metrics.LoadAverage)
	}
	if b.config.Monitors.CPUUsage {
		types = append(types, metrics.CPUUsage)
	}
	if b.config.Monitors.DiskUsage {
		types = append(types, metrics.DiskUsage)
	}
	if b.config.Monitors.NetworkStats {
		types = append(types, metrics.NetworkStats)
	}
	if b.config.Monitors.TopTalkers {
		types = append(types, metrics.TopTalkers)
	}
	if b.config.Monitors.ListeningSockets {
		types = append(types, metrics.ListeningSockets)
	}
	if b.config.Monitors.TCPConnectionStates {
		types = append(types, metrics.TCPConnectionStates)
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
		c, err := b.collectorBuilder.BuildCollector(metricType, 0)
		if err != nil {
			panic(err)
		}
		storage := storage.NewBufferedStorage(b.bufferCap)
		manager.Register(b.BuildRunner(c, storage))
	}
	return manager
}

func (b *Builder) BuildRunner(c collector.Collector, storage storage.Storage) *runner.Runner {
	return runner.NewRunner(c, NewStorageConsumer(storage), b.interval)
}

func (b *Builder) BuildStorageConsumer() runner.Consumer {
	return NewStorageConsumer(storage.NewBufferedStorage(b.bufferCap))
}

func (b *Builder) BuildDummyStorageConsumer() runner.Consumer {
	return NewStorageConsumer(storage.NewDummyStorage())
}
