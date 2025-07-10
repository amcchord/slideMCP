#!/bin/bash

# Quick verification that the reports tool is integrated

API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"

echo "Verifying slide_reports tool integration..."
echo ""

# Check 1: Tool is listed
echo "1. Checking if slide_reports is in the tools list:"
echo "=================================================="
TOOLS_LIST=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

HAS_REPORTS=$(echo "$TOOLS_LIST" | jq '.result.tools[] | select(.name == "slide_reports") | .name' -r)

if [ "$HAS_REPORTS" == "slide_reports" ]; then
    echo "✓ slide_reports tool is registered"
else
    echo "✗ slide_reports tool NOT found"
    exit 1
fi

# Check 2: Tool description
echo ""
echo "2. Tool description:"
echo "==================="
echo "$TOOLS_LIST" | jq '.result.tools[] | select(.name == "slide_reports") | .description' -r

# Check 3: Available operations
echo ""
echo "3. Available operations:"
echo "======================="
echo "$TOOLS_LIST" | jq '.result.tools[] | select(.name == "slide_reports") | .inputSchema.properties.operation.enum' -r

# Check 4: Test with minimal parameters (will likely error but shows the tool is callable)
echo ""
echo "4. Testing tool is callable (with minimal params):"
echo "================================================="
ERROR_RESPONSE=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"slide_reports","arguments":{}}}' | \
./slide-mcp-server --api-key "$API_KEY" --exit-after-first 2>&1 | grep '^{' | tail -n 1)

ERROR_MSG=$(echo "$ERROR_RESPONSE" | jq '.result.content[0].text' -r)
echo "Response: $ERROR_MSG"

if [[ "$ERROR_MSG" == *"operation parameter is required"* ]]; then
    echo "✓ Tool responds correctly to missing parameters"
else
    echo "✗ Unexpected response"
fi

echo ""
echo "Verification complete!" 