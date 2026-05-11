package main

// slide_backups: list/get/start backup runs PLUS the v4 status_for_client
// and status_for_device convenience ops that answer "did backups run last
// night for X?" in a single tool call.

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

func handleBackupsTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_backups", ToolOperations{
		"list":              handleBackupsList,
		"get":               handleBackupsGet,
		"start":             handleBackupsStart,
		"status_for_client": handleBackupsStatusForClient,
		"status_for_device": handleBackupsStatusForDevice,
		"recent_for_agent":  handleBackupsRecentForAgent,
	}), args)
}

var backupsOperationEnums = []string{"list", "get", "start", "status_for_client", "status_for_device", "recent_for_agent"}

func getBackupsToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        backupsOperationEnums,
		},
		"agent_id": map[string]interface{}{
			"type":        "string",
			"description": "Agent ID. Required for `start` and `recent_for_agent`. Optional filter for `list`.",
		},
		"device_id": map[string]interface{}{
			"type":        "string",
			"description": "Device ID. Required for `status_for_device`. Optional filter for `list`.",
		},
		"client_id": map[string]interface{}{
			"type":        "string",
			"description": "Client ID. Required for `status_for_client`.",
		},
		"snapshot_id": map[string]interface{}{
			"type":        "string",
			"description": "Snapshot ID filter (used with `list`).",
		},
		"backup_id": map[string]interface{}{
			"type":        "string",
			"description": "Backup ID. Required for `get`.",
		},
		"hours": map[string]interface{}{
			"type":        "number",
			"description": "Time window in hours for `status_for_*` and `recent_for_agent`. Default 24.",
			"minimum":     1,
			"maximum":     720,
		},
		"sort_by": map[string]interface{}{
			"type":        "string",
			"description": "Sort field for `list`.",
			"enum":        []string{"id", "start_time"},
		},
	}
	for k, v := range commonListProperties() {
		if _, exists := props[k]; !exists {
			props[k] = v
		}
	}

	return ToolInfo{
		Name: "slide_backups",
		Description: "Backup-run inspection and on-demand backup launch. " +
			"Operations: `list` (paginated history with filters), `get` (single backup detail), `start` (kick off a new backup), " +
			"`status_for_client` (last-N-hours summary for every agent under a client), " +
			"`status_for_device` (last-N-hours summary for every agent on a device), " +
			"`recent_for_agent` (last-N-hours runs for one agent). " +
			"The `status_for_*` ops answer \"did backups run last night for X?\" in one call.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("get"), "then": req("backup_id")},
				{"if": ifOp("start"), "then": req("agent_id")},
				{"if": ifOp("status_for_client"), "then": req("client_id")},
				{"if": ifOp("status_for_device"), "then": req("device_id")},
				{"if": ifOp("recent_for_agent"), "then": req("agent_id")},
			},
		},
	}
}

func handleBackupsList(args map[string]interface{}) (string, error) {
	return listBackups(args)
}
func handleBackupsGet(args map[string]interface{}) (string, error) {
	return getBackup(args)
}
func handleBackupsStart(args map[string]interface{}) (string, error) {
	return startBackup(args)
}

// agentBackupStatus is the per-agent rollup returned by status_for_*.
type agentBackupStatus struct {
	AgentID          string `json:"agent_id"`
	AgentName        string `json:"agent_name,omitempty"`
	Total            int    `json:"total"`
	Successful       int    `json:"successful"`
	Failed           int    `json:"failed"`
	InProgress       int    `json:"in_progress"`
	LastEndedAt      string `json:"last_ended_at,omitempty"`
	LastStatus       string `json:"last_status,omitempty"`
	LastErrorCode    *int   `json:"last_error_code,omitempty"`
	LastErrorMessage string `json:"last_error_message,omitempty"`
}

func handleBackupsStatusForClient(args map[string]interface{}) (string, error) {
	clientID, err := requireString(args, "client_id")
	if err != nil {
		return "", err
	}
	hours, ok := optionalInt(args, "hours")
	if !ok || hours <= 0 {
		hours = 24
	}

	devicesData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device?client_id=%s&limit=50", clientID), nil)
	if err != nil {
		return "", err
	}
	var devices PaginatedResponse[Device]
	if err := json.Unmarshal(devicesData, &devices); err != nil {
		return "", fmt.Errorf("parse devices: %w", err)
	}

	agents := []Agent{}
	for _, d := range devices.Data {
		ad, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent?device_id=%s&limit=50", d.DeviceID), nil)
		if err != nil {
			continue
		}
		var p PaginatedResponse[Agent]
		if json.Unmarshal(ad, &p) == nil {
			agents = append(agents, p.Data...)
		}
	}
	return runBackupsStatus(args, agents, hours, fmt.Sprintf("client %s", clientID))
}

func handleBackupsStatusForDevice(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}
	hours, ok := optionalInt(args, "hours")
	if !ok || hours <= 0 {
		hours = 24
	}
	agentsData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent?device_id=%s&limit=50", deviceID), nil)
	if err != nil {
		return "", err
	}
	var p PaginatedResponse[Agent]
	if err := json.Unmarshal(agentsData, &p); err != nil {
		return "", fmt.Errorf("parse agents: %w", err)
	}
	return runBackupsStatus(args, p.Data, hours, fmt.Sprintf("device %s", deviceID))
}

func handleBackupsRecentForAgent(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	hours, ok := optionalInt(args, "hours")
	if !ok || hours <= 0 {
		hours = 24
	}
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	data, err := makeAPIRequest("GET", fmt.Sprintf("/v1/backup?agent_id=%s&limit=50&sort_by=start_time", agentID), nil)
	if err != nil {
		return "", err
	}
	var p PaginatedResponse[Backup]
	if err := json.Unmarshal(data, &p); err != nil {
		return "", fmt.Errorf("parse backups: %w", err)
	}
	filtered := filterBackupsAfter(p.Data, cutoff)
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].StartedAt > filtered[j].StartedAt })
	return formatList(filtered, p.Pagination, args, formatSummary, summarizeBackup)
}

// runBackupsStatus is the shared implementation: fetch each agent's recent
// backups, summarise success/failure counts, and emit a per-client/device rollup.
func runBackupsStatus(args map[string]interface{}, agents []Agent, hours int, scope string) (string, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	statuses := make([]agentBackupStatus, 0, len(agents))
	totalSuccess, totalFail, totalProg := 0, 0, 0

	for _, a := range agents {
		data, err := makeAPIRequest("GET", fmt.Sprintf("/v1/backup?agent_id=%s&limit=50&sort_by=start_time", a.AgentID), nil)
		if err != nil {
			statuses = append(statuses, agentBackupStatus{
				AgentID: a.AgentID, AgentName: bestAgentName(a),
				LastStatus: "error fetching backups",
			})
			continue
		}
		var p PaginatedResponse[Backup]
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}
		recent := filterBackupsAfter(p.Data, cutoff)

		s := agentBackupStatus{AgentID: a.AgentID, AgentName: bestAgentName(a)}
		for _, b := range recent {
			s.Total++
			switch b.Status {
			case "succeeded":
				s.Successful++
				totalSuccess++
			case "failed":
				s.Failed++
				totalFail++
			default:
				s.InProgress++
				totalProg++
			}
		}
		// Latest backup overall for "last status"
		sort.Slice(recent, func(i, j int) bool { return recent[i].StartedAt > recent[j].StartedAt })
		if len(recent) > 0 {
			latest := recent[0]
			s.LastStatus = latest.Status
			if latest.EndedAt != nil {
				s.LastEndedAt = *latest.EndedAt
			}
			s.LastErrorCode = latest.ErrorCode
			if latest.ErrorMessage != nil {
				s.LastErrorMessage = *latest.ErrorMessage
			}
		}
		statuses = append(statuses, s)
	}

	out := map[string]interface{}{
		"scope":        scope,
		"window_hours": hours,
		"summary": map[string]interface{}{
			"agents":       len(agents),
			"successful":   totalSuccess,
			"failed":       totalFail,
			"in_progress":  totalProg,
			"window_start": cutoff.Format(time.RFC3339),
		},
		"agents_status": statuses,
	}
	return formatSingle(out, args, formatCompact)
}

func filterBackupsAfter(backups []Backup, cutoff time.Time) []Backup {
	out := backups[:0:0]
	for _, b := range backups {
		t, err := time.Parse(time.RFC3339, b.StartedAt)
		if err != nil {
			continue
		}
		if t.After(cutoff) {
			out = append(out, b)
		}
	}
	return out
}

func bestAgentName(a Agent) string {
	if a.DisplayName != "" {
		return a.DisplayName
	}
	return a.Hostname
}

func summarizeBackup(b Backup) map[string]interface{} {
	out := map[string]interface{}{
		"backup_id":  b.BackupID,
		"started_at": b.StartedAt,
		"status":     b.Status,
	}
	if b.EndedAt != nil {
		out["ended_at"] = *b.EndedAt
	}
	if b.SnapshotID != nil {
		out["snapshot_id"] = *b.SnapshotID
	}
	if b.ErrorCode != nil {
		out["error_code"] = *b.ErrorCode
	}
	if b.ErrorMessage != nil {
		out["error_message"] = *b.ErrorMessage
	}
	return out
}
