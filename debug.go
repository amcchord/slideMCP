package main

// Debug: comprehensive runtime + network + API diagnostics. Two ways to
// invoke:
//
//   - From a chat: `slide_help operation=debug` runs the same probes and
//     returns the result as a structured JSON payload the LLM (and the
//     operator) can read.
//   - From a shell: `slide-mcp-server --debug` prints the same payload as
//     pretty JSON and exits 0.
//
// The output is intentionally safe to paste back in a support thread:
// the API token is shown only as `<prefix>...<suffix>` (length + first 4
// and last 2 chars), no other secrets are emitted.

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// init wires log.Printf output through stderrTee so the in-memory
// capture buffer holds the same lines Claude Desktop sees in its
// extension log panel. Per Go's log package, SetOutput is goroutine-safe
// and applies to the default logger plus log.Default().
func init() {
	log.SetOutput(stderrTee{})
}

// startedAt records the process boot time so debug can report uptime.
var startedAt = time.Now()

// captureBuffer is a bounded ring of recent stderr lines. log.Printf calls
// (and any explicit stderr writes routed through stderrTee) get appended
// so the debug output can include "recent_logs" as a context block.
type captureBuffer struct {
	mu    sync.Mutex
	cap   int
	lines []string
}

func newCaptureBuffer(cap int) *captureBuffer {
	return &captureBuffer{cap: cap}
}

// addLines splits incoming chunks into trimmed lines and appends them.
func (c *captureBuffer) addLines(chunk []byte) {
	for _, raw := range strings.Split(string(chunk), "\n") {
		s := strings.TrimRight(raw, "\r")
		if s == "" {
			continue
		}
		c.mu.Lock()
		c.lines = append(c.lines, s)
		if len(c.lines) > c.cap {
			c.lines = c.lines[len(c.lines)-c.cap:]
		}
		c.mu.Unlock()
	}
}

// Snapshot returns a copy of the current line buffer for safe iteration.
func (c *captureBuffer) Snapshot() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.lines))
	copy(out, c.lines)
	return out
}

var logCapture = newCaptureBuffer(200)

// stderrTee is an io.Writer that mirrors output to os.Stderr AND the
// in-memory log capture. log.SetOutput is pointed at it from init() so
// every log.Printf call ends up in both places.
type stderrTee struct{}

func (stderrTee) Write(p []byte) (int, error) {
	logCapture.addLines(p)
	return os.Stderr.Write(p)
}

// stderrLogger is a separate log.Logger with no prefix/timestamp,
// intended for multi-line messages (warnings, remediation hints) that
// shouldn't have a `2026/05/15 17:00:00` smear in front of each line.
// Also writes to the capture buffer.
var stderrLogger = newPlainStderrLogger()

func newPlainStderrLogger() *plainLogger {
	return &plainLogger{}
}

// plainLogger is a minimal Print-style helper that writes raw text to
// stderr and to the capture buffer in one call.
type plainLogger struct{}

func (plainLogger) Println(args ...interface{}) {
	out := fmt.Sprintln(args...)
	logCapture.addLines([]byte(out))
	fmt.Fprint(os.Stderr, out)
}

func (plainLogger) Printf(format string, args ...interface{}) {
	out := fmt.Sprintf(format, args...)
	logCapture.addLines([]byte(out))
	fmt.Fprint(os.Stderr, out)
}

// maskedAPIKey returns a safe-to-paste rendering of the active token.
// Shape: <first4>...<last2> (len=<n>), or "<empty>" if no token is set.
func maskedAPIKey() string {
	if apiKey == "" {
		return "<empty>"
	}
	n := len(apiKey)
	if n <= 6 {
		return fmt.Sprintf("<short> (len=%d)", n)
	}
	return fmt.Sprintf("%s...%s (len=%d)", apiKey[:4], apiKey[n-2:], n)
}

// envSource describes how a config value was sourced (CLI flag, env, default).
type envSource struct {
	Value  string `json:"value"`
	Source string `json:"source"`
}

// envSnapshot inspects the process environment for our recognised vars
// and returns whether each is set. Values are masked appropriately.
func envSnapshot() map[string]interface{} {
	mask := func(v string) string {
		if v == "" {
			return ""
		}
		n := len(v)
		if n <= 6 {
			return fmt.Sprintf("<short> (len=%d)", n)
		}
		return fmt.Sprintf("%s...%s (len=%d)", v[:4], v[n-2:], n)
	}
	return map[string]interface{}{
		"SLIDE_API_KEY":         mask(os.Getenv("SLIDE_API_KEY")),
		"SLIDE_API_KEY_set":     os.Getenv("SLIDE_API_KEY") != "",
		"SLIDE_BASE_URL":        os.Getenv("SLIDE_BASE_URL"),
		"SLIDE_TOOLS":           os.Getenv("SLIDE_TOOLS"),
		"SLIDE_DISABLED_TOOLS":  os.Getenv("SLIDE_DISABLED_TOOLS"),
	}
}

// dnsProbe resolves api.slide.tech (or whatever host base_url points at)
// and reports the addresses. Useful to confirm DNS isn't being hijacked.
func dnsProbe() map[string]interface{} {
	host := APIBaseURL
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	start := time.Now()
	addrs, err := net.LookupHost(host)
	latency := time.Since(start).Milliseconds()
	out := map[string]interface{}{
		"host":       host,
		"latency_ms": latency,
	}
	if err != nil {
		out["error"] = err.Error()
		return out
	}
	out["addresses"] = addrs
	return out
}

// tlsProbe performs an unauthenticated HEAD against base_url to confirm
// TLS + reachability. Status + latency + cert subject.
func tlsProbe() map[string]interface{} {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	cli := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	start := time.Now()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", APIBaseURL+"/", nil)
	resp, err := cli.Do(req)
	latency := time.Since(start).Milliseconds()
	out := map[string]interface{}{
		"url":        APIBaseURL + "/",
		"latency_ms": latency,
	}
	if err != nil {
		out["error"] = err.Error()
		return out
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	out["status"] = resp.StatusCode
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		out["tls_subject"] = cert.Subject.String()
		out["tls_issuer"] = cert.Issuer.String()
		out["tls_not_after"] = cert.NotAfter.Format(time.RFC3339)
	}
	return out
}

// endpointProbe makes one authenticated GET against an endpoint, capturing
// status, latency, body excerpt, and error so the operator can see EXACTLY
// what the Slide API is returning right now (no guessing from the LLM).
func endpointProbe(method, endpoint string) map[string]interface{} {
	out := map[string]interface{}{
		"method":   method,
		"endpoint": endpoint,
	}
	start := time.Now()
	body, err := makeAPIRequest(method, endpoint, nil)
	latency := time.Since(start).Milliseconds()
	out["latency_ms"] = latency

	if err != nil {
		// Try to surface APIError fields structured.
		out["error"] = err.Error()
		if apiErr, ok := err.(*APIError); ok {
			out["status"] = apiErr.StatusCode
			brief := apiErr.Body
			if len(brief) > 300 {
				brief = brief[:300] + "..."
			}
			out["body_excerpt"] = brief
		}
		return out
	}

	out["status"] = 200
	brief := string(body)
	if len(brief) > 300 {
		brief = brief[:300] + "..."
	}
	out["body_excerpt"] = brief
	return out
}

// nameResolverCacheSnapshot returns a redacted view of the cache for debug.
func nameResolverCacheSnapshot() map[string]interface{} {
	nameCacheMu.RLock()
	defer nameCacheMu.RUnlock()
	out := map[string]interface{}{}
	for kind, entry := range nameCache {
		out[kind] = map[string]interface{}{
			"count":      len(entry.candidates),
			"fetched_at": entry.fetchedAt.Format(time.RFC3339),
			"age_ms":     time.Since(entry.fetchedAt).Milliseconds(),
		}
	}
	return out
}

// gatherDebugInfo is the workhorse: builds a single structured map with
// every diagnostic worth pasting back into a support thread. Safe to call
// from a chat (no secrets emitted) and from --debug on the CLI.
func gatherDebugInfo() map[string]interface{} {
	exe, _ := os.Executable()
	host, _ := os.Hostname()
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	cfg := map[string]interface{}{
		"tools_mode":     config.ToolsMode,
		"base_url":       config.BaseURL,
		"disabled_tools": config.DisabledTools,
		"api_key":        maskedAPIKey(),
		"api_key_set":    apiKey != "",
	}

	server := map[string]interface{}{
		"name":       ServerName,
		"version":    Version,
		"go_version": runtime.Version(),
		"goos":       runtime.GOOS,
		"goarch":     runtime.GOARCH,
		"executable": exe,
		"hostname":   host,
		"pid":        os.Getpid(),
	}

	runtimeInfo := map[string]interface{}{
		"uptime_sec":  int64(time.Since(startedAt).Seconds()),
		"goroutines":  runtime.NumGoroutine(),
		"cpus":        runtime.NumCPU(),
		"heap_mb":     memStats.HeapAlloc / 1024 / 1024,
		"sys_mb":      memStats.Sys / 1024 / 1024,
	}

	probes := map[string]interface{}{}
	if apiKey != "" {
		probes["account"] = endpointProbe("GET", "/v1/account?limit=1")
		probes["client"] = endpointProbe("GET", "/v1/client?limit=1")
		probes["device"] = endpointProbe("GET", "/v1/device?limit=1")
		probes["agent"] = endpointProbe("GET", "/v1/agent?limit=1")
	} else {
		probes["note"] = "No API token configured; skipping authenticated probes."
	}

	return map[string]interface{}{
		"server":              server,
		"runtime":             runtimeInfo,
		"config":              cfg,
		"environment":         envSnapshot(),
		"dns":                 dnsProbe(),
		"tls":                 tlsProbe(),
		"api_probes":          probes,
		"name_resolver_cache": nameResolverCacheSnapshot(),
		"recent_logs":         logCapture.Snapshot(),
		"generated_at":        time.Now().UTC().Format(time.RFC3339),
		"hint":                "Paste this whole payload into a support thread if you're seeing errors. The api_key field is masked; no other secrets are emitted.",
	}
}

// runDebug is the --debug CLI subcommand. Pretty-prints the debug bundle
// to stdout as JSON and exits 0.
func runDebug() {
	info := gatherDebugInfo()
	out, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal debug info: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
