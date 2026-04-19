package storage

import (
	"context"
	"runtime/metrics"
)

type Storage interface {
	Store(ctx context.Context, sample metrics.Sample) error
}

// the collector with buffer
type BufferedStorage struct {
	Storage
	buffer *RingBuffer[metrics.Sample]
}

func NewBufferedStore(capacity int) *BufferedStorage {
	return &BufferedStorage{
		buffer: NewRingBuffer[metrics.Sample](capacity),
	}
}

func (c *BufferedStorage) Store(ctx context.Context, sample metrics.Sample) error {
	c.buffer.Push(sample)
	return nil
}

func (c *BufferedStorage) Buffer() *RingBuffer[metrics.Sample] {
	return c.buffer
}

type DummyStorage struct {
	Storage
}

func (c *DummyStorage) Store(ctx context.Context, sample metrics.Sample) error {
	return nil
}
