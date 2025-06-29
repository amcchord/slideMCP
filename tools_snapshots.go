package main

import (
	"fmt"
)

// handleSnapshotsTool handles all snapshot-related operations through a single meta-tool
func handleSnapshotsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_snapshots", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_snapshots in '%s' mode", operation, toolsMode)
	}

	switch operation {
	case "list":
		return listSnapshots(args)
	case "list_deleted":
		return listDeletedSnapshots(args)
	case "get":
		return getSnapshot(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// listDeletedSnapshots lists deleted snapshots by calling listSnapshots with deleted location filters
func listDeletedSnapshots(args map[string]interface{}) (string, error) {
	// Set snapshot_location to show deleted snapshots if not already specified
	if _, exists := args["snapshot_location"]; !exists {
		args["snapshot_location"] = "exists_deleted"
	}
	return listSnapshots(args)
}

// getSnapshotsToolInfo returns the tool definition for the snapshots meta-tool
func getSnapshotsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_snapshots",
		Description: "Manage snapshots - completed backup data that can be used for restores and virtual machines. Supports list, list_deleted, and get operations.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "list_deleted", "get"},
				},
				// Parameters for list and list_deleted operations
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with 'list' and 'list_deleted' operations",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with 'list' and 'list_deleted' operations",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by agent ID - used with 'list' and 'list_deleted' operations",
				},
				"snapshot_location": map[string]interface{}{
					"type":        "string",
					"description": "Filter by snapshot location - used with 'list' operation (automatically set for 'list_deleted')",
					"enum":        []string{"exists_local", "exists_cloud", "exists_deleted", "exists_deleted_retention", "exists_deleted_manual", "exists_deleted_other"},
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with 'list' and 'list_deleted' operations",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with 'list' and 'list_deleted' operations",
					"enum":        []string{"backup_start_time", "backup_end_time", "created"},
				},
				// Parameters for get operation
				"snapshot_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the snapshot - required for 'get' operation",
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
						"required": []string{"snapshot_id"},
					},
				},
			},
		},
	}
}
