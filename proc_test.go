package proc

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestProcBasics(t *testing.T) {
	if Pid() <= 0 {
		t.Fatalf("Pid should be > 0, got %d", Pid())
	}
	if Name() == "" {
		t.Fatalf("Name should not be empty")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd error: %v", err)
	}
	if WorkDir() != wd {
		t.Fatalf("WorkDir mismatch: got %q want %q", WorkDir(), wd)
	}
	p := Path("a", "b", "c.txt")
	if !strings.HasSuffix(p, filepath.Join("a", "b", "c.txt")) {
		t.Fatalf("Path unexpected: %q", p)
	}
	pf := Pathf("%s/%s", "d", "e")
	if !strings.HasSuffix(pf, filepath.Join("d", "e")) {
		t.Fatalf("Pathf unexpected: %q", pf)
	}
	if Context() == nil {
		t.Fatalf("Context should not be nil")
	}
}

func TestDebugfWritesToLogger(t *testing.T) {
	var buf bytes.Buffer
	old := Logger
	Logger = &buf
	t.Cleanup(func() { Logger = old })

	debugf("hello %s", "world")
	if got := buf.String(); !strings.Contains(got, "hello world") {
		t.Fatalf("unexpected logger output: %q", got)
	}
}

func TestSignal_On_Once_Cancel_Notify(t *testing.T) {
	// Use SIGTERM which is registered by default in registerSignalListener.
	// We call Notify directly (not via OS) so no os.Exit occurs.
	var onCnt, onceCnt int
	onID := On(syscall.SIGTERM, func() { onCnt++ })
	Once(syscall.SIGTERM, func() { onceCnt++ })

	if !Notify(syscall.SIGTERM) {
		t.Fatalf("first Notify returned false")
	}
	if onCnt != 1 || onceCnt != 1 {
		t.Fatalf("after first notify: on=%d once=%d", onCnt, onceCnt)
	}

	// Once should not fire again; On should.
	if !Notify(syscall.SIGTERM) {
		t.Fatalf("second Notify returned false")
	}
	if onCnt != 2 || onceCnt != 1 {
		t.Fatalf("after second notify: on=%d once=%d", onCnt, onceCnt)
	}

	// Cancel the On listener and ensure no further calls.
	Cancel(onID)
	if Notify(syscall.SIGTERM) {
		t.Fatalf("third Notify should return false when no listeners remain")
	}
	if onCnt != 2 || onceCnt != 1 {
		t.Fatalf("after third notify (post-cancel): on=%d once=%d", onCnt, onceCnt)
	}
}

func TestExec_Success(t *testing.T) {
	// Use a trivial command that exits 0 on each platform
	cmd, args := trivialEcho()
	err := Exec(context.Background(), ExecOptions{Command: cmd, Args: args, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("Exec success expected, got error: %v", err)
	}
}

func TestExec_Timeout(t *testing.T) {
	// Use a short sleep with a shorter timeout
	cmd, args := trivialSleep(2 * time.Second)
	err := Exec(context.Background(), ExecOptions{Command: cmd, Args: args, Timeout: 50 * time.Millisecond, TTK: 50 * time.Millisecond})
	if err == nil {
		t.Fatalf("Exec should timeout, got nil error")
	}
}

// Helpers to select commands cross-platform
func trivialEcho() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", "echo", "ok"}
	}
	return "sh", []string{"-c", "echo ok"}
}

func trivialSleep(d time.Duration) (string, []string) {
	sec := int(d / time.Second)
	if sec <= 0 {
		sec = 1
	}
	if runtime.GOOS == "windows" {
		// powershell Start-Sleep -Seconds N
		return "powershell", []string{"-Command", "Start-Sleep", "-Seconds", strconv.Itoa(sec)}
	}
	return "sh", []string{"-c", "sleep " + strconv.Itoa(sec)}
}
