//go:build windows
// +build windows

package proc

import (
	"os"
	"syscall"
)

// kill terminates the current process on Windows.
// The signal parameter is ignored on Windows as the platform doesn't support
// Unix-style signals. Instead, it uses os.Process.Kill() to forcefully terminate
// the process.
//
// Alternative approaches that could be used:
// - taskkill command: exec.Command("taskkill", "/f", "/pid", strconv.Itoa(pid)).Run()
//   Reference: https://www.reddit.com/r/golang/comments/16k10ma/what_are_the_possible_alternatives_for/
//
// Current implementation uses os.FindProcess followed by Kill():
// Reference: https://samdlcong.com/posts/go_syscall_windows/
func kill(_ syscall.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
