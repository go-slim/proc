package proc

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestExec_OnStart_WorkDir_Env(t *testing.T) {
	td := t.TempDir()
	var gotDir string
	var envOK bool

	cb := func(cmd *exec.Cmd) {
		gotDir = cmd.Dir
		joined := strings.Join(cmd.Env, "\n")
		envOK = strings.Contains(joined, "FOO=BAR")
	}

	cmd, args := echoCmdArgs()
	err := Exec(context.Background(), ExecOptions{
		WorkDir: td,
		Env:     []string{"FOO=BAR"},
		Command: cmd,
		Args:    args,
		Timeout: 2 * time.Second,
		OnStart: cb,
	})
	if err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}
	if gotDir != td {
		t.Fatalf("OnStart saw Dir=%q want %q", gotDir, td)
	}
	if !envOK {
		t.Fatalf("OnStart did not see expected env var FOO=BAR")
	}
}

func echoCmdArgs() (string, []string) {
	if isWindows() {
		return "cmd", []string{"/C", "echo", "ok"}
	}
	return "sh", []string{"-c", "echo ok"}
}

func isWindows() bool { return os.PathSeparator == '\\' }

func TestExec_WithStdinStdout(t *testing.T) {
	// Test custom Stdin and Stdout
	stdin := strings.NewReader("test input\n")
	var stdout strings.Builder

	cmd, args := echoCmdArgs()
	err := Exec(context.Background(), ExecOptions{
		Command: cmd,
		Args:    args,
		Stdin:   stdin,
		Stdout:  &stdout,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec with custom IO failed: %v", err)
	}
}

func TestExec_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Start a long-running command and cancel it
	cmd, args := sleepCmd(5 * time.Second)

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := Exec(ctx, ExecOptions{
		Command: cmd,
		Args:    args,
		TTK:     50 * time.Millisecond,
	})

	if err == nil {
		t.Fatal("Expected error from cancelled context")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Fatalf("Expected context error, got: %v", err)
	}
}

func TestExec_DefaultWorkDir(t *testing.T) {
	// Test that empty WorkDir defaults to current directory
	cmd, args := echoCmdArgs()
	err := Exec(context.Background(), ExecOptions{
		Command: cmd,
		Args:    args,
		WorkDir: "", // Should default to workdir
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Exec with default WorkDir failed: %v", err)
	}
}

func TestExec_OnStartCallback(t *testing.T) {
	var callbackCalled bool
	var cmdPid int

	cmd, args := echoCmdArgs()
	err := Exec(context.Background(), ExecOptions{
		Command: cmd,
		Args:    args,
		Timeout: 2 * time.Second,
		OnStart: func(c *exec.Cmd) {
			callbackCalled = true
			if c.Process != nil {
				cmdPid = c.Process.Pid
			}
		},
	})

	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if !callbackCalled {
		t.Fatal("OnStart callback was not called")
	}
	if cmdPid <= 0 {
		t.Fatal("OnStart callback did not receive valid process")
	}
}

func TestExec_InvalidCommand(t *testing.T) {
	err := Exec(context.Background(), ExecOptions{
		Command: "nonexistent-command-12345",
		Args:    []string{},
		Timeout: 2 * time.Second,
	})

	if err == nil {
		t.Fatal("Expected error for invalid command")
	}
}

func TestExec_CommandFailure(t *testing.T) {
	// Run a command that exits with non-zero status
	var cmd string
	var args []string
	if isWindows() {
		cmd = "cmd"
		args = []string{"/C", "exit", "1"}
	} else {
		cmd = "sh"
		args = []string{"-c", "exit 1"}
	}

	err := Exec(context.Background(), ExecOptions{
		Command: cmd,
		Args:    args,
		Timeout: 2 * time.Second,
	})

	if err == nil {
		t.Fatal("Expected error for command that exits with non-zero status")
	}
}

// Helper for sleep commands in tests
func sleepCmd(d time.Duration) (string, []string) {
	sec := int(d / time.Second)
	if sec <= 0 {
		sec = 1
	}
	if isWindows() {
		return "powershell", []string{"-Command", "Start-Sleep", "-Seconds", strconv.Itoa(sec)}
	}
	return "sh", []string{"-c", "sleep " + strconv.Itoa(sec)}
}
