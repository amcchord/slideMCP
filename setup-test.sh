#!/bin/bash

# Set up environment variables
export SLIDE_API_KEY="tk_fzu45leaboc0_jqbDMMvhVMFOAv7nhq7M5kEA9tb2hl14"
export SLIDE_API_URL="https://api.slide.tech"
export SLIDE_API_VERSION="v1"

# Test if the MCP server can respond to initialize request
echo "Testing MCP server response..."
echo '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{"tools":true}}}' | node mcp-test.js

echo ""
echo "Testing tools/list request..."
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | node mcp-test.js

echo ""
echo "If you see valid JSON responses above, the MCP server is working correctly." 