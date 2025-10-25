#!/bin/bash

echo "📊 DEMO: Performance and Scaling Test"
echo "===================================="

# Base URL
BASE_URL="http://localhost"

echo "1. 🔄 Creating test URLs for load testing..."
SHORT_CODES=()
for i in {1..50}; do
    response=$(curl -s -X POST "$BASE_URL/api/shorten" \
        -H "Content-Type: application/json" \
        -d "{\"url\":\"https://example.com/test/$i?loadtest=true&timestamp=$(date +%s)\"}")
    
    short_code=$(echo "$response" | grep -o '"short_code":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$short_code" ]; then
        SHORT_CODES+=("$short_code")
        echo "✅ Created: $short_code"
    else
        echo "❌ Failed to create URL"
    fi
done

echo ""
echo "2. 🚀 Starting Load Test (Mixed Workload)..."
echo "   - 80% Redirects"
echo "   - 20% URL Shortening"
echo ""

get_random_short_code() {
    size=${#SHORT_CODES[@]}
    index=$((RANDOM % size))
    echo "${SHORT_CODES[$index]}"
}

echo "Testing mixed workload (5000 requests)..."
time {
    for i in {1..5000}; do
        if (( RANDOM % 100 < 80 )); then
            short_code=$(get_random_short_code)
            curl -s "$BASE_URL/$short_code" > /dev/null
        else
            curl -s -X POST "$BASE_URL/api/shorten" \
                -H "Content-Type: application/json" \
                -d "{\"url\":\"https://example.com/new/$(date +%s)/$RANDOM\"}" > /dev/null
        fi
    done
}

echo ""
echo "3. 📈 Testing High Redirect Volume..."
echo "Testing 5000 redirects..."
time {
    for i in {1..5000}; do
        short_code=$(get_random_short_code)
        curl -s "$BASE_URL/$short_code" > /dev/null
    done
}

echo ""
echo "4. 📝 Testing High Shortening Volume..."
echo "Testing 50 URL shortening requests..."
time {
    for i in {1..5000}; do
        curl -s -X POST "$BASE_URL/api/shorten" \
            -H "Content-Type: application/json" \
            -d "{\"url\":\"https://example.com/bulk/$i/$(date +%s)\"}" > /dev/null
    done
}

echo ""
echo "5. 📊 Checking Application Metrics..."
echo "Current performance metrics:"
if command -v jq &> /dev/null; then
    curl -s "$BASE_URL/metrics" | jq '.'
else
    curl -s "$BASE_URL/metrics"
fi

echo ""
echo "6. ✅ Health Check..."
curl -s "$BASE_URL/health" | grep -o '"status":"[^"]*"'

echo ""
echo "🎉 Load test completed!"
echo "📈 Check metrics above for performance data"