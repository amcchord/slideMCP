package main

// Tool handler type for cleaner dispatch
type ToolHandler func(map[string]interface{}) (string, error)

// Global tool registry for cleaner dispatch
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
		// Redirect to slide_meta for backward compatibility
		args["operation"] = "list_all_clients_devices_and_agents"
		return handleMetaTool(args)
	},
}

// getAllTools returns all available tools filtered by current configuration
func getAllTools() []ToolInfo {
	allTools := []ToolInfo{
		// Meta-tools
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
		getDocsToolInfo(), // Documentation access tool
		// Special tools (kept for backward compatibility)
		{
			Name:        "list_all_clients_devices_and_agents",
			Description: "Get a complete hierarchical view of all clients, their devices, and the agents on those devices. Use this tool when answers questions about how many agents, devices, or clients",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	// Filter tools based on current tools mode
	var filteredTools []ToolInfo
	for _, tool := range allTools {
		if config.IsToolAllowed(tool.Name) {
			// For reporting mode, update tool descriptions to indicate read-only access
			if config.ToolsMode == ToolsReporting {
				tool.Description = tool.Description + " (Read-only mode: only list/get operations available)"
			}
			filteredTools = append(filteredTools, tool)
		}
	}

	return filteredTools
}
