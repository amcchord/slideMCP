#!/usr/bin/env bash
# Verify the complete release contract before a tag or GitHub release exists.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
BINARY_NAME="slide-mcp-server"
VERSION="${1:-}"
REQUIRE_NOTARIZED="${REQUIRE_NOTARIZED:-false}"

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Usage: $0 vX.Y.Z" >&2
    exit 2
fi

cd "$BUILD_DIR"

if [ ! -s release-assets.txt ] || [ ! -s checksums.sha256 ]; then
    echo "ERROR: release manifest missing; run scripts/package-release.sh first" >&2
    exit 1
fi

echo "==> Verifying checksums"
shasum -a 256 -c checksums.sha256

echo "==> Verifying release asset manifest"
while IFS= read -r asset; do
    if [ ! -s "$asset" ]; then
        echo "FAIL: missing or empty asset: $asset" >&2
        exit 1
    fi
done < release-assets.txt

for required in \
    "$BINARY_NAME-$VERSION-linux-x64.tar.gz" \
    "$BINARY_NAME-$VERSION-macos-x64.tar.gz" \
    "$BINARY_NAME-$VERSION-macos-arm64.tar.gz"; do
    if [ ! -s "$required" ]; then
        echo "FAIL: legacy installer compatibility asset missing: $required" >&2
        exit 1
    fi
done

echo "==> Verifying archive contents"
while IFS= read -r archive; do
    tar -tzf "$archive" | grep -q "$BINARY_NAME"
done < <(find . -maxdepth 1 -name "$BINARY_NAME-$VERSION-*.tar.gz" -print | sort)
unzip -Z1 "$BINARY_NAME-$VERSION-windows-x64.zip" | grep -q "$BINARY_NAME-windows-amd64.exe"

if ! cmp -s "$BINARY_NAME.mcpb" "$BINARY_NAME-$VERSION.mcpb"; then
    echo "FAIL: versioned and stable-name MCPB assets differ" >&2
    exit 1
fi

echo "==> Verifying MCPB structure and embedded version"
"$ROOT_DIR/scripts/verify-dxt.sh" "$BUILD_DIR/$BINARY_NAME.mcpb"

if [ "$(uname -s)" = "Darwin" ] && [ "$REQUIRE_NOTARIZED" = "true" ]; then
    WORK_DIR="$(mktemp -d)"
    cleanup() { rm -rf "$WORK_DIR"; }
    trap cleanup EXIT
    unzip -q "$BINARY_NAME.mcpb" 'server/slide-mcp-server-darwin-universal' -d "$WORK_DIR"
    MAC_BINARY="$WORK_DIR/server/slide-mcp-server-darwin-universal"

    echo "==> Enforcing Developer ID signature and Apple notarization"
    codesign --verify --strict --verbose=2 "$MAC_BINARY"
    AUTHORITY="$(codesign -dvv "$MAC_BINARY" 2>&1 | sed -n 's/^Authority=//p' | head -1)"
    if [[ "$AUTHORITY" != Developer\ ID\ Application:* ]]; then
        echo "FAIL: unexpected macOS signing authority: $AUTHORITY" >&2
        exit 1
    fi
    codesign --test-requirement="=notarized" --verify "$MAC_BINARY"
fi

echo "==> verify-release: ok"
