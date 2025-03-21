//go:build windows
// +build windows

package proc

import (
	"os"
	"syscall"
)

func kill(_ syscall.Signal) error {
	// https://www.reddit.com/r/golang/comments/16k10ma/what_are_the_possible_alternatives_for/
	// return exec.Command("taskkill", "/f", "/pid", strconv.Itoa(pid)).Run()

	// https://samdlcong.com/posts/go_syscall_windows/
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
