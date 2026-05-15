package main

// slide_alerts: list/get/update + the v4 `triage` convenience op that
// groups unresolved alerts by severity hint and returns the worst-first.

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

func handleAlertsTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_alerts", ToolOperations{
		"list":   handleAlertsList,
		"get":    getAlert,
		"update": updateAlert,
		"triage": handleAlertsTriage,
	}), args)
}

var alertsOperationEnums = []string{"list", "get", "update", "triage"}

func getAlertsToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        alertsOperationEnums,
		},
		"alert_id": map[string]interface{}{
			"type":        "string",
			"description": "Alert ID. Required for `get`, `update`.",
		},
		"resolved": map[string]interface{}{
			"type":        "boolean",
			"description": "For `list`: filter by resolution status. For `update`: required, set true to resolve.",
		},
		"device_id": map[string]interface{}{
			"type":        "string",
			"description": "Filter by device.",
		},
		"agent_id": map[string]interface{}{
			"type":        "string",
			"description": "Filter by agent.",
		},
	}
	for k, v := range commonListProperties() {
		if _, exists := props[k]; !exists {
			props[k] = v
		}
	}

	return ToolInfo{
		Name: "slide_alerts",
		Description: "Slide MCP - Slide alert triage. " +
			"REACH FOR THIS whenever the user mentions a Slide alert, unresolved alert, 'triage alerts', " +
			"'what should I look at first', 'critical alerts', storage-low / backup-failed / not-checking-in alerts, " +
			"or 'is anything broken on the Slide side'. " +
			"Operations: `list`, `get`, `update` (resolve/unresolve), " +
			"`triage` (rolls up unresolved alerts by severity hint and returns the worst-first list - the answer to \"what should I look at first?\").",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": req("alert_id")},
				{"if": ifOp("update"), "then": req("alert_id", "resolved")},
			},
		},
	}
}

func handleAlertsList(args map[string]interface{}) (string, error) {
	return listAlerts(args)
}

// alertSeverity returns a rough priority hint based on the alert_type
// taxonomy v1.27.0 introduced. Higher = more important.
func alertSeverity(alertType string) int {
	switch alertType {
	case "device_storage_space_critical":
		return 100
	case "agent_backup_failed", "device_storage_not_healthy":
		return 90
	case "agent_not_backing_up":
		return 80
	case "device_storage_space_low":
		return 70
	case "agent_not_checking_in", "device_not_checking_in":
		return 60
	case "device_out_of_date":
		return 30
	}
	return 50
}

func alertSeverityLabel(score int) string {
	switch {
	case score >= 90:
		return "critical"
	case score >= 70:
		return "high"
	case score >= 50:
		return "medium"
	default:
		return "low"
	}
}

func handleAlertsTriage(args map[string]interface{}) (string, error) {
	limit, _ := optionalInt(args, "limit")
	if limit == 0 {
		limit = 50
	}

	// Build base list query.
	q := fmt.Sprintf("/v1/alert?resolved=false&limit=%d", limit)
	if did, ok := args["device_id"].(string); ok && did != "" {
		q += "&device_id=" + did
	}
	if aid, ok := args["agent_id"].(string); ok && aid != "" {
		q += "&agent_id=" + aid
	}

	data, err := makeAPIRequest("GET", q, nil)
	if err != nil {
		return "", err
	}
	var p PaginatedResponse[Alert]
	if err := json.Unmarshal(data, &p); err != nil {
		return "", fmt.Errorf("parse alerts: %w", err)
	}

	type ranked struct {
		Alert    Alert  `json:"alert"`
		Severity string `json:"severity"`
		Score    int    `json:"score"`
		AgeHours int    `json:"age_hours"`
	}
	now := time.Now().UTC()
	out := make([]ranked, 0, len(p.Data))
	bySeverity := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}
	for _, a := range p.Data {
		score := alertSeverity(a.AlertType)
		label := alertSeverityLabel(score)
		bySeverity[label]++
		var ageHours int
		if t, err := time.Parse(time.RFC3339, a.CreatedAt); err == nil {
			ageHours = int(now.Sub(t).Hours())
		}
		out = append(out, ranked{Alert: a, Severity: label, Score: score, AgeHours: ageHours})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].AgeHours > out[j].AgeHours
	})

	resp := map[string]interface{}{
		"summary": map[string]interface{}{
			"unresolved":   len(out),
			"by_severity":  bySeverity,
			"generated_at": now.Format(time.RFC3339),
		},
		"alerts": out,
	}
	return formatSingle(resp, args, formatCompact)
}
