//go:build !windows
// +build !windows

package proc

import (
	"os/exec"
	"syscall"
)

// SetSysProcAttribute sets the system-specific process attributes for Unix-like systems.
// It configures the command to use Setpgid to create a new process group,
// which ensures that child processes can be properly reaped when the parent
// is killed, preventing defunct (zombie) processes.
//
// This is particularly important when SubProcessA spawns SubProcessB and
// SubProcessA gets killed by context timeout - Setpgid ensures that
// SubProcessB can still be properly cleaned up.
func SetSysProcAttribute(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
