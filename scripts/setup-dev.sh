#!/usr/bin/env bash
# Idempotent dev environment setup for slide-mcp-server.
# Safe to re-run; only installs what's missing.

set -euo pipefail

if ! command -v brew >/dev/null 2>&1; then
    echo "ERROR: Homebrew is required. Install it from https://brew.sh and re-run." >&2
    exit 1
fi

ensure_brew_pkg() {
    local pkg="$1"
    if brew list --formula "$pkg" >/dev/null 2>&1; then
        echo "ok: $pkg already installed"
    else
        echo "installing: $pkg"
        brew install "$pkg"
    fi
}

ensure_brew_pkg go
ensure_brew_pkg jq
ensure_brew_pkg gh

echo
echo "--- toolchain ---"
go version
echo "GOPATH=$(go env GOPATH)"
echo "GOROOT=$(go env GOROOT)"

if command -v gh >/dev/null 2>&1; then
    echo "gh $(gh --version | head -n1 | awk '{print $3}')"
fi

GO_BIN="$(go env GOPATH)/bin"
case ":${PATH}:" in
    *":${GO_BIN}:"*) ;;
    *)
        echo
        echo "NOTE: $GO_BIN is not on your PATH."
        echo "      Add this to your ~/.zshrc:  export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
        ;;
esac

echo
echo "Dev environment ready. Try: make doctor"
