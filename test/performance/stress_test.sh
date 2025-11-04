#!/bin/bash
# Stress test - push system beyond normal limits

set -e

PROXY_URL="http://localhost:8080"
DURATION="60s"
EXTREME_CONNECTIONS=10000
THREADS=8

echo "==================================="
echo "ConnX Stress Test"
echo "==================================="
echo "⚠️  WARNING: This will push the system hard!"
echo "Duration: $DURATION"
echo "Connections: $EXTREME_CONNECTIONS"
echo "==================================="
echo ""

read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

echo "Starting stress test..."
echo ""

# Get baseline metrics
echo "Baseline metrics:"
curl -s "$PROXY_URL/metrics" | grep -E "(connx_requests_total|connx_active_connections)"
echo ""
echo "Starting in 3 seconds..."
sleep 3

# Run stress test
if command -v wrk &> /dev/null; then
    wrk -t$THREADS -c$EXTREME_CONNECTIONS -d$DURATION --latency "$PROXY_URL"
elif command -v hey &> /dev/null; then
    hey -n 1000000 -c $EXTREME_CONNECTIONS -t 30 "$PROXY_URL"
else
    echo "❌ Neither wrk nor hey found. Install one of them."
    exit 1
fi

echo ""
echo "==================================="
echo "Stress Test Complete"
echo "==================================="
echo ""
echo "Post-test metrics:"
curl -s "$PROXY_URL/metrics" | grep -E "(connx_requests_total|connx_errors_total|connx_active_connections)"
echo ""

echo "💡 Check system resources:"
echo "   - top or htop for CPU/Memory"
echo "   - dmesg for kernel messages"
echo "   - Check for 'too many open files' errors"
echo ""
echo "💡 If you see high error rates, reduce connections or increase limits:"
echo "   - ulimit -n 65536  # Increase file descriptors"
echo "   - Tune kernel parameters (net.ipv4.ip_local_port_range, etc.)"
