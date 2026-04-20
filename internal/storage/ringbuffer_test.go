package storage

import (
	"sync"
	"testing"
)

func TestNewRingBuffer(t *testing.T) {
	t.Run("positive capacity", func(t *testing.T) {
		rb := NewRingBuffer[int](5)
		if rb == nil {
			t.Fatal("expected non-nil ring buffer")
		}
		if rb.Capacity() != 5 {
			t.Errorf("Capacity: got %v, want 5", rb.Capacity())
		}
		if rb.Len() != 0 {
			t.Errorf("Len: got %v, want 0", rb.Len())
		}
	})

	t.Run("zero capacity panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for zero capacity")
			}
		}()
		NewRingBuffer[int](0)
	})

	t.Run("negative capacity panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for negative capacity")
			}
		}()
		NewRingBuffer[int](-1)
	})
}

func TestRingBufferPushAndLen(t *testing.T) {
	rb := NewRingBuffer[int](3)

	// Push first item
	rb.Push(10)
	if rb.Len() != 1 {
		t.Errorf("Len after first push: got %v, want 1", rb.Len())
	}

	// Push second item
	rb.Push(20)
	if rb.Len() != 2 {
		t.Errorf("Len after second push: got %v, want 2", rb.Len())
	}

	// Push third item (buffer full)
	rb.Push(30)
	if rb.Len() != 3 {
		t.Errorf("Len after third push: got %v, want 3", rb.Len())
	}

	// Push fourth item (overwrites oldest)
	rb.Push(40)
	if rb.Len() != 3 {
		t.Errorf("Len after overflow push: got %v, want 3", rb.Len())
	}
}

func TestRingBufferSlice(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		rb := NewRingBuffer[int](3)
		slice := rb.Slice()
		if slice != nil {
			t.Errorf("Slice of empty buffer: got %v, want nil", slice)
		}
	})

	t.Run("partially filled buffer", func(t *testing.T) {
		rb := NewRingBuffer[int](5)
		rb.Push(1)
		rb.Push(2)
		rb.Push(3)

		slice := rb.Slice()
		expected := []int{1, 2, 3}
		if len(slice) != len(expected) {
			t.Fatalf("Slice length: got %v, want %v", len(slice), len(expected))
		}
		for i, v := range expected {
			if slice[i] != v {
				t.Errorf("Slice[%d]: got %v, want %v", i, slice[i], v)
			}
		}
	})

	t.Run("full buffer no wrap", func(t *testing.T) {
		rb := NewRingBuffer[int](3)
		rb.Push(1)
		rb.Push(2)
		rb.Push(3)

		slice := rb.Slice()
		expected := []int{1, 2, 3}
		if len(slice) != len(expected) {
			t.Fatalf("Slice length: got %v, want %v", len(slice), len(expected))
		}
		for i, v := range expected {
			if slice[i] != v {
				t.Errorf("Slice[%d]: got %v, want %v", i, slice[i], v)
			}
		}
	})

	t.Run("full buffer with wrap", func(t *testing.T) {
		rb := NewRingBuffer[int](3)
		rb.Push(1)
		rb.Push(2)
		rb.Push(3)
		rb.Push(4) // overwrites 1, buffer now [4, 2, 3] with tail at 2

		slice := rb.Slice()
		expected := []int{2, 3, 4}
		if len(slice) != len(expected) {
			t.Fatalf("Slice length: got %v, want %v", len(slice), len(expected))
		}
		for i, v := range expected {
			if slice[i] != v {
				t.Errorf("Slice[%d]: got %v, want %v", i, slice[i], v)
			}
		}
	})

	t.Run("multiple wraps", func(t *testing.T) {
		rb := NewRingBuffer[int](3)
		// Fill buffer
		rb.Push(1)
		rb.Push(2)
		rb.Push(3)
		// Wrap once
		rb.Push(4)
		rb.Push(5)
		rb.Push(6)
		// Wrap again
		rb.Push(7)

		slice := rb.Slice()
		expected := []int{5, 6, 7}
		if len(slice) != len(expected) {
			t.Fatalf("Slice length: got %v, want %v", len(slice), len(expected))
		}
		for i, v := range expected {
			if slice[i] != v {
				t.Errorf("Slice[%d]: got %v, want %v", i, slice[i], v)
			}
		}
	})
}

func TestRingBufferWrap(t *testing.T) {
	t.Run("with wrap", func(t *testing.T) {
		rb2 := NewRingBuffer[int](3)
		rb2.Push(1)
		rb2.Push(2)
		rb2.Push(3)
		rb2.Push(4) // buffer: [4, 2, 3], tail at 2

		last := rb2.Last(2)
		expected := []int{3, 4}
		if len(last) != len(expected) {
			t.Fatalf("Last(2) with wrap length: got %v, want %v", len(last), len(expected))
		}
		for i, v := range expected {
			if last[i] != v {
				t.Errorf("Last(2) with wrap[%d]: got %v, want %v", i, last[i], v)
			}
		}
	})

	t.Run("complex wrap scenario", func(t *testing.T) {
		rb3 := NewRingBuffer[int](4)
		// Push 1,2,3,4
		for i := 1; i <= 4; i++ {
			rb3.Push(i)
		}
		// Push 5,6 (wrapping)
		rb3.Push(5)
		rb3.Push(6)
		// Buffer: [5, 6, 3, 4], tail at 3

		last := rb3.Last(3)
		expected := []int{4, 5, 6}
		if len(last) != len(expected) {
			t.Fatalf("Last(3) complex wrap length: got %v, want %v", len(last), len(expected))
		}
		for i, v := range expected {
			if last[i] != v {
				t.Errorf("Last(3) complex wrap[%d]: got %v, want %v", i, last[i], v)
			}
		}
	})
}

func TestRingBufferLast(t *testing.T) {
	rb := NewRingBuffer[int](5)
	// Fill with 1,2,3,4,5
	for i := 1; i <= 5; i++ {
		rb.Push(i)
	}

	t.Run("n larger than size", func(t *testing.T) {
		last := rb.Last(10)
		if len(last) != 5 {
			t.Fatalf("Last(10) length: got %v, want 5", len(last))
		}
		expected := []int{1, 2, 3, 4, 5}
		for i, v := range expected {
			if last[i] != v {
				t.Errorf("Last(10)[%d]: got %v, want %v", i, last[i], v)
			}
		}
	})

	t.Run("n equals size", func(t *testing.T) {
		last := rb.Last(5)
		if len(last) != 5 {
			t.Fatalf("Last(5) length: got %v, want 5", len(last))
		}
		expected := []int{1, 2, 3, 4, 5}
		for i, v := range expected {
			if last[i] != v {
				t.Errorf("Last(5)[%d]: got %v, want %v", i, last[i], v)
			}
		}
	})

	t.Run("n less than size", func(t *testing.T) {
		last := rb.Last(3)
		if len(last) != 3 {
			t.Fatalf("Last(3) length: got %v, want 3", len(last))
		}
		expected := []int{3, 4, 5}
		for i, v := range expected {
			if last[i] != v {
				t.Errorf("Last(3)[%d]: got %v, want %v", i, last[i], v)
			}
		}
	})
}

func TestRingBufferEdge(t *testing.T) {
	rb := NewRingBuffer[int](5)

	t.Run("empty buffer", func(t *testing.T) {
		last := rb.Last(1)
		if last != nil {
			t.Errorf("Last from empty buffer: got %v, want nil", last)
		}

		last = rb.Last(0)
		if last != nil {
			t.Errorf("Last with n=0: got %v, want nil", last)
		}

		last = rb.Last(-1)
		if last != nil {
			t.Errorf("Last with negative n: got %v, want nil", last)
		}
	})

	// Fill with 1,2,3,4,5
	for i := 1; i <= 5; i++ {
		rb.Push(i)
	}

	t.Run("n=1", func(t *testing.T) {
		last := rb.Last(1)
		if len(last) != 1 {
			t.Fatalf("Last(1) length: got %v, want 1", len(last))
		}
		if last[0] != 5 {
			t.Errorf("Last(1)[0]: got %v, want 5", last[0])
		}
	})
}

func TestRingBufferPeekOldest(t *testing.T) {
	rb := NewRingBuffer[int](3)

	t.Run("empty buffer", func(t *testing.T) {
		val, ok := rb.PeekOldest()
		if ok {
			t.Errorf("PeekOldest on empty buffer: got ok=true, want false")
		}
		if val != 0 {
			t.Errorf("PeekOldest on empty buffer: got %v, want zero value", val)
		}
	})

	rb.Push(10)
	val, ok := rb.PeekOldest()
	if !ok {
		t.Error("PeekOldest on non-empty buffer: got ok=false, want true")
	}
	if val != 10 {
		t.Errorf("PeekOldest: got %v, want 10", val)
	}

	rb.Push(20)
	val, ok = rb.PeekOldest()
	if !ok {
		t.Error("PeekOldest after second push: got ok=false, want true")
	}
	if val != 10 {
		t.Errorf("PeekOldest after second push: got %v, want 10", val)
	}

	// Fill buffer and cause wrap
	rb.Push(30)
	rb.Push(40) // overwrites 10, buffer: [40, 20, 30], tail at 20
	val, ok = rb.PeekOldest()
	if !ok {
		t.Error("PeekOldest after wrap: got ok=false, want true")
	}
	if val != 20 {
		t.Errorf("PeekOldest after wrap: got %v, want 20", val)
	}
}

func TestRingBufferPeekNewest(t *testing.T) {
	rb := NewRingBuffer[int](3)

	t.Run("empty buffer", func(t *testing.T) {
		val, ok := rb.PeekNewest()
		if ok {
			t.Errorf("PeekNewest on empty buffer: got ok=true, want false")
		}
		if val != 0 {
			t.Errorf("PeekNewest on empty buffer: got %v, want zero value", val)
		}
	})

	rb.Push(10)
	val, ok := rb.PeekNewest()
	if !ok {
		t.Error("PeekNewest on non-empty buffer: got ok=false, want true")
	}
	if val != 10 {
		t.Errorf("PeekNewest: got %v, want 10", val)
	}

	rb.Push(20)
	val, ok = rb.PeekNewest()
	if !ok {
		t.Error("PeekNewest after second push: got ok=false, want true")
	}
	if val != 20 {
		t.Errorf("PeekNewest after second push: got %v, want 20", val)
	}

	rb.Push(30)
	val, ok = rb.PeekNewest()
	if !ok {
		t.Error("PeekNewest after third push: got ok=false, want true")
	}
	if val != 30 {
		t.Errorf("PeekNewest after third push: got %v, want 30", val)
	}

	// Wrap
	rb.Push(40)
	val, ok = rb.PeekNewest()
	if !ok {
		t.Error("PeekNewest after wrap: got ok=false, want true")
	}
	if val != 40 {
		t.Errorf("PeekNewest after wrap: got %v, want 40", val)
	}
}

func TestRingBufferClear(t *testing.T) {
	rb := NewRingBuffer[int](3)
	rb.Push(1)
	rb.Push(2)
	rb.Push(3)

	if rb.Len() != 3 {
		t.Errorf("Len before clear: got %v, want 3", rb.Len())
	}

	rb.Clear()

	if rb.Len() != 0 {
		t.Errorf("Len after clear: got %v, want 0", rb.Len())
	}

	slice := rb.Slice()
	if slice != nil {
		t.Errorf("Slice after clear: got %v, want nil", slice)
	}

	// Verify we can push after clear
	rb.Push(10)
	if rb.Len() != 1 {
		t.Errorf("Len after push post-clear: got %v, want 1", rb.Len())
	}
	val, ok := rb.PeekNewest()
	if !ok || val != 10 {
		t.Errorf("PeekNewest after push post-clear: got %v (ok=%v), want 10", val, ok)
	}
}

func TestRingBufferConcurrentAccess(t *testing.T) {
	rb := NewRingBuffer[int](100)
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				rb.Push(id*1000 + j)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				rb.Slice()
				rb.Last(10)
				rb.Len()
				rb.PeekOldest()
				rb.PeekNewest()
			}
		}()
	}

	wg.Wait()

	// Buffer should be in a consistent state
	length := rb.Len()
	if length != 100 {
		t.Errorf("Final buffer length: got %v, want 100", length)
	}

	slice := rb.Slice()
	if len(slice) != 100 {
		t.Errorf("Final slice length: got %v, want 100", len(slice))
	}
}

func TestRingBufferWithStructType(t *testing.T) {
	type Point struct {
		X, Y int
	}

	rb := NewRingBuffer[Point](2)
	rb.Push(Point{X: 1, Y: 2})
	rb.Push(Point{X: 3, Y: 4})

	slice := rb.Slice()
	if len(slice) != 2 {
		t.Fatalf("Slice length with struct: got %v, want 2", len(slice))
	}
	if slice[0].X != 1 || slice[0].Y != 2 {
		t.Errorf("Slice[0]: got %v, want {1 2}", slice[0])
	}
	if slice[1].X != 3 || slice[1].Y != 4 {
		t.Errorf("Slice[1]: got %v, want {3 4}", slice[1])
	}

	// Test zero value after clear
	rb.Clear()
	rb.Push(Point{X: 5, Y: 6})
	val, ok := rb.PeekNewest()
	if !ok || val.X != 5 || val.Y != 6 {
		t.Errorf("PeekNewest after clear: got %v (ok=%v), want {5 6}", val, ok)
	}
}
