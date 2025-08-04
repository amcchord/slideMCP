package main

import (
	"fmt"
)

// OperationHandler represents a function that handles a specific operation
type OperationHandler func(map[string]interface{}) (string, error)

// ToolOperations maps operation names to their handler functions
type ToolOperations map[string]OperationHandler

// BaseToolConfig defines the configuration for a tool handler
type BaseToolConfig struct {
	ToolName   string
	Operations ToolOperations
}

// HandleToolWithOperations provides a shared base handler that eliminates boilerplate
// across all tool files by handling:
// 1. Operation parameter extraction and validation
// 2. Permission checking via config.IsOperationAllowed()
// 3. Operation dispatch to specific handlers
// 4. Standardized error handling
func HandleToolWithOperations(toolConfig BaseToolConfig, args map[string]interface{}) (string, error) {
	// Extract operation parameter
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !config.IsOperationAllowed(toolConfig.ToolName, operation) {
		return "", fmt.Errorf("operation '%s' not available for %s in '%s' mode", operation, toolConfig.ToolName, config.ToolsMode)
	}

	// Find and execute the operation handler
	handler, exists := toolConfig.Operations[operation]
	if !exists {
		return "", fmt.Errorf("unknown operation: %s", operation)
	}

	// Execute the specific operation handler
	return handler(args)
}

// CreateToolConfig is a helper function to create tool configurations more easily
func CreateToolConfig(toolName string, operations ToolOperations) BaseToolConfig {
	return BaseToolConfig{
		ToolName:   toolName,
		Operations: operations,
	}
}
