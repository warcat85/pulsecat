package app

import (
	"context"
	"pulsecat/internal/collector"
	"pulsecat/internal/storage"
	"runtime/metrics"
)

type StoreConsumer struct {
	collector.Consumer
	store storage.Storage
}

func (c StoreConsumer) Consume(ctx context.Context, sample metrics.Sample) error {
	return c.store.Store(ctx, sample)
}
