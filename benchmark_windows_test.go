//go:build windows

package proc

import (
	"syscall"
	"testing"
)

// BenchmarkNotify measures the performance of signal notification
// On Windows, we use SIGTERM instead of SIGUSR1
func BenchmarkNotify(b *testing.B) {
	// Register a single listener
	var counter int
	id := On(syscall.SIGTERM, func() { counter++ })
	defer Cancel(id)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Notify(syscall.SIGTERM)
	}
}

// BenchmarkNotifyParallel measures concurrent notification performance
// On Windows, we use SIGTERM instead of SIGUSR1
func BenchmarkNotifyParallel(b *testing.B) {
	// Register multiple listeners
	var counter int
	var ids []uint32
	for i := 0; i < 10; i++ {
		id := On(syscall.SIGTERM, func() { counter++ })
		ids = append(ids, id)
	}
	defer Cancel(ids...)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Notify(syscall.SIGTERM)
		}
	})
}
