package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
)

// buildMCPServer constructs the SDK-backed MCP server, registers every
// tool, prompt, and resource, and installs the mode-aware tool filter.
func buildMCPServer() (*server.MCPServer, error) {
	srv := server.NewMCPServer(
		ServerName,
		Version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
		server.WithInstructions(serverInstructions()),
		server.WithToolFilter(toolFilterForMode()),
	)

	if err := registerTools(srv); err != nil {
		return nil, fmt.Errorf("register tools: %w", err)
	}
	registerResources(srv)
	registerPrompts(srv)

	return srv, nil
}

// runStdioServer wires the MCP server to stdio. Blocks until stdin closes.
func runStdioServer() error {
	srv, err := buildMCPServer()
	if err != nil {
		return err
	}
	log.Printf("%s %s ready on stdio (mode=%s)", ServerName, Version, config.ToolsMode)
	return server.ServeStdio(srv)
}

// serverInstructions is shown to clients in the initialize response and
// tells the LLM about the v5 surface up-front. The FIRST sentence is the
// trigger-vocabulary list - it's the single highest-leverage place an MCP
// host shows the model context at conversation start, so we use it to
// loudly route the model toward this server on first-mention.
func serverInstructions() string {
	mode := ""
	if config != nil {
		mode = config.ToolsMode
	}
	modeLine := "Active permission tier: " + mode + "."
	switch mode {
	case ToolsReadOnly:
		modeLine += " You can list, get, search, browse, and triage. You cannot start backups, create restores, boot VMs, or delete anything."
	case ToolsSafe:
		modeLine += " You can read everything, start backups, create restores, boot recovery VMs, update agent/device/network settings, and resolve alerts. Deletes and device poweroff/reboot are blocked."
	case ToolsFull:
		modeLine += " You can do everything, including deletes and device poweroff/reboot. Ask the user before destructive operations."
	}

	return `Slide MCP server v` + Version + `.

USE THIS SERVER WHENEVER THE USER MENTIONS any of: Slide, slide.tech, a Slide box / Slide device / Slide appliance, BCDR, business continuity, disaster recovery (DR), DR runbook, failover, recovery point (RPO/RTO), backup, backups, "backed up", "did backups run", backup failed, backup schedule, pause/resume backups, snapshot, restore point, restore, recover, "recover a file", roll back, yesterday's copy, previous version, version history, recovery VM, "boot a VM from a snapshot", image export (VHD/VHDX/VMDK/QCOW2/RAW), RDP into a recovered server, DR network, VPN to a recovered VM, Slide alert, unresolved alert, triage alerts, "are all my Slide boxes OK", "is everything healthy", "what changed last night", audit log, an MSP client, a Slide agent on a protected server, a protected system, "find <file> on <person>'s laptop", or "push it back to the server". If you see any of these phrases, reach for a slide-mcp-server tool before asking the user clarifying questions.

Glossary (Slide-specific - terms below are always interpreted in the Slide sense, not the generic IT sense):
- client      = an MSP customer / organisational unit. ID prefix c_xxxxxxxxxxxx.
- device      = a physical Slide backup appliance (the "Slide box"). ID prefix d_xxxxxxxxxxxx.
- agent       = a backup agent installed on a protected server / endpoint. ID prefix a_xxxxxxxxxxxx.
- snapshot    = a point-in-time backup (restore point). ID prefix s_xxxxxxxxxxxx.
- restore     = a session that mounts files from a snapshot so they can be browsed or pushed back.
- VM          = a recovery VM booted from a snapshot on a Slide device. ID prefix v_xxxxxxxxxxxx.
- alert       = an unresolved condition raised by a device or agent (storage low, backup failed, etc.).

` + modeLine + `

When you don't know where to start, call slide_help operation=getting_started - it returns a concise tour of every capability for the current permission tier, plus example questions you can use to suggest follow-ups.

Cheap context you can read without burning a tool call:
- ` + resourceURIWelcome + ` - one-page primer with example questions.
- ` + resourceURIHealth + ` - one-screen health summary for every device + agent.
- ` + resourceURIInventory + ` - clients -> devices -> agents tree.
- ` + resourceURIAlertsOpen + ` - unresolved alerts, prioritised.
- ` + resourceURIHelpGlossary + ` - the glossary above, expanded.

Tool families:
- slide_help     -> discovery, examples, troubleshooting. Read-only, always available. Call this first if the user is vague.
- slide_overview -> inventory + health + per-client/per-device summaries. Read-only.
- slide_files    -> file search + restore + push (the headline "find and restore X" workflow).
- slide_recovery -> boot recovery VMs, export disk images, manage DR networks.
- slide_audit    -> account audit log queries (compliance / "what changed?").
- slide_clients / slide_admin                                -> client + user/account management.
- slide_devices / slide_agents / slide_snapshots / slide_backups / slide_alerts -> lower-level CRUD with task-oriented additions (triage, status_for_client, recent_for_agent).

Prompts (slash-command UI):
- slide.welcome          -> one-message intro for first-time users.
- slide.daily-status     -> daily ops summary across the account.
- slide.triage-alerts    -> prioritised unresolved-alert review.
- slide.restore-file     -> guided file recovery walkthrough.
- slide.boot-recovery-vm -> guided VM recovery walkthrough.
- slide.dr-runbook       -> tailored DR runbook for a device.

Worked examples (user question -> first tool call you should make):
- "Are all my Slide boxes healthy?"                        -> slide_overview operation=health
- "Did backups run last night for ACME?"                   -> slide_backups operation=status_for_client client_id=...
- "Find Q4-budget.xlsx on Bob's laptop"                    -> slide_files operation=search name_hint=Bob search_term=Q4-budget
- "I need yesterday's version of <file>"                   -> slide_files operation=versions ...
- "Boot a recovery VM for <server>"                        -> slide_recovery operation=boot_vm ...
- "What unresolved alerts do I have?"                      -> slide_alerts operation=triage
- "What changed in the last 24 hours?"                     -> slide_audit operation=recent

Identifying agents/devices/clients: every tool that needs an *_id ALSO accepts name_hint. Use name_hint when the user gives you a hostname, display name, or partial match ("Bob's laptop", "ACME", "DC-01"). The server resolves it server-side and returns the resolved object in the response. If multiple matches exist you'll get a structured "ambiguous" error with candidates - relay those to the user and re-call with the chosen id.

Every list-returning operation supports format=summary|compact|detailed and fields=a,b,c projection. Default is summary - opt up to detailed only when you need the full payload. Pass hints=off if you want to suppress the next_steps array we append to most responses.`
}

// runOneShotTool runs a single tool by name with the supplied JSON arguments
// and writes the result to stdout. Used by the --tool / --args CLI mode.
func runOneShotTool(name string, args map[string]interface{}) error {
	if args == nil {
		args = map[string]interface{}{}
	}
	if config.IsToolDisabled(name) {
		return fmt.Errorf("tool '%s' is disabled", name)
	}
	if !config.IsToolAllowed(name) {
		return fmt.Errorf("tool '%s' not available in '%s' mode", name, config.ToolsMode)
	}
	handler, ok := toolRegistry[name]
	if !ok {
		return fmt.Errorf("unknown tool: %s", name)
	}
	result := createToolResult(handler(args))
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	fmt.Println(string(out))
	return nil
}
