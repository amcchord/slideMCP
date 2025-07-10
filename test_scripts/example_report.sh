#!/bin/bash

# Example of using the slide_reports tool

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"

# Using the agent ID we found earlier
AGENT_ID="a_mbepgxrb629h"

echo "Example: Daily Backup/Snapshot Report"
echo "====================================="
echo ""
echo "Generating report for agent ID: $AGENT_ID"
echo "Date: Today (default)"
echo "Format: Markdown"
echo ""

# Generate the report
REPORT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_reports\",\"arguments\":{\"operation\":\"daily_backup_snapshot\",\"agent_id\":\"$AGENT_ID\",\"format\":\"markdown\"}}}" | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

# Extract and display the report
echo "$REPORT" | jq '.result.content[0].text' -r

echo ""
echo "====================================="
echo "Example complete!"
echo ""
echo "Note: This tool provides pre-calculated statistics for:"
echo "- Backup counts and success rates"
echo "- Snapshot counts (active snapshots)"
echo "- Storage locations (local vs cloud)"
echo ""
echo "The tool can filter by:"
echo "- agent_id: Specific agent"
echo "- device_id: All agents on a device"
echo "- client_id: All agents for a client" 
echo "- date: Specific date (YYYY-MM-DD format)"
echo ""
echo "Output formats: json or markdown" 