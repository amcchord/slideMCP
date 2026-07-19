package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestToolRegistryContract(t *testing.T) {
	setupTestEnv(t, ToolsFull)
	infos := allToolInfos()
	seen := make(map[string]struct{}, len(infos))
	validName := regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

	for _, info := range infos {
		if !validName.MatchString(info.Name) {
			t.Errorf("tool name %q is not portable across MCP hosts", info.Name)
		}
		if _, duplicate := seen[info.Name]; duplicate {
			t.Errorf("duplicate tool descriptor %q", info.Name)
		}
		seen[info.Name] = struct{}{}
		if _, ok := toolRegistry[info.Name]; !ok {
			t.Errorf("tool %q has a descriptor but no handler", info.Name)
		}
		if strings.TrimSpace(info.Description) == "" {
			t.Errorf("tool %q has no description", info.Name)
		}
		schema, ok := info.InputSchema.(map[string]interface{})
		if !ok || schema["type"] != "object" {
			t.Errorf("tool %q input schema is not an object: %#v", info.Name, info.InputSchema)
			continue
		}
		if _, err := json.Marshal(schema); err != nil {
			t.Errorf("tool %q input schema is not JSON serializable: %v", info.Name, err)
		}
		if raw := outputSchemaForTool(info.Name); len(raw) > 0 {
			var output map[string]interface{}
			if err := json.Unmarshal(raw, &output); err != nil || output["type"] != "object" {
				t.Errorf("tool %q has invalid object output schema: %s (%v)", info.Name, raw, err)
			}
		}
		annotations := annotationsForTool(info.Name)
		if annotations.Title == "" || annotations.ReadOnlyHint == nil {
			t.Errorf("tool %q has incomplete annotations: %+v", info.Name, annotations)
		}
	}
	for name := range toolRegistry {
		if _, ok := seen[name]; !ok {
			t.Errorf("tool %q has a handler but no descriptor", name)
		}
	}
	if len(seen) != len(toolRegistry) {
		t.Fatalf("descriptor count=%d handler count=%d", len(seen), len(toolRegistry))
	}
}

func TestOperationPermissionContract(t *testing.T) {
	setupTestEnv(t, ToolsFull)
	for _, info := range allToolInfos() {
		operations := operationsFromInfo(t, info)
		for _, operation := range operations {
			if isReadOperation(info.Name, operation) && isDestructiveOperation(info.Name, operation) {
				t.Errorf("%s/%s is classified as both read-only and destructive", info.Name, operation)
			}
			if strings.HasPrefix(operation, "delete") || operation == "poweroff" || operation == "reboot" {
				if !isDestructiveOperation(info.Name, operation) {
					t.Errorf("%s/%s looks destructive but is not full-tier-only", info.Name, operation)
				}
			}
		}
	}

	config.ToolsMode = ToolsReadOnly
	for _, info := range allToolInfos() {
		for _, operation := range operationsFromInfo(t, info) {
			if isReadOperation(info.Name, operation) {
				continue
			}
			_, err := toolRegistry[info.Name](map[string]interface{}{"operation": operation})
			if err == nil || !strings.Contains(err.Error(), "not available") {
				t.Errorf("read-only mode did not reject %s/%s before dispatch: %v", info.Name, operation, err)
			}
		}
	}

	config.ToolsMode = ToolsSafe
	for _, info := range allToolInfos() {
		for _, operation := range operationsFromInfo(t, info) {
			if !isDestructiveOperation(info.Name, operation) {
				continue
			}
			_, err := toolRegistry[info.Name](map[string]interface{}{"operation": operation})
			if err == nil || !strings.Contains(err.Error(), "not available") {
				t.Errorf("safe mode did not reject destructive %s/%s before dispatch: %v", info.Name, operation, err)
			}
		}
	}
}

func TestEveryOperationToolRequiresOperation(t *testing.T) {
	setupTestEnv(t, ToolsFull)
	srv, err := buildMCPServer()
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range allToolInfos() {
		if len(operationsFromInfo(t, info)) == 0 {
			continue
		}
		req := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":%q,"arguments":{}}}`, info.Name))
		response := marshalRPCResponse(t, srv.HandleMessage(context.Background(), req))
		if response["error"] != nil {
			continue // Server-side JSON Schema validation rejected it.
		}
		result := requireObject(t, response["result"], info.Name+" result")
		if result["isError"] != true {
			t.Errorf("%s accepted a call without operation: %v", info.Name, response)
		}
	}
}

func TestOperationEnumsAreStableAndUnique(t *testing.T) {
	for _, info := range allToolInfos() {
		operations := operationsFromInfo(t, info)
		if len(operations) == 0 {
			continue
		}
		copyOfOps := append([]string(nil), operations...)
		sort.Strings(copyOfOps)
		for i := 1; i < len(copyOfOps); i++ {
			if copyOfOps[i] == copyOfOps[i-1] {
				t.Errorf("%s operation enum contains duplicate %q", info.Name, copyOfOps[i])
			}
		}
	}
}

func operationsFromInfo(t *testing.T, info ToolInfo) []string {
	t.Helper()
	schema, ok := info.InputSchema.(map[string]interface{})
	if !ok {
		return nil
	}
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil
	}
	operationProperty, ok := properties["operation"].(map[string]interface{})
	if !ok {
		return nil
	}
	switch values := operationProperty["enum"].(type) {
	case []string:
		return values
	case []interface{}:
		out := make([]string, 0, len(values))
		for _, value := range values {
			text, ok := value.(string)
			if !ok {
				t.Fatalf("%s has non-string operation enum value %v", info.Name, value)
			}
			out = append(out, text)
		}
		return out
	default:
		return nil
	}
}
