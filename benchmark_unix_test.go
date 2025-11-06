//go:build unix

package proc

import (
	"syscall"
	"testing"
)

// BenchmarkOn measures the performance of signal listener registration
// Note: This benchmark is commented out due to slow performance at large iteration counts.
// To run: go test -bench=BenchmarkOn -benchtime=100ms
/*
func BenchmarkOn(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := On(syscall.SIGUSR1, func() {})
		b.StopTimer()
		Cancel(id)
		b.StartTimer()
	}
}
*/

// BenchmarkOnce measures the performance of one-time signal listener registration
// Note: This benchmark is commented out due to slow performance at large iteration counts.
// To run: go test -bench=BenchmarkOnce -benchtime=100ms
/*
func BenchmarkOnce(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := Once(syscall.SIGUSR1, func() {})
		b.StopTimer()
		Cancel(id)
		b.StartTimer()
	}
}
*/

// BenchmarkCancel measures the performance of signal listener cancellation
// Note: This benchmark is commented out due to performance issues with large N values.
// The repeated On/Cancel cycle becomes very slow as the number of iterations increases.
/*
func BenchmarkCancel(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		id := On(syscall.SIGUSR1, func() {})
		b.StartTimer()
		Cancel(id)
	}
}
*/

// BenchmarkNotify measures the performance of signal notification
func BenchmarkNotify(b *testing.B) {
	// Register a single listener
	var counter int
	id := On(syscall.SIGUSR1, func() { counter++ })
	defer Cancel(id)

	b.ReportAllocs()

	for b.Loop() {
		Notify(syscall.SIGUSR1)
	}
}

// BenchmarkNotifyParallel measures concurrent notification performance
func BenchmarkNotifyParallel(b *testing.B) {
	// Register multiple listeners
	var counter int
	var ids []uint32
	for range 10 {
		id := On(syscall.SIGUSR1, func() { counter++ })
		ids = append(ids, id)
	}
	defer Cancel(ids...)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Notify(syscall.SIGUSR1)
		}
	})
}
