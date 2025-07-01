package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// handleReportsTool handles all reports-related operations through a single meta-tool
func handleReportsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_reports", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_reports in '%s' mode", operation, toolsMode)
	}

	switch operation {
	case "get_runbook_template":
		return getRunbookTemplate(args)
	case "get_daily_report_template":
		return getDailyReportTemplate(args)
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
	case "markdown", "md":
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/runbook/runbook.md"
	default:
		return "", fmt.Errorf("unsupported format: %s. Supported formats: html, markdown, md", format)
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
	case "markdown", "md":
		// Assuming markdown version exists at similar path
		templateURL = "https://raw.githubusercontent.com/amcchord/slideReports/main/daily_report/daily_report.md"
	default:
		return "", fmt.Errorf("unsupported format: %s. Supported formats: html, markdown, md", format)
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

// getReportsToolInfo returns the tool definition for the reports meta-tool
func getReportsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_reports",
		Description: "Get report templates and guidance for generating comprehensive reports. Supports downloading runbook templates and daily report templates from GitHub. The daily report template is specifically designed for documenting what happened in a single day across agents, devices, or clients. Feel free to ask users followup questions when generating reports to gather specific information needed.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform: 'get_runbook_template' for runbook templates, 'get_daily_report_template' for daily activity reports",
					"enum":        []string{"get_runbook_template", "get_daily_report_template"},
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Format of the template to retrieve - works with both runbook and daily report templates",
					"enum":        []string{"html", "markdown", "md"},
					"default":     "markdown",
				},
			},
			"required": []string{"operation"},
		},
	}
}
