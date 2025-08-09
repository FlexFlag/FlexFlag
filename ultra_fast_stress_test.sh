#!/bin/bash

echo "⚡ FlexFlag ULTRA-FAST Stress Test - 1000 Evaluations"
echo "===================================================="
echo ""

# Warm up the cache first
echo "🔥 Warming up ultra-fast cache..."
for i in {1..5}; do
    curl -s -X POST http://localhost:8080/api/v1/evaluate/ultra \
        -H "Content-Type: application/json" \
        -d '{"flag_key":"new-feature","user_id":"warmup"}' > /dev/null
done

echo "✅ Ultra-fast cache warmed up"
echo ""

# Check initial stats
echo "📊 Initial Ultra-Fast Handler Stats:"
curl -s -X GET http://localhost:8080/api/v1/evaluate/ultra/stats
echo ""
echo ""

# Create results file
results_file=$(mktemp)
trap "rm -f $results_file" EXIT

echo "🚀 Starting 1000 ULTRA-FAST concurrent evaluations..."
start_time=$(python3 -c "import time; print(time.time())")

# Launch 1000 concurrent requests to ULTRA endpoint
for i in $(seq 1 1000); do
    {
        response=$(curl -s -X POST http://localhost:8080/api/v1/evaluate/ultra \
            -H "Content-Type: application/json" \
            -d "{
                \"flag_key\": \"new-feature\",
                \"user_id\": \"user$i\",
                \"attributes\": {
                    \"plan\": \"premium\",
                    \"region\": \"us-west-$((i % 3 + 1))\"
                }
            }" 2>/dev/null)
        
        eval_time=$(echo "$response" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(f'{data.get(\"evaluation_time_ms\", 0):.3f}')
except:
    print('0')
        " 2>/dev/null || echo "0")
        
        echo "$eval_time" >> "$results_file"
    } &
    
    # Batch control
    if (( i % 100 == 0 )); then
        echo "📊 Launched $i ultra-fast requests..."
        wait
    fi
done

echo "⏳ Waiting for all ultra-fast requests to complete..."
wait

end_time=$(python3 -c "import time; print(time.time())")
total_duration=$(python3 -c "print(f'{$end_time - $start_time:.3f}')")

echo ""
echo "✅ All 1000 ULTRA-FAST requests completed!"
echo "⏱️  Total test duration: ${total_duration}s"

# Check final stats
echo ""
echo "📊 Final Ultra-Fast Handler Stats:"
curl -s -X GET http://localhost:8080/api/v1/evaluate/ultra/stats
echo ""

# Calculate statistics
if [ -s "$results_file" ]; then
    echo ""
    echo "📊 ULTRA-FAST Performance Analysis:"
    echo "=================================="
    
    python3 << EOF
import statistics

with open('$results_file', 'r') as f:
    times = [float(line.strip()) for line in f if line.strip() and float(line.strip()) > 0]

if times:
    print(f"📈 Successful evaluations: {len(times)}")
    print(f"⚡ Average evaluation time: {statistics.mean(times):.6f}ms")
    print(f"🚀 Median evaluation time: {statistics.median(times):.6f}ms")
    print(f"📉 Min evaluation time: {min(times):.6f}ms")
    print(f"📈 Max evaluation time: {max(times):.6f}ms")
    print(f"📊 95th percentile: {sorted(times)[int(len(times)*0.95)]:.6f}ms")
    print(f"🎯 99th percentile: {sorted(times)[int(len(times)*0.99)]:.6f}ms")
    
    total_time = $total_duration
    throughput = 1000 / total_time
    print(f"")
    print(f"🚀 Throughput: {throughput:.0f} evaluations/second")
    print(f"⚡ Total test time: {total_time:.3f}s")
    
    avg_time = statistics.mean(times)
    p95_time = sorted(times)[int(len(times)*0.95)]
    p99_time = sorted(times)[int(len(times)*0.99)]
    
    print(f"")
    print(f"🏆 ULTRA-FAST Performance Verdict:")
    print(f"{'='*35}")
    
    if avg_time < 0.1:
        print("🥇 INCREDIBLE: Average < 0.1ms! (Ultra-sub-millisecond)")
    elif avg_time < 0.5:
        print("🥇 OUTSTANDING: Average < 0.5ms!")
    elif avg_time < 1:
        print("🥇 EXCELLENT: Average < 1ms!")
    elif avg_time < 5:
        print("✅ GOOD: Average < 5ms")
    else:
        print("⚠️  NEEDS OPTIMIZATION: Average > 5ms")
        
    if p95_time < 0.1:
        print("🥇 INCREDIBLE: 95th percentile < 0.1ms!")
    elif p95_time < 0.5:
        print("🥇 OUTSTANDING: 95th percentile < 0.5ms!")
    elif p95_time < 1:
        print("🥇 EXCELLENT: 95th percentile < 1ms!")
    elif p95_time < 5:
        print("✅ GOOD: 95th percentile < 5ms")
    else:
        print("⚠️  SLOW 95th PERCENTILE: > 5ms")
        
    if p99_time < 0.1:
        print("🥇 INCREDIBLE: 99th percentile < 0.1ms!")
    elif p99_time < 0.5:
        print("🥇 OUTSTANDING: 99th percentile < 0.5ms!")
    elif p99_time < 1:
        print("🥇 EXCELLENT: 99th percentile < 1ms!")
    elif p99_time < 5:
        print("✅ GOOD: 99th percentile < 5ms")
    else:
        print("⚠️  SLOW 99th PERCENTILE: > 5ms")
        
    if throughput > 1000:
        print("🥇 INCREDIBLE THROUGHPUT: > 1000 req/sec!")
    elif throughput > 500:
        print("🥇 OUTSTANDING THROUGHPUT: > 500 req/sec!")
    elif throughput > 200:
        print("✅ HIGH THROUGHPUT: > 200 req/sec")
    elif throughput > 100:
        print("✅ GOOD THROUGHPUT: > 100 req/sec")
    else:
        print("⚠️  LOW THROUGHPUT: < 100 req/sec")
        
    print(f"")
    print(f"🎯 IMPROVEMENT vs Previous Versions:")
    print(f"Original average: 22.691ms")
    print(f"Optimized average: 0.082ms")  
    print(f"Ultra-fast average: {avg_time:.6f}ms")
    improvement_vs_original = 22.691 / avg_time if avg_time > 0 else float('inf')
    improvement_vs_optimized = 0.082 / avg_time if avg_time > 0 else float('inf')
    print(f"🚀 vs Original: {improvement_vs_original:.0f}x FASTER!")
    print(f"⚡ vs Optimized: {improvement_vs_optimized:.1f}x FASTER!")
    
    print(f"")
    print(f"🎯 95th Percentile Analysis:")
    print(f"Previous 95th percentile: 33ms")
    print(f"Ultra-fast 95th percentile: {p95_time:.6f}ms")
    p95_improvement = 33.0 / p95_time if p95_time > 0 else float('inf')
    print(f"🚀 95th percentile improvement: {p95_improvement:.0f}x FASTER!")
    
    if p95_time < 1:
        print("✅ MISSION ACCOMPLISHED: 95th percentile now < 1ms!")
    else:
        print("⚠️  95th percentile still needs work")
    
else:
    print("❌ No valid evaluation times recorded")
EOF
fi

echo ""
echo "⚡ ULTRA-FAST Stress Test Complete!"
echo "================================="