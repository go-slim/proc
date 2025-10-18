package proc

import (
	"syscall"
	"time"
)

// delayTimeBeforeForceQuit specifies the duration to wait before forcefully
// killing the process. A value of 5500 milliseconds is typically used because
// most queues operate in blocking mode with a 5-second timeout.
var delayTimeBeforeForceQuit time.Duration

// killFn is the function used to kill the process. It can be stubbed in tests
// to verify shutdown behavior without actually killing the process.
var killFn = kill

// SetTimeToForceQuit sets the duration to wait before forcefully killing
// the process during shutdown. If set to 0, the process will be killed
// immediately without attempting graceful shutdown.
func SetTimeToForceQuit(duration time.Duration) {
	delayTimeBeforeForceQuit = duration
}

// Shutdown performs a graceful shutdown by notifying all registered signal
// listeners and optionally waiting for a configured delay before force killing.
//
// If delayTimeBeforeForceQuit > 0, it will:
//  1. Send SIGTERM to all registered listeners in a goroutine
//  2. Wait for delayTimeBeforeForceQuit duration
//  3. Force kill the process if still alive
//
// If delayTimeBeforeForceQuit == 0, it will:
//  1. Send SIGTERM to all registered listeners synchronously
//  2. Immediately kill the process
func Shutdown(sig syscall.Signal) error {
	debugf("Got signal %d, shutting down...", sig)

	if delayTimeBeforeForceQuit > 0 {
		go Notify(syscall.SIGTERM)
		time.Sleep(delayTimeBeforeForceQuit)
		debugf("Still alive after %v, going to force kill the process...", delayTimeBeforeForceQuit)
	} else {
		Notify(syscall.SIGTERM)
	}

	return killFn(sig)
}
