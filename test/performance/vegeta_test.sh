#!/bin/bash
# Vegeta-based benchmark (requires vegeta tool)

set -e

PROXY_URL="http://localhost:8080"
RATE=5000  # requests per second
DURATION="30s"

echo "==================================="
echo "Vegeta Constant Rate Test"
echo "==================================="
echo "Rate: $RATE req/sec"
echo "Duration: $DURATION"
echo "==================================="
echo ""

if ! command -v vegeta &> /dev/null; then
    echo "❌ 'vegeta' tool is required"
    echo "Install with: go install github.com/tsenart/vegeta@latest"
    exit 1
fi

# Generate target
echo "GET $PROXY_URL" | vegeta attack -rate=$RATE -duration=$DURATION | vegeta report -type=text

echo ""
echo "Generating plots..."

# Generate histogram
echo "GET $PROXY_URL" | vegeta attack -rate=$RATE -duration=$DURATION | vegeta plot > /tmp/connx_vegeta.html

echo ""
echo "✅ Results saved:"
echo "   - Text report: shown above"
echo "   - HTML plot: /tmp/connx_vegeta.html"
echo ""
echo "Open the HTML plot:"
echo "   open /tmp/connx_vegeta.html  # macOS"
echo "   xdg-open /tmp/connx_vegeta.html  # Linux"
