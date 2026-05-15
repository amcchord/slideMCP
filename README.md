# Slide MCP Server

An MCP server implementation that integrates with the Slide API, designed
around the questions an MSP technician actually asks Claude Desktop -
"are all my Slide boxes healthy?", "did backups run last night for ACME?",
"find Q4-budget.xlsx on Bob's laptop and restore Tuesday's version".

## What to ask first

If you've just installed the extension, try one of these prompts. Claude
will reach for `slide-mcp-server` automatically - you don't have to spell
out "use the Slide MCP".

- "Are all my Slide boxes healthy?"
- "Did backups run last night for ACME?"
- "Find Q4-budget.xlsx on Bob's laptop."
- "What unresolved alerts do I have? Sort worst first."
- "Boot a recovery VM for DC-01."
- "What changed in the last 24 hours?"
- "I don't know where to start - what can you do?"

You never need to know `a_xxxxxxxxxxxx` / `d_xxxxxxxxxxxx` IDs. Say
"Bob's laptop" or "the file server" - the server fuzzy-matches against
hostnames and display names and asks you to disambiguate when more than
one entity matches.

When in doubt: `/slide.welcome` (slash-command in Claude Desktop) or
just ask "what can you do?".

## What's new in v5.0.0

- **Novice-first overhaul**. The LLM now reaches for slide-mcp-server
  whenever the user mentions Slide, BCDR, backups, restores, snapshots,
  recovery VMs, file recovery, disaster recovery, audit logs, or
  unresolved alerts - without the user having to spell it out.
- **`slide_help`** - new always-available tool with `getting_started`,
  `examples`, `glossary`, `troubleshoot`, `what_can_you_do` operations.
- **`/slide.welcome`** slash-command prompt + `slide://welcome`,
  `slide://help/glossary`, `slide://help/troubleshoot` resources.
- **Server-side fuzzy name resolution**. Every tool that takes an
  `agent_id` / `device_id` / `client_id` also accepts `name_hint`. Pass
  a hostname, display name, or substring; the server resolves it.
  Ambiguous matches return a structured list of candidates.
- **Startup token validation** with clear, multi-line remediation
  guidance for invalid tokens or firewall issues.
- **`slide-mcp-server --doctor`** subcommand for self-diagnosis
  (token, network, sample reads). Idempotent and CI-friendly.
- **`next_steps`** array appended to every response with curated
  follow-up tool-call suggestions. Suppress with `hints=off`.
- **Friendlier errors**: 401/403/404 carry novice-friendly remediation
  hints, and unknown operations get did-you-mean suggestions.

Full migration table and per-file detail in [CHANGELOG.md](CHANGELOG.md).

## What's new in v4.0.0

- **Headline new capability: `slide_files`**. Search filenames across every
  snapshot for an agent, list snapshot versions of a path, then create a
  restore, browse it, or push files back to the protected system - all
  through one task-oriented tool. Wraps the previously-unexposed Slide
  API v1.27.0 `FileSearch` + `PathVersion` endpoints.
- **`slide_audit`** for compliance / "what changed in the last 24h?" -
  wraps the previously-unexposed `/v1/audit*` endpoint family.
- **Task-oriented tool surface**. New `slide_overview` (inventory +
  health + per-client/per-device summaries), `slide_recovery` (boot
  VMs + export images + DR networks), `slide_clients`, `slide_admin`.
- **5 MCP Prompts** in Claude Desktop's slash-command UI:
  `/slide.daily-status`, `/slide.triage-alerts`, `/slide.restore-file`,
  `/slide.boot-recovery-vm`, `/slide.dr-runbook`.
- **9 MCP Resources** (up from 1) including URI-templated
  `slide://client/{id}`, `slide://device/{id}`, `slide://agent/{id}`,
  plus static `slide://overview/health`, `slide://alerts/unresolved`,
  `slide://audit/recent`, and a live-cached `slide://docs/openapi`.
- **Token-conscious responses**. Every list op defaults to
  `format=summary` (one-line entries) instead of v3's full JSON dump;
  opt up to `compact` or `detailed` when you need more, plus
  `fields=a,b,c` projection.
- **Tool annotations** (`readOnlyHint`/`destructiveHint`/`idempotentHint`)
  so Claude Desktop skips confirmation prompts on safe reads.
- **Reliability**: HTTP client now honours `Retry-After` on 429s,
  retries transient 5xx once, and surfaces structured `APIError`s with
  per-status hints.
- **Permission tiers collapsed** from four to three (`read-only`/`safe`/
  `full`). Legacy names (`reporting`/`restores`/`full-safe`) are
  silently aliased.
- **Retired ~3K LoC** of dead-weight tooling (`slide_docs`,
  `slide_presentation`, `slide_reports`, `slide_meta`) that the LLM
  can compose from raw data + the new prompts.

Full migration table and per-file detail in [CHANGELOG.md](CHANGELOG.md).

## 🚀 Go Binary Implementation ⚡
- **Single binary**: No dependencies, just download and run
- **Fast startup**: ~50ms startup time
- **Low memory usage**: 10-20MB memory footprint
- **Cross-platform**: Linux, macOS (Apple Silicon + Intel), Windows
- **Task-oriented surface**: 11 meta-tools designed around real MSP workflows
- **MCP Resources + Prompts**: pre-baked context and slash-command workflows
  (`slide://overview`, `/slide.daily-status`, etc.)

---

## ⚡ Install

Pick the path for your host. Each is one step.

### 🟢 Claude Desktop - drag and drop (recommended)

1. Download **[slide-mcp-server.mcpb](https://github.com/austinmcchord/slideMCP/releases/latest/download/slide-mcp-server.mcpb)**.
2. Drag it onto Claude Desktop's **Settings -> Extensions** screen.
3. Paste your Slide API token when prompted (generate one at
   [console.slide.tech](https://console.slide.tech/) under
   **My Settings -> API Tokens**).

That's it. Claude Desktop owns the token storage, picks the right binary
for your OS, and starts the server. The bundle is signed and notarized,
so Gatekeeper won't get in the way on macOS.

### 🟦 Claude Code

```bash
claude mcp add slide --env SLIDE_API_KEY=tk_... -- /path/to/slide-mcp-server
```

Or, from a clone of this repo:

```bash
make build
SLIDE_API_KEY=tk_... make install-claude-code
```

The wrapper is idempotent - re-running it overwrites the prior `slide`
registration cleanly.

### 🟪 Cursor / other MCP hosts

Drop the binary somewhere on `$PATH` and add this to `~/.cursor/mcp.json`
(or your host's equivalent):

```json
{
  "mcpServers": {
    "slide": {
      "command": "slide-mcp-server",
      "env": { "SLIDE_API_KEY": "${SLIDE_API_KEY}" }
    }
  }
}
```

Claude Code, Cursor, and most other modern hosts support `${VAR}`
expansion so the literal token never has to live in the config file.

### 🛠️ From source / on a new dev box

```bash
git clone https://github.com/austinmcchord/slideMCP.git && cd slideMCP
make setup-dev   # idempotently installs go, jq, gh via Homebrew
make build       # produces build/slide-mcp-server
make pack-dxt    # produces build/slide-mcp-server.mcpb (unsigned, dev)
```

### 🔏 Code signing setup (release builds only)

`make pack-dxt-signed` (and the release scripts) need a valid Apple
Developer ID Application certificate plus a notarytool keychain profile.
A first-time setup helper handles the whole dance - generates a CSR,
opens the Apple Developer portal in your browser, imports the resulting
certificate, stores notarytool credentials, and writes
`scripts/release-env.sh` (gitignored) so subsequent builds Just Work:

```bash
./scripts/setup-signing.sh    # interactive, idempotent, ~5 minutes
make doctor-signing            # read-only check; run before each release
```

Once setup is done, on every new shell:

```bash
source scripts/release-env.sh   # exports DEVELOPER_ID, KEYCHAIN_PROFILE, etc.
make pack-dxt-signed            # signed + notarized .mcpb in build/
make verify-dxt                 # confirms Gatekeeper accepts the inner binary
```

The setup script is **not tracked in git** (covered by the existing
`**/setup-signing.sh` rule) and never writes plaintext secrets outside
of the local `scripts/release-env.sh` (also gitignored) and your macOS
keychain.

**Read [`docs/SIGNING.md`](docs/SIGNING.md) before touching anything in
the signing pipeline.** It documents 10+ macOS-specific quirks we
already debugged the hard way (DER vs PEM, PKCS#1 vs PKCS#8, why .p12
imports fail on modern macOS, why `add-trusted-cert -r trustAsRoot`
breaks chain validation, why `spctl --type execute` always rejects raw
Mach-O even when notarized, and more). LLMs working on this repo: also
see [`AGENTS.md`](AGENTS.md).

---

For more setup detail (CLI flags, environment variables, manual JSON snippets),
see the installation section further down.

## 🎯 Architecture

**Task-oriented meta-tools**: 11 focused meta-tools designed around the
questions an MSP technician actually asks Claude Desktop, not the raw API
endpoint surface. Each accepts an `operation` parameter that selects the
action; conditional `allOf`/`if`/`then` schemas validate per-operation
required parameters before the call ever leaves Claude.

Plus **MCP Prompts** (slash-command workflows in Claude Desktop) and
**URI-templated Resources** (cheap pre-baked context the LLM can pull
without a tool call).

### Example Usage Pattern

```json
{
  "name": "slide_files",
  "arguments": {
    "operation": "search",
    "agent_id": "a_0123456789ab",
    "search_term": "Q4-budget"
  }
}
```

## Available Meta-Tools (v5.0.0)

### 💡 Help & onboarding

0. **`slide_help`** — discovery / onboarding / troubleshooting
   - Operations: `getting_started`, `examples`, `glossary`, `troubleshoot`,
     `list_prompts`, `list_resources`, `what_can_you_do`
   - Always available, never blocked by tools mode, never disable-able.
   - First thing the LLM should call when the user is vague.

### 🧭 Overview & inventory

1. **`slide_overview`** — top-level inventory + health
   - Operations: `inventory`, `health`, `for_client`, `for_device`
   - One-screen answer to "what do we have?" and "is everything OK?"
   - Read-only; safe to call at the start of any conversation.

### 🔎 Files (the headline v4 capability)

2. **`slide_files`** — search filenames + restore + push back
   - Operations: `search`, `versions`, `list_restores`, `get_restore`,
     `create_restore`, `delete_restore`, `browse`, `list_pushes`,
     `create_push`, `update_push`, `get_push_status`
   - `search` finds a filename across every snapshot for an agent
     (Slide API v1.27.0 `FileSearch`); `versions` lists snapshots that
     contain a path. The whole "find Q4-budget.xlsx and restore Tuesday's
     copy" flow lives in this one tool.

### 🛟 Recovery

3. **`slide_recovery`** — boot VMs, export images, manage DR networks
   - VM operations: `list_vms`, `get_vm`, `boot_vm`, `update_vm`,
     `delete_vm`, `get_rdp_bookmark`
   - Image exports: `list_images`, `get_image`, `export_image`
     (VHD/VHDX/VMDK/QCOW2/RAW), `delete_image`, `browse_image`
   - DR networks: `list_networks`, `get_network`, `create_network`,
     `update_network`, `delete_network` plus IPSec / port-forward /
     WireGuard peer CRUD
   - The "I need to actually recover something" toolkit.

### 📜 Audit (compliance / "what changed?")

4. **`slide_audit`** — account audit log
   - Operations: `list`, `get`, `actions`, `resources`, `recent`
   - `recent` defaults to a 24h window for "what just changed?"

### 🗂️ Snapshots, backups, alerts, agents, devices

5. **`slide_snapshots`** — `list`, `list_deleted`, `get`,
   `get_service_verification`, `recent_for_agent` (one-call "last N days
   of restore points for X").
6. **`slide_backups`** — `list`, `get`, `start`, plus the v4
   `status_for_client` / `status_for_device` / `recent_for_agent` ops
   that answer "did backups run last night for X?" in one call.
7. **`slide_alerts`** — `list`, `get`, `update`, plus v4 `triage` that
   sorts unresolved alerts by severity hint.
8. **`slide_agents`** — full v1.27.0 surface (services, schedule,
   retention, restore-defaults, volumes, alert configs).
9. **`slide_devices`** — `list`, `get`, `update`, `poweroff`, `reboot`
   plus device network + per-device VLAN CRUD.

### 👥 Clients & admin

10. **`slide_clients`** — client CRUD (`list`, `get`, `create`,
    `update`, `delete`).
11. **`slide_admin`** — users + accounts + user avatar
    (`list_users`, `get_user`, `get_user_avatar`, `list_accounts`,
    `get_account`, `update_account`).

### 🔁 Backward-compat shim

- **`list_all_clients_devices_and_agents`** still works; it forwards to
  `slide_overview operation=inventory`.

## MCP Prompts (slash-command workflows)

Claude Desktop surfaces these in its slash-command UI:

- **`/slide.welcome`** — one-message intro for first-time users
- **`/slide.daily-status [client?] [hours?]`** — daily ops summary
- **`/slide.triage-alerts`** — prioritised unresolved-alert review
- **`/slide.restore-file [filename?] [agent?]`** — guided file recovery
- **`/slide.boot-recovery-vm [agent?]`** — guided VM recovery
- **`/slide.dr-runbook [device?]`** — DR runbook generated from the
  device's real agents and snapshots

## MCP Resources (cheap pre-baked context)

Claude can read these without burning a tool call:

- `slide://welcome` — one-page primer + example questions (no API calls)
- `slide://help/glossary` — Slide-specific terminology (no API calls)
- `slide://help/troubleshoot` — common setup / auth / network problems
- `slide://overview/inventory` — clients → devices → agents tree
- `slide://overview/health` — one-line-per-device-and-agent summary
- `slide://alerts/unresolved` — open alerts, prioritised
- `slide://audit/recent` — last 24h of audit log
- `slide://docs/openapi` — live Slide OpenAPI spec (cached 1h)
- `slide://client/{client_id}` — single client detail
- `slide://device/{device_id}` — single device detail
- `slide://agent/{agent_id}` — single agent detail
- `slide://agent/{agent_id}/snapshots/recent` — recent snapshots

The legacy `slide://context/clients-devices-agents` URI keeps working
as an alias for `slide://overview/inventory`.

## Token-conscious response shape

Every list/get operation accepts these cross-cutting parameters:

- `format` — `summary` (default for lists; one-line entries),
  `compact` (full payload, no indentation), or `detailed` (full
  payload, indented — equivalent to v3 behaviour).
- `fields` — comma-separated JSON projection, e.g.
  `fields=id,hostname,last_seen` drops everything else.
- `hints` — set to `off` to suppress the `next_steps` array v5 appends
  to most responses. Default is on; the hints are short and cheap.
- All list responses include `pagination.next_offset` and `count`.

## name_hint: no IDs to memorise

Every tool that needs an `agent_id`, `device_id`, or `client_id` also
accepts `name_hint`. Pass a hostname, display name, client name, or any
substring (`name_hint=bob`, `name_hint=ACME`, `name_hint=DC-01`) and
the server fuzzy-resolves it server-side.

- **0 matches** → the response is a `{"name_hint_error":"no_match",...}`
  payload with a suggestion to call `slide_overview operation=inventory`.
- **1 match** → the resolved entity appears in the response under
  `_resolved` so the assistant can confirm to the user which one it
  picked, then the operation proceeds normally.
- **2+ matches** → the response is a
  `{"name_hint_error":"ambiguous","candidates":[...]}` payload with up
  to 10 candidates. The assistant should ask the user to pick one,
  then re-call with the explicit `*_id`.

## Key Features

### 🔐 Infrastructure Management
- **Device Control**: Remote power operations, hostname management, client assignment
- **Agent Deployment**: Automated pairing, display name management  
- **Network Isolation**: Client-specific networks with VPN access
- **Advanced Networking**: IPSec tunnels, port forwarding, WireGuard peers

### 💽 Data Protection & Recovery
- **Automated Backups**: Agent-based backup initiation and monitoring
- **Point-in-Time Recovery**: Snapshot browsing with location filtering
- **Flexible Restores**: File-level and full disk image exports
- **Multiple Formats**: VHD, VHDX (dynamic/fixed), Raw disk images
- **Boot Modifications**: Optional passwordless admin account creation

### ☁️ Virtualization
- **VM Creation**: Create VMs from any snapshot
- **Resource Control**: Configurable CPU/RAM allocation  
- **Network Integration**: Connect VMs to isolated networks
- **Console Access**: Browser-based VNC for direct VM interaction
- **RDP Bookmarks**: Generate downloadable .rdp files for easy Windows Remote Desktop access

### 📊 Monitoring & Organization
- **Alert Management**: Centralized alert monitoring and resolution
- **Client Organization**: Group resources by client for better management
- **User Management**: Account access and permissions
- **Comprehensive Filtering**: Advanced pagination and sorting across all resources

### 📋 Professional Data Presentation
- **Smart Cards**: Individual and table-based cards for agents, clients, devices, and snapshots
- **Report Templates**: Runbook procedures, daily summaries, and monthly analysis reports
- **Multiple Formats**: HTML, Markdown, and HAML output for different use cases
- **Dashboard Ready**: Pre-formatted cards perfect for status displays and monitoring
- **Documentation Support**: Professional templates for operational procedures and troubleshooting

All meta-tools support pagination (`limit`, `offset`), sort options where
applicable, and the cross-cutting `format` / `fields` parameters described
above.

## 📦 Detailed installation reference

The "⚡ Install" section near the top of this README covers the
recommended paths. This section documents the underlying configuration so
you can hand-roll setups for unusual hosts or CI.

### Getting an API Key

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to **My Settings -> API Tokens**
3. Generate your API token. Treat it like a password - it grants full
   access to your Slide account.

### Pre-built downloads

The headline release asset is the `.mcpb` (Claude Desktop Extension):

- **slide-mcp-server.mcpb** (one file, every platform):
  `https://github.com/austinmcchord/slideMCP/releases/latest/download/slide-mcp-server.mcpb`

Standalone binaries (for `claude mcp add` / Cursor / CI):

- **macOS** (universal Apple Silicon + Intel): `slide-mcp-server-vX.Y.Z-darwin-universal.tar.gz`
- **Linux x64**: `slide-mcp-server-vX.Y.Z-linux-amd64.tar.gz`
- **Linux ARM64**: `slide-mcp-server-vX.Y.Z-linux-arm64.tar.gz`
- **Windows x64**: `slide-mcp-server-vX.Y.Z-windows-x64.zip`

All macOS artifacts are signed with a Developer ID and notarized.

### Build from source

```bash
git clone https://github.com/austinmcchord/slideMCP.git
cd slideMCP
make build      # build/slide-mcp-server (current platform)
make pack-dxt   # build/slide-mcp-server.mcpb (universal bundle, unsigned)
```

### Manual Claude Desktop config

If you can't use the `.mcpb` (older Claude Desktop, locked-down corp
machine, etc.), edit `claude_desktop_config.json` directly:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": { "SLIDE_API_KEY": "YOUR_API_KEY_HERE" }
    }
  }
}
```

With non-default tool mode and disabled tools:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE",
        "SLIDE_TOOLS": "read-only",
        "SLIDE_DISABLED_TOOLS": "slide_recovery"
      }
    }
  }
}
```

Or pass everything as CLI args:

```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "args": [
        "--api-key", "YOUR_API_KEY_HERE",
        "--tools", "full-safe",
        "--disabled-tools", "slide_agents,slide_backups"
      ]
    }
  }
}
```

### Test Your Installation

Use the official MCP Inspector to drive the server interactively:

```bash
export SLIDE_API_KEY=tk_...
npx @modelcontextprotocol/inspector ./slide-mcp-server
```

Or run the bundled smoke test (covers initialize, tools/list, resources/list, the
v1.27.0 validation paths, and a sampling of live tool calls):

```bash
SLIDE_API_KEY=tk_... ./smoke_test.sh
```

## 🔧 CLI Arguments & Configuration

The Slide MCP Server supports several command-line arguments for flexible configuration:

### Command Line Arguments

```bash
# Basic usage with API key
./slide-mcp-server --api-key YOUR_API_KEY

# All available flags
./slide-mcp-server [OPTIONS]
```

| Flag | Description | Environment Variable | Default |
|------|-------------|---------------------|---------|
| `--api-key` | Slide API key for authentication | `SLIDE_API_KEY` | Required |
| `--base-url` | Base URL for Slide API endpoint | `SLIDE_BASE_URL` | `https://api.slide.tech` |
| `--tools` | Permission mode (`read-only` / `safe` / `full`; legacy aliases `reporting`/`restores`/`full-safe` still work) | `SLIDE_TOOLS` | `safe` |
| `--disabled-tools` | Comma-separated list of tools to disable (`slide_help` cannot be disabled) | `SLIDE_DISABLED_TOOLS` | None |
| `--doctor` | Run self-diagnostic checks (token, network, sample reads) and exit | - | - |
| `--skip-startup-validation` | Skip the startup probe of `/v1/account` (useful when launching offline) | - | - |
| `--tool` | Run a single tool then exit (one-shot mode) | - | - |
| `--args` | JSON arguments for `--tool` | - | - |
| `--version` | Show version information and exit | - | - |

**Priority**: CLI flags take precedence over environment variables.

### Examples

```bash
# Using CLI flags (default 'safe' mode)
./slide-mcp-server --api-key sk_test_123 --base-url https://custom.api.endpoint

# Read-only mode (only list/get/search/browse)
./slide-mcp-server --api-key sk_test_123 --tools read-only

# Using environment variables
export SLIDE_API_KEY="sk_test_123"
export SLIDE_TOOLS="read-only"
./slide-mcp-server

# Mixed usage (CLI overrides environment)
export SLIDE_TOOLS="full"
./slide-mcp-server --api-key sk_test_123 --tools read-only

# Disable specific tools
./slide-mcp-server --api-key sk_test_123 --disabled-tools "slide_recovery,slide_files"

# One-shot tool invocation (handy for scripts / smoke tests)
./slide-mcp-server --api-key sk_test_123 --tool slide_overview --args '{"operation":"health"}'

# Show version
./slide-mcp-server --version
# Output: slide-mcp-server version 4.0.0
```

### 🚫 Disabling Specific Tools

In addition to permission modes, you can disable specific tools entirely using the `--disabled-tools` flag or `SLIDE_DISABLED_TOOLS` environment variable. This provides fine-grained control over which tools are available.

#### Usage Examples

```bash
# Disable specific tools via CLI flag
./slide-mcp-server --api-key YOUR_KEY --disabled-tools "slide_agents,slide_backups"

# Disable tools via environment variable
export SLIDE_DISABLED_TOOLS="slide_devices,slide_users"
./slide-mcp-server --api-key YOUR_KEY

# Combined with permission modes
./slide-mcp-server --api-key YOUR_KEY --tools reporting --disabled-tools "slide_snapshots"

# Whitespace is handled gracefully
./slide-mcp-server --api-key YOUR_KEY --disabled-tools " slide_agents , slide_backups , slide_devices "
```

#### Available Tool Names (v4.0.0)

- `slide_overview` — inventory + health + per-client/per-device summaries
- `slide_files` — file search + restore + push
- `slide_recovery` — VMs + image exports + DR networks
- `slide_audit` — account audit log
- `slide_clients` — client CRUD
- `slide_admin` — users + accounts
- `slide_agents`, `slide_devices`, `slide_snapshots`, `slide_backups`, `slide_alerts` — lower-level CRUD with v4 task-oriented additions
- `list_all_clients_devices_and_agents` — backward-compat alias for `slide_overview operation=inventory`

#### Key Features

- **Precedence**: CLI flags take precedence over environment variables
- **Whitespace Handling**: Extra spaces around tool names are automatically trimmed
- **Error Messages**: Clear error messages when attempting to use disabled tools
- **Combined Filtering**: Works alongside permission modes for layered access control
- **Transparency**: Logs which tools are disabled on server startup

#### Use Cases

```bash
# Read-only server (only list/get/search/browse operations)
./slide-mcp-server --tools read-only --disabled-tools "slide_admin"

# Allow restores but hide the recovery tool entirely
./slide-mcp-server --tools safe --disabled-tools "slide_recovery"

# Monitoring-only setup
./slide-mcp-server --tools read-only --disabled-tools "slide_recovery,slide_files"
```

When a disabled tool is called, the server returns:
```json
{
  "error": {
    "code": -32601,
    "message": "Tool 'slide_agents' is disabled"
  }
}
```

## 🔒 Permission Modes

v4.0.0 collapsed the permission system from four tiers down to three.
Legacy mode names (`reporting`, `restores`, `full-safe`) are silently
aliased so existing CLI flags / Claude Desktop user configs keep
working.

### Permission Levels

#### `read-only` — Pure read access

**Use Case**: monitoring dashboards, on-call sanity checks, anything
that should never accidentally mutate state.

- ✅ **Allowed**: every list/get/search/browse op, every Resource
  read, every Prompt
- ❌ **Blocked**: every create/update/delete/start/restore/etc.
- Legacy alias: `reporting`

```bash
./slide-mcp-server --api-key YOUR_KEY --tools read-only
```

#### `safe` — Default; everything except destructive ops

**Use Case**: general MSP / IT admin workflows including restores,
agent settings, backup launches, alert resolution.

- ✅ **Allowed**: read-only + create/update for restores, agents,
  devices, networks, alerts, settings, backup launch
- ❌ **Blocked**: deletes (`delete`, `delete_vm`, `delete_image`,
  `delete_network`, `delete_restore`, `delete_vlan`, ...) and device
  power control (`poweroff`, `reboot`)
- Legacy aliases: `restores`, `full-safe`

```bash
./slide-mcp-server --api-key YOUR_KEY --tools safe
# (this is the default; the flag is optional)
```

#### `full` — Complete access

**Use Case**: advanced admins who need to delete agents, delete
snapshots, or remotely power-cycle a device.

- ✅ **Allowed**: every operation, including the irreversible ones
- ⚠️ Includes agent deletion, snapshot deletion, device poweroff/reboot

```bash
./slide-mcp-server --api-key YOUR_KEY --tools full
```

### Permission Matrix

| Operation category                | `read-only` | `safe` | `full` |
|-----------------------------------|-------------|--------|--------|
| List / get / search / browse      | ✅          | ✅     | ✅     |
| File search + restore + push      | ❌          | ✅     | ✅     |
| Boot recovery VMs / export images | ❌          | ✅     | ✅     |
| DR network create / update        | ❌          | ✅     | ✅     |
| Agent / device settings updates   | ❌          | ✅     | ✅     |
| Alert resolution                  | ❌          | ✅     | ✅     |
| Backup launch                     | ❌          | ✅     | ✅     |
| **Deletes** (agents, snapshots, VMs, networks, port-forwards, ...)  | ❌ | ❌ | ✅ |
| **Device poweroff / reboot**      | ❌          | ❌     | ✅     |

### Security Recommendations

- **Production monitoring**: `read-only`
- **Day-to-day MSP admin**: `safe` (default)
- **Cleanup / decommissioning workflows**: `full`

### Usage with VS Code

For VS Code integration, add the following JSON block to your User Settings (JSON) file. You can do this by pressing `Ctrl + Shift + P` and typing `Preferences: Open User Settings (JSON)`.

Optionally, you can add it to a file called `.vscode/mcp.json` in your workspace. This will allow you to share the configuration with others.

> Note that the `mcp` key is not needed in the `.vscode/mcp.json` file.

```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "slide_api_key",
        "description": "Slide API Key",
        "password": true
      }
    ],
    "servers": {
      "slide": {
        "command": "/path/to/slide-mcp-server",
        "env": {
          "SLIDE_API_KEY": "${input:slide_api_key}"
        }
      }
    }
  }
}
```

With custom configuration and disabled tools:
```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "slide_api_key",
        "description": "Slide API Key",
        "password": true
      }
    ],
    "servers": {
      "slide": {
        "command": "/path/to/slide-mcp-server",
        "env": {
          "SLIDE_API_KEY": "${input:slide_api_key}",
          "SLIDE_TOOLS": "read-only",
          "SLIDE_DISABLED_TOOLS": "slide_admin"
        }
      }
    }
  }
}
```

## 💡 Usage Examples

### Health summary

```json
{
  "name": "slide_overview",
  "arguments": {"operation": "health"}
}
```

### "Did backups run last night for ACME?"

```json
{
  "name": "slide_backups",
  "arguments": {
    "operation": "status_for_client",
    "client_id": "c_0123456789ab",
    "hours": 24
  }
}
```

### Find a file across every snapshot for an agent

```json
{
  "name": "slide_files",
  "arguments": {
    "operation": "search",
    "agent_id": "a_0123456789ab",
    "search_term": "Q4-budget"
  }
}
```

### Find every snapshot that contains a specific path

```json
{
  "name": "slide_files",
  "arguments": {
    "operation": "versions",
    "agent_id": "a_0123456789ab",
    "path": "C:\\Users\\bob\\Documents\\Q4-budget.xlsx"
  }
}
```

### Boot a recovery VM

```json
{
  "name": "slide_recovery",
  "arguments": {
    "operation": "boot_vm",
    "snapshot_id": "s_0123456789ab",
    "device_id": "d_0123456789ab",
    "cpu_count": 4,
    "memory_in_mb": 8192,
    "network_type": "network-id",
    "network_source": "n_0123456789ab"
  }
}
```

### Generate an RDP bookmark for a running VM

```json
{
  "name": "slide_recovery",
  "arguments": {
    "operation": "get_rdp_bookmark",
    "virt_id": "v_0123456789ab"
  }
}
```

### Audit log: what changed in the last 24 hours?

```json
{
  "name": "slide_audit",
  "arguments": {"operation": "recent", "hours": 24}
}
```

### Triage open alerts (sorted worst-first)

```json
{
  "name": "slide_alerts",
  "arguments": {"operation": "triage"}
}
```

### Start a backup

```json
{
  "name": "slide_backups",
  "arguments": {"operation": "start", "agent_id": "a_0123456789ab"}
}
```

### Export a snapshot as a VHDX image

```json
{
  "name": "slide_recovery",
  "arguments": {
    "operation": "export_image",
    "snapshot_id": "s_0123456789ab",
    "device_id": "d_0123456789ab",
    "image_type": "vhdx"
  }
}
```

## Build

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release packages
make release

# View available commands
make help
```

## Release Process

### Unified Release Script (Recommended)

The `release-all-in-one.sh` script automates the entire release workflow:

```bash
# Auto-increment patch version and create release
./scripts/release-all-in-one.sh

# Release specific version
./scripts/release-all-in-one.sh v2.5.0

# Test release process without making changes
./scripts/release-all-in-one.sh --dry-run v2.5.0

# Build locally without pushing to GitHub
./scripts/release-all-in-one.sh --no-push

# Re-release existing version (overwrite tag and GitHub release)
./scripts/release-all-in-one.sh --force v2.5.0
```

**What it does:**
1. Checks prerequisites (Go, GitHub CLI, jq)
2. Validates git status (clean working directory, on main branch)
3. Updates version numbers in Makefile and config.go
4. Runs tests (optional with `--skip-tests`)
5. Commits version changes
6. Builds binaries for all platforms
7. Signs and notarizes macOS binaries (if credentials provided)
8. Creates release packages (.tar.gz for Unix, .zip for Windows)
9. Generates SHA256 checksums
10. Creates release notes from git history
11. Tags the release
12. Pushes to GitHub
13. Creates GitHub release with all assets

**Environment Variables (REQUIRED for releases):**
- `APPLE_ID`: Apple ID for macOS notarization
- `APP_SPECIFIC_PASSWORD`: App-specific password for notarization

⚠️ **Important**: macOS binaries MUST be signed and notarized to work on user systems. The script enforces this requirement and will fail if credentials are not provided. For local testing only, use `--no-push --skip-macos-signing`.

**Setting Up Credentials (One-Time Setup):**

```bash
# Option 1: Use credentials file (Recommended - saves credentials locally)
# File is already created at scripts/release-env.sh and gitignored for security
source scripts/release-env.sh

# Option 2: Set manually each time
export APPLE_ID='your-apple-id@example.com'
export APP_SPECIFIC_PASSWORD='xxxx-xxxx-xxxx-xxxx'
```

**Getting Apple Credentials:**
1. **Apple ID**: Your Apple Developer account email
2. **App-Specific Password**:
   - Go to [appleid.apple.com](https://appleid.apple.com)
   - Sign in and navigate to Security section
   - Click "Generate Password" under App-Specific Passwords
   - Copy the generated password (format: xxxx-xxxx-xxxx-xxxx)
3. **Developer ID Certificate**: Install from Apple Developer account
   - Required for code signing
   - Must be "Developer ID Application" certificate

**Security Note**: The `scripts/release-env.sh` file is gitignored and will never be committed to the repository.

**Options:**
- `--dry-run`: Simulate without making changes
- `--skip-tests`: Skip running tests
- `--no-push`: Build locally without GitHub operations
- `--skip-macos-signing`: Skip signing (TESTING ONLY, must use with --no-push)
- `--force` or `-f`: Force overwrite existing release (deletes tag and GitHub release)
- `--help`: Show detailed help

### Manual Release (Legacy)

For manual control, use the existing `automated-release.sh` script (see `scripts/RELEASE_README.md`).

## Architecture Benefits

### For LLMs
- **Reduced Complexity**: 14 meta-tools vs 52+ individual tools
- **Logical Grouping**: Related operations organized together
- **Consistent Interface**: All meta-tools follow the same operation pattern
- **Better Context**: Less tool switching, more focused conversations

### For Developers  
- **Maintainable**: Each meta-tool in its own file
- **Extensible**: Easy to add new operations to existing categories
- **Backward Compatible**: All original functionality preserved
- **Schema Validation**: Conditional parameter validation per operation

## License

This MCP server is licensed under the MIT License. This means you are free to use, modify, and distribute the software, subject to the terms and conditions of the MIT License. For more details, please see the LICENSE file in the project repository.
