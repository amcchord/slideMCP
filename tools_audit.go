package main

// slide_audit: query audit logs (Slide API v1.27.0). Critical for
// compliance + "what just changed in our account?" questions.

import (
	"fmt"
	"time"
)

func handleAuditTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_audit", ToolOperations{
		"list":      handleAuditList,
		"get":       handleAuditGet,
		"actions":   handleAuditActions,
		"resources": handleAuditResources,
		"recent":    handleAuditRecent,
	}), args)
}

var auditOperationEnums = []string{"list", "get", "actions", "resources", "recent"}

func getAuditToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        auditOperationEnums,
		},
		"audit_id": map[string]interface{}{
			"type":        "string",
			"description": "Audit log ID. Required for `get`.",
		},
		"audit_action_name": map[string]interface{}{
			"type":        "string",
			"description": "Filter to a specific action (e.g. `user_created`). Used with `list`. Use `actions` to discover the valid set.",
		},
		"audit_resource_type_name": map[string]interface{}{
			"type":        "string",
			"description": "Filter to a specific resource type (e.g. `user`, `agent`). Used with `list`. Use `resources` to discover the valid set.",
		},
		"audit_time_before": map[string]interface{}{
			"type":        "string",
			"description": "RFC3339 cutoff (return entries strictly before this time).",
			"format":      "date-time",
		},
		"audit_time_after": map[string]interface{}{
			"type":        "string",
			"description": "RFC3339 cutoff (return entries strictly after this time).",
			"format":      "date-time",
		},
		"hours": map[string]interface{}{
			"type":        "number",
			"description": "For `recent`: how many hours back to fetch (default 24).",
			"minimum":     1,
			"maximum":     720,
		},
		"sort_by": map[string]interface{}{
			"type":        "string",
			"description": "Sort field for `list` (currently only `audit_time` is supported).",
			"enum":        []string{"audit_time"},
		},
	}
	for k, v := range commonListProperties() {
		if _, exists := props[k]; !exists {
			props[k] = v
		}
	}

	return ToolInfo{
		Name: "slide_audit",
		Description: "Slide MCP - query the Slide account audit log. " +
			"REACH FOR THIS whenever the user mentions 'audit log', 'what changed', 'who did X', 'last night's changes', " +
			"'compliance report', 'who deleted/created Y', or any change-tracking / forensic question about Slide. " +
			"Operations: `list` (paginated query with optional action/resource/time filters), `get` (single audit entry by ID), " +
			"`actions` (list valid action names for the `audit_action_name` filter), `resources` (list valid resource type names " +
			"for the `audit_resource_type_name` filter), `recent` (convenience: last N hours, default 24). " +
			"Use this for compliance, change-tracking, and \"who did X to Y\" investigations.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": req("audit_id")},
			},
		},
	}
}

func handleAuditList(args map[string]interface{}) (string, error) {
	limit, _ := optionalInt(args, "limit")
	offset, _ := optionalInt(args, "offset")
	actionName, _ := optionalString(args, "audit_action_name")
	resourceType, _ := optionalString(args, "audit_resource_type_name")
	sortBy, _ := optionalString(args, "sort_by")
	timeBefore, _ := optionalString(args, "audit_time_before")
	timeAfter, _ := optionalString(args, "audit_time_after")
	var sortAscPtr *bool
	if v, ok := optionalBool(args, "sort_asc"); ok {
		sortAscPtr = &v
	}

	resp, err := listAudits(auditQueryOpts{
		Limit:           limit,
		Offset:          offset,
		ActionName:      actionName,
		ResourceType:    resourceType,
		SortBy:          sortBy,
		SortAsc:         sortAscPtr,
		AuditTimeBefore: timeBefore,
		AuditTimeAfter:  timeAfter,
	})
	if err != nil {
		return "", err
	}
	return formatList(resp.Data, resp.Pagination, args, formatSummary, summarizeAudit)
}

func handleAuditGet(args map[string]interface{}) (string, error) {
	auditID, err := requireString(args, "audit_id")
	if err != nil {
		return "", err
	}
	a, err := getAudit(auditID)
	if err != nil {
		return "", err
	}
	return formatSingle(a, args, formatCompact)
}

func handleAuditActions(args map[string]interface{}) (string, error) {
	limit, _ := optionalInt(args, "limit")
	offset, _ := optionalInt(args, "offset")
	resp, err := listAuditActions(limit, offset)
	if err != nil {
		return "", err
	}
	return formatList(resp.Data, resp.Pagination, args, formatSummary, func(a AuditAction) map[string]interface{} {
		return map[string]interface{}{"name": a.Name, "description": a.Description}
	})
}

func handleAuditResources(args map[string]interface{}) (string, error) {
	limit, _ := optionalInt(args, "limit")
	offset, _ := optionalInt(args, "offset")
	resp, err := listAuditResourceTypes(limit, offset)
	if err != nil {
		return "", err
	}
	return formatList(resp.Data, resp.Pagination, args, formatSummary, func(r AuditResourceType) map[string]interface{} {
		return map[string]interface{}{"name": r.Name}
	})
}

// handleAuditRecent is a convenience wrapper: "show me the last N hours".
// Defaults to 24h, capped at 30 days.
func handleAuditRecent(args map[string]interface{}) (string, error) {
	hours, ok := optionalInt(args, "hours")
	if !ok || hours <= 0 {
		hours = 24
	}
	if hours > 720 {
		hours = 720
	}
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour).Format(time.RFC3339)

	limit, _ := optionalInt(args, "limit")
	if limit == 0 {
		limit = 50
	}
	offset, _ := optionalInt(args, "offset")
	actionName, _ := optionalString(args, "audit_action_name")
	resourceType, _ := optionalString(args, "audit_resource_type_name")

	resp, err := listAudits(auditQueryOpts{
		Limit:          limit,
		Offset:         offset,
		ActionName:     actionName,
		ResourceType:   resourceType,
		SortBy:         "audit_time",
		AuditTimeAfter: cutoff,
	})
	if err != nil {
		return "", err
	}
	body, err := formatList(resp.Data, resp.Pagination, args, formatSummary, summarizeAudit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`{"window_hours":%d,"cutoff_after":%q,"result":%s}`, hours, cutoff, body), nil
}

func summarizeAudit(a Audit) map[string]interface{} {
	out := map[string]interface{}{
		"audit_id":      a.AuditID,
		"audit_time":    a.AuditTime,
		"action":        a.Action,
		"resource_type": a.ResourceType,
		"resource_id":   a.ResourceID,
	}
	if a.UserDisplayName != nil && *a.UserDisplayName != "" {
		out["user"] = *a.UserDisplayName
	} else if a.UserID != nil && *a.UserID != "" {
		out["user_id"] = *a.UserID
	}
	if a.Description != "" {
		out["description"] = a.Description
	}
	return out
}
