//go:build windows

package proc

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
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
	for range 10 {
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

	for range goroutines {
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

func TestWait_BlocksUntilSignal(t *testing.T) {
	// Test that Wait blocks until the signal is received
	// On Windows, we use os.Interrupt instead of SIGUSR1
	done := make(chan struct{})
	received := false

	go func() {
		Wait(os.Interrupt)
		received = true
		close(done)
	}()

	// Give Wait time to register the listener
	time.Sleep(10 * time.Millisecond)

	// Trigger the signal using Notify instead of syscall.Kill
	// (Windows doesn't support sending signals to self the same way)
	Notify(os.Interrupt)

	// Wait should unblock
	select {
	case <-done:
		if !received {
			t.Fatal("Wait should have unblocked after signal")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not unblock within timeout")
	}
}

func TestWait_MultipleWaiters(t *testing.T) {
	// Test that multiple goroutines can Wait for the same signal
	// On Windows, we use syscall.SIGTERM
	const numWaiters = 5
	var wg sync.WaitGroup
	wg.Add(numWaiters)

	for range numWaiters {
		go func() {
			defer wg.Done()
			Wait(syscall.SIGTERM)
		}()
	}

	// Give all Wait calls time to register
	time.Sleep(10 * time.Millisecond)

	// Trigger the signal using Notify
	Notify(syscall.SIGTERM)

	// All waiters should be unblocked
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Not all waiters were unblocked within timeout")
	}
}
