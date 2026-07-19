#!/usr/bin/env bash
# Idempotently register the local Slide server with OpenAI Codex CLI/Desktop.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

if ! command -v codex >/dev/null 2>&1; then
    echo "ERROR: 'codex' CLI not found on PATH." >&2
    exit 1
fi

BIN="${1:-$ROOT_DIR/build/slide-mcp-server}"
if [ ! -x "$BIN" ]; then
    echo "ERROR: $BIN is not executable. Run 'make build' first." >&2
    exit 1
fi
if [ -z "${SLIDE_API_KEY:-}" ]; then
    echo "ERROR: SLIDE_API_KEY is not set." >&2
    exit 1
fi

# Codex CLI and the desktop app share MCP configuration. Remove first so the
# command is safe to re-run after replacing the binary or rotating the token.
codex mcp remove slide >/dev/null 2>&1 || true
codex mcp add slide --env "SLIDE_API_KEY=$SLIDE_API_KEY" -- "$BIN"

echo "Registered 'slide' for Codex. Verify with: codex mcp get slide"
