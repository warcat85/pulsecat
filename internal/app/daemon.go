package app

import (
	"context"
	"log"
	"pulsecat/internal/collector"
	"pulsecat/internal/config"
	"pulsecat/internal/metrics"
	"pulsecat/internal/server"
	"time"
)

// represents the system monitoring daemon
type Daemon struct {
	config   *config.Config
	server   *server.Server
	registry *collector.Runner
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func NewDaemon(config *config.Config) *Daemon {
	srv := server.New(config)

	// Compute buffer capacity based on collection interval and buffer duration
	interval := time.Duration(config.CollectionInterval) * time.Second
	bufferCap := config.BufferDuration / config.CollectionInterval
	if bufferCap < 1 {
		panic("buffer capacity must be positive")
	}
	_ = interval
	_ = bufferCap

	runner := collector.NewRunner()

	// Register placeholder collectors for enabled monitors
	if config.Monitors.LoadAverage {
		runner.Register(collector.NewDummyCollector(metrics.LOAD_AVERAGE))
	}
	if config.Monitors.CPUUsage {
		runner.Register(collector.NewDummyCollector(metrics.CPU_USAGE))
	}
	if config.Monitors.DiskUsage {
		runner.Register(collector.NewDummyCollector(metrics.DISK_USAGE))
	}
	if config.Monitors.NetworkStats {
		runner.Register(collector.NewDummyCollector(metrics.NETWORK_STATS))
	}
	if config.Monitors.TopTalkers {
		runner.Register(collector.NewDummyCollector(metrics.TOP_TALKERS))
	}
	if config.Monitors.ListeningSockets {
		runner.Register(collector.NewDummyCollector(metrics.LISTENING_SOCKETS))
	}
	if config.Monitors.TCPConnectionStates {
		runner.Register(collector.NewDummyCollector(metrics.TCP_CONNECTION_STATES))
	}
	// Always register meow collector (it's a special metric)
	/*
		runner.Register(&collector.DummyCollector{
			metricType: collector.MEOW,
			name:       "meow",
		})*/

	return &Daemon{
		config:   config,
		server:   srv,
		registry: runner,
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

	if err := d.server.Run(d.stopCh); err != nil {
		return err
	}

	// Stop collector registry
	d.registry.Stop()

	close(d.doneCh)
	return nil
}

func (d *Daemon) Stop() {
	log.Println("Stopping daemon...")
	close(d.stopCh)
	<-d.doneCh
	log.Println("Daemon stopped")
}
