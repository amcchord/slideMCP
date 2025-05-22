# Slide MCP Server

An MCP server implementation that integrates with the Slide API, providing device and agent management capabilities.

## Features

- **Device Management**: List all devices with pagination and filtering options
- **Agent Management**: List, create, pair, and update agents
- **Backup Management**: List, retrieve, and initiate backups
- **Detailed Information**: Get comprehensive details about each device and agent including status, storage, and network information
- **Flexible Filtering**: Filter devices and agents by client ID, device ID, and other parameters
- **Pagination Support**: Control results per page with offset and limit parameters

## Tools

### Device Tools

- **slide_list_devices**
  - List all devices with pagination and filtering
  - Inputs:
    - `limit` (number, optional): Results per page (max 50)
    - `offset` (number, optional): Pagination offset
    - `client_id` (string, optional): Filter by client ID
    - `sort_asc` (boolean, optional): Sort in ascending order

### Agent Tools

- **slide_list_agents**
  - List all agents with pagination and filtering
  - Inputs:
    - `limit` (number, optional): Results per page (max 50)
    - `offset` (number, optional): Pagination offset
    - `device_id` (string, optional): Filter by device ID
    - `client_id` (string, optional): Filter by client ID
    - `sort_asc` (boolean, optional): Sort in ascending order
    - `sort_by` (string, optional): Sort by field (id, hostname, name)

- **slide_get_agent**
  - Get detailed information about a specific agent
  - Inputs:
    - `agent_id` (string, required): ID of the agent to retrieve

- **slide_create_agent**
  - Create an agent for auto-pair installation
  - Inputs:
    - `display_name` (string, required): Display name for the agent
    - `device_id` (string, required): ID of the device to associate with the agent

- **slide_pair_agent**
  - Pair an agent with a device using a pair code
  - Inputs:
    - `pair_code` (string, required): Pair code generated during agent creation
    - `device_id` (string, required): ID of the device to pair with

- **slide_update_agent**
  - Update an agent's properties
  - Inputs:
    - `agent_id` (string, required): ID of the agent to update
    - `display_name` (string, required): New display name for the agent

### Backup Tools

- **slide_list_backups**
  - List all backups with pagination and filtering
  - Inputs:
    - `limit` (number, optional): Results per page (max 50)
    - `offset` (number, optional): Pagination offset
    - `agent_id` (string, optional): Filter by agent ID
    - `device_id` (string, optional): Filter by device ID
    - `snapshot_id` (string, optional): Filter by snapshot ID
    - `sort_asc` (boolean, optional): Sort in ascending order
    - `sort_by` (string, optional): Sort by field (id, start_time)

- **slide_get_backup**
  - Get detailed information about a specific backup
  - Inputs:
    - `backup_id` (string, required): ID of the backup to retrieve

- **slide_start_backup**
  - Start a backup for a specific agent
  - Inputs:
    - `agent_id` (string, required): ID of the agent to backup

## Configuration

### Getting an API Key

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to your account settings
3. Generate your API key from the API section

### Usage with Claude Desktop

Add this to your `claude_desktop_config.json`:

### Docker

```json
{
  "mcpServers": {
    "slide": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "SLIDE_API_KEY",
        "mcp/slide"
      ],
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE"
      }
    }
  }
}
```

### NPX

```json
{
  "mcpServers": {
    "slide": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-slide"
      ],
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE"
      }
    }
  }
}
```

### Usage with VS Code

For quick installation, use the one-click installation buttons below...

[![Install with NPX in VS Code](https://img.shields.io/badge/VS_Code-NPM-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=slide&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22apiKey%22%7D%5D&config=%7B%22command%22%3A%22npx%22%2C%22args%22%3A%5B%22-y%22%2C%22%40modelcontextprotocol%2Fserver-slide%22%5D%2C%22env%22%3A%7B%22SLIDE_API_KEY%22%3A%22%24%7Binput%3Aslide_api_key%7D%22%7D%7D) [![Install with NPX in VS Code Insiders](https://img.shields.io/badge/VS_Code_Insiders-NPM-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=slide&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22apiKey%22%7D%5D&config=%7B%22command%22%3A%22npx%22%2C%22args%22%3A%5B%22-y%22%2C%22%40modelcontextprotocol%2Fserver-slide%22%5D%2C%22env%22%3A%7B%22SLIDE_API_KEY%22%3A%22%24%7Binput%3Aslide_api_key%7D%22%7D%7D&quality=insiders)

[![Install with Docker in VS Code](https://img.shields.io/badge/VS_Code-Docker-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=slide&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22apiKey%22%7D%5D&config=%7B%22command%22%3A%22docker%22%2C%22args%22%3A%5B%22run%22%2C%22-i%22%2C%22--rm%22%2C%22-e%22%2C%22SLIDE_API_KEY%22%2C%22mcp%2Fslide%22%5D%2C%22env%22%3A%7B%22SLIDE_API_KEY%22%3A%22%24%7Binput%3Aslide_api_key%7D%22%7D%7D) [![Install with Docker in VS Code Insiders](https://img.shields.io/badge/VS_Code_Insiders-Docker-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=slide&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22apiKey%22%7D%5D&config=%7B%22command%22%3A%22docker%22%2C%22args%22%3A%5B%22run%22%2C%22-i%22%2C%22--rm%22%2C%22-e%22%2C%22SLIDE_API_KEY%22%2C%22mcp%2Fslide%22%5D%2C%22env%22%3A%7B%22SLIDE_API_KEY%22%3A%22%24%7Binput%3Aslide_api_key%7D%22%7D%7D&quality=insiders)

For manual installation, add the following JSON block to your User Settings (JSON) file in VS Code. You can do this by pressing `Ctrl + Shift + P` and typing `Preferences: Open User Settings (JSON)`.

Optionally, you can add it to a file called `.vscode/mcp.json` in your workspace. This will allow you to share the configuration with others.

> Note that the `mcp` key is not needed in the `.vscode/mcp.json` file.

#### Docker

```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "slide_api_key",
        "description": "Slide API Key",
        "password": true
      }
    ],
    "servers": {
      "slide": {
        "command": "docker",
        "args": [
          "run",
          "-i",
          "--rm",
          "-e",
          "SLIDE_API_KEY",
          "mcp/slide"
        ],
        "env": {
          "SLIDE_API_KEY": "${input:slide_api_key}"
        }
      }
    }
  }
}
```

#### NPX

```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "slide_api_key",
        "description": "Slide API Key",
        "password": true
      }
    ],
    "servers": {
      "slide": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-slide"],
        "env": {
          "SLIDE_API_KEY": "${input:slide_api_key}"
        }
      }
    }
  }
}
```

## Build

Docker build:

```bash
docker build -t mcp/slide:latest -f Dockerfile .
```

## License

This MCP server is licensed under the MIT License. This means you are free to use, modify, and distribute the software, subject to the terms and conditions of the MIT License. For more details, please see the LICENSE file in the project repository.

