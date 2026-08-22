# FlexFlag Benchmark Results

Canonical performance numbers for the FlexFlag flag evaluation system.
All numbers in this file are measured, not estimated, unless flagged as `[estimated]`.

## Machine Specs

```
[FILL IN: cpu, ram, go version output]
```

Example:
```
cpu: Apple M2 (arm64, 8-core)
go version go1.24.5 darwin/arm64
```

## Canonical Latency Ladder

| Tier | Latency | What it measures | Source |
|------|---------|-----------------|--------|
| Standard handler | ~22ms | DB query + json.Unmarshal per request | [server-reported, not Go benchmark] |
| Optimized handler | ~500µs | Warm cache, json.Unmarshal per request | [estimated from handler code] |
| Ultra-fast, cache miss | ~13µs | Precomputed flag hit, json.Marshal once per unique request | BenchmarkLayer2, measured |
| Ultra-fast, cache hit | ~3.5µs | Pre-serialized bytes returned from response cache | BenchmarkLayer1, measured |
| Flag store read, atomic | ~16ns | atomic.Pointer.Load() + 10k-map lookup (DCE-proof) | BenchmarkPropagation_AtomicSnapshot, measured |
| Flag store read, RWMutex | ~114ns | sync.RWMutex.RLock() + 10k-map lookup (DCE-proof) | BenchmarkPropagation_RWMutex, measured |

**Important corrections vs prior claims:**
- The previously-cited "0.1ms / 317x" referred to the ultra-fast handler broadly.
  Measured: response-cache HIT is ~3.5µs; cache MISS is ~13µs.
  Under concurrent mixed load (BenchmarkLayer3) it is ~8µs.
- The 317x figure cannot be reproduced from these benchmarks. It was likely
  measured against a different baseline or under different conditions.
- The ratio (atomic vs RWMutex) at 8 parallel goroutines: ~7x from stable runs, 4.8x from
  benchstat geometric mean at count=10 (inflated by macOS timer quantization — the atomic
  snapshot operation at ~16ns is below macOS timer resolution of ~41ns, causing 17/24/41ns
  quantization noise).
- The single-goroutine vs 8-goroutine RWMutex ratio: 43ns (single) vs 122ns (8 goroutines)
  = **2.8x** degradation from cache-line contention. Previous claim of 4.2x was from a
  deleted benchmark that did not include the 10k map lookup (likely DCE artifact).

## Reproduction Commands

### Full propagation + hash + response suite (count=5, benchtime=2s)

```bash
go test ./internal/api/handlers/ -run=^$ \
    -bench="BenchmarkPropagation|BenchmarkHash|BenchmarkResponse" \
    -benchmem -benchtime=2s -count=5 -timeout=120s \
    | tee benchmarks/results/bench_raw.txt
```

### Layer benchmarks (full HTTP handler stack)

```bash
go test ./internal/api/handlers/ -run=^$ \
    -bench="BenchmarkLayer" \
    -benchmem -benchtime=2s -count=5 -timeout=120s \
    | tee benchmarks/results/bench_layer.txt
```

### Batch benchmarks

```bash
go test ./internal/api/handlers/ -run=^$ \
    -bench="BenchmarkBatch" \
    -benchmem -benchtime=1s -count=5 -timeout=300s \
    | tee benchmarks/results/bench_batch.txt
```

### Statistical analysis with benchstat

```bash
go install golang.org/x/perf/cmd/benchstat@latest
benchstat benchmarks/results/bench_raw.txt > benchmarks/results/benchstat.txt
```

### CPU profile (parallel RWMutex, shows lock contention)

```bash
go test ./internal/api/handlers/ -run=^$ \
    -bench=BenchmarkAudit_RWMutex_Sink \
    -benchtime=10s -count=1 \
    -cpuprofile=benchmarks/results/rwmutex_cpu.prof
go tool pprof -top benchmarks/results/rwmutex_cpu.prof | head -20
```

### Standalone latency harness (wall-clock p50/p99/p999 with writers)

```bash
for impl in rwmutex snapshot; do
    for scenario in none steady burst highwrite; do
        go run ./benchmarks/latencyharness/ \
            -impl=$impl -scenario=$scenario \
            -duration=10 -readers=8 \
            -csv=benchmarks/results/latency.csv
    done
done
```

### Race detector

```bash
go test -race ./... | tee benchmarks/results/race_test.txt
```

## Benchstat Table

Measured on [FILL IN machine specs]. Run `count=5` with `benchtime=2s` (propagation/hash/response)
and `benchtime=1s` (batch), single run for Layer benchmarks.

```
goos: darwin
goarch: arm64
pkg: github.com/flexflag/flexflag/internal/api/handlers
cpu: Apple M2
                                          │  sec/op   │
Propagation_RWMutex-8                     │ 114.0n    │  ← reader-reader contention (8 goroutines, 10k map)
Propagation_AtomicSnapshot-8             │  16.39n   │  ← atomic.Pointer.Load() + map lookup
Propagation_RWMutex_NoWriter-8           │ 115.6n    │  ← same cost even with no writer (cache-line bouncing)
Propagation_AtomicSnapshot_NoWriter-8    │  12.71n   │  ← floor: one atomic load + map lookup
Hash_ByteConversion-8                    │  82.59n   │  ← 0 allocs (compiler devirtualizes h.Write)
Hash_IOWriteString-8                     │ 187.0n    │  ← 2 allocs (interface dispatch prevents escape analysis)
Hash_UnsafeStringData-8                  │  83.68n   │  ← 0 allocs, same as ByteConversion (unsafe was unnecessary)
Response_MarshalPerRequest-8             │   2.794µ  │  ← 15 allocs, json.Marshal every request
Response_PreSerialized-8                 │   1.405µ  │  ← 8 allocs, write pre-built bytes to ResponseWriter
Layer1_ResponseCacheHit-8               │   3.491µ  │  ← full HTTP stack, response cache HIT path
Layer2_PrecomputedHit_CacheMiss-8       │  12.65µ   │  ← full HTTP stack, precomputed flag, cache MISS
Layer3_DualMutex_ConcurrentMixed-8      │   8.043µ  │  ← concurrent mixed load (parallel readers)
Batch_Simple_Sequential/N=1-8           │  134.5n   │
Batch_Simple_Sequential/N=32-8          │   3.102µ  │
Batch_Targeting_Sequential/N=1-8        │  197.2n   │
Batch_Targeting_Sequential/N=32-8       │   5.491µ  │
```

Full output: see `benchmarks/results/benchstat.txt`

## Latency Harness Table

Wall-clock measurements. Includes `time.Now()/time.Since()` overhead (~50-100ns).
Use for relative comparison, not as raw lock costs.

```
impl=rwmutex  scenario=none       readers=8  p50=512ns  p95=1024ns  p99=1024ns   p999=2048ns    total=73789317
impl=rwmutex  scenario=steady     readers=8  p50=512ns  p95=1024ns  p99=1024ns   p999=2048ns    total=72672815
impl=rwmutex  scenario=burst      readers=8  p50=512ns  p95=1024ns  p99=1024ns   p999=2048ns    total=74364502
impl=rwmutex  scenario=highwrite  readers=8  p50=128ns  p95=512ns   p99=16384ns  p999=131072ns  total=49051195
impl=snapshot scenario=none       readers=8  p50=64ns   p95=64ns    p99=128ns    p999=128ns     total=297074386
impl=snapshot scenario=steady     readers=8  p50=64ns   p95=128ns   p99=128ns    p999=256ns     total=281690753
impl=snapshot scenario=burst      readers=8  p50=64ns   p95=128ns   p99=128ns    p999=256ns     total=271557824
impl=snapshot scenario=highwrite  readers=8  p50=64ns   p95=128ns   p99=128ns    p999=256ns     total=218317810
```

Key finding: RWMutex p99 does NOT degrade at 100 writes/sec (steady/burst).
Under saturated continuous writes (highwrite), p99 spikes 16x (1024ns → 16384ns)
and p999 spikes 64x (2048ns → 131072ns). Atomic snapshot p99/p999 remain flat.

## CPU Profile Finding

From `go tool pprof -top benchmarks/results/rwmutex_cpu.prof`:

```
flat  flat%   sum%        cum   cum%
58.93s 93.84% 93.84%   sync/atomic.(*Int32).Add (inline)
 0.68s  1.08% 94.92%   sync.(*RWMutex).RUnlock (inline)
 0.21s  0.33% 97.77%   sync.(*RWMutex).RLock (inline)
```

93.84% of CPU time in the parallel RWMutex benchmark is spent in `sync/atomic.(*Int32).Add`
— this is the reader-count increment/decrement inside RLock/RUnlock bouncing a cache line
between 8 CPU cores. The application logic (map lookup) is noise.

## Single-Goroutine vs Parallel RWMutex

```
BenchmarkPropagation_RWMutex_SingleGoroutine-8: ~43ns/op (10 runs: 34–60ns, median ~43ns)
BenchmarkPropagation_RWMutex_NoWriter-8:        ~122ns/op (10 runs: 113–140ns, median ~122ns)
```

Ratio: **~2.8x**. This proves the 122ns cost is reader-reader cache-line contention
under parallelism, NOT reader-writer blocking. With one goroutine, the atomic reader-count
never bounces: ~43ns (RLock+RUnlock+10k map lookup). With 8 goroutines, each RLock/RUnlock
pair broadcasts the reader-count cache line to all cores: ~122ns.

Note: the single-goroutine benchmark is DCE-proof (10k map, rotating keys, sinkFlag assignment).
The 43ns includes the 10k map lookup (~20-25ns); pure RLock/RUnlock overhead is ~20ns alone.
