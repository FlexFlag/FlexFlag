#!/bin/bash

echo "⚡ FlexFlag OPTIMIZED Stress Test - 1000 Evaluations"
echo "===================================================="
echo ""

# Warm up the cache first
echo "🔥 Warming up cache..."
for i in {1..5}; do
    curl -s -X POST http://localhost:8080/api/v1/evaluate/fast \
        -H "Content-Type: application/json" \
        -d '{"flag_key":"new-feature","user_id":"warmup"}' > /dev/null
done

echo "✅ Cache warmed up"
echo ""

# Create results file
results_file=$(mktemp)
trap "rm -f $results_file" EXIT

echo "🚀 Starting 1000 OPTIMIZED concurrent evaluations..."
start_time=$(python3 -c "import time; print(time.time())")

# Launch 1000 concurrent requests to FAST endpoint
for i in $(seq 1 1000); do
    {
        response=$(curl -s -X POST http://localhost:8080/api/v1/evaluate/fast \
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
        echo "📊 Launched $i optimized requests..."
        wait
    fi
done

echo "⏳ Waiting for all optimized requests to complete..."
wait

end_time=$(python3 -c "import time; print(time.time())")
total_duration=$(python3 -c "print(f'{$end_time - $start_time:.3f}')")

echo ""
echo "✅ All 1000 OPTIMIZED requests completed!"
echo "⏱️  Total test duration: ${total_duration}s"

# Calculate statistics
if [ -s "$results_file" ]; then
    echo ""
    echo "📊 OPTIMIZED Performance Analysis:"
    echo "=================================="
    
    python3 << EOF
import statistics

with open('$results_file', 'r') as f:
    times = [float(line.strip()) for line in f if line.strip() and float(line.strip()) > 0]

if times:
    print(f"📈 Successful evaluations: {len(times)}")
    print(f"⚡ Average evaluation time: {statistics.mean(times):.3f}ms")
    print(f"🚀 Median evaluation time: {statistics.median(times):.3f}ms")
    print(f"📉 Min evaluation time: {min(times):.3f}ms")
    print(f"📈 Max evaluation time: {max(times):.3f}ms")
    print(f"📊 95th percentile: {sorted(times)[int(len(times)*0.95)]:.3f}ms")
    print(f"🎯 99th percentile: {sorted(times)[int(len(times)*0.99)]:.3f}ms")
    
    total_time = $total_duration
    throughput = 1000 / total_time
    print(f"")
    print(f"🚀 Throughput: {throughput:.0f} evaluations/second")
    print(f"⚡ Total test time: {total_time:.3f}s")
    
    avg_time = statistics.mean(times)
    p99_time = sorted(times)[int(len(times)*0.99)]
    
    print(f"")
    print(f"🏆 OPTIMIZED Performance Verdict:")
    print(f"{'='*35}")
    
    if avg_time < 1:
        print("🥇 OUTSTANDING: Average < 1ms!")
    elif avg_time < 5:
        print("✅ EXCELLENT: Average < 5ms")
    elif avg_time < 10:
        print("✅ GOOD: Average < 10ms")
    else:
        print("⚠️  NEEDS OPTIMIZATION: Average > 10ms")
        
    if p99_time < 5:
        print("🥇 OUTSTANDING: 99th percentile < 5ms!")
    elif p99_time < 10:
        print("✅ EXCELLENT: 99th percentile < 10ms")
    elif p99_time < 20:
        print("✅ GOOD: 99th percentile < 20ms")
    else:
        print("⚠️  NEEDS OPTIMIZATION: 99th percentile > 20ms")
        
    if throughput > 500:
        print("🥇 OUTSTANDING THROUGHPUT: > 500 req/sec!")
    elif throughput > 200:
        print("✅ HIGH THROUGHPUT: > 200 req/sec")
    elif throughput > 100:
        print("✅ GOOD THROUGHPUT: > 100 req/sec")
    else:
        print("⚠️  LOW THROUGHPUT: < 100 req/sec")
        
    print(f"")
    print(f"🎯 IMPROVEMENT vs Original:")
    print(f"Original average: 22.691ms")
    print(f"Optimized average: {avg_time:.3f}ms")
    improvement = 22.691 / avg_time
    print(f"🚀 Performance improvement: {improvement:.1f}x FASTER!")
    
else:
    print("❌ No valid evaluation times recorded")
EOF
fi

echo ""
echo "⚡ OPTIMIZED Stress Test Complete!"
echo "=================================="