package main

import (
	"fmt"
)

// handleBackupsTool handles all backup-related operations through a single meta-tool
func handleBackupsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_backups", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_backups in '%s' mode", operation, toolsMode)
	}

	switch operation {
	case "list":
		return listBackups(args)
	case "get":
		return getBackup(args)
	case "start":
		return startBackup(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getBackupsToolInfo returns the tool definition for the backups meta-tool
func getBackupsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_backups",
		Description: "Manage backup operations - view backup status and start new backups. Supports list, get, and start operations.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get", "start"},
				},
				// Parameters for list operation
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with 'list' operation",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with 'list' operation",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by agent ID - used with 'list' operation, or required for 'start' operation",
				},
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by device ID - used with 'list' operation",
				},
				"snapshot_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by snapshot ID - used with 'list' operation",
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with 'list' operation",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with 'list' operation",
					"enum":        []string{"id", "start_time"},
				},
				// Parameters for get operation
				"backup_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the backup - required for 'get' operation",
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"backup_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "start"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"agent_id"},
					},
				},
			},
		},
	}
}
