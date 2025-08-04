package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

// Global configuration instance
var config *ServerConfig

func main() {
	// Parse command line flags
	var cliAPIKey = flag.String("api-key", "", "API key for Slide service (overrides SLIDE_API_KEY environment variable)")
	var cliBaseURL = flag.String("base-url", "", "Base URL for Slide API (overrides SLIDE_BASE_URL environment variable)")
	var cliTools = flag.String("tools", "", "Tools mode: reporting, restores, full-safe, full (overrides SLIDE_TOOLS environment variable)")
	var cliDisabledTools = flag.String("disabled-tools", "", "Comma-separated list of tool names to disable (overrides SLIDE_DISABLED_TOOLS environment variable)")
	var cliEnablePresentation = flag.Bool("enable-presentation", false, "Enable the slide_presentation tool (overrides SLIDE_ENABLE_PRESENTATION environment variable)")
	var cliEnableReports = flag.Bool("enable-reports", false, "Enable the slide_reports tool (overrides SLIDE_ENABLE_REPORTS environment variable)")
	var showVersion = flag.Bool("version", false, "Show version information and exit")
	var exitAfterFirst = flag.Bool("exit-after-first", false, "Exit after processing the first request instead of running continuously")
	// One-shot tool execution flags
	var cliOneShotTool = flag.String("tool", "", "Run a single tool then exit (e.g. --tool slide_reports)")
	var cliToolArgs = flag.String("args", "", "JSON string with arguments for --tool (e.g. '{\"operation\":\"daily_backup_snapshot\"}')")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("%s version %s\n", ServerName, Version)
		os.Exit(0)
	}

	// Initialize configuration
	config = NewServerConfig()

	// Configure tools mode from CLI flag or environment variable
	if *cliTools != "" {
		config.ToolsMode = *cliTools
	} else if envTools := os.Getenv("SLIDE_TOOLS"); envTools != "" {
		config.ToolsMode = envTools
	}

	// Configure disabled tools from CLI flag or environment variable
	var disabledToolsStr string
	if *cliDisabledTools != "" {
		disabledToolsStr = *cliDisabledTools
	} else if envDisabledTools := os.Getenv("SLIDE_DISABLED_TOOLS"); envDisabledTools != "" {
		disabledToolsStr = envDisabledTools
	}
	config.SetDisabledTools(disabledToolsStr)

	// Configure presentation tool enablement
	if *cliEnablePresentation {
		config.EnablePresentation = true
	} else if envEnablePresentation := os.Getenv("SLIDE_ENABLE_PRESENTATION"); envEnablePresentation != "" {
		config.EnablePresentation = envEnablePresentation == "true" || envEnablePresentation == "1"
	}

	// Configure reports tool enablement
	if *cliEnableReports {
		config.EnableReports = true
	} else if envEnableReports := os.Getenv("SLIDE_ENABLE_REPORTS"); envEnableReports != "" {
		config.EnableReports = envEnableReports == "true" || envEnableReports == "1"
	}

	// Configure base URL
	if *cliBaseURL != "" {
		config.BaseURL = *cliBaseURL
	} else if envBaseURL := os.Getenv("SLIDE_BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}

	// Configure API key
	if *cliAPIKey != "" {
		config.APIKey = *cliAPIKey
	} else {
		config.APIKey = os.Getenv("SLIDE_API_KEY")
	}

	// Validate configuration
	if err := config.ValidateToolsMode(); err != nil {
		log.Fatal(err)
	}

	if config.APIKey == "" {
		log.Fatalf("Error: API key not provided. Use --api-key flag or set SLIDE_API_KEY environment variable.")
	}

	// Set API base URL for compatibility with existing api.go
	APIBaseURL = config.BaseURL
	apiKey = config.APIKey

	// Log enabled tools
	if len(config.DisabledTools) > 0 {
		log.Printf("Disabled tools: %v", config.DisabledTools)
	}
	if config.EnablePresentation {
		log.Printf("Presentation tool enabled")
	}
	if config.EnableReports {
		log.Printf("Reports tool enabled")
	}

	// -------------------------------------------------
	// One-shot tool execution (if --tool flag provided)
	// -------------------------------------------------
	if *cliOneShotTool != "" {
		// Parse JSON args (if provided)
		var argsMap map[string]interface{}
		if *cliToolArgs != "" {
			if err := json.Unmarshal([]byte(*cliToolArgs), &argsMap); err != nil {
				log.Fatalf("Invalid JSON for --args: %v", err)
			}
		} else {
			argsMap = make(map[string]interface{})
		}

		// Construct a mock MCPRequest to reuse existing handler logic
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      *cliOneShotTool,
				"arguments": argsMap,
			},
		}

		resp := handleToolCall(req)

		// Marshal and print the response so callers can parse
		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			log.Fatalf("Failed to encode response: %v", err)
		}
		fmt.Println(string(out))
		return // Exit after one-shot execution
	}

	log.Println("Slide MCP Server starting...")

	// Start MCP server
	startMCPServer(*exitAfterFirst)
}
