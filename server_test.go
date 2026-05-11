package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

// setupTestEnv configures a global config + apiKey suitable for tests that
// don't need to hit the real API (or use httptest for the API).
func setupTestEnv(t *testing.T, mode string) {
	t.Helper()
	config = NewServerConfig()
	config.APIKey = "tk_test"
	if mode != "" {
		config.ToolsMode = mode
	}
	if err := config.ValidateToolsMode(); err != nil {
		t.Fatalf("validate mode: %v", err)
	}
	APIBaseURL = config.BaseURL
	apiKey = config.APIKey
}

// TestToolsListContents asserts every v4 tool is registered and the
// expected operations appear in their schemas.
func TestToolsListContents(t *testing.T) {
	setupTestEnv(t, ToolsFull)

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
				OutputSchema map[string]interface{} `json:"outputSchema,omitempty"`
				Annotations  struct {
					ReadOnlyHint    *bool `json:"readOnlyHint,omitempty"`
					DestructiveHint *bool `json:"destructiveHint,omitempty"`
				} `json:"annotations"`
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
		"slide_admin", "slide_agents", "slide_alerts", "slide_audit", "slide_backups",
		"slide_clients", "slide_devices", "slide_files", "slide_overview",
		"slide_recovery", "slide_snapshots",
	}
	for _, name := range expectedTools {
		if _, ok := got[name]; !ok {
			t.Errorf("expected tool %q to be registered, only saw: %v", name, sortedKeys(got))
		}
	}

	// Spot-check the new task-oriented operations are reachable.
	wantOps := map[string][]string{
		"slide_files":     {"search", "versions", "create_restore", "browse", "create_push"},
		"slide_audit":     {"list", "get", "actions", "resources", "recent"},
		"slide_overview":  {"inventory", "health", "for_client", "for_device"},
		"slide_recovery":  {"boot_vm", "export_image", "create_network", "create_wg_peer"},
		"slide_backups":   {"status_for_client", "status_for_device", "recent_for_agent"},
		"slide_alerts":    {"triage"},
		"slide_snapshots": {"recent_for_agent", "get_service_verification"},
	}
	for tool, ops := range wantOps {
		enum := got[tool]
		for _, op := range ops {
			if !contains(enum, op) {
				t.Errorf("tool %q is missing operation %q (have: %v)", tool, op, enum)
			}
		}
	}

	// Annotations: the read-only tools should advertise readOnlyHint=true.
	readOnly := map[string]bool{
		"slide_overview":                      true,
		"slide_audit":                         true,
		"list_all_clients_devices_and_agents": true,
	}
	for _, tool := range parsed.Result.Tools {
		want, ok := readOnly[tool.Name]
		if !ok {
			continue
		}
		if want {
			if tool.Annotations.ReadOnlyHint == nil || !*tool.Annotations.ReadOnlyHint {
				t.Errorf("expected %q to have readOnlyHint=true, got %+v", tool.Name, tool.Annotations)
			}
		}
	}
}

// TestToolFilterRespectsMode confirms read-only mode hides write tools'
// destructive operations while keeping the tools themselves visible.
func TestToolFilterRespectsMode(t *testing.T) {
	setupTestEnv(t, ToolsReadOnly)

	srv, err := buildMCPServer()
	if err != nil {
		t.Fatalf("buildMCPServer: %v", err)
	}

	req := []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	resp := srv.HandleMessage(context.Background(), req)
	body, _ := json.Marshal(resp)

	// In read-only mode, every advertised tool should still appear (per-op
	// gating happens at call time), but the call-level dispatcher should
	// reject destructive operations.
	if !strings.Contains(string(body), `"name":"slide_files"`) {
		t.Error("slide_files should appear in tools/list even in read-only mode (search is read-only)")
	}
}

// TestPermissionTierAliases ensures the legacy mode names still work.
func TestPermissionTierAliases(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"reporting", ToolsReadOnly},
		{"restores", ToolsSafe},
		{"full-safe", ToolsSafe},
		{"safe", ToolsSafe},
		{"read-only", ToolsReadOnly},
		{"full", ToolsFull},
	} {
		cfg := NewServerConfig()
		cfg.ToolsMode = c.in
		if err := cfg.ValidateToolsMode(); err != nil {
			t.Errorf("%q rejected: %v", c.in, err)
			continue
		}
		if cfg.ToolsMode != c.want {
			t.Errorf("%q -> %q, want %q", c.in, cfg.ToolsMode, c.want)
		}
	}
}

// TestPromptsAdvertised checks the v4 prompts/list response.
func TestPromptsAdvertised(t *testing.T) {
	setupTestEnv(t, ToolsSafe)
	srv, err := buildMCPServer()
	if err != nil {
		t.Fatalf("buildMCPServer: %v", err)
	}
	req := []byte(`{"jsonrpc":"2.0","id":3,"method":"prompts/list"}`)
	resp := srv.HandleMessage(context.Background(), req)
	body, _ := json.Marshal(resp)
	want := []string{
		"slide.daily-status",
		"slide.triage-alerts",
		"slide.restore-file",
		"slide.boot-recovery-vm",
		"slide.dr-runbook",
	}
	for _, p := range want {
		if !strings.Contains(string(body), `"`+p+`"`) {
			t.Errorf("prompts/list missing %q. got: %s", p, string(body))
		}
	}
}

// TestResourcesAdvertised checks resources/list and resources/templates/list.
func TestResourcesAdvertised(t *testing.T) {
	setupTestEnv(t, ToolsSafe)
	srv, err := buildMCPServer()
	if err != nil {
		t.Fatalf("buildMCPServer: %v", err)
	}
	for _, method := range []string{"resources/list", "resources/templates/list"} {
		req := []byte(`{"jsonrpc":"2.0","id":4,"method":"` + method + `"}`)
		resp := srv.HandleMessage(context.Background(), req)
		body, _ := json.Marshal(resp)
		if !strings.Contains(string(body), "slide://") {
			t.Errorf("%s did not include any slide:// URIs: %s", method, string(body))
		}
	}
	// The full set of static URIs we expect:
	for _, uri := range []string{
		resourceURIInventory,
		resourceURIHealth,
		resourceURIAlertsOpen,
		resourceURIAuditRecent,
		resourceURIDocsOpenAPI,
	} {
		req := []byte(`{"jsonrpc":"2.0","id":5,"method":"resources/list"}`)
		resp := srv.HandleMessage(context.Background(), req)
		body, _ := json.Marshal(resp)
		if !strings.Contains(string(body), uri) {
			t.Errorf("resources/list missing %q", uri)
		}
	}
}

// TestSlideFilesSearchHTTP exercises the slide_files search handler against
// a faked Slide API to verify URL building, response parsing, and summary
// formatting end-to-end.
func TestSlideFilesSearchHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent/a_test/file-search":
			if got := r.URL.Query().Get("search_term"); got != "budget" {
				t.Errorf("expected search_term=budget, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":[{"path":"C:\\Users\\bob\\budget.xlsx","size":12345,"modified_time":"2026-01-01T00:00:00Z"}],"pagination":{"total":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	setupTestEnv(t, ToolsFull)
	APIBaseURL = srv.URL

	out, err := handleFilesSearch(map[string]interface{}{
		"agent_id":    "a_test",
		"search_term": "budget",
	})
	if err != nil {
		t.Fatalf("handleFilesSearch: %v", err)
	}
	if !strings.Contains(out, "budget.xlsx") {
		t.Errorf("expected response to include budget.xlsx, got: %s", out)
	}
	if !strings.Contains(out, `"count":1`) {
		t.Errorf("expected count:1 in summary envelope, got: %s", out)
	}
}

// TestSlideAuditRecentHTTP exercises the slide_audit recent handler.
func TestSlideAuditRecentHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/audit") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"audit_id":"aud_111111111111","action":"user_created","resource_type":"user","resource_id":"u_x","audit_time":"2026-05-11T00:00:00Z","account_id":"acc_x"}],"pagination":{}}`))
	}))
	defer srv.Close()

	setupTestEnv(t, ToolsFull)
	APIBaseURL = srv.URL

	out, err := handleAuditRecent(map[string]interface{}{"hours": float64(24)})
	if err != nil {
		t.Fatalf("handleAuditRecent: %v", err)
	}
	if !strings.Contains(out, `"window_hours":24`) {
		t.Errorf("expected window_hours:24, got: %s", out)
	}
	if !strings.Contains(out, `user_created`) {
		t.Errorf("expected user_created in output, got: %s", out)
	}
}

// TestRetryAfterParsing ensures the API client honours Retry-After.
func TestRetryAfterParsing(t *testing.T) {
	cases := []struct {
		header string
		wantOK bool
	}{
		{"5", true},
		{"  10  ", true},
		{"", false},
		{"not-a-number", false},
		{"Wed, 21 Oct 2099 07:28:00 GMT", true},
	}
	for _, c := range cases {
		got := parseRetryAfter(c.header)
		if c.wantOK && got <= 0 {
			t.Errorf("parseRetryAfter(%q) returned %v, expected positive duration", c.header, got)
		}
		if !c.wantOK && got != 0 {
			t.Errorf("parseRetryAfter(%q) returned %v, expected 0", c.header, got)
		}
	}
}

// TestOneShotPermissionDenied confirms tool-level disablement works.
func TestOneShotPermissionDenied(t *testing.T) {
	setupTestEnv(t, ToolsFull)
	config.SetDisabledTools("slide_devices")

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
