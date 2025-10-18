package proc

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestDebugf_WithNilLogger(t *testing.T) {
	// Test that debugf doesn't panic with nil Logger
	old := Logger
	Logger = nil
	defer func() { Logger = old }()

	// Should not panic
	debugf("test message")
}

func TestDebugf_WithDiscard(t *testing.T) {
	// Test that debugf doesn't output when Logger is io.Discard
	old := Logger
	Logger = io.Discard
	defer func() { Logger = old }()

	// Should not panic and not output
	debugf("test message")
}

func TestDebugf_FormatsCorrectly(t *testing.T) {
	var buf bytes.Buffer
	old := Logger
	Logger = &buf
	defer func() { Logger = old }()

	debugf("hello %s, number %d", "world", 42)
	output := buf.String()

	if !strings.Contains(output, "hello world") {
		t.Fatalf("Expected 'hello world' in output, got: %q", output)
	}
	if !strings.Contains(output, "number 42") {
		t.Fatalf("Expected 'number 42' in output, got: %q", output)
	}
}

func TestDebugf_AddsNewline(t *testing.T) {
	var buf bytes.Buffer
	old := Logger
	Logger = &buf
	defer func() { Logger = old }()

	debugf("test")
	output := buf.String()

	if !strings.HasSuffix(output, "\n") {
		t.Fatalf("Expected output to end with newline, got: %q", output)
	}
}

func TestDebugf_MultipleMessages(t *testing.T) {
	var buf bytes.Buffer
	old := Logger
	Logger = &buf
	defer func() { Logger = old }()

	debugf("message 1")
	debugf("message 2")
	debugf("message 3")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d: %q", len(lines), output)
	}
	if !strings.Contains(lines[0], "message 1") {
		t.Fatalf("Expected 'message 1' in first line, got: %q", lines[0])
	}
	if !strings.Contains(lines[1], "message 2") {
		t.Fatalf("Expected 'message 2' in second line, got: %q", lines[1])
	}
	if !strings.Contains(lines[2], "message 3") {
		t.Fatalf("Expected 'message 3' in third line, got: %q", lines[2])
	}
}
