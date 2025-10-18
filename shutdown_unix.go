//go:build !windows
// +build !windows

package proc

import (
	"syscall"
)

// kill sends the specified signal to the current process on Unix-like systems.
// This is the platform-specific implementation that uses syscall.Kill.
func kill(sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}
