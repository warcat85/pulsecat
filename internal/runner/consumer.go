package runner

import (
	"context"
	"pulsecat/internal/metrics"
)

type Consumer interface {
	Consume(ctx context.Context, sample metrics.Sample) error
}

type DummyConsumer struct{}

func NewDummyConsumer() *DummyConsumer {
	return &DummyConsumer{}
}

func (c *DummyConsumer) Consume(_ context.Context, _ metrics.Sample) error {
	return nil
}
