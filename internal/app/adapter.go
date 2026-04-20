package app

import (
	"context"
	"pulsecat/internal/metrics"
	"pulsecat/internal/storage"
)

type StorageConsumer struct {
	storage.Storage
}

func NewStorageConsumer(s storage.Storage) *StorageConsumer {
	return &StorageConsumer{
		Storage: s,
	}
}

func (c StorageConsumer) Consume(ctx context.Context, sample metrics.Sample) error {
	return c.Store(ctx, sample)
}
