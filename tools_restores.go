package main

import (
	"fmt"
)

// handleRestoresTool handles all restore-related operations (file restores and image exports) through a single meta-tool
func handleRestoresTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_restores", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_restores in '%s' mode", operation, toolsMode)
	}

	switch operation {
	// File restore operations
	case "list_files":
		return listFileRestores(args)
	case "get_file":
		return getFileRestore(args)
	case "create_file":
		return createFileRestore(args)
	case "delete_file":
		return deleteFileRestore(args)
	case "browse_file":
		return browseFileRestore(args)
	// Image export operations
	case "list_images":
		return listImageExports(args)
	case "get_image":
		return getImageExport(args)
	case "create_image":
		return createImageExport(args)
	case "delete_image":
		return deleteImageExport(args)
	case "browse_image":
		return browseImageExport(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// getRestoresToolInfo returns the tool definition for the restores meta-tool
func getRestoresToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_restores",
		Description: "Manage restore operations - both file restores (for browsing/downloading individual files) and image exports (for downloading full disk images). Supports operations for both file restores and image exports.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"list_files", "get_file", "create_file", "delete_file", "browse_file", "list_images", "get_image", "create_image", "delete_image", "browse_image"},
				},
				// Common parameters for list operations
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of results per page (max 50) - used with list operations",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Pagination offset - used with list operations",
				},
				"sort_asc": map[string]interface{}{
					"type":        "boolean",
					"description": "Sort in ascending order - used with list operations",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"description": "Sort by field - used with list operations",
					"enum":        []string{"id"},
				},
				// Parameters for file restore operations
				"file_restore_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the file restore - required for get_file, delete_file, and browse_file operations",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to browse within the restore - required for browse_file operation",
				},
				// Parameters for image export operations
				"image_export_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the image export - required for get_image, delete_image, and browse_image operations",
				},
				"image_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of image to create - required for create_image operation",
					"enum":        []string{"vhdx", "vhdx-dynamic", "vhd", "vmdk", "vmdk-flat", "qcow2", "raw"},
				},
				// Common parameters for create operations
				"snapshot_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the snapshot to restore from - required for create_file and create_image operations",
				},
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the device to restore to - required for create_file and create_image operations",
				},
				"boot_mods": map[string]interface{}{
					"type":        "array",
					"description": "Optional boot modifications - used with create_image operation",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"required": []string{"operation"},
			"allOf": []map[string]interface{}{
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get_file"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"file_restore_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create_file"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"snapshot_id", "device_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete_file"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"file_restore_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "browse_file"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"file_restore_id", "path"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "get_image"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"image_export_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "create_image"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"snapshot_id", "device_id", "image_type"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "delete_image"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"image_export_id"},
					},
				},
				{
					"if": map[string]interface{}{
						"properties": map[string]interface{}{
							"operation": map[string]interface{}{"const": "browse_image"},
						},
					},
					"then": map[string]interface{}{
						"required": []string{"image_export_id"},
					},
				},
			},
		},
	}
}
