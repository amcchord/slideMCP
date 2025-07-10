#!/bin/bash

# Performance test for reports tool - shows why operations can be slow

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"

echo "Performance Analysis for Slide Reports Tool"
echo "=========================================="
echo ""

# Enable verbose mode
export SLIDE_REPORTS_VERBOSE=true

# Get system stats
echo "1. Checking system scale..."
AGENTS_RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_agents","arguments":{"operation":"list","limit":1}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

TOTAL_AGENTS=$(echo "$AGENTS_RESPONSE" | jq -r '.result.content[0].text' | jq -r '.pagination.total')
echo "   Total agents in system: $TOTAL_AGENTS"

# Get a sample agent
FIRST_AGENT=$(echo "$AGENTS_RESPONSE" | jq -r '.result.content[0].text' | jq -r '.data[0].agent_id')
FIRST_AGENT_NAME=$(echo "$AGENTS_RESPONSE" | jq -r '.result.content[0].text' | jq -r '.data[0].display_name')
echo "   Sample agent: $FIRST_AGENT_NAME (ID: $FIRST_AGENT)"
echo ""

# Test single day report
echo "2. Testing single day report for one agent..."
echo "   Expected API calls: ~3 (agent details, backups, snapshots)"
START_TIME=$(date +%s)

DAILY_RESULT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"daily_backup_snapshot\",\"agent_id\":\"$FIRST_AGENT\",\"format\":\"json\"}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | tee /tmp/daily_single.log | grep '^{' | tail -n 1)

END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
echo "   Completed in $ELAPSED seconds"
echo ""

# Calculate estimates
echo "3. Performance Estimates:"
echo "   ----------------------"
echo "   Single agent/day: ~$ELAPSED seconds"
echo ""
echo "   Weekly report estimates:"
echo "   - Single agent: 7 days × $ELAPSED sec = ~$((7 * ELAPSED)) seconds"
echo "   - All agents: 7 days × $TOTAL_AGENTS agents × $ELAPSED sec = ~$((7 * TOTAL_AGENTS * ELAPSED)) seconds ($(((7 * TOTAL_AGENTS * ELAPSED) / 60)) minutes)"
echo ""
echo "   Monthly report estimates:"
echo "   - Single agent: 30 days × $ELAPSED sec = ~$((30 * ELAPSED)) seconds"
echo "   - All agents: 30 days × $TOTAL_AGENTS agents × $ELAPSED sec = ~$((30 * TOTAL_AGENTS * ELAPSED)) seconds ($(((30 * TOTAL_AGENTS * ELAPSED) / 60)) minutes)"
echo ""

# Show actual work being done
echo "4. What happens during report generation:"
echo "   --------------------------------------"
echo "   For each day in the report period:"
echo "     - Fetch all agents (if no filter specified)"
echo "     - For each agent:"
echo "       - Get agent details"
echo "       - Fetch all backups for that date"
echo "       - Fetch all snapshots (active + deleted)"
echo "       - Calculate statistics"
echo ""

# Recommendations
echo "5. Recommendations for faster reports:"
echo "   -----------------------------------"
echo "   - Use filters (agent_id, device_id, client_id) to reduce scope"
echo "   - Start with daily reports before weekly/monthly"
echo "   - Consider implementing caching for frequently accessed data"
echo "   - For production use, consider background processing with progress updates"
echo ""

# Show log contents
echo "6. Sample verbose output from daily report:"
echo "   ----------------------------------------"
grep "^\[" /tmp/daily_single.log | head -n 20 || echo "   (No verbose output captured)" 