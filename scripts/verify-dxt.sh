#!/usr/bin/env bash
# Lightweight verifier for build/slide-mcp-server.mcpb.
#
# Confirms:
#   1. The .mcpb file exists and is a valid zip.
#   2. manifest.json parses and has the required fields (manifest_version 0.3
#      schema basics).
#   3. Every binary referenced by mcp_config.command (and platform_overrides)
#      exists inside the bundle.
#   4. The binary for the current host's OS+arch can run --version and prints
#      the expected version string.
#   5. On macOS, the universal binary is signed (codesign --verify) and ideally
#      Gatekeeper-approved (spctl --assess); the spctl check is reported but
#      not fatal because notarization may have been skipped.

set -euo pipefail

MCPB="${1:-build/slide-mcp-server.mcpb}"

if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: 'jq' is required (brew install jq)." >&2
    exit 1
fi

if [ ! -f "$MCPB" ]; then
    echo "ERROR: $MCPB not found. Run 'make pack-dxt' first." >&2
    exit 1
fi

WORK="$(mktemp -d)"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

echo "==> Unpacking $MCPB"
unzip -q "$MCPB" -d "$WORK"

echo "==> Validating manifest.json"
MANIFEST="$WORK/manifest.json"
if [ ! -f "$MANIFEST" ]; then
    echo "FAIL: manifest.json missing inside the .mcpb" >&2
    exit 1
fi

if ! jq -e . "$MANIFEST" >/dev/null 2>&1; then
    echo "FAIL: manifest.json is not valid JSON" >&2
    exit 1
fi

require_field() {
    local field="$1"
    if ! jq -e "$field" "$MANIFEST" >/dev/null 2>&1; then
        echo "FAIL: manifest is missing required field $field" >&2
        exit 1
    fi
}

require_field '.manifest_version'
require_field '.name'
require_field '.version'
require_field '.description'
require_field '.author.name'
require_field '.server.type'
require_field '.server.entry_point'
require_field '.server.mcp_config.command'

MANIFEST_VERSION="$(jq -r '.manifest_version' "$MANIFEST")"
case "$MANIFEST_VERSION" in
    0.1|0.2|0.3) ;;
    *) echo "FAIL: unexpected manifest_version: $MANIFEST_VERSION" >&2; exit 1 ;;
esac

echo "    manifest_version=$MANIFEST_VERSION"
echo "    name=$(jq -r '.name' "$MANIFEST")"
echo "    version=$(jq -r '.version' "$MANIFEST")"
echo "    server.type=$(jq -r '.server.type' "$MANIFEST")"

echo "==> Checking bundle-relative references exist"
ENTRY_POINT="$(jq -r '.server.entry_point' "$MANIFEST")"
if [ ! -f "$WORK/$ENTRY_POINT" ]; then
    echo "FAIL: manifest entry_point $ENTRY_POINT is not in the bundle" >&2
    exit 1
fi
echo "    ok: $ENTRY_POINT"

# Search the full manifest rather than command fields alone: a launcher may be
# passed as an argument to a host executable such as /bin/sh.
BUNDLE_PATHS="$(jq -r '
    .. | strings
    | select(startswith("${__dirname}/"))
    | sub("^\\$\\{__dirname\\}/"; "")
' "$MANIFEST" | sort -u)"

while IFS= read -r path; do
    [ -z "$path" ] && continue
    if [ ! -f "$WORK/$path" ]; then
        echo "FAIL: manifest references $path but it is not in the bundle" >&2
        exit 1
    fi
    echo "    ok: $path"
done <<EOF
$BUNDLE_PATHS
EOF

echo "==> Checking host commands are explicit"
jq -r '
    [
        .server.mcp_config.command,
        (.server.mcp_config.platform_overrides // {} | to_entries[] | .value.command)
    ]
    | map(select(. != null and (startswith("${__dirname}/") | not)))
    | unique[]
' "$MANIFEST" | while IFS= read -r command; do
    case "$command" in
        /*) echo "    host command: $command" ;;
        *)
            if [ ! -f "$WORK/$command" ]; then
                echo "FAIL: relative command $command is not in the bundle" >&2
                exit 1
            fi
            echo "    ok: $command"
            ;;
    esac
done

echo "==> Verifying the host binary actually runs"
HOST_OS="$(uname -s)"
HOST_ARCH="$(uname -m)"
case "$HOST_OS" in
    Darwin)
        HOST_BIN="$WORK/server/slide-mcp-server-darwin-universal"
        ;;
    Linux)
        case "$HOST_ARCH" in
            x86_64|amd64) HOST_BIN="$WORK/server/slide-mcp-server-linux-amd64" ;;
            aarch64|arm64) HOST_BIN="$WORK/server/slide-mcp-server-linux-arm64" ;;
            *) echo "WARN: unsupported Linux arch $HOST_ARCH; skipping run check"; HOST_BIN="" ;;
        esac
        ;;
    *)
        echo "WARN: cannot exercise the binary on $HOST_OS from this script; skipping run check"
        HOST_BIN=""
        ;;
esac

if [ -n "$HOST_BIN" ]; then
    if [ ! -x "$HOST_BIN" ]; then
        chmod +x "$HOST_BIN"
    fi
    OUT="$("$HOST_BIN" --version 2>&1 || true)"
    EXPECTED_VERSION="$(jq -r '.version' "$MANIFEST")"
    echo "    binary --version: $OUT"
    if ! echo "$OUT" | grep -qF "$EXPECTED_VERSION"; then
        echo "FAIL: binary did not report version $EXPECTED_VERSION" >&2
        exit 1
    fi
fi

if [ "$HOST_OS" = "Darwin" ]; then
    echo "==> Checking macOS code signing on the universal binary"
    UNIV="$WORK/server/slide-mcp-server-darwin-universal"
    if codesign --verify --verbose=2 "$UNIV" >/dev/null 2>&1; then
        echo "    codesign:     signed and verified"
    else
        echo "    codesign:     NOT signed (ok for dev builds; signed builds use 'make pack-dxt-signed')"
    fi

    # Show the signing chain so it's obvious whether this is Developer ID
    # (release-grade) or ad-hoc.
    AUTH=$(codesign -dvv "$UNIV" 2>&1 | grep -E "^Authority=" | head -1 | sed 's/^Authority=//' || true)
    if [ -n "$AUTH" ]; then
        echo "    chain:        $AUTH"
    fi

    # Notarization check. spctl --type execute is for .app bundles and
    # always says "rejected ... does not seem to be an app" on raw
    # Mach-O, which is misleading. The right test is whether the binary
    # satisfies Apple's =notarized requirement (which goes online to
    # check Apple's records the first time).
    if codesign --test-requirement="=notarized" --verify "$UNIV" >/dev/null 2>&1; then
        echo "    notarization: notarized by Apple (verified online)"
    else
        echo "    notarization: NOT notarized (signed builds need 'make pack-dxt-signed' with valid notarytool credentials)"
    fi

    # Optional: stapled ticket for offline Gatekeeper. Apple can't staple
    # to raw Mach-O binaries (Error 73) - that's a known limitation. The
    # binary still works because Gatekeeper falls back to online check.
    if xcrun stapler validate "$UNIV" >/dev/null 2>&1; then
        echo "    stapled:      yes (works offline)"
    else
        echo "    stapled:      no (Apple can't staple raw Mach-O; online verification works)"
    fi
fi

echo "==> verify-dxt: ok"
