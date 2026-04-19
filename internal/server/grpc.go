package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"pulsecat/internal/config"
	"sync"
	"time"

	v1 "pulsecat/pkg/api/v1"

	"google.golang.org/grpc"
)

// Server implements the PulseCatServer interface.
type Server struct {
	v1.UnimplementedPulseCatServer
	config *config.Config

	mu sync.RWMutex

	// Individual metric data (simulated)
	loadAverage         *v1.LoadAverage
	cpuUsage            *v1.CpuUsage
	diskUsages          []*v1.DiskUsage
	networkStats        *v1.NetworkStats
	topTalkers          []*v1.NetworkTalker
	listeningSockets    []*v1.ListeningSocket
	tcpConnectionStates *v1.TcpConnectionStates
}

// New creates a new gRPC server instance.
func New(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

// Subscribe implements the gRPC streaming endpoint.
func (s *Server) Subscribe(req *v1.SubscribeRequest, stream v1.PulseCat_SubscribeServer) error {
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

// createMetricPulse creates a MetricPulse message for the requested metric type.
func (s *Server) createMetricPulse(metricType v1.MetricType) (*v1.MetricPulse, error) {
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

// CollectStatistics collects system statistics and updates server fields (simulated).
func (s *Server) CollectStatistics() {
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

// StartStatsCollection starts periodic statistics collection.
func (s *Server) StartStatsCollection(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.CollectStatistics()

			if s.config.LogLevel == "debug" {
				log.Printf("Updated system statistics at %v", time.Now().Format(time.RFC3339))
			}

		case <-ctx.Done():
			log.Println("Stopping statistics collection")
			return
		}
	}
}

func (s *Server) Run(stopCh chan struct{}) error {
	grpcServer := grpc.NewServer()
	v1.RegisterPulseCatServer(grpcServer, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.Port, err)
	}

	// Start gRPC server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("gRPC server listening on port %d", s.config.Port)
		if err := grpcServer.Serve(lis); err != nil {
			serverErr <- err
		}
	}()

	// Wait for stop signal or server error
	select {
	case <-stopCh:
		log.Println("Received stop signal")
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
		return err
	}

	log.Println("Initiating graceful shutdown...")
	grpcServer.GracefulStop()
	return nil
}
