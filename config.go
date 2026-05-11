package main

import (
	"fmt"
	"strings"
)

// Server configuration constants
const (
	ServerName = "slide-mcp-server"
	Version    = "3.1.0"
)

// Tools filtering modes
const (
	ToolsReporting = "reporting"
	ToolsRestores  = "restores"
	ToolsFullSafe  = "full-safe"
	ToolsFull      = "full"
)

// ServerConfig holds all configuration for the MCP server
type ServerConfig struct {
	// Core configuration
	APIKey  string
	BaseURL string

	// Tool access configuration
	ToolsMode          string
	DisabledTools      []string
	EnablePresentation bool
	EnableReports      bool
}

// NewServerConfig creates a new configuration with defaults
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		BaseURL:            "https://api.slide.tech",
		ToolsMode:          ToolsFullSafe,
		DisabledTools:      []string{},
		EnablePresentation: false,
		EnableReports:      false,
	}
}

// SetDisabledTools parses a comma-separated string of tool names
func (c *ServerConfig) SetDisabledTools(disabledToolsStr string) {
	c.DisabledTools = []string{}
	if disabledToolsStr != "" {
		toolsList := strings.Split(disabledToolsStr, ",")
		for _, tool := range toolsList {
			tool = strings.TrimSpace(tool)
			if tool != "" {
				c.DisabledTools = append(c.DisabledTools, tool)
			}
		}
	}
}

// ValidateToolsMode validates that the tools mode is valid
func (c *ServerConfig) ValidateToolsMode() error {
	switch c.ToolsMode {
	case ToolsReporting, ToolsRestores, ToolsFullSafe, ToolsFull:
		return nil
	default:
		return fmt.Errorf("invalid tools mode '%s'. Valid options: reporting, restores, full-safe, full", c.ToolsMode)
	}
}

// IsToolDisabled checks if a tool is explicitly disabled
func (c *ServerConfig) IsToolDisabled(toolName string) bool {
	for _, disabled := range c.DisabledTools {
		if disabled == toolName {
			return true
		}
	}
	return false
}

// IsToolAllowed checks if a tool is allowed based on current configuration
func (c *ServerConfig) IsToolAllowed(toolName string) bool {
	// First check if tool is explicitly disabled
	if c.IsToolDisabled(toolName) {
		return false
	}

	// Check if presentation or reports tools need explicit enablement
	if toolName == "slide_presentation" && !c.EnablePresentation {
		return false
	}
	if toolName == "slide_reports" && !c.EnableReports {
		return false
	}

	// Then check tools mode permissions
	switch c.ToolsMode {
	case ToolsReporting:
		// Only allow read-only tools
		return isReadOnlyTool(toolName)
	case ToolsRestores:
		// Allow reporting tools + restore/VM/network management
		return isReadOnlyTool(toolName) || isRestoreManagementTool(toolName)
	case ToolsFullSafe:
		// Allow everything except dangerous operations
		return !isDangerousTool(toolName)
	case ToolsFull:
		// Allow all tools
		return true
	default:
		return false
	}
}

// IsOperationAllowed checks if a specific operation on a tool is allowed
func (c *ServerConfig) IsOperationAllowed(toolName, operation string) bool {
	switch c.ToolsMode {
	case ToolsReporting:
		// Only allow read operations
		return isReadOperation(operation)
	case ToolsRestores:
		// Allow read operations + specific management operations
		return isReadOperation(operation) || isRestoreManagementOperation(toolName, operation)
	case ToolsFullSafe:
		// Allow everything except dangerous operations
		return !isDangerousOperation(toolName, operation)
	case ToolsFull:
		// Allow all operations
		return true
	default:
		return false
	}
}

// Helper functions for permission checking

func isReadOnlyTool(toolName string) bool {
	readOnlyTools := []string{
		"slide_agents", "slide_backups", "slide_snapshots", "slide_user_management",
		"slide_alerts", "slide_devices", "slide_networks",
		"slide_vms", "slide_restores", "slide_presentation", "slide_meta", "slide_docs", "list_all_clients_devices_and_agents",
	}
	for _, tool := range readOnlyTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

func isRestoreManagementTool(toolName string) bool {
	restoreTools := []string{
		"slide_vms", "slide_restores", "slide_networks",
	}
	for _, tool := range restoreTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

func isDangerousTool(toolName string) bool {
	// No tools are completely dangerous in full-safe mode
	// Danger is at the operation level
	return false
}

func isReadOperation(operation string) bool {
	readOps := []string{
		"list", "get", "browse", "list_deleted",
		// Restores tool read operations
		"list_files", "get_file", "browse_file", "list_images", "get_image", "browse_image",
		// User management tool read operations
		"list_users", "get_user", "get_user_avatar", "list_accounts", "get_account", "list_clients", "get_client",
		// Reports tool read operations
		"get_runbook_template",
		// Slide API v1.27.0 read operations
		"list_services", "get_network", "list_vlans", "get_vlan",
		"get_service_verification",
	}
	for _, op := range readOps {
		if op == operation {
			return true
		}
	}
	return false
}

func isRestoreManagementOperation(toolName, operation string) bool {
	switch toolName {
	case "slide_vms":
		return operation == "create" || operation == "update" || operation == "delete"
	case "slide_restores":
		return operation == "create_file" || operation == "delete_file" || operation == "browse_file" ||
			operation == "create_push" || operation == "update_push" || operation == "list_pushes" ||
			operation == "create_image" || operation == "delete_image" || operation == "browse_image"
	case "slide_networks":
		return operation == "create" || operation == "update" || operation == "delete" ||
			operation == "create_ipsec" || operation == "update_ipsec" || operation == "delete_ipsec" ||
			operation == "create_port_forward" || operation == "update_port_forward" || operation == "delete_port_forward" ||
			operation == "create_wg_peer" || operation == "update_wg_peer" || operation == "delete_wg_peer"
	case "slide_devices":
		// Removed device power control (poweroff, reboot) from restores mode.
		// Slide API v1.27.0: device network/VLAN edits are part of restore prep.
		return operation == "update" ||
			operation == "update_network" ||
			operation == "create_vlan" || operation == "update_vlan"
	case "slide_agents":
		// Slide API v1.27.0: schedule + retention + restore-defaults + service
		// verification + alert config edits are valuable while preparing
		// restores; passphrase/volumes/timezone changes stay full-safe-only.
		return operation == "create" || operation == "pair" || operation == "update" ||
			operation == "update_services" ||
			operation == "set_schedule" || operation == "clear_schedule" ||
			operation == "pause_backups" || operation == "resume_backups" ||
			operation == "set_retention" || operation == "set_restore_defaults" ||
			operation == "update_alert_config"
	case "slide_backups":
		return operation == "start"
	case "slide_user_management":
		return operation == "update_account" || operation == "create_client" || operation == "update_client" || operation == "delete_client"
		// Removed alert resolution from restores mode
	}
	return false
}

func isDangerousOperation(toolName, operation string) bool {
	// Define dangerous operations that are blocked in full-safe mode
	if toolName == "slide_agents" && operation == "delete" {
		return true
	}
	if toolName == "slide_snapshots" && operation == "delete" {
		return true
	}
	// Block device power control operations in full-safe mode
	if toolName == "slide_devices" && (operation == "poweroff" || operation == "reboot") {
		return true
	}
	// Slide API v1.27.0: deleting a VLAN can break VPN/port-forward routes that
	// reference its IPs; require explicit "full" mode for it.
	if toolName == "slide_devices" && operation == "delete_vlan" {
		return true
	}
	return false
}
