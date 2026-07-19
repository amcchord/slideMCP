package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type bundleManifest struct {
	ManifestVersion  string `json:"manifest_version"`
	Name             string `json:"name"`
	Version          string `json:"version"`
	Support          string `json:"support"`
	ToolsGenerated   bool   `json:"tools_generated"`
	PromptsGenerated bool   `json:"prompts_generated"`
	Tools            []struct {
		Name string `json:"name"`
	} `json:"tools"`
	Server struct {
		Type       string `json:"type"`
		EntryPoint string `json:"entry_point"`
		MCPConfig  struct {
			Command           string `json:"command"`
			PlatformOverrides struct {
				Linux struct {
					Command string   `json:"command"`
					Args    []string `json:"args"`
				} `json:"linux"`
			} `json:"platform_overrides"`
		} `json:"mcp_config"`
	} `json:"server"`
}

func TestDistributionManifestMatchesServer(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("dxt", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest bundleManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}

	if manifest.ManifestVersion != "0.3" {
		t.Errorf("manifest spec=%q, want current stable MCPB 0.3", manifest.ManifestVersion)
	}
	if manifest.Name != ServerName || manifest.Version != Version {
		t.Errorf("manifest identity=%s@%s, server=%s@%s", manifest.Name, manifest.Version, ServerName, Version)
	}
	if !strings.Contains(manifest.Support, "github.com/amcchord/slideMCP") {
		t.Errorf("manifest support points at the wrong repository: %q", manifest.Support)
	}
	if manifest.Server.Type != "binary" || manifest.Server.EntryPoint == "" || manifest.Server.MCPConfig.Command == "" {
		t.Errorf("manifest has incomplete binary server config: %+v", manifest.Server)
	}
	if manifest.ToolsGenerated {
		t.Error("fixed tool registry must use tools_generated=false")
	}
	if !manifest.PromptsGenerated {
		t.Error("runtime MCP prompts must use prompts_generated=true")
	}
	if manifest.Server.MCPConfig.PlatformOverrides.Linux.Command != "/bin/sh" {
		t.Error("Linux MCPB must use the architecture-selecting launcher")
	}
	linuxArgs := strings.Join(manifest.Server.MCPConfig.PlatformOverrides.Linux.Args, " ")
	if !strings.Contains(linuxArgs, "slide-mcp-server-linux") {
		t.Errorf("Linux launcher missing from platform override: %q", linuxArgs)
	}

	manifestTools := make([]string, 0, len(manifest.Tools))
	for _, tool := range manifest.Tools {
		manifestTools = append(manifestTools, tool.Name)
	}
	serverTools := make([]string, 0, len(allToolInfos()))
	for _, tool := range allToolInfos() {
		serverTools = append(serverTools, tool.Name)
	}
	sort.Strings(manifestTools)
	sort.Strings(serverTools)
	if strings.Join(manifestTools, "\n") != strings.Join(serverTools, "\n") {
		t.Errorf("manifest tools do not match runtime\nmanifest=%v\nruntime=%v", manifestTools, serverTools)
	}
}

func TestVersionDeclarationsStayInLockstep(t *testing.T) {
	makefile, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?m)^VERSION\s*=\s*(v[0-9]+\.[0-9]+\.[0-9]+)\s*$`)
	match := re.FindStringSubmatch(string(makefile))
	if len(match) != 2 {
		t.Fatal("Makefile VERSION declaration not found")
	}
	if match[1] != "v"+Version {
		t.Errorf("Makefile VERSION=%s, server Version=%s", match[1], Version)
	}

	launcher, err := os.Stat(filepath.Join("dxt", "launchers", "slide-mcp-server-linux"))
	if err != nil {
		t.Fatal(err)
	}
	if launcher.Mode()&0o111 == 0 {
		t.Error("Linux MCPB launcher is not executable")
	}
}
