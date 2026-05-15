package main

// Doctor: startup validation + the `--doctor` self-check subcommand.
//
// Two entry points share the same probe set:
//
//   - runStartupValidation()  - called from main() before runStdioServer.
//     Fails fast on 401/403 (clearly broken token); logs+continues on
//     network errors so the server stays alive long enough that
//     slide_help operation=troubleshoot is still callable.
//
//   - runDoctor()             - called when the user passes --doctor.
//     Runs every probe, prints a checklist, exits non-zero on any
//     failure. Idempotent and CI-friendly.
//
// The probe set:
//   1. Authentication: a HEAD-ish GET on /v1/account succeeds with 200.
//   2. Per-endpoint sample reads: /v1/client, /v1/device, /v1/agent.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// probeAccount calls /v1/account with the current token and returns
// either (account_count, nil) on 200 or an APIError on non-2xx.
// The count comes from the response data length, not pagination.total,
// because the Slide API omits pagination.total for single-item responses.
func probeAccount() (int, error) {
	body, err := makeAPIRequest("GET", "/v1/account?limit=1", nil)
	if err != nil {
		return 0, err
	}
	var p PaginatedResponse[Account]
	if err := json.Unmarshal(body, &p); err != nil {
		return 0, fmt.Errorf("parse account response: %w", err)
	}
	if p.Pagination.Total > 0 {
		return p.Pagination.Total, nil
	}
	return len(p.Data), nil
}

// accountSummary returns a short "as <name>" string for the startup log
// from the same /v1/account probe. Empty string if no account is visible.
func accountSummary() string {
	body, err := makeAPIRequest("GET", "/v1/account?limit=1", nil)
	if err != nil {
		return ""
	}
	var p PaginatedResponse[Account]
	if err := json.Unmarshal(body, &p); err != nil {
		return ""
	}
	if len(p.Data) == 0 {
		return ""
	}
	a := p.Data[0]
	if a.AccountName != "" {
		return a.AccountName
	}
	return a.AccountID
}

// probeCount runs `GET <endpoint>?limit=1` and returns the total count
// from the pagination envelope, or an error.
func probeCount(endpoint string) (int, error) {
	body, err := makeAPIRequest("GET", endpoint+"?limit=1", nil)
	if err != nil {
		return 0, err
	}
	var p struct {
		Pagination Pagination `json:"pagination"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return 0, fmt.Errorf("parse %s response: %w", endpoint, err)
	}
	return p.Pagination.Total, nil
}

// authErrorHint returns a novice-friendly multi-line message for 401/403
// responses; the empty string otherwise. Centralised so startup
// validation and --doctor share the wording.
func authErrorHint(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return ""
	}
	if apiErr.StatusCode != http.StatusUnauthorized && apiErr.StatusCode != http.StatusForbidden {
		return ""
	}
	return strings.TrimSpace(`
Your Slide API token was rejected by api.slide.tech (HTTP ` + fmt.Sprintf("%d", apiErr.StatusCode) + `).

To fix:
  1. Open https://console.slide.tech and sign in.
  2. Navigate to My Settings -> API Tokens.
  3. Generate a fresh token (or revoke + reissue the existing one).
  4. In Claude Desktop: Settings -> Extensions -> Slide Backup -> paste
     the new token and restart the extension.
  5. From a shell: re-run with --api-key <NEW_TOKEN> or set
     SLIDE_API_KEY in your environment.

The token gives full access to your Slide account; it is stored by
Claude Desktop and only forwarded to slide-mcp-server at launch.`) + "\n"
}

// networkErrorHint detects "no internet / firewall" style failures so
// the operator gets useful guidance rather than a Go stack-trace-style
// "dial tcp: i/o timeout" line.
func networkErrorHint(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if strings.Contains(s, "no such host") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "no route to host") ||
		strings.Contains(s, "TLS handshake") {
		return strings.TrimSpace(`
Could not reach api.slide.tech. The slide-mcp-server tried to call the
Slide API and the connection failed before any HTTPS response. Likely
causes:

  - No internet on the machine running Claude Desktop.
  - A corporate firewall / zero-trust agent is blocking outbound HTTPS
    to api.slide.tech. Ask your network admin to allow that host.
  - A captive portal is intercepting requests (hotel/airport wifi).
  - DNS isn't resolving api.slide.tech.

You can verify connectivity with:
  curl -sSI https://api.slide.tech/v1/account -H "Authorization: Bearer $SLIDE_API_KEY"`) + "\n"
	}
	return ""
}

// runStartupValidation is the lightweight, single-probe check that runs
// alongside runStdioServer(). It NEVER crashes the process on ANY
// failure - including 401/403 - because that would kill the MCP
// transport from Claude Desktop's perspective, which manifests as
// "Failed to call tool" with no useful diagnostic for the operator.
//
// Instead we log a clear remediation message to stderr (which surfaces
// in Claude Desktop's extension log panel) and let the server come up.
// Each subsequent tool call surfaces the same friendly auth error via
// the per-status APIError hints, and slide_help operation=troubleshoot
// remains callable so the LLM can guide the user to a fix.
func runStartupValidation() {
	total, err := probeAccount()
	if err == nil {
		summary := accountSummary()
		if summary != "" {
			stderrLogger.Printf("slide-mcp-server: connected to Slide as %s (%d account(s) visible)\n", summary, total)
		} else {
			stderrLogger.Printf("slide-mcp-server: connected to Slide successfully (%d account(s) visible)\n", total)
		}
		return
	}

	if hint := authErrorHint(err); hint != "" {
		stderrLogger.Println("")
		stderrLogger.Println("slide-mcp-server: WARNING - Slide API token was rejected at startup.")
		stderrLogger.Println(hint)
		stderrLogger.Println("The server is still running so slide_help (and any cached responses) remain callable, but every API-backed tool will return an auth error until the token is fixed.")
		stderrLogger.Println("Underlying error:", err)
		stderrLogger.Println("Run `slide_help operation=debug` from a chat (or `slide-mcp-server --debug` from a shell) for full diagnostics.")
		return
	}

	if hint := networkErrorHint(err); hint != "" {
		stderrLogger.Println("")
		stderrLogger.Println("slide-mcp-server: WARNING - could not reach Slide API at startup.")
		stderrLogger.Println(hint)
		stderrLogger.Println("The server will start anyway; slide_help operation=troubleshoot remains callable. Restart once connectivity is restored.")
		return
	}

	stderrLogger.Println("slide-mcp-server: WARNING - startup probe of /v1/account failed:", err)
	stderrLogger.Println("The server will start anyway; tool calls will surface the underlying error.")
}

// doctorCheck is one row in the doctor checklist.
type doctorCheck struct {
	Name   string
	Status string // "OK" | "FAIL" | "WARN"
	Detail string
}

// runDoctor is the --doctor subcommand. Probes everything that
// matters, prints a clean checklist, exits non-zero on any FAIL.
// Reusable from CI (--doctor --json would be a future addition).
func runDoctor() {
	fmt.Printf("slide-mcp-server v%s -- doctor\n", Version)
	fmt.Printf("Base URL: %s\n", APIBaseURL)
	fmt.Printf("Tools mode: %s\n", config.ToolsMode)
	fmt.Println(strings.Repeat("-", 60))

	checks := []doctorCheck{}
	addCheck := func(name, status, detail string) {
		checks = append(checks, doctorCheck{Name: name, Status: status, Detail: detail})
	}

	if apiKey == "" {
		addCheck("API token configured", "FAIL", "no token provided via --api-key or SLIDE_API_KEY")
	} else {
		shown := apiKey
		if len(shown) > 8 {
			shown = shown[:4] + "..." + shown[len(shown)-2:]
		}
		addCheck("API token configured", "OK", fmt.Sprintf("token=%s", shown))
	}

	// Connectivity probe (DNS + TLS to api.slide.tech). Cheap GET with
	// no auth payload required - we just need *some* response.
	netStart := time.Now()
	netCheck := doctorCheck{Name: "Network reachability (api.slide.tech)"}
	resp, netErr := http.Get(APIBaseURL + "/")
	netLatency := time.Since(netStart)
	if netErr != nil {
		netCheck.Status = "FAIL"
		netCheck.Detail = netErr.Error()
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		netCheck.Status = "OK"
		netCheck.Detail = fmt.Sprintf("HTTP %d in %s", resp.StatusCode, netLatency.Round(time.Millisecond))
	}
	checks = append(checks, netCheck)

	// Auth probe (only run if token + connectivity look OK).
	if apiKey != "" && netErr == nil {
		_, err := probeAccount()
		switch {
		case err == nil:
			addCheck("Authentication (/v1/account)", "OK", "200 - token accepted")
		default:
			detail := err.Error()
			if hint := authErrorHint(err); hint != "" {
				detail = "401/403 - token rejected (run --doctor for full guidance)"
			}
			addCheck("Authentication (/v1/account)", "FAIL", detail)
		}
	}

	// Per-endpoint sample reads. Skipped if auth failed.
	if apiKey != "" && netErr == nil && hasOK(checks, "Authentication (/v1/account)") {
		for _, ep := range []struct {
			name, path string
		}{
			{"Clients endpoint", "/v1/client"},
			{"Devices endpoint", "/v1/device"},
			{"Agents endpoint", "/v1/agent"},
		} {
			total, err := probeCount(ep.path)
			if err != nil {
				addCheck(ep.name, "FAIL", err.Error())
			} else {
				addCheck(ep.name, "OK", fmt.Sprintf("total=%d", total))
			}
		}
	}

	// Print results.
	failed := 0
	for _, c := range checks {
		mark := "OK   "
		switch c.Status {
		case "FAIL":
			mark = "FAIL "
			failed++
		case "WARN":
			mark = "WARN "
		}
		if c.Detail != "" {
			fmt.Printf("[%s] %-40s %s\n", mark, c.Name, c.Detail)
		} else {
			fmt.Printf("[%s] %s\n", mark, c.Name)
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	if failed == 0 {
		fmt.Println("All checks passed.")
		os.Exit(0)
	}

	// Print remediation hints for the first failure that has one.
	for _, c := range checks {
		if c.Status != "FAIL" {
			continue
		}
		if strings.HasPrefix(c.Name, "Authentication") {
			fmt.Println("")
			fmt.Println(strings.TrimSpace(`
Authentication failed. To fix:
  1. Open https://console.slide.tech and sign in.
  2. My Settings -> API Tokens -> generate a fresh token.
  3. Re-run with --api-key <NEW_TOKEN> or update SLIDE_API_KEY.`))
			break
		}
		if strings.HasPrefix(c.Name, "Network") {
			fmt.Println("")
			fmt.Println(strings.TrimSpace(`
Network probe failed. Check internet connectivity, then verify that
outbound HTTPS to api.slide.tech is allowed by any firewall or
zero-trust agent on this machine.`))
			break
		}
	}

	fmt.Printf("\n%d check(s) failed.\n", failed)
	os.Exit(1)
}

// hasOK is a tiny helper used by runDoctor to gate downstream probes.
func hasOK(checks []doctorCheck, name string) bool {
	for _, c := range checks {
		if c.Name == name && c.Status == "OK" {
			return true
		}
	}
	return false
}
