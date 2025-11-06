//go:build unix

package proc

import (
	"sync"
	"syscall"
	"testing"
)

func TestSignal_Cancel_MultipleIDs(t *testing.T) {
	// Test cancelling multiple listeners at once
	// Use actual syscall.Signal
	id1 := On(syscall.SIGUSR1, func() {})
	id2 := On(syscall.SIGUSR1, func() {})
	id3 := On(syscall.SIGUSR1, func() {})

	// Cancel all three
	Cancel(id1, id2, id3)

	// Verify they're all removed
	if Notify(syscall.SIGUSR1) {
		t.Fatal("All listeners should be cancelled")
	}
}

func TestSignal_On_ReturnsUniqueIDs(t *testing.T) {
	// Verify that each On call returns a unique ID
	// Use actual syscall.Signal to ensure valid IDs are returned
	id1 := On(syscall.SIGUSR1, func() {})
	id2 := On(syscall.SIGUSR1, func() {})
	id3 := On(syscall.SIGUSR1, func() {})

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
	// Use actual syscall.Signal to ensure valid IDs are returned
	id1 := Once(syscall.SIGUSR2, func() {})
	id2 := Once(syscall.SIGUSR2, func() {})

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
	// Use actual syscall.Signal
	var counter int
	var mu sync.Mutex

	var ids []uint32
	for range 10 {
		id := On(syscall.SIGUSR1, func() {
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

	for range goroutines {
		go func() {
			defer wg.Done()
			Notify(syscall.SIGUSR1)
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
