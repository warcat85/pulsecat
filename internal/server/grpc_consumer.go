package server

import (
	"context"
	"fmt"
	"log"
	"pulsecat/internal/metrics"
	v1 "pulsecat/pkg/api/v1"
)

type GRPCConsumer struct {
	stream    v1.PulseCat_SubscribeServer
	converter Converter
	count     int
	logLevel  string
}

func NewGRPCConsumer(stream v1.PulseCat_SubscribeServer, converter Converter, logLevel string) *GRPCConsumer {
	return &GRPCConsumer{
		stream:    stream,
		converter: converter,
		logLevel:  logLevel,
	}
}

func (c *GRPCConsumer) Consume(_ context.Context, sample metrics.Sample) error {
	pulse, err := c.converter.Convert(sample)
	c.count++
	snapshotCount := c.count
	if err != nil {
		return fmt.Errorf("failed to create metric pulse for snapshot #%d: %w", snapshotCount, err)
	}

	if err := c.stream.Send(pulse); err != nil {
		log.Printf("Failed to send snapshot #%d: %v", snapshotCount, err)
		return err
	}

	if c.logLevel == "debug" {
		log.Printf("Sent snapshot #%d to client", snapshotCount)
	}
	return nil
}

func (c *GRPCConsumer) SnapshotCount() int {
	return c.count
}
