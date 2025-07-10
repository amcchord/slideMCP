#!/bin/bash

# Diagnostic script to understand the hanging issue

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"

echo "Diagnosing Report Tool Performance"
echo "=================================="
echo ""

# Enable verbose logging
export SLIDE_REPORTS_VERBOSE=true

# Test 1: Check basic connectivity
echo "1. Testing basic API connectivity..."
echo -n "   Fetching agents list... "
START=$(date +%s)
RESULT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_agents","arguments":{"operation":"list","limit":1}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>/dev/null | grep '^{' | tail -n 1)
END=$(date +%s)
echo "OK ($(($END - $START))s)"

# Test 2: Get first agent details
AGENT_ID=$(echo "$RESULT" | jq -r '.result.content[0].text' | jq -r '.data[0].agent_id')
echo "   First agent ID: $AGENT_ID"
echo ""

# Test 3: Test single day report with explicit verbose
echo "2. Testing daily report with verbose output..."
echo "   This should show progress messages:"
echo "   --------------------------------"

# Run with both stderr and stdout visible
echo "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"daily_backup_snapshot\",\"agent_id\":\"$AGENT_ID\",\"verbose\":true}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | \
while IFS= read -r line; do
    if [[ "$line" == "["* ]]; then
        # Progress message
        echo "   PROGRESS: $line"
    elif [[ "$line" == "{"* ]]; then
        # Save JSON result
        echo "$line" > /tmp/diagnose_result.json
    else
        # Other output
        echo "   LOG: $line"
    fi
done

echo ""
echo "3. Checking if we got a valid result..."
if [ -f /tmp/diagnose_result.json ]; then
    BACKUP_COUNT=$(cat /tmp/diagnose_result.json | jq -r '.result.content[0].text' | jq -r '.reports[0].backups.total' 2>/dev/null || echo "parse error")
    echo "   Result received. Backup count: $BACKUP_COUNT"
else
    echo "   ERROR: No result received"
fi

echo ""
echo "4. Testing what happens with multiple agents (3 days)..."
echo "   Running a 3-day report for single agent (should be faster):"

# Use a shorter date range
THREE_DAYS_AGO=$(date -v-3d '+%Y-%m-%d' 2>/dev/null || date -d "3 days ago" '+%Y-%m-%d')

echo "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"daily_backup_snapshot\",\"agent_id\":\"$AGENT_ID\",\"date\":\"$THREE_DAYS_AGO\",\"verbose\":true}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | \
while IFS= read -r line; do
    if [[ "$line" == "["* ]]; then
        echo "   PROGRESS: $line"
    fi
done

echo ""
echo "5. Recommendations:"
echo "   - If you see no PROGRESS messages above, verbose logging isn't working"
echo "   - If you see progress but it's slow, the API calls are the bottleneck"
echo "   - For 10 agents × 7 days = 70 combinations, expect several minutes"
echo ""
echo "   Quick workarounds:"
echo "   a) Test with single agent first:"
echo "      {\"operation\": \"weekly_backup_snapshot\", \"agent_id\": \"$AGENT_ID\", \"verbose\": true}"
echo "   b) Test with shorter timeframe (yesterday only):"
echo "      {\"operation\": \"daily_backup_snapshot\", \"date\": \"$(date -v-1d '+%Y-%m-%d' 2>/dev/null || date -d 'yesterday' '+%Y-%m-%d')\"}" 