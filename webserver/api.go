package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const APIBaseURL = "https://api.slide.tech"

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	apiKey string
)

// Function to set API key for each request
func setAPIKey(key string) {
	apiKey = key
}

// Data structures
type Device struct {
	DeviceID              string   `json:"device_id"`
	DisplayName           string   `json:"display_name"`
	LastSeenAt            string   `json:"last_seen_at"`
	Hostname              string   `json:"hostname"`
	IPAddresses           []string `json:"ip_addresses"`
	Addresses             []any    `json:"addresses"`
	PublicIPAddress       string   `json:"public_ip_address"`
	ImageVersion          string   `json:"image_version"`
	PackageVersion        string   `json:"package_version"`
	StorageUsedBytes      int64    `json:"storage_used_bytes"`
	StorageTotalBytes     int64    `json:"storage_total_bytes"`
	SerialNumber          string   `json:"serial_number"`
	HardwareModelName     string   `json:"hardware_model_name"`
	ServiceModelName      string   `json:"service_model_name"`
	ServiceModelNameShort string   `json:"service_model_name_short"`
	ServiceStatus         string   `json:"service_status"`
	NFR                   bool     `json:"nfr"`
	ClientID              *string  `json:"client_id,omitempty"`
	BootedAt              *string  `json:"booted_at,omitempty"`
}

type Agent struct {
	AgentID             string   `json:"agent_id"`
	DeviceID            string   `json:"device_id"`
	DisplayName         string   `json:"display_name"`
	LastSeenAt          string   `json:"last_seen_at"`
	Hostname            string   `json:"hostname"`
	IPAddresses         []string `json:"ip_addresses"`
	Addresses           []any    `json:"addresses"`
	PublicIPAddress     string   `json:"public_ip_address"`
	AgentVersion        string   `json:"agent_version"`
	Platform            string   `json:"platform"`
	OS                  string   `json:"os"`
	OSVersion           string   `json:"os_version"`
	FirmwareType        string   `json:"firmware_type"`
	EncryptionAlgorithm *string  `json:"encryption_algorithm,omitempty"`
	Manufacturer        *string  `json:"manufacturer,omitempty"`
	ClientID            *string  `json:"client_id,omitempty"`
	BootedAt            *string  `json:"booted_at,omitempty"`
}

type Pagination struct {
	Total      int  `json:"total"`
	NextOffset *int `json:"next_offset,omitempty"`
}

type PaginatedResponse[T any] struct {
	Pagination Pagination `json:"pagination"`
	Data       []T        `json:"data"`
}

type AgentPairCode struct {
	AgentID     string `json:"agent_id"`
	DisplayName string `json:"display_name"`
	PairCode    string `json:"pair_code"`
}

type Backup struct {
	BackupID     string  `json:"backup_id"`
	AgentID      string  `json:"agent_id"`
	StartedAt    string  `json:"started_at"`
	EndedAt      *string `json:"ended_at,omitempty"`
	Status       string  `json:"status"`
	ErrorCode    *int    `json:"error_code,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
	SnapshotID   *string `json:"snapshot_id,omitempty"`
}

type Location struct {
	Type     string `json:"type"`
	DeviceID string `json:"device_id"`
}

type Deletion struct {
	Type             string  `json:"type"`
	Deleted          string  `json:"deleted"`
	DeletedBy        string  `json:"deleted_by"`
	FirstAndLastName *string `json:"first_and_last_name,omitempty"`
}

type Snapshot struct {
	SnapshotID              string     `json:"snapshot_id"`
	AgentID                 string     `json:"agent_id"`
	Locations               []Location `json:"locations"`
	BackupStartedAt         string     `json:"backup_started_at"`
	BackupEndedAt           string     `json:"backup_ended_at"`
	Deleted                 *string    `json:"deleted,omitempty"`
	Deletions               []Deletion `json:"deletions,omitempty"`
	VerifyBootStatus        *string    `json:"verify_boot_status,omitempty"`
	VerifyFsStatus          *string    `json:"verify_fs_status,omitempty"`
	VerifyBootScreenshotURL *string    `json:"verify_boot_screenshot_url,omitempty"`
}

type FileRestore struct {
	FileRestoreID string  `json:"file_restore_id"`
	DeviceID      string  `json:"device_id"`
	AgentID       string  `json:"agent_id"`
	SnapshotID    string  `json:"snapshot_id"`
	CreatedAt     string  `json:"created_at"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
}

type DownloadURI struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type FileRestoreEntry struct {
	Name              string        `json:"name"`
	Path              string        `json:"path"`
	Size              int64         `json:"size"`
	Type              string        `json:"type"`
	ModifiedAt        string        `json:"modified_at"`
	DownloadURIs      []DownloadURI `json:"download_uris"`
	SymlinkTargetPath *string       `json:"symlink_target_path,omitempty"`
}

type ImageExport struct {
	ImageExportID string `json:"image_export_id"`
	DeviceID      string `json:"device_id"`
	AgentID       string `json:"agent_id"`
	SnapshotID    string `json:"snapshot_id"`
	ImageType     string `json:"image_type"`
	CreatedAt     string `json:"created_at"`
}

type ImageExportEntry struct {
	DiskID       string        `json:"disk_id"`
	Name         string        `json:"name"`
	Size         int64         `json:"size"`
	DownloadURIs []DownloadURI `json:"download_uris"`
}

type VNC struct {
	Type         string  `json:"type"`
	Host         *string `json:"host,omitempty"`
	Port         *int    `json:"port,omitempty"`
	WebsocketURI *string `json:"websocket_uri,omitempty"`
}

type VirtualMachine struct {
	VirtID        string  `json:"virt_id"`
	DeviceID      string  `json:"device_id"`
	AgentID       string  `json:"agent_id"`
	SnapshotID    string  `json:"snapshot_id"`
	State         string  `json:"state"`
	CreatedAt     string  `json:"created_at"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	CPUCount      int     `json:"cpu_count"`
	MemoryInMB    int     `json:"memory_in_mb"`
	DiskBus       string  `json:"disk_bus"`
	NetworkModel  string  `json:"network_model"`
	NetworkType   *string `json:"network_type,omitempty"`
	NetworkSource *string `json:"network_source,omitempty"`
	VNC           []VNC   `json:"vnc"`
	VNCPassword   string  `json:"vnc_password"`
}

type User struct {
	UserID      string `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	RoleID      string `json:"role_id"`
}

type Alert struct {
	AlertID     string  `json:"alert_id"`
	AlertType   string  `json:"alert_type"`
	AlertFields string  `json:"alert_fields"`
	CreatedAt   string  `json:"created_at"`
	Resolved    bool    `json:"resolved"`
	ResolvedAt  *string `json:"resolved_at,omitempty"`
	ResolvedBy  *string `json:"resolved_by,omitempty"`
	DeviceID    *string `json:"device_id,omitempty"`
	AgentID     *string `json:"agent_id,omitempty"`
}

type BillingAddress struct {
	Line1      string  `json:"Line1"`
	Line2      *string `json:"Line2,omitempty"`
	City       string  `json:"City"`
	State      string  `json:"State"`
	PostalCode string  `json:"PostalCode"`
	Country    string  `json:"Country"`
}

type Account struct {
	AccountID      string         `json:"account_id"`
	AccountName    string         `json:"account_name"`
	PrimaryContact string         `json:"primary_contact"`
	PrimaryEmail   string         `json:"primary_email"`
	PrimaryPhone   string         `json:"primary_phone"`
	BillingAddress BillingAddress `json:"billing_address"`
	AlertEmails    []string       `json:"alert_emails"`
}

type Client struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
	Comments string `json:"comments"`
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
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	} else {
		// Set default sort_by as per API spec
		params.Set("sort_by", "hostname")
	}

	endpoint := "/v1/device"
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
			"primary_identifier":    "hostname",
			"presentation_guidance": "When referring to devices, use the hostname as the primary identifier. Device IDs are internal identifiers not commonly used by humans.",
			"workflow_guidance":     "Devices are the physical machines running the Slide appliance. Agents are the backup software installed on computers that get backed up to these devices.",
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

	endpoint := "/v1/agent"
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
			"primary_identifier":    "display_name",
			"presentation_guidance": "When referring to agents, use the display name as the primary identifier. If display name is blank, use hostname instead. Agent IDs are internal identifiers not commonly used by humans.",
			"workflow_guidance":     "Agents are backup software installed on computers. They connect to devices (Slide appliances) to store backups.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Agent API functions
func getAgent(args map[string]interface{}) (string, error) {
	agentID, ok := args["agent_id"].(string)
	if !ok {
		return "", fmt.Errorf("agent_id is required")
	}

	endpoint := fmt.Sprintf("/v1/agent/%s", agentID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Agent
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func createAgent(args map[string]interface{}) (string, error) {
	displayName, ok := args["display_name"].(string)
	if !ok {
		return "", fmt.Errorf("display_name is required")
	}
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	payload := map[string]interface{}{
		"display_name": displayName,
		"device_id":    deviceID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/agent", body)
	if err != nil {
		return "", err
	}

	var result AgentPairCode
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func pairAgent(args map[string]interface{}) (string, error) {
	pairCode, ok := args["pair_code"].(string)
	if !ok {
		return "", fmt.Errorf("pair_code is required")
	}
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	payload := map[string]interface{}{
		"pair_code": pairCode,
		"device_id": deviceID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/agent/pair", body)
	if err != nil {
		return "", err
	}

	var result Agent
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateAgent(args map[string]interface{}) (string, error) {
	agentID, ok := args["agent_id"].(string)
	if !ok {
		return "", fmt.Errorf("agent_id is required")
	}
	displayName, ok := args["display_name"].(string)
	if !ok {
		return "", fmt.Errorf("display_name is required")
	}

	payload := map[string]interface{}{
		"display_name": displayName,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/agent/%s", agentID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result Agent
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Backup API functions
func listBackups(args map[string]interface{}) (string, error) {
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
	if agentID, ok := args["agent_id"]; ok {
		if aid, ok := agentID.(string); ok {
			params.Set("agent_id", aid)
		}
	}
	if deviceID, ok := args["device_id"]; ok {
		if did, ok := deviceID.(string); ok {
			params.Set("device_id", did)
		}
	}
	if snapshotID, ok := args["snapshot_id"]; ok {
		if sid, ok := snapshotID.(string); ok {
			params.Set("snapshot_id", sid)
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
	} else {
		params.Set("sort_by", "start_time")
	}

	endpoint := "/v1/backup"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Backup]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "backup_id",
			"presentation_guidance": "Backups represent backup jobs. Status indicates success/failure. If successful, snapshot_id will be present.",
			"workflow_guidance":     "Backups create snapshots when successful. Failed backups will have error_code and error_message.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getBackup(args map[string]interface{}) (string, error) {
	backupID, ok := args["backup_id"].(string)
	if !ok {
		return "", fmt.Errorf("backup_id is required")
	}

	endpoint := fmt.Sprintf("/v1/backup/%s", backupID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Backup
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func startBackup(args map[string]interface{}) (string, error) {
	agentID, ok := args["agent_id"].(string)
	if !ok {
		return "", fmt.Errorf("agent_id is required")
	}

	payload := map[string]interface{}{
		"agent_id": agentID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/backup", body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Snapshot API functions
func listSnapshots(args map[string]interface{}) (string, error) {
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
	if agentID, ok := args["agent_id"]; ok {
		if aid, ok := agentID.(string); ok {
			params.Set("agent_id", aid)
		}
	}
	if snapshotLocation, ok := args["snapshot_location"]; ok {
		if sl, ok := snapshotLocation.(string); ok {
			params.Set("snapshot_location", sl)
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
	} else {
		params.Set("sort_by", "created")
	}

	endpoint := "/v1/snapshot"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Snapshot]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "snapshot_id",
			"presentation_guidance": "Snapshots are point-in-time backups that can be used for restores. Check locations to see where stored.",
			"workflow_guidance":     "Snapshots can be restored as files, images, or virtual machines. Verify status shows boot/filesystem verification results.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getSnapshot(args map[string]interface{}) (string, error) {
	snapshotID, ok := args["snapshot_id"].(string)
	if !ok {
		return "", fmt.Errorf("snapshot_id is required")
	}

	endpoint := fmt.Sprintf("/v1/snapshot/%s", snapshotID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Snapshot
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// File restore API functions
func listFileRestores(args map[string]interface{}) (string, error) {
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
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	} else {
		params.Set("sort_by", "id")
	}

	endpoint := "/v1/restore/file"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[FileRestore]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "file_restore_id",
			"presentation_guidance": "File restores allow browsing and downloading files from snapshots.",
			"workflow_guidance":     "Create file restores from snapshots, then browse to find and download specific files.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getFileRestore(args map[string]interface{}) (string, error) {
	fileRestoreID, ok := args["file_restore_id"].(string)
	if !ok {
		return "", fmt.Errorf("file_restore_id is required")
	}

	endpoint := fmt.Sprintf("/v1/restore/file/%s", fileRestoreID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result FileRestore
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func createFileRestore(args map[string]interface{}) (string, error) {
	snapshotID, ok := args["snapshot_id"].(string)
	if !ok {
		return "", fmt.Errorf("snapshot_id is required")
	}
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	payload := map[string]interface{}{
		"snapshot_id": snapshotID,
		"device_id":   deviceID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/restore/file", body)
	if err != nil {
		return "", err
	}

	var result FileRestore
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteFileRestore(args map[string]interface{}) (string, error) {
	fileRestoreID, ok := args["file_restore_id"].(string)
	if !ok {
		return "", fmt.Errorf("file_restore_id is required")
	}

	endpoint := fmt.Sprintf("/v1/restore/file/%s", fileRestoreID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "File restore deleted successfully", nil
}

func browseFileRestore(args map[string]interface{}) (string, error) {
	fileRestoreID, ok := args["file_restore_id"].(string)
	if !ok {
		return "", fmt.Errorf("file_restore_id is required")
	}
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path is required")
	}

	params := url.Values{}
	params.Set("path", path)

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

	endpoint := fmt.Sprintf("/v1/restore/file/%s/browse?%s", fileRestoreID, params.Encode())
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[FileRestoreEntry]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "path",
			"presentation_guidance": "File listing from the restored snapshot. Use download_uris to download files.",
			"workflow_guidance":     "Navigate directories by changing the path parameter. Files have download URIs for retrieval.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Image export API functions
func listImageExports(args map[string]interface{}) (string, error) {
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
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	} else {
		params.Set("sort_by", "id")
	}

	endpoint := "/v1/restore/image"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[ImageExport]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "image_export_id",
			"presentation_guidance": "Image exports create downloadable disk images from snapshots.",
			"workflow_guidance":     "Create image exports to get VHDX, VHD, or RAW disk images for importing into other systems.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getImageExport(args map[string]interface{}) (string, error) {
	imageExportID, ok := args["image_export_id"].(string)
	if !ok {
		return "", fmt.Errorf("image_export_id is required")
	}

	endpoint := fmt.Sprintf("/v1/restore/image/%s", imageExportID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result ImageExport
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func createImageExport(args map[string]interface{}) (string, error) {
	snapshotID, ok := args["snapshot_id"].(string)
	if !ok {
		return "", fmt.Errorf("snapshot_id is required")
	}
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}
	imageType, ok := args["image_type"].(string)
	if !ok {
		return "", fmt.Errorf("image_type is required")
	}

	payload := map[string]interface{}{
		"snapshot_id": snapshotID,
		"device_id":   deviceID,
		"image_type":  imageType,
	}

	if bootMods, ok := args["boot_mods"]; ok {
		payload["boot_mods"] = bootMods
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/restore/image", body)
	if err != nil {
		return "", err
	}

	var result ImageExport
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteImageExport(args map[string]interface{}) (string, error) {
	imageExportID, ok := args["image_export_id"].(string)
	if !ok {
		return "", fmt.Errorf("image_export_id is required")
	}

	endpoint := fmt.Sprintf("/v1/restore/image/%s", imageExportID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "Image export deleted successfully", nil
}

func browseImageExport(args map[string]interface{}) (string, error) {
	imageExportID, ok := args["image_export_id"].(string)
	if !ok {
		return "", fmt.Errorf("image_export_id is required")
	}

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

	endpoint := fmt.Sprintf("/v1/restore/image/%s/browse", imageExportID)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[ImageExportEntry]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "disk_id",
			"presentation_guidance": "Available disk images from the export. Use download_uris to download image files.",
			"workflow_guidance":     "Each disk from the original system becomes a separate downloadable image file.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Virtual machine API functions
func listVirtualMachines(args map[string]interface{}) (string, error) {
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
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	} else {
		params.Set("sort_by", "created")
	}

	endpoint := "/v1/restore/virt"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[VirtualMachine]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Process each VM to add VNC viewer URLs
	enhancedVMs := make([]map[string]interface{}, len(result.Data))
	for i, vm := range result.Data {
		vmMap := map[string]interface{}{
			"virt_id":        vm.VirtID,
			"device_id":      vm.DeviceID,
			"agent_id":       vm.AgentID,
			"snapshot_id":    vm.SnapshotID,
			"state":          vm.State,
			"created_at":     vm.CreatedAt,
			"expires_at":     vm.ExpiresAt,
			"cpu_count":      vm.CPUCount,
			"memory_in_mb":   vm.MemoryInMB,
			"disk_bus":       vm.DiskBus,
			"network_model":  vm.NetworkModel,
			"network_type":   vm.NetworkType,
			"network_source": vm.NetworkSource,
			"vnc":            vm.VNC,
			"vnc_password":   vm.VNCPassword,
		}

		// Generate VNC viewer URL if websocket URI is available
		for _, vnc := range vm.VNC {
			if vnc.WebsocketURI != nil {
				vncViewerURL := generateVNCViewerURL(vm.VirtID, *vnc.WebsocketURI, vm.VNCPassword)
				vmMap["_vnc_viewer_url"] = vncViewerURL
				break
			}
		}

		enhancedVMs[i] = vmMap
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       enhancedVMs,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "virt_id",
			"presentation_guidance": "Virtual machines created from snapshots for testing or disaster recovery.",
			"workflow_guidance":     "VMs can be started/stopped and accessed via VNC. Great for testing backups before full restore.",
			"vnc_guidance":          "Each virtual machine includes a _vnc_viewer_url property that provides a direct link to access its console through a browser-based VNC client.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getVirtualMachine(args map[string]interface{}) (string, error) {
	virtID, ok := args["virt_id"].(string)
	if !ok {
		return "", fmt.Errorf("virt_id is required")
	}

	endpoint := fmt.Sprintf("/v1/restore/virt/%s", virtID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result VirtualMachine
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate VNC viewer URL if websocket URI is available
	enhancedResult := map[string]interface{}{
		"virt_id":        result.VirtID,
		"device_id":      result.DeviceID,
		"agent_id":       result.AgentID,
		"snapshot_id":    result.SnapshotID,
		"state":          result.State,
		"created_at":     result.CreatedAt,
		"expires_at":     result.ExpiresAt,
		"cpu_count":      result.CPUCount,
		"memory_in_mb":   result.MemoryInMB,
		"disk_bus":       result.DiskBus,
		"network_model":  result.NetworkModel,
		"network_type":   result.NetworkType,
		"network_source": result.NetworkSource,
		"vnc":            result.VNC,
		"vnc_password":   result.VNCPassword,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "virt_id",
			"presentation_guidance": "When referring to the virtual machine, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
		},
	}

	for _, vnc := range result.VNC {
		if vnc.WebsocketURI != nil {
			vncViewerURL := generateVNCViewerURL(result.VirtID, *vnc.WebsocketURI, result.VNCPassword)
			enhancedResult["_vnc_viewer_url"] = vncViewerURL
			// Add to metadata as well
			enhancedResult["_metadata"].(map[string]interface{})["vnc_guidance"] = "Use the _vnc_viewer_url to access the virtual machine's console via a browser-based VNC client."
			enhancedResult["_metadata"].(map[string]interface{})["vnc_viewer_url"] = vncViewerURL
			break
		}
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func createVirtualMachine(args map[string]interface{}) (string, error) {
	snapshotID, ok := args["snapshot_id"].(string)
	if !ok {
		return "", fmt.Errorf("snapshot_id is required")
	}
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	payload := map[string]interface{}{
		"snapshot_id": snapshotID,
		"device_id":   deviceID,
	}

	// Add optional parameters
	if cpuCount, ok := args["cpu_count"]; ok {
		payload["cpu_count"] = cpuCount
	}
	if memoryInMB, ok := args["memory_in_mb"]; ok {
		payload["memory_in_mb"] = memoryInMB
	}
	if diskBus, ok := args["disk_bus"]; ok {
		payload["disk_bus"] = diskBus
	}
	if networkModel, ok := args["network_model"]; ok {
		payload["network_model"] = networkModel
	}
	if networkType, ok := args["network_type"]; ok {
		payload["network_type"] = networkType
	}
	if networkSource, ok := args["network_source"]; ok {
		payload["network_source"] = networkSource
	}
	if bootMods, ok := args["boot_mods"]; ok {
		payload["boot_mods"] = bootMods
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/restore/virt", body)
	if err != nil {
		return "", err
	}

	var result VirtualMachine
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate VNC viewer URL if websocket URI is available
	enhancedResult := map[string]interface{}{
		"virt_id":        result.VirtID,
		"device_id":      result.DeviceID,
		"agent_id":       result.AgentID,
		"snapshot_id":    result.SnapshotID,
		"state":          result.State,
		"created_at":     result.CreatedAt,
		"expires_at":     result.ExpiresAt,
		"cpu_count":      result.CPUCount,
		"memory_in_mb":   result.MemoryInMB,
		"disk_bus":       result.DiskBus,
		"network_model":  result.NetworkModel,
		"network_type":   result.NetworkType,
		"network_source": result.NetworkSource,
		"vnc":            result.VNC,
		"vnc_password":   result.VNCPassword,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "virt_id",
			"presentation_guidance": "When referring to the virtual machine, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
			"next_steps":            "Now that you've created a virtual machine, you can control it using slide_update_virtual_machine to change its state (running, stopped, paused) or update resources.",
			"resource_guidance":     "For optimal performance, 8192MB of RAM is recommended for most VMs. You can adjust this as needed using slide_update_virtual_machine.",
		},
	}

	for _, vnc := range result.VNC {
		if vnc.WebsocketURI != nil {
			vncViewerURL := generateVNCViewerURL(result.VirtID, *vnc.WebsocketURI, result.VNCPassword)
			enhancedResult["_vnc_viewer_url"] = vncViewerURL
			// Add to metadata as well
			enhancedResult["_metadata"].(map[string]interface{})["vnc_guidance"] = "Use the _vnc_viewer_url to access the virtual machine's console via a browser-based VNC client."
			enhancedResult["_metadata"].(map[string]interface{})["vnc_viewer_url"] = vncViewerURL
			break
		}
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateVirtualMachine(args map[string]interface{}) (string, error) {
	virtID, ok := args["virt_id"].(string)
	if !ok {
		return "", fmt.Errorf("virt_id is required")
	}

	payload := make(map[string]interface{})

	if state, ok := args["state"]; ok {
		payload["state"] = state
	}
	if expiresAt, ok := args["expires_at"]; ok {
		payload["expires_at"] = expiresAt
	}
	if memoryInMB, ok := args["memory_in_mb"]; ok {
		payload["memory_in_mb"] = memoryInMB
	}
	if cpuCount, ok := args["cpu_count"]; ok {
		payload["cpu_count"] = cpuCount
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/restore/virt/%s", virtID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result VirtualMachine
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate VNC viewer URL if websocket URI is available
	enhancedResult := map[string]interface{}{
		"virt_id":        result.VirtID,
		"device_id":      result.DeviceID,
		"agent_id":       result.AgentID,
		"snapshot_id":    result.SnapshotID,
		"state":          result.State,
		"created_at":     result.CreatedAt,
		"expires_at":     result.ExpiresAt,
		"cpu_count":      result.CPUCount,
		"memory_in_mb":   result.MemoryInMB,
		"disk_bus":       result.DiskBus,
		"network_model":  result.NetworkModel,
		"network_type":   result.NetworkType,
		"network_source": result.NetworkSource,
		"vnc":            result.VNC,
		"vnc_password":   result.VNCPassword,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "virt_id",
			"presentation_guidance": "When referring to the virtual machine, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
		},
	}

	for _, vnc := range result.VNC {
		if vnc.WebsocketURI != nil {
			vncViewerURL := generateVNCViewerURL(result.VirtID, *vnc.WebsocketURI, result.VNCPassword)
			enhancedResult["_vnc_viewer_url"] = vncViewerURL
			// Add to metadata as well
			enhancedResult["_metadata"].(map[string]interface{})["vnc_guidance"] = "Use the _vnc_viewer_url to access the virtual machine's console via a browser-based VNC client."
			enhancedResult["_metadata"].(map[string]interface{})["vnc_viewer_url"] = vncViewerURL
			break
		}
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteVirtualMachine(args map[string]interface{}) (string, error) {
	virtID, ok := args["virt_id"].(string)
	if !ok {
		return "", fmt.Errorf("virt_id is required")
	}

	endpoint := fmt.Sprintf("/v1/restore/virt/%s", virtID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "Virtual machine deleted successfully", nil
}

// User API functions
func listUsers(args map[string]interface{}) (string, error) {
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
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	} else {
		params.Set("sort_by", "id")
	}

	endpoint := "/v1/user"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[User]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "display_name",
			"presentation_guidance": "Users with access to the Slide system. Check role_id for permissions level.",
			"workflow_guidance":     "Users can be account owners, admins, technicians, or read-only. Email is used for notifications.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getUser(args map[string]interface{}) (string, error) {
	userID, ok := args["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("user_id is required")
	}

	endpoint := fmt.Sprintf("/v1/user/%s", userID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result User
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Alert API functions
func listAlerts(args map[string]interface{}) (string, error) {
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
	if agentID, ok := args["agent_id"]; ok {
		if aid, ok := agentID.(string); ok {
			params.Set("agent_id", aid)
		}
	}
	if resolved, ok := args["resolved"]; ok {
		if r, ok := resolved.(bool); ok {
			params.Set("resolved", strconv.FormatBool(r))
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
	} else {
		params.Set("sort_by", "created")
	}

	endpoint := "/v1/alert"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Alert]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "alert_id",
			"presentation_guidance": "System alerts for devices not checking in, failed backups, storage issues, etc.",
			"workflow_guidance":     "Alerts can be resolved manually. Check alert_type for the specific issue and alert_fields for details.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getAlert(args map[string]interface{}) (string, error) {
	alertID, ok := args["alert_id"].(string)
	if !ok {
		return "", fmt.Errorf("alert_id is required")
	}

	endpoint := fmt.Sprintf("/v1/alert/%s", alertID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Alert
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateAlert(args map[string]interface{}) (string, error) {
	alertID, ok := args["alert_id"].(string)
	if !ok {
		return "", fmt.Errorf("alert_id is required")
	}
	resolved, ok := args["resolved"].(bool)
	if !ok {
		return "", fmt.Errorf("resolved is required")
	}

	payload := map[string]interface{}{
		"resolved": resolved,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/alert/%s", alertID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result Alert
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Account API functions
func listAccounts(args map[string]interface{}) (string, error) {
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
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	} else {
		params.Set("sort_by", "name")
	}

	endpoint := "/v1/account"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Account]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "account_name",
			"presentation_guidance": "Billing accounts that contain devices and users. Each account has contact info and alert settings.",
			"workflow_guidance":     "Accounts are the top-level organization unit. Devices and users belong to accounts.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getAccount(args map[string]interface{}) (string, error) {
	accountID, ok := args["account_id"].(string)
	if !ok {
		return "", fmt.Errorf("account_id is required")
	}

	endpoint := fmt.Sprintf("/v1/account/%s", accountID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Account
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateAccount(args map[string]interface{}) (string, error) {
	accountID, ok := args["account_id"].(string)
	if !ok {
		return "", fmt.Errorf("account_id is required")
	}
	alertEmails, ok := args["alert_emails"].([]interface{})
	if !ok {
		return "", fmt.Errorf("alert_emails is required")
	}

	// Convert []interface{} to []string
	emailStrings := make([]string, len(alertEmails))
	for i, email := range alertEmails {
		if emailStr, ok := email.(string); ok {
			emailStrings[i] = emailStr
		} else {
			return "", fmt.Errorf("alert_emails must be an array of strings")
		}
	}

	payload := map[string]interface{}{
		"alert_emails": emailStrings,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/account/%s", accountID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result Account
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Client API functions
func listClients(args map[string]interface{}) (string, error) {
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
	if sortAsc, ok := args["sort_asc"]; ok {
		if sa, ok := sortAsc.(bool); ok {
			params.Set("sort_asc", strconv.FormatBool(sa))
		}
	}
	if sortBy, ok := args["sort_by"]; ok {
		if sb, ok := sortBy.(string); ok {
			params.Set("sort_by", sb)
		}
	} else {
		params.Set("sort_by", "id")
	}

	endpoint := "/v1/client"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Client]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       result.Data,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "name",
			"presentation_guidance": "Clients represent end customers or organizational units within an MSP environment.",
			"workflow_guidance":     "Clients group devices and agents for easier management. Useful for MSPs managing multiple customers.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getClient(args map[string]interface{}) (string, error) {
	clientID, ok := args["client_id"].(string)
	if !ok {
		return "", fmt.Errorf("client_id is required")
	}

	endpoint := fmt.Sprintf("/v1/client/%s", clientID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Client
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func createClient(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("name is required")
	}

	payload := map[string]interface{}{
		"name": name,
	}

	if comments, ok := args["comments"]; ok {
		payload["comments"] = comments
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/client", body)
	if err != nil {
		return "", err
	}

	var result Client
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateClient(args map[string]interface{}) (string, error) {
	clientID, ok := args["client_id"].(string)
	if !ok {
		return "", fmt.Errorf("client_id is required")
	}

	payload := make(map[string]interface{})

	if name, ok := args["name"]; ok {
		payload["name"] = name
	}
	if comments, ok := args["comments"]; ok {
		payload["comments"] = comments
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/client/%s", clientID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result Client
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteClient(args map[string]interface{}) (string, error) {
	clientID, ok := args["client_id"].(string)
	if !ok {
		return "", fmt.Errorf("client_id is required")
	}

	endpoint := fmt.Sprintf("/v1/client/%s", clientID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "Client deleted successfully", nil
}

// MCP Protocol structures
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Configuration
const (
	ServerName = "slide-mcp-web-server"
	Version    = "0.1.0"
)

// Web server main function
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Serve static files from current directory
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// Serve index.html for root path
			http.ServeFile(w, r, "index.html")
			return
		}

		// For other paths, check if file exists and serve it
		fileName := r.URL.Path[1:] // Remove leading slash
		if fileName == "slide-mcp-bridge.js" || fileName == "index.html" {
			http.ServeFile(w, r, fileName)
			return
		}

		// If not a static file, return 404
		http.NotFound(w, r)
	})

	http.HandleFunc("/mcp", handleMCPRequest)
	http.HandleFunc("/health", handleHealth)

	log.Printf("Slide MCP Web Server starting on port %s...", port)
	log.Printf("MCP endpoint: http://localhost:%s/mcp", port)
	log.Printf("Documentation: http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"server":  ServerName,
		"version": Version,
	})
}

func handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract API key from request
	var userAPIKey string

	// Try Authorization header first (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		userAPIKey = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try X-API-Key header
	if userAPIKey == "" {
		userAPIKey = r.Header.Get("X-API-Key")
	}

	// Try query parameter
	if userAPIKey == "" {
		userAPIKey = r.URL.Query().Get("api_key")
	}

	if userAPIKey == "" {
		errorResponse := MCPResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: map[string]interface{}{
				"code":    -32602,
				"message": "API key required. Provide via Authorization header (Bearer token), X-API-Key header, or api_key query parameter",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Parse the MCP request
	var request MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errorResponse := MCPResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: map[string]interface{}{
				"code":    -32700,
				"message": "Parse error",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Handle the request
	response := handleRequest(request, userAPIKey)

	// Send response
	json.NewEncoder(w).Encode(response)
}

func handleRequest(request MCPRequest, userAPIKey string) MCPResponse {
	switch request.Method {
	case "initialize":
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    ServerName,
					"version": Version,
				},
			},
		}

	case "tools/list":
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result: map[string]interface{}{
				"tools": getAllTools(),
			},
		}

	case "tools/call":
		return handleToolCall(request, userAPIKey)

	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
		}
	}
}

func handleToolCall(request MCPRequest, userAPIKey string) MCPResponse {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return sendError(request.ID, -32602, "Invalid params", nil)
	}

	name, ok := params["name"].(string)
	if !ok {
		return sendError(request.ID, -32602, "Tool name required", nil)
	}

	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	var result ToolResult

	// Set the API key for this request
	setAPIKey(userAPIKey)

	switch name {
	case "slide_list_devices":
		data, err := listDevices(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_agents":
		data, err := listAgents(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_agent":
		data, err := getAgent(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_create_agent":
		data, err := createAgent(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_pair_agent":
		data, err := pairAgent(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_update_agent":
		data, err := updateAgent(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_backups":
		data, err := listBackups(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_backup":
		data, err := getBackup(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_start_backup":
		data, err := startBackup(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_snapshots":
		data, err := listSnapshots(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_snapshot":
		data, err := getSnapshot(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_file_restores":
		data, err := listFileRestores(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_file_restore":
		data, err := getFileRestore(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_create_file_restore":
		data, err := createFileRestore(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_delete_file_restore":
		data, err := deleteFileRestore(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_browse_file_restore":
		data, err := browseFileRestore(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_image_exports":
		data, err := listImageExports(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_image_export":
		data, err := getImageExport(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_create_image_export":
		data, err := createImageExport(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_delete_image_export":
		data, err := deleteImageExport(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_browse_image_export":
		data, err := browseImageExport(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_virtual_machines":
		data, err := listVirtualMachines(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_virtual_machine":
		data, err := getVirtualMachine(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_create_virtual_machine":
		data, err := createVirtualMachine(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_update_virtual_machine":
		data, err := updateVirtualMachine(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_delete_virtual_machine":
		data, err := deleteVirtualMachine(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_users":
		data, err := listUsers(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_user":
		data, err := getUser(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_alerts":
		data, err := listAlerts(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_alert":
		data, err := getAlert(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_update_alert":
		data, err := updateAlert(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_accounts":
		data, err := listAccounts(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_account":
		data, err := getAccount(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_update_account":
		data, err := updateAccount(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_list_clients":
		data, err := listClients(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_get_client":
		data, err := getClient(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_create_client":
		data, err := createClient(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_update_client":
		data, err := updateClient(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	case "slide_delete_client":
		data, err := deleteClient(args)
		if err != nil {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}
		} else {
			result = ToolResult{
				Content: []ToolContent{{Type: "text", Text: data}},
				IsError: false,
			}
		}

	default:
		result = ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", name)}},
			IsError: true,
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

func sendError(id interface{}, code int, message string, data interface{}) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
	}
}

func getAllTools() []ToolInfo {
	return []ToolInfo{
		{
			Name:        "slide_list_devices",
			Description: "List all devices with pagination and filtering options. Hostname is the primary identifier for devices.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by client ID",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (hostname)",
						"enum":        []string{"hostname"},
					},
				},
			},
		},
		{
			Name:        "slide_list_agents",
			Description: "List all agents with pagination and filtering options. Display Name is the primary identifier for agents.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of results per page (max 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Pagination offset",
					},
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by device ID",
					},
					"client_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by client ID",
					},
					"sort_asc": map[string]interface{}{
						"type":        "boolean",
						"description": "Sort in ascending order",
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Sort by field (id, hostname, name)",
						"enum":        []string{"id", "hostname", "name"},
					},
				},
			},
		},
		{
			Name:        "slide_get_agent",
			Description: "Get detailed information about a specific agent by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the agent to retrieve",
					},
				},
				"required": []string{"agent_id"},
			},
		},
	}
}
