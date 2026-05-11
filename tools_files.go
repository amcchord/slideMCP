package main

// slide_files: search, version, restore, push. Consolidates the v1.27.0
// FileSearch + PathVersion endpoints with the existing file-restore CRUD
// + push so a single tool covers the entire "find and recover a file"
// workflow an MSP tech actually executes.

import (
	"encoding/json"
	"fmt"
)

func handleFilesTool(args map[string]interface{}) (string, error) {
	return HandleToolWithOperations(CreateToolConfig("slide_files", ToolOperations{
		"search":          handleFilesSearch,
		"versions":        handleFilesVersions,
		"list_restores":   listFileRestores,
		"get_restore":     getFileRestore,
		"create_restore":  createFileRestore,
		"delete_restore":  deleteFileRestore,
		"browse":          handleFilesBrowse,
		"list_pushes":     listFileRestorePushes,
		"create_push":     createFileRestorePush,
		"update_push":     updateFileRestorePush,
		"get_push_status": handleFilesGetPushStatus,
	}), args)
}

var filesOperationEnums = []string{
	"search", "versions",
	"list_restores", "get_restore", "create_restore", "delete_restore",
	"browse",
	"list_pushes", "create_push", "update_push", "get_push_status",
}

func getFilesToolInfo() ToolInfo {
	props := map[string]interface{}{
		"operation": map[string]interface{}{
			"type":        "string",
			"description": "The operation to perform.",
			"enum":        filesOperationEnums,
		},
		"agent_id": map[string]interface{}{
			"type":        "string",
			"description": "ID of the agent to search. Required for `search` and `versions`.",
		},
		"search_term": map[string]interface{}{
			"type":        "string",
			"description": "File name (or substring) to search for. Required for `search`. Matches anywhere in the path.",
		},
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Full file path to look up versions for, e.g. `C:\\Users\\bob\\Documents\\Q4-budget.xlsx`. Required for `versions`.",
		},
		"sort_by_search": map[string]interface{}{
			"type":        "string",
			"description": "Sort field for `search` (default `path`).",
			"enum":        []string{"path", "last_modified", "file_size"},
		},
		"sort_by_versions": map[string]interface{}{
			"type":        "string",
			"description": "Sort field for `versions` (default `created_time`).",
			"enum":        []string{"created_time", "last_modified", "file_size"},
		},
		"snapshot_id": map[string]interface{}{
			"type":        "string",
			"description": "Snapshot ID to restore from. Required for `create_restore`.",
		},
		"device_id": map[string]interface{}{
			"type":        "string",
			"description": "Device that owns the snapshot. Required for `create_restore`.",
		},
		"file_restore_id": map[string]interface{}{
			"type":        "string",
			"description": "Restore session ID. Required for `get_restore`, `delete_restore`, `browse`, `list_pushes`, `create_push`, `update_push`, `get_push_status`.",
		},
		"browse_path": map[string]interface{}{
			"type":        "string",
			"description": "Folder path inside the restore to browse, e.g. `/C:/Users`. Required for `browse`.",
		},
		"source_file_path": map[string]interface{}{
			"type":        "string",
			"description": "Path of the file inside the restore to push back. Required for `create_push`.",
		},
		"destination_folder": map[string]interface{}{
			"type":        "string",
			"description": "Destination on the protected system. Must end in `\\SlideRestore` (e.g. `C:\\SlideRestore`). Required for `create_push`.",
		},
		"file_restore_push_id": map[string]interface{}{
			"type":        "string",
			"description": "Push operation ID. Required for `update_push` and `get_push_status`.",
		},
		"state": map[string]interface{}{
			"type":        "string",
			"description": "Push state. Only `canceled` is accepted. Required for `update_push`.",
			"enum":        []string{"canceled"},
		},
	}
	for k, v := range commonListProperties() {
		props[k] = v
	}

	return ToolInfo{
		Name: "slide_files",
		Description: "Search agent file indexes and run file-level restores. " +
			"Operations: `search` (find a file across an agent's snapshots), `versions` (list snapshots that contain a path), " +
			"`list_restores`/`get_restore`/`create_restore`/`delete_restore` (manage restore sessions), `browse` (walk a restore), " +
			"`list_pushes`/`create_push`/`update_push`/`get_push_status` (push a file back to the protected system). " +
			"Typical recovery flow: `search` -> pick a path -> `versions` -> pick a snapshot -> `create_restore` -> `browse` -> `create_push`.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": props,
			"required":   []string{"operation"},
			"allOf": []map[string]interface{}{
				{"if": ifOp("search"), "then": req("agent_id", "search_term")},
				{"if": ifOp("versions"), "then": req("agent_id", "path")},
				{"if": ifOp("get_restore"), "then": req("file_restore_id")},
				{"if": ifOp("create_restore"), "then": req("snapshot_id", "device_id")},
				{"if": ifOp("delete_restore"), "then": req("file_restore_id")},
				{"if": ifOp("browse"), "then": req("file_restore_id", "browse_path")},
				{"if": ifOp("list_pushes"), "then": req("file_restore_id")},
				{"if": ifOp("create_push"), "then": req("file_restore_id", "source_file_path", "destination_folder")},
				{"if": ifOp("update_push"), "then": req("file_restore_id", "file_restore_push_id", "state")},
				{"if": ifOp("get_push_status"), "then": req("file_restore_id", "file_restore_push_id")},
			},
		},
	}
}

// --- search & versions handlers ----------------------------------------

func handleFilesSearch(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	searchTerm, err := requireString(args, "search_term")
	if err != nil {
		return "", err
	}
	limit, _ := optionalInt(args, "limit")
	offset, _ := optionalInt(args, "offset")
	sortBy, _ := optionalString(args, "sort_by_search")
	if sortBy == "" {
		// fall back to the cross-cutting `sort_by`
		sortBy, _ = optionalString(args, "sort_by")
	}
	var sortAscPtr *bool
	if v, ok := optionalBool(args, "sort_asc"); ok {
		sortAscPtr = &v
	}

	resp, err := searchAgentFiles(agentID, searchTerm, limit, offset, sortBy, sortAscPtr)
	if err != nil {
		return "", err
	}
	return formatList(resp.Data, resp.Pagination, args, formatSummary, func(f FileIndexSearch) map[string]interface{} {
		return map[string]interface{}{
			"path":          f.Path,
			"size":          f.Size,
			"modified_time": f.ModifiedTime,
		}
	})
}

func handleFilesVersions(args map[string]interface{}) (string, error) {
	agentID, err := requireString(args, "agent_id")
	if err != nil {
		return "", err
	}
	path, err := requireString(args, "path")
	if err != nil {
		return "", err
	}
	limit, _ := optionalInt(args, "limit")
	offset, _ := optionalInt(args, "offset")
	sortBy, _ := optionalString(args, "sort_by_versions")
	if sortBy == "" {
		sortBy, _ = optionalString(args, "sort_by")
	}
	var sortAscPtr *bool
	if v, ok := optionalBool(args, "sort_asc"); ok {
		sortAscPtr = &v
	}

	resp, err := listPathVersions(agentID, path, limit, offset, sortBy, sortAscPtr)
	if err != nil {
		return "", err
	}
	return formatList(resp.Data, resp.Pagination, args, formatSummary, func(v PathVersion) map[string]interface{} {
		return map[string]interface{}{
			"snapshot_id":   v.SnapshotID,
			"size":          v.Size,
			"modified_time": v.ModifiedTime,
		}
	})
}

// handleFilesBrowse forwards to browseFileRestore, mapping our `browse_path`
// arg onto the legacy `path` parameter so the slide_files tool stays
// internally consistent.
func handleFilesBrowse(args map[string]interface{}) (string, error) {
	if bp, ok := args["browse_path"].(string); ok {
		args["path"] = bp
	}
	return browseFileRestore(args)
}

func handleFilesGetPushStatus(args map[string]interface{}) (string, error) {
	fileRestoreID, err := requireString(args, "file_restore_id")
	if err != nil {
		return "", err
	}
	pushID, err := requireString(args, "file_restore_push_id")
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("/v1/restore/file/%s/push/%s", fileRestoreID, pushID)
	body, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	var p FileRestorePush
	if err := json.Unmarshal(body, &p); err != nil {
		return "", fmt.Errorf("failed to parse push status response: %w", err)
	}
	return formatSingle(p, args, formatCompact)
}
