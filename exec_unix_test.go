//go:build !windows
// +build !windows

package proc

import (
	"os/exec"
	"syscall"
	"testing"
)

func TestSetSysProcAttribute_SetsSetpgid(t *testing.T) {
	cmd := &exec.Cmd{}
	SetSysProcAttribute(cmd)
	if cmd.SysProcAttr == nil {
		t.Fatalf("SysProcAttr should not be nil")
	}
	if !cmd.SysProcAttr.Setpgid {
		t.Fatalf("expected Setpgid to be true, got: %#v", cmd.SysProcAttr)
	}
	// Ensure type is *syscall.SysProcAttr
	if _, ok := any(cmd.SysProcAttr).(*syscall.SysProcAttr); !ok {
		t.Fatalf("SysProcAttr has unexpected type")
	}
}
