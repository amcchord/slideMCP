package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// handleMetaTool handles all meta-tool operations
func handleMetaTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	switch operation {
	case "list_all_clients_devices_and_agents":
		return listAllClientsDevicesAndAgents(args)
	case "get_snapshot_changes":
		return getSnapshotChanges(args)
	case "get_reporting_data":
		return getReportingData(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getMetaToolInfo returns the tool definition for the meta tools
func getMetaToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_meta",
		Description: "Meta tools for reporting and aggregated data views. Provides hierarchical views and time-based snapshot analysis.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list_all_clients_devices_and_agents", "get_snapshot_changes", "get_reporting_data"},
				},
				// Parameters for get_snapshot_changes
				"period": map[string]interface{}{
					"type":        "string",
					"description": "Time period for snapshot changes - used with 'get_snapshot_changes' operation",
					"enum":        []string{"day", "week", "month"},
				},
				"client_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by client ID - optional for time-based operations",
				},
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by device ID - optional for time-based operations",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by agent ID - optional for time-based operations",
				},
				// Parameters for get_reporting_data
				"report_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of report data to generate - used with 'get_reporting_data' operation",
					"enum":        []string{"daily", "weekly", "monthly"},
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get_snapshot_changes"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"period"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get_reporting_data"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"report_type"},
					},
				},
			},
		},
	}
}

// getSnapshotChanges returns snapshot changes (new and deleted) over a time period
func getSnapshotChanges(args map[string]interface{}) (string, error) {
	period, ok := args["period"].(string)
	if !ok {
		return "", fmt.Errorf("period is required")
	}

	// Calculate the time range based on period
	now := time.Now()
	var startTime time.Time
	switch period {
	case "day":
		startTime = now.AddDate(0, 0, -1)
	case "week":
		startTime = now.AddDate(0, 0, -7)
	case "month":
		startTime = now.AddDate(0, -1, 0)
	default:
		return "", fmt.Errorf("invalid period: %s", period)
	}

	// Build filters for snapshot queries
	filters := make(map[string]interface{})
	filters["limit"] = 50 // We'll need to handle pagination for large datasets

	// Add optional filters
	if clientID, ok := args["client_id"].(string); ok && clientID != "" {
		filters["client_id"] = clientID
	}
	if deviceID, ok := args["device_id"].(string); ok && deviceID != "" {
		filters["device_id"] = deviceID
	}
	if agentID, ok := args["agent_id"].(string); ok && agentID != "" {
		filters["agent_id"] = agentID
	}

	// Get all snapshots (active and deleted)
	allSnapshots := make([]Snapshot, 0)

	// Get active snapshots
	activeFilters := make(map[string]interface{})
	for k, v := range filters {
		activeFilters[k] = v
	}
	activeFilters["snapshot_location"] = "exists_cloud"

	activeData, err := makeAPIRequest("GET", buildQueryString("/v1/snapshot", activeFilters), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get active snapshots: %w", err)
	}

	var activeResult PaginatedResponse[Snapshot]
	if err := json.Unmarshal(activeData, &activeResult); err != nil {
		return "", fmt.Errorf("failed to parse active snapshots: %w", err)
	}
	allSnapshots = append(allSnapshots, activeResult.Data...)

	// Get deleted snapshots
	deletedFilters := make(map[string]interface{})
	for k, v := range filters {
		deletedFilters[k] = v
	}
	deletedFilters["snapshot_location"] = "exists_deleted"

	deletedData, err := makeAPIRequest("GET", buildQueryString("/v1/snapshot", deletedFilters), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get deleted snapshots: %w", err)
	}

	var deletedResult PaginatedResponse[Snapshot]
	if err := json.Unmarshal(deletedData, &deletedResult); err != nil {
		return "", fmt.Errorf("failed to parse deleted snapshots: %w", err)
	}
	allSnapshots = append(allSnapshots, deletedResult.Data...)

	// Process snapshots to find changes within the time period
	newSnapshots := make([]map[string]interface{}, 0)
	deletedSnapshots := make([]map[string]interface{}, 0)

	for _, snapshot := range allSnapshots {
		// Check if snapshot was created in the time period
		backupTime, err := time.Parse(time.RFC3339, snapshot.BackupEndedAt)
		if err == nil && backupTime.After(startTime) {
			snapshotInfo := map[string]interface{}{
				"snapshot_id":       snapshot.SnapshotID,
				"agent_id":          snapshot.AgentID,
				"backup_started_at": snapshot.BackupStartedAt,
				"backup_ended_at":   snapshot.BackupEndedAt,
				"locations":         snapshot.Locations,
			}

			// Get agent info for context
			agentData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent/%s", snapshot.AgentID), nil)
			if err == nil {
				var agent Agent
				if json.Unmarshal(agentData, &agent) == nil {
					snapshotInfo["agent_name"] = agent.DisplayName
					snapshotInfo["agent_hostname"] = agent.Hostname
					snapshotInfo["device_id"] = agent.DeviceID

					// Get device info
					deviceData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device/%s", agent.DeviceID), nil)
					if err == nil {
						var device Device
						if json.Unmarshal(deviceData, &device) == nil {
							snapshotInfo["device_name"] = device.DisplayName
							snapshotInfo["device_hostname"] = device.Hostname
							if device.ClientID != nil {
								snapshotInfo["client_id"] = *device.ClientID
								snapshotInfo["client_name"] = getClientName(*device.ClientID)
							}
						}
					}
				}
			}

			newSnapshots = append(newSnapshots, snapshotInfo)
		}

		// Check if snapshot was deleted in the time period
		if snapshot.Deleted != nil && len(snapshot.Deletions) > 0 {
			for _, deletion := range snapshot.Deletions {
				deletedTime, err := time.Parse(time.RFC3339, deletion.Deleted)
				if err == nil && deletedTime.After(startTime) {
					snapshotInfo := map[string]interface{}{
						"snapshot_id":       snapshot.SnapshotID,
						"agent_id":          snapshot.AgentID,
						"backup_started_at": snapshot.BackupStartedAt,
						"backup_ended_at":   snapshot.BackupEndedAt,
						"deleted_at":        deletion.Deleted,
						"deleted_by":        deletion.DeletedBy,
						"deletion_type":     deletion.Type,
					}

					// Get agent info for context
					agentData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent/%s", snapshot.AgentID), nil)
					if err == nil {
						var agent Agent
						if json.Unmarshal(agentData, &agent) == nil {
							snapshotInfo["agent_name"] = agent.DisplayName
							snapshotInfo["agent_hostname"] = agent.Hostname
							snapshotInfo["device_id"] = agent.DeviceID

							// Get device info
							deviceData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device/%s", agent.DeviceID), nil)
							if err == nil {
								var device Device
								if json.Unmarshal(deviceData, &device) == nil {
									snapshotInfo["device_name"] = device.DisplayName
									snapshotInfo["device_hostname"] = device.Hostname
									if device.ClientID != nil {
										snapshotInfo["client_id"] = *device.ClientID
										snapshotInfo["client_name"] = getClientName(*device.ClientID)
									}
								}
							}
						}
					}

					deletedSnapshots = append(deletedSnapshots, snapshotInfo)
				}
			}
		}
	}

	// Build the result
	result := map[string]interface{}{
		"period":            period,
		"start_time":        startTime.Format(time.RFC3339),
		"end_time":          now.Format(time.RFC3339),
		"new_snapshots":     newSnapshots,
		"deleted_snapshots": deletedSnapshots,
		"summary": map[string]interface{}{
			"total_new":     len(newSnapshots),
			"total_deleted": len(deletedSnapshots),
		},
		"_metadata": map[string]interface{}{
			"description": fmt.Sprintf("Snapshot changes over the last %s", period),
			"usage":       "Use this data to populate reporting templates with accurate snapshot creation and deletion information",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// getReportingData returns pre-formatted data for report templates
func getReportingData(args map[string]interface{}) (string, error) {
	reportType, ok := args["report_type"].(string)
	if !ok {
		return "", fmt.Errorf("report_type is required")
	}

	// Get the full hierarchy first
	hierarchyData, err := listAllClientsDevicesAndAgents(map[string]interface{}{})
	if err != nil {
		return "", fmt.Errorf("failed to get hierarchy: %w", err)
	}

	var hierarchy map[string]interface{}
	if err := json.Unmarshal([]byte(hierarchyData), &hierarchy); err != nil {
		return "", fmt.Errorf("failed to parse hierarchy: %w", err)
	}

	// Calculate time periods based on report type
	now := time.Now()
	var period string
	var daysBack int
	switch reportType {
	case "daily":
		period = "day"
		daysBack = 1
	case "weekly":
		period = "week"
		daysBack = 7
	case "monthly":
		period = "month"
		daysBack = 30
	default:
		return "", fmt.Errorf("invalid report_type: %s", reportType)
	}

	// Get snapshot changes for the period
	changesData, err := getSnapshotChanges(map[string]interface{}{
		"period": period,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get snapshot changes: %w", err)
	}

	var changes map[string]interface{}
	if err := json.Unmarshal([]byte(changesData), &changes); err != nil {
		return "", fmt.Errorf("failed to parse changes: %w", err)
	}

	// Process hierarchy to calculate metrics
	clients := hierarchy["clients"].([]interface{})
	totalClients := len(clients)
	totalDevices := 0
	totalAgents := 0
	devicesActive := 0
	agentsOnline := 0

	clientMetrics := make([]map[string]interface{}, 0)

	for _, clientInterface := range clients {
		client := clientInterface.(map[string]interface{})
		devices := client["devices"].([]interface{})
		clientDeviceCount := len(devices)
		clientAgentCount := 0
		clientActiveDevices := 0
		clientOnlineAgents := 0

		for _, deviceInterface := range devices {
			device := deviceInterface.(map[string]interface{})
			agents := device["agents"].([]interface{})
			clientAgentCount += len(agents)

			// Check if device is active (seen in last 24 hours)
			if lastSeen, ok := device["last_seen_at"].(string); ok {
				lastSeenTime, err := time.Parse(time.RFC3339, lastSeen)
				if err == nil && time.Since(lastSeenTime) < 24*time.Hour {
					clientActiveDevices++
					devicesActive++
				}
			}

			// Check agent status
			for _, agentInterface := range agents {
				agent := agentInterface.(map[string]interface{})
				if lastSeen, ok := agent["last_seen_at"].(string); ok {
					lastSeenTime, err := time.Parse(time.RFC3339, lastSeen)
					if err == nil && time.Since(lastSeenTime) < 24*time.Hour {
						clientOnlineAgents++
						agentsOnline++
					}
				}
			}
		}

		totalDevices += clientDeviceCount
		totalAgents += clientAgentCount

		clientMetrics = append(clientMetrics, map[string]interface{}{
			"client_id":      client["client_id"],
			"client_name":    client["name"],
			"device_count":   clientDeviceCount,
			"agent_count":    clientAgentCount,
			"active_devices": clientActiveDevices,
			"online_agents":  clientOnlineAgents,
		})
	}

	// Build formatted data for templates
	result := map[string]interface{}{
		"report_type":   reportType,
		"report_date":   now.Format("January 2, 2006"),
		"report_period": fmt.Sprintf("Last %d days", daysBack),
		"metrics": map[string]interface{}{
			"total_clients":     totalClients,
			"total_devices":     totalDevices,
			"total_agents":      totalAgents,
			"devices_active":    devicesActive,
			"agents_online":     agentsOnline,
			"new_snapshots":     changes["summary"].(map[string]interface{})["total_new"],
			"deleted_snapshots": changes["summary"].(map[string]interface{})["total_deleted"],
		},
		"client_metrics":   clientMetrics,
		"snapshot_changes": changes,
		"hierarchy":        hierarchy,
		"_metadata": map[string]interface{}{
			"description": fmt.Sprintf("Pre-formatted data for %s report template", reportType),
			"usage":       "Use this data to populate report templates. All metrics are pre-calculated and formatted for easy insertion.",
			"template_placeholders": map[string]interface{}{
				"daily": []string{
					"REPORT_DATE", "TOTAL_SUCCESSFUL_SNAPSHOTS", "TOTAL_FAILED_SNAPSHOTS",
					"TOTAL_CLIENTS", "TOTAL_DEVICES", "AGENTS_ONLINE", "TOTAL_AGENTS",
					"DEVICES_ACTIVE", "SNAPSHOTS_TODAY", "CLIENT_NAME", "DEVICE_NAME",
				},
				"monthly": []string{
					"REPORT_MONTH_YEAR", "TOTAL_MONTHLY_SNAPSHOTS", "TOTAL_FAILED_SNAPSHOTS",
					"TOTAL_DELETED_SNAPSHOTS", "ACTIVE_DAYS", "TOTAL_DATA_SIZE",
					"RETENTION_RULES", "SNAPSHOTS_DELETED_COUNT", "SPACE_FREED",
				},
			},
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// buildQueryString builds a URL with query parameters
func buildQueryString(base string, params map[string]interface{}) string {
	if len(params) == 0 {
		return base
	}

	queryParts := make([]string, 0)
	for key, value := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%v", key, value))
	}

	return fmt.Sprintf("%s?%s", base, strings.Join(queryParts, "&"))
}

// Combined hierarchy function for easier LLM consumption (moved from api.go)
func listAllClientsDevicesAndAgents(args map[string]interface{}) (string, error) {
	// Get all clients first
	clientsData, err := makeAPIRequest("GET", "/v1/client?limit=50", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get clients: %w", err)
	}

	var clientsResult PaginatedResponse[Client]
	if err := json.Unmarshal(clientsData, &clientsResult); err != nil {
		return "", fmt.Errorf("failed to parse clients response: %w", err)
	}

	// Build the hierarchical structure
	clientsHierarchy := make([]map[string]interface{}, 0)

	// Handle clients with no ClientID (empty string clients)
	unassignedDevicesData, err := makeAPIRequest("GET", "/v1/device?limit=50", nil)
	if err != nil {
		return "", fmt.Errorf("failed to get unassigned devices: %w", err)
	}

	var unassignedDevicesResult PaginatedResponse[Device]
	if err := json.Unmarshal(unassignedDevicesData, &unassignedDevicesResult); err != nil {
		return "", fmt.Errorf("failed to parse unassigned devices response: %w", err)
	}

	// Filter devices that have empty/nil client_id
	unassignedDevices := make([]Device, 0)
	for _, device := range unassignedDevicesResult.Data {
		if device.ClientID == nil || *device.ClientID == "" {
			unassignedDevices = append(unassignedDevices, device)
		}
	}

	// Build unassigned client group if there are unassigned devices
	if len(unassignedDevices) > 0 {
		unassignedClient := map[string]interface{}{
			"client_id": "",
			"name":      "[Unassigned]",
			"comments":  "Devices not assigned to any specific client",
			"devices":   make([]map[string]interface{}, 0),
		}

		for _, device := range unassignedDevices {
			// Get agents for this device
			agentsEndpoint := fmt.Sprintf("/v1/agent?device_id=%s&limit=50", device.DeviceID)
			agentsData, err := makeAPIRequest("GET", agentsEndpoint, nil)
			if err != nil {
				// Continue even if agents call fails - just log the device without agents
				deviceInfo := map[string]interface{}{
					"device_id":      device.DeviceID,
					"hostname":       device.Hostname,
					"display_name":   device.DisplayName,
					"last_seen_at":   device.LastSeenAt,
					"service_status": device.ServiceStatus,
					"agents":         []map[string]interface{}{},
					"agents_error":   fmt.Sprintf("Failed to get agents: %v", err),
				}
				unassignedClient["devices"] = append(unassignedClient["devices"].([]map[string]interface{}), deviceInfo)
				continue
			}

			var agentsResult PaginatedResponse[Agent]
			if err := json.Unmarshal(agentsData, &agentsResult); err != nil {
				// Continue even if parsing fails
				deviceInfo := map[string]interface{}{
					"device_id":      device.DeviceID,
					"hostname":       device.Hostname,
					"display_name":   device.DisplayName,
					"last_seen_at":   device.LastSeenAt,
					"service_status": device.ServiceStatus,
					"agents":         []map[string]interface{}{},
					"agents_error":   fmt.Sprintf("Failed to parse agents: %v", err),
				}
				unassignedClient["devices"] = append(unassignedClient["devices"].([]map[string]interface{}), deviceInfo)
				continue
			}

			// Build agents list for this device
			agentsList := make([]map[string]interface{}, 0)
			for _, agent := range agentsResult.Data {
				agentInfo := map[string]interface{}{
					"agent_id":     agent.AgentID,
					"display_name": agent.DisplayName,
					"hostname":     agent.Hostname,
					"last_seen_at": agent.LastSeenAt,
					"platform":     agent.Platform,
					"os":           agent.OS,
					"os_version":   agent.OSVersion,
				}
				agentsList = append(agentsList, agentInfo)
			}

			deviceInfo := map[string]interface{}{
				"device_id":      device.DeviceID,
				"hostname":       device.Hostname,
				"display_name":   device.DisplayName,
				"last_seen_at":   device.LastSeenAt,
				"service_status": device.ServiceStatus,
				"agents":         agentsList,
			}
			unassignedClient["devices"] = append(unassignedClient["devices"].([]map[string]interface{}), deviceInfo)
		}

		clientsHierarchy = append(clientsHierarchy, unassignedClient)
	}

	// Process each client
	for _, client := range clientsResult.Data {
		clientInfo := map[string]interface{}{
			"client_id": client.ClientID,
			"name":      client.Name,
			"comments":  client.Comments,
			"devices":   make([]map[string]interface{}, 0),
		}

		// Get devices for this client
		devicesEndpoint := fmt.Sprintf("/v1/device?client_id=%s&limit=50", client.ClientID)
		devicesData, err := makeAPIRequest("GET", devicesEndpoint, nil)
		if err != nil {
			// Continue even if devices call fails
			clientInfo["devices_error"] = fmt.Sprintf("Failed to get devices: %v", err)
		} else {
			var devicesResult PaginatedResponse[Device]
			if err := json.Unmarshal(devicesData, &devicesResult); err != nil {
				clientInfo["devices_error"] = fmt.Sprintf("Failed to parse devices: %v", err)
			} else {
				// Process each device for this client
				for _, device := range devicesResult.Data {
					// Get agents for this device
					agentsEndpoint := fmt.Sprintf("/v1/agent?device_id=%s&limit=50", device.DeviceID)
					agentsData, err := makeAPIRequest("GET", agentsEndpoint, nil)
					if err != nil {
						// Continue even if agents call fails
						deviceInfo := map[string]interface{}{
							"device_id":      device.DeviceID,
							"hostname":       device.Hostname,
							"display_name":   device.DisplayName,
							"last_seen_at":   device.LastSeenAt,
							"service_status": device.ServiceStatus,
							"agents":         []map[string]interface{}{},
							"agents_error":   fmt.Sprintf("Failed to get agents: %v", err),
						}
						clientInfo["devices"] = append(clientInfo["devices"].([]map[string]interface{}), deviceInfo)
						continue
					}

					var agentsResult PaginatedResponse[Agent]
					if err := json.Unmarshal(agentsData, &agentsResult); err != nil {
						// Continue even if parsing fails
						deviceInfo := map[string]interface{}{
							"device_id":      device.DeviceID,
							"hostname":       device.Hostname,
							"display_name":   device.DisplayName,
							"last_seen_at":   device.LastSeenAt,
							"service_status": device.ServiceStatus,
							"agents":         []map[string]interface{}{},
							"agents_error":   fmt.Sprintf("Failed to parse agents: %v", err),
						}
						clientInfo["devices"] = append(clientInfo["devices"].([]map[string]interface{}), deviceInfo)
						continue
					}

					// Build agents list for this device
					agentsList := make([]map[string]interface{}, 0)
					for _, agent := range agentsResult.Data {
						agentInfo := map[string]interface{}{
							"agent_id":     agent.AgentID,
							"display_name": agent.DisplayName,
							"hostname":     agent.Hostname,
							"last_seen_at": agent.LastSeenAt,
							"platform":     agent.Platform,
							"os":           agent.OS,
							"os_version":   agent.OSVersion,
						}
						agentsList = append(agentsList, agentInfo)
					}

					deviceInfo := map[string]interface{}{
						"device_id":      device.DeviceID,
						"hostname":       device.Hostname,
						"display_name":   device.DisplayName,
						"last_seen_at":   device.LastSeenAt,
						"service_status": device.ServiceStatus,
						"agents":         agentsList,
					}
					clientInfo["devices"] = append(clientInfo["devices"].([]map[string]interface{}), deviceInfo)
				}
			}
		}

		clientsHierarchy = append(clientsHierarchy, clientInfo)
	}

	// Create the final response with metadata
	enhancedResult := map[string]interface{}{
		"clients": clientsHierarchy,
		"_metadata": map[string]interface{}{
			"description":           "Complete hierarchy of clients, devices, and agents",
			"structure":             "clients -> devices -> agents",
			"client_identification": "Clients are identified by name. '[Unassigned]' represents devices not assigned to any client.",
			"device_identification": "Devices are identified by hostname (the primary human-readable identifier).",
			"agent_identification":  "Agents are identified by display_name (or hostname if display_name is empty).",
			"relationships": map[string]interface{}{
				"client_to_device": "One client can have multiple devices",
				"device_to_agent":  "One device can have multiple agents (backup software instances)",
			},
			"workflow_guidance": "This gives you a complete overview of the backup infrastructure. Clients are organizational units (often customers in MSP scenarios), devices are the Slide appliances, and agents are the backup software installed on computers.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}
