package main

// Hierarchy helpers used by slide_overview's inventory operation and the
// slide://overview resource. Large MSP accounts must stay below host tool-call
// deadlines, so complete inventory is assembled from three bulk paginated
// reads rather than one API request per client and device.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// fetchAllPaginated follows Slide's next_offset cursor and refuses malformed
// pagination loops. The API caps limit at 50, so account-wide operations must
// not silently treat the first page as the whole account.
func fetchAllPaginated[T any](endpoint string) ([]T, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse paginated endpoint %q: %w", endpoint, err)
	}
	query := u.Query()
	query.Set("limit", "50")
	offset := 0
	if raw := query.Get("offset"); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 0 {
			return nil, fmt.Errorf("invalid pagination offset %q in %s", raw, endpoint)
		}
		offset = parsed
	}

	items := make([]T, 0)
	for {
		query.Set("offset", strconv.Itoa(offset))
		u.RawQuery = query.Encode()
		body, requestErr := makeAPIRequest("GET", u.String(), nil)
		if requestErr != nil {
			return nil, requestErr
		}
		var page PaginatedResponse[T]
		if unmarshalErr := json.Unmarshal(body, &page); unmarshalErr != nil {
			return nil, fmt.Errorf("parse %s page at offset %d: %w", u.Path, offset, unmarshalErr)
		}
		items = append(items, page.Data...)
		if len(items) > maxPaginatedEntities {
			return nil, fmt.Errorf("%s returned more than the safety limit of %d entities", u.Path, maxPaginatedEntities)
		}
		if page.Pagination.NextOffset == nil {
			return items, nil
		}
		next := *page.Pagination.NextOffset
		if next <= offset {
			return nil, fmt.Errorf("%s returned non-advancing next_offset %d after %d", u.Path, next, offset)
		}
		offset = next
	}
}

func fetchInventoryEntities() ([]Client, []Device, []Agent, error) {
	var (
		clients []Client
		devices []Device
		agents  []Agent
		errs    = make([]error, 3)
		wg      sync.WaitGroup
	)
	wg.Add(3)
	go func() {
		defer wg.Done()
		clients, errs[0] = fetchAllPaginated[Client]("/v1/client")
	}()
	go func() {
		defer wg.Done()
		devices, errs[1] = fetchAllPaginated[Device]("/v1/device")
	}()
	go func() {
		defer wg.Done()
		agents, errs[2] = fetchAllPaginated[Agent]("/v1/agent")
	}()
	wg.Wait()

	labels := []string{"clients", "devices", "agents"}
	for i, fetchErr := range errs {
		if fetchErr != nil {
			return nil, nil, nil, fmt.Errorf("failed to get %s: %w", labels[i], fetchErr)
		}
	}
	return clients, devices, agents, nil
}

func listAllClientsDevicesAndAgents(_ map[string]interface{}) (string, error) {
	clients, devices, agents, err := fetchInventoryEntities()
	if err != nil {
		return "", err
	}

	sort.Slice(clients, func(i, j int) bool {
		return strings.ToLower(clients[i].Name)+clients[i].ClientID < strings.ToLower(clients[j].Name)+clients[j].ClientID
	})
	sort.Slice(devices, func(i, j int) bool {
		return strings.ToLower(devices[i].Hostname)+devices[i].DeviceID < strings.ToLower(devices[j].Hostname)+devices[j].DeviceID
	})
	sort.Slice(agents, func(i, j int) bool {
		return strings.ToLower(agents[i].Hostname)+agents[i].AgentID < strings.ToLower(agents[j].Hostname)+agents[j].AgentID
	})

	knownDevices := make(map[string]struct{}, len(devices))
	for _, device := range devices {
		knownDevices[device.DeviceID] = struct{}{}
	}
	agentsByDevice := make(map[string][]map[string]interface{}, len(devices))
	orphanAgents := make([]map[string]interface{}, 0)
	for _, agent := range agents {
		agentInfo := map[string]interface{}{
			"agent_id":     agent.AgentID,
			"display_name": agent.DisplayName,
			"hostname":     agent.Hostname,
			"last_seen_at": agent.LastSeenAt,
			"platform":     agent.Platform,
			"os":           agent.OS,
			"os_version":   agent.OSVersion,
		}
		if _, ok := knownDevices[agent.DeviceID]; !ok {
			agentInfo["device_id"] = agent.DeviceID
			orphanAgents = append(orphanAgents, agentInfo)
			continue
		}
		agentsByDevice[agent.DeviceID] = append(agentsByDevice[agent.DeviceID], agentInfo)
	}

	devicesByClient := make(map[string][]map[string]interface{})
	for _, device := range devices {
		clientID := ""
		if device.ClientID != nil {
			clientID = *device.ClientID
		}
		deviceAgents := agentsByDevice[device.DeviceID]
		if deviceAgents == nil {
			deviceAgents = []map[string]interface{}{}
		}
		devicesByClient[clientID] = append(devicesByClient[clientID], map[string]interface{}{
			"device_id":      device.DeviceID,
			"hostname":       device.Hostname,
			"display_name":   device.DisplayName,
			"last_seen_at":   device.LastSeenAt,
			"service_status": device.ServiceStatus,
			"agents":         deviceAgents,
		})
	}

	clientsHierarchy := make([]map[string]interface{}, 0, len(clients)+1)
	if unassigned := devicesByClient[""]; len(unassigned) > 0 {
		clientsHierarchy = append(clientsHierarchy, map[string]interface{}{
			"client_id": "",
			"name":      "[Unassigned]",
			"comments":  "Devices not assigned to any specific client",
			"devices":   unassigned,
		})
		delete(devicesByClient, "")
	}
	for _, client := range clients {
		clientDevices := devicesByClient[client.ClientID]
		if clientDevices == nil {
			clientDevices = []map[string]interface{}{}
		}
		clientsHierarchy = append(clientsHierarchy, map[string]interface{}{
			"client_id": client.ClientID,
			"name":      client.Name,
			"comments":  client.Comments,
			"devices":   clientDevices,
		})
		delete(devicesByClient, client.ClientID)
	}

	// Preserve devices that reference a client hidden from or deleted out of
	// the client listing instead of silently dropping them.
	unknownClientIDs := make([]string, 0, len(devicesByClient))
	for clientID := range devicesByClient {
		unknownClientIDs = append(unknownClientIDs, clientID)
	}
	sort.Strings(unknownClientIDs)
	for _, clientID := range unknownClientIDs {
		clientsHierarchy = append(clientsHierarchy, map[string]interface{}{
			"client_id": clientID,
			"name":      "[Client unavailable]",
			"comments":  "Devices reference a client not returned by the current API token",
			"devices":   devicesByClient[clientID],
		})
	}

	enhancedResult := map[string]interface{}{
		"clients":       clientsHierarchy,
		"orphan_agents": orphanAgents,
		"_metadata": map[string]interface{}{
			"description": "Complete hierarchy of clients, devices, and agents",
			"structure":   "clients -> devices -> agents",
			"complete":    true,
			"counts": map[string]int{
				"clients":       len(clients),
				"devices":       len(devices),
				"agents":        len(agents),
				"orphan_agents": len(orphanAgents),
			},
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
