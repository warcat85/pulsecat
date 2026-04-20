package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"pulsecat/internal/collector/builder"
	"pulsecat/internal/config"
	v1 "pulsecat/pkg/api/v1"
	"time"

	"google.golang.org/grpc"
)

// implements the PulseCatServer interface.
type Server struct {
	v1.UnimplementedPulseCatServer
	config  *config.Config
	builder builder.CollectorBuilder
}

// creates a new gRPC server instance.
func New(config *config.Config, builder builder.CollectorBuilder) *Server {
	return &Server{
		config:  config,
		builder: builder,
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

	protoType := req.MetricType
	metricType, ok := ConvertMetricTypeFromProto(protoType)
	if !ok {
		return fmt.Errorf("unsupported metric type: %v", protoType)
	}
	c, err := s.builder.BuildCollector(metricType, startDelay)
	if err != nil {
		return err
	}
	converter, ok := converters[metricType]
	if !ok {
		return fmt.Errorf("converter is not registered for metric type: %v", metricType)
	}

	// If frequency is 0, send a single snapshot and return
	if frequency == 0 {
		data, err := c.Collect(context.Background())
		if err != nil {
			return fmt.Errorf("%s collection failed: %w", converter.Name(), err)
		}
		pulse, err := converter.Convert(data)
		if err != nil {
			return err
		}
		if err := stream.Send(pulse); err != nil {
			log.Printf("Failed to send single snapshot: %v", err)
			return err
		}
		log.Printf("Sent single snapshot for metric type=%s", metricType)
		return nil
	}

	// additional objects for pipeline
	consumer := NewGRPCConsumer(stream, converter, s.config.LogLevel)
	runner := s.builder.BuildRunner(c, consumer, frequency)
	runner.Start(ctx)

	log.Printf("Subscription ended after %d snapshots", consumer.SnapshotCount())
	return ctx.Err()
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
