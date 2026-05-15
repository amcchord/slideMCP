package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolHandler is the in-process handler signature shared by every
// tools_*.go file. It returns a JSON string body plus an error.
type ToolHandler func(map[string]interface{}) (string, error)

// toolRegistry maps tool names to their in-process handlers. The SDK server
// calls these via adaptToolHandler(); the one-shot --tool CLI mode also
// calls them directly without any MCP transport.
//
// The v4.0.0 surface is task-oriented: ten meta-tools an MSP tech can
// reach for naturally, plus the bare list_all shim that older Desktop
// installs may still call.
var toolRegistry = map[string]ToolHandler{
	// v5 discovery / onboarding tool. Always available, regardless of
	// tools mode, so a novice can self-rescue.
	"slide_help": handleHelpTool,

	// v4 task-oriented tools
	"slide_overview": handleOverviewTool,
	"slide_files":    handleFilesTool,
	"slide_recovery": handleRecoveryTool,
	"slide_audit":    handleAuditTool,
	"slide_clients":  handleClientsTool,
	"slide_admin":    handleAdminTool,

	// Refined v4 versions of v3 tools
	"slide_devices":   handleDevicesTool,
	"slide_agents":    handleAgentsTool,
	"slide_snapshots": handleSnapshotsTool,
	"slide_backups":   handleBackupsTool,
	"slide_alerts":    handleAlertsTool,

	// Backward-compat shim (delegates to slide_overview inventory)
	"list_all_clients_devices_and_agents": func(args map[string]interface{}) (string, error) {
		return listAllClientsDevicesAndAgents(args)
	},
}

// allToolInfos returns the legacy ToolInfo for every tool. Filtering happens
// later (in the SDK tool filter) so the descriptors here are unconditional.
func allToolInfos() []ToolInfo {
	return []ToolInfo{
		getHelpToolInfo(),
		getOverviewToolInfo(),
		getFilesToolInfo(),
		getRecoveryToolInfo(),
		getAuditToolInfo(),
		getClientsToolInfo(),
		getAdminToolInfo(),
		getDevicesToolInfo(),
		getAgentsToolInfo(),
		getSnapshotsToolInfo(),
		getBackupsToolInfo(),
		getAlertsToolInfo(),
		{
			Name: "list_all_clients_devices_and_agents",
			Description: "Slide MCP - backward-compat alias. " +
				"REACH FOR THIS whenever the user mentions Slide inventory, 'list my Slide clients/devices/agents', " +
				"or the v3 question 'what's in my Slide account'. New conversations should prefer `slide_overview operation=inventory` " +
				"(which this is just a thin alias for).",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// toolInfoToSDKTool converts a legacy ToolInfo descriptor into an mcp.Tool
// using NewToolWithRawSchema, preserving the original JSON Schema (including
// conditional allOf/if/then/required blocks). Adds tool annotations
// (readOnlyHint / destructiveHint / idempotentHint / openWorldHint) so
// Claude Desktop renders confirmation prompts only when actually needed.
func toolInfoToSDKTool(info ToolInfo) (mcp.Tool, error) {
	schemaBytes, err := json.Marshal(info.InputSchema)
	if err != nil {
		return mcp.Tool{}, fmt.Errorf("marshal schema for %s: %w", info.Name, err)
	}
	desc := info.Description
	if config != nil && config.ToolsMode == ToolsReadOnly {
		desc += " (Read-only mode: only list/get/search operations available.)"
	}
	t := mcp.NewToolWithRawSchema(info.Name, desc, schemaBytes)
	t.Annotations = annotationsForTool(info.Name)
	if outSchema := outputSchemaForTool(info.Name); len(outSchema) > 0 {
		t.RawOutputSchema = outSchema
	}
	return t, nil
}

// adaptToolHandler wraps an in-process ToolHandler so it satisfies the SDK's
// ToolHandlerFunc shape. Permission checks mirror handleToolCall().
//
// When the tool declares an outputSchema (via outputSchemaForTool), the
// MCP 2025-11-25 spec requires the response to include `structuredContent`
// matching that schema in addition to the text content. Without it,
// Claude.ai's tool-execution layer rejects the response as a schema
// violation and surfaces a generic "Tool execution failed" to the LLM,
// even though the raw JSON-RPC response left the server cleanly.
//
// Every handler in this codebase already produces a JSON string, so we
// parse it back into a map at the dispatcher boundary and hand BOTH the
// original text (as fallback content) and the parsed map (as
// structuredContent) to the SDK. If the parse fails (which should never
// happen for our handlers), we fall back to text-only - still valid per
// spec, just won't satisfy outputSchema clients.
func adaptToolHandler(name string, handler ToolHandler) server.ToolHandlerFunc {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if config.IsToolDisabled(name) {
			return mcp.NewToolResultErrorf("tool '%s' is disabled", name), nil
		}
		if !config.IsToolAllowed(name) {
			return mcp.NewToolResultErrorf("tool '%s' not available in '%s' mode", name, config.ToolsMode), nil
		}
		args := req.GetArguments()
		if args == nil {
			args = map[string]any{}
		}
		text, err := handler(args)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return toolResultWithStructured(text), nil
	}
}

// toolResultWithStructured returns a CallToolResult populated with BOTH
// the original text content AND the parsed JSON as structuredContent.
// On parse failure (text isn't valid JSON), returns text-only.
func toolResultWithStructured(text string) *mcp.CallToolResult {
	var structured interface{}
	if err := json.Unmarshal([]byte(text), &structured); err == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: text},
			},
			StructuredContent: structured,
		}
	}
	return mcp.NewToolResultText(text)
}

// registerTools wires every tool descriptor + handler onto the SDK server.
func registerTools(s *server.MCPServer) error {
	infos := allToolInfos()
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	for _, info := range infos {
		handler, ok := toolRegistry[info.Name]
		if !ok {
			return fmt.Errorf("no handler registered for tool %s", info.Name)
		}
		sdkTool, err := toolInfoToSDKTool(info)
		if err != nil {
			return err
		}
		s.AddTool(sdkTool, adaptToolHandler(info.Name, handler))
	}
	return nil
}

// toolFilterForMode hides tools that are not allowed in the current mode so
// they do not appear in tools/list.
func toolFilterForMode() server.ToolFilterFunc {
	return func(_ context.Context, tools []mcp.Tool) []mcp.Tool {
		filtered := make([]mcp.Tool, 0, len(tools))
		for _, t := range tools {
			if config.IsToolAllowed(t.Name) {
				filtered = append(filtered, t)
			}
		}
		return filtered
	}
}
