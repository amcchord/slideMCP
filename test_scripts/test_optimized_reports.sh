#!/bin/bash

# Test script to compare original vs optimized report generation performance

API_KEY="$1"

if [ -z "$API_KEY" ]; then
    echo "Usage: $0 <API_KEY>"
    echo "Please provide your API key as the first argument"
    exit 1
fi

echo "Performance Comparison: Original vs Optimized Report Generation"
echo "=============================================================="
echo ""

# First, get device ID for testing
echo "Finding devices..."
DEVICE_INFO=$(echo '{"jsonrpc":"2.0","id":0,"method":"tools/call","params":{"name":"slide_devices","arguments":{"operation":"list","limit":2}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>/dev/null | grep '^{' | tail -n 1 | \
jq -r '.result.content[0].text' | jq -r '.data[0] | {device_id, hostname}')

DEVICE_ID=$(echo "$DEVICE_INFO" | jq -r '.device_id')
DEVICE_NAME=$(echo "$DEVICE_INFO" | jq -r '.hostname')

echo "Testing with device: $DEVICE_NAME (ID: $DEVICE_ID)"
echo ""

# Test 1: Original Weekly Report
echo "Test 1: Original Weekly Report Implementation"
echo "----------------------------------------"
START_TIME=$(date +%s)

echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_reports","arguments":{"operation":"weekly_backup_snapshot","device_id":"'$DEVICE_ID'","format":"json"}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep -E "(Fetching|Processing|Completed)" | head -20

END_TIME=$(date +%s)
ORIGINAL_TIME=$((END_TIME - START_TIME))
echo "Original implementation took: ${ORIGINAL_TIME} seconds"
echo ""

# Test 2: Optimized Weekly Report (would need to be implemented in the actual tool)
echo "Test 2: Proposed Optimizations"
echo "----------------------------------------"
echo "The optimized implementation would include:"
echo "1. Prefetching all backup data for the week in one pass per agent"
echo "2. Caching results to avoid redundant API calls"
echo "3. Higher parallelism (20 concurrent agents vs 10)"
echo "4. Batch processing of date ranges"
echo ""
echo "Expected performance improvements:"
echo "- 50-70% reduction in API calls"
echo "- 3-5x faster for weekly/monthly reports"
echo "- Much better scalability with more agents"
echo ""

# Show API call analysis
echo "API Call Analysis"
echo "----------------------------------------"
echo "For a weekly report with 4 agents:"
echo ""
echo "Original approach:"
echo "- Per agent per day: ~10-20 API calls (pagination)"
echo "- Total: 4 agents × 7 days × 15 avg calls = ~420 API calls"
echo ""
echo "Optimized approach:"
echo "- Per agent: ~20-30 API calls (fetching full week)"
echo "- Total: 4 agents × 25 avg calls = ~100 API calls"
echo "- Reduction: ~76% fewer API calls"
echo ""

# Additional optimization suggestions
echo "Additional Optimization Strategies"
echo "----------------------------------------"
echo "1. Implement server-side date filtering in the API"
echo "2. Add bulk data export endpoints for reporting"
echo "3. Implement report caching with TTL"
echo "4. Use database views or materialized views for common queries"
echo "5. Add async report generation with progress tracking"
echo "" 