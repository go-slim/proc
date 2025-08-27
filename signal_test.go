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
	var wg sync.WaitGroup
	run := safeRunner(&wg)
	run(func() { panic("boom") })
	wg.Wait()
	// If recovery failed, test would panic; reaching here is success.
}
