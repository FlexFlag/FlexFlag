# GopherCon India 2026 — CFP Submission

> **Status**: Ready to submit  
> **Deadline**: August 23, 2026 IST  
> **Track**: Technical (standard 30-min slot)  
> **Fill in before submitting**: `[YOUR NAME]`, `[YOUR CITY]`, `[YOUR YEARS OF GO]`

---

## Title

**Sub-microsecond Feature Flags: What Go's Memory Model Actually Gives You**

---

## Abstract (≤300 words — paste this into the CFP form)

Feature flag evaluation looks simple on paper: look up a key, return a value. So why do most systems struggle past 5ms? And what does Go's memory model offer that genuinely changes the answer?

I built FlexFlag, an open-source feature flag system, while trying to answer this question. The talk follows three concrete Go insights that came out of that process — each one grounded in benchmarks, race-detector output, and `go tool pprof` profiles.

**Insight 1 — `atomic.Pointer[Snapshot]` beats `sync.RWMutex` under read-heavy load.**  
With 8 reader goroutines and 10,000 flags, a snapshot read (atomic Load + map lookup) costs ~16ns. The equivalent `sync.RWMutex.RLock()` + map lookup costs 116ns — a 7x difference. The reason is not writer blocking. Even with no writer running, RWMutex degrades ~3x under parallelism because `RLock()` does an atomic increment on a shared reader-count field that bounces between CPU cores (confirmed by pprof: 93.84% of time in `sync/atomic.(*Int32).Add`). Under saturated writes, RWMutex p999 spikes 64x, from 2µs to 131µs. The atomic snapshot p999 stays at 256ns. The GC is the reclamation mechanism — in C you would need hazard pointers.

**Insight 2 — `io.WriteString` is a 2-allocation trap.**  
`io.WriteString(h, key)` on an `fnv.sum64` writer allocates 72 bytes and takes ~190ns. Direct `h.Write([]byte(key))` allocates nothing and takes ~83ns. The reason is two separate escape-analysis violations: the concrete `fnv.sum64` struct escapes to the heap when `h` is boxed into `io.Writer` (8 bytes), and the `[]byte(s)` backing array escapes through the opaque `w.Write()` dispatch that the compiler cannot inline (64 bytes). Total: 72 bytes, 2 allocs — consistently measured at count=10 with zero variance.

**Insight 3 — goroutines belong at the connection level, not the per-request level.**  
Per-item goroutines are 3–9x *slower* than sequential at every small batch size (N=1–32 flags). A goroutine costs ~1µs. In-memory flag evaluation costs ~170ns. The math does not work. Chunked parallelism (GOMAXPROCS goroutines each handling N/GOMAXPROCS flags) crosses over at N≈256 — when each goroutine has enough work to amortise its creation cost. The right place for goroutines is at the connection level: one goroutine per SSE client scales to thousands of concurrent streams.

Attendees leave with: concrete benchmark methodology (DCE-proof patterns, log2 histogram, benchstat), three rules about Go's memory model they can apply to their own systems, and a skepticism for "unsafe is always faster" claims.

---

## Talk Outline (for reviewer, not submitted)

**0:00–2:00 — Hook**  
Live demo: feature flag evaluation taking 22ms per call (standard DB query). The claim: "we can get this to 16ns." How?

**2:00–8:00 — Insight 1: Copy-on-Write with `atomic.Pointer[Snapshot]`**  
- What RWMutex actually does: reader-count as shared mutable state  
- pprof live: 93.84% of parallel RWMutex time is `sync/atomic.(*Int32).Add`  
- The snapshot design: immutable map, atomic pointer swap, GC reclamation  
- Benchmark: 116ns (RWMutex) → 16–24ns (snapshot, timer-resolution noise). NoWriter baseline shows it is NOT writer blocking  
- Latency tail: p999 64x spike under saturation (2µs → 131µs RWMutex) vs 256ns snapshot  

**8:00–14:00 — Insight 2: The `io.WriteString` allocation trap**  
- FNV hashing for cache-key generation: the 2-alloc surprise  
- Escape analysis: why `h.Write([]byte(s))` = 0 allocs, `io.WriteString(h, s)` = 2 allocs  
- The unsafe.StringData trap: same allocs as the safe version — unnecessary complexity  
- Rule: check whether your writer implements `io.StringWriter` before using `io.WriteString`  

**14:00–20:00 — Insight 3: Goroutines at the right level**  
- Per-item goroutines: 170ns/flag work vs 1µs goroutine cost — always lose (N=1–1000)  
- Chunked parallel (GOMAXPROCS workers): crossover at N≈256 (32 flags/goroutine)  
- This is the actual lesson: goroutine granularity matters — chunk the work correctly  
- SSE fan-out: one goroutine per client, non-blocking channel writes vs blocking `fmt.Fprintf`  
- The rule: goroutines pay off when work per goroutine >> 1µs  

**20:00–26:00 — Benchmark methodology: how to not lie to yourself**  
- Dead-code elimination: `_ = f.Enabled` vs `sinkFlag = f` (package-level variable)  
- 10k-entry maps and rotating keys to defeat constant-key optimisation  
- benchstat for stability: geometric mean, variance, noise floor  
- Log2-bucket histogram for wall-clock latency without allocation in the measurement loop  
- Race detector: `go test -race ./...` — required before trusting any concurrent design  

**26:00–30:00 — What the numbers actually mean**  
- The honest attribution: 22ms (DB) → 500µs (warm cache) → 13µs (precomputed flag) → 3.5µs (response cache) → 16ns (flag store read)  
- Each reduction required a different technique. Caching did most of the work. Go-specific tools did the last two.  
- When does this matter? High-frequency evaluation (millions/sec), low-latency services  
- Q&A  

---

## Speaker Bio (≤100 words)

[YOUR NAME] is a software engineer based in [YOUR CITY] who has been writing Go for [YOUR YEARS OF GO]. They built FlexFlag as an open-source feature flag system to explore high-performance Go patterns in a real product context. When not benchmarking flag evaluation, they are investigating why the Go compiler sometimes knows more about escape analysis than they do.

GitHub: [YOUR GITHUB HANDLE]

---

## Notes for the Program Committee

The benchmarks referenced in this proposal are reproducible from the open-source repository. The key files:

- `internal/api/handlers/propagation_bench_test.go` — `atomic.Pointer` vs RWMutex, DCE-proof
- `internal/api/handlers/attribution_bench_test.go` — `io.WriteString` trap, response serialisation layers  
- `internal/api/handlers/batch_bench_test.go` — goroutine crossover benchmarks
- `benchmarks/latencyharness/main.go` — wall-clock p99/p999 latency harness
- `BENCHMARKS.md` — canonical numbers with machine specs and reproduction commands

All benchmark results cited were measured with `-race ./...` passing (zero data races). The `unsafe.StringData` claim was validated as unnecessary, which is mentioned explicitly in the talk — the goal is honest engineering, not performance hype.

---

## Key Numbers (for quick reviewer reference)

| Claim | Evidence | File | Status |
|-------|----------|------|--------|
| Snapshot read ~16–24ns vs RWMutex 116ns | BenchmarkPropagation_*, count=10, benchstat | propagation_bench_test.go | ✓ measured |
| RWMutex ~3x worse under parallelism alone | BenchmarkPropagation_RWMutex_SingleGoroutine (43ns) vs NoWriter (122ns) | propagation_bench_test.go | ✓ measured, ratio 2.8x |
| RWMutex p999 64x worse under saturation | latency harness highwrite: 2µs→131µs | benchmarks/latencyharness/ | ✓ measured |
| Snapshot p999 stays at 256ns | latency harness highwrite snapshot row | benchmarks/latencyharness/ | ✓ measured |
| io.WriteString = 2 allocs, 72B | BenchmarkHash_IOWriteString, count=10 (100% consistent) | attribution_bench_test.go | ✓ measured |
| h.Write([]byte(s)) = 0 allocs | BenchmarkHash_ByteConversion, count=10 (100% consistent) | attribution_bench_test.go | ✓ measured |
| Per-item goroutines always lose (N=1–1000) | BenchmarkBatch_*_LargeN_PerItem | batch_bench_test.go | ✓ measured |
| Chunked parallel crossover at N≈256 | BenchmarkBatch_Targeting_LargeN_Chunked: 43µs vs 49µs sequential at N=256 | batch_bench_test.go | ✓ measured |
| Response cache HIT = 3.5µs | BenchmarkLayer1 | attribution_bench_test.go | ✓ measured |
| Race detector: PASS | go test -race ./... | all test files | ✓ verified |
| Go version: 1.24.5 darwin/arm64 | go version | — | required context for alloc claims |
