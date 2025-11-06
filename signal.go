package proc

import (
	"os"
	"os/signal"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	// seq is an atomic counter for generating unique listener IDs
	seq uint32
	// lock protects the listeners slice during concurrent access
	lock sync.Mutex
	// lns stores all registered signal listeners
	lns []*listener
	// mask is a bitmask tracking which signals have been registered with the OS
	mask uint32
	// sigch is the channel that receives OS signals
	sigch chan os.Signal
)

// registerSignalListener initializes the signal handling system.
// It creates a signal channel and starts a goroutine to handle incoming signals.
// The following signals are handled:
// - SIGHUP, SIGINT, SIGQUIT, SIGTERM: Trigger graceful shutdown
// - Other signals: Dispatched to registered listeners
//
// References:
// - https://golang.org/pkg/os/signal/#Notify
// - https://colobu.com/2015/10/09/Linux-Signals/
func registerSignalListener() {
	// https://golang.org/pkg/os/signal/#Notify
	sigch = make(chan os.Signal, 1)

	// https://colobu.com/2015/10/09/Linux-Signals/
	signal.Notify(
		sigch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		for {
			sig := <-sigch
			debugf("PID: %d. Received %v.", pid, sig)
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				// gracefully shuts down the process.
				Shutdown(syscall.SIGTERM)
				signal.Stop(sigch)
				os.Exit(0)
			default:
				if !Notify(sig) {
					debugf("PID %d. Got unregistered signal: %v.", pid, sig)
				}
			}
		}
	}()
}

// numSig is the maximum number of signals supported across all systems.
// This value is defined to match the implementation in go/src/os/signal/signal.go.
const numSig = 65

// signum converts an os.Signal to its numeric representation.
// Returns -1 if the signal is not a valid syscall.Signal or is out of range.
func signum(sig os.Signal) int {
	switch sig := sig.(type) {
	case syscall.Signal:
		i := int(sig)
		if i < 0 || i >= numSig {
			return -1
		}
		return i
	default:
		return -1
	}
}

// listener represents a registered signal handler.
type listener struct {
	// id is the unique identifier for this listener
	id uint32
	// fn is the callback function to execute when the signal is received
	fn func()
	// sig is the numeric representation of the signal to listen for
	sig int
	// once indicates whether this listener should execute only once
	once bool
}

// add registers a new signal listener with the specified behavior.
// It handles the signal registration with the OS if needed and returns
// a unique ID that can be used to cancel the listener later.
// Returns 0 if the signal is invalid.
func add(sig os.Signal, fn func(), once bool) uint32 {
	if n := signum(sig); n > -1 {
		lock.Lock()
		defer lock.Unlock()

		// see go/src/os/signal/signal.go
		if (mask>>uint(n&31))&1 == 0 {
			mask |= 1 << uint(n&31)
			signal.Notify(sigch, sig)
		}

		id := atomic.AddUint32(&seq, 1)
		lns = append(lns, &listener{
			id:   id,
			fn:   wrap(fn, once),
			sig:  n,
			once: once,
		})
		return id
	}
	return 0
}

// wrap returns a function that optionally ensures single execution.
// If once is true, the returned function will execute fn at most once,
// even if called multiple times. If once is false, returns fn unchanged.
func wrap(fn func(), once bool) func() {
	if !once {
		return fn
	}
	var so sync.Once
	return func() { so.Do(fn) }
}

// On registers a signal handler that will be called every time the specified
// signal is received. Returns a unique ID that can be used with Cancel to
// remove the listener.
func On(sig os.Signal, fn func()) uint32 {
	return add(sig, fn, false)
}

// Once registers a signal handler that will be called at most once when the
// specified signal is received. After execution, the listener is automatically
// removed. Returns a unique ID that can be used with Cancel to remove the
// listener before it executes.
func Once(sig os.Signal, fn func()) uint32 {
	return add(sig, fn, true)
}

// Cancel removes the signal listeners with the specified IDs.
// It's safe to pass IDs that don't exist or have already been removed.
// Zero IDs are ignored.
func Cancel(ids ...uint32) {
	n := len(ids)
	for _, id := range ids {
		if id == 0 {
			n--
		}
	}
	if n == 0 {
		return
	}
	lock.Lock()
	lns = slices.DeleteFunc(lns, func(l *listener) bool {
		return slices.Contains(ids, l.id)
	})
	lock.Unlock()
}

// Wait blocks until the specified signal is received.
// It registers a one-time signal handler and blocks the current goroutine
// until the signal arrives. This is useful for waiting for specific signals
// in a synchronous manner.
//
// Example:
//
//	// Wait for SIGUSR1 signal
//	proc.Wait(syscall.SIGUSR1)
//	fmt.Println("Received SIGUSR1")
func Wait(sig os.Signal) {
	wait := make(chan struct{})
	Once(sig, func() { close(wait) })
	<-wait
}

// Notify dispatches a signal to all registered listeners for that signal.
// It executes all matching listeners concurrently in separate goroutines,
// with panic recovery. Listeners registered with Once are automatically
// removed after execution.
//
// Returns true if at least one listener was notified, false if no listeners
// were registered for the signal or if the signal is invalid.
func Notify(sig os.Signal) bool {
	n := signum(sig)
	if n == -1 {
		return false
	}

	lock.Lock()
	l := len(lns)
	fs := make([]func(), 0, l)

	for i := l - 1; i >= 0; i-- {
		if l := lns[i]; l.sig == n {
			fs = append(fs, l.fn)
			if l.once {
				lns = slices.Delete(lns, i, i+1)
			}
		}
	}
	lock.Unlock()

	if len(fs) == 0 {
		return false
	}

	var wg sync.WaitGroup
	var run = safeRunner(&wg)
	for _, fn := range fs {
		if fn != nil {
			run(fn)
		}
	}
	wg.Wait()

	return true
}

// safeRunner creates a function that executes callbacks in separate goroutines
// with panic recovery. Each callback execution is tracked by the provided
// WaitGroup.
func safeRunner(wg *sync.WaitGroup) func(func()) {
	return func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer recovery()
			fn()
		}()
	}
}

// recovery handles panics that occur during signal listener execution.
// It logs the panic value and stack trace for debugging purposes.
func recovery() {
	if p := recover(); p != nil {
		debugf("%+v\n%s", p, debug.Stack())
	}
}
