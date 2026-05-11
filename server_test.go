package main

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

// TestToolsListContents builds the same MCP server we serve over stdio and
// asserts every meta-tool we expect to ship in v3.0.0 is present, including
// every operation added in Slide API v1.27.0. No network calls.
func TestToolsListContents(t *testing.T) {
	// Bootstrap the global config the same way main() would, but with a
	// dummy API key so we never have to touch the real Slide API. We keep
	// presentation/reports enabled to cover the full tool surface.
	config = NewServerConfig()
	config.APIKey = "tk_test"
	config.ToolsMode = ToolsFull
	config.EnablePresentation = true
	config.EnableReports = true
	APIBaseURL = config.BaseURL
	apiKey = config.APIKey

	srv, err := buildMCPServer()
	if err != nil {
		t.Fatalf("buildMCPServer: %v", err)
	}

	req := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	resp := srv.HandleMessage(context.Background(), req)
	if resp == nil {
		t.Fatal("HandleMessage returned nil response")
	}

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var parsed struct {
		Result struct {
			Tools []struct {
				Name        string `json:"name"`
				InputSchema struct {
					Properties struct {
						Operation struct {
							Enum []string `json:"enum"`
						} `json:"operation"`
					} `json:"properties"`
				} `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal response: %v\nbody: %s", err, string(body))
	}
	if parsed.Error != nil {
		t.Fatalf("tools/list returned error: %v", parsed.Error)
	}

	got := map[string][]string{}
	for _, tool := range parsed.Result.Tools {
		got[tool.Name] = tool.InputSchema.Properties.Operation.Enum
	}

	expectedTools := []string{
		"list_all_clients_devices_and_agents",
		"slide_agents", "slide_alerts", "slide_backups", "slide_devices",
		"slide_docs", "slide_meta", "slide_networks", "slide_presentation",
		"slide_reports", "slide_restores", "slide_snapshots",
		"slide_user_management", "slide_vms",
	}
	for _, name := range expectedTools {
		if _, ok := got[name]; !ok {
			t.Errorf("expected tool %q to be registered, only saw: %v", name, sortedKeys(got))
		}
	}

	// Every Slide API v1.27.0 operation should be reachable through the
	// existing meta-tools. If a name is missing from the operation enum the
	// SDK validator would reject the call entirely, so we want a hard fail.
	wantOps := map[string][]string{
		"slide_agents": {
			"list_services", "update_services",
			"set_schedule", "clear_schedule",
			"pause_backups", "resume_backups",
			"set_retention", "set_restore_defaults",
			"set_volumes", "set_file_index_enabled",
			"set_timezone", "set_comments",
			"update_alert_config",
		},
		"slide_devices": {
			"get_network", "update_network",
			"list_vlans", "get_vlan", "create_vlan", "update_vlan", "delete_vlan",
		},
		"slide_snapshots": {"get_service_verification"},
		"slide_user_management": {"get_user_avatar"},
	}
	for tool, ops := range wantOps {
		enum := got[tool]
		for _, op := range ops {
			if !contains(enum, op) {
				t.Errorf("tool %q is missing operation %q (have: %v)", tool, op, enum)
			}
		}
	}
}

// TestToolFilterRespectsMode confirms the mode-aware filter hides operations
// that should not be visible to the LLM in the active permission tier.
func TestToolFilterRespectsMode(t *testing.T) {
	config = NewServerConfig()
	config.APIKey = "tk_test"
	config.ToolsMode = ToolsReporting
	APIBaseURL = config.BaseURL
	apiKey = config.APIKey

	srv, err := buildMCPServer()
	if err != nil {
		t.Fatalf("buildMCPServer: %v", err)
	}

	req := []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	resp := srv.HandleMessage(context.Background(), req)
	body, _ := json.Marshal(resp)

	// Reporting mode should never advertise the presentation or reports
	// tools (gated by enable flags) and should not advertise any tool that
	// is not read-only.
	if strings.Contains(string(body), `"name":"slide_presentation"`) {
		t.Errorf("slide_presentation should be hidden when EnablePresentation=false")
	}
	if strings.Contains(string(body), `"name":"slide_reports"`) {
		t.Errorf("slide_reports should be hidden when EnableReports=false")
	}
}

// TestOneShotPermissionDenied confirms tool-level disablement is enforced by
// the one-shot CLI dispatcher (operation-level denial flows through the
// regular tool result with isError=true; tool-level denial is a hard error).
func TestOneShotPermissionDenied(t *testing.T) {
	config = NewServerConfig()
	config.APIKey = "tk_test"
	config.ToolsMode = ToolsFull
	config.SetDisabledTools("slide_devices")
	APIBaseURL = config.BaseURL
	apiKey = config.APIKey

	if err := runOneShotTool("slide_devices", map[string]interface{}{
		"operation": "list",
	}); err == nil {
		t.Fatal("expected error when calling a disabled tool, got nil")
	}

	if err := runOneShotTool("slide_does_not_exist", map[string]interface{}{}); err == nil {
		t.Fatal("expected error when calling an unknown tool, got nil")
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
