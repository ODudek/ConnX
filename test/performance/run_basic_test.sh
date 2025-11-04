#!/bin/bash
# Basic performance test using wrk or hey

set -e

PROXY_URL="http://localhost:8080"
DURATION="30s"
CONNECTIONS=100
THREADS=4

echo "==================================="
echo "ConnX Basic Performance Test"
echo "==================================="
echo "Target: $PROXY_URL"
echo "Duration: $DURATION"
echo "Connections: $CONNECTIONS"
echo "Threads: $THREADS"
echo "==================================="
echo ""

# Check if proxy is running
if ! curl -s -f "$PROXY_URL/metrics" > /dev/null; then
    echo "❌ Error: ConnX proxy is not running on $PROXY_URL"
    echo "Start it with: ./bin/proxy -config=configs/config.yaml"
    exit 1
fi

echo "✅ Proxy is running"
echo ""

# Function to test with different tools
test_with_wrk() {
    if command -v wrk &> /dev/null; then
        echo "📊 Running wrk benchmark..."
        wrk -t$THREADS -c$CONNECTIONS -d$DURATION --latency "$PROXY_URL"
        echo ""
    else
        echo "⚠️  wrk not installed. Install with: brew install wrk (macOS) or apt install wrk (Ubuntu)"
    fi
}

test_with_hey() {
    if command -v hey &> /dev/null; then
        echo "📊 Running hey benchmark..."
        hey -n 100000 -c $CONNECTIONS -t 10 "$PROXY_URL"
        echo ""
    else
        echo "⚠️  hey not installed. Install with: go install github.com/rakyll/hey@latest"
    fi
}

test_with_ab() {
    if command -v ab &> /dev/null; then
        echo "📊 Running Apache Bench..."
        ab -n 10000 -c $CONNECTIONS -g /dev/null "$PROXY_URL/"
        echo ""
    else
        echo "⚠️  ab not installed. Install with: apt install apache2-utils"
    fi
}

# Run available tools
test_with_wrk
test_with_hey
test_with_ab

# Show metrics
echo "==================================="
echo "Current ConnX Metrics:"
echo "==================================="
curl -s "$PROXY_URL/metrics" | grep -E "(connx_requests_total|connx_requests_per_second|connx_response_time|connx_errors_total)" || echo "Metrics not available"

echo ""
echo "✅ Test completed!"
echo ""
echo "Next steps:"
echo "  - Run load test: ./test/performance/load_test.sh"
echo "  - Run stress test: ./test/performance/stress_test.sh"
echo "  - View Prometheus metrics: curl http://localhost:8080/metrics"
