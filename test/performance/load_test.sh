#!/bin/bash
# Load test - gradually increase load to find limits

set -e

PROXY_URL="http://localhost:8080"
DURATION="30s"

echo "==================================="
echo "ConnX Load Test"
echo "==================================="
echo ""

# Check for hey tool
if ! command -v hey &> /dev/null; then
    echo "❌ 'hey' tool is required for load testing"
    echo "Install with: go install github.com/rakyll/hey@latest"
    exit 1
fi

# Test with increasing load
LOADS=(10 50 100 200 500 1000 2000 5000)

echo "Testing with increasing concurrent connections..."
echo ""

for LOAD in "${LOADS[@]}"; do
    echo "===================================   "
    echo "Testing with $LOAD concurrent connections"
    echo "==================================="

    # Calculate requests (10 req/conn)
    REQUESTS=$((LOAD * 10))

    # Run test
    hey -n $REQUESTS -c $LOAD -t 10 "$PROXY_URL" | grep -E "(Requests/sec|Average|Fastest|Slowest|Status code)"

    echo ""

    # Small delay between tests
    sleep 2
done

echo "==================================="
echo "Load Test Complete"
echo "==================================="
echo ""
echo "Results Summary:"
curl -s "$PROXY_URL/metrics" | grep -E "connx_requests_per_second"
echo ""

echo "💡 Check logs for any errors or warnings"
echo "💡 Look for the point where latency significantly increases"
echo "💡 That's your maximum sustainable load"
