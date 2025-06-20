# Slide MCP HTTP Server Mode

The Slide MCP Server can run as an HTTP server, providing a RESTful API interface with Server-Sent Events (SSE) support for streaming responses.

For practical examples, see [HTTP_EXAMPLES.md](HTTP_EXAMPLES.md).

## Starting HTTP Server Mode

```bash
# Start with default port 8080
slide-mcp-server -http

# Specify custom port
slide-mcp-server -http -port 3000

# Use a custom API endpoint
slide-mcp-server -http -api-url="https://custom-api.example.com"

# Combine options
slide-mcp-server -http -port 3000 -api-url="https://custom-api.example.com"
```

## API Endpoints

The HTTP server exposes a single endpoint:

- `POST /mcp` - For all client-to-server requests and notifications
- `GET /mcp` - For Server-Sent Events (SSE) streaming connections

## HTTP Methods

- `POST` - Use for all client-to-server messages
- `GET` - Use only for SSE streaming connections

## Session Management

The server implements session management with the following characteristics:

1. A new session is automatically created on the first request
2. The session ID is returned in the `Mcp-Session-Id` header
3. Clients should include the `Mcp-Session-Id` header in subsequent requests
4. Sessions are validated on each request

Example:

```
POST /mcp HTTP/1.1
Content-Type: application/json

{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {}
}

HTTP/1.1 200 OK
Content-Type: application/json
Mcp-Session-Id: 123e4567e89b12d3a456426614174000

{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "slide-mcp-server",
      "version": "0.1.0"
    }
  }
}
```

## Content Negotiation

The server checks the `Accept` header to determine the response format:

- `application/json` (default) - JSON responses
- `text/event-stream` - Server-Sent Events stream (only valid with GET method)

## Notification Handling

Notifications (requests without an ID) receive a `202 Accepted` response with no body.

```
POST /mcp HTTP/1.1
Content-Type: application/json
Mcp-Session-Id: 123e4567e89b12d3a456426614174000

{
  "jsonrpc": "2.0",
  "method": "notifications/initialized"
}

HTTP/1.1 202 Accepted
```

## Server-Sent Events (SSE) Support

To use SSE streaming:

1. Establish a session using POST requests
2. Connect to the SSE stream using a GET request with the same session ID
3. Specify `Accept: text/event-stream` header

Example:

```
GET /mcp HTTP/1.1
Accept: text/event-stream
Mcp-Session-Id: 123e4567e89b12d3a456426614174000

HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

: connected

id: 6b86b273ff34fce1
data: {"jsonrpc":"2.0","method":"server/event","params":{"type":"status","message":"Connected"}}

: heartbeat 2025-06-19T10:15:30Z
```

### SSE Features

- Heartbeats every 30 seconds
- Event IDs for resuming connections
- Event typing for client-side filtering

## Security Considerations

- Always use HTTPS in production
- Consider adding authentication mechanisms
- Validate session IDs
- Set appropriate CORS headers for browser clients
