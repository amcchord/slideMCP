package main

// handleSnapshotsTool handles all snapshot-related operations through a single meta-tool
func handleSnapshotsTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_snapshots", ToolOperations{
		"list":                     listSnapshots,
		"list_deleted":             listDeletedSnapshots,
		"get":                      getSnapshot,
		"get_service_verification": handleSnapshotGetServiceVerification,
	}), args)
}

// listDeletedSnapshots lists deleted snapshots by calling listSnapshots with deleted location filters
func listDeletedSnapshots(args map[string]interface{}) (string, error) {
	if _, exists := args["snapshot_location"]; !exists {
		args["snapshot_location"] = "exists_deleted"
	}
	return listSnapshots(args)
}

// getSnapshotsToolInfo returns the tool definition for the snapshots meta-tool
func getSnapshotsToolInfo() ToolInfo {
	return ToolInfo{
		Name: "slide_snapshots",
		Description: "Manage backup snapshots (point-in-time recovery points). Operations: list, list_deleted, get, " +
			"get_service_verification (Slide API v1.27.0 - per-service verification results for the snapshot). " +
			"Get/list responses now also include verify_service_status. Filter by location, date ranges, deletion status.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "list_deleted", "get", "get_service_verification"},
				},
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
					"description": "Filter by snapshot location - used with 'list' operation (automatically set for 'list_deleted'). Defaults to 'location_any'.",
					"enum":        []string{"exists_local", "exists_cloud", "exists_deleted", "exists_deleted_retention", "exists_deleted_manual", "exists_deleted_other", "location_any"},
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
				"snapshot_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the snapshot - required for 'get' and 'get_service_verification' operations",
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": req("snapshot_id")},
				{"if": ifOp("get_service_verification"), "then": req("snapshot_id")},
			},
		},
	}
}
