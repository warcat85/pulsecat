package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	v1 "pulsecat/pkg/api/v1"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Version information set by build flags
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

// Default configuration values
const (
	DefaultPort     = 25225
	DefaultLogLevel = "info"
)

// Config holds the application configuration
type Config struct {
	Port     int    // Port for gRPC server
	LogLevel string // Log level (debug, info, warn, error)

	// Monitor enable/disable settings
	Monitors MonitorsConfig
}

// MonitorsConfig holds enable/disable settings for each monitor type
type MonitorsConfig struct {
	LoadAverage         bool `yaml:"load_average"`
	CPUUsage            bool `yaml:"cpu_usage"`
	DiskUsage           bool `yaml:"disk_usage"`
	NetworkStats        bool `yaml:"network_stats"`
	TopTalkers          bool `yaml:"top_talkers"`
	ListeningSockets    bool `yaml:"listening_sockets"`
	TCPConnectionStates bool `yaml:"tcp_connection_states"`
}

// server implements the PulseCatServer interface
type server struct {
	v1.UnimplementedPulseCatServer
	config *Config
	mu     sync.RWMutex

	// Individual metric data
	loadAverage         *v1.LoadAverage
	cpuUsage            *v1.CpuUsage
	diskUsages          []*v1.DiskUsage
	networkStats        *v1.NetworkStats
	topTalkers          []*v1.NetworkTalker
	listeningSockets    []*v1.ListeningSocket
	tcpConnectionStates *v1.TcpConnectionStates
}

// Subscribe implements the gRPC streaming endpoint
func (s *server) Subscribe(req *v1.SubscribeRequest, stream v1.PulseCat_SubscribeServer) error {
	ctx := stream.Context()

	// Validate request parameters
	if req.MetricType == v1.MetricType_METRIC_TYPE_UNSPECIFIED {
		return fmt.Errorf("metric_type must be specified")
	}

	// Use request parameters directly (0 means no delay / single snapshot)
	startDelay := int(req.StartDelay)
	frequency := int(req.Frequency)

	log.Printf("New subscription: start_delay=%ds, frequency=%ds, metric_type=%v",
		startDelay, frequency, req.MetricType)

	// Wait for initial delay
	if startDelay > 0 {
		log.Printf("Waiting %d seconds before first snapshot for client", startDelay)
		select {
		case <-time.After(time.Duration(startDelay) * time.Second):
			log.Printf("Initial delay completed, starting stream")
		case <-ctx.Done():
			log.Printf("Stream cancelled during initial delay")
			return ctx.Err()
		}
	}

	// If frequency is 0, send a single snapshot and return
	if frequency == 0 {
		pulse, err := s.createMetricPulse(req.MetricType)
		if err != nil {
			return err
		}
		if err := stream.Send(pulse); err != nil {
			log.Printf("Failed to send single snapshot: %v", err)
			return err
		}
		log.Printf("Sent single snapshot for metric_type=%v", req.MetricType)
		return nil
	}

	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	defer ticker.Stop()

	snapshotCount := 0
	for {
		select {
		case <-ticker.C:
			snapshotCount++

			pulse, err := s.createMetricPulse(req.MetricType)
			if err != nil {
				log.Printf("Failed to create metric pulse for snapshot #%d: %v", snapshotCount, err)
				continue
			}

			// Send metric pulse to client
			if err := stream.Send(pulse); err != nil {
				log.Printf("Failed to send snapshot #%d: %v", snapshotCount, err)
				return err
			}

			if s.config.LogLevel == "debug" {
				log.Printf("Sent snapshot #%d to client", snapshotCount)
			}

		case <-ctx.Done():
			log.Printf("Subscription ended after %d snapshots", snapshotCount)
			return ctx.Err()
		}
	}
}

// createMetricPulse creates a MetricPulse message for the requested metric type
func (s *server) createMetricPulse(metricType v1.MetricType) (*v1.MetricPulse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pulse := &v1.MetricPulse{
		Timestamp: time.Now().Unix(),
	}

	switch metricType {
	case v1.MetricType_METRIC_TYPE_MEOW:
		pulse.Metric = &v1.MetricPulse_Meow{
			Meow: &v1.Meow{
				Message: "Meow!",
			},
		}
	case v1.MetricType_METRIC_TYPE_LOAD_AVERAGE:
		if s.loadAverage == nil {
			return nil, fmt.Errorf("load average data not available")
		}
		pulse.Metric = &v1.MetricPulse_LoadAverage{
			LoadAverage: s.loadAverage,
		}
	case v1.MetricType_METRIC_TYPE_CPU_USAGE:
		if s.cpuUsage == nil {
			return nil, fmt.Errorf("CPU usage data not available")
		}
		pulse.Metric = &v1.MetricPulse_CpuUsage{
			CpuUsage: s.cpuUsage,
		}
	case v1.MetricType_METRIC_TYPE_DISK_USAGE:
		if s.diskUsages == nil {
			return nil, fmt.Errorf("disk usage data not available")
		}
		pulse.Metric = &v1.MetricPulse_DiskUsage{
			DiskUsage: &v1.DiskUsages{
				Disks: s.diskUsages,
			},
		}
	case v1.MetricType_METRIC_TYPE_NETWORK_STATS:
		if s.networkStats == nil {
			return nil, fmt.Errorf("network stats data not available")
		}
		pulse.Metric = &v1.MetricPulse_NetworkStats{
			NetworkStats: s.networkStats,
		}
	case v1.MetricType_METRIC_TYPE_TOP_TALKERS:
		if s.topTalkers == nil {
			return nil, fmt.Errorf("top talkers data not available")
		}
		pulse.Metric = &v1.MetricPulse_TopTalkers{
			TopTalkers: &v1.NetworkTalkers{
				Talkers: s.topTalkers,
			},
		}
	case v1.MetricType_METRIC_TYPE_LISTENING_SOCKETS:
		if s.listeningSockets == nil {
			return nil, fmt.Errorf("listening sockets data not available")
		}
		pulse.Metric = &v1.MetricPulse_ListeningSockets{
			ListeningSockets: &v1.ListeningSockets{
				Sockets: s.listeningSockets,
			},
		}
	case v1.MetricType_METRIC_TYPE_TCP_CONNECTION_STATES:
		if s.tcpConnectionStates == nil {
			return nil, fmt.Errorf("TCP connection states data not available")
		}
		pulse.Metric = &v1.MetricPulse_TcpConnectionStates{
			TcpConnectionStates: s.tcpConnectionStates,
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %v", metricType)
	}

	return pulse, nil
}

// collectStatistics collects system statistics and updates server fields
func (s *server) collectStatistics() {
	// TODO: Implement actual system statistics collection
	// For now, generate placeholder data with some variation to simulate real data

	now := time.Now()
	second := now.Second()

	// Simulate some data variation
	baseLoad := 0.1 + float64(second%30)*0.01

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update individual metric fields
	s.loadAverage = &v1.LoadAverage{
		OneMin:     baseLoad,
		FiveMin:    baseLoad * 0.9,
		FifteenMin: baseLoad * 0.8,
	}

	s.cpuUsage = &v1.CpuUsage{
		User:   10.5 + float64(second%10),
		System: 5.2 + float64(second%5),
		Idle:   84.3 - float64(second%15),
	}

	s.diskUsages = []*v1.DiskUsage{
		{
			Filesystem:  "/dev/sda1",
			TotalMb:     102400,
			UsedMb:      51200 + uint64(second%100),
			UsedPercent: 50.0 + float64(second%10)*0.1,
			MountPoint:  "/",
		},
	}

	s.networkStats = &v1.NetworkStats{
		TotalBytesReceived: 1000000 + uint64(second%1000)*1000,
		TotalBytesSent:     500000 + uint64(second%500)*1000,
	}

	s.topTalkers = []*v1.NetworkTalker{
		{
			Identifier: &v1.NetworkTalker_Protocol{
				Protocol: &v1.ProtocolTalker{
					Protocol: "TCP",
					Port:     80,
				},
			},
			BytesPerSecond: 100000 + uint64(second%10000),
			Percentage:     80.0 + float64(second%5),
		},
	}

	s.listeningSockets = []*v1.ListeningSocket{
		{
			Command:  "sshd",
			Pid:      1234,
			User:     "root",
			Protocol: "tcp",
			Port:     22,
			Address:  "0.0.0.0",
		},
	}

	s.tcpConnectionStates = &v1.TcpConnectionStates{
		Established: 10 + uint32(second%5),
		Listen:      5,
	}

	if s.config.LogLevel == "debug" {
		log.Printf("Collected statistics: CPU User=%.1f%%, System=%.1f%%, Idle=%.1f%%",
			s.cpuUsage.User, s.cpuUsage.System, s.cpuUsage.Idle)
	}
}

// startStatsCollection starts periodic statistics collection
func (s *server) startStatsCollection(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collectStatistics()

			if s.config.LogLevel == "debug" {
				log.Printf("Updated system statistics at %v", time.Now().Format(time.RFC3339))
			}

		case <-ctx.Done():
			log.Println("Stopping statistics collection")
			return
		}
	}
}

// Daemon represents the system monitoring daemon
type Daemon struct {
	config     *Config
	grpcServer *grpc.Server
	server     *server
	stopCh     chan struct{}
	doneCh     chan struct{}
}

// NewDaemon creates a new daemon instance
func NewDaemon(config *Config) *Daemon {
	srv := &server{
		config: config,
	}

	return &Daemon{
		config: config,
		server: srv,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// Run starts the daemon
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

	// Create gRPC server
	d.grpcServer = grpc.NewServer()
	v1.RegisterPulseCatServer(d.grpcServer, d.server)

	// Start TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", d.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", d.config.Port, err)
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start statistics collection
	go d.server.startStatsCollection(ctx)

	// Start gRPC server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("gRPC server listening on port %d", d.config.Port)
		if err := d.grpcServer.Serve(lis); err != nil {
			serverErr <- err
		}
	}()

	// Wait for stop signal or server error
	select {
	case <-d.stopCh:
		log.Println("Received stop signal")
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
		return err
	}

	// Graceful shutdown
	log.Println("Initiating graceful shutdown...")
	d.grpcServer.GracefulStop()
	cancel()

	close(d.doneCh)
	return nil
}

// Stop gracefully stops the daemon
func (d *Daemon) Stop() {
	log.Println("Stopping daemon...")
	close(d.stopCh)
	<-d.doneCh
	log.Println("Daemon stopped")
}

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("Starting PulseCat daemon v%s (build: %s, commit: %s)",
		Version, BuildTime, CommitHash)

	// Parse command-line arguments
	var config Config
	var configFile string

	// Command-line flags (overrides config file)
	flag.IntVar(&config.Port, "port", 0, "Port for gRPC server (overrides config file)")
	flag.StringVar(&config.LogLevel, "log-level", "", "Log level (debug, info, warn, error) (overrides config file)")
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to YAML configuration file")

	// Add help flag
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	// Track which flags were explicitly set
	visitedFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		visitedFlags[f.Name] = true
	})

	// Load configuration from YAML file if it exists
	var yamlConfig *Config
	if _, err := os.Stat(configFile); err == nil {
		var err error
		yamlConfig, err = loadConfigFromYAML(configFile)
		if err != nil {
			log.Fatalf("Failed to load config from %s: %v", configFile, err)
		}
		log.Printf("Loaded configuration from %s", configFile)
	} else {
		log.Printf("Config file %s not found, using defaults and command-line flags", configFile)
	}

	// Merge YAML config with command-line flags (flags override YAML)
	if yamlConfig != nil {
		// Port: use flag value if visited, else YAML value
		if !visitedFlags["port"] {
			config.Port = yamlConfig.Port
		} else {
			log.Printf("Overriding port from config file with command-line value: %d", config.Port)
		}
		// LogLevel: use flag value if visited, else YAML value
		if !visitedFlags["log-level"] {
			config.LogLevel = yamlConfig.LogLevel
		} else {
			log.Printf("Overriding log level from config file with command-line value: %s", config.LogLevel)
		}
		// Monitors: always use YAML value (no flag for monitors)
		config.Monitors = yamlConfig.Monitors
	}

	// Set defaults if still zero
	if config.Port == 0 {
		config.Port = DefaultPort
	}
	if config.LogLevel == "" {
		config.LogLevel = DefaultLogLevel
	}

	// Initialize monitors with defaults if not set from YAML
	// (MonitorsConfig fields already have zero values which are false, but we want them true by default)
	// We need to check if monitors were loaded from YAML by checking if any monitor is set
	// A simple approach: if all monitors are false, set them all to true
	allFalse := !config.Monitors.LoadAverage && !config.Monitors.CPUUsage &&
		!config.Monitors.DiskUsage && !config.Monitors.NetworkStats &&
		!config.Monitors.TopTalkers && !config.Monitors.ListeningSockets &&
		!config.Monitors.TCPConnectionStates
	if allFalse {
		config.Monitors = MonitorsConfig{
			LoadAverage:         true,
			CPUUsage:            true,
			DiskUsage:           true,
			NetworkStats:        true,
			TopTalkers:          true,
			ListeningSockets:    true,
			TCPConnectionStates: true,
		}
		log.Printf("No monitor configuration found, enabling all monitors by default")
	}

	// Validate configuration
	if config.Port <= 0 || config.Port > 65535 {
		log.Fatal("Error: Port must be between 1 and 65535")
	}

	// Log configuration
	enabledMonitors := 0
	if config.Monitors.LoadAverage {
		enabledMonitors++
	}
	if config.Monitors.CPUUsage {
		enabledMonitors++
	}
	if config.Monitors.DiskUsage {
		enabledMonitors++
	}
	if config.Monitors.NetworkStats {
		enabledMonitors++
	}
	if config.Monitors.TopTalkers {
		enabledMonitors++
	}
	if config.Monitors.ListeningSockets {
		enabledMonitors++
	}
	if config.Monitors.TCPConnectionStates {
		enabledMonitors++
	}

	log.Printf("Configuration loaded: port=%d, log-level=%s, monitors=%d/7 enabled",
		config.Port, config.LogLevel, enabledMonitors)

	// Create daemon
	daemon := NewDaemon(&config)

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Start daemon in a goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Println("Daemon starting...")
		if err := daemon.Run(); err != nil {
			errCh <- fmt.Errorf("daemon failed: %w", err)
		} else {
			errCh <- nil
		}
	}()

	// Wait for either a signal or daemon error
	select {
	case sig := <-sigCh:
		log.Printf("Received signal: %v", sig)
		log.Println("Initiating graceful shutdown...")
		daemon.Stop()
		log.Println("Shutdown complete")
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Fatal error: %v", err)
		}
		log.Println("Daemon stopped normally")
	}

	log.Println("System monitor daemon exited")
}

// YAMLConfig represents the structure of the YAML configuration file
type YAMLConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
	Monitors MonitorsConfig `yaml:"monitors"`
	Advanced AdvancedConfig `yaml:"advanced,omitempty"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file,omitempty"`
}

type AdvancedConfig struct {
	CollectionInterval int      `yaml:"collection_interval,omitempty"`
	MaxTopTalkers      int      `yaml:"max_top_talkers,omitempty"`
	NetworkInterface   string   `yaml:"network_interface,omitempty"`
	ExcludeFilesystems []string `yaml:"exclude_filesystems,omitempty"`
}

// loadConfigFromYAML loads configuration from a YAML file
func loadConfigFromYAML(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var yamlConfig YAMLConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Convert YAML config to internal Config
	config := &Config{
		Port:     yamlConfig.Server.Port,
		LogLevel: yamlConfig.Logging.Level,
		Monitors: yamlConfig.Monitors,
	}

	// Set default values if not specified in YAML
	if config.Port == 0 {
		config.Port = DefaultPort
	}
	if config.LogLevel == "" {
		config.LogLevel = DefaultLogLevel
	}

	return config, nil
}

// Helper function to print usage
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}
