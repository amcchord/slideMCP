package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolHandler is the legacy in-process handler signature shared by every
// tools_*.go file. It returns a JSON string body plus an error.
type ToolHandler func(map[string]interface{}) (string, error)

// toolRegistry maps tool names to their in-process handlers. The SDK server
// calls these via adaptToolHandler(); the one-shot --tool CLI mode also calls
// them directly without any MCP transport.
var toolRegistry = map[string]ToolHandler{
	"slide_agents":          handleAgentsTool,
	"slide_backups":         handleBackupsTool,
	"slide_snapshots":       handleSnapshotsTool,
	"slide_restores":        handleRestoresTool,
	"slide_networks":        handleNetworksTool,
	"slide_user_management": handleUserManagementTool,
	"slide_alerts":          handleAlertsTool,
	"slide_devices":         handleDevicesTool,
	"slide_vms":             handleVMsTool,
	"slide_presentation":    handlePresentationTool,
	"slide_reports":         handleReportsTool,
	"slide_meta":            handleMetaTool,
	"slide_docs":            handleDocsTool,
	"list_all_clients_devices_and_agents": func(args map[string]interface{}) (string, error) {
		args["operation"] = "list_all_clients_devices_and_agents"
		return handleMetaTool(args)
	},
}

// allToolInfos returns the legacy ToolInfo for every tool. Filtering happens
// later (in the SDK tool filter) so the descriptors here are unconditional.
func allToolInfos() []ToolInfo {
	return []ToolInfo{
		getAgentsToolInfo(),
		getBackupsToolInfo(),
		getSnapshotsToolInfo(),
		getRestoresToolInfo(),
		getNetworksToolInfo(),
		getUserManagementToolInfo(),
		getAlertsToolInfo(),
		getDevicesToolInfo(),
		getVMsToolInfo(),
		getPresentationToolInfo(),
		getReportsToolInfo(),
		getMetaToolInfo(),
		getDocsToolInfo(),
		{
			Name:        "list_all_clients_devices_and_agents",
			Description: "Get hierarchical view of all clients, devices, and agents. Use when answering questions about infrastructure counts and organization. Returns complete system overview with relationship data.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// toolInfoToSDKTool converts a legacy ToolInfo descriptor into an mcp.Tool
// using NewToolWithRawSchema, preserving the original JSON Schema (including
// conditional allOf/if/then/required blocks).
func toolInfoToSDKTool(info ToolInfo) (mcp.Tool, error) {
	schemaBytes, err := json.Marshal(info.InputSchema)
	if err != nil {
		return mcp.Tool{}, fmt.Errorf("marshal schema for %s: %w", info.Name, err)
	}
	desc := info.Description
	if config != nil && config.ToolsMode == ToolsReporting {
		desc += " (Read-only mode: only list/get operations available)"
	}
	return mcp.NewToolWithRawSchema(info.Name, desc, schemaBytes), nil
}

// adaptToolHandler wraps a legacy ToolHandler so it satisfies the SDK's
// ToolHandlerFunc shape. Permission checks mirror the original handleToolCall()
// dispatcher in server.go.
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
		return mcp.NewToolResultText(text), nil
	}
}

// registerTools wires every tool descriptor + handler onto the SDK server.
// Tool filtering (per current ToolsMode / disabled list) happens in
// toolFilterForMode() so tools/list always reflects the live configuration.
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
