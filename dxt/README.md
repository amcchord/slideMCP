# Slide MCP - Claude Desktop Extension (.mcpb)

This directory holds everything that makes up
`build/slide-mcp-server.mcpb`, the drag-and-drop install bundle for
Claude Desktop.

## Layout

```
dxt/
  manifest.json         MCPB spec 0.3 manifest for the bundle
  icon.png              256x256 default icon
  icons/                Light/dark icons at 16, 32, 128, 256 px
  icon-source-1024.png  Original 1024x1024 source (used by `sips` to regen icons)
  README.md             This file
```

## Building the bundle

From the repo root:

```bash
# Unsigned, fast, for dev iteration
make pack-dxt

# Signed + notarized, for release. Requires:
#   - DEVELOPER_ID  -> "Developer ID Application: <Name> (<TEAMID>)"
#   - KEYCHAIN_PROFILE -> a notarytool keychain profile created via
#     `xcrun notarytool store-credentials`
make pack-dxt-signed \
     DEVELOPER_ID='Developer ID Application: Your Name (TEAMID)' \
     KEYCHAIN_PROFILE='YourProfile'
```

Both targets produce `build/slide-mcp-server.mcpb`. The bundle contains:

- `manifest.json` (this file's `manifest.json`)
- `icon.png` and `icons/`
- `server/slide-mcp-server-darwin-universal` (lipo'd Apple Silicon + Intel)
- `server/slide-mcp-server-linux-amd64`
- `server/slide-mcp-server-linux-arm64`
- `server/slide-mcp-server-linux` (selects the correct Linux architecture)
- `server/slide-mcp-server-windows-amd64.exe`

The manifest selects by OS; the Linux launcher then selects amd64 or arm64.
macOS uses one universal Mach-O containing both Intel and Apple Silicon slices.

## Installing locally

1. Run `make pack-dxt`.
2. Open Claude Desktop -> Settings -> Extensions.
3. Drag `build/slide-mcp-server.mcpb` onto the Extensions window.
4. Paste your Slide API token when prompted.

To pick up code changes after iterating, rebuild, remove the existing private
extension from Claude Desktop's UI, and drag the new bundle in. Private
sideloaded MCPBs do not auto-update; official-directory extensions do.

## Verifying the bundle

```bash
make verify-dxt
```

Runs `scripts/verify-dxt.sh`, which:

1. Confirms the .mcpb is a valid zip.
2. Validates `manifest.json` has the required v0.3 fields and parses.
3. Confirms every binary referenced by `command` / `platform_overrides`
   exists in the bundle.
4. Runs the binary appropriate for the current host with `--version` and
   asserts the version string matches the manifest.
5. On macOS, runs `codesign --verify` and the correct online raw-Mach-O
   notarization requirement. `spctl --type execute` is intentionally not used:
   it assesses app bundles and misleadingly rejects raw executables.

## User config exposed in Claude Desktop's UI

| Field        | Required | Notes                                                                       |
|--------------|----------|-----------------------------------------------------------------------------|
| `api_key`    | yes      | Sensitive. Generated at console.slide.tech > My Settings > API Tokens.      |
| `tools_mode` | no       | One of `read-only`, `safe` (default), or `full`; legacy names still work.   |

`SLIDE_BASE_URL` is intentionally not exposed in the UI; power users
who need to point at a staging API can set the env var directly via
their host config.

## Regenerating the icon set

The icons live in [icons/](icons/) at 16/32/128/256 in both light and
dark variants. They're generated from `icon-source-1024.png` with macOS's
built-in `sips`:

```bash
cd dxt
for size in 16 32 128 256; do
    sips -Z $size icon-source-1024.png --out icons/icon-$size-light.png
    cp icons/icon-$size-light.png icons/icon-$size-dark.png
done
sips -Z 256 icon-source-1024.png --out icon.png
```

(Light and dark are identical right now - the source is a colored shield
that reads well on both backgrounds. If you want true light/dark variants
later, replace one of the two passes.)
