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
	var (
		cliAPIKey             = flag.String("api-key", "", "API key for Slide service (overrides SLIDE_API_KEY environment variable)")
		cliBaseURL            = flag.String("base-url", "", "Base URL for Slide API (overrides SLIDE_BASE_URL environment variable)")
		cliTools              = flag.String("tools", "", "Tools mode: reporting, restores, full-safe, full (overrides SLIDE_TOOLS environment variable)")
		cliDisabledTools      = flag.String("disabled-tools", "", "Comma-separated list of tool names to disable (overrides SLIDE_DISABLED_TOOLS environment variable)")
		cliEnablePresentation = flag.Bool("enable-presentation", false, "Enable the slide_presentation tool (overrides SLIDE_ENABLE_PRESENTATION environment variable)")
		cliEnableReports      = flag.Bool("enable-reports", false, "Enable the slide_reports tool (overrides SLIDE_ENABLE_REPORTS environment variable)")
		showVersion           = flag.Bool("version", false, "Show version information and exit")
		// One-shot tool execution flags
		cliOneShotTool = flag.String("tool", "", "Run a single tool then exit (e.g. --tool slide_reports)")
		cliToolArgs    = flag.String("args", "", "JSON string with arguments for --tool (e.g. '{\"operation\":\"daily_backup_snapshot\"}')")
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

	if *cliEnablePresentation {
		config.EnablePresentation = true
	} else if envEnablePresentation := os.Getenv("SLIDE_ENABLE_PRESENTATION"); envEnablePresentation != "" {
		config.EnablePresentation = envEnablePresentation == "true" || envEnablePresentation == "1"
	}

	if *cliEnableReports {
		config.EnableReports = true
	} else if envEnableReports := os.Getenv("SLIDE_ENABLE_REPORTS"); envEnableReports != "" {
		config.EnableReports = envEnableReports == "true" || envEnableReports == "1"
	}

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
	if config.EnablePresentation {
		log.Printf("Presentation tool enabled")
	}
	if config.EnableReports {
		log.Printf("Reports tool enabled")
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

	log.Println("Slide MCP Server starting...")

	if err := runStdioServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
