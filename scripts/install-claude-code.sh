#!/usr/bin/env bash
# Idempotent installer that registers slide-mcp-server with Claude Code.
#
# Usage:
#   SLIDE_API_KEY=tk_... ./scripts/install-claude-code.sh [path/to/slide-mcp-server]
#
# If no binary path is supplied, the script will prefer:
#   1) build/slide-mcp-server in the repo
#   2) /usr/local/bin/slide-mcp-server
#   3) the first slide-mcp-server on $PATH

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

if ! command -v claude >/dev/null 2>&1; then
    echo "ERROR: 'claude' CLI not found on PATH. Install Claude Code first: https://claude.com/code" >&2
    exit 1
fi

BIN="${1:-}"
if [ -z "$BIN" ]; then
    if [ -x "$ROOT_DIR/build/slide-mcp-server" ]; then
        BIN="$ROOT_DIR/build/slide-mcp-server"
    elif [ -x "/usr/local/bin/slide-mcp-server" ]; then
        BIN="/usr/local/bin/slide-mcp-server"
    elif command -v slide-mcp-server >/dev/null 2>&1; then
        BIN="$(command -v slide-mcp-server)"
    else
        echo "ERROR: could not find a slide-mcp-server binary." >&2
        echo "       Run 'make build' first, or pass a path:" >&2
        echo "         $0 /path/to/slide-mcp-server" >&2
        exit 1
    fi
fi

if [ ! -x "$BIN" ]; then
    echo "ERROR: $BIN is not executable." >&2
    exit 1
fi

if [ -z "${SLIDE_API_KEY:-}" ]; then
    echo "ERROR: SLIDE_API_KEY is not set in your environment." >&2
    echo "       Generate one at https://console.slide.tech/ and re-run:" >&2
    echo "         SLIDE_API_KEY=tk_... $0" >&2
    exit 1
fi

# `claude mcp add` is the supported way to register an MCP server with Claude
# Code. If a server with the same name already exists we remove it first so
# this script stays idempotent.
if claude mcp list 2>/dev/null | grep -q '^slide '; then
    echo "Removing existing 'slide' MCP server registration..."
    claude mcp remove slide >/dev/null
fi

echo "Registering 'slide' MCP server -> $BIN"
claude mcp add slide \
    --env SLIDE_API_KEY="$SLIDE_API_KEY" \
    -- "$BIN"

echo
echo "Done. Verify with: claude mcp list"
