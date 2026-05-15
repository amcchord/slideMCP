package main

import (
	"fmt"
	"strings"
)

// Server constants. The version drives the manifest, the User-Agent header,
// and the `--version` flag.
const (
	ServerName = "slide-mcp-server"
	Version    = "5.0.1"
)

// Permission tiers. v4.0.0 collapses the old four-tier system
// (reporting/restores/full-safe/full) down to three, with backward-compat
// aliases so existing CLI flags / env vars / .mcpb user configs keep working.
const (
	ToolsReadOnly = "read-only" // pure read; lists, gets, searches, browses
	ToolsSafe     = "safe"      // default; read + restores + settings; no destructive ops
	ToolsFull     = "full"      // everything, including delete/poweroff/reboot
)

// legacy mode aliases (silently mapped to the new names)
var legacyToolsModes = map[string]string{
	"reporting": ToolsReadOnly,
	"restores":  ToolsSafe,
	"full-safe": ToolsSafe,
	"read":      ToolsReadOnly,
	"readonly":  ToolsReadOnly,
}

// ServerConfig holds all configuration for the MCP server.
type ServerConfig struct {
	APIKey        string
	BaseURL       string
	ToolsMode     string
	DisabledTools []string
}

// NewServerConfig creates a new configuration with defaults.
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		BaseURL:       "https://api.slide.tech",
		ToolsMode:     ToolsSafe,
		DisabledTools: []string{},
	}
}

// SetDisabledTools parses a comma-separated string of tool names.
func (c *ServerConfig) SetDisabledTools(disabledToolsStr string) {
	c.DisabledTools = []string{}
	if disabledToolsStr != "" {
		for _, tool := range strings.Split(disabledToolsStr, ",") {
			tool = strings.TrimSpace(tool)
			if tool != "" {
				c.DisabledTools = append(c.DisabledTools, tool)
			}
		}
	}
}

// ValidateToolsMode validates that the tools mode is valid, mapping legacy
// names through the alias table.
func (c *ServerConfig) ValidateToolsMode() error {
	if c.ToolsMode == "" {
		c.ToolsMode = ToolsSafe
		return nil
	}
	if mapped, ok := legacyToolsModes[c.ToolsMode]; ok {
		c.ToolsMode = mapped
		return nil
	}
	switch c.ToolsMode {
	case ToolsReadOnly, ToolsSafe, ToolsFull:
		return nil
	}
	return fmt.Errorf("invalid tools mode '%s'. Valid options: read-only, safe, full (legacy aliases reporting/restores/full-safe also accepted)", c.ToolsMode)
}

// IsToolDisabled checks if a tool is explicitly disabled.
func (c *ServerConfig) IsToolDisabled(toolName string) bool {
	for _, disabled := range c.DisabledTools {
		if disabled == toolName {
			return true
		}
	}
	return false
}

// IsToolAllowed checks if a tool is allowed at the tool level. Per-operation
// gating happens inside HandleToolWithOperations via IsOperationAllowed.
//
// slide_help is special-cased to be allowed in every mode and never
// disable-able: it's the LLM's escape hatch when a user is stuck.
func (c *ServerConfig) IsToolAllowed(toolName string) bool {
	if toolName == "slide_help" {
		return true
	}
	if c.IsToolDisabled(toolName) {
		return false
	}
	switch c.ToolsMode {
	case ToolsReadOnly:
		return isReadOnlyTool(toolName)
	case ToolsSafe, ToolsFull:
		return true
	}
	return false
}

// IsOperationAllowed checks if a specific operation on a tool is allowed in
// the active mode. Read operations are always permitted; safe-mode blocks
// destructive ops; full unlocks everything.
func (c *ServerConfig) IsOperationAllowed(toolName, operation string) bool {
	switch c.ToolsMode {
	case ToolsReadOnly:
		return isReadOperation(toolName, operation)
	case ToolsSafe:
		return !isDestructiveOperation(toolName, operation)
	case ToolsFull:
		return true
	}
	return false
}

// isReadOnlyTool returns true for tools whose entire surface is reads.
func isReadOnlyTool(toolName string) bool {
	switch toolName {
	case "slide_help",
		"slide_overview", "slide_audit", "list_all_clients_devices_and_agents":
		return true
	}
	// Mixed-purpose tools are still allowed in read-only mode at the tool
	// level; per-op gating filters writes via isReadOperation.
	switch toolName {
	case "slide_files", "slide_recovery", "slide_clients", "slide_admin",
		"slide_devices", "slide_agents", "slide_snapshots", "slide_backups", "slide_alerts":
		return true
	}
	return false
}

// isReadOperation returns true if (tool, op) is a read.
//
//nolint:gocyclo
func isReadOperation(toolName, op string) bool {
	// slide_help is read-only end-to-end: every operation serves baked-in
	// content or non-mutating capability metadata.
	if toolName == "slide_help" {
		return true
	}
	switch op {
	case "list", "get", "browse", "search", "versions",
		"recent", "actions", "resources",
		"list_users", "get_user", "get_user_avatar",
		"list_accounts", "get_account",
		"list_clients", "get_client",
		"list_services", "list_vlans", "get_vlan",
		"get_network", "list_networks",
		"get_service_verification",
		"recent_for_agent", "status_for_client", "status_for_device",
		"inventory", "health", "for_client", "for_device",
		"list_restores", "get_restore", "list_pushes", "get_push_status",
		"list_vms", "get_vm", "get_rdp_bookmark",
		"list_images", "get_image", "browse_image",
		"list_deleted",
		"triage",
		// slide_help operations
		"getting_started", "examples", "glossary", "troubleshoot",
		"list_prompts", "list_resources", "what_can_you_do":
		return true
	}
	return false
}

// isDestructiveOperation returns true for ops that delete, power-cycle, or
// otherwise irreversibly mutate. These are blocked outside `full` mode.
func isDestructiveOperation(toolName, op string) bool {
	switch toolName {
	case "slide_devices":
		switch op {
		case "poweroff", "reboot", "delete_vlan":
			return true
		}
	case "slide_agents":
		if op == "delete" {
			return true
		}
	case "slide_clients":
		if op == "delete" {
			return true
		}
	case "slide_recovery":
		switch op {
		case "delete_vm", "delete_image", "delete_network",
			"delete_ipsec", "delete_port_forward", "delete_wg_peer":
			return true
		}
	case "slide_files":
		// File restore deletion is destructive.
		if op == "delete_restore" {
			return true
		}
	case "slide_snapshots":
		if op == "delete" {
			return true
		}
	}
	return false
}
