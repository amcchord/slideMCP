package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// handlePresentationTool handles all presentation and formatting operations through a single meta-tool
func handlePresentationTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !config.IsOperationAllowed("slide_presentation", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_presentation in '%s' mode", operation, config.ToolsMode)
	}

	switch operation {
	case "get_runbook_template":
		return getRunbookTemplate(args)
	case "get_daily_report_template":
		return getDailyReportTemplate(args)
	case "get_monthly_report_template":
		return getMonthlyReportTemplate(args)
	case "get_card":
		return getCard(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// downloadContext downloads context instructions from GitHub
func downloadContext(contextURL string) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the context
	resp, err := client.Get(contextURL)
	if err != nil {
		return "", fmt.Errorf("failed to download context: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download context: HTTP %d", resp.StatusCode)
	}

	// Read the content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read context content: %v", err)
	}

	return string(content), nil
}

// getRunbookTemplate downloads and returns runbook templates from GitHub
func getRunbookTemplate(args map[string]interface{}) (string, error) {
	format, ok := args["format"].(string)
	if !ok {
		format = "markdown" // Default to markdown
	}

	var templateURL string
	switch format {
	case "html":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/runbook/runbook.html"
	case "haml":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/runbook/runbook.haml"
	case "markdown", "md":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/runbook/runbook.md"
	default:
		return "", fmt.Errorf("unsupported format: %s. Supported formats: html, haml, markdown, md", format)
	}

	// Download context and template in parallel
	contextURL := "https://raw.githubusercontent.com/amcchord/slideReports/main/runbook/context.txt"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the context
	context, err := downloadContext(contextURL)
	if err != nil {
		return "", fmt.Errorf("failed to download runbook context: %v", err)
	}

	// Download the template
	resp, err := client.Get(templateURL)
	if err != nil {
		return "", fmt.Errorf("failed to download template: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download template: HTTP %d", resp.StatusCode)
	}

	// Read the template content
	templateContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read template content: %v", err)
	}

	// Combine context and template
	response := fmt.Sprintf(`%s

TEMPLATE CONTENT:
%s`, context, string(templateContent))

	return response, nil
}

// getDailyReportTemplate downloads and returns daily report templates from GitHub
func getDailyReportTemplate(args map[string]interface{}) (string, error) {
	format, ok := args["format"].(string)
	if !ok {
		format = "html" // Default to html for daily reports
	}

	var templateURL string
	switch format {
	case "html":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/daily_report/daily_report.html"
	case "haml":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/daily_report/daily_report.haml"
	case "markdown", "md":
		// Assuming markdown version exists at similar path
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/daily_report/daily_report.md"
	default:
		return "", fmt.Errorf("unsupported format: %s. Supported formats: html, haml, markdown, md", format)
	}

	// Download context and template
	contextURL := "https://raw.githubusercontent.com/amcchord/slideReports/main/daily_report/context.txt"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the context
	context, err := downloadContext(contextURL)
	if err != nil {
		return "", fmt.Errorf("failed to download daily report context: %v", err)
	}

	// Download the template
	resp, err := client.Get(templateURL)
	if err != nil {
		return "", fmt.Errorf("failed to download daily report template: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download daily report template: HTTP %d", resp.StatusCode)
	}

	// Read the template content
	templateContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read daily report template content: %v", err)
	}

	// Combine context and template
	response := fmt.Sprintf(`%s

TEMPLATE CONTENT:
%s`, context, string(templateContent))

	return response, nil
}

// getMonthlyReportTemplate downloads and returns monthly report templates from GitHub
func getMonthlyReportTemplate(args map[string]interface{}) (string, error) {
	format, ok := args["format"].(string)
	if !ok {
		format = "html" // Default to html for monthly reports
	}

	var templateURL string
	switch format {
	case "html":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/monthly_report/monthly_report.html"
	case "haml":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/monthly_report/monthly_report.haml"
	case "markdown", "md":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/monthly_report/monthly_report.md"
	default:
		return "", fmt.Errorf("unsupported format: %s. Supported formats: html, haml, markdown, md", format)
	}

	// Download context and template
	contextURL := "https://raw.githubusercontent.com/amcchord/slideReports/main/monthly_report/context.txt"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the context
	context, err := downloadContext(contextURL)
	if err != nil {
		return "", fmt.Errorf("failed to download monthly report context: %v", err)
	}

	// Download the template
	resp, err := client.Get(templateURL)
	if err != nil {
		return "", fmt.Errorf("failed to download monthly report template: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download monthly report template: HTTP %d", resp.StatusCode)
	}

	// Read the template content
	templateContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read monthly report template content: %v", err)
	}

	// Combine context and template
	response := fmt.Sprintf(`%s

TEMPLATE CONTENT:
%s`, context, string(templateContent))

	return response, nil
}

// getCard downloads and returns card templates from GitHub
func getCard(args map[string]interface{}) (string, error) {
	cardType, ok := args["card_type"].(string)
	if !ok {
		return "", fmt.Errorf("card_type parameter is required")
	}

	// Validate card type and get description
	cardDescriptions := map[string]string{
		"agent":           "Individual backup agent card - Shows detailed info for a single agent including hostname, OS, status, recent backups, and client assignment. Use for agent status pages or detailed agent views.",
		"agents_table":    "Multiple agents table card - Shows overview table of multiple agents with status, last seen, and assignments. Use for agent dashboard, status overview, or multi-agent comparison.",
		"client":          "Individual client card - Shows detailed info for a single client including name, agent count, device assignments, and stats. Use for client detail pages or client status views.",
		"clients_table":   "Multiple clients table card - Shows overview table of multiple clients with agent counts and assignments. Use for client dashboard or multi-client management views.",
		"device":          "Individual backup device card - Shows detailed info for a single backup device including capacity, assignments, and storage info. Use for device status pages or detailed device views.",
		"devices_table":   "Multiple devices table card - Shows overview table of multiple backup devices with capacity and assignments. Use for device dashboard or storage management views.",
		"snapshot":        "Individual snapshot card - Shows detailed info for a single backup snapshot including date, size, status, and retention. Use for backup detail pages or snapshot analysis.",
		"snapshots_table": "Multiple snapshots table card - Shows chronological table of multiple backup snapshots with sizes and status. Use for backup history, snapshot dashboard, or backup timeline views.",
	}

	description, isValid := cardDescriptions[cardType]
	if !isValid {
		validCards := make([]string, 0, len(cardDescriptions))
		for card := range cardDescriptions {
			validCards = append(validCards, card)
		}
		return "", fmt.Errorf("invalid card_type: %s. Valid types: %v", cardType, validCards)
	}

	// Construct the card URL
	cardURL := fmt.Sprintf("https://raw.githubusercontent.com/amcchord/slideReports/refs/heads/main/cards/%s.md", cardType)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the card
	resp, err := client.Get(cardURL)
	if err != nil {
		return "", fmt.Errorf("failed to download card: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download card: HTTP %d", resp.StatusCode)
	}

	// Read the card content
	cardContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read card content: %v", err)
	}

	// Return the card content with enhanced context
	response := fmt.Sprintf(`CARD TYPE: %s
DESCRIPTION: %s
SOURCE: %s

CARD CONTENT:
%s`, cardType, description, cardURL, string(cardContent))

	return response, nil
}

// getPresentationToolInfo returns the tool definition for the presentation meta-tool
func getPresentationToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_presentation",
		Description: "Format data professionally for user display. Use whenever showing system status, lists, reports, or documentation. Provides report templates (runbooks, daily/monthly reports) and display cards (individual items or tables). Supports HTML, Markdown, and HAML output. Operations: get_card (display agents, clients, devices, snapshots), get_runbook_template (procedures), get_daily_report_template (summaries), get_monthly_report_template (analysis).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "Operation to perform: get_card (display data), get_runbook_template (procedures), get_daily_report_template (daily summary), get_monthly_report_template (monthly analysis)",
					"enum":        []string{"get_runbook_template", "get_daily_report_template", "get_monthly_report_template", "get_card"},
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Template format - use 'html' for web display, 'haml' for Ruby/Rails apps, 'markdown' for documentation. Only applies to report templates, not cards.",
					"enum":        []string{"html", "haml", "markdown", "md"},
					"default":     "markdown",
				},
				"card_type": map[string]interface{}{
					"type":        "string",
					"description": "Required for get_card operation. Single item cards: agent, client, device, snapshot. Table cards for multiple items: agents_table, clients_table, devices_table, snapshots_table.",
					"enum":        []string{"agent", "agents_table", "client", "clients_table", "device", "devices_table", "snapshot", "snapshot_table"},
				},
			},
			"required": []string{"operation"},
		},
	}
}
