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

type Network struct {
	NetworkID        string               `json:"network_id"`
	Type             string               `json:"type"`
	Name             string               `json:"name"`
	Comments         string               `json:"comments"`
	BridgeDeviceID   string               `json:"bridge_device_id"`
	RouterPrefix     string               `json:"router_prefix"`
	DHCP             bool                 `json:"dhcp"`
	DHCPRangeStart   string               `json:"dhcp_range_start"`
	DHCPRangeEnd     string               `json:"dhcp_range_end"`
	Nameservers      []string             `json:"nameservers"`
	Internet         bool                 `json:"internet"`
	ConnectedVirtIDs []string             `json:"connected_virt_ids"`
	IPSecConns       []NetworkIPSecConn   `json:"ipsec_conns"`
	PortForwards     []NetworkPortForward `json:"port_forwards"`
	WGPeers          []NetworkWGPeer      `json:"wg_peers"`
	WG               bool                 `json:"wg"`
	WGPrefix         string               `json:"wg_prefix"`
	WGPublicKey      string               `json:"wg_public_key"`
	ClientID         *string              `json:"client_id,omitempty"`
}

type NetworkIPSecConn struct {
	IPSecID        string   `json:"ipsec_id"`
	Name           string   `json:"name"`
	PSK            string   `json:"psk"`
	LocalID        string   `json:"local_id"`
	LocalAddrs     []string `json:"local_addrs"`
	LocalNetworks  []string `json:"local_networks"`
	RemoteID       string   `json:"remote_id"`
	RemoteAddrs    []string `json:"remote_addrs"`
	RemoteNetworks []string `json:"remote_networks"`
}

type NetworkPortForward struct {
	PortForwardID string `json:"port_forward_id"`
	Proto         string `json:"proto"`
	Port          int    `json:"port"`
	Dest          string `json:"dest"`
	Endpoint      string `json:"endpoint"`
}

type NetworkWGPeer struct {
	WGPeerID       string   `json:"wg_peer_id"`
	PeerName       string   `json:"peer_name"`
	WGPublicKey    string   `json:"wg_public_key"`
	WGPrivateKey   string   `json:"wg_private_key"`
	WGAddress      string   `json:"wg_address"`
	WGEndpoint     string   `json:"wg_endpoint"`
	RemoteNetworks []string `json:"remote_networks"`
}

// Helper function to generate VNC viewer URL
func generateVNCViewerURL(virtID, websocketURI, vncPassword string) string {
	encodedWebsocketURI := url.QueryEscape(websocketURI)
	base64Password := base64.StdEncoding.EncodeToString([]byte(vncPassword))
	return fmt.Sprintf("https://slide.recipes/mcpTools/vncViewer.php?id=%s&ws=%s&password=%s&encoding=base64",
		virtID, encodedWebsocketURI, base64Password)
}

// Helper function to generate WireGuard config file
func generateWireGuardConfig(peer NetworkWGPeer, serverPublicKey, networkPrefix string) string {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = 25
`,
		peer.WGPrivateKey,
		peer.WGAddress,
		serverPublicKey,
		peer.WGEndpoint,
		joinNetworks(peer.RemoteNetworks, networkPrefix))

	return config
}

// Helper function to join networks with commas for WireGuard config
func joinNetworks(networks []string, defaultNetwork string) string {
	if len(networks) == 0 {
		if defaultNetwork != "" {
			return defaultNetwork // Default to the network scope if no specific networks
		}
		return "0.0.0.0/0, ::/0" // Fallback to all traffic if no network prefix available
	}
	result := ""
	for i, network := range networks {
		if i > 0 {
			result += ", "
		}
		result += network
	}
	return result
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

func getDevice(args map[string]interface{}) (string, error) {
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	endpoint := fmt.Sprintf("/v1/device/%s", deviceID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Device
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateDevice(args map[string]interface{}) (string, error) {
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	payload := make(map[string]interface{})

	if displayName, ok := args["display_name"]; ok {
		payload["display_name"] = displayName
	}
	if hostname, ok := args["hostname"]; ok {
		payload["hostname"] = hostname
	}
	if clientID, ok := args["client_id"]; ok {
		payload["client_id"] = clientID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/device/%s", deviceID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result Device
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func powerOffDevice(args map[string]interface{}) (string, error) {
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	endpoint := fmt.Sprintf("/v1/device/%s/shutdown/poweroff", deviceID)
	data, err := makeAPIRequest("POST", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Device
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func rebootDevice(args map[string]interface{}) (string, error) {
	deviceID, ok := args["device_id"].(string)
	if !ok {
		return "", fmt.Errorf("device_id is required")
	}

	endpoint := fmt.Sprintf("/v1/device/%s/shutdown/reboot", deviceID)
	data, err := makeAPIRequest("POST", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Device
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
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
			"primary_identifier":              "snapshot_id",
			"presentation_guidance":           "Snapshots are point-in-time backups that can be used for restores. Check locations to see where stored.",
			"workflow_guidance":               "Snapshots can be restored as files, images, or virtual machines. Verify status shows boot/filesystem verification results.",
			"location_guidance":               "Each snapshot has a locations array showing where it's stored. If a location's device_id matches the agent's device_id, it's stored locally on that device. If the device_id is different, it's likely stored in the cloud. This is important for choosing where to deploy virtual machines.",
			"virtualization_device_selection": "When creating virtual machines from this snapshot, you can choose any device_id from the locations array. If you use the agent's original device_id, the VM will run locally on that device. If you use a different device_id from locations (cloud device), the VM will run in the cloud. Always inform the user whether their VM will be local or cloud-based.",
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

	// Add metadata for better LLM interaction
	enhancedResult := map[string]interface{}{
		"snapshot_id":                result.SnapshotID,
		"agent_id":                   result.AgentID,
		"locations":                  result.Locations,
		"backup_started_at":          result.BackupStartedAt,
		"backup_ended_at":            result.BackupEndedAt,
		"deleted":                    result.Deleted,
		"deletions":                  result.Deletions,
		"verify_boot_status":         result.VerifyBootStatus,
		"verify_fs_status":           result.VerifyFsStatus,
		"verify_boot_screenshot_url": result.VerifyBootScreenshotURL,
		"_metadata": map[string]interface{}{
			"primary_identifier":              "snapshot_id",
			"presentation_guidance":           "This snapshot can be restored as files, images, or virtual machines.",
			"location_guidance":               "The locations array shows where this snapshot is stored. If a location's device_id matches the agent_id's device, it's stored locally. If different, it's likely in the cloud.",
			"virtualization_device_selection": "When creating VMs from this snapshot, choose device_id from locations. Agent's original device = local VM, different device = cloud VM. Always tell the user if their VM will be local or cloud-based.",
			"restore_options":                 "This snapshot can be used for file restores, image exports, or virtual machine creation depending on your recovery needs.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
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
			"primary_identifier":     "virt_id",
			"presentation_guidance":  "Virtual machines created from snapshots for testing or disaster recovery.",
			"workflow_guidance":      "VMs can be started/stopped and accessed via VNC. Great for testing backups before full restore.",
			"vnc_guidance":           "The easiest way to access a virtual machine is through the _vnc_viewer_url property - this provides a direct browser link to the VM console that requires no additional software or configuration.",
			"console_access":         "Always use the _vnc_viewer_url for immediate browser-based console access. This is much easier than configuring a separate VNC client.",
			"network_configuration":  "When creating new VMs, always use network_type: 'network-nat-shared' for most use cases. This provides NAT networking with internet access.",
			"network_type_reference": "Valid network_type values: 'network-nat-shared' (recommended), 'network-nat-isolated', 'bridge', 'network-id'",
			"network_dependencies":   "IMPORTANT: If you need to create a VM with network_type: 'network-id', you MUST create the custom network first using slide_create_network before creating the VM. Built-in network types do not require pre-existing networks.",
			"network_clarification":  "IMPORTANT: If networks allready exist ask the user for clarification if they want the VM to be connected to the network. They don't specify at creation time okay to just go with network-nat-shared",
			"deployment_awareness":   "Each VM's device_id indicates where it's running. When referencing VMs to users, be aware that some may be running locally on their devices while others may be running in the cloud, depending on which device_id was used during creation.",
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
			"console_access":        "The easiest way to access this virtual machine is through the _vnc_viewer_url property - this provides a direct browser link to the VM console that requires no additional software or configuration.",
			"deployment_location":   "The device_id field indicates where this VM is running. If this matches the original agent's device_id, it's running locally. If different, it's likely running in the cloud.",
		},
	}

	for _, vnc := range result.VNC {
		if vnc.WebsocketURI != nil {
			vncViewerURL := generateVNCViewerURL(result.VirtID, *vnc.WebsocketURI, result.VNCPassword)
			enhancedResult["_vnc_viewer_url"] = vncViewerURL
			// Add to metadata as well
			enhancedResult["_metadata"].(map[string]interface{})["vnc_guidance"] = "The _vnc_viewer_url is the easiest way to access this VM console - simply click the link to open the VM in your browser. No VNC client setup required."
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
			"primary_identifier":       "virt_id",
			"presentation_guidance":    "When referring to the virtual machine, use the virt_id as the primary identifier. Virtual machine IDs are internal identifiers not commonly used by humans.",
			"next_steps":               "Now that you've created a virtual machine, you can control it using slide_update_virtual_machine to change its state (running, stopped, paused) or update resources.",
			"resource_guidance":        "For optimal performance, 8192MB of RAM is recommended for most VMs. You can adjust this as needed using slide_update_virtual_machine.",
			"console_access":           "The easiest way to access this virtual machine is through the _vnc_viewer_url property - this provides a direct browser link to the VM console that requires no additional software or configuration.",
			"network_configuration":    "When creating VMs, easiest to use network_type: 'network-nat-shared' for most use cases. This provides NAT networking with internet access.",
			"network_type_options":     "Valid network_type values: 'network-nat-shared' (recommended default), 'network-nat-isolated' (no internet), 'bridge' (direct LAN access), 'network-id' (connect to specific network)",
			"network_best_practices":   "Always specify network_type explicitly. Use 'network-nat-shared' unless you have specific requirements for isolation or LAN bridging. The network_model should typically be 'virtio' for best performance.",
			"network_dependencies":     "IMPORTANT: If using network_type: 'network-id', you MUST create the custom network first using slide_create_network before creating the VM. The network_source field should reference an existing network_id. Built-in network types ('network-nat-shared', 'network-nat-isolated', 'bridge') do not require pre-existing networks.",
			"network_clarification":    "IMPORTANT: If the user hasn't specified their networking requirements clearly, ask for clarification about whether they need internet access, LAN connectivity, or custom network isolation before choosing network_type. Don't assume their networking needs.",
			"client_id_matching":       "CRITICAL: When using network_type: 'network-id', the VM's client_id MUST match the network's client_id. A network with client_id '' (empty string) can only be used with VMs that also have client_id ''. A network with a specific client_id can only be used with VMs that have the same client_id. Always verify client ID compatibility before creating VMs with custom networks.",
			"deployment_location":      "IMPORTANT: Always inform the user whether this VM was deployed locally or in the cloud. Check if the device_id used matches the original agent's device (local) or is a different device from the snapshot locations (cloud). Users need to know where their VM is running.",
			"deployment_communication": "When presenting VM creation results to users, clearly state: 'Your virtual machine has been created and will run [locally on your device / in the cloud]' based on the device_id selection.",
		},
	}

	for _, vnc := range result.VNC {
		if vnc.WebsocketURI != nil {
			vncViewerURL := generateVNCViewerURL(result.VirtID, *vnc.WebsocketURI, result.VNCPassword)
			enhancedResult["_vnc_viewer_url"] = vncViewerURL
			// Add to metadata as well
			enhancedResult["_metadata"].(map[string]interface{})["vnc_guidance"] = "The _vnc_viewer_url is the easiest way to access this VM console - simply click the link to open the VM in your browser. No VNC client setup required."
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
			"console_access":        "The easiest way to access this virtual machine is through the _vnc_viewer_url property - this provides a direct browser link to the VM console that requires no additional software or configuration.",
			"deployment_location":   "The device_id field indicates where this VM is running. If this matches the original agent's device_id, it's running locally. If different, it's likely running in the cloud.",
		},
	}

	for _, vnc := range result.VNC {
		if vnc.WebsocketURI != nil {
			vncViewerURL := generateVNCViewerURL(result.VirtID, *vnc.WebsocketURI, result.VNCPassword)
			enhancedResult["_vnc_viewer_url"] = vncViewerURL
			// Add to metadata as well
			enhancedResult["_metadata"].(map[string]interface{})["vnc_guidance"] = "The _vnc_viewer_url is the easiest way to access this VM console - simply click the link to open the VM in your browser. No VNC client setup required."
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

// Network API functions
func listNetworks(args map[string]interface{}) (string, error) {
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

	endpoint := "/v1/network"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result PaginatedResponse[Network]
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Process networks to include WireGuard configs for peers
	enhancedNetworks := make([]map[string]interface{}, len(result.Data))
	for i, network := range result.Data {
		enhancedNetwork := map[string]interface{}{
			"network_id":         network.NetworkID,
			"type":               network.Type,
			"name":               network.Name,
			"comments":           network.Comments,
			"bridge_device_id":   network.BridgeDeviceID,
			"router_prefix":      network.RouterPrefix,
			"dhcp":               network.DHCP,
			"dhcp_range_start":   network.DHCPRangeStart,
			"dhcp_range_end":     network.DHCPRangeEnd,
			"nameservers":        network.Nameservers,
			"internet":           network.Internet,
			"connected_virt_ids": network.ConnectedVirtIDs,
			"ipsec_conns":        network.IPSecConns,
			"port_forwards":      network.PortForwards,
			"wg":                 network.WG,
			"wg_prefix":          network.WGPrefix,
			"wg_public_key":      network.WGPublicKey,
			"client_id":          network.ClientID,
		}

		// Process WireGuard peers to include config files
		if len(network.WGPeers) > 0 {
			enhancedWGPeers := make([]map[string]interface{}, len(network.WGPeers))

			for j, peer := range network.WGPeers {
				enhancedPeer := map[string]interface{}{
					"wg_peer_id":        peer.WGPeerID,
					"peer_name":         peer.PeerName,
					"wg_public_key":     peer.WGPublicKey,
					"wg_private_key":    peer.WGPrivateKey,
					"wg_address":        peer.WGAddress,
					"wg_endpoint":       peer.WGEndpoint,
					"remote_networks":   peer.RemoteNetworks,
					"_wireguard_config": generateWireGuardConfig(peer, network.WGPublicKey, network.WGPrefix),
				}
				enhancedWGPeers[j] = enhancedPeer
			}
			enhancedNetwork["wg_peers"] = enhancedWGPeers
		} else {
			enhancedNetwork["wg_peers"] = network.WGPeers
		}

		enhancedNetworks[i] = enhancedNetwork
	}

	enhancedResult := map[string]interface{}{
		"pagination": result.Pagination,
		"data":       enhancedNetworks,
		"_metadata": map[string]interface{}{
			"primary_identifier":     "name",
			"presentation_guidance":  "Networks enable disaster recovery and isolated networking for virtual machines.",
			"workflow_guidance":      "Networks can be standard (isolated) or bridge-lan (connected to device LAN). Virtual machines can be connected to networks.",
			"wireguard_guidance":     "Networks with WireGuard enabled will have WG peers that include ready-to-use configuration files in the _wireguard_config field.",
			"creation_guidance":      "When creating new networks with slide_create_network, ask users for clarification about configuration details rather than guessing. Key areas that often need clarification: network type (standard vs bridge-lan), bridge device selection, IP address ranges, DHCP configuration, internet access requirements, and WireGuard VPN needs.",
			"clarification_guidance": "IMPORTANT: Network configuration errors can cause serious connectivity issues. Always ask users to specify their requirements clearly before creating networks. Don't assume default values for critical settings like network type, IP addressing, or bridge device selection.",
			"client_id_matching":     "CRITICAL: When working with networks, agents and VMs assigned to the network MUST be part of the same client as the network. A network with client_id '' (empty string) can only be used with VMs that also have client_id ''. A network with a specific client_id can only be used with VMs that have the same client_id. Mismatched client IDs will cause network assignment failures.",
		},
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getNetwork(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}

	endpoint := fmt.Sprintf("/v1/network/%s", networkID)
	data, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	var result Network
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Enhanced result with WireGuard configs for peers
	enhancedResult := map[string]interface{}{
		"network_id":         result.NetworkID,
		"type":               result.Type,
		"name":               result.Name,
		"comments":           result.Comments,
		"bridge_device_id":   result.BridgeDeviceID,
		"router_prefix":      result.RouterPrefix,
		"dhcp":               result.DHCP,
		"dhcp_range_start":   result.DHCPRangeStart,
		"dhcp_range_end":     result.DHCPRangeEnd,
		"nameservers":        result.Nameservers,
		"internet":           result.Internet,
		"connected_virt_ids": result.ConnectedVirtIDs,
		"ipsec_conns":        result.IPSecConns,
		"port_forwards":      result.PortForwards,
		"wg":                 result.WG,
		"wg_prefix":          result.WGPrefix,
		"wg_public_key":      result.WGPublicKey,
		"client_id":          result.ClientID,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "network_id",
			"presentation_guidance": "Network configuration with associated services and peers.",
			"wireguard_guidance":    "If this network has WireGuard enabled, WG peers will include ready-to-use configuration files in the _wireguard_config field.",
		},
	}

	// Process WireGuard peers to include config files
	if len(result.WGPeers) > 0 {
		enhancedWGPeers := make([]map[string]interface{}, len(result.WGPeers))

		for i, peer := range result.WGPeers {
			enhancedPeer := map[string]interface{}{
				"wg_peer_id":        peer.WGPeerID,
				"peer_name":         peer.PeerName,
				"wg_public_key":     peer.WGPublicKey,
				"wg_private_key":    peer.WGPrivateKey,
				"wg_address":        peer.WGAddress,
				"wg_endpoint":       peer.WGEndpoint,
				"remote_networks":   peer.RemoteNetworks,
				"_wireguard_config": generateWireGuardConfig(peer, result.WGPublicKey, result.WGPrefix),
			}
			enhancedWGPeers[i] = enhancedPeer
		}
		enhancedResult["wg_peers"] = enhancedWGPeers
	} else {
		enhancedResult["wg_peers"] = result.WGPeers
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func createNetwork(args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("name is required")
	}
	networkType, ok := args["type"].(string)
	if !ok {
		return "", fmt.Errorf("type is required")
	}

	payload := map[string]interface{}{
		"name": name,
		"type": networkType,
	}

	// Add optional parameters
	if bridgeDeviceID, ok := args["bridge_device_id"]; ok {
		payload["bridge_device_id"] = bridgeDeviceID
	}
	if clientID, ok := args["client_id"]; ok {
		payload["client_id"] = clientID
	}
	if comments, ok := args["comments"]; ok {
		payload["comments"] = comments
	}
	if dhcp, ok := args["dhcp"]; ok {
		payload["dhcp"] = dhcp
	}
	if dhcpRangeStart, ok := args["dhcp_range_start"]; ok {
		payload["dhcp_range_start"] = dhcpRangeStart
	}
	if dhcpRangeEnd, ok := args["dhcp_range_end"]; ok {
		payload["dhcp_range_end"] = dhcpRangeEnd
	}
	if internet, ok := args["internet"]; ok {
		payload["internet"] = internet
	}
	if nameservers, ok := args["nameservers"]; ok {
		payload["nameservers"] = nameservers
	}
	if routerPrefix, ok := args["router_prefix"]; ok {
		if prefixStr, ok := routerPrefix.(string); ok {
			// Ensure router_prefix includes CIDR notation
			if prefixStr != "" && !strings.Contains(prefixStr, "/") {
				return "", fmt.Errorf("router_prefix is expecting CIDR notation for the routers IP address (e.g., '192.168.1.1/24')")
			}
			// Validate that router_prefix is not the network address
			if prefixStr != "" && strings.Contains(prefixStr, "/") {
				parts := strings.Split(prefixStr, "/")
				if len(parts) == 2 {
					ip := parts[0]
					// Check if the IP ends with .0, which would indicate a network address
					if strings.HasSuffix(ip, ".0") {
						return "", fmt.Errorf("router_prefix cannot be the network address (e.g., use '192.168.1.1/24' not '192.168.1.0/24')")
					}
				}
			}
		}
		payload["router_prefix"] = routerPrefix
	}
	if wg, ok := args["wg"]; ok {
		payload["wg"] = wg
	}
	if wgPrefix, ok := args["wg_prefix"]; ok {
		if prefixStr, ok := wgPrefix.(string); ok {
			// Ensure wg_prefix includes CIDR notation
			if prefixStr != "" && !strings.Contains(prefixStr, "/") {
				return "", fmt.Errorf("wg_prefix is expecting CIDR notation for the WireGuard IP address (e.g., '10.0.0.1/24')")
			}
		}
		payload["wg_prefix"] = wgPrefix
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := makeAPIRequest("POST", "/v1/network", body)
	if err != nil {
		return "", err
	}

	var result Network
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Enhanced result with WireGuard configs for peers
	enhancedResult := map[string]interface{}{
		"network_id":         result.NetworkID,
		"type":               result.Type,
		"name":               result.Name,
		"comments":           result.Comments,
		"bridge_device_id":   result.BridgeDeviceID,
		"router_prefix":      result.RouterPrefix,
		"dhcp":               result.DHCP,
		"dhcp_range_start":   result.DHCPRangeStart,
		"dhcp_range_end":     result.DHCPRangeEnd,
		"nameservers":        result.Nameservers,
		"internet":           result.Internet,
		"connected_virt_ids": result.ConnectedVirtIDs,
		"ipsec_conns":        result.IPSecConns,
		"port_forwards":      result.PortForwards,
		"wg":                 result.WG,
		"wg_prefix":          result.WGPrefix,
		"wg_public_key":      result.WGPublicKey,
		"client_id":          result.ClientID,
		"_metadata": map[string]interface{}{
			"primary_identifier":     "network_id",
			"presentation_guidance":  "Network created successfully. You can now connect virtual machines to this network or configure additional services.",
			"next_steps":             "You can now create virtual machines using this network_id with network_type: 'network-id' and network_source: '" + result.NetworkID + "'",
			"router_prefix_guidance": "The router_prefix is the IP address of the router that will be used to connect to the network. It should NOT be the same as the network address (the first IP in the subnet). For example, use '192.168.1.1/24' not '192.168.1.0/24'.",
			"wireguard_guidance":     "If WireGuard is enabled, you can create WG peers to allow VPN access to this network using slide_create_network_wg_peer.",
			"client_id_constraint":   "CRITICAL: This network can only be used with VMs and agents that have the same client_id. If this network has client_id '" + fmt.Sprintf("%v", result.ClientID) + "', then any VMs using network_type: 'network-id' with this network MUST also have the same client_id. Empty string client_id can only work with other empty string client_ids.",
			"clarification_guidance": "IMPORTANT: When creating networks, if you are unsure about any configuration details (network type, bridge device selection, DHCP settings, IP addressing, WireGuard configuration, etc.), it is always better to ask the user for clarification rather than guessing. Network configuration mistakes can cause connectivity issues that are difficult to troubleshoot.",
		},
	}

	// Process WireGuard peers to include config files (if any returned)
	if len(result.WGPeers) > 0 {
		enhancedWGPeers := make([]map[string]interface{}, len(result.WGPeers))

		for i, peer := range result.WGPeers {
			enhancedPeer := map[string]interface{}{
				"wg_peer_id":        peer.WGPeerID,
				"peer_name":         peer.PeerName,
				"wg_public_key":     peer.WGPublicKey,
				"wg_private_key":    peer.WGPrivateKey,
				"wg_address":        peer.WGAddress,
				"wg_endpoint":       peer.WGEndpoint,
				"remote_networks":   peer.RemoteNetworks,
				"_wireguard_config": generateWireGuardConfig(peer, result.WGPublicKey, result.WGPrefix),
			}
			enhancedWGPeers[i] = enhancedPeer
		}
		enhancedResult["wg_peers"] = enhancedWGPeers
	} else {
		enhancedResult["wg_peers"] = result.WGPeers
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateNetwork(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}

	payload := make(map[string]interface{})

	// Add optional parameters
	if bridgeDeviceID, ok := args["bridge_device_id"]; ok {
		payload["bridge_device_id"] = bridgeDeviceID
	}
	if clientID, ok := args["client_id"]; ok {
		payload["client_id"] = clientID
	}
	if comments, ok := args["comments"]; ok {
		payload["comments"] = comments
	}
	if dhcp, ok := args["dhcp"]; ok {
		payload["dhcp"] = dhcp
	}
	if dhcpRangeStart, ok := args["dhcp_range_start"]; ok {
		payload["dhcp_range_start"] = dhcpRangeStart
	}
	if dhcpRangeEnd, ok := args["dhcp_range_end"]; ok {
		payload["dhcp_range_end"] = dhcpRangeEnd
	}
	if internet, ok := args["internet"]; ok {
		payload["internet"] = internet
	}
	if name, ok := args["name"]; ok {
		payload["name"] = name
	}
	if nameservers, ok := args["nameservers"]; ok {
		payload["nameservers"] = nameservers
	}
	if routerPrefix, ok := args["router_prefix"]; ok {
		if prefixStr, ok := routerPrefix.(string); ok {
			// Ensure router_prefix includes CIDR notation
			if prefixStr != "" && !strings.Contains(prefixStr, "/") {
				return "", fmt.Errorf("router_prefix must include subnet mask (e.g., '192.168.1.1/24')")
			}
			// Validate that router_prefix is not the network address
			if prefixStr != "" && strings.Contains(prefixStr, "/") {
				parts := strings.Split(prefixStr, "/")
				if len(parts) == 2 {
					ip := parts[0]
					// Check if the IP ends with .0, which would indicate a network address
					if strings.HasSuffix(ip, ".0") {
						return "", fmt.Errorf("router_prefix cannot be the network address (e.g., use '192.168.1.1/24' not '192.168.1.0/24')")
					}
				}
			}
		}
		payload["router_prefix"] = routerPrefix
	}
	if wg, ok := args["wg"]; ok {
		payload["wg"] = wg
	}
	if wgPrefix, ok := args["wg_prefix"]; ok {
		if prefixStr, ok := wgPrefix.(string); ok {
			// Ensure wg_prefix includes CIDR notation
			if prefixStr != "" && !strings.Contains(prefixStr, "/") {
				return "", fmt.Errorf("wg_prefix must include subnet mask (e.g., '10.0.0.1/24')")
			}
		}
		payload["wg_prefix"] = wgPrefix
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/network/%s", networkID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result Network
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Enhanced result with WireGuard configs for peers
	enhancedResult := map[string]interface{}{
		"network_id":         result.NetworkID,
		"type":               result.Type,
		"name":               result.Name,
		"comments":           result.Comments,
		"bridge_device_id":   result.BridgeDeviceID,
		"router_prefix":      result.RouterPrefix,
		"dhcp":               result.DHCP,
		"dhcp_range_start":   result.DHCPRangeStart,
		"dhcp_range_end":     result.DHCPRangeEnd,
		"nameservers":        result.Nameservers,
		"internet":           result.Internet,
		"connected_virt_ids": result.ConnectedVirtIDs,
		"ipsec_conns":        result.IPSecConns,
		"port_forwards":      result.PortForwards,
		"wg":                 result.WG,
		"wg_prefix":          result.WGPrefix,
		"wg_public_key":      result.WGPublicKey,
		"client_id":          result.ClientID,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "network_id",
			"presentation_guidance": "Network updated successfully.",
			"wireguard_guidance":    "If WireGuard is enabled, existing WG peers will include ready-to-use configuration files in the _wireguard_config field.",
		},
	}

	// Process WireGuard peers to include config files (if any returned)
	if len(result.WGPeers) > 0 {
		enhancedWGPeers := make([]map[string]interface{}, len(result.WGPeers))

		for i, peer := range result.WGPeers {
			enhancedPeer := map[string]interface{}{
				"wg_peer_id":        peer.WGPeerID,
				"peer_name":         peer.PeerName,
				"wg_public_key":     peer.WGPublicKey,
				"wg_private_key":    peer.WGPrivateKey,
				"wg_address":        peer.WGAddress,
				"wg_endpoint":       peer.WGEndpoint,
				"remote_networks":   peer.RemoteNetworks,
				"_wireguard_config": generateWireGuardConfig(peer, result.WGPublicKey, result.WGPrefix),
			}
			enhancedWGPeers[i] = enhancedPeer
		}
		enhancedResult["wg_peers"] = enhancedWGPeers
	} else {
		enhancedResult["wg_peers"] = result.WGPeers
	}

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteNetwork(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}

	endpoint := fmt.Sprintf("/v1/network/%s", networkID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "Network deleted successfully", nil
}

// Network IPsec Connection functions
func createNetworkIPSecConn(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("name is required")
	}
	remoteAddrs, ok := args["remote_addrs"].([]interface{})
	if !ok {
		return "", fmt.Errorf("remote_addrs is required")
	}
	remoteNetworks, ok := args["remote_networks"].([]interface{})
	if !ok {
		return "", fmt.Errorf("remote_networks is required")
	}

	// Convert []interface{} to []string
	remoteAddrStrings := make([]string, len(remoteAddrs))
	for i, addr := range remoteAddrs {
		if addrStr, ok := addr.(string); ok {
			remoteAddrStrings[i] = addrStr
		} else {
			return "", fmt.Errorf("remote_addrs must be an array of strings")
		}
	}

	remoteNetworkStrings := make([]string, len(remoteNetworks))
	for i, network := range remoteNetworks {
		if networkStr, ok := network.(string); ok {
			remoteNetworkStrings[i] = networkStr
		} else {
			return "", fmt.Errorf("remote_networks must be an array of strings")
		}
	}

	payload := map[string]interface{}{
		"name":            name,
		"remote_addrs":    remoteAddrStrings,
		"remote_networks": remoteNetworkStrings,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/network/%s/ipsec", networkID)
	data, err := makeAPIRequest("POST", endpoint, body)
	if err != nil {
		return "", err
	}

	var result NetworkIPSecConn
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateNetworkIPSecConn(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	ipsecID, ok := args["ipsec_id"].(string)
	if !ok {
		return "", fmt.Errorf("ipsec_id is required")
	}

	payload := make(map[string]interface{})

	if name, ok := args["name"]; ok {
		payload["name"] = name
	}
	if remoteAddrs, ok := args["remote_addrs"]; ok {
		if addrs, ok := remoteAddrs.([]interface{}); ok {
			// Convert []interface{} to []string
			remoteAddrStrings := make([]string, len(addrs))
			for i, addr := range addrs {
				if addrStr, ok := addr.(string); ok {
					remoteAddrStrings[i] = addrStr
				} else {
					return "", fmt.Errorf("remote_addrs must be an array of strings")
				}
			}
			payload["remote_addrs"] = remoteAddrStrings
		}
	}
	if remoteNetworks, ok := args["remote_networks"]; ok {
		if networks, ok := remoteNetworks.([]interface{}); ok {
			// Convert []interface{} to []string
			remoteNetworkStrings := make([]string, len(networks))
			for i, network := range networks {
				if networkStr, ok := network.(string); ok {
					remoteNetworkStrings[i] = networkStr
				} else {
					return "", fmt.Errorf("remote_networks must be an array of strings")
				}
			}
			payload["remote_networks"] = remoteNetworkStrings
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/network/%s/ipsec/%s", networkID, ipsecID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result NetworkIPSecConn
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteNetworkIPSecConn(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	ipsecID, ok := args["ipsec_id"].(string)
	if !ok {
		return "", fmt.Errorf("ipsec_id is required")
	}

	endpoint := fmt.Sprintf("/v1/network/%s/ipsec/%s", networkID, ipsecID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "Network IPsec connection deleted successfully", nil
}

// Network Port Forward functions
func createNetworkPortForward(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	proto, ok := args["proto"].(string)
	if !ok {
		return "", fmt.Errorf("proto is required")
	}
	dest, ok := args["dest"].(string)
	if !ok {
		return "", fmt.Errorf("dest is required")
	}

	payload := map[string]interface{}{
		"proto": proto,
		"dest":  dest,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/network/%s/port-forward", networkID)
	data, err := makeAPIRequest("POST", endpoint, body)
	if err != nil {
		return "", err
	}

	var result NetworkPortForward
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateNetworkPortForward(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	portForwardID, ok := args["port_forward_id"].(string)
	if !ok {
		return "", fmt.Errorf("port_forward_id is required")
	}

	payload := make(map[string]interface{})

	if proto, ok := args["proto"]; ok {
		payload["proto"] = proto
	}
	if dest, ok := args["dest"]; ok {
		payload["dest"] = dest
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/network/%s/port-forward/%s", networkID, portForwardID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result NetworkPortForward
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteNetworkPortForward(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	portForwardID, ok := args["port_forward_id"].(string)
	if !ok {
		return "", fmt.Errorf("port_forward_id is required")
	}

	endpoint := fmt.Sprintf("/v1/network/%s/port-forward/%s", networkID, portForwardID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "Network port forward deleted successfully", nil
}

// Network WireGuard Peer functions
func createNetworkWGPeer(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	peerName, ok := args["peer_name"].(string)
	if !ok {
		return "", fmt.Errorf("peer_name is required")
	}

	payload := map[string]interface{}{
		"peer_name": peerName,
	}

	if remoteNetworks, ok := args["remote_networks"]; ok {
		if networks, ok := remoteNetworks.([]interface{}); ok {
			// Convert []interface{} to []string
			remoteNetworkStrings := make([]string, len(networks))
			for i, network := range networks {
				if networkStr, ok := network.(string); ok {
					remoteNetworkStrings[i] = networkStr
				} else {
					return "", fmt.Errorf("remote_networks must be an array of strings")
				}
			}
			payload["remote_networks"] = remoteNetworkStrings
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/network/%s/wg-peer", networkID)
	data, err := makeAPIRequest("POST", endpoint, body)
	if err != nil {
		return "", err
	}

	var result NetworkWGPeer
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Enhanced result with WireGuard config
	enhancedResult := map[string]interface{}{
		"wg_peer_id":      result.WGPeerID,
		"peer_name":       result.PeerName,
		"wg_public_key":   result.WGPublicKey,
		"wg_private_key":  result.WGPrivateKey,
		"wg_address":      result.WGAddress,
		"wg_endpoint":     result.WGEndpoint,
		"remote_networks": result.RemoteNetworks,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "wg_peer_id",
			"presentation_guidance": "WireGuard peer configuration for VPN access to the network.",
			"config_file_guidance":  "Use the _wireguard_config field to get a ready-to-use WireGuard configuration file. Save this as a .conf file and import it into your WireGuard client.",
			"usage_instructions":    "1. Save the _wireguard_config as a .conf file, 2. Import into WireGuard client, 3. Connect to access the specified remote networks.",
		},
	}

	// Get network details to retrieve server public key
	networkEndpoint := fmt.Sprintf("/v1/network/%s", networkID)
	networkData, err := makeAPIRequest("GET", networkEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get network details: %w", err)
	}

	var network Network
	if err := json.Unmarshal(networkData, &network); err != nil {
		return "", fmt.Errorf("failed to parse network response: %w", err)
	}

	// Generate WireGuard config with actual server public key
	wireguardConfig := generateWireGuardConfig(result, network.WGPublicKey, network.WGPrefix)
	enhancedResult["_wireguard_config"] = wireguardConfig

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func updateNetworkWGPeer(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	wgPeerID, ok := args["wg_peer_id"].(string)
	if !ok {
		return "", fmt.Errorf("wg_peer_id is required")
	}

	payload := make(map[string]interface{})

	if peerName, ok := args["peer_name"]; ok {
		payload["peer_name"] = peerName
	}
	if remoteNetworks, ok := args["remote_networks"]; ok {
		if networks, ok := remoteNetworks.([]interface{}); ok {
			// Convert []interface{} to []string
			remoteNetworkStrings := make([]string, len(networks))
			for i, network := range networks {
				if networkStr, ok := network.(string); ok {
					remoteNetworkStrings[i] = networkStr
				} else {
					return "", fmt.Errorf("remote_networks must be an array of strings")
				}
			}
			payload["remote_networks"] = remoteNetworkStrings
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("/v1/network/%s/wg-peer/%s", networkID, wgPeerID)
	data, err := makeAPIRequest("PATCH", endpoint, body)
	if err != nil {
		return "", err
	}

	var result NetworkWGPeer
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Enhanced result with WireGuard config
	enhancedResult := map[string]interface{}{
		"wg_peer_id":      result.WGPeerID,
		"peer_name":       result.PeerName,
		"wg_public_key":   result.WGPublicKey,
		"wg_private_key":  result.WGPrivateKey,
		"wg_address":      result.WGAddress,
		"wg_endpoint":     result.WGEndpoint,
		"remote_networks": result.RemoteNetworks,
		"_metadata": map[string]interface{}{
			"primary_identifier":    "wg_peer_id",
			"presentation_guidance": "Updated WireGuard peer configuration for VPN access to the network.",
			"config_file_guidance":  "Use the _wireguard_config field to get a ready-to-use WireGuard configuration file. Save this as a .conf file and import it into your WireGuard client.",
			"usage_instructions":    "1. Save the _wireguard_config as a .conf file, 2. Import into WireGuard client, 3. Connect to access the specified remote networks.",
		},
	}

	// Get network details to retrieve server public key
	networkEndpoint := fmt.Sprintf("/v1/network/%s", networkID)
	networkData, err := makeAPIRequest("GET", networkEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get network details: %w", err)
	}

	var network Network
	if err := json.Unmarshal(networkData, &network); err != nil {
		return "", fmt.Errorf("failed to parse network response: %w", err)
	}

	// Generate WireGuard config with actual server public key
	wireguardConfig := generateWireGuardConfig(result, network.WGPublicKey, network.WGPrefix)
	enhancedResult["_wireguard_config"] = wireguardConfig

	jsonData, err := json.MarshalIndent(enhancedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func deleteNetworkWGPeer(args map[string]interface{}) (string, error) {
	networkID, ok := args["network_id"].(string)
	if !ok {
		return "", fmt.Errorf("network_id is required")
	}
	wgPeerID, ok := args["wg_peer_id"].(string)
	if !ok {
		return "", fmt.Errorf("wg_peer_id is required")
	}

	endpoint := fmt.Sprintf("/v1/network/%s/wg-peer/%s", networkID, wgPeerID)
	_, err := makeAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return "", err
	}

	return "Network WireGuard peer deleted successfully", nil
}
