#!/bin/bash

# Demo script to show verbose progress logging

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"
AGENT_ID="a_mbepgxrb629h"

echo "Slide Reports Tool - Verbose Progress Demo"
echo "=========================================="
echo ""
echo "This demonstrates the verbose logging feature that shows progress"
echo "during report generation. The tool is not hanging - it's working!"
echo ""
echo "Starting daily report with verbose=true..."
echo ""

# Run with verbose flag to show progress
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_reports","arguments":{"operation":"daily_backup_snapshot","agent_id":"'$AGENT_ID'","verbose":true}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | \
while IFS= read -r line; do
    if [[ "$line" == *"[Backup Stats]"* ]] || [[ "$line" == *"[Agent"* ]]; then
        echo ">>> PROGRESS: $line"
    elif [[ "$line" == "{"*"jsonrpc"* ]]; then
        echo ">>> COMPLETE: Received report data"
        # Extract summary
        echo "$line" | jq -r '.result.content[0].text' 2>/dev/null | \
        jq '.reports[0] | {
            date: .date,
            agent: .agent_name,
            backups: {
                total: .backups.total,
                success_rate: .backups.success_rate
            },
            snapshots: {
                total: .snapshots.total,
                active: .snapshots.active
            }
        }' 2>/dev/null || echo "Could not parse result"
    elif [[ "$line" == *"Slide MCP Server starting"* ]]; then
        echo ">>> Server started"
    elif [[ "$line" == *"Exiting after"* ]]; then
        echo ">>> Server exiting normally"
    fi
done

echo ""
echo "As you can see, the verbose flag shows the progress of API calls."
echo "This is especially helpful for longer operations like weekly/monthly reports." 