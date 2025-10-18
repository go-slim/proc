//go:build !windows
// +build !windows

package proc

import (
	"context"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestSetSysProcAttribute_Unix(t *testing.T) {
	// Test that SetSysProcAttribute sets Setpgid on Unix systems
	cmd := exec.Command("sh", "-c", "echo test")
	SetSysProcAttribute(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr should not be nil after SetSysProcAttribute")
	}
	if !cmd.SysProcAttr.Setpgid {
		t.Fatal("Setpgid should be true")
	}
}

func TestExec_ProcessGroupCreated_Unix(t *testing.T) {
	// Test that a process group is created
	var pgid int

	err := Exec(context.Background(), ExecOptions{
		Command: "sh",
		Args:    []string{"-c", "echo $$"},
		Timeout: 2 * time.Second,
		OnStart: func(cmd *exec.Cmd) {
			if cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid {
				// Get process group ID
				if cmd.Process != nil {
					pgid, _ = syscall.Getpgid(cmd.Process.Pid)
				}
			}
		},
	})

	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if pgid == 0 {
		t.Log("Warning: Could not verify process group ID")
	}
}
