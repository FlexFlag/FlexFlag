#!/bin/bash

echo "🔥 FlexFlag Stress Test - 1000 Concurrent Evaluations"
echo "======================================================"

# Create a temporary file for results
results_file=$(mktemp)
trap "rm -f $results_file" EXIT

echo "🚀 Starting 1000 concurrent flag evaluations..."
start_time=$(python3 -c "import time; print(time.time())")

# Launch 1000 concurrent requests
for i in $(seq 1 1000); do
    {
        response=$(curl -s -w "%{time_total}" -X POST http://localhost:8080/api/v1/evaluate \
            -H "Content-Type: application/json" \
            -d "{
                \"flag_key\": \"new-feature\",
                \"user_id\": \"user$i\",
                \"attributes\": {
                    \"email\": \"user$i@example.com\",
                    \"plan\": \"premium\",
                    \"region\": \"us-west-$((i % 3 + 1))\"
                }
            }" 2>/dev/null)
        
        # Extract evaluation time from JSON response
        eval_time=$(echo "$response" | python3 -c "
import sys, json, re
content = sys.stdin.read()
# Split curl timing from JSON
parts = content.rsplit('}', 1)
if len(parts) == 2:
    json_part = parts[0] + '}'
    curl_time = parts[1]
    try:
        data = json.loads(json_part)
        print(f'{data.get(\"evaluation_time_ms\", 0):.3f}')
    except:
        print('0')
else:
    print('0')
        " 2>/dev/null || echo "0")
        
        echo "$eval_time" >> "$results_file"
    } &
    
    # Limit concurrent processes to avoid overwhelming the system
    if (( i % 100 == 0 )); then
        echo "📊 Launched $i requests..."
        wait  # Wait for this batch to complete
    fi
done

echo "⏳ Waiting for all requests to complete..."
wait

end_time=$(python3 -c "import time; print(time.time())")
total_duration=$(python3 -c "print(f'{$end_time - $start_time:.3f}')")

echo ""
echo "✅ All 1000 requests completed!"
echo "⏱️  Total test duration: ${total_duration}s"

# Calculate statistics
if [ -s "$results_file" ]; then
    echo ""
    echo "📊 Performance Analysis:"
    echo "========================"
    
    python3 << EOF
import statistics

# Read evaluation times
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
    
    # Calculate throughput
    total_time = $total_duration
    throughput = 1000 / total_time
    print(f"")
    print(f"🚀 Throughput: {throughput:.0f} evaluations/second")
    print(f"⚡ Total test time: {total_time:.3f}s")
    
    # Performance verdict
    avg_time = statistics.mean(times)
    p99_time = sorted(times)[int(len(times)*0.99)]
    
    print(f"")
    print(f"🏆 Performance Verdict:")
    print(f"{'='*25}")
    if avg_time < 5:
        print("✅ EXCELLENT: Average < 5ms")
    elif avg_time < 10:
        print("✅ GOOD: Average < 10ms")
    else:
        print("⚠️  NEEDS OPTIMIZATION: Average > 10ms")
        
    if p99_time < 10:
        print("✅ EXCELLENT: 99th percentile < 10ms")
    elif p99_time < 20:
        print("✅ GOOD: 99th percentile < 20ms")
    else:
        print("⚠️  NEEDS OPTIMIZATION: 99th percentile > 20ms")
        
    if throughput > 100:
        print("✅ HIGH THROUGHPUT: > 100 req/sec")
    elif throughput > 50:
        print("✅ GOOD THROUGHPUT: > 50 req/sec")
    else:
        print("⚠️  LOW THROUGHPUT: < 50 req/sec")
else:
    print("❌ No valid evaluation times recorded")
EOF

else
    echo "❌ No results recorded - check server status"
fi

echo ""
echo "🎯 Stress Test Complete!"
echo "========================"