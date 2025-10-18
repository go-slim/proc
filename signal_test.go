package proc

import (
	"sync"
	"syscall"
	"testing"
)

// A custom signal type that is NOT syscall.Signal to force signum() -> -1
// which makes Notify return false.
type bogusSignal struct{}

func (bogusSignal) String() string { return "bogus" }
func (bogusSignal) Signal()        {}

func TestNotify_UnknownSignal_ReturnsFalse(t *testing.T) {
	if Notify(bogusSignal{}) {
		t.Fatalf("Notify should return false for unknown signals")
	}
}

func TestSafeRunner_RecoveryDoesNotPanic(t *testing.T) {
	var wg sync.WaitGroup
	run := safeRunner(&wg)
	run(func() { panic("boom") })
	wg.Wait()
	// If recovery failed, test would panic; reaching here is success.
}

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

func TestSignal_Cancel_ZeroIDs(t *testing.T) {
	// Test that cancelling with zero IDs is safe
	Cancel(0, 0, 0)
	// Should not panic
}

func TestSignal_Cancel_InvalidIDs(t *testing.T) {
	// Test cancelling non-existent IDs is safe
	Cancel(99999, 88888, 77777)
	// Should not panic
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

func TestSignal_InvalidSignalReturnsZeroID(t *testing.T) {
	// When adding an invalid signal, should return 0
	// This happens when signum returns -1
	// bogusSignal is not a syscall.Signal, so signum returns -1, and add() returns 0
	id := On(bogusSignal{}, func() {})
	if id != 0 {
		t.Fatalf("Invalid signal should return ID 0, got %d", id)
	}
}

func TestSignal_ConcurrentNotify(t *testing.T) {
	// Test that concurrent Notify calls are safe
	// Use actual syscall.Signal
	var counter int
	var mu sync.Mutex

	var ids []uint32
	for i := 0; i < 10; i++ {
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

	for i := 0; i < goroutines; i++ {
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
