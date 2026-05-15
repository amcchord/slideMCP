package main

// MCP Prompts - reusable templated workflows that Claude Desktop surfaces
// in its slash-command UI. Each prompt seeds the conversation with a
// system-level instruction grounded in the user's own data; the model
// then drives the actual tool calls.
//
// Wired via srv.AddPrompt() in server.go. Five prompts in the v4 surface:
//
//   /slide.daily-status [client]      - daily ops summary
//   /slide.triage-alerts              - prioritised unresolved-alert review
//   /slide.restore-file               - guided file recovery walkthrough
//   /slide.boot-recovery-vm           - guided VM recovery walkthrough
//   /slide.dr-runbook  [device]       - DR runbook for a device

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerPrompts attaches every MCP Prompt to the server.
func registerPrompts(s *server.MCPServer) {
	s.AddPrompt(promptWelcome(), handlePromptWelcome)
	s.AddPrompt(promptDailyStatus(), handlePromptDailyStatus)
	s.AddPrompt(promptTriageAlerts(), handlePromptTriageAlerts)
	s.AddPrompt(promptRestoreFile(), handlePromptRestoreFile)
	s.AddPrompt(promptBootRecoveryVM(), handlePromptBootRecoveryVM)
	s.AddPrompt(promptDRRunbook(), handlePromptDRRunbook)
}

// --- welcome ----------------------------------------------------------

func promptWelcome() mcp.Prompt {
	return mcp.NewPrompt("slide.welcome",
		mcp.WithPromptDescription("One-message intro for first-time users of the Slide MCP extension. Lists what the server can do, the trigger vocabulary that should route to Slide, and 5 example questions to try. No arguments."),
	)
}

func handlePromptWelcome(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	mode := ""
	if config != nil {
		mode = config.ToolsMode
	}
	body := `You are connected to slide-mcp-server, the MCP extension for the Slide BCDR / backup platform. ` +
		`Greet the operator with a concise, friendly intro that covers:

1. ONE-LINE PURPOSE. "Slide is the BCDR / backup platform that protects servers and endpoints; this extension lets you ask Claude about your Slide environment, run restores, boot recovery VMs, and triage alerts."

2. WHEN TO REACH FOR ME. Say something like: "If you mention Slide, BCDR, backups, snapshots, restores, recovery VMs, file recovery, disaster recovery, audit logs, or unresolved alerts, I will use this extension automatically."

3. PERMISSION TIER. The server is currently running in the ` + "`" + mode + "`" + ` tier. Briefly explain what that allows.

4. FIVE EXAMPLE QUESTIONS the operator can try right now. Pick from this list and pre-fill them as bulleted suggestions:
   - "Are all my Slide boxes healthy?"
   - "Did backups run last night for <client name>?"
   - "Find <filename> on <person>'s laptop."
   - "What unresolved alerts do I have? Sort worst first."
   - "What changed in the last 24 hours?"

5. FALLBACK. "If you're not sure where to start, just say 'show me what you can do' and I'll call slide_help getting_started."

Do not call any tools yet. The next user message will say what they actually want.`

	return mcp.NewGetPromptResult(
		"Welcome to the Slide MCP extension",
		[]mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body))},
	), nil
}

// --- daily-status ------------------------------------------------------

func promptDailyStatus() mcp.Prompt {
	return mcp.NewPrompt("slide.daily-status",
		mcp.WithPromptDescription("Generate a daily ops summary across the account, optionally scoped to one client. Walks alerts, last-24h backups, stale agents."),
		mcp.WithArgument("client", mcp.ArgumentDescription("Optional client name or client_id to scope the report.")),
		mcp.WithArgument("hours", mcp.ArgumentDescription("Hours of history to include (default 24).")),
	)
}

func handlePromptDailyStatus(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	client := strings.TrimSpace(req.Params.Arguments["client"])
	hours := strings.TrimSpace(req.Params.Arguments["hours"])
	if hours == "" {
		hours = "24"
	}
	scope := "the whole account"
	if client != "" {
		scope = "client " + client
	}

	body := fmt.Sprintf(`Produce a daily ops summary for %s covering the last %s hours.

Steps:
1. Call slide_overview operation=health to get current health.
2. If a client was named, resolve it via slide_clients list (match the name) then call slide_overview for_client.
3. Call slide_alerts triage to get unresolved alerts grouped by severity.
4. Call slide_backups status_for_client (or status_for_device for each device) for the time window.

Then write a concise summary in this exact structure:

## Daily Status: %s

### Health
- Total devices/agents seen, healthy / stale / unknown counts.

### Open alerts
- Critical: ...
- High: ...
(omit medium/low unless critical/high are empty)

### Backups (last %s hours)
- Successful / failed / in-progress totals.
- List any agent that had a failure (with the error message).

### Action items
- 1-3 concrete next steps the operator should take, in priority order.

Be terse. No filler. Use the data, do not invent it.`, scope, hours, scope, hours)

	return mcp.NewGetPromptResult(
		"Daily Slide ops summary",
		[]mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body))},
	), nil
}

// --- triage-alerts -----------------------------------------------------

func promptTriageAlerts() mcp.Prompt {
	return mcp.NewPrompt("slide.triage-alerts",
		mcp.WithPromptDescription("Review unresolved alerts in priority order and propose remediation for each."),
	)
}

func handlePromptTriageAlerts(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	body := `Walk the unresolved Slide alerts and propose action.

Steps:
1. Call slide_alerts triage to get alerts sorted by severity score.
2. For each critical or high severity alert (cap at 5):
   - Identify the affected device or agent (use slide_overview for_device or slide_overview for_client if needed for context).
   - Briefly explain what the alert_type means.
   - Propose a concrete action.
3. Group medium and low severity alerts as a single line summary.

Output format:

### Critical
- <alert summary> on <device/agent name>: <action>
### High
...
### Medium / Low
- <count> alerts; nothing immediately actionable.

Do not call slide_alerts update unless the operator explicitly asks you to resolve something.`
	return mcp.NewGetPromptResult(
		"Triage unresolved alerts",
		[]mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body))},
	), nil
}

// --- restore-file ------------------------------------------------------

func promptRestoreFile() mcp.Prompt {
	return mcp.NewPrompt("slide.restore-file",
		mcp.WithPromptDescription("Guided walkthrough for finding and restoring a single file from a snapshot."),
		mcp.WithArgument("filename", mcp.ArgumentDescription("File name or substring to search for.")),
		mcp.WithArgument("agent", mcp.ArgumentDescription("Agent name, hostname, or agent_id (will be resolved if not an exact ID).")),
	)
}

func handlePromptRestoreFile(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	filename := strings.TrimSpace(req.Params.Arguments["filename"])
	agent := strings.TrimSpace(req.Params.Arguments["agent"])

	body := fmt.Sprintf(`Help the operator restore a file from a Slide snapshot.

Inputs (may be empty - ask the operator):
- filename: %q
- agent:    %q

Steps:
1. If the agent is empty, ask the operator which agent to search.
2. Resolve the agent: if it does not look like an a_xxxxxxxxxxxx ID, call slide_overview inventory and match by display_name or hostname (case-insensitive substring).
3. Call slide_files search agent_id=<id> search_term=<filename>.
4. Show the matches; ask the operator which path to restore (if multiple).
5. Call slide_files versions agent_id=<id> path=<chosen path> to list snapshot versions.
6. Help pick a snapshot (newest is usually right; flag any that look stale).
7. Confirm with the operator before mutating anything, then:
   - Call slide_files create_restore snapshot_id=<id> device_id=<id from snapshot>.
   - Either guide the operator to download via the resulting download URIs, or call slide_files create_push to push back to the protected system at C:\\SlideRestore.
8. Stop and report the restore_id, push_id (if any), and download URIs.

Be cautious with create_push: confirm destination_folder explicitly and remind the operator the file lands under <drive>:\\SlideRestore.`, filename, agent)

	return mcp.NewGetPromptResult(
		"Guided file restore",
		[]mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body))},
	), nil
}

// --- boot-recovery-vm --------------------------------------------------

func promptBootRecoveryVM() mcp.Prompt {
	return mcp.NewPrompt("slide.boot-recovery-vm",
		mcp.WithPromptDescription("Guided walkthrough to boot a recovery VM from a snapshot."),
		mcp.WithArgument("agent", mcp.ArgumentDescription("Agent name, hostname, or agent_id.")),
	)
}

func handlePromptBootRecoveryVM(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	agent := strings.TrimSpace(req.Params.Arguments["agent"])

	body := fmt.Sprintf(`Help the operator boot a recovery VM from a Slide snapshot.

Input:
- agent: %q

Steps:
1. Resolve the agent: if not an a_xxxxxxxxxxxx ID, use slide_overview inventory to match by name/hostname.
2. Call slide_snapshots recent_for_agent to get recent snapshots; help the operator pick one (newest healthy one is usually right).
3. Confirm with the operator the device that should host the VM (defaults to the device that created the snapshot).
4. Call slide_recovery boot_vm snapshot_id=<id> device_id=<id> with sensible defaults (cpu_count=4, memory_in_mb=8192, disk_bus=virtio, network_model=virtio, network_type=network).
5. Once the VM is created, call slide_recovery get_vm to confirm it's running, then call slide_recovery get_rdp_bookmark to give the operator an RDP URI.
6. Optionally walk through DR-network setup if the operator needs the VM reachable from outside.
7. Stop and report the virt_id, VM state, and any access URIs.

Always confirm before booting; this consumes resources on the device.`, agent)

	return mcp.NewGetPromptResult(
		"Guided VM recovery",
		[]mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body))},
	), nil
}

// --- dr-runbook --------------------------------------------------------

func promptDRRunbook() mcp.Prompt {
	return mcp.NewPrompt("slide.dr-runbook",
		mcp.WithPromptDescription("Generate a tailored disaster-recovery runbook for a device, grounded in its current agents and snapshots."),
		mcp.WithArgument("device", mcp.ArgumentDescription("Device name, hostname, or device_id.")),
	)
}

func handlePromptDRRunbook(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	device := strings.TrimSpace(req.Params.Arguments["device"])

	body := fmt.Sprintf(`Generate a disaster-recovery runbook for the named Slide-protected device.

Input:
- device: %q

Steps:
1. Resolve the device: if not d_xxxxxxxxxxxx, use slide_overview inventory to match by name/hostname.
2. Call slide_overview for_device to get the device + agents + open alerts snapshot.
3. For each agent, call slide_snapshots recent_for_agent (last 7 days) to confirm there are valid restore points.
4. Synthesize a runbook in markdown with the following sections:

# DR Runbook: <device name>
Generated: <RFC3339>

## Inventory
- Device: <name> (<id>) - last seen <ts>, status <status>
- Agents protected:
  - <name> (<id>) - <os>, last seen <ts>, latest snapshot <ts>

## Open issues to resolve before DR is invoked
- <list each open alert with severity hint and resolution>

## DR procedure
For each agent:
1. Boot a recovery VM via slide_recovery boot_vm using the most recent verified snapshot.
2. Verify the VM via the VNC console / RDP bookmark.
3. (Optional) Set up a DR network so the VM is reachable.
4. (Optional) Push individual files back via slide_files create_push.

## Known constraints
- Tools mode currently in effect; some destructive operations may require switching to 'full' mode.
- Image exports are large; budget time for download.

## Validation cadence
- Recommended: boot one snapshot per agent every 30 days as a fire drill.

Keep the runbook concrete; reference real IDs from the data, not placeholders.`, device)

	return mcp.NewGetPromptResult(
		"Generate DR runbook for a device",
		[]mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body))},
	), nil
}
