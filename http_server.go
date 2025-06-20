package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"
)

// Session represents an active client session
type Session struct {
	ID        string
	CreatedAt time.Time
	LastUsed  time.Time
}

// SessionManager handles session creation and validation
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// generateSessionID creates a random session ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		httpLogger.Printf("ERROR: Failed to generate random bytes for session ID: %v", err)
		// Fallback to timestamp
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	sessionID := hex.EncodeToString(bytes)
	httpLogger.Printf("DEBUG: Generated new session ID: %s", sessionID)
	return sessionID
}

// CreateSession creates a new session and returns its ID
func (sm *SessionManager) CreateSession() string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionID := generateSessionID()
	sm.sessions[sessionID] = &Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}
	httpLogger.Printf("INFO: Created new session: %s", sessionID)

	return sessionID
}

// ValidateSession checks if a session is valid and updates its last used time
func (sm *SessionManager) ValidateSession(sessionID string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		httpLogger.Printf("WARN: Session validation failed for ID: %s", sessionID)
		return false
	}

	// Update last used time
	session.LastUsed = time.Now()
	httpLogger.Printf("DEBUG: Session validated: %s, last used: %s", sessionID, session.LastUsed)
	return true
}

// SSEClient represents a client connected via Server-Sent Events
type SSEClient struct {
	ResponseWriter http.ResponseWriter
	Flusher        http.Flusher
	LastEventID    string
	Done           chan bool
}

// HTTPServer handles HTTP MCP requests
type HTTPServer struct {
	sessionManager  *SessionManager
	sseClients      map[string]*SSEClient
	sseClientsMutex sync.RWMutex
}

// NewHTTPServer creates a new HTTP MCP server
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		sessionManager: NewSessionManager(),
		sseClients:     make(map[string]*SSEClient),
	}
}

// StartHTTPServer starts the HTTP server on the specified port
func StartHTTPServer(port int) {
	server := NewHTTPServer()

	// Set up route handlers
	http.HandleFunc("/mcp", server.loggingMiddleware(server.handleMCPRequest))

	// Add a health check endpoint for debugging
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	httpLogger.Printf("INFO: Starting HTTP MCP server on port %d...", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		httpLogger.Fatalf("ERROR: Failed to start HTTP server: %v", err)
	}
}

// loggingMiddleware wraps an HTTP handler with request/response logging
func (s *HTTPServer) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the incoming request details
		httpLogger.Printf("INFO: %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		httpLogger.Printf("DEBUG: Request headers: %v", r.Header)

		// Log the request body if available
		if r.Body != nil && r.Method == http.MethodPost {
			// Clone the body so we can still read it later
			bodyBytes, err := httputil.DumpRequest(r, true)
			if err == nil {
				httpLogger.Printf("DEBUG: Request body: %s", string(bodyBytes))
			}
		}

		// Create a response wrapper to capture the response
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           new(bytes.Buffer),
		}

		// Call the wrapped handler
		next(lrw, r)

		// Log the response
		duration := time.Since(start)
		httpLogger.Printf("INFO: %s %s - %d %s - %v", r.Method, r.URL.Path, lrw.statusCode,
			http.StatusText(lrw.statusCode), duration)

		// Don't log response body for streaming responses
		if r.Method != http.MethodGet || r.Header.Get("Accept") != "text/event-stream" {
			responseBody := lrw.body.String()
			if len(responseBody) > 1000 {
				responseBody = responseBody[:1000] + "... [truncated]"
			}
			httpLogger.Printf("DEBUG: Response body: %s", responseBody)
		}
	}
}

// loggingResponseWriter is a wrapper for http.ResponseWriter that captures the status code and response body
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

// handleMCPRequest handles incoming MCP requests
func (s *HTTPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	requestID := generateRequestID()
	httpLogger.Printf("INFO [ReqID:%s]: Processing MCP request from %s", requestID, r.RemoteAddr)

	// Check HTTP method
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		httpLogger.Printf("ERROR [ReqID:%s]: Unsupported method: %s", requestID, r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// GET method is only for SSE streaming
	if r.Method == http.MethodGet {
		// Check if client wants SSE
		acceptHeader := r.Header.Get("Accept")
		if acceptHeader != "text/event-stream" {
			httpLogger.Printf("ERROR [ReqID:%s]: Invalid Accept header for GET: %s", requestID, acceptHeader)
			http.Error(w, "Unsupported Accept header for GET", http.StatusNotAcceptable)
			return
		}

		// Check session
		sessionID := r.Header.Get("Mcp-Session-Id")
		httpLogger.Printf("DEBUG [ReqID:%s]: SSE connection attempt with session ID: %s", requestID, sessionID)

		if sessionID == "" || !s.sessionManager.ValidateSession(sessionID) {
			httpLogger.Printf("ERROR [ReqID:%s]: Invalid or missing session for SSE: %s", requestID, sessionID)
			http.Error(w, "Invalid or missing session", http.StatusUnauthorized)
			return
		}

		httpLogger.Printf("INFO [ReqID:%s]: Starting SSE handler for session: %s", requestID, sessionID)
		s.handleSSE(w, r, sessionID)
		return
	}

	// Process POST request
	var sessionID string
	// Check if there's an existing session
	sessionID = r.Header.Get("Mcp-Session-Id")
	httpLogger.Printf("DEBUG [ReqID:%s]: Processing POST with session ID: %s", requestID, sessionID)

	if sessionID == "" || !s.sessionManager.ValidateSession(sessionID) {
		// Create a new session
		sessionID = s.sessionManager.CreateSession()
		httpLogger.Printf("INFO [ReqID:%s]: Created new session: %s", requestID, sessionID)
	} else {
		httpLogger.Printf("DEBUG [ReqID:%s]: Using existing session: %s", requestID, sessionID)
	}

	// Set session ID in response header
	w.Header().Set("Mcp-Session-Id", sessionID)

	// Content negotiation
	acceptHeader := r.Header.Get("Accept")
	// Default to JSON if not specified
	if acceptHeader == "" {
		acceptHeader = "application/json"
	}
	httpLogger.Printf("DEBUG [ReqID:%s]: Content negotiation - Accept: %s", requestID, acceptHeader)

	// Parse the request body
	var request MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpLogger.Printf("ERROR [ReqID:%s]: Failed to parse request body: %v", requestID, err)

		// Send error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		errorResponse := sendError(nil, -32700, "Parse error", nil)
		err = json.NewEncoder(w).Encode(errorResponse)
		if err != nil {
			httpLogger.Printf("ERROR [ReqID:%s]: Failed to encode error response: %v", requestID, err)
		}
		return
	}

	// Log request details
	httpLogger.Printf("INFO [ReqID:%s]: Request Method: %s", requestID, request.Method)
	if request.ID != nil {
		httpLogger.Printf("INFO [ReqID:%s]: Request ID: %v", requestID, request.ID)
	}

	// Marshal params for logging
	paramsJSON, err := json.Marshal(request.Params)
	if err == nil {
		// Truncate large params
		paramsStr := string(paramsJSON)
		if len(paramsStr) > 1000 {
			paramsStr = paramsStr[:1000] + "... [truncated]"
		}
		httpLogger.Printf("DEBUG [ReqID:%s]: Request Params: %s", requestID, paramsStr)
	}

	// Check if this is a notification (no ID field)
	if request.ID == nil {
		httpLogger.Printf("INFO [ReqID:%s]: Processing notification: %s", requestID, request.Method)

		// This is a notification - handle it but don't send a response
		handleNotification(request)
		httpLogger.Printf("DEBUG [ReqID:%s]: Notification processed, no response required", requestID)

		// Return 202 Accepted with no body for notifications
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// This is a request - handle it and send a response
	httpLogger.Printf("INFO [ReqID:%s]: Processing request: %s with ID: %v", requestID, request.Method, request.ID)
	response := handleRequest(request)

	// Log the response
	responseJSON, err := json.Marshal(response)
	if err == nil {
		// Truncate large responses
		responseStr := string(responseJSON)
		if len(responseStr) > 1000 {
			responseStr = responseStr[:1000] + "... [truncated]"
		}
		httpLogger.Printf("DEBUG [ReqID:%s]: Response: %s", requestID, responseStr)
	}

	// Check if there was an error in the response
	if response.Error != nil {
		httpLogger.Printf("ERROR [ReqID:%s]: Error in response: %v", requestID, response.Error)
	}

	// Set content type header based on Accept header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		httpLogger.Printf("ERROR [ReqID:%s]: Failed to encode response: %v", requestID, err)
	} else {
		httpLogger.Printf("INFO [ReqID:%s]: Response sent successfully", requestID)
	}
}

// generateRequestID creates a unique ID for request tracking
func generateRequestID() string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req-%s", hex.EncodeToString(bytes))
}

// handleSSE handles Server-Sent Events connections
func (s *HTTPServer) handleSSE(w http.ResponseWriter, r *http.Request, sessionID string) {
	requestID := generateRequestID()
	httpLogger.Printf("INFO [ReqID:%s]: Setting up SSE connection for session %s", requestID, sessionID)

	// Check if the response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpLogger.Printf("ERROR [ReqID:%s]: Client doesn't support SSE (no Flusher)", requestID)
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	httpLogger.Printf("DEBUG [ReqID:%s]: SSE headers set", requestID)

	// Get the last event ID if provided
	lastEventID := r.Header.Get("Last-Event-ID")
	if lastEventID != "" {
		httpLogger.Printf("DEBUG [ReqID:%s]: Last-Event-ID: %s", requestID, lastEventID)
	}

	// Create a new SSE client
	client := &SSEClient{
		ResponseWriter: w,
		Flusher:        flusher,
		LastEventID:    lastEventID,
		Done:           make(chan bool),
	}

	// Register the client
	s.sseClientsMutex.Lock()
	s.sseClients[sessionID] = client
	clientCount := len(s.sseClients)
	s.sseClientsMutex.Unlock()
	httpLogger.Printf("INFO [ReqID:%s]: Client registered for SSE, total clients: %d", requestID, clientCount)

	// Remove client when done
	defer func() {
		s.sseClientsMutex.Lock()
		delete(s.sseClients, sessionID)
		remainingClients := len(s.sseClients)
		s.sseClientsMutex.Unlock()
		httpLogger.Printf("INFO [ReqID:%s]: Client disconnected from SSE, remaining clients: %d", requestID, remainingClients)
	}()

	// Send an initial comment to establish the connection
	connectedMsg := fmt.Sprintf(": connected at %s\n\n", time.Now().Format(time.RFC3339))
	_, err := fmt.Fprintf(w, connectedMsg)
	if err != nil {
		httpLogger.Printf("ERROR [ReqID:%s]: Failed to write initial SSE message: %v", requestID, err)
		return
	}
	flusher.Flush()
	httpLogger.Printf("DEBUG [ReqID:%s]: SSE connection established, sent initial message", requestID)

	// Start a heartbeat goroutine
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	go func() {
		heartbeatCount := 0
		for {
			select {
			case <-heartbeatTicker.C:
				// Send a heartbeat comment
				heartbeatTime := time.Now().Format(time.RFC3339)
				heartbeatMsg := fmt.Sprintf(": heartbeat %s\n\n", heartbeatTime)
				_, err := fmt.Fprintf(client.ResponseWriter, heartbeatMsg)

				heartbeatCount++
				if err != nil {
					httpLogger.Printf("ERROR [ReqID:%s]: Failed to send heartbeat #%d: %v",
						requestID, heartbeatCount, err)
					return
				}

				client.Flusher.Flush()
				httpLogger.Printf("DEBUG [ReqID:%s]: Sent SSE heartbeat #%d at %s",
					requestID, heartbeatCount, heartbeatTime)

			case <-client.Done:
				httpLogger.Printf("DEBUG [ReqID:%s]: Heartbeat goroutine terminating", requestID)
				return
			}
		}
	}()

	httpLogger.Printf("INFO [ReqID:%s]: SSE handler waiting for client disconnect", requestID)

	// Wait for client disconnect or context done
	<-r.Context().Done()
	httpLogger.Printf("INFO [ReqID:%s]: SSE connection context done, closing", requestID)
	close(client.Done)
}

// generateEventID creates a random event ID for SSE events
func generateEventID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		httpLogger.Printf("ERROR: Failed to generate random bytes for event ID: %v", err)
		// Fallback to timestamp
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	eventID := hex.EncodeToString(bytes)
	httpLogger.Printf("DEBUG: Generated event ID: %s", eventID)
	return eventID
}

// SendSSEEvent sends an event to a specific SSE client
func (s *HTTPServer) SendSSEEvent(sessionID, eventType, data string) {
	requestID := generateRequestID()
	httpLogger.Printf("INFO [ReqID:%s]: Sending SSE event to session %s, type: %s",
		requestID, sessionID, eventType)

	s.sseClientsMutex.RLock()
	client, exists := s.sseClients[sessionID]
	s.sseClientsMutex.RUnlock()

	if !exists {
		httpLogger.Printf("WARN [ReqID:%s]: Failed to send SSE event - client not found for session: %s",
			requestID, sessionID)
		return
	}

	// Generate a unique event ID
	eventID := generateEventID()

	// Prepare the message
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("id: %s\n", eventID))
	if eventType != "" {
		buffer.WriteString(fmt.Sprintf("event: %s\n", eventType))
	}

	// Log a truncated version of the data if it's large
	logData := data
	if len(logData) > 1000 {
		logData = logData[:1000] + "... [truncated]"
	}
	httpLogger.Printf("DEBUG [ReqID:%s]: SSE data: %s", requestID, logData)

	buffer.WriteString(fmt.Sprintf("data: %s\n\n", data))

	// Send the event
	_, err := fmt.Fprint(client.ResponseWriter, buffer.String())
	if err != nil {
		httpLogger.Printf("ERROR [ReqID:%s]: Failed to send SSE event: %v", requestID, err)
		return
	}

	client.Flusher.Flush()
	httpLogger.Printf("INFO [ReqID:%s]: SSE event sent successfully to session %s", requestID, sessionID)
}

// BroadcastSSEEvent sends an event to all connected SSE clients
func (s *HTTPServer) BroadcastSSEEvent(eventType, data string) {
	requestID := generateRequestID()

	s.sseClientsMutex.RLock()
	clientCount := len(s.sseClients)
	s.sseClientsMutex.RUnlock()

	httpLogger.Printf("INFO [ReqID:%s]: Broadcasting SSE event to %d clients, type: %s",
		requestID, clientCount, eventType)

	if clientCount == 0 {
		httpLogger.Printf("WARN [ReqID:%s]: No clients connected for broadcast", requestID)
		return
	}

	// Log a truncated version of the data if it's large
	logData := data
	if len(logData) > 1000 {
		logData = logData[:1000] + "... [truncated]"
	}
	httpLogger.Printf("DEBUG [ReqID:%s]: Broadcast SSE data: %s", requestID, logData)

	// Track successful and failed sends
	successCount := 0
	failCount := 0

	s.sseClientsMutex.RLock()
	defer s.sseClientsMutex.RUnlock()

	for sessionID, client := range s.sseClients {
		// Generate a unique event ID
		eventID := generateEventID()

		// Prepare the message
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("id: %s\n", eventID))
		if eventType != "" {
			buffer.WriteString(fmt.Sprintf("event: %s\n", eventType))
		}
		buffer.WriteString(fmt.Sprintf("data: %s\n\n", data))

		// Send the event
		_, err := fmt.Fprint(client.ResponseWriter, buffer.String())
		if err != nil {
			httpLogger.Printf("ERROR [ReqID:%s]: Failed to send broadcast to session %s: %v",
				requestID, sessionID, err)
			failCount++
			continue
		}

		client.Flusher.Flush()
		successCount++
	}

	httpLogger.Printf("INFO [ReqID:%s]: SSE broadcast complete - %d successful, %d failed",
		requestID, successCount, failCount)
}

// Initialize more detailed logger
var (
	httpLogger = log.New(os.Stdout, "[HTTP SERVER] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
)

// Custom HTTP handler with detailed logging
func loggingHandler(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request
	httpLogger.Printf("Incoming request: %s %s", r.Method, r.URL.Path)

	// Log request headers
	for name, values := range r.Header {
		for _, value := range values {
			httpLogger.Printf("Header: %s: %s", name, value)
		}
	}

	// Log the request body if it's a POST or PUT request
	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		// Dump the request body
		if body, err := httputil.DumpRequest(r, true); err == nil {
			httpLogger.Printf("Request body: %s", body)
		} else {
			httpLogger.Printf("Error dumping request body: %v", err)
		}
	}
}
