//go:build !windows
// +build !windows

package proc

import (
	"os/exec"
	"syscall"
)

// SetSysProcAttribute sets the common SysProcAttrs for commands
func SetSysProcAttribute(cmd *exec.Cmd) {
	// When Gitea runs SubProcessA -> SubProcessB and SubProcessA gets killed by context timeout, use setpgid to make sure the sub processes can be reaped instead of leaving defunct(zombie) processes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
