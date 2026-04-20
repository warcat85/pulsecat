package storage

import (
	"context"
	"log"
	"pulsecat/internal/metrics"
)

type Storage interface {
	Store(ctx context.Context, sample metrics.Sample) error
	Last(ctx context.Context, numSamples int) metrics.Samples
}

// the storage with buffer.
type BufferedStorage struct {
	Storage
	buffer *RingBuffer[metrics.Sample]
}

func NewBufferedStorage(capacity int) *BufferedStorage {
	return &BufferedStorage{
		buffer: NewRingBuffer[metrics.Sample](capacity),
	}
}

func (c *BufferedStorage) Store(_ context.Context, sample metrics.Sample) error {
	c.buffer.Push(sample)
	return nil
}

func (c *BufferedStorage) Last(_ context.Context, numSamples int) metrics.Samples {
	return c.buffer.Last(numSamples)
}

// DummyStorage prints every sample it receives to stdout.
type DummyStorage struct{}

func NewDummyStorage() *DummyStorage {
	return &DummyStorage{}
}

func (d *DummyStorage) Store(_ context.Context, sample metrics.Sample) error {
	log.Printf("Dummy storage is storing: %+v\n", sample)
	return nil
}

func (d *DummyStorage) Last(_ context.Context, numSamples int) metrics.Samples {
	log.Printf("Dummy storage returning last %d samples\n", numSamples)
	return nil
}
