package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// These profiles cover the protocol revisions used by current Anthropic and
// OpenAI local MCP hosts. Claude Desktop/MCPB compatibility still needs the
// 2025-06-18 path, while current Codex uses the official Rust MCP SDK and the
// 2025-11-25 tool/output-schema surface.
func TestMCPClientCompatibilityProfiles(t *testing.T) {
	profiles := []struct {
		name            string
		clientName      string
		protocolVersion string
	}{
		{name: "anthropic-claude-desktop", clientName: "claude-desktop", protocolVersion: "2025-06-18"},
		{name: "openai-codex", clientName: "codex-mcp-client", protocolVersion: "2025-11-25"},
	}

	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			setupTestEnv(t, ToolsSafe)
			srv, err := buildMCPServer()
			if err != nil {
				t.Fatalf("buildMCPServer: %v", err)
			}

			initialize := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "initialize",
				"params": map[string]interface{}{
					"protocolVersion": profile.protocolVersion,
					"capabilities":    map[string]interface{}{},
					"clientInfo": map[string]interface{}{
						"name":    profile.clientName,
						"version": "compat-harness",
					},
				},
			}
			initBody, _ := json.Marshal(initialize)
			initResponse := marshalRPCResponse(t, srv.HandleMessage(context.Background(), initBody))
			result := requireObject(t, initResponse["result"], "initialize.result")
			if got := result["protocolVersion"]; got != profile.protocolVersion {
				t.Fatalf("negotiated protocol = %v, want %s", got, profile.protocolVersion)
			}
			serverInfo := requireObject(t, result["serverInfo"], "initialize.serverInfo")
			if serverInfo["name"] != ServerName || serverInfo["version"] != Version {
				t.Fatalf("unexpected serverInfo: %v", serverInfo)
			}
			instructions, _ := result["instructions"].(string)
			if !strings.Contains(instructions, "USE THIS SERVER") || !strings.Contains(instructions, "name_hint") {
				t.Fatalf("initialize instructions missing routing contract")
			}

			initialized := []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
			if response := srv.HandleMessage(context.Background(), initialized); response != nil {
				t.Fatalf("initialized notification unexpectedly returned %v", response)
			}

			ping := marshalRPCResponse(t, srv.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":2,"method":"ping"}`)))
			if ping["error"] != nil {
				t.Fatalf("ping error: %v", ping["error"])
			}

			toolsResponse := marshalRPCResponse(t, srv.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`)))
			toolsResult := requireObject(t, toolsResponse["result"], "tools/list.result")
			tools, ok := toolsResult["tools"].([]interface{})
			if !ok || len(tools) != len(allToolInfos()) {
				t.Fatalf("tools/list returned %T len=%d, want %d tools", toolsResult["tools"], len(tools), len(allToolInfos()))
			}
			for _, rawTool := range tools {
				tool := requireObject(t, rawTool, "tool")
				schema := requireObject(t, tool["inputSchema"], "tool.inputSchema")
				if schema["type"] != "object" {
					t.Fatalf("tool %v has non-object input schema: %v", tool["name"], schema)
				}
				if _, ok := tool["annotations"].(map[string]interface{}); !ok {
					t.Fatalf("tool %v is missing MCP annotations", tool["name"])
				}
			}

			helpResponse := marshalRPCResponse(t, srv.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"slide_help","arguments":{"operation":"getting_started"}}}`)))
			helpResult := requireObject(t, helpResponse["result"], "tools/call.result")
			if helpResult["isError"] == true {
				t.Fatalf("slide_help failed: %v", helpResult)
			}
			content, ok := helpResult["content"].([]interface{})
			if !ok || len(content) == 0 {
				t.Fatalf("slide_help omitted content: %v", helpResult)
			}
		})
	}
}

func marshalRPCResponse(t *testing.T, response interface{}) map[string]interface{} {
	t.Helper()
	if response == nil {
		t.Fatal("expected JSON-RPC response, got nil")
	}
	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal JSON-RPC response: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("parse JSON-RPC response: %v; body=%s", err, body)
	}
	return parsed
}

func requireObject(t *testing.T, value interface{}, label string) map[string]interface{} {
	t.Helper()
	object, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("%s = %T, want object (%v)", label, value, value)
	}
	return object
}

// TestStdioTransportBlackBox launches the production binary and drives raw
// JSONL over stdio. It catches startup/logging regressions that in-process SDK
// tests cannot: stdout contamination, missed notifications, broken framing,
// and failure to shut down cleanly when the host closes stdin.
func TestStdioTransportBlackBox(t *testing.T) {
	if testing.Short() {
		t.Skip("black-box binary build skipped in short mode")
	}

	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "slide-mcp-server")
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build production binary: %v\n%s", err, output)
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"codex-mcp-client","version":"black-box"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"slide_help","arguments":{"operation":"what_can_you_do"}}}`,
	}, "\n") + "\n"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--skip-startup-validation")
	cmd.Dir = root
	cmd.Env = withoutEnvironmentKey(os.Environ(), "SLIDE_API_KEY")
	cmd.Env = append(cmd.Env, "SLIDE_API_KEY=tk_black_box_test")
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("stdio server failed: %v\nstderr:\n%s\nstdout:\n%s", err, stderr.String(), stdout.String())
	}
	if ctx.Err() != nil {
		t.Fatalf("stdio server did not exit after EOF: %v", ctx.Err())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("stdout contained %d lines, want 4 JSON-RPC responses:\n%s\nstderr:\n%s", len(lines), stdout.String(), stderr.String())
	}
	seenIDs := map[float64]bool{}
	for _, line := range lines {
		var message map[string]interface{}
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			t.Fatalf("stdout is not pure JSONL; invalid line %q: %v", line, err)
		}
		if message["jsonrpc"] != "2.0" || message["error"] != nil {
			t.Fatalf("bad JSON-RPC response: %v", message)
		}
		if id, ok := message["id"].(float64); ok {
			seenIDs[id] = true
		}
	}
	for _, id := range []float64{1, 2, 3, 4} {
		if !seenIDs[id] {
			t.Errorf("missing response id %.0f", id)
		}
	}
	if !strings.Contains(stderr.String(), "ready on stdio") {
		t.Fatalf("stderr did not contain readiness signal: %s", stderr.String())
	}
}

func withoutEnvironmentKey(env []string, key string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
