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
		// Check if reporting or presentation tools are enabled
		if !config.EnableReports && !config.EnablePresentation {
			return "", fmt.Errorf("get_snapshot_changes operation requires reporting or presentation tools to be enabled")
		}
		return getSnapshotChanges(args)
	case "get_reporting_data":
		// Check if reporting or presentation tools are enabled
		if !config.EnableReports && !config.EnablePresentation {
			return "", fmt.Errorf("get_reporting_data operation requires reporting or presentation tools to be enabled")
		}
		return getReportingData(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getMetaToolInfo returns the tool definition for the meta tools
func getMetaToolInfo() ToolInfo {
	// Base operations that are always available
	baseOperations := []string{"list_all_clients_devices_and_agents"}

	// Add reporting operations only if reporting or presentation tools are enabled
	if config.EnableReports || config.EnablePresentation {
		baseOperations = append(baseOperations, "get_snapshot_changes", "get_reporting_data")
	}

	return ToolInfo{
		Name:        "slide_meta",
		Description: "Meta tools for reporting and aggregated data views. Provides hierarchical views and time-based snapshot analysis.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        baseOperations,
				},
				// Parameters for get_snapshot_changes (only shown if operation is available)
				"period": map[string]interface{}{
					"type":        "string",
					"description": "Time period for snapshot changes - used with 'get_snapshot_changes' operation",
					"enum":        []string{"day", "week", "month"},
				},
				"summary_only": map[string]interface{}{
					"type":        "boolean",
					"description": "Return only summary counts without detailed snapshot lists (reduces output size) - default: false",
				},
				"include_metadata": map[string]interface{}{
					"type":        "boolean",
					"description": "Include agent and device names in detailed results (increases output size) - default: false",
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
				// Parameters for get_reporting_data (only shown if operation is available)
				"report_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of report data to generate - used with 'get_reporting_data' operation",
					"enum":        []string{"daily", "weekly", "monthly"},
				},
			},
			"required": []string{"operation"},
			"allOf":    buildConditionalRequirements(baseOperations),
		},
	}
}

// buildConditionalRequirements creates the conditional schema requirements based on available operations
func buildConditionalRequirements(availableOperations []string) []map[string]interface{} {
	var conditions []map[string]interface{}

	// Only add get_snapshot_changes requirements if the operation is available
	for _, op := range availableOperations {
		if op == "get_snapshot_changes" {
			conditions = append(conditions, map[string]interface{}{
				"if": map[string]interface{}{
					"properties": map[string]interface{}{
						"operation": map[string]interface{}{"const": "get_snapshot_changes"},
					},
				},
				"then": map[string]interface{}{
					"required": []string{"period"},
				},
			})
			break
		}
	}

	// Only add get_reporting_data requirements if the operation is available
	for _, op := range availableOperations {
		if op == "get_reporting_data" {
			conditions = append(conditions, map[string]interface{}{
				"if": map[string]interface{}{
					"properties": map[string]interface{}{
						"operation": map[string]interface{}{"const": "get_reporting_data"},
					},
				},
				"then": map[string]interface{}{
					"required": []string{"report_type"},
				},
			})
			break
		}
	}

	return conditions
}

// getSnapshotChanges returns snapshot changes (new and deleted) over a time period
func getSnapshotChanges(args map[string]interface{}) (string, error) {
	period, ok := args["period"].(string)
	if !ok {
		return "", fmt.Errorf("period is required")
	}

	// Check if summary_only is requested (default: false for backward compatibility)
	summaryOnly := false
	if val, ok := args["summary_only"].(bool); ok {
		summaryOnly = val
	}

	// Check if we should include metadata (default: false to reduce size)
	includeMetadata := false
	if val, ok := args["include_metadata"].(bool); ok {
		includeMetadata = val
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
	filters["limit"] = 50

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

	// Initialize counters and aggregated data
	newSnapshotCount := 0
	deletedSnapshotCount := 0
	newSnapshotsByAgent := make(map[string]int)
	deletedSnapshotsByAgent := make(map[string]int)

	// For detailed results (when not summary_only)
	newSnapshots := make([]map[string]interface{}, 0)
	deletedSnapshots := make([]map[string]interface{}, 0)

	// Process snapshots with proper pagination
	processSnapshots := func(location string, isDeleted bool) error {
		offset := 0
		for {
			queryFilters := make(map[string]interface{})
			for k, v := range filters {
				queryFilters[k] = v
			}
			queryFilters["snapshot_location"] = location
			queryFilters["offset"] = offset

			data, err := makeAPIRequest("GET", buildQueryString("/v1/snapshot", queryFilters), nil)
			if err != nil {
				return fmt.Errorf("failed to get snapshots: %w", err)
			}

			var result PaginatedResponse[Snapshot]
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("failed to parse snapshots: %w", err)
			}

			// Process each snapshot
			for _, snapshot := range result.Data {
				if isDeleted && snapshot.Deleted != nil && len(snapshot.Deletions) > 0 {
					// Check deleted snapshots
					for _, deletion := range snapshot.Deletions {
						deletedTime, err := time.Parse(time.RFC3339, deletion.Deleted)
						if err == nil && deletedTime.After(startTime) && deletedTime.Before(now) {
							deletedSnapshotCount++
							deletedSnapshotsByAgent[snapshot.AgentID]++

							if !summaryOnly && len(deletedSnapshots) < 100 { // Limit detailed results
								snapshotInfo := map[string]interface{}{
									"snapshot_id":     snapshot.SnapshotID,
									"agent_id":        snapshot.AgentID,
									"backup_ended_at": snapshot.BackupEndedAt,
									"deleted_at":      deletion.Deleted,
									"deletion_type":   deletion.Type,
								}

								if includeMetadata {
									// Only fetch additional metadata if requested
									if agentData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent/%s", snapshot.AgentID), nil); err == nil {
										var agent Agent
										if json.Unmarshal(agentData, &agent) == nil {
											snapshotInfo["agent_name"] = agent.DisplayName
											snapshotInfo["device_id"] = agent.DeviceID
										}
									}
								}

								deletedSnapshots = append(deletedSnapshots, snapshotInfo)
							}
						}
					}
				} else if !isDeleted {
					// Check new snapshots
					backupTime, err := time.Parse(time.RFC3339, snapshot.BackupEndedAt)
					if err == nil && backupTime.After(startTime) && backupTime.Before(now) {
						newSnapshotCount++
						newSnapshotsByAgent[snapshot.AgentID]++

						if !summaryOnly && len(newSnapshots) < 100 { // Limit detailed results
							snapshotInfo := map[string]interface{}{
								"snapshot_id":       snapshot.SnapshotID,
								"agent_id":          snapshot.AgentID,
								"backup_started_at": snapshot.BackupStartedAt,
								"backup_ended_at":   snapshot.BackupEndedAt,
							}

							if includeMetadata {
								// Only fetch additional metadata if requested
								if agentData, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent/%s", snapshot.AgentID), nil); err == nil {
									var agent Agent
									if json.Unmarshal(agentData, &agent) == nil {
										snapshotInfo["agent_name"] = agent.DisplayName
										snapshotInfo["device_id"] = agent.DeviceID
									}
								}
							}

							newSnapshots = append(newSnapshots, snapshotInfo)
						}
					}
				}
			}

			// Check if we have more pages
			if result.Pagination.NextOffset == nil {
				break
			}
			offset = *result.Pagination.NextOffset
		}
		return nil
	}

	// Process active snapshots for new ones
	if err := processSnapshots("exists_cloud", false); err != nil {
		return "", err
	}

	// Process deleted snapshots
	if err := processSnapshots("exists_deleted", true); err != nil {
		return "", err
	}

	// Build the result based on summary_only flag
	var result map[string]interface{}

	if summaryOnly {
		// Return only summary data
		result = map[string]interface{}{
			"period":     period,
			"start_time": startTime.Format(time.RFC3339),
			"end_time":   now.Format(time.RFC3339),
			"summary": map[string]interface{}{
				"total_new":              newSnapshotCount,
				"total_deleted":          deletedSnapshotCount,
				"new_by_agent_count":     len(newSnapshotsByAgent),
				"deleted_by_agent_count": len(deletedSnapshotsByAgent),
			},
			"_metadata": map[string]interface{}{
				"description": fmt.Sprintf("Summary of snapshot changes over the last %s", period),
				"mode":        "summary_only",
				"note":        "Use summary_only=false to get detailed snapshot lists (limited to 100 each)",
			},
		}
	} else {
		// Return detailed data (limited)
		result = map[string]interface{}{
			"period":            period,
			"start_time":        startTime.Format(time.RFC3339),
			"end_time":          now.Format(time.RFC3339),
			"new_snapshots":     newSnapshots,
			"deleted_snapshots": deletedSnapshots,
			"summary": map[string]interface{}{
				"total_new":              newSnapshotCount,
				"total_deleted":          deletedSnapshotCount,
				"shown_new":              len(newSnapshots),
				"shown_deleted":          len(deletedSnapshots),
				"new_by_agent_count":     len(newSnapshotsByAgent),
				"deleted_by_agent_count": len(deletedSnapshotsByAgent),
			},
			"_metadata": map[string]interface{}{
				"description":   fmt.Sprintf("Snapshot changes over the last %s", period),
				"mode":          "detailed",
				"limit_notice":  "Detailed results are limited to 100 snapshots each to prevent excessive data",
				"metadata_mode": includeMetadata,
			},
		}

		// Add notice if we hit the limit
		if newSnapshotCount > 100 || deletedSnapshotCount > 100 {
			result["_metadata"].(map[string]interface{})["truncated"] = true
			result["_metadata"].(map[string]interface{})["truncation_message"] = fmt.Sprintf("Showing %d of %d new and %d of %d deleted snapshots",
				len(newSnapshots), newSnapshotCount, len(deletedSnapshots), deletedSnapshotCount)
		}
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

	// Get snapshot changes for the period (summary only for performance)
	changesData, err := getSnapshotChanges(map[string]interface{}{
		"period":       period,
		"summary_only": true,
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
		"snapshot_summary": changes["summary"],
		"_metadata": map[string]interface{}{
			"description": fmt.Sprintf("Pre-formatted data for %s report template", reportType),
			"usage":       "Use this data to populate report templates. All metrics are pre-calculated and formatted for easy insertion.",
			"note":        "Full hierarchy data not included by default to reduce size. Use list_all_clients_devices_and_agents separately if needed.",
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
