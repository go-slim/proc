//go:build windows
// +build windows

package proc

import "os/exec"

// SetSysProcAttribute sets the system-specific process attributes for Windows.
// On Windows, no special process attributes are needed, so this is a no-op.
func SetSysProcAttribute(cmd *exec.Cmd) {
	// Do nothing
}
