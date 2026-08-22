// Package handlers (internal test) — propagation benchmarks.
//
// These benchmarks are in package handlers (not handlers_test) so they can
// access the unexported flagStore and Snapshot fields directly. This is
// intentional: we are benchmarking the synchronization mechanism itself,
// not the HTTP handler.
//
// Run:
//
//	go test ./internal/api/handlers/ -run=^$ \
//	    -bench=BenchmarkPropagation \
//	    -benchmem -benchtime=5s -count=3
//
// The two benchmarks model the same workload:
//   - GOMAXPROCS*8 goroutines reading the flag store continuously
//   - 1 background goroutine writing a new flag value every 1ms (~1000/sec)
//
// This ratio is realistic: flag updates are rare; evaluations are constant.
//
// What to look for in the output:
//   - RWMutex ns/op climbs relative to the no-writer baseline because
//     Lock() blocks all RLock() holders for the duration of the copy.
//   - AtomicSnapshot ns/op stays flat regardless of write frequency because
//     atomic.Pointer.Load() is never blocked by a writer.
//
// DCE-proof pattern: all variants use a 10k-entry map, rotating keys (i%10000),
// and assign the result to the package-level sinkFlag variable. This prevents
// the compiler from eliminating the map lookup via dead-code elimination.
package handlers

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// sinkFlag is a package-level variable. Assigning to it forces the compiler
// to materialise the pointer — it cannot be optimised away.
// (Moved here from dce_audit_test.go which has been deleted.)
var sinkFlag *PrecomputedFlag

// syntheticFlag returns a minimal PrecomputedFlag for benchmark use.
// Called in the writer goroutine only; not on the read hot path.
func syntheticFlag() *PrecomputedFlag {
	return &PrecomputedFlag{
		Key:          "bench-flag",
		Enabled:      true,
		DefaultValue: true,
		HasTargeting: false,
	}
}

// make10kFlags builds a 10,000-entry flag map used by all propagation benchmarks.
// 10k entries ensures the map is large enough to defeat constant-key optimisation
// and represent a realistic production flag store.
func make10kFlags() map[string]*PrecomputedFlag {
	m := make(map[string]*PrecomputedFlag, 10_000)
	for i := 0; i < 10_000; i++ {
		key := fmt.Sprintf("flag-%05d:production", i)
		m[key] = &PrecomputedFlag{Key: key, Enabled: true}
	}
	return m
}

// make10kKeys returns the pre-built key slice so the read hot path avoids
// fmt.Sprintf per iteration.
func make10kKeys() []string {
	keys := make([]string, 10_000)
	for i := range keys {
		keys[i] = fmt.Sprintf("flag-%05d:production", i)
	}
	return keys
}

// BenchmarkPropagation_RWMutex is the baseline: the approach UltraFastHandler
// used before this refactor. A sync.RWMutex protects a mutable map.
//
// Every write acquires a full write lock, which blocks every concurrent
// RLock() call. Under high read concurrency, writers cause latency spikes
// across all goroutines — not just the ones that happened to call RLock()
// at the wrong moment, but any that try to RLock() while the write lock
// is pending (Go's sync.RWMutex is fair: a waiting writer blocks new readers).
//
// DCE-proof: uses 10k map + rotating keys + sinkFlag assignment.
func BenchmarkPropagation_RWMutex(b *testing.B) {
	var mu sync.RWMutex
	m := make10kFlags()
	keys := make10kKeys()
	n := len(keys)

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond) // ~1000 writes/sec
		defer ticker.Stop()
		writeIdx := 0
		for {
			select {
			case <-ticker.C:
				key := fmt.Sprintf("flag-%05d:production", writeIdx%10_000)
				f := &PrecomputedFlag{Key: key, Enabled: true}
				mu.Lock()
				m[key] = f
				mu.Unlock()
				writeIdx++
			case <-stop:
				return
			}
		}
	}()
	defer close(stop)

	b.SetParallelism(8) // 8 * GOMAXPROCS reader goroutines
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mu.RLock()
			f := m[keys[i%n]]
			mu.RUnlock()
			sinkFlag = f // DCE-proof: forces pointer escape
			i++
		}
	})
}

// BenchmarkPropagation_AtomicSnapshot is the new approach: atomic.Pointer[Snapshot].
//
// Readers call flagStore.load() — one atomic.Pointer.Load(), zero locks.
// The writer calls flagStore.set() — acquires writeMu, copies the map,
// stores a new Snapshot. Readers and the writer never contend on any lock.
// The old Snapshot is collected by the GC once all readers holding it finish.
//
// This is why Go makes this pattern clean: in C/C++ you would need hazard
// pointers or RCU to safely reclaim the old snapshot. In Go, the GC is
// the reclamation mechanism and it requires no application-level code.
//
// DCE-proof: uses 10k map + rotating keys + sinkFlag assignment.
func BenchmarkPropagation_AtomicSnapshot(b *testing.B) {
	fs := newFlagStore()
	fs.reset(make10kFlags())
	keys := make10kKeys()
	n := len(keys)

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond) // same write rate as RWMutex benchmark
		defer ticker.Stop()
		writeIdx := 0
		for {
			select {
			case <-ticker.C:
				key := fmt.Sprintf("flag-%05d:production", writeIdx%10_000)
				fs.set(key, &PrecomputedFlag{Key: key, Enabled: true})
				writeIdx++
			case <-stop:
				return
			}
		}
	}()
	defer close(stop)

	b.SetParallelism(8)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			snap := fs.load()
			f := snap.flags[keys[i%n]]
			sinkFlag = f // DCE-proof: forces pointer escape
			i++
		}
	})
}

// BenchmarkPropagation_RWMutex_NoWriter shows the floor: read throughput
// with zero write pressure. Comparing this against BenchmarkPropagation_RWMutex
// shows the cost of the writer goroutine on reader throughput (should be significant
// for RWMutex but near zero for AtomicSnapshot).
// Comparing it against BenchmarkPropagation_AtomicSnapshot_NoWriter shows the base
// cost of the synchronization mechanism itself.
//
// DCE-proof: uses 10k map + rotating keys + sinkFlag assignment.
func BenchmarkPropagation_RWMutex_NoWriter(b *testing.B) {
	var mu sync.RWMutex
	m := make10kFlags()
	keys := make10kKeys()
	n := len(keys)

	b.SetParallelism(8)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mu.RLock()
			f := m[keys[i%n]]
			mu.RUnlock()
			sinkFlag = f // DCE-proof: forces pointer escape
			i++
		}
	})
}

// BenchmarkPropagation_RWMutex_SingleGoroutine is the single-goroutine floor for
// RWMutex. No cache-line bouncing — only one goroutine does RLock/RUnlock, so the
// atomic reader-count field is never contended. Compare against
// BenchmarkPropagation_RWMutex_NoWriter (8 goroutines) to quantify reader-reader
// cache-line contention directly. Expected ratio: ~4-5x.
//
// DCE-proof: uses 10k map + rotating keys + sinkFlag assignment.
func BenchmarkPropagation_RWMutex_SingleGoroutine(b *testing.B) {
	var mu sync.RWMutex
	m := make10kFlags()
	keys := make10kKeys()
	n := len(keys)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mu.RLock()
		f := m[keys[i%n]]
		mu.RUnlock()
		sinkFlag = f // DCE-proof: forces pointer escape
	}
}

// BenchmarkPropagation_AtomicSnapshot_NoWriter is the atomic snapshot floor:
// read throughput with zero write pressure.
//
// DCE-proof: uses 10k map + rotating keys + sinkFlag assignment.
func BenchmarkPropagation_AtomicSnapshot_NoWriter(b *testing.B) {
	fs := newFlagStore()
	fs.reset(make10kFlags())
	keys := make10kKeys()
	n := len(keys)

	b.SetParallelism(8)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			snap := fs.load()
			f := snap.flags[keys[i%n]]
			sinkFlag = f // DCE-proof: forces pointer escape
			i++
		}
	})
}
