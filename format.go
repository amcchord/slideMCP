package main

// v4.0.0 cross-cutting response shaping. Replaces the per-handler
// `json.MarshalIndent` calls with a token-conscious formatter that supports:
//   - format=summary (default for lists, ~10 LoC per entry)
//   - format=compact (one-line JSON, no indentation)
//   - format=detailed (full payload, MarshalIndent — equivalent to v3)
//   - fields=a,b,c projection to drop everything else
//
// All four come from the same `args` map every tool handler already
// receives, so no signature changes ripple through.

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	formatSummary  = "summary"
	formatCompact  = "compact"
	formatDetailed = "detailed"
)

// extractFormat picks `format` out of the tool args and returns the
// canonical lower-cased value, defaulting to `defaultFormat`.
func extractFormat(args map[string]interface{}, defaultFormat string) string {
	if raw, ok := args["format"].(string); ok && raw != "" {
		switch strings.ToLower(raw) {
		case formatSummary, formatCompact, formatDetailed:
			return strings.ToLower(raw)
		}
	}
	return defaultFormat
}

// extractFields parses `fields=a,b,c` into a non-empty slice or nil.
func extractFields(args map[string]interface{}) []string {
	raw, ok := args["fields"].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// projectMap returns a new map containing only the keys in `fields`. If
// `fields` is empty/nil the original map is returned unchanged.
func projectMap(m map[string]interface{}, fields []string) map[string]interface{} {
	if len(fields) == 0 {
		return m
	}
	out := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		if v, ok := m[f]; ok {
			out[f] = v
		}
	}
	return out
}

// projectStruct marshals + remarshals a struct through interface{} so
// `fields=` projection works on typed values too.
func projectStruct(v interface{}, fields []string) (map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return m, nil
	}
	return projectMap(m, fields), nil
}

// formatJSON serializes `v` according to the active `format` mode.
//   - detailed: MarshalIndent
//   - summary: same as compact (lists handle their own per-entry shaping)
//   - compact: Marshal (no indent)
func formatJSON(v interface{}, format string) (string, error) {
	switch format {
	case formatDetailed:
		body, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(body), nil
	default:
		body, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(body), nil
	}
}

// formatList renders a paginated list response in the requested format.
//
//	summarize:  fn(entry) -> summary map[string]interface{}, used in `summary` mode
//	full:       []entry verbatim, used in `compact` and `detailed` modes
//
// All modes preserve the pagination envelope so the LLM can iterate.
func formatList[T any](
	items []T,
	pagination Pagination,
	args map[string]interface{},
	defaultFormat string,
	summarize func(T) map[string]interface{},
) (string, error) {
	format := extractFormat(args, defaultFormat)
	fields := extractFields(args)

	var data interface{}
	switch format {
	case formatSummary:
		out := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			s := summarize(it)
			out = append(out, projectMap(s, fields))
		}
		data = out
	default:
		if fields == nil {
			data = items
		} else {
			out := make([]map[string]interface{}, 0, len(items))
			for _, it := range items {
				m, err := projectStruct(it, fields)
				if err != nil {
					return "", err
				}
				out = append(out, m)
			}
			data = out
		}
	}

	envelope := map[string]interface{}{
		"data":       data,
		"pagination": pagination,
		"count":      len(items),
	}
	return formatJSON(envelope, format)
}

// formatSingle renders a single struct response in the requested format.
func formatSingle(v interface{}, args map[string]interface{}, defaultFormat string) (string, error) {
	format := extractFormat(args, defaultFormat)
	fields := extractFields(args)

	if fields == nil {
		return formatJSON(v, format)
	}
	m, err := projectStruct(v, fields)
	if err != nil {
		return "", err
	}
	return formatJSON(m, format)
}

// commonListProperties returns the JSON-Schema fragment shared by every
// tool that supports limit/offset/sort/format/fields. Merge into the
// `properties` block of the tool input schema.
func commonListProperties() map[string]interface{} {
	return map[string]interface{}{
		"limit": map[string]interface{}{
			"type":        "number",
			"description": "Number of results per page (max 50). Default: API default.",
			"minimum":     1,
			"maximum":     50,
		},
		"offset": map[string]interface{}{
			"type":        "number",
			"description": "Pagination offset; pass `pagination.next_offset` from the prior response to continue.",
			"minimum":     0,
		},
		"sort_asc": map[string]interface{}{
			"type":        "boolean",
			"description": "Sort ascending if true, descending if false.",
		},
		"format": map[string]interface{}{
			"type":        "string",
			"description": "Response density. `summary` (default for lists; one-line entries optimised for LLM context), `compact` (full payload, no indentation), `detailed` (full payload, indented).",
			"enum":        []string{"summary", "compact", "detailed"},
		},
		"fields": map[string]interface{}{
			"type":        "string",
			"description": "Comma-separated field projection, e.g. `id,hostname,last_seen`. Applies to each entry in lists or to the response object itself.",
		},
	}
}

// commonSinglePropertiesNoList returns just the format/fields fragment for
// non-list operations.
func commonSinglePropertiesNoList() map[string]interface{} {
	return map[string]interface{}{
		"format": map[string]interface{}{
			"type":        "string",
			"description": "Response density: `compact` (default), `detailed` (indented), `summary` (one-line summary).",
			"enum":        []string{"summary", "compact", "detailed"},
		},
		"fields": map[string]interface{}{
			"type":        "string",
			"description": "Comma-separated field projection, e.g. `id,hostname,last_seen`.",
		},
	}
}
