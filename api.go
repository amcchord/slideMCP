package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const APIBaseURL = "https://api.slide.tech"

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	apiKey string
)

// Data structures
type Device struct {
	DeviceID                string   `json:"device_id"`
	DisplayName             string   `json:"display_name"`
	LastSeenAt              string   `json:"last_seen_at"`
	Hostname                string   `json:"hostname"`
	IPAddresses             []string `json:"ip_addresses"`
	Addresses               []any    `json:"addresses"`
	PublicIPAddress         string   `json:"public_ip_address"`
	ImageVersion            string   `json:"image_version"`
	PackageVersion          string   `json:"package_version"`
	StorageUsedBytes        int64    `json:"storage_used_bytes"`
	StorageTotalBytes       int64    `json:"storage_total_bytes"`
	SerialNumber            string   `json:"serial_number"`
	HardwareModelName       string   `json:"hardware_model_name"`
	ServiceModelName        string   `json:"service_model_name"`
	ServiceModelNameShort   string   `json:"service_model_name_short"`
	ServiceStatus           string   `json:"service_status"`
	NFR                     bool     `json:"nfr"`
	ClientID                *string  `json:"client_id,omitempty"`
	BootedAt                *string  `json:"booted_at,omitempty"`
}

type Agent struct {
	AgentID         string   `json:"agent_id"`
	DeviceID        string   `json:"device_id"`
	DisplayName     string   `json:"display_name"`
	LastSeenAt      string   `json:"last_seen_at"`
	Hostname        string   `json:"hostname"`
	IPAddresses     []string `json:"ip_addresses"`
	Addresses       []any    `json:"addresses"`
	PublicIPAddress string   `json:"public_ip_address"`
	AgentVersion    string   `json:"agent_version"`
	Platform        string   `json:"platform"`
	OS              string   `json:"os"`
	OSVersion       string   `json:"os_version"`
	FirmwareType    string   `json:"firmware_type"`
	Manufacturer    *string  `json:"manufacturer,omitempty"`
	ClientID        *string  `json:"client_id,omitempty"`
	BootedAt        *string  `json:"booted_at,omitempty"`
}

type Pagination struct {
	Total      int  `json:"total"`
	NextOffset *int `json:"next_offset,omitempty"`
}

type PaginatedResponse[T any] struct {
	Pagination Pagination `json:"pagination"`
	Data       []T        `json:"data"`
}

// Helper function to generate VNC viewer URL
func generateVNCViewerURL(virtID, websocketURI, vncPassword string) string {
	encodedWebsocketURI := url.QueryEscape(websocketURI)
	base64Password := base64.StdEncoding.EncodeToString([]byte(vncPassword))
	return fmt.Sprintf("https://slide.recipes/mcpTools/vncViewer.php?id=%s&ws=%s&password=%s&encoding=base64",
		virtID, encodedWebsocketURI, base64Password)
}

// API client helper
func makeAPIRequest(method, endpoint string, body []byte) ([]byte, error) {
	url := APIBaseURL + endpoint

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// API implementations
func listDevices(args map[string]interface{}) (string, error) {
	params := url.Values{}

	if limit, ok := args["limit"]; ok {
		if l, ok := limit.(float64); ok {
			params.Set("limit", strconv.Itoa(int(l)))
		}
	}
	if offset, ok := args["offset"]; ok {
		if o, ok := offset.(float64); ok {
			params.Set("offset", strconv.Itoa(int(o)))
		}
	}
	if clientID, ok := args["client_id"]; ok {
		if cid, ok := clientID.(string); ok {
			params.Set("client_id", cid)
		}
	}
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}

	endpoint := "/v1/devices"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Device]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Add metadata for better LLM interaction
	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":      "hostname",
			"presentation_guidance":   "When referring to devices, use the hostname as the primary identifier. Device IDs are internal identifiers not commonly used by humans.",
			"workflow_guidance":       "Devices are the physical machines running the Slide appliance. Agents are the backup software installed on computers that get backed up to these devices.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func listAgents(args map[string]interface{}) (string, error) {
	params := url.Values{}

	if limit, ok := args["limit"]; ok {
		if l, ok := limit.(float64); ok {
			params.Set("limit", strconv.Itoa(int(l)))
		}
	}
	if offset, ok := args["offset"]; ok {
		if o, ok := offset.(float64); ok {
			params.Set("offset", strconv.Itoa(int(o)))
		}
	}
	if deviceID, ok := args["device_id"]; ok {
		if did, ok := deviceID.(string); ok {
			params.Set("device_id", did)
		}
	}
	if clientID, ok := args["client_id"]; ok {
		if cid, ok := clientID.(string); ok {
			params.Set("client_id", cid)
		}
	}
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	}

	endpoint := "/v1/agents"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Agent]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Add metadata for better LLM interaction
	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":      "display_name",
			"presentation_guidance":   "When referring to agents, use the display name as the primary identifier. If display name is blank, use hostname instead. Agent IDs are internal identifiers not commonly used by humans.",
			"workflow_guidance":       "Agents are backup software installed on computers. They connect to devices (Slide appliances) to store backups.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}