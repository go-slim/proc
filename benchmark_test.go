package proc

import (
	"bytes"
	"io"
	"testing"
)

// BenchmarkPath measures the performance of Path function
func BenchmarkPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Path("dir1", "dir2", "file.txt")
	}
}

// BenchmarkPathf measures the performance of Pathf function
func BenchmarkPathf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Pathf("dir/%s/%d.txt", "subdir", 123)
	}
}

// BenchmarkDebugf measures logging performance
func BenchmarkDebugf(b *testing.B) {
	old := Logger
	Logger = io.Discard
	defer func() { Logger = old }()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debugf("test message %d", i)
	}
}

// BenchmarkDebugfWithBuffer measures logging performance with real buffer
func BenchmarkDebugfWithBuffer(b *testing.B) {
	old := Logger
	var buf bytes.Buffer
	Logger = &buf
	defer func() { Logger = old }()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debugf("test message %d", i)
	}
}

// BenchmarkPid measures process ID retrieval
func BenchmarkPid(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Pid()
	}
}

// BenchmarkName measures process name retrieval
func BenchmarkName(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Name()
	}
}

// BenchmarkWorkDir measures working directory retrieval
func BenchmarkWorkDir(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WorkDir()
	}
}

// BenchmarkContext measures context retrieval
func BenchmarkContext(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Context()
	}
}
