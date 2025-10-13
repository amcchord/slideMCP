# MCP Context Viewer

This script shows the initial context that the Slide MCP Server provides to LLMs during initialization.

## What It Shows

The MCP server provides rich context to LLMs at startup, including:

1. **Server Information**: Name, version, and protocol details
2. **Initial Context**: A complete hierarchical view of your Slide infrastructure:
   - All clients (organizations/groups)
   - All devices (Slide appliances) under each client
   - All agents (backup software) on each device
3. **Available Tools**: All tools the LLM can use to interact with your Slide system

## Usage

```bash
./scripts/show_mcp_context.sh <API_KEY> [PATH_TO_MCP_SERVER]
```

### Arguments

- `API_KEY` (required): Your Slide API key
- `PATH_TO_MCP_SERVER` (optional): Path to the slide-mcp-server binary

If you don't specify the path, the script will automatically search in common locations:
- Current directory (`./slide-mcp-server`)
- Parent directory (`../slide-mcp-server`)
- `/usr/local/bin/slide-mcp-server`
- `$HOME/.local/bin/slide-mcp-server`
- Recent release directory (`./release/v2.4.0/slide-mcp-server-*`)

### Examples

```bash
# Use with auto-detection of binary location
./scripts/show_mcp_context.sh your-api-key-here

# Specify exact path to binary
./scripts/show_mcp_context.sh your-api-key-here /usr/local/bin/slide-mcp-server

# Use local development binary
./scripts/show_mcp_context.sh your-api-key-here ./slide-mcp-server
```

## Requirements

- Bash shell
- Valid Slide API key
- slide-mcp-server binary

### Optional (for better formatting)

- `jq` - JSON processor for prettier output formatting
- `python3` - Fallback JSON formatting if jq is not available

## Output

The script provides a comprehensive view of what the LLM sees, including:

### Server Information
- Server name and version
- Protocol version
- Capabilities

### Infrastructure Overview
- Summary counts (total clients, devices, agents)
- Hierarchical tree view of your infrastructure
- Complete JSON data structure

### Available Tools
- List of all available tools
- Tool descriptions and capabilities
- Current tool mode restrictions (if any)

## Understanding the Context

This initial context is crucial for LLM performance because:

1. **Immediate Awareness**: The LLM knows your entire infrastructure without API calls
2. **Efficient Queries**: Can answer questions about counts, relationships, and status immediately  
3. **Smart Tool Usage**: Knows which tools are available and how to use them
4. **Context-Aware Responses**: Can provide specific, relevant answers about your environment

## Example Output

```
=================================
Slide MCP Server Initial Context Viewer
=================================
API Key: sk_test_1...
MCP Server: ./slide-mcp-server

--- Server Information ---
Name: slide-mcp-server
Version: 2.4.0

--- Initial Context Metadata ---
Description: Initial overview of all clients, devices, and agents loaded at startup for improved performance
Source Tool: list_all_clients_devices_and_agents
Usage Note: This data is also available via the list_all_clients_devices_and_agents tool and should be refreshed if needed

--- Clients, Devices, and Agents Overview ---
Total Clients: 3
Total Devices: 5
Total Agents: 12

Client: Acme Corp (client_123)
└── 2 device(s)
    └── Office Server (3 agent(s))
    └── Backup Appliance (2 agent(s))

Client: Tech Startup (client_456)
└── 1 device(s)
    └── Main Server (4 agent(s))
```

## Troubleshooting

1. **Binary not found**: Ensure the slide-mcp-server binary exists and is executable
2. **API key errors**: Verify your API key is valid and has appropriate permissions
3. **No response**: Check network connectivity and API endpoint availability
4. **JSON formatting issues**: Install `jq` for better output formatting

## Integration

This script can be used for:
- Debugging MCP context issues
- Understanding what information is available to LLMs
- Verifying infrastructure visibility
- API key testing
- Development and troubleshooting