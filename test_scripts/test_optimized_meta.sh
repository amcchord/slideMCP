#!/bin/bash

echo "Testing optimized meta tools output sizes..."
echo "========================================="

# Test get_snapshot_changes with different modes
echo -e "\n1. Testing get_snapshot_changes with summary_only=true..."
for period in day week month; do
    echo "   - Period: $period (summary only)"
    SUMMARY_OUTPUT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_meta\",\"arguments\":{\"operation\":\"get_snapshot_changes\",\"period\":\"$period\",\"summary_only\":true}}}" | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null | jq -r '.result.content[0].text' 2>/dev/null)
    SUMMARY_SIZE=${#SUMMARY_OUTPUT}
    echo "     Output size: $SUMMARY_SIZE characters"
    
    # Parse summary
    if [ "$period" = "month" ]; then
        echo "$SUMMARY_OUTPUT" | jq '.summary' 2>/dev/null || echo "     Failed to parse summary"
    fi
done

echo -e "\n2. Testing get_snapshot_changes with detailed mode (no metadata)..."
DETAILED_OUTPUT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"get_snapshot_changes","period":"day","summary_only":false,"include_metadata":false}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null | jq -r '.result.content[0].text' 2>/dev/null)
DETAILED_SIZE=${#DETAILED_OUTPUT}
echo "   Output size (day, no metadata): $DETAILED_SIZE characters"

echo -e "\n3. Testing get_snapshot_changes with detailed mode (with metadata)..."
DETAILED_META_OUTPUT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"get_snapshot_changes","period":"day","summary_only":false,"include_metadata":true}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null | jq -r '.result.content[0].text' 2>/dev/null)
DETAILED_META_SIZE=${#DETAILED_META_OUTPUT}
echo "   Output size (day, with metadata): $DETAILED_META_SIZE characters"

echo -e "\n4. Testing get_reporting_data (optimized)..."
for report_type in daily weekly monthly; do
    echo "   - Report type: $report_type"
    REPORT_OUTPUT=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"slide_meta\",\"arguments\":{\"operation\":\"get_reporting_data\",\"report_type\":\"$report_type\"}}}" | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null | jq -r '.result.content[0].text' 2>/dev/null)
    REPORT_SIZE=${#REPORT_OUTPUT}
    echo "     Output size: $REPORT_SIZE characters"
done

echo -e "\nDone!" 