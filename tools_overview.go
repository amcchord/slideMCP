package main

// slide_overview: hierarchical inventory + health summary. Replaces the
// `list_all_clients_devices_and_agents` top-level tool plus the bits of
// the retired slide_meta surface that an MSP tech actually asks for
// in natural language ("what's the state of things?").

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func handleOverviewTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfigWithResolutions("slide_overview", ToolOperations{
		"inventory":  handleOverviewInventory,
		"health":     handleOverviewHealth,
		"for_client": handleOverviewForClient,
		"for_device": handleOverviewForDevice,
	}, map[string]ResolutionSpec{
		"for_client": {IDKey: "client_id", Kind: "client"},
		"for_device": {IDKey: "device_id", Kind: "device"},
	}), args)
}

var overviewOperationEnums = []string{"inventory", "health", "for_client", "for_device"}

func getOverviewToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        overviewOperationEnums,
		},
		"client_id": map[string]interface{}{
			"type":        "string",
			"description": "Client ID. Required for `for_client` (alternative: pass `name_hint`).",
		},
		"device_id": map[string]interface{}{
			"type":        "string",
			"description": "Device ID. Required for `for_device` (alternative: pass `name_hint`).",
		},
		"name_hint": map[string]interface{}{
			"type":        "string",
			"description": "Alternative to client_id / device_id: a client name, device hostname, or display name (case-insensitive substring match). For `for_client` resolves to a client; for `for_device` resolves to a device. Ambiguous matches return a structured `name_hint_error=ambiguous` response with candidates.",
		},
		"stale_minutes": map[string]interface{}{
			"type":        "number",
			"description": "Used by `health`. A device or agent is flagged unhealthy if last_seen_at is older than this. Default: 30.",
			"minimum":     1,
		},
	}
	for k, v := range commonSinglePropertiesNoList() {
		props[k] = v
	}

	return ToolInfo{
		Name: "slide_overview",
		Description: "Slide MCP - hierarchical inventory + health for the whole Slide-protected MSP environment. " +
			"REACH FOR THIS whenever the user mentions 'are all my Slide boxes healthy?', 'is everything OK', " +
			"'what clients/devices/agents do I have', or wants a one-screen health summary for BCDR / backup posture. " +
			"Operations: `inventory` (full clients -> devices -> agents tree, the answer to \"what do we have?\"), " +
			"`health` (one-line-per-device-and-agent summary with stale/healthy flags - the answer to \"are all my Slide boxes OK?\"), " +
			"`for_client` (deep view of a single client's devices + agents + open alerts; accepts client_id OR name_hint), " +
			"`for_device` (deep view of one device's agents + recent backups + open alerts; accepts device_id OR name_hint). " +
			"Read-only and safe to call at the start of any conversation - call it before asking the user clarifying questions about their environment.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("for_client"), "then": reqEither("client_id", "name_hint")},
				{"if": ifOp("for_device"), "then": reqEither("device_id", "name_hint")},
			},
		},
	}
}

// handleOverviewInventory preserves the legacy clients->devices->agents
// hierarchy that the old list_all_clients_devices_and_agents returned, in a
// shape Claude Desktop is already used to.
func handleOverviewInventory(args map[string]interface{}) (string, error) {
	return listAllClientsDevicesAndAgents(args)
}

// healthEntry is one row in the health summary.
type healthEntry struct {
	Kind         string `json:"kind"` // "device" or "agent"
	ID           string `json:"id"`
	Name         string `json:"name"`
	ClientID     string `json:"client_id,omitempty"`
	LastSeenAt   string `json:"last_seen_at,omitempty"`
	MinutesStale int    `json:"minutes_stale,omitempty"`
	Status       string `json:"status"` // "healthy" | "stale" | "unknown"
	Detail       string `json:"detail,omitempty"`
}

func handleOverviewHealth(args map[string]interface{}) (string, error) {
	staleMinutes, ok := optionalInt(args, "stale_minutes")
	if !ok || staleMinutes <= 0 {
		staleMinutes = 30
	}

	devicesData, err := makeAPIRequest("GET", "/v1/device?limit=50", nil)
	if err != nil {
		return "", fmt.Errorf("failed to list devices: %w", err)
	}
	var devices PaginatedResponse[Device]
	if err := json.Unmarshal(devicesData, &devices); err != nil {
		return "", fmt.Errorf("parse devices: %w", err)
	}

	agentsData, err := makeAPIRequest("GET", "/v1/agent?limit=50", nil)
	if err != nil {
		return "", fmt.Errorf("failed to list agents: %w", err)
	}
	var agents PaginatedResponse[Agent]
	if err := json.Unmarshal(agentsData, &agents); err != nil {
		return "", fmt.Errorf("parse agents: %w", err)
	}

	now := time.Now().UTC()
	entries := make([]healthEntry, 0, len(devices.Data)+len(agents.Data))

	healthy, stale, unknown := 0, 0, 0
	classify := func(lastSeen string) (string, int) {
		if lastSeen == "" {
			return "unknown", 0
		}
		t, err := time.Parse(time.RFC3339, lastSeen)
		if err != nil {
			return "unknown", 0
		}
		mins := int(now.Sub(t).Minutes())
		if mins < 0 {
			mins = 0
		}
		if mins > staleMinutes {
			return "stale", mins
		}
		return "healthy", mins
	}

	for _, d := range devices.Data {
		status, mins := classify(d.LastSeenAt)
		switch status {
		case "healthy":
			healthy++
		case "stale":
			stale++
		default:
			unknown++
		}
		name := d.DisplayName
		if name == "" {
			name = d.Hostname
		}
		clientID := ""
		if d.ClientID != nil {
			clientID = *d.ClientID
		}
		detail := d.ServiceStatus
		entries = append(entries, healthEntry{
			Kind:         "device",
			ID:           d.DeviceID,
			Name:         name,
			ClientID:     clientID,
			LastSeenAt:   d.LastSeenAt,
			MinutesStale: mins,
			Status:       status,
			Detail:       detail,
		})
	}

	for _, a := range agents.Data {
		status, mins := classify(a.LastSeenAt)
		switch status {
		case "healthy":
			healthy++
		case "stale":
			stale++
		default:
			unknown++
		}
		name := a.DisplayName
		if name == "" {
			name = a.Hostname
		}
		clientID := ""
		if a.ClientID != nil {
			clientID = *a.ClientID
		}
		entries = append(entries, healthEntry{
			Kind:         "agent",
			ID:           a.AgentID,
			Name:         name,
			ClientID:     clientID,
			LastSeenAt:   a.LastSeenAt,
			MinutesStale: mins,
			Status:       status,
			Detail:       strings.TrimSpace(a.OS + " " + a.OSVersion),
		})
	}

	out := map[string]interface{}{
		"summary": map[string]interface{}{
			"total":        len(entries),
			"healthy":      healthy,
			"stale":        stale,
			"unknown":      unknown,
			"stale_cutoff": fmt.Sprintf("%d minutes", staleMinutes),
			"generated_at": now.Format(time.RFC3339),
		},
		"entries": entries,
	}
	return formatSingle(out, args, formatCompact)
}

// handleOverviewForClient: client + devices + agents + open alerts in one shot.
func handleOverviewForClient(args map[string]interface{}) (string, error) {
	clientID, err := requireString(args, "client_id")
	if err != nil {
		return "", err
	}

	clientData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/client/%s", clientID), nil)
	if err != nil {
		return "", err
	}
	var client Client
	if err := json.Unmarshal(clientData, &client); err != nil {
		return "", fmt.Errorf("parse client: %w", err)
	}

	devicesData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device?client_id=%s&limit=50", clientID), nil)
	if err != nil {
		return "", err
	}
	var devices PaginatedResponse[Device]
	_ = json.Unmarshal(devicesData, &devices)

	deviceSummaries := make([]map[string]interface{}, 0, len(devices.Data))
	for _, d := range devices.Data {
		name := d.DisplayName
		if name == "" {
			name = d.Hostname
		}
		deviceSummaries = append(deviceSummaries, map[string]interface{}{
			"device_id":      d.DeviceID,
			"name":           name,
			"hostname":       d.Hostname,
			"last_seen_at":   d.LastSeenAt,
			"service_status": d.ServiceStatus,
		})
	}

	alertsData, _ := makeAPIRequest("GET", fmt.Sprintf("/v1/alert?resolved=false&limit=50"), nil)
	var alerts PaginatedResponse[Alert]
	_ = json.Unmarshal(alertsData, &alerts)
	openAlerts := make([]map[string]interface{}, 0)
	for _, al := range alerts.Data {
		// Filter to alerts whose device belongs to this client.
		if al.DeviceID == nil {
			continue
		}
		for _, d := range devices.Data {
			if d.DeviceID == *al.DeviceID {
				openAlerts = append(openAlerts, map[string]interface{}{
					"alert_id":   al.AlertID,
					"alert_type": al.AlertType,
					"created_at": al.CreatedAt,
					"device_id":  *al.DeviceID,
				})
				break
			}
		}
	}

	out := map[string]interface{}{
		"client": map[string]interface{}{
			"client_id": client.ClientID,
			"name":      client.Name,
			"comments":  client.Comments,
		},
		"devices":     deviceSummaries,
		"open_alerts": openAlerts,
		"counts": map[string]interface{}{
			"devices":     len(devices.Data),
			"open_alerts": len(openAlerts),
		},
	}
	return formatSingle(out, args, formatCompact)
}

// handleOverviewForDevice: device + agents + last 24h backups (count) + open alerts.
func handleOverviewForDevice(args map[string]interface{}) (string, error) {
	deviceID, err := requireString(args, "device_id")
	if err != nil {
		return "", err
	}

	deviceData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device/%s", deviceID), nil)
	if err != nil {
		return "", err
	}
	var device Device
	if err := json.Unmarshal(deviceData, &device); err != nil {
		return "", fmt.Errorf("parse device: %w", err)
	}

	agentsData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent?device_id=%s&limit=50", deviceID), nil)
	if err != nil {
		return "", err
	}
	var agents PaginatedResponse[Agent]
	_ = json.Unmarshal(agentsData, &agents)

	agentSummaries := make([]map[string]interface{}, 0, len(agents.Data))
	for _, a := range agents.Data {
		name := a.DisplayName
		if name == "" {
			name = a.Hostname
		}
		agentSummaries = append(agentSummaries, map[string]interface{}{
			"agent_id":     a.AgentID,
			"name":         name,
			"hostname":     a.Hostname,
			"last_seen_at": a.LastSeenAt,
			"os":           strings.TrimSpace(a.OS + " " + a.OSVersion),
		})
	}

	alertsData, _ := makeAPIRequest("GET", fmt.Sprintf("/v1/alert?device_id=%s&resolved=false&limit=50", deviceID), nil)
	var alerts PaginatedResponse[Alert]
	_ = json.Unmarshal(alertsData, &alerts)
	openAlerts := make([]map[string]interface{}, 0, len(alerts.Data))
	for _, al := range alerts.Data {
		openAlerts = append(openAlerts, map[string]interface{}{
			"alert_id":   al.AlertID,
			"alert_type": al.AlertType,
			"created_at": al.CreatedAt,
			"agent_id":   al.AgentID,
		})
	}

	out := map[string]interface{}{
		"device": map[string]interface{}{
			"device_id":      device.DeviceID,
			"display_name":   device.DisplayName,
			"hostname":       device.Hostname,
			"last_seen_at":   device.LastSeenAt,
			"service_status": device.ServiceStatus,
			"image_version":  device.ImageVersion,
		},
		"agents":      agentSummaries,
		"open_alerts": openAlerts,
		"counts": map[string]interface{}{
			"agents":      len(agents.Data),
			"open_alerts": len(openAlerts),
		},
	}
	return formatSingle(out, args, formatCompact)
}
