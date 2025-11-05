#!/bin/bash
# Compare performance of different load balancing algorithms

set -e

PROXY_URL="http://localhost:8080"
CONFIG_PATH="configs/config.yaml"
REQUESTS=50000
CONNECTIONS=200

echo "==================================="
echo "Load Balancing Algorithm Comparison"
echo "==================================="
echo ""

if ! command -v hey &> /dev/null; then
    echo "❌ 'hey' tool is required"
    echo "Install with: go install github.com/rakyll/hey@latest"
    exit 1
fi

# Backup original config
cp "$CONFIG_PATH" "${CONFIG_PATH}.backup"

# Function to update algorithm and test
test_algorithm() {
    local algo=$1
    echo ""
    echo "==================================="
    echo "Testing: $algo"
    echo "==================================="

    # Update config (hot reload will pick it up)
    sed -i.tmp "s/algorithm: .*/algorithm: \"$algo\"/" "$CONFIG_PATH"
    rm -f "${CONFIG_PATH}.tmp"

    # Wait for hot reload
    echo "Waiting for hot reload..."
    sleep 3

    # Warm up
    echo "Warming up..."
    hey -n 1000 -c 50 "$PROXY_URL" > /dev/null 2>&1
    sleep 2

    # Run test
    echo "Running benchmark..."
    hey -n $REQUESTS -c $CONNECTIONS -t 10 "$PROXY_URL" > "/tmp/connx_bench_${algo}.txt"

    # Show results
    echo "Results:"
    grep -E "(Requests/sec|Average|Fastest|Slowest|P50|P95|P99)" "/tmp/connx_bench_${algo}.txt"
}

# Test each algorithm
test_algorithm "round-robin"
test_algorithm "least-connections"
test_algorithm "weighted-round-robin"

# Restore original config
mv "${CONFIG_PATH}.backup" "$CONFIG_PATH"
echo ""
echo "Config restored, waiting for hot reload..."
sleep 3

echo ""
echo "==================================="
echo "Comparison Summary"
echo "==================================="
echo ""

for algo in "round-robin" "least-connections" "weighted-round-robin"; do
    echo "--- $algo ---"
    grep "Requests/sec" "/tmp/connx_bench_${algo}.txt"
    grep "Average" "/tmp/connx_bench_${algo}.txt" | head -1
done

echo ""
echo "📊 Full results saved in /tmp/connx_bench_*.txt"
echo ""
echo "💡 Analysis tips:"
echo "   - round-robin: Baseline, simplest"
echo "   - least-connections: Better for varying response times"
echo "   - weighted-round-robin: Best when backends have different capacities"
