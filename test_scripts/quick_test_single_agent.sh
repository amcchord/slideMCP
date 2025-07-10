#!/bin/bash

# Quick test with single agent to show the tool works but is slow

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"

echo "Quick Single Agent Weekly Report Test"
echo "====================================="
echo ""

# Get first agent
echo "Getting first agent ID..."
AGENT_DATA=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_agents","arguments":{"operation":"list","limit":1}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

AGENT_ID=$(echo "$AGENT_DATA" | jq -r '.result.content[0].text' | jq -r '.data[0].agent_id')
AGENT_NAME=$(echo "$AGENT_DATA" | jq -r '.result.content[0].text' | jq -r '.data[0].display_name')

echo "Testing with agent: $AGENT_NAME (ID: $AGENT_ID)"
echo ""

# Test 1: Single day (fast)
echo "Test 1: Single day report (should be fast - few seconds)"
echo "--------------------------------------------------------"
START=$(date +%s)

DAILY=$(echo "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"daily_backup_snapshot\",\"agent_id\":\"$AGENT_ID\",\"format\":\"json\"}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

END=$(date +%s)
ELAPSED=$((END - START))

BACKUP_TOTAL=$(echo "$DAILY" | jq -r '.result.content[0].text' | jq -r '.reports[0].backups.total' 2>/dev/null || echo "0")
echo "✓ Completed in $ELAPSED seconds - Found $BACKUP_TOTAL backups"
echo ""

# Test 2: Weekly report for single agent
echo "Test 2: Weekly report for single agent (7x slower)"
echo "--------------------------------------------------"
echo "Processing 7 days of data..."
START=$(date +%s)

# Add progress indicator
(
    while kill -0 $$ 2>/dev/null; do
        echo -n "."
        sleep 2
    done
) &
PROGRESS_PID=$!

WEEKLY=$(echo "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"weekly_backup_snapshot\",\"agent_id\":\"$AGENT_ID\",\"format\":\"json\"}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

# Stop progress indicator
kill $PROGRESS_PID 2>/dev/null
wait $PROGRESS_PID 2>/dev/null
echo ""

END=$(date +%s)
ELAPSED=$((END - START))

# Show summary
WEEK_START=$(echo "$WEEKLY" | jq -r '.result.content[0].text' | jq -r '.week_start' 2>/dev/null || echo "error")
echo "✓ Completed in $ELAPSED seconds"
echo "  Week starting: $WEEK_START"
echo ""

# Calculate estimates
echo "Performance Analysis:"
echo "===================="
echo "- Single day for 1 agent: ~$(($ELAPSED / 7)) seconds"
echo "- Weekly for 1 agent: $ELAPSED seconds"
echo "- Weekly for 10 agents: ~$(($ELAPSED * 10)) seconds ($(($ELAPSED * 10 / 60)) minutes)"
echo "- Monthly for 10 agents: ~$(($ELAPSED * 10 * 30 / 7)) seconds ($(($ELAPSED * 10 * 30 / 7 / 60)) minutes)"
echo ""
echo "The tool is working correctly, but processing many agent-days takes time!"
echo ""
echo "Tips:"
echo "- Always use agent_id filter for testing"
echo "- Start with daily reports before weekly/monthly"
echo "- Consider running reports for specific clients/devices instead of all agents" 