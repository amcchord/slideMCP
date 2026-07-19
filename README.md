# Slide MCP Server

`slide-mcp-server` connects AI assistants to the [Slide BCDR API](https://docs.slide.tech/). It is a single Go binary with task-oriented tools for backup health, file recovery, snapshots, alerts, recovery VMs, image exports, DR networks, and account administration.

It supports the two main local-MCP ecosystems:

- Anthropic: a drag-and-drop Claude Desktop MCP Bundle plus Claude Code registration.
- OpenAI: Codex desktop, CLI, and IDE registration through their shared MCP configuration.

The server uses stdio JSON-RPC, has no runtime dependencies, and never writes the Slide API token to this repository.

## Install

### Claude Desktop: drag and drop

1. Download [slide-mcp-server.mcpb](https://github.com/amcchord/slideMCP/releases/latest/download/slide-mcp-server.mcpb).
2. Drag it into **Claude Desktop → Settings → Extensions**.
3. Paste a Slide API token from [console.slide.tech](https://console.slide.tech/) under **My Settings → API Tokens**.

The bundle contains macOS universal, Linux amd64/arm64, and Windows amd64 binaries. Its macOS binary is Developer ID signed and notarized by Apple, so a normal online Gatekeeper check runs without a warning.

The stable download URL always points at the newest GitHub release. The retired pre-MCPB installer also keeps working: releases publish its historical `macos-x64`, `macos-arm64`, `linux-x64`, `linux-arm64`, and `windows-x64` asset names.

Private, sideloaded Claude Desktop bundles do not currently auto-update themselves. Reinstall the stable `.mcpb` to update while the extension is private. Anthropic enables automatic extension updates for extensions distributed through its official directory; the manifest is prepared for that submission path. See [Anthropic's local extension guide](https://support.claude.com/en/articles/10949351-getting-started-with-local-mcp-servers-on-claude-desktop).

### OpenAI Codex

From a clone:

```bash
make build
SLIDE_API_KEY=tk_... make install-codex
codex mcp get slide
```

Or register any downloaded binary directly:

```bash
codex mcp add slide --env SLIDE_API_KEY=tk_... -- /absolute/path/to/slide-mcp-server
```

Codex desktop, CLI, and IDE surfaces share the MCP configuration. Re-running `make install-codex` replaces the existing `slide` registration safely.

### Claude Code

```bash
make build
SLIDE_API_KEY=tk_... make install-claude-code
claude mcp list
```

Direct registration also works:

```bash
claude mcp add slide --env SLIDE_API_KEY=tk_... -- /absolute/path/to/slide-mcp-server
```

### Cursor and other stdio MCP hosts

Download the binary for your platform and configure the equivalent of:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/absolute/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "tk_...",
        "SLIDE_TOOLS": "safe"
      }
    }
  }
}
```

## What to ask

- “Are all my Slide boxes healthy?”
- “Did backups run last night for ACME?”
- “Find Q4-budget.xlsx on Bob's laptop and restore Tuesday's version.”
- “What unresolved alerts do I have? Sort worst first.”
- “Boot a recovery VM for DC-01.”
- “What changed in the last 24 hours?”
- “I don't know where to start—what can you do?”

You do not need to memorize `a_...`, `d_...`, or `c_...` identifiers. User-facing operations accept `name_hint`; deterministic server-side matching resolves display names and hostnames or returns candidates when a name is ambiguous.

## MCP surface

The server exposes 12 task-oriented tools plus one compatibility alias. Each main tool uses an `operation` parameter instead of creating a separate top-level tool for every API endpoint.

| Tool | Purpose |
|---|---|
| `slide_help` | Embedded onboarding, examples, glossary, troubleshooting, and diagnostics |
| `slide_overview` | Complete inventory, health, and client/device summaries |
| `slide_files` | Search, browse, restore, and push files |
| `slide_recovery` | Recovery VMs, image exports, and DR networks |
| `slide_audit` | Account audit log queries |
| `slide_clients` | Client management |
| `slide_admin` | User and account management |
| `slide_devices` | Device details, networking, and appliance actions |
| `slide_agents` | Protected-system settings, services, schedules, and retention |
| `slide_snapshots` | Restore points and verification results |
| `slide_backups` | Backup status and backup launch |
| `slide_alerts` | Alert review, triage, and resolution |
| `list_all_clients_devices_and_agents` | Backward-compatible inventory alias |

It also serves six MCP Prompts—including `slide.welcome`, `slide.daily-status`, and `slide.restore-file`—and static/dynamic `slide://` Resources.

## Safety and reliability

The default `safe` permission tier allows operational recovery work but blocks deletes, poweroff, and reboot. Available modes are:

| Mode | Allowed operations |
|---|---|
| `read-only` | List, get, search, browse, and diagnostic operations |
| `safe` | Read operations plus restores, backups, and reversible settings |
| `full` | Everything, including delete and power-cycle operations |

Legacy names `reporting`, `restores`, and `full-safe` remain accepted so existing configurations continue to start.

Reliability protections include:

- startup validation that never blocks the MCP initialize handshake;
- 45-second operation deadlines and bounded upstream response bodies;
- retry/backoff only for idempotent reads, avoiding duplicate mutations;
- complete, loop-safe pagination for account-wide inventory and health;
- typed, token-redacted API errors with actionable remediation;
- server-side JSON Schema validation for inputs and structured outputs;
- `--doctor` and secret-masked `--debug` diagnostics.

## Configuration

CLI flags override environment variables.

| Flag | Environment | Default |
|---|---|---|
| `--api-key` | `SLIDE_API_KEY` | required for normal operation |
| `--base-url` | `SLIDE_BASE_URL` | `https://api.slide.tech` |
| `--tools` | `SLIDE_TOOLS` | `safe` |
| `--disabled-tools` | `SLIDE_DISABLED_TOOLS` | none |
| `--doctor` | — | run checks and exit |
| `--debug` | — | print a masked diagnostic bundle and exit |
| `--skip-startup-validation` | — | skip the background account probe |
| `--tool` / `--args` | — | execute one tool call without an MCP host |
| `--version` | — | print the version and exit |

Non-loopback API base URLs must use HTTPS. Plain HTTP is accepted only for localhost test servers.

## Test harness

The repository has three testing layers:

```bash
# Deterministic offline suite: API failure injection, permissions, schemas,
# MCP protocol profiles, distribution contracts, and a real stdio subprocess.
make test

# Live, read-only demo/account smoke. Loads ignored .env when present and
# always rebuilds the repository binary before testing.
./smoke_test.sh

# Cross-platform binaries, MCPB structure, checksums, compatibility aliases.
make release
```

CI runs race-enabled Go tests, `go vet`, cross-compilation, ShellCheck, release-asset verification, and the official Anthropic MCPB manifest validator. Protocol profiles cover Claude Desktop's established `2025-06-18` path and Codex/current MCP's `2025-11-25` path.

For interactive inspection:

```bash
SLIDE_API_KEY=tk_... npx @modelcontextprotocol/inspector ./build/slide-mcp-server
```

## Develop

```bash
git clone https://github.com/amcchord/slideMCP.git
cd slideMCP
make setup-dev
make test
make build
```

The API client tracks Slide API v1.27.0; the cached OpenAPI document is [docs/openapi.json](docs/openapi.json). Add new endpoint behavior as an operation on the closest existing meta-tool unless the workflow is better represented by an MCP Prompt or Resource.

Useful diagnostics:

```bash
SLIDE_API_KEY=tk_... build/slide-mcp-server --doctor
build/slide-mcp-server --debug | jq .
```

## Release

Read [docs/SIGNING.md](docs/SIGNING.md) before touching the signing pipeline. A production release is intentionally cut from an already-committed, clean `main` branch:

```bash
./scripts/setup-signing.sh       # first machine only
make doctor-signing
./scripts/release-all-in-one.sh vX.Y.Z
```

The release pipeline:

1. verifies `Makefile`, `config.go`, and `dxt/manifest.json` versions agree;
2. runs race tests and `go vet`;
3. builds every target from a clean `build/` directory;
4. Developer ID signs and notarizes the universal macOS executable;
5. packs and verifies the `.mcpb`;
6. builds current and legacy installer asset names plus SHA-256 checksums;
7. enforces signing/notarization again before tagging;
8. pushes `main`, creates the tag, and publishes the GitHub release.

Apple cannot staple a ticket to a raw Mach-O executable. `stapled: no` is expected; `codesign --test-requirement="=notarized"` performs the correct online notarization check.

## License

[MIT](LICENSE)
