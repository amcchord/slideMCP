#!/usr/bin/env bash
# Idempotent merge of the slide MCP server into Claude Desktop's config.
#
# Usage:
#   SLIDE_API_KEY=tk_... ./scripts/install-claude-desktop.sh [path/to/slide-mcp-server]
#
# Writes/updates the "slide" entry in mcpServers without disturbing other
# servers. Safe to re-run.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: 'jq' is required (brew install jq)." >&2
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
    exit 1
fi

case "$(uname -s)" in
    Darwin)
        CONFIG_DIR="$HOME/Library/Application Support/Claude"
        ;;
    Linux)
        CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/Claude"
        ;;
    *)
        echo "ERROR: This installer only supports macOS and Linux." >&2
        echo "       On Windows, edit %APPDATA%\\Claude\\claude_desktop_config.json by hand." >&2
        exit 1
        ;;
esac

CONFIG_FILE="$CONFIG_DIR/claude_desktop_config.json"
mkdir -p "$CONFIG_DIR"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating new $CONFIG_FILE"
    echo '{}' >"$CONFIG_FILE"
fi

# Make sure the file is valid JSON; if not, abort instead of clobbering it.
if ! jq -e . "$CONFIG_FILE" >/dev/null 2>&1; then
    echo "ERROR: $CONFIG_FILE is not valid JSON. Refusing to modify." >&2
    exit 1
fi

# Take a one-time backup the first time we touch the file in a given day.
BACKUP="$CONFIG_FILE.slide-backup"
if [ ! -f "$BACKUP" ]; then
    cp "$CONFIG_FILE" "$BACKUP"
    echo "Saved backup at $BACKUP"
fi

TMP="$(mktemp)"
jq \
    --arg bin "$BIN" \
    --arg key "$SLIDE_API_KEY" \
    '.mcpServers = (.mcpServers // {}) | .mcpServers.slide = {command: $bin, env: {SLIDE_API_KEY: $key}}' \
    "$CONFIG_FILE" >"$TMP"

mv "$TMP" "$CONFIG_FILE"
echo "Updated 'slide' entry in $CONFIG_FILE"
echo
echo "Restart Claude Desktop to pick up the change."
