package proc

import (
	"syscall"
	"time"
)

// we can use 5500 milliseconds is because most of
// our queue are blocking mode with 5 seconds
var delayTimeBeforeForceQuit time.Duration

// killFn allows tests to stub the real kill implementation.
var killFn = kill

// SetTimeToForceQuit sets the waiting time before force quitting.
func SetTimeToForceQuit(duration time.Duration) {
	delayTimeBeforeForceQuit = duration
}

// Shutdown calls the registered shutdown listeners, only for test purpose.
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
