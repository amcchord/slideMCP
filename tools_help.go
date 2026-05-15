package main

// slide_help: the discovery / onboarding tool. Read-only, always
// available (never blocked by tools mode), and intended to be the
// first thing the LLM calls when the user's request is too vague to
// route directly to a domain meta-tool.
//
// Operations are deliberately a small, finite set so the LLM can pick
// one without re-checking the schema:
//
//   getting_started   - one-page primer + active permission tier
//   examples          - copy-pasteable example user questions
//   glossary          - Slide-specific terminology
//   troubleshoot      - common setup / auth / connectivity problems
//   list_prompts      - the slash-command prompts the server publishes
//   list_resources    - the slide:// resources the server publishes
//   what_can_you_do   - terse capability summary tagged by permission
//
// The bodies live in docs/help/*.md and are baked into the binary via
// //go:embed; no network is touched.

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed docs/help/welcome.md
var helpWelcomeMD string

//go:embed docs/help/glossary.md
var helpGlossaryMD string

//go:embed docs/help/troubleshoot.md
var helpTroubleshootMD string

//go:embed docs/help/examples.md
var helpExamplesMD string

func handleHelpTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_help", ToolOperations{
		"getting_started":  handleHelpGettingStarted,
		"examples":         handleHelpExamples,
		"glossary":         handleHelpGlossary,
		"troubleshoot":     handleHelpTroubleshoot,
		"list_prompts":     handleHelpListPrompts,
		"list_resources":   handleHelpListResources,
		"what_can_you_do":  handleHelpWhatCanYouDo,
		"debug":            handleHelpDebug,
	}), args)
}

var helpOperationEnums = []string{
	"getting_started", "examples", "glossary", "troubleshoot",
	"list_prompts", "list_resources", "what_can_you_do", "debug",
}

func getHelpToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        helpOperationEnums,
		},
	}
	return ToolInfo{
		Name: "slide_help",
		Description: "Slide MCP - discovery, onboarding, and troubleshooting. " +
			"REACH FOR THIS whenever the user is vague ('can you help me with backups?', 'what can you do?', " +
			"'I don't know where to start', 'how do I restore something?'), whenever you need to remind yourself " +
			"of the canonical first tool call for a question, whenever the user reports a setup / authentication " +
			"problem with slide-mcp-server, or whenever ANY tool call fails and you need to understand why. " +
			"Read-only and always available, even in read-only permission mode. " +
			"Operations: `getting_started` (one-page primer + active permission tier), `examples` (copy-paste " +
			"example user questions paired with first tool calls), `glossary` (Slide-specific terminology), " +
			"`troubleshoot` (token / network / permission diagnostics), `list_prompts` and `list_resources` " +
			"(what slash-commands and slide:// URIs the server exposes), `what_can_you_do` (terse capability " +
			"summary tagged by permission tier), `debug` (FULL diagnostic dump: version, runtime, config, env, " +
			"DNS, TLS, live probes against /v1/account /v1/client /v1/device /v1/agent with HTTP status + body " +
			"excerpts, name_resolver cache state, recent stderr logs - paste-safe; api_key is masked). " +
			"If any tool call returns isError=true or you see a generic 'Tool execution failed', " +
			"immediately call slide_help operation=debug so the user can see what's actually wrong.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
		},
	}
}

// --- handlers ---------------------------------------------------------

func handleHelpGettingStarted(_ map[string]interface{}) (string, error) {
	mode := ""
	if config != nil {
		mode = config.ToolsMode
	}
	prefix := fmt.Sprintf("slide-mcp-server v%s, permission tier=%s\n\n", Version, mode)
	return prefix + helpWelcomeMD, nil
}

func handleHelpExamples(_ map[string]interface{}) (string, error) {
	return helpExamplesMD, nil
}

func handleHelpGlossary(_ map[string]interface{}) (string, error) {
	return helpGlossaryMD, nil
}

func handleHelpTroubleshoot(_ map[string]interface{}) (string, error) {
	return helpTroubleshootMD, nil
}

// helpPromptEntry mirrors what registerPrompts wires up so we don't
// drift from the actual surface. Update both together.
type helpPromptEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func handleHelpListPrompts(_ map[string]interface{}) (string, error) {
	prompts := []helpPromptEntry{
		{"slide.welcome", "One-message intro for first-time users. No arguments."},
		{"slide.daily-status", "Daily ops summary across the account. Optional client + hours."},
		{"slide.triage-alerts", "Prioritised unresolved-alert review."},
		{"slide.restore-file", "Guided file-recovery walkthrough. Optional filename + agent."},
		{"slide.boot-recovery-vm", "Guided VM recovery walkthrough. Optional agent."},
		{"slide.dr-runbook", "Generates a DR runbook tailored to one device's real agents and snapshots."},
	}
	out := map[string]interface{}{
		"prompts": prompts,
		"hint":    "Surface these to the user via Claude Desktop's slash-command UI. They seed the conversation with a system message you then drive.",
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type helpResourceEntry struct {
	URI         string `json:"uri"`
	Description string `json:"description"`
}

func handleHelpListResources(_ map[string]interface{}) (string, error) {
	resources := []helpResourceEntry{
		{resourceURIWelcome, "One-page primer with example questions; no API calls. Cheap, safe to read at conversation start."},
		{resourceURIHelpGlossary, "Slide-specific glossary."},
		{resourceURIHelpTroubleshoot, "Setup / auth / connectivity troubleshooting."},
		{resourceURIInventory, "Full clients -> devices -> agents tree (live data)."},
		{resourceURIHealth, "One-line-per-device-and-agent health summary (live data, 30m stale cutoff)."},
		{resourceURIAlertsOpen, "Unresolved alerts, prioritised by severity hint."},
		{resourceURIAuditRecent, "Last 24h of audit log."},
		{resourceURIDocsOpenAPI, "Live Slide OpenAPI spec (cached 1h)."},
		{resourceURITplClient, "URI template: slide://client/{client_id}."},
		{resourceURITplDevice, "URI template: slide://device/{device_id}."},
		{resourceURITplAgent, "URI template: slide://agent/{agent_id}."},
		{resourceURITplAgentRecents, "URI template: slide://agent/{agent_id}/snapshots/recent."},
	}
	out := map[string]interface{}{
		"resources": resources,
		"hint":      "Resources are read-only; prefer them over tool calls when you just need context.",
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// handleHelpDebug returns the full diagnostic bundle (same payload as the
// --debug CLI subcommand) as pretty-printed JSON. Read-only; safe to call
// from any permission tier. The api_key field is masked; no secrets are
// emitted, so the LLM can echo the entire response back to the user.
func handleHelpDebug(_ map[string]interface{}) (string, error) {
	info := gatherDebugInfo()
	out, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal debug info: %w", err)
	}
	return string(out), nil
}

func handleHelpWhatCanYouDo(_ map[string]interface{}) (string, error) {
	mode := ""
	if config != nil {
		mode = config.ToolsMode
	}

	always := []string{
		"Inventory + health summary of every client, device, and agent in the Slide account.",
		"File search across every snapshot for an agent ('find Q4-budget on Bob's laptop').",
		"List snapshot versions of a path, browse a created restore session.",
		"Triage unresolved alerts in severity order.",
		"Query the Slide audit log for compliance / 'what changed?' investigations.",
		"List clients, devices, agents, snapshots, backup runs, users, accounts.",
	}
	safeOnly := []string{
		"Kick off on-demand backups, create/manage restore sessions, push files back to protected systems.",
		"Boot recovery VMs from snapshots, export disk images (VHD/VHDX/VMDK/QCOW2/RAW), generate RDP bookmarks.",
		"Create and update DR networks, IPSec connections, port forwards, and WireGuard peers.",
		"Update agent backup schedules, retention policies, restore defaults, volume settings, alert configs.",
		"Update device hostnames / display names / network configuration. Create or modify VLANs.",
		"Resolve / re-open alerts; update account alert email recipients.",
	}
	fullOnly := []string{
		"Delete agents, clients, snapshots, restore sessions, VMs, image exports, DR networks, VLANs, port forwards, IPSec peers, WireGuard peers.",
		"Remotely power off or reboot a Slide device.",
	}

	out := map[string]interface{}{
		"permission_tier": mode,
		"always_available": map[string]interface{}{
			"description": "Available in every permission tier (read-only, safe, full).",
			"capabilities": always,
		},
		"requires_safe_or_full": map[string]interface{}{
			"description": "Available in `safe` (the default) and `full`. Blocked in `read-only`.",
			"capabilities": safeOnly,
		},
		"requires_full": map[string]interface{}{
			"description": "Available only in `full` (irreversible / power-cycle operations).",
			"capabilities": fullOnly,
		},
		"note": fmt.Sprintf("Current tier is %q. Change it in Claude Desktop -> Settings -> Extensions -> Slide Backup -> Tool permissions.", mode),
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
