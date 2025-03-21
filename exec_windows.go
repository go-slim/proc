//go:build windows
// +build windows

package proc

import "os/exec"

// SetSysProcAttribute sets the common SysProcAttrs for commands
func SetSysProcAttribute(cmd *exec.Cmd) {
	// Do nothing
}
