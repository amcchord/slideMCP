package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func useTestHTTPServer(t *testing.T, srv *httptest.Server) {
	t.Helper()
	oldBaseURL, oldClient, oldSleep, oldKey := APIBaseURL, httpClient, retrySleep, apiKey
	APIBaseURL = srv.URL
	httpClient = srv.Client()
	apiKey = "tk_test_reliability"
	retrySleep = func(context.Context, time.Duration) error { return nil }
	t.Cleanup(func() {
		APIBaseURL, httpClient, retrySleep, apiKey = oldBaseURL, oldClient, oldSleep, oldKey
	})
}

func TestAPIRequestRetriesOnlyIdempotentMethods(t *testing.T) {
	var getCalls, postCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/read":
			if getCalls.Add(1) == 1 {
				http.Error(w, "temporary", http.StatusServiceUnavailable)
				return
			}
			io.WriteString(w, `{"ok":true}`)
		case "/write":
			postCalls.Add(1)
			http.Error(w, "temporary", http.StatusServiceUnavailable)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	useTestHTTPServer(t, srv)

	if _, err := makeAPIRequest(http.MethodGet, "/read", nil); err != nil {
		t.Fatalf("GET should recover after one transient response: %v", err)
	}
	if got := getCalls.Load(); got != 2 {
		t.Fatalf("GET calls = %d, want 2", got)
	}
	if _, err := makeAPIRequest(http.MethodPost, "/write", []byte(`{}`)); err == nil {
		t.Fatal("POST should return the first transient error")
	}
	if got := postCalls.Load(); got != 1 {
		t.Fatalf("POST calls = %d, want exactly 1 to avoid duplicate mutations", got)
	}
}

func TestAPIRequestCapsRetryAfter(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "3600")
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()
	useTestHTTPServer(t, srv)

	var waited time.Duration
	retrySleep = func(_ context.Context, d time.Duration) error {
		waited = d
		return nil
	}
	if _, err := makeAPIRequest(http.MethodGet, "/rate-limited", nil); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if waited != maxRetryAfter {
		t.Fatalf("retry wait = %v, want capped %v", waited, maxRetryAfter)
	}
}

func TestAPIRequestHonorsContextDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	useTestHTTPServer(t, srv)
	retrySleep = sleepWithContext

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err := makeAPIRequestContext(ctx, http.MethodGet, "/slow", nil)
	if err == nil {
		t.Fatal("expected deadline error")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("deadline took %v to propagate", elapsed)
	}
}

func TestAPIRequestRejectsOversizedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, strings.NewReader(strings.Repeat("x", maxAPIResponseBytes+1)))
	}))
	defer srv.Close()
	useTestHTTPServer(t, srv)

	_, err := makeAPIRequest(http.MethodGet, "/oversized", nil)
	if err == nil || !strings.Contains(err.Error(), "response exceeded") {
		t.Fatalf("expected bounded-response error, got %v", err)
	}
}

func TestAPIErrorRedactsToken(t *testing.T) {
	oldKey := apiKey
	apiKey = "tk_should_never_escape"
	t.Cleanup(func() { apiKey = oldKey })
	err := (&APIError{
		Method:     http.MethodGet,
		Endpoint:   "/v1/account",
		StatusCode: http.StatusUnauthorized,
		Body:       `upstream echoed tk_should_never_escape`,
	}).Error()
	if strings.Contains(err, apiKey) || !strings.Contains(err, "[REDACTED]") {
		t.Fatalf("API error was not redacted: %s", err)
	}
}

func TestFetchAllPaginated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "50" {
			t.Errorf("limit = %q, want 50", got)
		}
		switch r.URL.Query().Get("offset") {
		case "0":
			io.WriteString(w, `{"data":[{"client_id":"c_1","name":"One"},{"client_id":"c_2","name":"Two"}],"pagination":{"total":3,"next_offset":2}}`)
		case "2":
			io.WriteString(w, `{"data":[{"client_id":"c_3","name":"Three"}],"pagination":{"total":3}}`)
		default:
			http.Error(w, "unexpected offset", http.StatusBadRequest)
		}
	}))
	defer srv.Close()
	useTestHTTPServer(t, srv)

	clients, err := fetchAllPaginated[Client]("/v1/client?limit=1")
	if err != nil {
		t.Fatalf("fetchAllPaginated: %v", err)
	}
	if len(clients) != 3 || clients[2].Name != "Three" {
		t.Fatalf("unexpected clients: %+v", clients)
	}
}

func TestFetchAllPaginatedRejectsLoop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, `{"data":[],"pagination":{"next_offset":0}}`)
	}))
	defer srv.Close()
	useTestHTTPServer(t, srv)

	_, err := fetchAllPaginated[Client]("/v1/client")
	if err == nil || !strings.Contains(err.Error(), "non-advancing") {
		t.Fatalf("expected pagination-loop error, got %v", err)
	}
}

func TestInventoryUsesThreeBulkReadsAndPreservesOrphans(t *testing.T) {
	var mu sync.Mutex
	requests := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/client":
			io.WriteString(w, `{"data":[{"client_id":"c_one","name":"One"}],"pagination":{}}`)
		case "/v1/device":
			io.WriteString(w, `{"data":[{"device_id":"d_one","client_id":"c_one","hostname":"box-1"},{"device_id":"d_two","hostname":"box-2"}],"pagination":{}}`)
		case "/v1/agent":
			io.WriteString(w, `{"data":[{"agent_id":"a_one","device_id":"d_one","hostname":"host-1"},{"agent_id":"a_orphan","device_id":"d_missing","hostname":"lost"}],"pagination":{}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	useTestHTTPServer(t, srv)

	out, err := listAllClientsDevicesAndAgents(nil)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 3 || requests["/v1/client"] != 1 || requests["/v1/device"] != 1 || requests["/v1/agent"] != 1 {
		t.Fatalf("inventory requests = %v, want one bulk request per entity type", requests)
	}
	var parsed struct {
		Clients []struct {
			Name    string `json:"name"`
			Devices []struct {
				DeviceID string  `json:"device_id"`
				Agents   []Agent `json:"agents"`
			} `json:"devices"`
		} `json:"clients"`
		OrphanAgents []Agent `json:"orphan_agents"`
		Metadata     struct {
			Complete bool `json:"complete"`
		} `json:"_metadata"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("parse inventory: %v", err)
	}
	if !parsed.Metadata.Complete || len(parsed.Clients) != 2 || len(parsed.OrphanAgents) != 1 {
		t.Fatalf("inventory lost entities: %+v", parsed)
	}
	if parsed.Clients[1].Name != "One" || len(parsed.Clients[1].Devices) != 1 || len(parsed.Clients[1].Devices[0].Agents) != 1 {
		t.Fatalf("client hierarchy is incorrect: %+v", parsed.Clients)
	}
}

func TestValidateBaseURL(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"https://api.slide.tech/", false},
		{"http://127.0.0.1:8080", false},
		{"http://localhost:8080", false},
		{"http://api.slide.tech", true},
		{"https://user:pass@api.slide.tech", true},
		{"https://api.slide.tech?token=x", true},
		{"not-a-url", true},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			cfg := NewServerConfig()
			cfg.BaseURL = tc.value
			err := cfg.ValidateBaseURL()
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateBaseURL(%q) error = %v, wantErr=%v", tc.value, err, tc.wantErr)
			}
			if err == nil && strings.HasSuffix(cfg.BaseURL, "/") {
				t.Fatalf("base URL was not normalized: %q", cfg.BaseURL)
			}
		})
	}
}

func ExampleAPIError() {
	err := &APIError{Method: http.MethodGet, Endpoint: "/v1/device", StatusCode: http.StatusServiceUnavailable, Body: "maintenance"}
	fmt.Println(strings.Contains(err.Error(), "status.slide.tech"))
	// Output: true
}
