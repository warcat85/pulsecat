package storage

import (
	"context"
	"log"
	"pulsecat/internal/metrics"
)

type Storage interface {
	Store(ctx context.Context, sample metrics.Sample) error
}

// the collector with buffer
type BufferedStorage struct {
	Storage
	buffer *RingBuffer[metrics.Sample]
}

func NewBufferedStorage(capacity int) *BufferedStorage {
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

// DummyStorage prints every sample it receives to stdout.
type DummyStorage struct{}

func NewDummyStorage() *DummyStorage {
	return &DummyStorage{}
}

func (d *DummyStorage) Store(ctx context.Context, sample metrics.Sample) error {
	log.Printf("Dummy storage is storing: %+v\n", sample)
	return nil
}
