# Slide MCP Web Server

A multi-user HTTP web server implementation of the Slide MCP (Model Context Protocol) server that allows users to access Slide API functionality through web requests with API key authentication.

## Features

- **Multi-user Support**: Multiple users can access the server simultaneously with their own API keys
- **HTTP/HTTPS Access**: Standard web protocols instead of stdin/stdout
- **API Key Authentication**: Secure access using Slide API keys
- **CORS Support**: Cross-origin requests enabled for web applications
- **Full Slide API Coverage**: All Slide API endpoints available as MCP tools
- **Systemd Service**: Runs as a system service with automatic restart

## Installation

The server is installed at `/var/www/www.slide.recipes/mcp/` and runs as a systemd service.

### Service Management

```bash
# Check status
sudo systemctl status slide-mcp-webserver

# Start service
sudo systemctl start slide-mcp-webserver

# Stop service
sudo systemctl stop slide-mcp-webserver

# Restart service
sudo systemctl restart slide-mcp-webserver

# View logs
sudo journalctl -u slide-mcp-webserver -f
```

## Usage

### Endpoints

- **Health Check**: `GET /health`
- **MCP Protocol**: `POST /mcp`

### Authentication

The server requires a valid Slide API key provided via one of these methods:

1. **Authorization Header**: `Authorization: Bearer YOUR_API_KEY`
2. **X-API-Key Header**: `X-API-Key: YOUR_API_KEY`
3. **Query Parameter**: `?api_key=YOUR_API_KEY`

### Examples

#### Initialize MCP Session

```bash
curl -X POST https://www.slide.recipes/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_SLIDE_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {}
  }'
```

#### List Available Tools

```bash
curl -X POST https://www.slide.recipes/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_SLIDE_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
```

#### Call a Tool (List Devices)

```bash
curl -X POST https://www.slide.recipes/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_SLIDE_API_KEY" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "slide_list_devices",
      "arguments": {
        "limit": 10
      }
    }
  }'
```

## Available Tools

The server provides access to all Slide API functionality through MCP tools:

### Devices
- `slide_list_devices` - List all devices with pagination and filtering

### Agents
- `slide_list_agents` - List all agents with pagination and filtering
- `slide_get_agent` - Get detailed information about a specific agent
- `slide_create_agent` - Create an agent for auto-pair installation
- `slide_pair_agent` - Pair an agent with a device using a pair code
- `slide_update_agent` - Update an agent's properties

### Backups
- `slide_list_backups` - List all backups with pagination and filtering
- `slide_get_backup` - Get detailed information about a specific backup
- `slide_start_backup` - Start a backup for a specific agent

### Snapshots
- `slide_list_snapshots` - List all snapshots with pagination and filtering
- `slide_get_snapshot` - Get detailed information about a specific snapshot

### File Restores
- `slide_list_file_restores` - List all file restores
- `slide_get_file_restore` - Get detailed information about a specific file restore
- `slide_create_file_restore` - Create a file restore from a snapshot
- `slide_delete_file_restore` - Delete a file restore
- `slide_browse_file_restore` - Browse files in a file restore

### Image Exports
- `slide_list_image_exports` - List all image exports
- `slide_get_image_export` - Get detailed information about a specific image export
- `slide_create_image_export` - Create an image export from a snapshot
- `slide_delete_image_export` - Delete an image export
- `slide_browse_image_export` - Browse images in an image export

### Virtual Machines
- `slide_list_virtual_machines` - List all virtual machines
- `slide_get_virtual_machine` - Get detailed information about a specific virtual machine
- `slide_create_virtual_machine` - Create a virtual machine from a snapshot
- `slide_update_virtual_machine` - Update a virtual machine's properties
- `slide_delete_virtual_machine` - Delete a virtual machine

### Users
- `slide_list_users` - List all users
- `slide_get_user` - Get detailed information about a specific user

### Alerts
- `slide_list_alerts` - List all alerts
- `slide_get_alert` - Get detailed information about a specific alert
- `slide_update_alert` - Update an alert's status (resolve/unresolve)

### Accounts
- `slide_list_accounts` - List all accounts
- `slide_get_account` - Get detailed information about a specific account
- `slide_update_account` - Update an account's properties

### Clients
- `slide_list_clients` - List all clients
- `slide_get_client` - Get detailed information about a specific client
- `slide_create_client` - Create a new client
- `slide_update_client` - Update a client's properties
- `slide_delete_client` - Delete a client

## Configuration

The server runs on port 8080 by default. This can be changed by setting the `PORT` environment variable in the systemd service file.

## Security

- All requests require valid Slide API keys
- CORS is enabled for web application access
- The server runs as the `www-data` user for security
- API keys are passed through to the Slide API for authentication

## Building from Source

```bash
cd webserver
go build -o slide-mcp-webserver .
```

## Documentation

For complete Slide API documentation, visit: https://api.slide.tech/docs 