package main

// Hierarchy helper preserved from the retired tools_meta.go. Used by
// slide_overview's `inventory` operation and by the slide://overview
// resource so the startup context is still cheap to fetch.

import (
	"encoding/json"
	"fmt"
)

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
