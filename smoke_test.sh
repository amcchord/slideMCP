#!/bin/bash

# Slide MCP Server Smoke Test
# Tests at least one operation from each tool to validate functionality

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
API_KEY="tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"
SERVER_BINARY="./slide-mcp-server"
TEST_TIMEOUT=10
SERVER_PID=""

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo -e "${BLUE}🧪 Slide MCP Server Smoke Test${NC}"
echo -e "${BLUE}=================================${NC}\n"

# Cleanup function
cleanup() {
    if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
        echo -e "\n${YELLOW}🔄 Cleaning up server (PID: $SERVER_PID)...${NC}"
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# Check if server binary exists
if [[ ! -f "$SERVER_BINARY" ]]; then
    echo -e "${RED}❌ Server binary not found: $SERVER_BINARY${NC}"
    echo -e "${YELLOW}💡 Run 'go build -o slide-mcp-server .' first${NC}"
    exit 1
fi

# Function to test a tool operation
test_tool() {
    local tool_name="$1"
    local operation="$2"
    local description="$3"
    local additional_args="${4:-{}}"
    
    echo -e "${BLUE}Testing ${tool_name} (${operation})...${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Create JSON-RPC request
    local request_id=$((RANDOM))
    local args_json="{\"operation\": \"$operation\""
    
    # Add additional arguments if provided
    if [[ "$additional_args" != "{}" ]]; then
        # Remove opening and closing braces from additional_args and add to request
        local cleaned_args="${additional_args#\{}"
        cleaned_args="${cleaned_args%\}}"
        if [[ -n "$cleaned_args" ]]; then
            args_json="$args_json, $cleaned_args"
        fi
    fi
    args_json="$args_json}"
    
    local json_request=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "id": $request_id,
    "method": "tools/call",
    "params": {
        "name": "$tool_name",
        "arguments": $args_json
    }
}
EOF
)
    
    # Send request and capture response with timeout
    local response=""
    if response=$(echo "$json_request" | timeout $TEST_TIMEOUT "$SERVER_BINARY" --exit-after-first 2>/dev/null); then
        # Check if response contains error
        if echo "$response" | grep -q '"error"'; then
            local error_msg=$(echo "$response" | grep -o '"message":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "Unknown error")
            echo -e "  ${RED}❌ FAILED: $error_msg${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            echo -e "  ${GREEN}✅ PASSED: $description${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        fi
    else
        echo -e "  ${RED}❌ FAILED: Timeout or server error${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Set environment variables
export SLIDE_API_KEY="$API_KEY"
export SLIDE_TOOLS="full"  # Enable all tools for comprehensive testing

echo -e "${YELLOW}🔧 Environment Setup:${NC}"
echo -e "  API Key: ${API_KEY:0:10}..."
echo -e "  Tools Mode: full"
echo -e "  Binary: $SERVER_BINARY"
echo -e ""

# Test server version first
echo -e "${BLUE}📋 Testing server version...${NC}"
if version_output=$("$SERVER_BINARY" --version 2>/dev/null); then
    echo -e "  ${GREEN}✅ Server version: $version_output${NC}"
else
    echo -e "  ${RED}❌ Failed to get server version${NC}"
    exit 1
fi
echo ""

echo -e "${YELLOW}🚀 Starting smoke tests...${NC}\n"

# Test each tool with a basic operation
echo -e "${BLUE}Testing Core Tools:${NC}"

# Test slide_devices (basic device listing)
test_tool "slide_devices" "list" "List devices" '{"limit": 5}'

# Test slide_agents (basic agent listing) 
test_tool "slide_agents" "list" "List agents" '{"limit": 5}'

# Test slide_backups (basic backup listing)
test_tool "slide_backups" "list" "List backups" '{"limit": 5}'

# Test slide_snapshots (basic snapshot listing)
test_tool "slide_snapshots" "list" "List snapshots" '{"limit": 5}'

# Test slide_alerts (basic alert listing)
test_tool "slide_alerts" "list" "List alerts" '{"limit": 5}'

# Test slide_networks (basic network listing)
test_tool "slide_networks" "list" "List networks" '{"limit": 5}'

# Test slide_vms (basic VM listing)
test_tool "slide_vms" "list" "List virtual machines" '{"limit": 5}'

# Test slide_user_management (basic user listing)
test_tool "slide_user_management" "list_users" "List users" '{"limit": 5}'

echo -e "\n${BLUE}Testing Restore Tools:${NC}"

# Test slide_restores (basic file restore listing)
test_tool "slide_restores" "list_file_restores" "List file restores" '{"limit": 5}'

echo -e "\n${BLUE}Testing Reporting Tools:${NC}"

# Test slide_reports (basic reporting)
test_tool "slide_reports" "daily_backup_snapshot" "Generate daily report" '{"date": "2024-01-01"}'

# Test slide_presentation (context download)
test_tool "slide_presentation" "download_context" "Download presentation context" '{}'

echo -e "\n${BLUE}Testing Meta Tools:${NC}"

# Test slide_meta (comprehensive overview)
test_tool "slide_meta" "list_all_clients_devices_and_agents" "Get system overview" '{}'

# Test slide_docs (documentation access)
test_tool "slide_docs" "get_openapi_spec" "Get API specification" '{}'

# Test backward compatibility tool
test_tool "list_all_clients_devices_and_agents" "" "Legacy compatibility tool" '{}'

echo -e "\n${BLUE}=================================${NC}"
echo -e "${BLUE}📊 Smoke Test Results Summary${NC}"
echo -e "${BLUE}=================================${NC}"

echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}✅ Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}❌ Failed: ${FAILED_TESTS}${NC}"

# Calculate success rate
if [[ $TOTAL_TESTS -gt 0 ]]; then
    success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo -e "Success Rate: ${success_rate}%"
    
    if [[ $success_rate -ge 80 ]]; then
        echo -e "\n${GREEN}🎉 SMOKE TEST PASSED! System appears to be working correctly.${NC}"
        exit 0
    elif [[ $success_rate -ge 50 ]]; then
        echo -e "\n${YELLOW}⚠️  PARTIAL SUCCESS: Some issues detected but core functionality works.${NC}"
        exit 1
    else
        echo -e "\n${RED}💥 SMOKE TEST FAILED! Major issues detected.${NC}"
        exit 1
    fi
else
    echo -e "\n${RED}💥 No tests were executed!${NC}"
    exit 1
fi