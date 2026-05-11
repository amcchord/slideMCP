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
// tells the LLM about the v4 surface up-front.
func serverInstructions() string {
	return `Slide MCP server v` + Version + `.

You are talking to a Slide-protected MSP backup environment. The fastest path to a useful first answer is usually:

1. Read the resource ` + resourceURIHealth + ` for a one-screen health summary.
2. Read ` + resourceURIInventory + ` for the clients -> devices -> agents tree.

Tool families:
- slide_overview: inventory + health + per-client/per-device summaries (read-only).
- slide_files: file search + restore + push (the headline v4 capability for "find and restore X").
- slide_recovery: boot VMs, export disk images, manage DR networks.
- slide_audit: account audit log queries (compliance / "what changed?").
- slide_clients / slide_admin: client + user/account management.
- slide_devices / slide_agents / slide_snapshots / slide_backups / slide_alerts: lower-level CRUD.

Prompts (slash-command UI):
- slide.daily-status, slide.triage-alerts, slide.restore-file, slide.boot-recovery-vm, slide.dr-runbook.

Mutations (create/update/delete/poweroff/reboot) are gated by the active tools mode (` + ToolsReadOnly + `/` + ToolsSafe + `/` + ToolsFull + `).

Every list-returning operation supports format=summary|compact|detailed and fields=a,b,c projection. Default is summary - opt up to detailed only when you need the full payload.`
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
