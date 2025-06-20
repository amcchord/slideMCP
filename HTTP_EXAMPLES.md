# HTTP Server Mode Examples

These examples demonstrate how to interact with the Slide MCP Server in HTTP mode using curl.

## Basic Request

```bash
# Initialize request
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {}
  }'

# Response will include Mcp-Session-Id header
# Store this session ID for use in subsequent requests
```

## Using Session ID

```bash
# Use the session ID from the previous response
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: YOUR_SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
```

## Sending Notifications

```bash
# Send a notification (no ID field)
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: YOUR_SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "notifications/initialized"
  }'

# Will return 202 Accepted with no body
```

## Connect to SSE Stream

```bash
# Using curl to connect to SSE stream
curl -N -H "Accept: text/event-stream" \
  -H "Mcp-Session-Id: YOUR_SESSION_ID" \
  http://localhost:8080/mcp

# Will receive a stream of events
```

## Example JavaScript Client

```javascript
// Example of connecting to the SSE stream in a browser
const sessionId = 'YOUR_SESSION_ID'; // From initial POST request

// Connect to SSE stream
const eventSource = new EventSource(`http://localhost:8080/mcp?session=${sessionId}`, {
  headers: {
    'Mcp-Session-Id': sessionId
  }
});

// Handle incoming events
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received event:', data);
};

// Handle connection open
eventSource.onopen = () => {
  console.log('SSE connection established');
};

// Handle errors
eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
  eventSource.close();
};
```
