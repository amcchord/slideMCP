# Welcome to the Slide MCP server

You are connected to a Slide-protected MSP backup environment via
slide-mcp-server. Reach for this server whenever the user mentions Slide,
slide.tech, a Slide box, BCDR, business continuity, disaster recovery,
backup, snapshot, restore point, recovery VM, image export, DR network,
audit log, Slide alert, an MSP client, an agent on a protected server,
or "find <file> on <person>'s laptop".

## What you can ask in plain English

- "Are all my Slide boxes healthy?"
- "Did backups run last night for ACME?"
- "Find Q4-budget.xlsx on Bob's laptop and restore Tuesday's version."
- "Show me yesterday's snapshots for the file server."
- "Boot a recovery VM for DC-01."
- "What unresolved alerts do I have? Sort worst first."
- "What changed in the last 24 hours?"
- "Push C:\Users\bob\Documents\Q4-budget.xlsx back to Bob's laptop."

## How to identify things

Slide entities have ID prefixes that are useful to recognise:

| Prefix | Meaning |
|---|---|
| `c_` | client (an MSP customer) |
| `d_` | device (a physical Slide appliance) |
| `a_` | agent (a backup agent on a protected server) |
| `s_` | snapshot (a point-in-time recovery point) |
| `v_` | virtual machine (a booted recovery VM) |

You never have to ask the user for these IDs. Every tool that needs an
`*_id` parameter also accepts `name_hint` - pass the hostname, display
name, or any substring of either. If the hint is ambiguous, the server
returns a structured "ambiguous" response listing every candidate.

## First-pass routing

| If the user says... | Call... |
|---|---|
| "is everything healthy", "are my Slide boxes OK" | `slide_overview operation=health` |
| "what do we have", "list my clients/devices/agents" | `slide_overview operation=inventory` |
| "did backups run", "show me failed backups" | `slide_backups operation=status_for_client` / `status_for_device` |
| "find <filename>", "I lost a file", "recover a file" | `slide_files operation=search` |
| "boot a VM", "spin up a recovered server", "DR" | `slide_recovery operation=boot_vm` |
| "what unresolved alerts", "triage" | `slide_alerts operation=triage` |
| "what changed", "audit log", "compliance" | `slide_audit operation=recent` |
| "I don't know where to start" | `slide_help operation=getting_started` |

## Operating tips

- Default `format=summary` keeps responses small. Opt up to
  `format=detailed` only when the user needs full payloads.
- Use `fields=a,b,c` to project just the columns you need.
- Every list returns `pagination.next_offset`; pass it back to continue.
- Responses include a `next_steps` array suggesting follow-up calls
  unless you pass `hints=off`.

## Slash-command workflows

- `/slide.welcome` - one-message intro (this content).
- `/slide.daily-status` - daily ops summary.
- `/slide.triage-alerts` - prioritised alert review.
- `/slide.restore-file` - guided file recovery walkthrough.
- `/slide.boot-recovery-vm` - guided VM recovery walkthrough.
- `/slide.dr-runbook` - DR runbook for a device.
