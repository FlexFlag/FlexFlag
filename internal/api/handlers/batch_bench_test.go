// Batch evaluation benchmarks — package handlers (internal).
//
// These benchmarks answer the question: at what batch size do goroutines
// beat sequential evaluation, and what does the Snapshot design enable?
//
// Two flag types are benchmarked:
//   - Simple: enabled flag, no targeting. Evaluation is a struct field read.
//   - Targeting: flag with 8 attribute rules + percentage rollout.
//     Evaluation involves string matching and FNV hash computation.
//
// For simple flags, goroutine overhead (~1µs per goroutine) exceeds the
// parallelism benefit until N is very large. For targeting flags, each
// evaluation costs enough that goroutines start winning around N=5.
//
// The key insight the Snapshot design unlocks: goroutines read from the
// same immutable Snapshot with zero locks. With a RWMutex map, N parallel
// readers contend on the mutex's internal reader count — cache-line bouncing
// that grows with GOMAXPROCS.
//
// Run:
//
//	go test ./internal/api/handlers/ -run=^$ \
//	    -bench=BenchmarkBatch \
//	    -benchmem -benchtime=3s -count=3
package handlers

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/flexflag/flexflag/pkg/types"
)

// ── flag constructors ─────────────────────────────────────────────────────────

func simpleFlag(key string) *PrecomputedFlag {
	return &PrecomputedFlag{
		Key:          key,
		Enabled:      true,
		DefaultValue: true,
		HasTargeting: false,
	}
}

func targetingFlag(key string) *PrecomputedFlag {
	// 8 attribute-based rules + a percentage rollout.
	// Each evaluation: 8 map lookups + up to 8 string comparisons + FNV hash.
	rules := make([]types.TargetingRule, 8)
	for i := range rules {
		rules[i] = types.TargetingRule{
			Attribute: fmt.Sprintf("attr-%d", i),
			Operator:  "eq",
			Values:    []string{fmt.Sprintf("value-%d", i)},
			Variation: "v-enabled",
		}
	}
	return &PrecomputedFlag{
		Key:          key,
		Enabled:      true,
		DefaultValue: false,
		HasTargeting: true,
		VariationValues: map[string]interface{}{
			"v-enabled": true,
		},
		Targeting: &types.TargetingConfig{
			Rules: rules,
			Rollout: &types.RolloutConfig{
				Variations: []types.VariationRollout{
					{VariationID: "v-enabled", Weight: 50},
				},
			},
		},
	}
}

// ── snapshot population ───────────────────────────────────────────────────────

func snapshotWithN(n int, constructor func(string) *PrecomputedFlag) *flagStore {
	fs := newFlagStore()
	m := make(map[string]*PrecomputedFlag, n)
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("flag-%04d:production", i)
		m[key] = constructor(fmt.Sprintf("flag-%04d", i))
	}
	fs.reset(m)
	return fs
}

func batchKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("flag-%04d", i)
	}
	return keys
}

// ── sequential helper (what UltraFastBatchEvaluate does below threshold) ──────

func runSequential(h *UltraFastHandler, snap *Snapshot, keys []string, req *EvaluateRequest) []batchResult {
	results := make([]batchResult, len(keys))
	for i, key := range keys {
		results[i] = h.evaluateOne(snap, key, "production", req)
	}
	return results
}

// ── parallel helper (what UltraFastBatchEvaluate does above threshold) ────────

func runParallel(h *UltraFastHandler, snap *Snapshot, keys []string, req *EvaluateRequest) []batchResult {
	results := make([]batchResult, len(keys))
	var wg sync.WaitGroup
	wg.Add(len(keys))
	for i, key := range keys {
		i, key := i, key
		go func() {
			defer wg.Done()
			results[i] = h.evaluateOne(snap, key, "production", req)
		}()
	}
	wg.Wait()
	return results
}

// ── chunked-parallel helper ───────────────────────────────────────────────────
//
// runChunked splits keys into GOMAXPROCS chunks and evaluates each chunk
// sequentially within one goroutine. This amortises goroutine-creation cost
// (~1µs per goroutine) across many flags, unlike per-item goroutines.
// At small N the overhead dominates; at large N the parallelism should win.

func runChunked(h *UltraFastHandler, snap *Snapshot, keys []string, req *EvaluateRequest) []batchResult {
	n := len(keys)
	nWorkers := runtime.GOMAXPROCS(0)
	if nWorkers > n {
		nWorkers = n
	}
	results := make([]batchResult, n)
	var wg sync.WaitGroup
	wg.Add(nWorkers)
	chunkSize := (n + nWorkers - 1) / nWorkers
	for w := 0; w < nWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > n {
			end = n
		}
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				results[i] = h.evaluateOne(snap, keys[i], "production", req)
			}
		}(start, end)
	}
	wg.Wait()
	return results
}

// ── benchmark matrix ──────────────────────────────────────────────────────────

var batchSizes = []int{1, 4, 8, 16, 32}

// largeBatchSizes tests batch sizes large enough to amortise goroutine overhead.
// At N=256 (GOMAXPROCS*32), chunked-parallel should start winning for targeting flags.
var largeBatchSizes = []int{64, 128, 256, 512, 1000}

// BenchmarkBatch_Simple_Sequential: N simple flags, no goroutines.
func BenchmarkBatch_Simple_Sequential(b *testing.B) {
	h := &UltraFastHandler{store: newFlagStore()}
	req := &EvaluateRequest{UserID: "user-bench"}

	for _, n := range batchSizes {
		n := n
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			fs := snapshotWithN(n, simpleFlag)
			snap := fs.load()
			keys := batchKeys(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = runSequential(h, snap, keys, req)
			}
		})
	}
}

// BenchmarkBatch_Simple_Parallel: N simple flags, one goroutine per flag.
// Shows goroutine overhead relative to work per flag.
func BenchmarkBatch_Simple_Parallel(b *testing.B) {
	h := &UltraFastHandler{store: newFlagStore()}
	req := &EvaluateRequest{UserID: "user-bench"}

	for _, n := range batchSizes {
		n := n
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			fs := snapshotWithN(n, simpleFlag)
			snap := fs.load()
			keys := batchKeys(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = runParallel(h, snap, keys, req)
			}
		})
	}
}

// BenchmarkBatch_Targeting_Sequential: N flags with targeting rules, sequential.
// Each evaluation: 8 map lookups + string comparisons + FNV hash.
func BenchmarkBatch_Targeting_Sequential(b *testing.B) {
	h := &UltraFastHandler{store: newFlagStore()}
	req := &EvaluateRequest{
		UserID: "user-bench",
		Attributes: map[string]interface{}{
			"attr-0": "value-0", // matches rule 0, stops early
		},
	}

	for _, n := range batchSizes {
		n := n
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			fs := snapshotWithN(n, targetingFlag)
			snap := fs.load()
			keys := batchKeys(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = runSequential(h, snap, keys, req)
			}
		})
	}
}

// BenchmarkBatch_Targeting_Parallel: N flags with targeting rules, goroutines.
// Compare ns/op against Sequential at each N to find the crossover point.
func BenchmarkBatch_Targeting_Parallel(b *testing.B) {
	h := &UltraFastHandler{store: newFlagStore()}
	req := &EvaluateRequest{
		UserID: "user-bench",
		Attributes: map[string]interface{}{
			"attr-0": "value-0",
		},
	}

	for _, n := range batchSizes {
		n := n
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			fs := snapshotWithN(n, targetingFlag)
			snap := fs.load()
			keys := batchKeys(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = runParallel(h, snap, keys, req)
			}
		})
	}
}

// BenchmarkBatch_Targeting_RWMutex_Sequential: same work but readers must
// acquire RLock on every flag lookup. Shows the cost of bringing a mutex into
// the per-flag evaluation loop vs reading from an immutable Snapshot.
func BenchmarkBatch_Targeting_RWMutex_Sequential(b *testing.B) {
	// Simulate the pre-Snapshot approach: a mutable map + RWMutex.
	var mu sync.RWMutex
	m := make(map[string]*PrecomputedFlag)

	for _, n := range batchSizes {
		n := n
		for i := 0; i < n; i++ {
			key := fmt.Sprintf("flag-%04d:production", i)
			m[key] = targetingFlag(fmt.Sprintf("flag-%04d", i))
		}

		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			keys := batchKeys(n)
			req := &EvaluateRequest{
				UserID:     "user-bench",
				Attributes: map[string]interface{}{"attr-0": "value-0"},
			}
			h := &UltraFastHandler{store: newFlagStore()}
			b.ResetTimer()
			b.ReportAllocs()
			for iter := 0; iter < b.N; iter++ {
				results := make([]batchResult, len(keys))
				for i, key := range keys {
					mu.RLock()
					flag := m[key+":production"]
					mu.RUnlock()
					// evaluate inline (same logic as evaluateOne)
					if flag == nil {
						results[i] = batchResult{key: key, found: false}
						continue
					}
					req.FlagKey = key
					var value interface{}
					var reason string
					if !flag.Enabled {
						value, reason = flag.DefaultValue, "flag_disabled"
					} else if !flag.HasTargeting {
						value, reason = flag.DefaultValue, "default"
					} else {
						if matched := h.fastTargetingEvaluation(flag, req); matched != nil {
							value, reason = matched, "rule_match"
						} else {
							value, reason = flag.DefaultValue, "default"
						}
					}
					results[i] = batchResult{key: key, value: value, reason: reason, found: true}
				}
				_ = results
			}
		})
	}
}

// ── Large-N benchmarks: crossover point for per-item vs chunked parallelism ──
//
// At small N, goroutine creation cost (~1µs each) swamps evaluation time.
// At large N, amortised goroutine cost becomes negligible, and chunked-parallel
// should win if evaluation work is CPU-bound. For in-memory flag evaluation
// (~170ns/flag), the crossover requires very large N.

// BenchmarkBatch_Targeting_LargeN_Sequential: sequential baseline at large N.
func BenchmarkBatch_Targeting_LargeN_Sequential(b *testing.B) {
	h := &UltraFastHandler{store: newFlagStore()}
	req := &EvaluateRequest{
		UserID: "user-bench",
		Attributes: map[string]interface{}{
			"attr-0": "value-0",
		},
	}

	for _, n := range largeBatchSizes {
		n := n
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			fs := snapshotWithN(n, targetingFlag)
			snap := fs.load()
			keys := batchKeys(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = runSequential(h, snap, keys, req)
			}
		})
	}
}

// BenchmarkBatch_Targeting_LargeN_PerItem: one goroutine per flag (unbounded).
// Expected: still worse than sequential at all sizes due to scheduler overhead.
func BenchmarkBatch_Targeting_LargeN_PerItem(b *testing.B) {
	h := &UltraFastHandler{store: newFlagStore()}
	req := &EvaluateRequest{
		UserID: "user-bench",
		Attributes: map[string]interface{}{
			"attr-0": "value-0",
		},
	}

	for _, n := range largeBatchSizes {
		n := n
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			fs := snapshotWithN(n, targetingFlag)
			snap := fs.load()
			keys := batchKeys(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = runParallel(h, snap, keys, req)
			}
		})
	}
}

// BenchmarkBatch_Targeting_LargeN_Chunked: GOMAXPROCS goroutines, each handling
// N/GOMAXPROCS flags sequentially. Amortises goroutine cost across many flags.
// This is the crossover-finding benchmark. If it beats Sequential at any N,
// report the crossover point.
func BenchmarkBatch_Targeting_LargeN_Chunked(b *testing.B) {
	h := &UltraFastHandler{store: newFlagStore()}
	req := &EvaluateRequest{
		UserID: "user-bench",
		Attributes: map[string]interface{}{
			"attr-0": "value-0",
		},
	}

	for _, n := range largeBatchSizes {
		n := n
		b.Run(fmt.Sprintf("N=%d_workers=%d", n, runtime.GOMAXPROCS(0)), func(b *testing.B) {
			fs := snapshotWithN(n, targetingFlag)
			snap := fs.load()
			keys := batchKeys(n)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = runChunked(h, snap, keys, req)
			}
		})
	}
}
