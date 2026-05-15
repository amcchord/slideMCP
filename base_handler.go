package main

import (
	"fmt"
	"sort"
	"strings"
)

// OperationHandler represents a function that handles a specific operation
type OperationHandler func(map[string]interface{}) (string, error)

// ToolOperations maps operation names to their handler functions
type ToolOperations map[string]OperationHandler

// BaseToolConfig defines the configuration for a tool handler.
//
// Resolutions optionally maps operation -> ResolutionSpec so the
// dispatcher resolves args["name_hint"] into args[spec.IDKey] before
// the operation handler runs. Operations without an entry behave the
// same as v4.
type BaseToolConfig struct {
	ToolName    string
	Operations  ToolOperations
	Resolutions map[string]ResolutionSpec
}

// HandleToolWithOperations provides a shared base handler that eliminates boilerplate
// across all tool files by handling:
// 1. Operation parameter extraction and validation
// 2. Permission checking via config.IsOperationAllowed()
// 3. Optional name_hint -> *_id resolution before dispatch
// 4. Operation dispatch to specific handlers
// 5. Standardized error handling
func HandleToolWithOperations(toolConfig BaseToolConfig, args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	if !config.IsOperationAllowed(toolConfig.ToolName, operation) {
		return "", fmt.Errorf("operation '%s' not available for %s in '%s' mode", operation, toolConfig.ToolName, config.ToolsMode)
	}

	handler, exists := toolConfig.Operations[operation]
	if !exists {
		return "", unknownOperationError(toolConfig.ToolName, operation, toolConfig.Operations)
	}

	if spec, ok := toolConfig.Resolutions[operation]; ok {
		hintResp, err := resolveNameHint(args, spec)
		if err != nil {
			return "", err
		}
		if hintResp != "" {
			// Structured "no_match" / "ambiguous" response - surface as
			// the tool's successful response so the LLM can paraphrase
			// candidates back to the user.
			return hintResp, nil
		}
	}

	// Stash the tool name so format.go can compute next_steps hints
	// without each handler having to thread it through.
	args["_tool"] = toolConfig.ToolName

	return handler(args)
}

// CreateToolConfig is a helper function to create tool configurations more easily
func CreateToolConfig(toolName string, operations ToolOperations) BaseToolConfig {
	return BaseToolConfig{
		ToolName:   toolName,
		Operations: operations,
	}
}

// CreateToolConfigWithResolutions is the v5 variant that wires up the
// dispatcher-level name_hint resolver per-operation.
func CreateToolConfigWithResolutions(toolName string, operations ToolOperations, resolutions map[string]ResolutionSpec) BaseToolConfig {
	return BaseToolConfig{
		ToolName:    toolName,
		Operations:  operations,
		Resolutions: resolutions,
	}
}

// unknownOperationError produces a helpful "unknown operation" error
// including a Levenshtein-style nearest match (if one exists within
// edit distance 2) and the full list of valid operations.
func unknownOperationError(toolName, op string, operations ToolOperations) error {
	known := make([]string, 0, len(operations))
	for k := range operations {
		known = append(known, k)
	}
	sort.Strings(known)

	best := ""
	bestDist := 3
	for _, k := range known {
		d := levenshtein(op, k)
		if d < bestDist {
			best = k
			bestDist = d
		}
	}

	suggestion := ""
	if best != "" {
		suggestion = fmt.Sprintf(" (did you mean %q?)", best)
	}
	return fmt.Errorf("unknown operation %q for %s%s. Valid operations: %s",
		op, toolName, suggestion, strings.Join(known, ", "))
}

// levenshtein is a small DP implementation for did-you-mean suggestions.
// Operation names are short so we don't need anything fancy.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			min := del
			if ins < min {
				min = ins
			}
			if sub < min {
				min = sub
			}
			curr[j] = min
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}
