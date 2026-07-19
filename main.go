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
		runDoctorFlag    = flag.Bool("doctor", false, "Run self-diagnostic checks (token, network, sample reads) and exit. Idempotent and CI-friendly.")
		runDebugFlag     = flag.Bool("debug", false, "Dump a full diagnostic bundle (version, runtime, config, env, DNS, TLS, live API probes, recent logs) as JSON and exit. Safe to paste into a support thread; API token is masked.")
		skipValidation   = flag.Bool("skip-startup-validation", false, "Skip the startup probe of /v1/account. Useful when launching offline.")

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

	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}

	if config.APIKey == "" && !*runDoctorFlag && !*runDebugFlag {
		log.Fatalf(`Error: Slide API token not provided.

Use one of:
  - Pass --api-key <token>
  - Set the SLIDE_API_KEY environment variable
  - Configure the token in Claude Desktop's Slide Backup extension settings

Generate a token at https://console.slide.tech under My Settings -> API Tokens.

To diagnose without a token: slide-mcp-server --debug`)
	}

	APIBaseURL = config.BaseURL
	apiKey = config.APIKey

	if len(config.DisabledTools) > 0 {
		log.Printf("Disabled tools: %v", config.DisabledTools)
	}

	if *runDoctorFlag {
		runDoctor()
		return
	}

	if *runDebugFlag {
		runDebug()
		return
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

	// Validate the API token in the background so we never block the MCP
	// initialize handshake on a network round-trip to Slide. The goroutine
	// only logs to stderr - it can NEVER kill the process. If the token
	// is bad, each tool call surfaces the same friendly auth error via
	// APIError, and the warning lands in Claude Desktop's extension log.
	if !*skipValidation {
		go runStartupValidation()
	}

	if err := runStdioServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
