//go:build windows

package proc

import (
	"sync"
	"syscall"
	"testing"
)

// On Windows, we use SIGTERM and SIGINT instead of SIGUSR1 and SIGUSR2
// since Windows doesn't support user-defined signals

func TestSignal_Cancel_MultipleIDs(t *testing.T) {
	// Test cancelling multiple listeners at once
	id1 := On(syscall.SIGTERM, func() {})
	id2 := On(syscall.SIGTERM, func() {})
	id3 := On(syscall.SIGTERM, func() {})

	// Cancel all three
	Cancel(id1, id2, id3)

	// Verify they're all removed
	if Notify(syscall.SIGTERM) {
		t.Fatal("All listeners should be cancelled")
	}
}

func TestSignal_On_ReturnsUniqueIDs(t *testing.T) {
	// Verify that each On call returns a unique ID
	id1 := On(syscall.SIGTERM, func() {})
	id2 := On(syscall.SIGTERM, func() {})
	id3 := On(syscall.SIGTERM, func() {})

	if id1 == 0 || id2 == 0 || id3 == 0 {
		t.Fatalf("IDs should not be zero: %d, %d, %d", id1, id2, id3)
	}
	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Fatalf("IDs should be unique: %d, %d, %d", id1, id2, id3)
	}

	Cancel(id1, id2, id3)
}

func TestSignal_Once_ReturnsUniqueIDs(t *testing.T) {
	// Verify that each Once call returns a unique ID
	// Use SIGINT instead of SIGUSR2
	id1 := Once(syscall.SIGINT, func() {})
	id2 := Once(syscall.SIGINT, func() {})

	if id1 == 0 || id2 == 0 {
		t.Fatalf("IDs should not be zero: %d, %d", id1, id2)
	}
	if id1 == id2 {
		t.Fatalf("IDs should be unique: %d, %d", id1, id2)
	}

	Cancel(id1, id2)
}

func TestSignal_ConcurrentNotify(t *testing.T) {
	// Test that concurrent Notify calls are safe
	var counter int
	var mu sync.Mutex

	var ids []uint32
	for i := 0; i < 10; i++ {
		id := On(syscall.SIGTERM, func() {
			mu.Lock()
			counter++
			mu.Unlock()
		})
		ids = append(ids, id)
	}
	defer Cancel(ids...)

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			Notify(syscall.SIGTERM)
		}()
	}

	wg.Wait()

	mu.Lock()
	expected := 10 * goroutines
	if counter != expected {
		t.Fatalf("Expected counter=%d, got %d", expected, counter)
	}
	mu.Unlock()
}
