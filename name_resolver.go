package main

// Server-side fuzzy name resolution for agent_id / device_id / client_id.
//
// Slide IDs are opaque (`a_xxxxxxxxxxxx`, `d_...`, `c_...`); operators
// never know them. v5 lets every tool accept `name_hint` as an
// alternative: pass a hostname, display name, client name, or any
// substring, and the server resolves it server-side before dispatching.
//
// Match precedence (deterministic, no fuzzy distance metric for v5.0):
//   1. Case-insensitive exact match against display_name OR hostname
//      (or client.name). If any matches, stop at this tier.
//   2. Case-insensitive prefix match. Stop at this tier if any matches.
//   3. Case-insensitive substring match. Stop at this tier if any matches.
//   4. No match.
//
// On exactly one match the resolved ID is written into args[idKey] so
// per-operation handlers continue to work unchanged.
//
// On zero or multiple matches the resolver returns a JSON body
// describing the failure (a "name_hint_error" payload) that
// HandleToolWithOperations surfaces as the tool's response with
// isError=false - the LLM gets structured candidates it can paraphrase
// back to the user, then re-call with an explicit *_id.

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

const nameResolverTTL = 15 * time.Minute

// Slide IDs look like `<prefix>_<base32-ish-suffix>`. We accept the
// well-known prefixes for the kinds we resolve; anything else is treated
// as a candidate name_hint instead of an ID.
var slideIDPattern = regexp.MustCompile(`^[a-z]_[A-Za-z0-9]+$`)

// looksLikeSlideID is a cheap guard so an exact ID short-circuits
// before we hit any cache or fuzzy logic.
func looksLikeSlideID(s string) bool {
	return slideIDPattern.MatchString(s)
}

// nameCandidate is one entry in the cache.
type nameCandidate struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
}

type nameCacheEntry struct {
	candidates []nameCandidate
	fetchedAt  time.Time
}

var (
	nameCacheMu sync.RWMutex
	nameCache   = map[string]nameCacheEntry{}
)

// nameCacheGet returns a non-empty, non-stale cache entry for kind, or
// the zero value if the cache should be refreshed.
func nameCacheGet(kind string) (nameCacheEntry, bool) {
	nameCacheMu.RLock()
	defer nameCacheMu.RUnlock()
	e, ok := nameCache[kind]
	if !ok {
		return nameCacheEntry{}, false
	}
	if time.Since(e.fetchedAt) > nameResolverTTL {
		return nameCacheEntry{}, false
	}
	return e, true
}

func nameCachePut(kind string, candidates []nameCandidate) {
	nameCacheMu.Lock()
	defer nameCacheMu.Unlock()
	nameCache[kind] = nameCacheEntry{candidates: candidates, fetchedAt: time.Now()}
}

// resetNameCache is used by tests to force a refresh between scenarios.
func resetNameCache() {
	nameCacheMu.Lock()
	defer nameCacheMu.Unlock()
	nameCache = map[string]nameCacheEntry{}
}

// fetchCandidates pulls the live list of agents / devices / clients and
// maps each one into one or more nameCandidate rows. Agents and devices
// publish both display_name and hostname; we emit both as candidates so
// either lookup style works.
func fetchCandidates(kind string) ([]nameCandidate, error) {
	switch kind {
	case "agent":
		body, err := makeAPIRequest("GET", "/v1/agent?limit=50", nil)
		if err != nil {
			return nil, err
		}
		var p PaginatedResponse[Agent]
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, fmt.Errorf("parse agents: %w", err)
		}
		out := make([]nameCandidate, 0, len(p.Data)*2)
		seen := map[string]bool{}
		for _, a := range p.Data {
			detail := strings.TrimSpace(a.OS + " " + a.OSVersion)
			if a.DisplayName != "" && !seen[a.AgentID+"|display|"+strings.ToLower(a.DisplayName)] {
				out = append(out, nameCandidate{ID: a.AgentID, Name: a.DisplayName, Kind: "agent", Detail: detail})
				seen[a.AgentID+"|display|"+strings.ToLower(a.DisplayName)] = true
			}
			if a.Hostname != "" && strings.ToLower(a.Hostname) != strings.ToLower(a.DisplayName) {
				out = append(out, nameCandidate{ID: a.AgentID, Name: a.Hostname, Kind: "agent", Detail: detail})
			}
		}
		return out, nil
	case "device":
		body, err := makeAPIRequest("GET", "/v1/device?limit=50", nil)
		if err != nil {
			return nil, err
		}
		var p PaginatedResponse[Device]
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, fmt.Errorf("parse devices: %w", err)
		}
		out := make([]nameCandidate, 0, len(p.Data)*2)
		for _, d := range p.Data {
			detail := d.ServiceStatus
			if d.DisplayName != "" {
				out = append(out, nameCandidate{ID: d.DeviceID, Name: d.DisplayName, Kind: "device", Detail: detail})
			}
			if d.Hostname != "" && strings.ToLower(d.Hostname) != strings.ToLower(d.DisplayName) {
				out = append(out, nameCandidate{ID: d.DeviceID, Name: d.Hostname, Kind: "device", Detail: detail})
			}
		}
		return out, nil
	case "client":
		body, err := makeAPIRequest("GET", "/v1/client?limit=50", nil)
		if err != nil {
			return nil, err
		}
		var p PaginatedResponse[Client]
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, fmt.Errorf("parse clients: %w", err)
		}
		out := make([]nameCandidate, 0, len(p.Data))
		for _, c := range p.Data {
			out = append(out, nameCandidate{ID: c.ClientID, Name: c.Name, Kind: "client"})
		}
		return out, nil
	}
	return nil, fmt.Errorf("unknown resolution kind: %s", kind)
}

// ensureCandidates returns the cached list for kind, fetching if stale.
func ensureCandidates(kind string) ([]nameCandidate, error) {
	if e, ok := nameCacheGet(kind); ok {
		return e.candidates, nil
	}
	candidates, err := fetchCandidates(kind)
	if err != nil {
		return nil, err
	}
	nameCachePut(kind, candidates)
	return candidates, nil
}

// matchByName applies the deterministic 3-tier matching: exact > prefix
// > substring, returning the highest-quality non-empty tier.
func matchByName(candidates []nameCandidate, hint string) []nameCandidate {
	needle := strings.ToLower(strings.TrimSpace(hint))
	if needle == "" {
		return nil
	}

	exact := []nameCandidate{}
	prefix := []nameCandidate{}
	substr := []nameCandidate{}
	seenID := map[string]bool{}
	addOnce := func(slice []nameCandidate, c nameCandidate) []nameCandidate {
		if seenID[c.ID] {
			return slice
		}
		seenID[c.ID] = true
		return append(slice, c)
	}

	for _, c := range candidates {
		name := strings.ToLower(c.Name)
		switch {
		case name == needle:
			exact = addOnce(exact, c)
		case strings.HasPrefix(name, needle):
			prefix = addOnce(prefix, c)
		case strings.Contains(name, needle):
			substr = addOnce(substr, c)
		}
	}

	if len(exact) > 0 {
		return exact
	}
	if len(prefix) > 0 {
		return prefix
	}
	return substr
}

// ResolutionSpec is attached to a meta-tool's BaseToolConfig (one per
// operation that should auto-resolve a *_id from name_hint). The
// dispatcher consumes it in HandleToolWithOperations.
type ResolutionSpec struct {
	IDKey string // e.g. "agent_id", "device_id", "client_id"
	Kind  string // "agent" | "device" | "client"
}

// resolveNameHint is the dispatcher-level helper. Behaviour:
//
//   - If args[idKey] is already a non-empty string, return ("", "", nil)
//     (no resolution attempted - the existing requireString-in-handler
//     path will use that value).
//   - Else if args["name_hint"] is a non-empty string, fuzzy-match and:
//     * 0 matches  -> hintResp is a JSON payload describing the miss.
//     * 1 match    -> args[idKey] gets the resolved ID written into it,
//                     return ("", "", nil) so the handler proceeds.
//     * 2+ matches -> hintResp is a JSON payload describing the candidates.
//   - Else: return ("", "", nil) - the handler's own missing-id error
//     fires when it does its requireString check.
func resolveNameHint(args map[string]interface{}, spec ResolutionSpec) (hintResp string, err error) {
	if cur, ok := args[spec.IDKey].(string); ok && strings.TrimSpace(cur) != "" {
		return "", nil
	}
	hint, _ := args["name_hint"].(string)
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return "", nil
	}

	// If someone passes a literal Slide ID through name_hint (e.g. they
	// copy/paste from the inventory output), accept it.
	if looksLikeSlideID(hint) {
		args[spec.IDKey] = hint
		return "", nil
	}

	candidates, ferr := ensureCandidates(spec.Kind)
	if ferr != nil {
		return "", fmt.Errorf("name_hint resolution: %w", ferr)
	}
	matches := matchByName(candidates, hint)

	switch len(matches) {
	case 0:
		// Force-refresh once on a miss in case the cache is stale -
		// new agents/devices/clients show up frequently in MSP usage.
		resetNameCache()
		candidates, ferr = ensureCandidates(spec.Kind)
		if ferr == nil {
			matches = matchByName(candidates, hint)
		}
	}

	switch len(matches) {
	case 0:
		payload := map[string]interface{}{
			"name_hint_error": "no_match",
			"name_hint":       hint,
			"kind":            spec.Kind,
			"suggestion": fmt.Sprintf(
				"No %s matched %q. Call slide_overview operation=inventory to see the full list, then re-call with the explicit %s.",
				spec.Kind, hint, spec.IDKey),
		}
		b, _ := json.MarshalIndent(payload, "", "  ")
		return string(b), nil
	case 1:
		args[spec.IDKey] = matches[0].ID
		// Stash the resolution so format.go can surface it in the response.
		args["_resolution"] = map[string]interface{}{
			"name_hint": hint,
			"resolved": map[string]interface{}{
				"id":     matches[0].ID,
				"name":   matches[0].Name,
				"kind":   matches[0].Kind,
				"detail": matches[0].Detail,
			},
		}
		return "", nil
	default:
		// Trim ambiguous candidate list to 10 to avoid context blowups.
		shown := matches
		if len(shown) > 10 {
			shown = shown[:10]
		}
		payload := map[string]interface{}{
			"name_hint_error": "ambiguous",
			"name_hint":       hint,
			"kind":            spec.Kind,
			"candidates":      shown,
			"hint": fmt.Sprintf(
				"%d %ss matched %q (showing up to 10). Ask the user to pick one, then re-call with %s=<id from candidates>.",
				len(matches), spec.Kind, hint, spec.IDKey),
		}
		b, _ := json.MarshalIndent(payload, "", "  ")
		return string(b), nil
	}
}
