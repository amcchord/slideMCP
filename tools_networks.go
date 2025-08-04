package main

import (
	"fmt"
)

// handleNetworksTool handles all network-related operations through a single meta-tool
func handleNetworksTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !config.IsOperationAllowed("slide_networks", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_networks in '%s' mode", operation, config.ToolsMode)
	}

	switch operation {
	// Basic network operations
	case "list":
		return listNetworks(args)
	case "get":
		return getNetwork(args)
	case "create":
		return createNetwork(args)
	case "update":
		return updateNetwork(args)
	case "delete":
		return deleteNetwork(args)
	// IPSec connection operations
	case "create_ipsec":
		return createNetworkIPSecConn(args)
	case "update_ipsec":
		return updateNetworkIPSecConn(args)
	case "delete_ipsec":
		return deleteNetworkIPSecConn(args)
	// Port forward operations
	case "create_port_forward":
		return createNetworkPortForward(args)
	case "update_port_forward":
		return updateNetworkPortForward(args)
	case "delete_port_forward":
		return deleteNetworkPortForward(args)
	// WireGuard peer operations
	case "create_wg_peer":
		return createNetworkWGPeer(args)
	case "update_wg_peer":
		return updateNetworkWGPeer(args)
	case "delete_wg_peer":
		return deleteNetworkWGPeer(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getNetworksToolInfo returns the tool definition for the networks meta-tool
func getNetworksToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_networks",
		Description: "Manage networks and their configurations including IPSec connections, port forwards, and WireGuard peers. Networks enable virtual machines to communicate with each other and external networks.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list", "get", "create", "update", "delete", "create_ipsec", "update_ipsec", "delete_ipsec", "create_port_forward", "update_port_forward", "delete_port_forward", "create_wg_peer", "update_wg_peer", "delete_wg_peer"},
				},
				// Common parameters for list operations
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with 'list' operation",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with 'list' operation",
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with 'list' operation",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with 'list' operation",
					"enum":        []string{"id"},
				},
				// Basic network parameters
				"network_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the network - required for get, update, delete operations and all sub-resource operations",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the network - required for 'create' operation",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Network type - required for 'create' operation",
					"enum":        []string{"standard", "bridge-lan"},
				},
				"bridge_device_id": map[string]interface{}{
					"type":        "string",
					"description": "Device ID for bridge networks - used with 'create' and 'update' operations",
				},
				"client_id": map[string]interface{}{
					"type":        "string",
					"description": "Client ID for the network - should match the client_id of VMs that will use this network",
				},
				"comments": map[string]interface{}{
					"type":        "string",
					"description": "Comments about the network - used with 'create' and 'update' operations",
				},
				"dhcp": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable DHCP server - used with 'create' and 'update' operations",
				},
				"dhcp_range_start": map[string]interface{}{
					"type":        "string",
					"description": "DHCP range start address - used with 'create' and 'update' operations",
				},
				"dhcp_range_end": map[string]interface{}{
					"type":        "string",
					"description": "DHCP range end address - used with 'create' and 'update' operations",
				},
				"internet": map[string]interface{}{
					"type":        "boolean",
					"description": "Allow internet access - used with 'create' and 'update' operations",
				},
				"nameservers": map[string]interface{}{
					"type":        "array",
					"description": "DNS servers - used with 'create' and 'update' operations",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"router_prefix": map[string]interface{}{
					"type":        "string",
					"description": "Router IP address with prefix (e.g., '192.168.1.1/24') - must be a usable IP address within the network, not the network address itself",
				},
				"wg": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable WireGuard VPN - used with 'create' and 'update' operations",
				},
				"wg_prefix": map[string]interface{}{
					"type":        "string",
					"description": "WireGuard network prefix - must not overlap with other networks",
				},
				// IPSec connection parameters
				"ipsec_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the IPSec connection - required for 'update_ipsec' and 'delete_ipsec' operations",
				},
				"remote_addrs": map[string]interface{}{
					"type":        "array",
					"description": "Remote addresses for IPSec - required for 'create_ipsec' operation",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"remote_networks": map[string]interface{}{
					"type":        "array",
					"description": "Remote networks for IPSec - required for 'create_ipsec' operation, or for WireGuard peers",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				// Port forward parameters
				"port_forward_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the port forward - required for 'update_port_forward' and 'delete_port_forward' operations",
				},
				"proto": map[string]interface{}{
					"type":        "string",
					"description": "Protocol for port forward - required for 'create_port_forward' operation",
					"enum":        []string{"tcp", "udp"},
				},
				"dest": map[string]interface{}{
					"type":        "string",
					"description": "Destination address:port for port forward - required for 'create_port_forward' operation",
				},
				// WireGuard peer parameters
				"wg_peer_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the WireGuard peer - required for 'update_wg_peer' and 'delete_wg_peer' operations",
				},
				"peer_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the WireGuard peer - required for 'create_wg_peer' operation",
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"name", "type"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create_ipsec"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "name", "remote_addrs", "remote_networks"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update_ipsec"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "ipsec_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete_ipsec"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "ipsec_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create_port_forward"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "proto", "dest"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update_port_forward"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "port_forward_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete_port_forward"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "port_forward_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create_wg_peer"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "peer_name"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "update_wg_peer"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "wg_peer_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete_wg_peer"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"network_id", "wg_peer_id"},
					},
				},
			},
		},
	}
}
