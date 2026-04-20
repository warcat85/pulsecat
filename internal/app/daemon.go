package app

import (
	"context"
	"log"
	"pulsecat/internal/collector"
	"pulsecat/internal/config"
	"pulsecat/internal/metrics"
	"pulsecat/internal/runner"
	"pulsecat/internal/server"
)

// represents the system monitoring daemon
type Daemon struct {
	config  *config.Config
	server  *server.Server
	manager *runner.Manager
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func NewDaemon(config *config.Config) *Daemon {
	builder := NewBuilder(config)

	collectors, manager := builder.BuildComponents()
	collectors[metrics.MEOW] = collector.NewMeowCollector()
	srv := server.New(config, collectors)
	/*
		manager := collector.NewManager()

		// Register placeholder collectors for enabled monitors
		if config.Monitors.LoadAverage {
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
		}*/
	// Always register meow collector (it's a special metric)
	/*
		runner.Register(&collector.DummyCollector{
			metricType: collector.MEOW,
			name:       "meow",
		})*/

	return &Daemon{
		config:  config,
		server:  srv,
		manager: manager,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
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
	d.manager.Start(ctx)

	if err := d.server.Run(d.stopCh); err != nil {
		return err
	}

	// Stop collector registry
	d.manager.Stop()

	close(d.doneCh)
	return nil
}

func (d *Daemon) Stop() {
	log.Println("Stopping daemon...")
	close(d.stopCh)
	<-d.doneCh
	log.Println("Daemon stopped")
}
