package proc

import (
	"context"
	"os"
	"os/exec"
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
