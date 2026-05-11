# AGENTS.md

Repository instructions for AI coding assistants (Claude Code, Cursor,
Codex, etc). Humans, see [README.md](README.md) for the regular intro.

## What this repo is

`slide-mcp-server` is a Go binary that exposes the [Slide
API](https://api.slide.tech/openapi.json) as a Model Context Protocol
server. The headline distribution artifact is
`build/slide-mcp-server.mcpb` - a single signed + notarized Claude
Desktop Extension bundle that contains platform binaries for macOS
(universal Apple Silicon + Intel), Linux (amd64 + arm64), and
Windows (amd64).

## Critical conventions

### 1. Drag-and-drop is the headline install path

The `.mcpb` is the front door for users. `make pack-dxt-signed`
produces it; `make verify-dxt` confirms it works. Before suggesting
any other install flow (CLI wizard, separate installer app, etc.),
make sure your suggestion makes the drag-drop experience better, not
worse. The retired Fyne installer in the deleted `Installer/` dir
is the cautionary tale.

### 2. macOS signing + notarization is non-trivial - read the doc

**[docs/SIGNING.md](docs/SIGNING.md) is required reading** before you
touch anything in:

- [`Makefile`](Makefile) (signing / pack-dxt / release targets)
- [`scripts/setup-signing.sh`](scripts/setup-signing.sh) (gitignored
  per-machine setup script)
- [`scripts/verify-dxt.sh`](scripts/verify-dxt.sh)
- [`scripts/release-all-in-one.sh`](scripts/release-all-in-one.sh)
- [`dxt/manifest.json`](dxt/manifest.json)

That doc is the canonical source of truth for the 10+ macOS-specific
quirks we hit getting the pipeline working. Do **not** try the
"obvious" approach (`.p12` imports, `spctl --type execute` checks,
`add-trusted-cert -r trustAsRoot`, etc.) without first reading why
those don't work. Every quirk listed is something we burned hours on.

The 30-second sniff test for "is the bundle release-grade?":

```bash
source scripts/release-env.sh
make pack-dxt-signed
make verify-dxt
```

`make verify-dxt` should print:

```
codesign:     signed and verified
chain:        Developer ID Application: ... (TEAMID)
notarization: notarized by Apple (verified online)
stapled:      no (Apple can't staple raw Mach-O; online verification works)
```

`stapled: no` is **expected and correct** for raw Mach-O - see
SIGNING.md section 3.6. Do not "fix" it.

### 3. Tool architecture: meta-tools, not one-tool-per-endpoint

The server exposes ~13 meta-tools (`slide_devices`, `slide_agents`,
etc.), each with an `operation` parameter that selects the action. Do
not add new top-level tools when you can add a new operation to an
existing meta-tool. See [registry.go](registry.go) and the
`tools_*.go` files for the pattern. Conditional `allOf` / `if` /
`then` / `required` blocks in each tool's schema enforce per-operation
parameter requirements.

### 4. SDK migration is done, do not undo it

The server is built on
[`github.com/mark3labs/mcp-go`](https://pkg.go.dev/github.com/mark3labs/mcp-go).
The pre-v3.0.0 hand-rolled `bufio.Scanner` JSON-RPC loop is gone.
[`registry.go`](registry.go) converts the legacy `ToolInfo`
descriptors via `mcp.NewToolWithRawSchema` to preserve the rich JSON
schemas without rewriting them. Don't replace `mcp.NewToolWithRawSchema`
with structured `mcp.NewTool` builders unless you've checked that the
allOf/if/then conditional schemas survive.

### 5. Slide API version

We track Slide API v1.27.0 (refresh
[`docs/openapi.json`](docs/openapi.json) from
`https://api.slide.tech/openapi.json` when bumping). New API surfaces
should be added under existing meta-tools as new operations; new
client wrappers go in [`api_v127.go`](api_v127.go) (or
[`api.go`](api.go) for older endpoints).

### 6. The `--api-key` flag, `$SLIDE_API_KEY` env, and `.mcpb`'s `user_config.api_key` all exist

The Slide API token can come from three places:

1. CLI: `slide-mcp-server --api-key tk_...` (highest precedence)
2. Env: `SLIDE_API_KEY=tk_... slide-mcp-server`
3. Claude Desktop's `user_config.api_key` UI (templated into env via
   the `.mcpb` manifest)

OS keychain support is **not** wired in (was a v3.2.0 stretch goal we
deferred because the `.mcpb` `user_config` UI handles secret storage
for the drag-drop audience, and CLI users have shell rc / CI vault /
Docker secret patterns already). Don't add it without first checking
whether the new use case is actually distinct from these three.

### 7. Permission tiers are real - respect them

`config.go` defines four `--tools` modes: `reporting` (read-only),
`restores` (read + restore/VM/network), `full-safe` (default,
everything except dangerous ops), `full` (unrestricted). Every new
tool operation must be classified in `isReadOperation` /
`isRestoreManagementOperation` / `isDangerousOperation`. New
destructive ops (delete-style) default to `full` only.

### 8. Tests + smoke

Before any meaningful change, run:

```bash
go test ./...
./smoke_test.sh        # needs a working SLIDE_API_KEY in .env
make verify-dxt        # if you touched anything build/manifest related
```

The smoke test correctly distinguishes pure-code paths from live API
paths failing on a stale token (skips with a warning rather than
failing).

## What lives where

| Path | Purpose |
|---|---|
| [`main.go`](main.go), [`server.go`](server.go), [`registry.go`](registry.go), [`mcp.go`](mcp.go), [`base_handler.go`](base_handler.go) | SDK wiring, tool registration, request dispatch |
| [`tools_*.go`](.) | Per-tool descriptors + operation handlers |
| [`api.go`](api.go), [`api_v127.go`](api_v127.go) | HTTP client wrappers for the Slide API |
| [`config.go`](config.go) | Server config + permission tier logic |
| [`server_test.go`](server_test.go) | In-process tests via SDK's `HandleMessage` |
| [`smoke_test.sh`](smoke_test.sh) | Live-API smoke covering tools + handshake |
| [`Makefile`](Makefile) | Build / pack-dxt / sign / release targets |
| [`dxt/`](dxt/) | `.mcpb` manifest, icons, packaging assets |
| [`docs/SIGNING.md`](docs/SIGNING.md) | **Required reading for any signing/release work** |
| [`docs/MODERNIZATION_NOTES.md`](docs/MODERNIZATION_NOTES.md) | Per-release migration writeups |
| [`docs/openapi.json`](docs/openapi.json) | Cached Slide API spec (refresh from upstream when bumping) |
| [`scripts/setup-signing.sh`](scripts/setup-signing.sh) | Personal, gitignored, per-machine signing setup |
| [`scripts/release-env.sh`](scripts/release-env.sh) | Personal, gitignored, sourced before signed builds |
| [`scripts/verify-dxt.sh`](scripts/verify-dxt.sh) | `.mcpb` validation (called by `make verify-dxt`) |
| [`scripts/release-all-in-one.sh`](scripts/release-all-in-one.sh) | Production release pipeline (sign + notarize + upload) |
| [`CHANGELOG.md`](CHANGELOG.md) | What changed between releases |

## Common tasks - don't reinvent

- **"How do I add a Slide API endpoint that's new in v1.28?"**
  Add a struct + client method to `api_v127.go`, add an operation +
  handler to the relevant `tools_*.go` file, classify in `config.go`,
  add an entry to the operation enum + `allOf` conditional schema,
  add a smoke test line, add a CHANGELOG entry.
- **"How do I cut a release?"**
  `./scripts/release-all-in-one.sh vX.Y.Z` (needs Apple credentials;
  see SIGNING.md). It signs, notarizes, packs the `.mcpb`, generates
  release notes, and uploads to GitHub Releases. The unversioned
  `slide-mcp-server.mcpb` URL stays stable across releases.
- **"How do I add macOS keychain support for the API key?"**
  See conventions item 6 above. Probably don't, but if you do: add
  `github.com/zalando/go-keyring`, resolve key in order
  flag > env > keychain, never write the key to disk yourself.
- **"How do I sign + notarize on a fresh Mac?"**
  `./scripts/setup-signing.sh`. Read [docs/SIGNING.md](docs/SIGNING.md)
  before touching the script.
