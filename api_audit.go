package main

// Slide API v1.27.0 audit-log endpoints (Audits, AuditByID, AuditActions,
// AuditResources). Critical for compliance + "what just changed" questions.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Audit is the response shape for GET /v1/audit/{audit_id} and one entry
// in the GET /v1/audit list.
type Audit struct {
	AuditID          string  `json:"audit_id"`
	AccountID        string  `json:"account_id"`
	ClientID         *string `json:"client_id,omitempty"`
	UserID           *string `json:"user_id,omitempty"`
	UserDisplayName  *string `json:"user_display_name,omitempty"`
	Action           string  `json:"action"`
	ActionFieldsJSON string  `json:"action_fields_json,omitempty"`
	ResourceID       string  `json:"resource_id"`
	ResourceType     string  `json:"resource_type"`
	Description      string  `json:"description,omitempty"`
	AuditTime        string  `json:"audit_time"`
}

// AuditAction is one entry returned by GET /v1/audit/action.
type AuditAction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AuditResourceType is one entry returned by GET /v1/audit/resource.
type AuditResourceType struct {
	Name string `json:"name"`
}

// listAudits wraps GET /v1/audit.
func listAudits(opts auditQueryOpts) (*PaginatedResponse[Audit], error) {
	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.ActionName != "" {
		params.Set("audit_action_name", opts.ActionName)
	}
	if opts.ResourceType != "" {
		params.Set("audit_resource_type_name", opts.ResourceType)
	}
	if opts.SortBy != "" {
		params.Set("sort_by", opts.SortBy)
	}
	if opts.SortAsc != nil {
		params.Set("sort_asc", strconv.FormatBool(*opts.SortAsc))
	}
	if opts.AuditTimeBefore != "" {
		params.Set("audit_time_before", opts.AuditTimeBefore)
	}
	if opts.AuditTimeAfter != "" {
		params.Set("audit_time_after", opts.AuditTimeAfter)
	}
	endpoint := "/v1/audit"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	body, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var resp PaginatedResponse[Audit]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse audit list response: %w", err)
	}
	return &resp, nil
}

type auditQueryOpts struct {
	Limit           int
	Offset          int
	ActionName      string
	ResourceType    string
	SortBy          string
	SortAsc         *bool
	AuditTimeBefore string
	AuditTimeAfter  string
}

// getAudit wraps GET /v1/audit/{audit_id}.
func getAudit(auditID string) (*Audit, error) {
	if auditID == "" {
		return nil, fmt.Errorf("audit_id is required")
	}
	body, err := makeAPIRequest("GET", fmt.Sprintf("/v1/audit/%s", url.PathEscape(auditID)), nil)
	if err != nil {
		return nil, err
	}
	var a Audit
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("failed to parse audit response: %w", err)
	}
	return &a, nil
}

// listAuditActions wraps GET /v1/audit/action.
func listAuditActions(limit, offset int) (*PaginatedResponse[AuditAction], error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	endpoint := "/v1/audit/action"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	body, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var resp PaginatedResponse[AuditAction]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse audit actions response: %w", err)
	}
	return &resp, nil
}

// listAuditResourceTypes wraps GET /v1/audit/resource.
func listAuditResourceTypes(limit, offset int) (*PaginatedResponse[AuditResourceType], error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	endpoint := "/v1/audit/resource"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	body, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	var resp PaginatedResponse[AuditResourceType]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse audit resource types response: %w", err)
	}
	return &resp, nil
}
