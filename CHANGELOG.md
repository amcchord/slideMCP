# Slide MCP Changes

## 2026-05-14 - v5.0.0 - Novice-first pass: discoverability + fuzzy names + doctor

### Headline

v5.0.0 makes slide-mcp-server reach for itself. The LLM now picks this
extension automatically whenever the user mentions Slide, BCDR,
backups, snapshots, restores, recovery VMs, file recovery, disaster
recovery, alerts, or any MSP workflow - no "use the Slide MCP"
prompting required. Operators never have to memorise `a_xxxxxxxxxxxx`
IDs: every tool that wants an `agent_id` / `device_id` / `client_id`
also accepts `name_hint` and the server fuzzy-resolves it. First-run
errors carry novice-friendly remediation steps; `slide-mcp-server
--doctor` self-diagnoses token / network / endpoint problems.

### Headline new surfaces

- **`slide_help`** - new top-level tool with operations
  `getting_started`, `examples`, `glossary`, `troubleshoot`,
  `list_prompts`, `list_resources`, `what_can_you_do`. Read-only and
  always available, even in `read-only` mode; cannot be disabled via
  `--disabled-tools`. Content lives in [`docs/help/*.md`](docs/help/)
  and is embedded via `//go:embed` (no network).
- **`/slide.welcome`** - new slash-command prompt that seeds Claude
  with a one-message intro listing the trigger vocabulary, active
  permission tier, and 5 example questions to try.
- **`slide://welcome`**, **`slide://help/glossary`**,
  **`slide://help/troubleshoot`** resources serve the same markdown
  primer / glossary / troubleshooting FAQ at conversation start, no
  API calls required.

### Discoverability ("when should the LLM reach for me?")

- **Rewritten `serverInstructions()`** ([server.go](server.go)). The
  first sentence is now the literal "use slide-mcp-server whenever the
  user mentions ..." trigger-vocabulary list (Slide / BCDR /
  business continuity / disaster recovery / backup / snapshot / restore
  / recovery VM / image export / DR network / Slide alert / unresolved
  alert / audit log / MSP client / Slide agent / protected system /
  ...). Also adds an inline glossary, an active-permission-tier note,
  6 worked example user prompts mapped to their canonical first tool
  call, and an explicit "call slide_help getting_started if you don't
  know where to start" fallback.
- **Every tool's `Description` now leads with a
  "REACH FOR THIS whenever the user mentions ..." sentence** so the
  trigger vocabulary appears in tool-relevance scoring too. Reworded
  consistently across `slide_overview`, `slide_files`,
  `slide_recovery`, `slide_audit`, `slide_clients`, `slide_admin`,
  `slide_agents`, `slide_devices`, `slide_snapshots`, `slide_backups`,
  `slide_alerts`, and `list_all_clients_devices_and_agents`.
- **`TestTriggerVocabularyCoverage`** in
  [`server_test.go`](server_test.go) is a regression check that asserts
  every canonical phrase still appears in at least one description or
  the server instructions. Catches accidental deletions during future
  refactors.

### Server-side fuzzy name resolution (`name_hint`)

- **New [`name_resolver.go`](name_resolver.go)**. Resolves a free-form
  `name_hint` (hostname, display name, client name, or any substring,
  case-insensitive) to the canonical `agent_id` / `device_id` /
  `client_id` server-side. 15-minute cache, automatic refresh on a
  miss. Match precedence: exact > prefix > substring, stop at first
  non-empty tier (deterministic; no Levenshtein for predictability).
- **0 matches** -> structured `{"name_hint_error":"no_match",...}`
  response with a suggestion to call `slide_overview inventory`.
- **1 match** -> the resolved ID is written into args, the operation
  proceeds normally, and the response carries a top-level `_resolved`
  block describing what got picked so the LLM can confirm to the user.
- **2+ matches** -> structured
  `{"name_hint_error":"ambiguous","candidates":[...]}` response with
  up to 10 candidates so the LLM can ask the user to pick one.
- **Wired into** `slide_overview` (for_client / for_device),
  `slide_files` (search / versions), `slide_backups` (start /
  status_for_client / status_for_device / recent_for_agent),
  `slide_snapshots` (recent_for_agent), `slide_clients`
  (get/update/delete), `slide_devices` (every single-device op), and
  `slide_agents` (every single-agent op plus create/pair which
  resolve a device).
- Schema updates: every affected tool's properties include a
  `name_hint` field; conditional `allOf` blocks use a new `reqEither`
  helper so either `*_id` OR `name_hint` satisfies the requirement.
- Dispatcher-level: [`base_handler.go`](base_handler.go) consumes
  per-operation `ResolutionSpec` entries from each tool's
  `CreateToolConfigWithResolutions` call. Individual operation
  handlers continue to `requireString(args, "agent_id")` unchanged -
  the resolver populates args before dispatch.

### Response affordances

- **`next_steps`** array appended to every response with 1-3
  curated follow-up suggestions, e.g. after `slide_files search` the
  caller gets a hint to call `slide_files versions agent_id=... path=...`
  next. Suppressed with the new cross-cutting `hints=off` argument.
- **`_resolved`** block (see name_hint above) carries
  `{"id","name","kind","detail"}` so the LLM can show the user which
  entity it auto-resolved when name_hint was used.
- Both fields are spliced into the JSON envelope by
  [`format.go`](format.go) via [`hints.go`](hints.go); per-handler
  code does not change.

### Startup validation + `--doctor`

- **Token validation on startup** ([`doctor.go`](doctor.go)). Before
  the stdio server comes up, a single `GET /v1/account?limit=1` probe
  confirms the token is accepted. On 401/403 the server exits with a
  multi-line, novice-friendly remediation message pointing at
  console.slide.tech. On network errors it logs a warning and stays
  alive so `slide_help operation=troubleshoot` remains callable. Skip
  with `--skip-startup-validation`.
- **`slide-mcp-server --doctor`** - new self-check subcommand that
  probes token + connectivity + a sample read against each major
  endpoint family (`/v1/account`, `/v1/client`, `/v1/device`,
  `/v1/agent`), prints an `OK / FAIL` checklist, and exits non-zero on
  any failure. Idempotent; safe to wire into CI.

### Error message polish

- **`APIError.Error()`** ([`api.go`](api.go)) now carries novice-
  friendly per-status hints:
  - 401: "your Slide API token was rejected. Generate a fresh one at
    https://console.slide.tech under My Settings -> API Tokens ..."
  - 403: "your token doesn't have access to this resource, or your
    user role lacks the required permission ..."
  - 404: detects `a_` / `d_` / `c_` / `s_` / `v_` prefixes in the
    endpoint and explains what kind of ID was 404'd.
  - 429 / 5xx: pointer at status.slide.tech for incident lookup.
- **Unknown operation errors** ([`base_handler.go`](base_handler.go))
  now include a Levenshtein-based did-you-mean (`unknown operation
  "searh" ... did you mean "search"?`) plus the full list of valid
  operations for the tool.

### Cross-cutting parameter additions

- New `hints=on|off` on every list/single response shape. Default
  `on`; `off` suppresses the `next_steps` array described above.
- New `name_hint=<string>` on every tool that takes an `agent_id`,
  `device_id`, or `client_id` (see fuzzy resolution section above).

### Manifest + README

- [`dxt/manifest.json`](dxt/manifest.json): punchier `description`,
  rewritten `long_description` highlighting trigger vocabulary,
  example questions, and `name_hint`; `user_config.tools_mode` now
  has an `enum` constraint + friendlier copy.
- [`README.md`](README.md): new "What to ask first" section near the
  top with 7 copy-paste example questions, plus a dedicated
  "name_hint: no IDs to memorise" subsection. `slide_help` listed in
  the meta-tools table; `--doctor` and `--skip-startup-validation`
  documented in the CLI argument table.

### Things that did NOT change

- All 11 v4 meta-tools' operation enums are preserved (additive only).
- Permission tier semantics, signing pipeline, `.mcpb` build/release
  flow, Slide API endpoint coverage (still v1.27.0).
- All v4 prompts and resources keep their URIs; v3
  `slide://context/clients-devices-agents` is still an alias for
  `slide://overview/inventory`.

### Migration table

| If you were doing this in v4...           | In v5...                                          |
|--------------------------------------------|---------------------------------------------------|
| Resolving hostnames via `slide_overview inventory` then matching | Just pass `name_hint=<hostname>` directly         |
| Parsing the bare response JSON                | Same shape - now with optional `next_steps` and `_resolved` siblings. `hints=off` to suppress |
| Manually testing the token after install     | `slide-mcp-server --doctor` (or wait for startup validation to fail loudly) |
| Telling Claude "use the Slide MCP for this"  | Not needed - the trigger vocabulary covers it     |

---

## 2026-05-11 - v4.0.0 - Ground-up overhaul for Claude Desktop end users

### Headline

v4.0.0 reshapes the entire tool surface around the actual questions
an MSP technician asks Claude Desktop ("did backups run last night for
ACME?", "find Q4-budget.xlsx on Bob's laptop and restore it",
"are all my Slide boxes healthy?"), adopts modern MCP primitives
(Prompts, URI-templated Resources, structured outputs, tool
annotations), plugs the highest-value Slide API gaps (file search and
audit logs were not exposed in v3), and trims ~3,000 LoC of legacy
docs/reports/presentation tooling that the LLM can compose from raw
data anyway.

### New capabilities

- **`slide_files`** - the headline new tool. Search filenames across
  every snapshot for an agent (Slide API v1.27.0 `FileSearch`), list
  snapshot versions of a path (`PathVersion`), then create a restore,
  browse it, or push files back to the protected system. Single tool
  covers the entire "find and recover X" workflow.
- **`slide_audit`** - account audit log queries (Slide API v1.27.0
  `Audits` / `AuditByID` / `AuditActions` / `AuditResources`). Includes
  a `recent` convenience op for "what changed in the last N hours?".
- **`slide_overview`** - inventory + per-account health summary, plus
  `for_client` and `for_device` deep views that join devices, agents,
  and open alerts in one call. Replaces the v3
  `list_all_clients_devices_and_agents` top-level tool (the old name
  still works as a backward-compat alias).
- **`slide_recovery`** - consolidates v3's `slide_vms`, image-export
  surface from `slide_restores`, and `slide_networks` into one
  task-oriented tool. `boot_vm`, `export_image`, `create_network`,
  `create_wg_peer`, and friends.
- **`slide_clients`** + **`slide_admin`** - splits the v3
  `slide_user_management` mega-tool into the two halves an MSP
  actually thinks about: client CRUD vs user/account admin.
- **`slide_backups status_for_client` / `status_for_device` /
  `recent_for_agent`** - one-call answers to "did backups run for X?".
- **`slide_alerts triage`** - groups unresolved alerts by severity hint
  and returns the worst-first list for prioritised remediation.
- **`slide_snapshots recent_for_agent`** - one-call "what restore
  points do I have for X?".

### Modern MCP primitives

- **5 MCP Prompts** wired via `srv.AddPrompt`. Surface in Claude
  Desktop's slash-command UI:
  - `/slide.daily-status [client] [hours]` - daily ops summary
  - `/slide.triage-alerts` - prioritised unresolved alert review
  - `/slide.restore-file [filename] [agent]` - guided file recovery
  - `/slide.boot-recovery-vm [agent]` - guided VM recovery
  - `/slide.dr-runbook [device]` - generates a tailored DR runbook
    grounded in the device's real agents and snapshots.
- **9 MCP Resources** (up from 1), including URI-templated
  `slide://client/{client_id}`, `slide://device/{device_id}`,
  `slide://agent/{agent_id}`, plus static `slide://overview`,
  `slide://overview/health`, `slide://alerts/unresolved`,
  `slide://audit/recent`, and a live-cached `slide://docs/openapi`.
  The legacy `slide://context/clients-devices-agents` URI keeps
  working as an alias.
- **Tool annotations** on every tool: `readOnlyHint`,
  `destructiveHint`, `idempotentHint`, `openWorldHint`. Claude Desktop
  uses these to skip confirmation prompts on safe reads.
- **Output schemas** on every tool. Permissive but signal "structured
  object response" to clients that render `outputSchema`.
- **Token-conscious response shape**. Every list/get op accepts:
  - `format=summary|compact|detailed` - default `summary` for lists
    (one-line entries) and `compact` for single-object responses.
    `detailed` is the v3 `MarshalIndent` behaviour.
  - `fields=a,b,c` - JSON projection. Drops everything else from each
    entry / from the response object.
  - Lists always include `pagination.next_offset` and `count`.

### Reliability

- Reworked HTTP client (in [`api.go`](api.go)):
  - Honours `Retry-After` header on 429s (header preferred,
    exponential backoff fallback).
  - Single retry on transient 5xx (502 / 503 / 504).
  - Sets `User-Agent: slide-mcp-server/4.0.0` and `Accept:
    application/json` on every request.
  - Errors return a structured `*APIError` with `method`, `endpoint`,
    `status`, body, and a per-status hint (e.g. "check that
    SLIDE_API_KEY is valid" on 401/403).

### Permission tiers

Collapsed from four (`reporting`/`restores`/`full-safe`/`full`) to
three (`read-only`/`safe`/`full`). Legacy mode names are silently
mapped:

  - `reporting` -> `read-only`
  - `restores` -> `safe`
  - `full-safe` -> `safe`
  - `full` -> unchanged

`safe` is the new default. `slide-mcp-server --tools <legacy-name>`
keeps working unchanged.

### Retired (~3,000 LoC removed)

- **`slide_docs`** ([`tools_docs.go`](tools_docs.go), 911 LoC) was a
  hardcoded fake doc index. Replaced by the live-cached
  `slide://docs/openapi` resource.
- **`slide_presentation`** ([`tools_presentation.go`](tools_presentation.go),
  333 LoC) - GitHub-fetched runbook templates. Replaced by the
  `/slide.dr-runbook` and `/slide.daily-status` prompts that ground in
  the user's own data.
- **`slide_reports`** ([`tools_reports.go`](tools_reports.go), 1033
  LoC) - power-user pre-baked reports. The LLM composes these from
  `slide_overview` + `slide_backups status_for_*` now.
- **`slide_meta`** ([`tools_meta.go`](tools_meta.go), 746 LoC) -
  fully absorbed into `slide_overview` + Resources. The old
  `listAllClientsDevicesAndAgents` helper is preserved in
  [`overview_helpers.go`](overview_helpers.go) and reachable via the
  backward-compat `list_all_clients_devices_and_agents` tool name.
- **`slide_vms`**, **`slide_restores`** (image export bits),
  **`slide_networks`** - merged into `slide_recovery`.
- **`slide_user_management`** - split into `slide_clients` +
  `slide_admin`.
- The `--enable-reports` and `--enable-presentation` CLI flags +
  `SLIDE_ENABLE_REPORTS` / `SLIDE_ENABLE_PRESENTATION` env vars are
  gone (the tools they gated are gone). The flags being absent
  silently no-ops; nobody's `claude_desktop_config.json` will fail to
  start because of it.

### Migration table (old -> new)

| v3 tool / op                                  | v4 equivalent                                               |
|-----------------------------------------------|-------------------------------------------------------------|
| `list_all_clients_devices_and_agents`         | `slide_overview operation=inventory` (legacy name still works) |
| `slide_meta operation=get_snapshot_changes`   | compose via `slide_snapshots recent_for_agent`              |
| `slide_meta operation=get_reporting_data`     | compose via `slide_backups status_for_client`               |
| `slide_user_management list_clients/...`      | `slide_clients list/...`                                    |
| `slide_user_management list_users/get_account/...` | `slide_admin list_users/get_account/...`               |
| `slide_vms ...`                               | `slide_recovery list_vms/boot_vm/...`                       |
| `slide_restores list_images/create_image/...` | `slide_recovery list_images/export_image/...`               |
| `slide_restores list_files/create_file/...`   | `slide_files list_restores/create_restore/...` *(new file-search ops added on top)* |
| `slide_networks ...`                          | `slide_recovery list_networks/create_network/...`           |
| `slide_docs ...`                              | resource read of `slide://docs/openapi` (or just ask)       |
| `slide_presentation get_runbook_template`     | `/slide.dr-runbook` MCP prompt                              |
| `slide_reports get_daily_report`              | `/slide.daily-status` MCP prompt                            |

### Tests

- New [`server_test.go`](server_test.go) covers: every v4 tool +
  operation is registered, prompts/list and resources/list advertise
  the v4 surface, legacy permission-tier names are aliased correctly,
  `slide_files search` end-to-end against an httptest server,
  `slide_audit recent` end-to-end against an httptest server,
  `parseRetryAfter` handles delta-seconds and HTTP-date forms,
  one-shot CLI rejects disabled and unknown tools.

### Files touched

- Added: [`tools_overview.go`](tools_overview.go),
  [`tools_files.go`](tools_files.go),
  [`tools_recovery.go`](tools_recovery.go),
  [`tools_audit.go`](tools_audit.go),
  [`tools_clients.go`](tools_clients.go),
  [`tools_admin.go`](tools_admin.go),
  [`api_files.go`](api_files.go), [`api_audit.go`](api_audit.go),
  [`format.go`](format.go), [`annotations.go`](annotations.go),
  [`output_schemas.go`](output_schemas.go),
  [`prompts.go`](prompts.go), [`resources.go`](resources.go),
  [`overview_helpers.go`](overview_helpers.go).
- Rewritten: [`registry.go`](registry.go), [`server.go`](server.go),
  [`config.go`](config.go), [`main.go`](main.go),
  [`tools_alerts.go`](tools_alerts.go),
  [`tools_backups.go`](tools_backups.go),
  [`tools_snapshots.go`](tools_snapshots.go),
  [`server_test.go`](server_test.go).
- Updated HTTP client: [`api.go`](api.go) (Retry-After + 5xx retry +
  structured `APIError`).
- Bumped: [`Makefile`](Makefile) `VERSION = v4.0.0`,
  [`dxt/manifest.json`](dxt/manifest.json) version + long_description
  + tools_mode default + description, [`config.go`](config.go)
  `Version = "4.0.0"`.
- Deleted: `tools_docs.go`, `tools_presentation.go`,
  `tools_reports.go`, `tools_meta.go`, `tools_restores.go`,
  `tools_vms.go`, `tools_networks.go`, `tools_user_management.go`.

---

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