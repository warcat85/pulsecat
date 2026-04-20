package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"pulsecat/internal/collector"
	"pulsecat/internal/config"
	"pulsecat/internal/metrics"
	"sync"
	"time"

	v1 "pulsecat/pkg/api/v1"

	"google.golang.org/grpc"
)

// implements the PulseCatServer interface.
type Server struct {
	v1.UnimplementedPulseCatServer
	config     *config.Config
	collectors metrics.MetricMap[collector.Collector]

	mu sync.RWMutex
}

// creates a new gRPC server instance.
func New(config *config.Config, collectors metrics.MetricMap[collector.Collector]) *Server {
	return &Server{
		config:     config,
		collectors: collectors,
	}
}

// implements the gRPC streaming endpoint.
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
func (s *Server) createMetricPulse(protoType v1.MetricType) (*v1.MetricPulse, error) {
	metricType, ok := ConvertMetricTypeFromProto(protoType)
	if !ok {
		return nil, fmt.Errorf("unsupported metric type: %v", protoType)
	}

	current, ok := s.collectors[metricType]
	if !ok {
		return nil, fmt.Errorf("collector is disabled for metric type: %v", metricType)
	}
	converter, ok := converters[metricType]
	if !ok {
		return nil, fmt.Errorf("converter is not registered for metric type: %v", metricType)
	}
	data, err := current.Collect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s collection failed: %w", converter.Name(), err)
	}

	return converter.Convert(data)
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
