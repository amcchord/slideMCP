#!/usr/bin/env bash
# Assemble the exact GitHub release asset set, including names consumed by the
# retired Fyne installer. The legacy installer asks GitHub's latest-release API
# for macos-x64/linux-x64 assets, so those aliases are an upgrade contract.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
BINARY_NAME="slide-mcp-server"
VERSION="${1:-}"

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Usage: $0 vX.Y.Z" >&2
    exit 2
fi

require_file() {
    if [ ! -f "$1" ]; then
        echo "ERROR: required release input is missing: $1" >&2
        exit 1
    fi
}

LINUX_AMD64="$BUILD_DIR/$BINARY_NAME-linux-amd64"
LINUX_ARM64="$BUILD_DIR/$BINARY_NAME-linux-arm64"
MAC_UNIVERSAL="$BUILD_DIR/$BINARY_NAME-darwin-universal"
WINDOWS_AMD64="$BUILD_DIR/$BINARY_NAME-windows-amd64.exe"
MCPB="$BUILD_DIR/$BINARY_NAME.mcpb"

for input in "$LINUX_AMD64" "$LINUX_ARM64" "$MAC_UNIVERSAL" "$WINDOWS_AMD64" "$MCPB"; do
    require_file "$input"
done

cd "$BUILD_DIR"

archive_tar() {
    local output="$1"
    local input="$2"
    rm -f "$output"
    tar -czf "$output" "$(basename "$input")"
}

# Canonical archives.
archive_tar "$BINARY_NAME-$VERSION-linux-amd64.tar.gz" "$LINUX_AMD64"
archive_tar "$BINARY_NAME-$VERSION-linux-arm64.tar.gz" "$LINUX_ARM64"
archive_tar "$BINARY_NAME-$VERSION-macos-universal.tar.gz" "$MAC_UNIVERSAL"
rm -f "$BINARY_NAME-$VERSION-windows-x64.zip"
zip -q "$BINARY_NAME-$VERSION-windows-x64.zip" "$(basename "$WINDOWS_AMD64")"

# Stable compatibility aliases. Every macOS alias contains the signed and
# notarized universal binary, so old Intel and Apple Silicon installers both
# receive the exact release-grade executable.
cp "$BINARY_NAME-$VERSION-linux-amd64.tar.gz" "$BINARY_NAME-$VERSION-linux-x64.tar.gz"
for alias in \
    darwin-amd64 darwin-arm64 darwin-universal \
    macos-amd64 macos-arm64 macos-x64; do
    cp "$BINARY_NAME-$VERSION-macos-universal.tar.gz" "$BINARY_NAME-$VERSION-$alias.tar.gz"
done

cp "$MCPB" "$BINARY_NAME-$VERSION.mcpb"

ASSETS=(
    "$BINARY_NAME.mcpb"
    "$BINARY_NAME-$VERSION.mcpb"
    "$BINARY_NAME-$VERSION-linux-amd64.tar.gz"
    "$BINARY_NAME-$VERSION-linux-x64.tar.gz"
    "$BINARY_NAME-$VERSION-linux-arm64.tar.gz"
    "$BINARY_NAME-$VERSION-darwin-amd64.tar.gz"
    "$BINARY_NAME-$VERSION-darwin-arm64.tar.gz"
    "$BINARY_NAME-$VERSION-darwin-universal.tar.gz"
    "$BINARY_NAME-$VERSION-macos-amd64.tar.gz"
    "$BINARY_NAME-$VERSION-macos-arm64.tar.gz"
    "$BINARY_NAME-$VERSION-macos-x64.tar.gz"
    "$BINARY_NAME-$VERSION-macos-universal.tar.gz"
    "$BINARY_NAME-$VERSION-windows-x64.zip"
)

rm -f checksums.sha256 release-assets.txt
for asset in "${ASSETS[@]}"; do
    require_file "$BUILD_DIR/$asset"
    shasum -a 256 "$asset" >> checksums.sha256
    printf '%s\n' "$BUILD_DIR/$asset" >> release-assets.txt
done
printf '%s\n' "$BUILD_DIR/checksums.sha256" >> release-assets.txt

echo "Packaged ${#ASSETS[@]} release assets for $VERSION"
echo "Compatibility aliases: linux-x64, macos-x64, macos-arm64"
