package proc

import (
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestShutdown_Immediate_NotifiesAndKills(t *testing.T) {
	oldKill := killFn
	defer func() { killFn = oldKill }()

	var called int32
	var gotSig syscall.Signal
	killFn = func(sig syscall.Signal) error {
		atomic.AddInt32(&called, 1)
		gotSig = sig
		return nil
	}

	SetTimeToForceQuit(0)

	var notified int32
	Once(syscall.SIGTERM, func() { atomic.AddInt32(&notified, 1) })

	if err := Shutdown(syscall.SIGTERM); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("killFn should be called once, got %d", called)
	}
	if gotSig != syscall.SIGTERM {
		t.Fatalf("killFn got sig=%v want SIGTERM", gotSig)
	}
	if atomic.LoadInt32(&notified) != 1 {
		t.Fatalf("notify Once should be called once, got %d", notified)
	}
}

func TestShutdown_Delayed_WaitsAndKills(t *testing.T) {
	oldKill := killFn
	defer func() { killFn = oldKill }()

	var called int32
	killCh := make(chan struct{}, 1)
	killFn = func(sig syscall.Signal) error {
		atomic.AddInt32(&called, 1)
		killCh <- struct{}{}
		return nil
	}

	delay := 60 * time.Millisecond
	SetTimeToForceQuit(delay)

	// observe notify fired too (optional)
	done := make(chan struct{}, 1)
	Once(syscall.SIGTERM, func() { done <- struct{}{} })

	start := time.Now()
	if err := Shutdown(syscall.SIGTERM); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	elapsed := time.Since(start)

	select {
	case <-killCh:
		// ok
	default:
		t.Fatalf("killFn not called")
	}
	if elapsed < delay {
		t.Fatalf("expected elapsed >= %v, got %v", delay, elapsed)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("killFn should be called once, got %d", called)
	}
	select {
	case <-done:
		// ok
	default:
		// notify might have raced, but with delay we expect it to happen
		t.Fatalf("expected SIGTERM notify to run")
	}

	// reset delay for other tests
	SetTimeToForceQuit(0)
}

func TestShutdown_KillError(t *testing.T) {
	// Test that Shutdown returns error if kill fails
	oldKill := killFn
	defer func() { killFn = oldKill }()

	expectedErr := syscall.EPERM
	killFn = func(sig syscall.Signal) error {
		return expectedErr
	}

	SetTimeToForceQuit(0)

	err := Shutdown(syscall.SIGTERM)
	if err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestSetTimeToForceQuit(t *testing.T) {
	// Test that SetTimeToForceQuit updates the delay
	oldDelay := delayTimeBeforeForceQuit
	defer func() { delayTimeBeforeForceQuit = oldDelay }()

	newDelay := 123 * time.Millisecond
	SetTimeToForceQuit(newDelay)

	if delayTimeBeforeForceQuit != newDelay {
		t.Fatalf("Expected delay %v, got %v", newDelay, delayTimeBeforeForceQuit)
	}
}

func TestShutdown_MultipleListeners(t *testing.T) {
	// Test that all listeners are notified during shutdown
	oldKill := killFn
	defer func() { killFn = oldKill }()

	var killCalled int32
	killFn = func(sig syscall.Signal) error {
		atomic.AddInt32(&killCalled, 1)
		return nil
	}

	SetTimeToForceQuit(0)

	var count1, count2, count3 int32
	Once(syscall.SIGTERM, func() { atomic.AddInt32(&count1, 1) })
	Once(syscall.SIGTERM, func() { atomic.AddInt32(&count2, 1) })
	Once(syscall.SIGTERM, func() { atomic.AddInt32(&count3, 1) })

	if err := Shutdown(syscall.SIGTERM); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if atomic.LoadInt32(&killCalled) != 1 {
		t.Fatalf("kill should be called once, got %d", killCalled)
	}
	if atomic.LoadInt32(&count1) != 1 {
		t.Fatalf("listener 1 should be notified once, got %d", count1)
	}
	if atomic.LoadInt32(&count2) != 1 {
		t.Fatalf("listener 2 should be notified once, got %d", count2)
	}
	if atomic.LoadInt32(&count3) != 1 {
		t.Fatalf("listener 3 should be notified once, got %d", count3)
	}
}
