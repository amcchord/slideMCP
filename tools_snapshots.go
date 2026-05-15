package main

// slide_snapshots: list / list_deleted / get / get_service_verification +
// v4 `recent_for_agent` convenience.

import (
	"encoding/json"
	"fmt"
)

func handleSnapshotsTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfigWithResolutions("slide_snapshots", ToolOperations{
		"list":                     listSnapshots,
		"list_deleted":             listDeletedSnapshots,
		"get":                      getSnapshot,
		"get_service_verification": handleSnapshotGetServiceVerification,
		"recent_for_agent":         handleSnapshotsRecentForAgent,
	}, map[string]ResolutionSpec{
		"recent_for_agent": {IDKey: "agent_id", Kind: "agent"},
	}), args)
}

func listDeletedSnapshots(args map[string]interface{}) (string, error) {
	if _, exists := args["snapshot_location"]; !exists {
		args["snapshot_location"] = "exists_deleted"
	}
	return listSnapshots(args)
}

var snapshotsOperationEnums = []string{"list", "list_deleted", "get", "get_service_verification", "recent_for_agent"}

func getSnapshotsToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        snapshotsOperationEnums,
		},
		"agent_id": map[string]interface{}{
			"type":        "string",
			"description": "Filter by agent. Required for `recent_for_agent` (alternative: pass `name_hint`).",
		},
		"name_hint": map[string]interface{}{
			"type":        "string",
			"description": "Alternative to agent_id for `recent_for_agent`: an agent hostname or display name.",
		},
		"snapshot_id": map[string]interface{}{
			"type":        "string",
			"description": "Required for `get` and `get_service_verification`.",
		},
		"snapshot_location": map[string]interface{}{
			"type":        "string",
			"description": "Filter by location. Set automatically for `list_deleted`.",
			"enum":        []string{"exists_local", "exists_cloud", "exists_deleted", "exists_deleted_retention", "exists_deleted_manual", "exists_deleted_other", "location_any"},
		},
		"days": map[string]interface{}{
			"type":        "number",
			"description": "Window in days for `recent_for_agent`. Default 14.",
			"minimum":     1,
			"maximum":     90,
		},
		"sort_by": map[string]interface{}{
			"type":        "string",
			"description": "Sort field for `list`/`list_deleted`/`recent_for_agent`.",
			"enum":        []string{"backup_start_time", "backup_end_time", "created"},
		},
	}
	for k, v := range commonListProperties() {
		if _, exists := props[k]; !exists {
			props[k] = v
		}
	}

	return ToolInfo{
		Name: "slide_snapshots",
		Description: "Slide MCP - inspect backup snapshots (point-in-time recovery points). " +
			"REACH FOR THIS whenever the user mentions a snapshot, restore point, recovery point, RPO, " +
			"'what backups do I have for X', 'last successful backup', 'show me snapshots from yesterday', " +
			"'when was the last verified boot', or wants to inspect (not restore) historical recovery points. " +
			"Operations: `list`, `list_deleted`, `get`, `get_service_verification` (Slide API v1.27.0 per-service results), " +
			"`recent_for_agent` (last N days for a single agent, default 14 - the answer to \"what restore points do I have for X?\"; accepts agent_id OR name_hint). " +
			"Get/list responses include verify_service_status.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": req("snapshot_id")},
				{"if": ifOp("get_service_verification"), "then": req("snapshot_id")},
				{"if": ifOp("recent_for_agent"), "then": reqEither("agent_id", "name_hint")},
			},
		},
	}
}

func handleSnapshotsRecentForAgent(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	days, ok := optionalInt(args, "days")
	if !ok || days <= 0 {
		days = 14
	}
	limit, _ := optionalInt(args, "limit")
	if limit == 0 {
		limit = 50
	}
	endpoint := fmt.Sprintf("/v1/snapshot?agent_id=%s&limit=%d&sort_by=backup_start_time&sort_asc=false", agentID, limit)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	var p PaginatedResponse[Snapshot]
	if err := json.Unmarshal(data, &p); err != nil {
		return "", fmt.Errorf("parse snapshots: %w", err)
	}
	out := map[string]interface{}{
		"agent_id":    agentID,
		"window_days": days,
		"snapshots":   p.Data,
		"pagination":  p.Pagination,
		"count":       len(p.Data),
	}
	return formatSingle(out, args, formatCompact)
}
