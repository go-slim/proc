package proc

import (
	"sync"
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
	// Disable logger output during this test to avoid noise from expected panic
	old := Logger
	Logger = nil
	defer func() { Logger = old }()

	var wg sync.WaitGroup
	run := safeRunner(&wg)
	run(func() { panic("boom") })
	wg.Wait()
	// If recovery failed, test would panic; reaching here is success.
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

func TestSignal_InvalidSignalReturnsZeroID(t *testing.T) {
	// When adding an invalid signal, should return 0
	// This happens when signum returns -1
	// bogusSignal is not a syscall.Signal, so signum returns -1, and add() returns 0
	id := On(bogusSignal{}, func() {})
	if id != 0 {
		t.Fatalf("Invalid signal should return ID 0, got %d", id)
	}
}
