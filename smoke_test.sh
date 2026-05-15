#!/usr/bin/env bash
# Slide MCP Server Smoke Test
#
# Drives the freshly built binary in two modes:
#   1. `--tool / --args` one-shot mode - cheapest sanity check on each tool.
#   2. Full MCP stdio handshake - confirms the SDK transport still serves
#      initialize / tools/list / resources/list / a real tool call.
#
# Idempotent: safe to re-run any number of times. Only side effect is reading
# from your Slide account.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SERVER_BINARY="${SERVER_BINARY:-$ROOT_DIR/build/slide-mcp-server}"
TOOLS_MODE="${SLIDE_TOOLS:-full}"

if [ -z "${SLIDE_API_KEY:-}" ] && [ -f "$ROOT_DIR/.env" ]; then
    set -a
    # shellcheck disable=SC1090,SC1091
    . "$ROOT_DIR/.env"
    set +a
fi

if [ ! -x "$SERVER_BINARY" ]; then
    echo -e "${YELLOW}Building $SERVER_BINARY...${NC}"
    (cd "$ROOT_DIR" && go build -o "$SERVER_BINARY" .)
fi

if [ -z "${SLIDE_API_KEY:-}" ]; then
    echo -e "${RED}SLIDE_API_KEY is not set. Add it to .env or export it.${NC}" >&2
    exit 1
fi

export SLIDE_API_KEY
export SLIDE_TOOLS="$TOOLS_MODE"

TOTAL=0
PASSED=0
FAILED=0
SKIPPED_AUTH=0
FAILURES=()

echo -e "${BLUE}=================================${NC}"
echo -e "${BLUE}Slide MCP Server smoke test${NC}"
echo -e "${BLUE}=================================${NC}"
"$SERVER_BINARY" --version
echo "Tools mode: $SLIDE_TOOLS"
echo

run_one_shot() {
    local label="$1"
    local tool="$2"
    local args="$3"
    local accept_error_substr="${4:-}"

    TOTAL=$((TOTAL + 1))
    printf "%-48s" "$label"

    local out
    if ! out=$("$SERVER_BINARY" --tool "$tool" --args "$args" 2>/dev/null); then
        echo -e "${RED}FAIL${NC} (binary exited non-zero)"
        FAILED=$((FAILED + 1))
        FAILURES+=("$label: binary error")
        return
    fi

    if echo "$out" | grep -q '"isError": true'; then
        if [ -n "$accept_error_substr" ] && echo "$out" | grep -q "$accept_error_substr"; then
            echo -e "${GREEN}PASS${NC} (expected error: $accept_error_substr)"
            PASSED=$((PASSED + 1))
        elif echo "$out" | grep -q 'API error 401\|err_unauthorized'; then
            echo -e "${YELLOW}SKIP${NC} (API auth failed - stale SLIDE_API_KEY)"
            SKIPPED_AUTH=$((SKIPPED_AUTH + 1))
        else
            local snippet
            snippet=$(echo "$out" | sed -n 's/.*"text": "\(Error: [^"]\{0,160\}\).*/\1/p' | head -n1)
            echo -e "${RED}FAIL${NC} - ${snippet:-tool returned isError}"
            FAILED=$((FAILED + 1))
            FAILURES+=("$label: ${snippet:-tool returned isError}")
        fi
    else
        echo -e "${GREEN}PASS${NC}"
        PASSED=$((PASSED + 1))
    fi
}

echo -e "${BLUE}Core list operations${NC}"
run_one_shot "slide_devices list"           "slide_devices"   '{"operation":"list","limit":3}'
run_one_shot "slide_agents list"            "slide_agents"    '{"operation":"list","limit":3}'
run_one_shot "slide_backups list"           "slide_backups"   '{"operation":"list","limit":3}'
run_one_shot "slide_snapshots list"         "slide_snapshots" '{"operation":"list","limit":3}'
run_one_shot "slide_alerts list"            "slide_alerts"    '{"operation":"list","limit":3}'
run_one_shot "slide_clients list"           "slide_clients"   '{"operation":"list","limit":3}'
run_one_shot "slide_admin list_users"       "slide_admin"     '{"operation":"list_users","limit":3}'
run_one_shot "slide_recovery list_vms"      "slide_recovery"  '{"operation":"list_vms","limit":3}'
run_one_shot "slide_recovery list_images"   "slide_recovery"  '{"operation":"list_images","limit":3}'
run_one_shot "slide_recovery list_networks" "slide_recovery"  '{"operation":"list_networks","limit":3}'

echo
echo -e "${BLUE}v4 task-oriented operations${NC}"
run_one_shot "slide_overview health"        "slide_overview"  '{"operation":"health"}'
run_one_shot "slide_overview inventory"     "slide_overview"  '{"operation":"inventory"}'
run_one_shot "slide_audit recent"           "slide_audit"     '{"operation":"recent","hours":24}'
run_one_shot "slide_audit actions"          "slide_audit"     '{"operation":"actions","limit":5}'
run_one_shot "slide_alerts triage"          "slide_alerts"    '{"operation":"triage"}'
run_one_shot "slide_files restores"         "slide_files"     '{"operation":"list_restores","limit":3}'

echo
echo -e "${BLUE}v5 help / discovery surface${NC}"
run_one_shot "slide_help getting_started"   "slide_help"      '{"operation":"getting_started"}'
run_one_shot "slide_help examples"          "slide_help"      '{"operation":"examples"}'
run_one_shot "slide_help glossary"          "slide_help"      '{"operation":"glossary"}'
run_one_shot "slide_help troubleshoot"      "slide_help"      '{"operation":"troubleshoot"}'
run_one_shot "slide_help list_prompts"      "slide_help"      '{"operation":"list_prompts"}'
run_one_shot "slide_help list_resources"    "slide_help"      '{"operation":"list_resources"}'
run_one_shot "slide_help what_can_you_do"   "slide_help"      '{"operation":"what_can_you_do"}'

echo
echo -e "${BLUE}v5 name_hint resolution (live)${NC}"
run_one_shot "slide_overview for_client by name"   "slide_overview" '{"operation":"for_client","name_hint":"a"}'
run_one_shot "slide_backups recent by name"        "slide_backups"  '{"operation":"recent_for_agent","name_hint":"a","hours":24}'

echo
echo -e "${BLUE}Validation paths (expected errors)${NC}"
run_one_shot "slide_agents missing arg"     "slide_agents"   '{"operation":"get"}'                                  "agent_id is required"
run_one_shot "slide_devices missing arg"    "slide_devices"  '{"operation":"get_network"}'                          "device_id is required"
run_one_shot "slide_snapshots missing arg"  "slide_snapshots" '{"operation":"get_service_verification"}'            "snapshot_id is required"
run_one_shot "slide_admin avatar missing"   "slide_admin"    '{"operation":"get_user_avatar"}'                      "user_id is required"
run_one_shot "slide_files search missing"   "slide_files"    '{"operation":"search"}'                                "agent_id is required"
run_one_shot "slide_audit get missing"      "slide_audit"    '{"operation":"get"}'                                   "audit_id is required"
run_one_shot "slide_overview for_client"    "slide_overview" '{"operation":"for_client"}'                            "client_id is required"
run_one_shot "did-you-mean for typo'd op"   "slide_files"    '{"operation":"searh"}'                                 'did you mean "search"'

echo
echo -e "${BLUE}MCP stdio handshake${NC}"
TOTAL=$((TOTAL + 1))
HANDSHAKE_OUT=$(printf '%s\n' \
    '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
    '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
    '{"jsonrpc":"2.0","id":3,"method":"resources/list"}' \
    '{"jsonrpc":"2.0","id":4,"method":"resources/templates/list"}' \
    '{"jsonrpc":"2.0","id":5,"method":"prompts/list"}' \
    | "$SERVER_BINARY" 2>/dev/null || true)

if echo "$HANDSHAKE_OUT" | grep -q 'slide-mcp-server' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide_help' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide_overview' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide_files' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide_audit' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide://welcome' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide://overview/inventory' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide.welcome' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide.daily-status' \
   && echo "$HANDSHAKE_OUT" | grep -q 'slide.restore-file'; then
    echo -e "  initialize/tools/list/resources/list/templates/prompts ${GREEN}PASS${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "  ${RED}FAIL${NC} (missing expected fields in handshake)"
    FAILED=$((FAILED + 1))
    FAILURES+=("MCP handshake: missing expected fields")
fi

echo
echo -e "${BLUE}--doctor self-check${NC}"
TOTAL=$((TOTAL + 1))
DOCTOR_OUT=$("$SERVER_BINARY" --doctor 2>&1 || true)
if echo "$DOCTOR_OUT" | grep -q 'All checks passed'; then
    echo -e "  --doctor ${GREEN}PASS${NC}"
    PASSED=$((PASSED + 1))
elif echo "$DOCTOR_OUT" | grep -q 'Authentication.*FAIL\|err_unauthorized'; then
    echo -e "  --doctor ${YELLOW}SKIP${NC} (API auth failed - stale SLIDE_API_KEY)"
    SKIPPED_AUTH=$((SKIPPED_AUTH + 1))
else
    echo -e "  --doctor ${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
    FAILURES+=("--doctor: did not complete cleanly")
fi

echo
echo -e "${BLUE}=================================${NC}"
echo "Total:    $TOTAL"
echo -e "Passed:   ${GREEN}$PASSED${NC}"
echo -e "Failed:   ${RED}$FAILED${NC}"
echo -e "Skipped:  ${YELLOW}$SKIPPED_AUTH${NC} (live tests skipped due to API auth failure)"
if [ "$FAILED" -gt 0 ]; then
    echo
    echo "Failures:"
    for f in "${FAILURES[@]}"; do
        echo "  - $f"
    done
    exit 1
fi
if [ "$SKIPPED_AUTH" -gt 0 ]; then
    echo
    echo -e "${YELLOW}Note: $SKIPPED_AUTH live tests were skipped because SLIDE_API_KEY appears stale.${NC}"
    echo -e "${YELLOW}Update .env with a working token and re-run for full coverage.${NC}"
fi
echo -e "${GREEN}smoke test passed${NC}"
