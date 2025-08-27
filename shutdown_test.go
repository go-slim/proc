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
