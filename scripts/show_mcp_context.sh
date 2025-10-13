#!/bin/bash

# Show MCP Initial Context Script
# This script shows the initial context that the Slide MCP Server provides to LLMs
# Usage: ./show_mcp_context.sh <API_KEY> [PATH_TO_MCP_SERVER]

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper function to print colored output
print_header() {
    echo -e "${CYAN}=================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}=================================${NC}"
}

print_section() {
    echo -e "\n${YELLOW}--- $1 ---${NC}"
}

print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

# Check command line arguments
if [ $# -lt 1 ]; then
    echo "Usage: $0 <API_KEY> [PATH_TO_MCP_SERVER]"
    echo
    echo "Arguments:"
    echo "  API_KEY           Your Slide API key"
    echo "  PATH_TO_MCP_SERVER Optional path to slide-mcp-server binary"
    echo "                    If not provided, will search in common locations"
    echo
    echo "Examples:"
    echo "  $0 your-api-key-here"
    echo "  $0 your-api-key-here /usr/local/bin/slide-mcp-server"
    echo "  $0 your-api-key-here ./slide-mcp-server"
    exit 1
fi

API_KEY="$1"
MCP_SERVER_PATH="${2:-}"

# Find the MCP server binary if not provided
if [ -z "$MCP_SERVER_PATH" ]; then
    print_section "Searching for slide-mcp-server binary"
    
    # Common locations to search
    SEARCH_PATHS=(
        "./slide-mcp-server"
        "../slide-mcp-server"
        "/usr/local/bin/slide-mcp-server"
        "$HOME/.local/bin/slide-mcp-server"
        "./release/v2.4.0/slide-mcp-server-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
    )
    
    # Handle macOS architecture naming
    if [ "$(uname -s)" = "Darwin" ]; then
        if [ "$(uname -m)" = "arm64" ]; then
            SEARCH_PATHS+=("./release/v2.4.0/slide-mcp-server-macos-arm64")
        else
            SEARCH_PATHS+=("./release/v2.4.0/slide-mcp-server-macos-x64")
        fi
    fi
    
    for path in "${SEARCH_PATHS[@]}"; do
        if [ -f "$path" ] && [ -x "$path" ]; then
            MCP_SERVER_PATH="$path"
            echo "Found MCP server at: $MCP_SERVER_PATH"
            break
        fi
    done
    
    if [ -z "$MCP_SERVER_PATH" ]; then
        print_error "Could not find slide-mcp-server binary in common locations."
        echo "Please specify the path as the second argument."
        echo "Searched in:"
        for path in "${SEARCH_PATHS[@]}"; do
            echo "  $path"
        done
        exit 1
    fi
fi

# Verify the binary exists and is executable
if [ ! -f "$MCP_SERVER_PATH" ]; then
    print_error "MCP server binary not found at: $MCP_SERVER_PATH"
    exit 1
fi

if [ ! -x "$MCP_SERVER_PATH" ]; then
    print_error "MCP server binary is not executable: $MCP_SERVER_PATH"
    exit 1
fi

print_header "Slide MCP Server Initial Context Viewer"
echo "API Key: ${API_KEY:0:8}..."
echo "MCP Server: $MCP_SERVER_PATH"
echo

# Create temporary files for communication
TEMP_DIR=$(mktemp -d)
INIT_REQUEST="$TEMP_DIR/init_request.json"
TOOLS_REQUEST="$TEMP_DIR/tools_request.json"
INIT_RESPONSE="$TEMP_DIR/init_response.json"
TOOLS_RESPONSE="$TEMP_DIR/tools_response.json"

# Cleanup function
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Create the initialize request
cat > "$INIT_REQUEST" << 'EOF'
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"context-viewer","version":"1.0.0"}}}
EOF

# Create the tools/list request  
cat > "$TOOLS_REQUEST" << 'EOF'
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
EOF

print_section "Starting MCP Server and fetching initial context"

# Start the MCP server and send requests
{
    echo "Sending initialize request..."
    cat "$INIT_REQUEST"
    echo
    echo "Sending tools/list request..."
    cat "$TOOLS_REQUEST"
    echo
} | SLIDE_API_KEY="$API_KEY" "$MCP_SERVER_PATH" > "$TEMP_DIR/responses.json" 2>"$TEMP_DIR/server.log"

# Check if server produced output
if [ ! -s "$TEMP_DIR/responses.json" ]; then
    print_error "No response from MCP server"
    echo "Server log:"
    cat "$TEMP_DIR/server.log"
    exit 1
fi

# Split the responses (assuming they come on separate lines)
head -n 1 "$TEMP_DIR/responses.json" > "$INIT_RESPONSE"
if [ $(wc -l < "$TEMP_DIR/responses.json") -gt 1 ]; then
    tail -n 1 "$TEMP_DIR/responses.json" > "$TOOLS_RESPONSE"
fi

print_success "Successfully received responses from MCP server"

# Parse and display the initialize response
print_header "INITIALIZATION RESPONSE"

if command -v jq &> /dev/null; then
    # Use jq for pretty formatting if available
    print_section "Server Information"
    jq -r '.result.serverInfo | "Name: \(.name)\nVersion: \(.version)"' "$INIT_RESPONSE"
    
    print_section "Protocol Information"  
    jq -r '.result | "Protocol Version: \(.protocolVersion)"' "$INIT_RESPONSE"
    
    print_section "Initial Context Metadata"
    jq -r '.result.initialContext._metadata | "Description: \(.description)\nSource Tool: \(.source_tool)\nUsage Note: \(.usage_note)\nTimestamp: \(.timestamp)"' "$INIT_RESPONSE"
    
    print_section "Clients, Devices, and Agents Overview"
    echo "This is the hierarchical data structure provided to the LLM:"
    echo
    
    # Display a summary of the data structure
    if jq -e '.result.initialContext.clients_devices_agents.data' "$INIT_RESPONSE" > /dev/null 2>&1; then
        CLIENT_COUNT=$(jq '.result.initialContext.clients_devices_agents.data | length' "$INIT_RESPONSE")
        echo -e "${GREEN}Total Clients: $CLIENT_COUNT${NC}"
        
        # Count total devices and agents
        TOTAL_DEVICES=$(jq '[.result.initialContext.clients_devices_agents.data[].devices | length] | add // 0' "$INIT_RESPONSE")
        TOTAL_AGENTS=$(jq '[.result.initialContext.clients_devices_agents.data[].devices[].agents | length] | add // 0' "$INIT_RESPONSE")
        
        echo -e "${GREEN}Total Devices: $TOTAL_DEVICES${NC}"
        echo -e "${GREEN}Total Agents: $TOTAL_AGENTS${NC}"
        echo
        
        # Show client structure
        jq -r '.result.initialContext.clients_devices_agents.data[] | 
        "Client: \(.client_name // "Unassigned") (\(.client_id // "N/A"))
        └── \(.devices | length) device(s)
        \(if .devices then [.devices[] | "    └── \(.display_name // .hostname) (\(.agents | length) agent(s))"] | join("\n") else "" end)
        "' "$INIT_RESPONSE"
    else
        echo "No client data found in response"
    fi
    
    print_section "Full Initial Context Data (JSON)"
    echo "This is the complete data structure the LLM receives:"
    jq '.result.initialContext.clients_devices_agents' "$INIT_RESPONSE"
    
else
    # Fallback to basic text parsing if jq is not available
    print_section "Raw Initialize Response"
    echo "Note: Install 'jq' for better formatting"
    cat "$INIT_RESPONSE" | python3 -m json.tool 2>/dev/null || cat "$INIT_RESPONSE"
fi

# Display tools list if available
if [ -f "$TOOLS_RESPONSE" ] && [ -s "$TOOLS_RESPONSE" ]; then
    print_header "AVAILABLE TOOLS"
    
    if command -v jq &> /dev/null; then
        TOOL_COUNT=$(jq '.result.tools | length' "$TOOLS_RESPONSE")
        echo -e "${GREEN}Total Available Tools: $TOOL_COUNT${NC}"
        echo
        
        print_section "Tool Summary"
        jq -r '.result.tools[] | "• \(.name): \(.description)"' "$TOOLS_RESPONSE"
        
        print_section "Full Tools Response (JSON)"
        jq '.result.tools' "$TOOLS_RESPONSE"
    else
        print_section "Raw Tools Response"
        cat "$TOOLS_RESPONSE" | python3 -m json.tool 2>/dev/null || cat "$TOOLS_RESPONSE"
    fi
fi

print_header "SUMMARY"
echo "The MCP server provides the following context to LLMs upon initialization:"
echo
echo "1. ${YELLOW}Server Information${NC}: Name, version, protocol version"
echo "2. ${YELLOW}Initial Context${NC}: Complete hierarchy of clients → devices → agents"
echo "3. ${YELLOW}Available Tools${NC}: All tools the LLM can use to interact with Slide"
echo
echo "This context allows the LLM to immediately understand the infrastructure"
echo "without needing to make additional API calls for basic information."

print_success "Context display complete!"