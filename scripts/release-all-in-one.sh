#!/usr/bin/env bash
# Production release pipeline. It deliberately does not edit or commit source:
# a release must be built from a clean, reviewed commit whose three version
# declarations already agree.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
REPO="amcchord/slideMCP"
VERSION="${1:-}"

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Usage: $0 vX.Y.Z" >&2
    exit 2
fi

if [ -f "$ROOT_DIR/scripts/release-env.sh" ]; then
    # Local, gitignored Apple credentials and signing identity.
    # shellcheck disable=SC1091
    source "$ROOT_DIR/scripts/release-env.sh"
fi

DEVELOPER_ID="${DEVELOPER_ID:-}"
KEYCHAIN_PROFILE="${KEYCHAIN_PROFILE:-slide-mcp-release}"

fail() {
    echo "ERROR: $*" >&2
    exit 1
}

for command in go git gh jq make shasum unzip zip codesign xcrun; do
    command -v "$command" >/dev/null 2>&1 || fail "required command not found: $command"
done

[ "$(uname -s)" = "Darwin" ] || fail "production releases must run on macOS for signing and notarization"
[ -n "$DEVELOPER_ID" ] || fail "DEVELOPER_ID is unset; run scripts/setup-signing.sh"
[ -n "$KEYCHAIN_PROFILE" ] || fail "KEYCHAIN_PROFILE is unset; run scripts/setup-signing.sh"

cd "$ROOT_DIR"
[ "$(git branch --show-current)" = "main" ] || fail "releases must be cut from main"
[ -z "$(git status --porcelain)" ] || fail "working tree is not clean; commit the release first"

git fetch origin main --tags
git merge-base --is-ancestor origin/main HEAD || fail "local main does not contain origin/main"
git rev-parse "$VERSION" >/dev/null 2>&1 && fail "tag already exists locally: $VERSION"
git ls-remote --exit-code --tags origin "refs/tags/$VERSION" >/dev/null 2>&1 && fail "tag already exists on origin: $VERSION"
gh release view "$VERSION" --repo "$REPO" >/dev/null 2>&1 && fail "GitHub release already exists: $VERSION"

VERSION_NO_V="${VERSION#v}"
MAKE_VERSION="$(awk '$1 == "VERSION" {print $3; exit}' Makefile)"
GO_VERSION="$(sed -n 's/^[[:space:]]*Version[[:space:]]*=[[:space:]]*"\([^"]*\)"/\1/p' config.go)"
MANIFEST_VERSION="$(jq -r '.version' dxt/manifest.json)"

[ "$MAKE_VERSION" = "$VERSION" ] || fail "Makefile says $MAKE_VERSION, expected $VERSION"
[ "$GO_VERSION" = "$VERSION_NO_V" ] || fail "config.go says $GO_VERSION, expected $VERSION_NO_V"
[ "$MANIFEST_VERSION" = "$VERSION_NO_V" ] || fail "manifest says $MANIFEST_VERSION, expected $VERSION_NO_V"

echo "==> Signing readiness"
make doctor-signing

echo "==> Tests"
go test -race -shuffle=on ./...
go vet ./...

echo "==> Clean signed build"
make clean
make pack-dxt-signed DEVELOPER_ID="$DEVELOPER_ID" KEYCHAIN_PROFILE="$KEYCHAIN_PROFILE"

echo "==> Release assets"
./scripts/package-release.sh "$VERSION"
REQUIRE_NOTARIZED=true ./scripts/verify-release.sh "$VERSION"

NOTES_FILE="$(mktemp)"
cleanup() { rm -f "$NOTES_FILE"; }
trap cleanup EXIT

{
    echo "# Slide MCP Server $VERSION"
    echo
    awk -v version="$VERSION" '
        $0 ~ "^## .* - " version {printing=1; next}
        printing && /^## / {exit}
        printing {print}
    ' CHANGELOG.md
    echo
    echo "## Install"
    echo
    echo "Download **slide-mcp-server.mcpb** and drag it into Claude Desktop's Extensions screen."
    echo "The stable download name is preserved, and legacy linux-x64/macos-x64 installer assets are included."
    echo
    echo "Private sideloaded MCPB installs still require reinstalling the new bundle; automatic extension updates are available only through Claude's official extension directory."
    echo
    echo "The macOS binary is universal, Developer ID signed, and notarized by Apple."
} > "$NOTES_FILE"

echo "==> Tag and publish"
git tag -a "$VERSION" -m "Release $VERSION"
git push origin main
git push origin "$VERSION"

RELEASE_ASSETS=()
while IFS= read -r asset; do
    RELEASE_ASSETS+=("$asset")
done < "$BUILD_DIR/release-assets.txt"
gh release create "$VERSION" \
    --repo "$REPO" \
    --title "Slide MCP Server $VERSION" \
    --notes-file "$NOTES_FILE" \
    --verify-tag \
    --latest \
    "${RELEASE_ASSETS[@]}"

echo "Release published: https://github.com/$REPO/releases/tag/$VERSION"
