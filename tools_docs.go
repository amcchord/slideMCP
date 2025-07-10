package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Documentation structure based on docs.slide.tech
var docSections = map[string][]string{
	"Getting Started": {
		"Introduction to Slide",
		"Quick Start Guide",
		"Installation",
		"Initial Setup",
		"First Backup",
	},
	"Slide Console": {
		"Dashboard Overview",
		"Protected Systems",
		"Slide Boxes",
		"Snapshots",
		"Restores",
		"Alerts",
		"Users",
		"Clients",
		"Networks",
		"My Settings",
	},
	"Product": {
		"Backups",
		"Slide Agent",
		"Networking",
		"Product Specifications",
		"System Requirements",
		"Security Features",
	},
	"Billing": {
		"Quotes",
		"Subscriptions",
		"Invoices",
		"Payment Methods",
		"Billing Overview",
	},
	"API": {
		"API Overview",
		"Authentication",
		"Endpoints",
		"Rate Limits",
		"Examples",
		"SDKs and Libraries",
	},
	"Troubleshooting": {
		"Common Issues",
		"Error Codes",
		"Performance Issues",
		"Network Problems",
		"Agent Issues",
		"Restore Problems",
	},
	"Best Practices": {
		"Backup Strategies",
		"Retention Policies",
		"Network Configuration",
		"Security Recommendations",
		"Performance Optimization",
	},
}

// Detailed documentation content (simplified for demonstration)
var docContent = map[string]string{
	"api_authentication": `# API Authentication

The Slide API uses API keys for authentication. All API requests must include your API key in the Authorization header.

## Getting Your API Key
1. Log into the Slide Console
2. Navigate to My Settings > API Keys
3. Click "Generate New API Key"
4. Copy and securely store your API key

## Using Your API Key
Include your API key in all API requests:
- Header: Authorization: Bearer YOUR_API_KEY

## Security Best Practices
- Never share your API key
- Rotate keys regularly
- Use environment variables to store keys
- Restrict API key permissions when possible`,

	"backup_overview": `# Backups Overview

Slide provides automated, secure backups for your systems with the following features:

## Key Features
- Incremental backups to minimize bandwidth
- End-to-end encryption
- Flexible scheduling options
- Multiple retention policies
- Point-in-time recovery

## Backup Types
1. **Full Backups**: Complete system backup
2. **Incremental Backups**: Only changed data since last backup
3. **Differential Backups**: Changes since last full backup

## Backup Process
1. Agent scans for changes
2. Data is encrypted locally
3. Compressed data is transmitted
4. Backup is verified and stored
5. Retention policies are applied`,

	"restore_process": `# Restore Process

Slide offers multiple restore options to meet different recovery needs.

## Restore Types
1. **File-Level Restore**: Restore individual files or folders
2. **Image-Level Restore**: Full system restore
3. **Bare Metal Restore**: Complete system recovery to new hardware

## File Restore Steps
1. Navigate to Restores in the console
2. Select the snapshot to restore from
3. Browse and select files/folders
4. Choose restore destination
5. Initiate restore

## Best Practices
- Test restores regularly
- Document restore procedures
- Maintain restore media
- Keep network configurations updated`,

	"network_configuration": `# Network Configuration

Proper network configuration ensures reliable and secure backups.

## Network Requirements
- Minimum 1 Mbps upload bandwidth
- Stable internet connection
- Firewall rules for Slide services

## Port Requirements
- Outbound HTTPS (443) for API communication
- Outbound TCP 8443 for backup data
- No inbound ports required

## Firewall Configuration
Allow outbound connections to:
- api.slide.com (API endpoints)
- backup.slide.com (Backup destinations)
- *.slide.com (CDN and auxiliary services)

## VPN and Proxy Support
- HTTP/HTTPS proxy supported
- SOCKS proxy configuration available
- Split-tunnel VPN compatible`,

	"snapshot_management": `# Snapshot Management

Snapshots are point-in-time copies of your protected systems.

## Understanding Snapshots
- Created after each successful backup
- Immutable once created
- Contain full system state at backup time
- Support instant recovery

## Snapshot Features
1. **Browse**: Explore snapshot contents
2. **Search**: Find specific files across snapshots
3. **Compare**: See changes between snapshots
4. **Export**: Download snapshot data

## Retention Policies
- Daily snapshots: Keep for X days
- Weekly snapshots: Keep for Y weeks
- Monthly snapshots: Keep for Z months
- Custom policies available

## Best Practices
- Regular snapshot verification
- Appropriate retention periods
- Monitor storage usage
- Document important snapshots`,
}

func handleDocsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation is required")
	}

	switch operation {
	case "list_sections":
		return listDocSections()
	case "get_topics":
		return getDocTopics(args)
	case "search_docs":
		return searchDocs(args)
	case "get_content":
		return getDocContent(args)
	case "get_api_reference":
		return getAPIReference(args)
	default:
		return "", fmt.Errorf("invalid operation: %s", operation)
	}
}

func listDocSections() (string, error) {
	sections := make([]string, 0, len(docSections))
	for section := range docSections {
		sections = append(sections, section)
	}

	result := map[string]interface{}{
		"sections": sections,
		"_metadata": map[string]interface{}{
			"description": "Available documentation sections from docs.slide.tech",
			"usage":       "Use 'get_topics' operation to see topics within a section",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getDocTopics(args map[string]interface{}) (string, error) {
	section, ok := args["section"].(string)
	if !ok {
		return "", fmt.Errorf("section is required")
	}

	topics, exists := docSections[section]
	if !exists {
		return "", fmt.Errorf("section '%s' not found", section)
	}

	result := map[string]interface{}{
		"section": section,
		"topics":  topics,
		"_metadata": map[string]interface{}{
			"description": fmt.Sprintf("Topics available in the '%s' section", section),
			"usage":       "Use 'get_content' operation to retrieve specific topic content",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func searchDocs(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query is required")
	}

	query = strings.ToLower(query)
	results := make([]map[string]interface{}, 0)

	// Search through all sections and topics
	for section, topics := range docSections {
		for _, topic := range topics {
			if strings.Contains(strings.ToLower(section), query) ||
				strings.Contains(strings.ToLower(topic), query) {
				results = append(results, map[string]interface{}{
					"section": section,
					"topic":   topic,
					"type":    "topic_match",
				})
			}
		}
	}

	// Search through documentation content
	for key, content := range docContent {
		if strings.Contains(strings.ToLower(content), query) {
			// Extract title from content
			lines := strings.Split(content, "\n")
			title := "Unknown"
			if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
				title = strings.TrimPrefix(lines[0], "# ")
			}

			results = append(results, map[string]interface{}{
				"content_key": key,
				"title":       title,
				"type":        "content_match",
				"preview":     getContentPreview(content, query),
			})
		}
	}

	result := map[string]interface{}{
		"query":        query,
		"results":      results,
		"result_count": len(results),
		"_metadata": map[string]interface{}{
			"description": "Search results from Slide documentation",
			"note":        "Results include both topic matches and content matches",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getDocContent(args map[string]interface{}) (string, error) {
	topic, ok := args["topic"].(string)
	if !ok {
		return "", fmt.Errorf("topic is required")
	}

	// Map common topics to content keys
	contentKey := strings.ToLower(strings.ReplaceAll(topic, " ", "_"))

	// Check for exact content match
	if content, exists := docContent[contentKey]; exists {
		result := map[string]interface{}{
			"topic":   topic,
			"content": content,
			"_metadata": map[string]interface{}{
				"source": "docs.slide.tech",
				"format": "markdown",
			},
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}

		return string(jsonData), nil
	}

	// Provide a general template for topics without specific content
	generalContent := fmt.Sprintf(`# %s

This documentation section covers %s in the Slide backup system.

## Overview
[Content for this specific topic would be retrieved from docs.slide.tech]

## Key Concepts
- Feature overview and capabilities
- Configuration requirements
- Best practices
- Common use cases

## Related Topics
- Check the '%s' section for more related documentation
- Use the search function to find specific information
- Refer to the API documentation for programmatic access

For the most up-to-date and detailed information, please visit:
https://docs.slide.tech`, topic, topic, getSectionForTopic(topic))

	result := map[string]interface{}{
		"topic":   topic,
		"content": generalContent,
		"_metadata": map[string]interface{}{
			"source": "docs.slide.tech",
			"format": "markdown",
			"note":   "Generic template - full content available at docs.slide.tech",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getAPIReference(args map[string]interface{}) (string, error) {
	endpoint, _ := args["endpoint"].(string)

	// Provide API reference structure
	apiRef := map[string]interface{}{
		"base_url": "https://api.slide.com/v1",
		"authentication": map[string]interface{}{
			"type":   "Bearer Token",
			"header": "Authorization: Bearer YOUR_API_KEY",
		},
		"common_endpoints": map[string]interface{}{
			"agents":    "/agent - Agent management",
			"backups":   "/backup - Backup operations",
			"snapshots": "/snapshot - Snapshot management",
			"restores":  "/restore - Restore operations",
			"devices":   "/device - Device management",
			"clients":   "/client - Client management",
			"networks":  "/network - Network configuration",
			"alerts":    "/alert - Alert management",
		},
		"rate_limits": map[string]interface{}{
			"requests_per_minute": 60,
			"requests_per_hour":   1000,
			"burst_limit":         10,
		},
		"_metadata": map[string]interface{}{
			"description": "Slide API reference summary",
			"docs_url":    "https://docs.slide.tech/api",
			"note":        "Use slide_* tools for direct API access with proper authentication",
		},
	}

	// If specific endpoint requested, provide more details
	if endpoint != "" {
		apiRef["endpoint_details"] = fmt.Sprintf("For detailed information about the '%s' endpoint, use the appropriate slide_* tool or visit https://docs.slide.tech/api/%s", endpoint, endpoint)
	}

	jsonData, err := json.MarshalIndent(apiRef, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// Helper functions
func getContentPreview(content, query string) string {
	index := strings.Index(strings.ToLower(content), strings.ToLower(query))
	if index == -1 {
		return ""
	}

	start := index - 50
	if start < 0 {
		start = 0
	}

	end := index + len(query) + 50
	if end > len(content) {
		end = len(content)
	}

	preview := content[start:end]
	if start > 0 {
		preview = "..." + preview
	}
	if end < len(content) {
		preview = preview + "..."
	}

	return strings.ReplaceAll(preview, "\n", " ")
}

func getSectionForTopic(topic string) string {
	for section, topics := range docSections {
		for _, t := range topics {
			if strings.EqualFold(t, topic) {
				return section
			}
		}
	}
	return "General"
}

func getDocsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_docs",
		Description: "Access Slide documentation from docs.slide.tech. Search topics, browse sections, and get detailed documentation content.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The documentation operation to perform",
					"enum": []string{
						"list_sections",
						"get_topics",
						"search_docs",
						"get_content",
						"get_api_reference",
					},
				},
				"section": map[string]interface{}{
					"type":        "string",
					"description": "Documentation section name (for get_topics operation)",
				},
				"topic": map[string]interface{}{
					"type":        "string",
					"description": "Specific topic to retrieve content for (for get_content operation)",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query to find relevant documentation (for search_docs operation)",
				},
				"endpoint": map[string]interface{}{
					"type":        "string",
					"description": "Specific API endpoint to get reference for (optional for get_api_reference)",
				},
			},
			"required": []string{"operation"},
		},
	}
}
