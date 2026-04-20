package storage

import (
	"context"
	"pulsecat/internal/collector"
	"pulsecat/internal/metrics"
	"testing"
)

func TestBufferedStorage(t *testing.T) {
	ctx := context.Background()
	storage := NewBufferedStorage(3)

	// Test Store
	sample1 := &collector.LoadAverage{OneMin: 1.0, FiveMin: 2.0, FifteenMin: 3.0}
	err := storage.Store(ctx, sample1)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Test Last with single sample
	samples := storage.Last(ctx, 1)
	if len(samples) != 1 {
		t.Fatalf("Last(1) length: got %v, want 1", len(samples))
	}
	if samples[0] != sample1 {
		t.Error("Last(1) returned wrong sample")
	}

	// Store more samples
	sample2 := &collector.LoadAverage{OneMin: 2.0, FiveMin: 3.0, FifteenMin: 4.0}
	sample3 := &collector.LoadAverage{OneMin: 3.0, FiveMin: 4.0, FifteenMin: 5.0}
	storage.Store(ctx, sample2)
	storage.Store(ctx, sample3)

	// Test Last with multiple samples
	samples = storage.Last(ctx, 3)
	if len(samples) != 3 {
		t.Fatalf("Last(3) length: got %v, want 3", len(samples))
	}
	if samples[0] != sample1 || samples[1] != sample2 || samples[2] != sample3 {
		t.Error("Last(3) returned wrong samples or wrong order")
	}

	// Test Last with n larger than stored samples
	samples = storage.Last(ctx, 5)
	if len(samples) != 3 {
		t.Fatalf("Last(5) length: got %v, want 3", len(samples))
	}

	// Test overflow (buffer capacity is 3)
	sample4 := &collector.LoadAverage{OneMin: 4.0, FiveMin: 5.0, FifteenMin: 6.0}
	storage.Store(ctx, sample4)

	samples = storage.Last(ctx, 3)
	if len(samples) != 3 {
		t.Fatalf("Last(3) after overflow length: got %v, want 3", len(samples))
	}
	// After overflow, sample1 should be evicted
	if samples[0] != sample2 || samples[1] != sample3 || samples[2] != sample4 {
		t.Error("Last(3) after overflow returned wrong samples")
	}
}

func TestDummyStorage(t *testing.T) {
	ctx := context.Background()
	storage := NewDummyStorage()

	// Test Store doesn't error
	sample := &collector.LoadAverage{OneMin: 1.0, FiveMin: 2.0, FifteenMin: 3.0}
	err := storage.Store(ctx, sample)
	if err != nil {
		t.Fatalf("DummyStorage.Store failed: %v", err)
	}

	// Test Last always returns nil
	samples := storage.Last(ctx, 5)
	if samples != nil {
		t.Errorf("DummyStorage.Last: got %v, want nil", samples)
	}

	samples = storage.Last(ctx, 0)
	if samples != nil {
		t.Errorf("DummyStorage.Last(0): got %v, want nil", samples)
	}
}

func TestStorageInterface(_ *testing.T) {
	// Verify that both BufferedStorage and DummyStorage implement Storage interface
	var _ Storage = (*BufferedStorage)(nil)
	var _ Storage = (*DummyStorage)(nil)
}

func TestBufferedStorageWithDifferentSampleTypes(t *testing.T) {
	ctx := context.Background()
	storage := NewBufferedStorage(5)

	// Mix different sample types (they all implement metrics.Sample)
	samples := []metrics.Sample{
		&collector.LoadAverage{OneMin: 1.0},
		&collector.CPUUsage{User: 10.0},
		&collector.NetworkStats{TotalBytesReceived: 1000},
		&collector.TCPConnectionStates{Established: 5},
	}

	for _, sample := range samples {
		err := storage.Store(ctx, sample)
		if err != nil {
			t.Fatalf("Store failed for %T: %v", sample, err)
		}
	}

	// Retrieve all samples
	retrieved := storage.Last(ctx, 5)
	if len(retrieved) != 4 {
		t.Fatalf("Last(5) length: got %v, want 4", len(retrieved))
	}

	// Verify types
	for i, sample := range retrieved {
		expectedType := samples[i]
		// We can't directly compare because they're different types,
		// but we can verify they're not nil
		if sample == nil {
			t.Errorf("Sample %d is nil", i)
		}
		// Check that it's the same concrete type
		if i < len(samples) {
			// This is a basic type check
			switch expectedType.(type) {
			case *collector.LoadAverage:
				if _, ok := sample.(*collector.LoadAverage); !ok {
					t.Errorf("Sample %d: expected *collector.LoadAverage, got %T", i, sample)
				}
			case *collector.CPUUsage:
				if _, ok := sample.(*collector.CPUUsage); !ok {
					t.Errorf("Sample %d: expected *collector.CPUUsage, got %T", i, sample)
				}
			case *collector.NetworkStats:
				if _, ok := sample.(*collector.NetworkStats); !ok {
					t.Errorf("Sample %d: expected *collector.NetworkStats, got %T", i, sample)
				}
			case *collector.TCPConnectionStates:
				if _, ok := sample.(*collector.TCPConnectionStates); !ok {
					t.Errorf("Sample %d: expected *collector.TCPConnectionStates, got %T", i, sample)
				}
			}
		}
	}
}
