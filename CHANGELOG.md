# Slide MCP Changes

## 2026-05-10 - v3.1.0 - Drag-and-drop install + retire Fyne installer

### Headline

The recommended install path for Claude Desktop is now a single
`slide-mcp-server.mcpb` file you drag onto Claude Desktop's Extensions
screen. The bundle is signed and notarized; Claude Desktop prompts for
your Slide API token, stores it in its own per-user state, and selects
the right binary for your OS at runtime.

### Build

- One universal `.mcpb` ([dxt/manifest.json](dxt/manifest.json),
  manifest_version `0.3`) that contains:
  - macOS universal binary (lipo'd Apple Silicon + Intel)
  - Linux amd64 + arm64
  - Windows amd64
  Selected at runtime via the manifest's `platform_overrides`. ~17 MB.
- New [`Makefile`](Makefile) targets:
  - `build-universal-darwin` - lipo'd Mac binary
  - `pack-dxt` - unsigned bundle for dev iteration
  - `pack-dxt-signed` - signed + notarized bundle for release
    (requires `DEVELOPER_ID` + `KEYCHAIN_PROFILE`)
  - `verify-dxt` - validates the manifest and runs the host binary
  - `notarize-darwin-universal` - signs + notarizes + staples the
    universal binary that ships inside the .mcpb
- File targets for every per-platform binary so `make` only rebuilds
  what's stale.
- Polished `.mcpb` metadata: 4-size + light/dark icon set
  ([dxt/icons/](dxt/icons/)), repository / license / privacy URLs,
  long_description covering the v1.27.0 feature surface, softer
  `api_key` prompt explaining what the token authorizes.

### API key handling

- The `.mcpb` flow uses Claude Desktop's `user_config.api_key` UI
  (`sensitive: true, required: true`). Claude Desktop owns the secret;
  the MCP never writes the API key to `claude_desktop_config.json` itself.
- The `--api-key` flag and `SLIDE_API_KEY` env var continue to work for
  Claude Code, Cursor, CI, and Docker. Modern hosts (Claude Code, Cursor)
  also support `${SLIDE_API_KEY}` expansion in their JSON configs so the
  literal token doesn't have to live in the file.
- Native OS keychain support is queued as a v3.2.0 stretch goal; not
  needed for the drag-drop UX so it's deferred.

### Distribution

- `scripts/release-all-in-one.sh` now signs + notarizes the universal
  binary, builds the .mcpb, and uploads it as the headline release asset.
  The README links to the unversioned
  `releases/latest/download/slide-mcp-server.mcpb` URL so the link never
  has to change per release.
- Per-platform tarballs (including a new `darwin-universal.tar.gz`) are
  still published for `claude mcp add` / Cursor / CI users.

### Retired

- The entire `Installer/` directory (Fyne GUI app, ~660 LOC, separate
  `go.mod`, CGO/OpenGL build deps, MCP-Installer.png + icon.go, separate
  signing scripts, `releases/`). Its only differentiator was Claude
  Desktop install, which the .mcpb does better and with no extra app to
  ship. The icon source lives at `dxt/icon-source-1024.png` for posterity.
- `Makefile` targets `build-installer`, `build-installer-all`, and
  `build-complete` are gone.
- README "GUI Installer (Recommended)" section + screenshots replaced by
  the new "⚡ Install" section at the top.

### Documentation

- New [`AGENTS.md`](AGENTS.md) at the repo root. Repository-level
  instructions for AI coding assistants (Claude Code, Cursor, Codex,
  etc.) using the universal `AGENTS.md` convention. Documents the
  meta-tools architecture, drag-drop install philosophy, permission
  tiers, and links to the canonical reference docs.
- New [`docs/SIGNING.md`](docs/SIGNING.md) - the canonical reference for
  the macOS sign + notarize + verify pipeline. Documents all 10+
  macOS quirks that bit us (Apple's DER `.cer` vs PEM, OpenSSL 3.x
  PKCS#8 vs `security`'s PKCS#1, `.p12` rejection on modern macOS,
  the `add-trusted-cert -r trustAsRoot` chain-validation trap, the
  stapler Error 73 limitation on raw Mach-O, the misleading
  `spctl --type execute` rejection, app-specific password expiry,
  Make subshell env-var inheritance, Makefile quoting). Includes a
  troubleshooting matrix and a known-good cert-rotation procedure.
- [`docs/CODE_SIGNING.md`](docs/CODE_SIGNING.md) is now a thin redirect
  to `SIGNING.md`.

### Code signing helper

- New `scripts/setup-signing.sh` (personal, gitignored automatically by
  the existing `**/setup-signing.sh` rule). Idempotent walkthrough that:
  - Detects any existing `Developer ID Application` identity and skips
    the cert flow if present.
  - Otherwise generates a 2048-bit RSA key + CSR locally, opens the
    Apple Developer portal in the browser, waits for the resulting
    `.cer` to land in `~/Downloads/`, and imports both into the login
    keychain via a temporary password-less `.p12`.
  - Reads `APPLE_ID` + `APP_SPECIFIC_PASSWORD` from
    `scripts/release-env.sh` if it exists; otherwise prompts (with
    masked input for the password).
  - Auto-detects the Team ID from the installed identity.
  - Creates a `slide-mcp-release` notarytool keychain profile.
  - Atomically rewrites `scripts/release-env.sh` (mode 600, gitignored)
    with `DEVELOPER_ID`, `KEYCHAIN_PROFILE`, `APPLE_TEAM_ID`,
    `APPLE_ID`, `APP_SPECIFIC_PASSWORD`.
  - Smoke-signs a throwaway binary as proof, then offers to run
    `make pack-dxt-signed` immediately.
- Flags: `--check` (read-only sanity check), `--force` (re-create
  notarytool profile), `--cleanup` (escape hatch).
- New `make doctor-signing` target that runs `setup-signing.sh --check`.
  Also called as a non-fatal section of `make doctor`.
- `make sign-macos` and `make notarize-darwin-universal` errors now
  point at `source scripts/release-env.sh` and
  `./scripts/setup-signing.sh` instead of just printing CLI args.

### Files touched

- Edited: [`Makefile`](Makefile) (rewritten as drag-drop centric;
  doctor-signing target; friendlier signing-prereq errors),
  [`dxt/manifest.json`](dxt/manifest.json) (multi-platform overrides +
  metadata polish), [`dxt/README.md`](dxt/README.md),
  [`scripts/release-all-in-one.sh`](scripts/release-all-in-one.sh)
  (signs+notarizes the universal binary, packs and uploads .mcpb),
  [`README.md`](README.md) (rewritten install section, retired GUI
  installer, new code-signing setup subsection),
  [`config.go`](config.go) (`Version = 3.1.0`).
- New: `dxt/icons/icon-{16,32,128,256}-{light,dark}.png`,
  `dxt/icon.png` (256), `dxt/icon-source-1024.png` (preserved from the
  retired Installer), [`scripts/verify-dxt.sh`](scripts/verify-dxt.sh),
  `scripts/setup-signing.sh` (gitignored).
- Deleted: `Installer/` (entire directory).

## 2026-05-10 - v3.0.0 - SDK migration + Slide API v1.27.0 coverage

### Architecture

- **Migrated to the official Go MCP SDK** ([`github.com/mark3labs/mcp-go v0.52.0`](https://pkg.go.dev/github.com/mark3labs/mcp-go)).
  - The hand-rolled `bufio.Scanner` JSON-RPC loop in the previous `server.go`
    is gone; we now wire tools via `server.NewMCPServer` + `server.ServeStdio`.
  - Each `getXxxToolInfo()` descriptor is converted to an `mcp.Tool` via
    `mcp.NewToolWithRawSchema` so the existing rich JSON schemas (with
    `allOf` / `if` / `then` / `required`) are preserved untouched.
  - Tool handlers continue to use the same `func(map[string]interface{}) (string, error)`
    signature; they are wrapped by `adaptToolHandler` for the SDK.
  - The mode-aware permission filter is now an `mcp.ToolFilterFunc` so
    `tools/list` always reflects the live `--tools` configuration.
- **Initial context is now an MCP resource** at
  `slide://context/clients-devices-agents`. The non-spec `initialContext`
  field we used to stuff into the `initialize` response is removed.
- Latest MCP protocol version (`2025-06-18`+) is advertised by the SDK.
- The `--tool` / `--args` one-shot CLI mode is preserved (calls handlers
  directly without going through the SDK transport), so existing
  `scripts/show_mcp_context.sh` style integrations keep working.

### Slide API v1.27.0 surface coverage

New operations added without growing the tool count:

- `slide_agents`: `list_services`, `update_services`, `set_schedule`,
  `clear_schedule`, `pause_backups`, `resume_backups`, `set_retention`,
  `set_restore_defaults`, `set_volumes`, `set_file_index_enabled`,
  `set_timezone`, `set_comments`, `update_alert_config`.
- `slide_devices`: `get_network`, `update_network`, `list_vlans`, `get_vlan`,
  `create_vlan`, `update_vlan`, `delete_vlan` (plus list/get responses now
  expose `device_warranty_expiration_date` and `network_update_pending`).
- `slide_snapshots`: `get_service_verification` (and list/get responses now
  expose `verify_service_status`).
- `slide_user_management`: `get_user_avatar`.
- `slide_networks`: new `vlan_tag` parameter on `create` / `update` for
  bridge-lan networks bound to a specific VLAN tag on the bridge device.

The `Agent` and `Device` Go structs now decode every new field from
v1.27.0 (`backup_schedule`, `local_retention_policy`, `default_restore_settings`,
`volumes`, `volumes_include_default`, `alert_configs`, `comments`, `timezone`,
`file_index_enabled`, `backup_paused_indefinite`, `backup_paused_until`,
`network_update_pending`, etc).

### Permissions

`config.go` now classifies the new operations:

- Read-only (always allowed): `list_services`, `get_network`, `list_vlans`,
  `get_vlan`, `get_service_verification`, `get_user_avatar`.
- Restores-mode safe writes: schedule, retention, restore defaults, alert
  config, agent service updates, device network update, VLAN create/update.
- Reserved for `--tools full`: `delete_vlan` (can break VPN/port-forward
  routes that reference VLAN IPs).

### Distribution & install ergonomics

- `scripts/install-claude-code.sh` and `make install-claude-code` register
  the server with Claude Code via `claude mcp add` (idempotent).
- `scripts/install-claude-desktop.sh` and `make install-claude-desktop`
  merge a `slide` entry into Claude Desktop's `claude_desktop_config.json`
  with a one-time backup; safe to re-run.
- New `dxt/manifest.json` (MCPB spec 0.3) plus `make pack-dxt` produces a
  Claude Desktop Extension bundle (`build/slide-mcp-server-v3.0.0.mcpb`)
  with `api_key`, `tools_mode`, and `base_url` user-config fields.
- `scripts/setup-dev.sh` and `make setup-dev` idempotently install Go,
  jq, and gh via Homebrew on a fresh dev box. `make doctor` runs the same
  setup, then `go vet`, then a build smoke.
- Version bumped to `v3.0.0` everywhere (`Makefile`, `Installer/Makefile`,
  `scripts/build-and-sign.sh`, `config.go`, `dxt/manifest.json`).

### Tests

- `server_test.go` adds in-process tests using the SDK's `HandleMessage`:
  - `TestToolsListContents` asserts every expected tool name and v1.27.0
    operation is registered.
  - `TestToolFilterRespectsMode` confirms the mode filter hides tools that
    require explicit enable.
  - `TestOneShotPermissionDenied` exercises tool-level disablement and
    unknown-tool dispatch through the one-shot CLI mode.
- `smoke_test.sh` rewritten to drive the SDK server via the `--tool` mode
  plus a real stdio handshake. Live API failures from a stale token are
  reported as SKIP rather than FAIL.

### Breaking changes

- The `initialize` response no longer contains a non-standard
  `initialContext` field. Read the `slide://context/clients-devices-agents`
  resource (or call `slide_meta` with `list_all_clients_devices_and_agents`).
- The `--exit-after-first` flag is removed; `--tool` / `--args` is the
  supported way to run a single tool and exit.

## 2025-01-17 - v2.6.0 - Push Restore Support

### Added
- **Push Restore Functionality**: New operations to restore files directly to protected systems
  - `list_pushes` - List push operations for a file restore
  - `create_push` - Create a push operation to restore files to protected system
  - `update_push` - Update push operation state (cancel operations)
- Push restore destination folder validation (must be `X:\SlideRestore` where X is drive letter)
- Enhanced metadata and guidance for push restore operations
- State tracking for push operations (created, in_progress, completed, failed, canceled)

### Changed
- **Snapshot Location Default**: Updated default snapshot listing to use `location_any` filter
  - Now shows snapshots from all locations (local, cloud, and deleted) by default
  - Added `location_any` to the `snapshot_location` enum in tool definitions
- Updated `slide_restores` tool description to include push restore functionality
- Updated permissions in config.go to include push restore operations

### Technical Details
- Added `FileRestorePush` struct with fields: file_restore_push_id, file_restore_id, start_time, end_time, state, source_file_path, destination_folder, size
- Implemented API functions: `listFileRestorePushes()`, `createFileRestorePush()`, `updateFileRestorePush()`
- Added strict validation for destination folder format with helpful error messages
- Updated tools_restores.go with push operations and parameter definitions
- Updated tools_snapshots.go to include location_any in enum

## 2025-01-10 - Documentation Access Tool

### Added
- New `slide_docs` tool for accessing Slide documentation
  - `list_sections` - Browse available documentation categories
  - `get_topics` - View topics within a specific section
  - `search_docs` - Search across all documentation
  - `get_content` - Get detailed content for a specific topic
  - `get_api_reference` - Quick access to API reference information
- Added `tools_docs.go` implementing documentation access functionality
- Created `DOCS_TOOL_README.md` explaining the new tool usage

### Purpose
The slide_docs tool allows LLMs to reference official Slide documentation from docs.slide.tech when answering questions about Slide backup and recovery features. This provides contextual help, best practices, troubleshooting guidance, and API reference information directly within the MCP agent.

## 2025-01-09 - Meta Tool Performance Improvements 