package app

import (
	"context"
	"log"
	"pulsecat/internal/collector"
	"pulsecat/internal/config"
	"pulsecat/internal/server"
	v1 "pulsecat/pkg/api/v1"
	"time"
)

// placeholderCollector is a dummy collector that logs collection events.
// It returns empty data; used for testing the registry.
type placeholderCollector struct {
	metricType v1.MetricType
	name       string
}

func (c *placeholderCollector) Type() v1.MetricType { return c.metricType }
func (c *placeholderCollector) Name() string        { return c.name }
func (c *placeholderCollector) Collect(ctx context.Context) (any, error) {
	log.Printf("Placeholder collector %s collecting", c.name)
	// Return an empty struct appropriate for the metric type.
	// This is just a placeholder; real collectors will return proper data.
	switch c.metricType {
	case v1.MetricType_METRIC_TYPE_LOAD_AVERAGE:
		return &v1.LoadAverage{}, nil
	case v1.MetricType_METRIC_TYPE_CPU_USAGE:
		return &v1.CpuUsage{}, nil
	case v1.MetricType_METRIC_TYPE_DISK_USAGE:
		return &v1.DiskUsages{}, nil
	case v1.MetricType_METRIC_TYPE_NETWORK_STATS:
		return &v1.NetworkStats{}, nil
	case v1.MetricType_METRIC_TYPE_TOP_TALKERS:
		return &v1.NetworkTalkers{}, nil
	case v1.MetricType_METRIC_TYPE_LISTENING_SOCKETS:
		return &v1.ListeningSockets{}, nil
	case v1.MetricType_METRIC_TYPE_TCP_CONNECTION_STATES:
		return &v1.TcpConnectionStates{}, nil
	case v1.MetricType_METRIC_TYPE_MEOW:
		return &v1.Meow{Message: "meow"}, nil
	default:
		return struct{}{}, nil
	}
}

// Daemon represents the system monitoring daemon
type Daemon struct {
	config   *config.Config
	server   *server.Server
	registry *collector.Registry
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func NewDaemon(config *config.Config) *Daemon {
	srv := server.New(config)

	// Compute buffer capacity based on collection interval and buffer duration
	interval := time.Duration(config.CollectionInterval) * time.Second
	bufferCap := config.BufferDuration / config.CollectionInterval
	if bufferCap < 1 {
		bufferCap = 1
	}
	registry := collector.NewRegistry(interval, bufferCap)

	// Register placeholder collectors for enabled monitors
	if config.Monitors.LoadAverage {
		registry.Register(&placeholderCollector{
			metricType: v1.MetricType_METRIC_TYPE_LOAD_AVERAGE,
			name:       "load_average",
		})
	}
	if config.Monitors.CPUUsage {
		registry.Register(&placeholderCollector{
			metricType: v1.MetricType_METRIC_TYPE_CPU_USAGE,
			name:       "cpu_usage",
		})
	}
	if config.Monitors.DiskUsage {
		registry.Register(&placeholderCollector{
			metricType: v1.MetricType_METRIC_TYPE_DISK_USAGE,
			name:       "disk_usage",
		})
	}
	if config.Monitors.NetworkStats {
		registry.Register(&placeholderCollector{
			metricType: v1.MetricType_METRIC_TYPE_NETWORK_STATS,
			name:       "network_stats",
		})
	}
	if config.Monitors.TopTalkers {
		registry.Register(&placeholderCollector{
			metricType: v1.MetricType_METRIC_TYPE_TOP_TALKERS,
			name:       "top_talkers",
		})
	}
	if config.Monitors.ListeningSockets {
		registry.Register(&placeholderCollector{
			metricType: v1.MetricType_METRIC_TYPE_LISTENING_SOCKETS,
			name:       "listening_sockets",
		})
	}
	if config.Monitors.TCPConnectionStates {
		registry.Register(&placeholderCollector{
			metricType: v1.MetricType_METRIC_TYPE_TCP_CONNECTION_STATES,
			name:       "tcp_connection_states",
		})
	}
	// Always register meow collector (it's a special metric)
	registry.Register(&placeholderCollector{
		metricType: v1.MetricType_METRIC_TYPE_MEOW,
		name:       "meow",
	})

	return &Daemon{
		config:   config,
		server:   srv,
		registry: registry,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

func (d *Daemon) Run() error {
	log.Printf("Starting PulseCat daemon")

	// Format monitor status
	monitorStatus := []string{}
	if d.config.Monitors.LoadAverage {
		monitorStatus = append(monitorStatus, "load_average")
	}
	if d.config.Monitors.CPUUsage {
		monitorStatus = append(monitorStatus, "cpu_usage")
	}
	if d.config.Monitors.DiskUsage {
		monitorStatus = append(monitorStatus, "disk_usage")
	}
	if d.config.Monitors.NetworkStats {
		monitorStatus = append(monitorStatus, "network_stats")
	}
	if d.config.Monitors.TopTalkers {
		monitorStatus = append(monitorStatus, "top_talkers")
	}
	if d.config.Monitors.ListeningSockets {
		monitorStatus = append(monitorStatus, "listening_sockets")
	}
	if d.config.Monitors.TCPConnectionStates {
		monitorStatus = append(monitorStatus, "tcp_connection_states")
	}

	log.Printf("Configuration: port=%d, log-level=%s",
		d.config.Port, d.config.LogLevel)
	log.Printf("Monitors enabled: %v", monitorStatus)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start collector registry
	d.registry.Start(ctx)

	// Start statistics collection
	go d.server.StartStatsCollection(ctx)

	if err := d.server.Run(d.stopCh); err != nil {
		return err
	}

	// Stop collector registry
	d.registry.Stop()

	cancel()
	close(d.doneCh)
	return nil
}

func (d *Daemon) Stop() {
	log.Println("Stopping daemon...")
	close(d.stopCh)
	<-d.doneCh
	log.Println("Daemon stopped")
}
