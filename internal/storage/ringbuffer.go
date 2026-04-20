package storage

import (
	"sync"
)

// thread-safe fixed-size circular buffer that stores items of type T.
// When the buffer is full, the oldest item is overwritten.
type RingBuffer[T any] struct {
	mu       sync.RWMutex
	buffer   []T
	capacity int
	head     int // index of the next write position
	tail     int // index of the oldest item (if size > 0)
	size     int // number of items
}

// Capacity must be positive.
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		panic("ring buffer capacity must be positive")
	}
	return &RingBuffer[T]{
		buffer:   make([]T, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		size:     0,
	}
}

// add an item to the buffer, overwriting the oldest item if the buffer is full.
func (rb *RingBuffer[T]) Push(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.head] = item
	rb.head = (rb.head + 1) % rb.capacity
	if rb.size < rb.capacity {
		rb.size++
	} else {
		// buffer is full, increase tail - if over capacity, wrap around
		rb.tail = (rb.tail + 1) % rb.capacity
	}
}

// a copy of all items in the buffer in chronological order (oldest first).
func (rb *RingBuffer[T]) Slice() []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return nil
	}
	result := make([]T, rb.size)
	if rb.head > rb.tail || rb.size < rb.capacity {
		// contiguous segment
		copy(result, rb.buffer[rb.tail:rb.head])
	} else {
		// wrapped around
		n := copy(result, rb.buffer[rb.tail:])
		copy(result[n:], rb.buffer[:rb.head])
	}
	return result
}

// copies the last n items in the buffer in chronological order (oldest first).
// If n is greater than the number of items stored, all items are returned.
// If n <= 0 or buffer is empty, returns nil.
func (rb *RingBuffer[T]) Last(n int) []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 || n <= 0 {
		return nil
	}
	if n > rb.size {
		n = rb.size
	}

	result := make([]T, n)
	// start index of the oldest among the last n items
	start := (rb.tail + (rb.size - n)) % rb.capacity
	if start+n <= rb.capacity {
		// contiguous segment
		copy(result, rb.buffer[start:start+n])
	} else {
		// wraps around the end of the buffer
		first := rb.capacity - start
		copy(result, rb.buffer[start:])
		copy(result[first:], rb.buffer[:n-first])
	}
	return result
}

// number of items currently stored in the buffer.
func (rb *RingBuffer[T]) Len() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// the maximum number of items the buffer can hold.
func (rb *RingBuffer[T]) Capacity() int {
	return rb.capacity
}

// removes all items from the buffer.
func (rb *RingBuffer[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.head = 0
	rb.tail = 0
	rb.size = 0
	// zero out the buffer to allow garbage collection
	for i := range rb.buffer {
		var zero T
		rb.buffer[i] = zero
	}
}

// Returns the oldest item without removing it.
// Returns false if the buffer is empty.
func (rb *RingBuffer[T]) PeekOldest() (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if rb.size == 0 {
		var zero T
		return zero, false
	}
	return rb.buffer[rb.tail], true
}

// returns the most recently added item without removing it.
// Returns false if the buffer is empty.
func (rb *RingBuffer[T]) PeekNewest() (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if rb.size == 0 {
		var zero T
		return zero, false
	}
	// head points to the next write position, so the newest item is at head-1.
	// Adding capacity before modulo ensures positive index when head == 0.
	idx := (rb.head - 1 + rb.capacity) % rb.capacity
	return rb.buffer[idx], true
}
