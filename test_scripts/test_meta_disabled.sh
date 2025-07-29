#!/bin/bash

# Test script to verify meta tools operations are disabled by default and enabled when reporting tools are on

echo "Testing slide_meta tools conditional availability..."
echo "=================================================="

# Check if API key is set
if [ -z "$SLIDE_API_KEY" ]; then
    echo "Error: SLIDE_API_KEY environment variable is not set"
    echo "This test requires an API key to run"
    exit 1
fi

echo -e "\n1. Testing with default settings (reporting/presentation tools disabled)..."
echo "Expected: get_snapshot_changes and get_reporting_data should not be available"

# Test list_all_clients_devices_and_agents (should work)
echo "   - Testing list_all_clients_devices_and_agents (should work)..."
LIST_RESULT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"list_all_clients_devices_and_agents"}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null)
if echo "$LIST_RESULT" | jq -e '.result.content[0].text' >/dev/null 2>&1; then
    echo "     ✓ list_all_clients_devices_and_agents works as expected"
else
    echo "     ✗ list_all_clients_devices_and_agents failed unexpectedly"
    echo "     Error: $(echo "$LIST_RESULT" | jq -r '.error.message // .result.content[0].text' 2>/dev/null)"
fi

# Test get_snapshot_changes (should fail)
echo "   - Testing get_snapshot_changes (should fail)..."
CHANGES_RESULT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"get_snapshot_changes","period":"day"}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null)
if echo "$CHANGES_RESULT" | jq -e '.result.content[0].text' | grep -q "requires reporting or presentation tools" 2>/dev/null; then
    echo "     ✓ get_snapshot_changes correctly disabled"
elif echo "$CHANGES_RESULT" | jq -e '.error.message' | grep -q "unknown operation" 2>/dev/null; then
    echo "     ✓ get_snapshot_changes correctly not available"
else
    echo "     ✗ get_snapshot_changes should be disabled but wasn't"
    echo "     Result: $(echo "$CHANGES_RESULT" | jq -r '.result.content[0].text // .error.message' 2>/dev/null)"
fi

# Test get_reporting_data (should fail)
echo "   - Testing get_reporting_data (should fail)..."
REPORT_RESULT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"get_reporting_data","report_type":"daily"}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --exit-after-first 2>/dev/null)
if echo "$REPORT_RESULT" | jq -e '.result.content[0].text' | grep -q "requires reporting or presentation tools" 2>/dev/null; then
    echo "     ✓ get_reporting_data correctly disabled"
elif echo "$REPORT_RESULT" | jq -e '.error.message' | grep -q "unknown operation" 2>/dev/null; then
    echo "     ✓ get_reporting_data correctly not available"
else
    echo "     ✗ get_reporting_data should be disabled but wasn't"
    echo "     Result: $(echo "$REPORT_RESULT" | jq -r '.result.content[0].text // .error.message' 2>/dev/null)"
fi

echo -e "\n2. Testing with reporting tools enabled..."
echo "Expected: get_snapshot_changes and get_reporting_data should be available"

# Test with --enable-reports flag
echo "   - Testing get_snapshot_changes with --enable-reports..."
CHANGES_ENABLED=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"get_snapshot_changes","period":"day","summary_only":true}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --enable-reports --exit-after-first 2>/dev/null)
if echo "$CHANGES_ENABLED" | jq -e '.result.content[0].text' | jq -e '.summary' >/dev/null 2>&1; then
    echo "     ✓ get_snapshot_changes works when reporting enabled"
else
    echo "     ✗ get_snapshot_changes failed when reporting enabled"
    echo "     Result: $(echo "$CHANGES_ENABLED" | jq -r '.result.content[0].text // .error.message' 2>/dev/null)"
fi

echo "   - Testing get_reporting_data with --enable-reports..."
REPORT_ENABLED=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"get_reporting_data","report_type":"daily"}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --enable-reports --exit-after-first 2>/dev/null)
if echo "$REPORT_ENABLED" | jq -e '.result.content[0].text' | jq -e '.metrics' >/dev/null 2>&1; then
    echo "     ✓ get_reporting_data works when reporting enabled"
else
    echo "     ✗ get_reporting_data failed when reporting enabled"
    echo "     Result: $(echo "$REPORT_ENABLED" | jq -r '.result.content[0].text // .error.message' 2>/dev/null)"
fi

echo -e "\n3. Testing with presentation tools enabled..."
echo "Expected: get_snapshot_changes and get_reporting_data should be available"

echo "   - Testing get_snapshot_changes with --enable-presentation..."
CHANGES_PRES=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"slide_meta","arguments":{"operation":"get_snapshot_changes","period":"day","summary_only":true}}}' | ./slide-mcp-server --api-key "$SLIDE_API_KEY" --enable-presentation --exit-after-first 2>/dev/null)
if echo "$CHANGES_PRES" | jq -e '.result.content[0].text' | jq -e '.summary' >/dev/null 2>&1; then
    echo "     ✓ get_snapshot_changes works when presentation enabled"
else
    echo "     ✗ get_snapshot_changes failed when presentation enabled"
    echo "     Result: $(echo "$CHANGES_PRES" | jq -r '.result.content[0].text // .error.message' 2>/dev/null)"
fi

echo -e "\nTest completed!"
echo "===============" 