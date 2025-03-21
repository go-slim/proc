//go:build !windows
// +build !windows

package proc

import (
	"syscall"
)

func kill(sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}
