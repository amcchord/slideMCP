#!/bin/bash

# Test script for the weekly and monthly report operations

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"

# Enable verbose logging
export SLIDE_REPORTS_VERBOSE=true

echo "Testing slide_reports tool with weekly and monthly operations..."
echo "Note: Verbose mode enabled - you'll see progress messages"
echo ""

# First, let's get a count of agents to understand the scale
echo "Checking number of agents in the system..."
AGENT_COUNT=$(echo '{"jsonrpc":"2.0","id":0,"method":"tools/call","params":{"name":"slide_agents","arguments":{"operation":"list","limit":1}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1 | \
jq -r '.result.content[0].text' | jq -r '.pagination.total')
echo "Found $AGENT_COUNT total agents"
echo ""

# Test 1: Weekly report in JSON format
echo "Test 1: Weekly backup/snapshot report (JSON format)"
echo "==================================================="
echo "This will fetch data for 7 days × $AGENT_COUNT agents = ~$(($AGENT_COUNT * 7)) agent-day combinations"
echo "Starting at $(date)..."
WEEKLY_JSON=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_reports","arguments":{"operation":"weekly_backup_snapshot","format":"json","verbose":true}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | tee /tmp/weekly_output.log | grep '^{' | tail -n 1)

echo "Completed at $(date)"
echo "$WEEKLY_JSON" | jq '.result.content[0].text' -r | jq '{week_start, week_end, metadata: ._metadata}'

echo ""
echo "Test 2: Weekly backup/snapshot report (Markdown format) - Skipping to save time"
echo "==============================================================================="
echo "(Would process same 7 days × $AGENT_COUNT agents)"

echo ""
echo "Test 3: Testing with a single agent to show performance"
echo "======================================================="
# Get first agent ID
FIRST_AGENT=$(echo '{"jsonrpc":"2.0","id":0,"method":"tools/call","params":{"name":"slide_agents","arguments":{"operation":"list","limit":1}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1 | \
jq -r '.result.content[0].text' | jq -r '.data[0].agent_id')

echo "Running weekly report for just agent: $FIRST_AGENT"
echo "This will fetch data for 7 days × 1 agent = 7 agent-day combinations"
echo "Starting at $(date)..."

SINGLE_AGENT_WEEKLY=$(echo "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"weekly_backup_snapshot\",\"agent_id\":\"$FIRST_AGENT\",\"format\":\"json\",\"verbose\":true}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | tee /tmp/single_agent_weekly.log | grep '^{' | tail -n 1)

echo "Completed at $(date)"
echo "$SINGLE_AGENT_WEEKLY" | jq '.result.content[0].text' -r | jq '{week_start, week_end, reports_count: ".daily_reports | length"}'

echo ""
echo "Test 4: Monthly report for single agent (with calendar)"
echo "======================================================"
CURRENT_MONTH_DAYS=$(date +%d)
echo "Running monthly report for agent: $FIRST_AGENT"
echo "This will fetch data for ~$CURRENT_MONTH_DAYS days × 1 agent = $CURRENT_MONTH_DAYS agent-day combinations"
echo "Starting at $(date)..."

MONTHLY_MD=$(echo "{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"monthly_backup_snapshot\",\"agent_id\":\"$FIRST_AGENT\",\"format\":\"markdown\",\"verbose\":true}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | tee /tmp/monthly_output.log | grep '^{' | tail -n 1)

# Extract and show the calendar portion
echo "Completed at $(date)"
echo "$MONTHLY_MD" | jq '.result.content[0].text' -r | sed -n '/## Calendar View/,/## Daily Details/p' | head -n -1

echo ""
echo "Test complete!"
echo ""
echo "Performance Summary:"
echo "- Total agents in system: $AGENT_COUNT"
echo "- Weekly report for all agents processes: 7 days × $AGENT_COUNT agents = $(($AGENT_COUNT * 7)) data points"
echo "- Monthly report for all agents would process: ~30 days × $AGENT_COUNT agents = $(($AGENT_COUNT * 30)) data points"
echo ""
echo "To see detailed progress logs, check:"
echo "- /tmp/weekly_output.log"
echo "- /tmp/single_agent_weekly.log" 
echo "- /tmp/monthly_output.log"
echo ""
echo "Tips for faster testing:"
echo "1. Use agent_id, device_id, or client_id filters to reduce scope"
echo "2. Test with shorter date ranges first"
echo "3. Enable verbose mode to see progress" 