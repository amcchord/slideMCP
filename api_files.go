package main

// Slide API v1.27.0 file-search endpoints (FileSearch + PathVersion).
// These are the highest-value LLM-facing operations in the API surface
// because they let the model jump straight from a natural-language file
// reference ("the Q4 budget spreadsheet on Bob's laptop") to a snapshot
// + path tuple ready for restore.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// FileIndexSearch is one entry in the file-search result set.
// Response shape from GET /v1/agent/{agent_id}/file-search.
type FileIndexSearch struct {
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modified_time"`
}

// PathVersion is one historical version of a path on a single agent.
// Response shape from GET /v1/agent/{agent_id}/file-search/version.
type PathVersion struct {
	SnapshotID   string `json:"snapshot_id"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modified_time"`
}

// searchAgentFiles wraps GET /v1/agent/{agent_id}/file-search.
func searchAgentFiles(agentID, searchTerm string, limit, offset int, sortBy string, sortAsc *bool) (*PaginatedResponse[FileIndexSearch], error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if searchTerm == "" {
		return nil, fmt.Errorf("search_term is required")
	}
	params := url.Values{}
	params.Set("search_term", searchTerm)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	if sortBy != "" {
		params.Set("sort_by", sortBy)
	}
	if sortAsc != nil {
		params.Set("sort_asc", strconv.FormatBool(*sortAsc))
	}
	endpoint := fmt.Sprintf("/v1/agent/%s/file-search?%s", url.PathEscape(agentID), params.Encode())
	body, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var resp PaginatedResponse[FileIndexSearch]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse file-search response: %w", err)
	}
	return &resp, nil
}

// listPathVersions wraps GET /v1/agent/{agent_id}/file-search/version.
func listPathVersions(agentID, path string, limit, offset int, sortBy string, sortAsc *bool) (*PaginatedResponse[PathVersion], error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	params := url.Values{}
	params.Set("path", path)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	if sortBy != "" {
		params.Set("sort_by", sortBy)
	}
	if sortAsc != nil {
		params.Set("sort_asc", strconv.FormatBool(*sortAsc))
	}
	endpoint := fmt.Sprintf("/v1/agent/%s/file-search/version?%s", url.PathEscape(agentID), params.Encode())
	body, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var resp PaginatedResponse[PathVersion]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse path-version response: %w", err)
	}
	return &resp, nil
}
