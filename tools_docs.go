package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Documentation Tool Usage Guide for LLMs
// ========================================
//
// The slide_docs tool provides access to Slide's documentation system. Here's how to use it effectively:
//
// 1. DISCOVERING DOCUMENTATION:
//    - Start with "list_sections" to see all available documentation categories
//    - Use "get_topics" with a specific section to drill down into topics
//    - Example flow: list_sections → get_topics (section: "API") → get_content (topic: "Authentication")
//
// 2. SEARCHING FOR INFORMATION:
//    - Use "search_docs" when you know what you're looking for but not where it is
//    - The search looks through both topic names and content
//    - Search is case-insensitive and finds partial matches
//    - Example: search_docs (query: "backup retention") finds all retention-related docs
//
// 3. GETTING SPECIFIC CONTENT:
//    - Use "get_content" when you know the exact topic name
//    - Topic names come from the "get_topics" operation
//    - Content is returned in Markdown format for easy reading
//
// 4. API REFERENCE:
//    - Use "get_api_reference" to get the complete OpenAPI specification
//    - This fetches the live OpenAPI spec from http://api.slide.tech/openapi.json
//    - The spec includes all endpoints, parameters, responses, and schemas
//    - For specific endpoint details, the OpenAPI spec is the authoritative source
//
// 5. BEST PRACTICES FOR LLMS:
//    - Always start broad (list_sections) if you're unsure where to find something
//    - Use search_docs for general queries before drilling into specific sections
//    - The API reference contains the complete, up-to-date API documentation
//    - Combine multiple operations to build comprehensive answers
//    - Remember that actual API calls should use the slide_* tools, not raw HTTP
//
// 6. COMMON PATTERNS:
//    - For "How do I..." questions: search_docs first, then get_content for details
//    - For API questions: get_api_reference for the complete OpenAPI spec
//    - For feature exploration: list_sections → get_topics → get_content
//    - For troubleshooting: search_docs with error keywords
//
// 7. RESPONSE FORMATS:
//    - All responses are JSON with a consistent structure
//    - Look for "_metadata" fields for additional context
//    - Content is typically in Markdown format for readability

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
		"Networks (Managing Networks)",
		"My Settings",
	},
	"Product": {
		"Backups",
		"Slide Agent",
		"Networking (Requirements)",
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

// Detailed descriptions for each documentation section to help with disambiguation
var docSectionDescriptions = map[string]string{
	"Getting Started": "Initial setup guides and tutorials for new users getting started with Slide backup solutions",
	"Slide Console":   "Web console interface documentation - how to manage and configure your Slide infrastructure through the UI. Includes managing networks on Slide devices/cloud",
	"Product":         "Technical product documentation including system requirements, networking prerequisites, and core product features",
	"Billing":         "Billing, subscription management, quotes, invoices, and payment-related documentation",
	"API":             "Developer documentation for the Slide API - endpoints, authentication, examples, and SDKs",
	"Troubleshooting": "Common issues, error codes, and problem-solving guides for various Slide components",
	"Best Practices":  "Recommended approaches for backup strategies, retention, security, and performance optimization",
}

// Topic descriptions for ambiguous items
var topicDescriptions = map[string]string{
	"Networks (Managing Networks)": "How to create, configure, and manage virtual networks on Slide devices and in the Slide cloud through the console",
	"Networking (Requirements)":    "Network infrastructure requirements, firewall rules, port configurations, and connectivity prerequisites for using Slide devices",
	"Backups":                      "Core backup functionality - types of backups, backup processes, scheduling, and verification",
	"Restores":                     "How to restore data from backups - file-level, image-level, and bare metal restore procedures",
	"Alerts":                       "Configuring and managing system alerts, notifications, and monitoring thresholds",
	"Users":                        "User management within the Slide console - creating users, permissions, and access control",
	"Clients":                      "Managing client organizations and multi-tenancy features in Slide",
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
	sections := make([]map[string]interface{}, 0, len(docSections))
	for section := range docSections {
		sectionInfo := map[string]interface{}{
			"name":        section,
			"description": docSectionDescriptions[section],
			"topic_count": len(docSections[section]),
		}
		sections = append(sections, sectionInfo)
	}

	result := map[string]interface{}{
		"sections": sections,
		"_metadata": map[string]interface{}{
			"description": "Available documentation sections from docs.slide.tech with contextual descriptions",
			"usage":       "Use 'get_topics' operation to see topics within a section. Pay attention to descriptions to choose the right section.",
			"note":        "Some sections have similar names but different purposes - check descriptions carefully",
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

	// Create topic list with descriptions where available
	topicList := make([]map[string]interface{}, 0, len(topics))
	for _, topic := range topics {
		topicInfo := map[string]interface{}{
			"name": topic,
		}
		if desc, hasDesc := topicDescriptions[topic]; hasDesc {
			topicInfo["description"] = desc
		}
		topicList = append(topicList, topicInfo)
	}

	result := map[string]interface{}{
		"section":             section,
		"section_description": docSectionDescriptions[section],
		"topics":              topicList,
		"_metadata": map[string]interface{}{
			"description": fmt.Sprintf("Topics available in the '%s' section", section),
			"usage":       "Use 'get_content' operation to retrieve specific topic content",
			"note":        "Topics with descriptions have been clarified to avoid ambiguity",
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
				searchResult := map[string]interface{}{
					"section":             section,
					"section_description": docSectionDescriptions[section],
					"topic":               topic,
					"type":                "topic_match",
				}
				// Add topic description if available
				if desc, hasDesc := topicDescriptions[topic]; hasDesc {
					searchResult["topic_description"] = desc
				}
				results = append(results, searchResult)
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
			"description": "Search results from Slide documentation with contextual information",
			"note":        "Results include section and topic descriptions to help identify the correct documentation",
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
	// Fetch the actual OpenAPI specification from the API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("http://api.slide.tech/openapi.json")
	if err != nil {
		// Fallback to basic information if fetch fails
		return getAPIReferenceFallback(args)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to basic information if fetch fails
		return getAPIReferenceFallback(args)
	}

	// Read the OpenAPI spec
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return getAPIReferenceFallback(args)
	}

	// Parse to ensure it's valid JSON
	var openAPISpec interface{}
	if err := json.Unmarshal(body, &openAPISpec); err != nil {
		return getAPIReferenceFallback(args)
	}

	// Return the complete OpenAPI specification
	result := map[string]interface{}{
		"openapi_spec": openAPISpec,
		"_metadata": map[string]interface{}{
			"source":      "http://api.slide.tech/openapi.json",
			"description": "Complete OpenAPI 3.0 specification for the Slide API",
			"usage_notes": []string{
				"This is the authoritative API documentation",
				"All endpoints, parameters, and schemas are defined here",
				"Use the slide_* tools to make actual API calls",
				"Authentication is handled automatically by the slide_* tools",
			},
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

func getAPIReferenceFallback(args map[string]interface{}) (string, error) {
	endpoint, _ := args["endpoint"].(string)

	// Provide API reference structure as fallback
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
			"description": "Basic API reference (fallback - OpenAPI spec unavailable)",
			"docs_url":    "https://docs.slide.tech/api",
			"openapi_url": "http://api.slide.tech/openapi.json",
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
		Name: "slide_docs",
		Description: `Access Slide documentation and API reference with enhanced contextual information. This tool provides comprehensive access to:
- Documentation sections and topics from docs.slide.tech with disambiguating descriptions
- Search functionality across all documentation with context-aware results
- Complete OpenAPI specification from http://api.slide.tech/openapi.json

Key features:
- Section descriptions help distinguish between similar-sounding sections (e.g., "Networks" in Console vs "Networking" in Product)
- Topic descriptions clarify ambiguous terms
- Search results include contextual information to help identify the correct documentation

Usage patterns:
- Start with 'list_sections' to explore available documentation with descriptions
- Use 'search_docs' to find information on specific topics with contextual results
- Use 'get_api_reference' to retrieve the complete, authoritative OpenAPI spec
- For API questions, always fetch the OpenAPI spec rather than relying on summaries
- Remember to use slide_* tools for actual API calls, not raw HTTP requests`,
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
					"description": "Specific API endpoint to get reference for (optional for get_api_reference) - Note: This parameter is ignored as the full OpenAPI spec is always returned",
				},
			},
			"required": []string{"operation"},
		},
	}
}
