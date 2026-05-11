package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

// Global configuration instance. Plumbed through tools_*.go via package
// scope rather than DI because slide-mcp-server is a single-process,
// single-tenant binary; the indirection cost outweighs the cleanliness
// benefit and would touch every handler signature.
var config *ServerConfig

func main() {
	var (
		cliAPIKey        = flag.String("api-key", "", "API key for the Slide API (overrides SLIDE_API_KEY environment variable)")
		cliBaseURL       = flag.String("base-url", "", "Base URL for the Slide API (overrides SLIDE_BASE_URL environment variable)")
		cliTools         = flag.String("tools", "", "Tools mode: read-only, safe, full (overrides SLIDE_TOOLS environment variable). Legacy aliases reporting/restores/full-safe still work.")
		cliDisabledTools = flag.String("disabled-tools", "", "Comma-separated list of tool names to disable (overrides SLIDE_DISABLED_TOOLS environment variable)")
		showVersion      = flag.Bool("version", false, "Show version information and exit")

		// One-shot tool execution flags
		cliOneShotTool = flag.String("tool", "", "Run a single tool then exit (e.g. --tool slide_overview)")
		cliToolArgs    = flag.String("args", "", "JSON string with arguments for --tool (e.g. '{\"operation\":\"health\"}')")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s version %s\n", ServerName, Version)
		os.Exit(0)
	}

	config = NewServerConfig()

	if *cliTools != "" {
		config.ToolsMode = *cliTools
	} else if envTools := os.Getenv("SLIDE_TOOLS"); envTools != "" {
		config.ToolsMode = envTools
	}

	var disabledToolsStr string
	if *cliDisabledTools != "" {
		disabledToolsStr = *cliDisabledTools
	} else if envDisabledTools := os.Getenv("SLIDE_DISABLED_TOOLS"); envDisabledTools != "" {
		disabledToolsStr = envDisabledTools
	}
	config.SetDisabledTools(disabledToolsStr)

	if *cliBaseURL != "" {
		config.BaseURL = *cliBaseURL
	} else if envBaseURL := os.Getenv("SLIDE_BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}

	if *cliAPIKey != "" {
		config.APIKey = *cliAPIKey
	} else {
		config.APIKey = os.Getenv("SLIDE_API_KEY")
	}

	if err := config.ValidateToolsMode(); err != nil {
		log.Fatal(err)
	}

	if config.APIKey == "" {
		log.Fatalf("Error: API key not provided. Use --api-key flag or set SLIDE_API_KEY environment variable.")
	}

	APIBaseURL = config.BaseURL
	apiKey = config.APIKey

	if len(config.DisabledTools) > 0 {
		log.Printf("Disabled tools: %v", config.DisabledTools)
	}

	if *cliOneShotTool != "" {
		argsMap := map[string]interface{}{}
		if *cliToolArgs != "" {
			if err := json.Unmarshal([]byte(*cliToolArgs), &argsMap); err != nil {
				log.Fatalf("Invalid JSON for --args: %v", err)
			}
		}
		if err := runOneShotTool(*cliOneShotTool, argsMap); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("Slide MCP Server v%s starting (mode=%s)...", Version, config.ToolsMode)

	if err := runStdioServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
