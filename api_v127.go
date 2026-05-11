package main

// API client wrappers for endpoints introduced in Slide API v1.27.0.
// Older endpoints (devices, agents, alerts, etc.) live in api.go; this file
// exists only to keep the diff readable when extending coverage.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// -------------------------------------------------------------------------
// Agent services (GET/PATCH /v1/agent/{agent_id}/service)
// -------------------------------------------------------------------------

type agentServicesResponse struct {
	Data []AgentService `json:"data"`
}

type agentServicesPatchBody struct {
	Services []AgentServiceUpdateItem `json:"services"`
}

type agentServicesPatchResponse struct {
	Data []AgentServiceUpdateItem `json:"data"`
}

func listAgentServices(agentID string) ([]AgentService, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	body, err := makeAPIRequest("GET", fmt.Sprintf("/v1/agent/%s/service", agentID), nil)
	if err != nil {
		return nil, err
	}
	var resp agentServicesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse agent services response: %w", err)
	}
	return resp.Data, nil
}

func updateAgentServices(agentID string, items []AgentServiceUpdateItem) ([]AgentServiceUpdateItem, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("services must contain at least one entry")
	}
	payload, err := json.Marshal(agentServicesPatchBody{Services: items})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal services payload: %w", err)
	}
	body, err := makeAPIRequest("PATCH", fmt.Sprintf("/v1/agent/%s/service", agentID), payload)
	if err != nil {
		return nil, err
	}
	var resp agentServicesPatchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse agent services patch response: %w", err)
	}
	return resp.Data, nil
}

// -------------------------------------------------------------------------
// Snapshot service verification (GET /v1/snapshot/{id}/service-verification)
// -------------------------------------------------------------------------

type SnapshotServiceVerification struct {
	Status   string                      `json:"status,omitempty"`
	Services []ServiceVerificationResult `json:"services"`
}

func getSnapshotServiceVerification(snapshotID string) (*SnapshotServiceVerification, error) {
	if snapshotID == "" {
		return nil, fmt.Errorf("snapshot_id is required")
	}
	body, err := makeAPIRequest("GET", fmt.Sprintf("/v1/snapshot/%s/service-verification", snapshotID), nil)
	if err != nil {
		return nil, err
	}
	var resp SnapshotServiceVerification
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse service verification response: %w", err)
	}
	return &resp, nil
}

// -------------------------------------------------------------------------
// User avatar (GET /v1/user/{user_id}/avatar)
// -------------------------------------------------------------------------

func getUserAvatar(userID string) (*Avatar, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	body, err := makeAPIRequest("GET", fmt.Sprintf("/v1/user/%s/avatar", userID), nil)
	if err != nil {
		return nil, err
	}
	var resp Avatar
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse avatar response: %w", err)
	}
	return &resp, nil
}

// -------------------------------------------------------------------------
// Device network (GET/PATCH /v1/device/{device_id}/network)
// -------------------------------------------------------------------------

func getDeviceNetwork(deviceID string) (*DeviceNetwork, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	body, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device/%s/network", deviceID), nil)
	if err != nil {
		return nil, err
	}
	var resp DeviceNetwork
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device network response: %w", err)
	}
	return &resp, nil
}

func updateDeviceNetwork(deviceID string, payload map[string]interface{}) (*DeviceNetwork, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device network update: %w", err)
	}
	resp, err := makeAPIRequest("PATCH", fmt.Sprintf("/v1/device/%s/network", deviceID), body)
	if err != nil {
		return nil, err
	}
	var out DeviceNetwork
	if err := json.Unmarshal(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to parse device network update response: %w", err)
	}
	return &out, nil
}

// -------------------------------------------------------------------------
// Device VLANs (CRUD /v1/device/{device_id}/vlan[/{vlan_id}])
// -------------------------------------------------------------------------

type deviceVLANListResponse struct {
	Data []DeviceVLAN `json:"data"`
}

func listDeviceVLANs(deviceID string) ([]DeviceVLAN, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	body, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device/%s/vlan", deviceID), nil)
	if err != nil {
		return nil, err
	}
	var resp deviceVLANListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device VLAN list response: %w", err)
	}
	return resp.Data, nil
}

func getDeviceVLAN(deviceID, vlanID string) (*DeviceVLAN, error) {
	if deviceID == "" || vlanID == "" {
		return nil, fmt.Errorf("device_id and vlan_id are required")
	}
	body, err := makeAPIRequest("GET", fmt.Sprintf("/v1/device/%s/vlan/%s", deviceID, vlanID), nil)
	if err != nil {
		return nil, err
	}
	var resp DeviceVLAN
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device VLAN response: %w", err)
	}
	return &resp, nil
}

func createDeviceVLAN(deviceID string, payload map[string]interface{}) (*DeviceVLAN, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VLAN create payload: %w", err)
	}
	resp, err := makeAPIRequest("POST", fmt.Sprintf("/v1/device/%s/vlan", deviceID), body)
	if err != nil {
		return nil, err
	}
	var out DeviceVLAN
	if err := json.Unmarshal(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to parse VLAN create response: %w", err)
	}
	return &out, nil
}

func updateDeviceVLAN(deviceID, vlanID string, payload map[string]interface{}) (*DeviceVLAN, error) {
	if deviceID == "" || vlanID == "" {
		return nil, fmt.Errorf("device_id and vlan_id are required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VLAN update payload: %w", err)
	}
	resp, err := makeAPIRequest("PATCH", fmt.Sprintf("/v1/device/%s/vlan/%s", deviceID, vlanID), body)
	if err != nil {
		return nil, err
	}
	var out DeviceVLAN
	if err := json.Unmarshal(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to parse VLAN update response: %w", err)
	}
	return &out, nil
}

func deleteDeviceVLAN(deviceID, vlanID string) error {
	if deviceID == "" || vlanID == "" {
		return fmt.Errorf("device_id and vlan_id are required")
	}
	if _, err := makeAPIRequest("DELETE", fmt.Sprintf("/v1/device/%s/vlan/%s", deviceID, vlanID), nil); err != nil {
		return err
	}
	return nil
}

// -------------------------------------------------------------------------
// Agent PATCH helpers (the agent endpoint accepts many new fields)
// -------------------------------------------------------------------------

// patchAgent issues PATCH /v1/agent/{id} with the provided fields and returns
// the resulting Agent. Centralizing this avoids each tools_agents.go operation
// re-marshaling and re-parsing.
func patchAgent(agentID string, payload map[string]interface{}) (*Agent, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("at least one field must be provided")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent patch: %w", err)
	}
	resp, err := makeAPIRequest("PATCH", fmt.Sprintf("/v1/agent/%s", agentID), body)
	if err != nil {
		return nil, err
	}
	var out Agent
	if err := json.Unmarshal(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to parse agent patch response: %w", err)
	}
	return &out, nil
}

// patchAgentJSON is a convenience that returns the full Agent as a pretty
// JSON string, suitable for direct use as a tool result body.
func patchAgentJSON(agentID string, payload map[string]interface{}) (string, error) {
	agent, err := patchAgent(agentID, payload)
	if err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal agent: %w", err)
	}
	return string(out), nil
}

// validateBackupSchedule sanity-checks BackupSchedule values before sending.
func validateBackupSchedule(s BackupSchedule) error {
	allowedIntervals := map[int]bool{30: true, 60: true, 120: true, 180: true, 360: true}
	if !allowedIntervals[s.IntervalInMinutes] {
		return fmt.Errorf("interval_in_minutes must be one of 30, 60, 120, 180, 360 (got %d)", s.IntervalInMinutes)
	}
	if s.StartHour < 0 || s.StartHour > 23 {
		return fmt.Errorf("start_hour must be 0-23 (got %d)", s.StartHour)
	}
	if s.EndHour < 0 || s.EndHour > 23 {
		return fmt.Errorf("end_hour must be 0-23 (got %d)", s.EndHour)
	}
	if len(s.Days) == 0 {
		return fmt.Errorf("days must include at least one day-of-week (0=Sun..6=Sat)")
	}
	seen := map[int]bool{}
	for _, d := range s.Days {
		if d < 0 || d > 6 {
			return fmt.Errorf("days entries must be 0-6 (got %d)", d)
		}
		if seen[d] {
			return fmt.Errorf("days must not contain duplicates (got %d twice)", d)
		}
		seen[d] = true
	}
	return nil
}

// validateRetentionPolicyName ensures the policy name matches one of the
// API-allowed values.
func validateRetentionPolicyName(name string) error {
	switch strings.ToLower(name) {
	case "lean", "balanced", "comprehensive":
		return nil
	default:
		return fmt.Errorf("retention_policy_name must be one of: lean, balanced, comprehensive (got %q)", name)
	}
}
