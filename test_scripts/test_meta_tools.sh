#!/bin/bash

# Test script for meta tools output size

echo "Testing slide_meta tools output sizes..."
echo "======================================="

# Test list_all_clients_devices_and_agents
echo -e "\n1. Testing list_all_clients_devices_and_agents..."
HIERARCHY_OUTPUT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"list_all_clients_devices_and_agents"}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null | jq -r '.result.content[0].text' 2>/dev/null)
HIERARCHY_SIZE=${#HIERARCHY_OUTPUT}
echo "   Output size: $HIERARCHY_SIZE characters"

# Test get_snapshot_changes for different periods
echo -e "\n2. Testing get_snapshot_changes..."
for period in day week month; do
    echo "   - Period: $period"
    CHANGES_OUTPUT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_meta\",\"arguments\":{\"operation\":\"get_snapshot_changes\",\"period\":\"$period\"}}}" | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null | jq -r '.result.content[0].text' 2>/dev/null)
    CHANGES_SIZE=${#CHANGES_OUTPUT}
    echo "     Output size: $CHANGES_SIZE characters"
done

# Test get_reporting_data for different report types
echo -e "\n3. Testing get_reporting_data..."
for report_type in daily weekly monthly; do
    echo "   - Report type: $report_type"
    REPORT_OUTPUT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_meta\",\"arguments\":{\"operation\":\"get_reporting_data\",\"report_type\":\"$report_type\"}}}" | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null | jq -r '.result.content[0].text' 2>/dev/null)
    REPORT_SIZE=${#REPORT_OUTPUT}
    echo "     Output size: $REPORT_SIZE characters"
    
    # Save monthly output for analysis
    if [ "$report_type" = "monthly" ]; then
        echo "$REPORT_OUTPUT" > monthly_report_output.json
        echo "     Saved to monthly_report_output.json for analysis"
    fi
done

echo -e "\n4. Analyzing monthly report structure..."
if [ -f monthly_report_output.json ]; then
    # Count different sections
    echo "   - Number of clients: $(echo "$REPORT_OUTPUT" | jq '.hierarchy.clients | length' 2>/dev/null || echo "N/A")"
    echo "   - Number of new snapshots: $(echo "$REPORT_OUTPUT" | jq '.snapshot_changes.summary.total_new' 2>/dev/null || echo "N/A")"
    echo "   - Number of deleted snapshots: $(echo "$REPORT_OUTPUT" | jq '.snapshot_changes.summary.total_deleted' 2>/dev/null || echo "N/A")"
fi

echo -e "\nDone!" 