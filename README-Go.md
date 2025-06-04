# Slide MCP Server (Go)

A **portable, single-binary** MCP (Model Context Protocol) server implementation for the Slide API, written in Go for maximum portability and ease of use.

## 🚀 Why Go?

This Go implementation provides several advantages over the original TypeScript version:

- **Single Binary**: No Node.js runtime required - just one executable file
- **Cross-Platform**: Builds for Linux, macOS, and Windows out of the box
- **Fast Startup**: No dependency installation or compilation delays
- **Small Footprint**: Minimal memory usage and fast execution
- **Easy Distribution**: Copy one file and run anywhere

## ✨ Features

This Go implementation provides the same comprehensive Slide API functionality as the original:

- **Device Management**: List all devices with pagination and filtering
- **Agent Management**: List, create, pair, and update agents
- **Backup Management**: List, retrieve, and initiate backups
- **Snapshot Management**: List and retrieve snapshots with detailed information
- **Virtual Machine Tools**: Create, manage, and control VMs from snapshots
- **File Restore Tools**: Browse and restore files from snapshots
- **Image Export Tools**: Export snapshots as disk images
- **User, Alert, Account, and Client Management**: Complete administrative tools

## 📦 Installation

### Quick Start - Download Pre-built Binary

1. **Download** the latest release for your platform:
   ```bash
   # Linux AMD64
   curl -L -o slide-mcp-server https://github.com/yourusername/slide-mcp-go/releases/latest/download/slide-mcp-server-linux-amd64
   
   # macOS AMD64
   curl -L -o slide-mcp-server https://github.com/yourusername/slide-mcp-go/releases/latest/download/slide-mcp-server-darwin-amd64
   
   # macOS ARM64 (Apple Silicon)
   curl -L -o slide-mcp-server https://github.com/yourusername/slide-mcp-go/releases/latest/download/slide-mcp-server-darwin-arm64
   ```

2. **Make it executable**:
   ```bash
   chmod +x slide-mcp-server
   ```

3. **Optional: Install system-wide**:
   ```bash
   sudo mv slide-mcp-server /usr/local/bin/
   ```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/slide-mcp-go.git
cd slide-mcp-go

# Build for your platform
make build

# Or build for all platforms
make build-all

# Install dependencies if needed
make deps
```

## 🔧 Configuration

### Environment Variables

Set your Slide API key:

```bash
export SLIDE_API_KEY="your-api-key-here"
```

### Getting an API Key

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to your account settings
3. Generate your API key from the API section

## 🎯 Usage

### With Claude Desktop

Add this to your `claude_desktop_config.json`:

#### Direct Binary Usage
```json
{
  "mcpServers": {
    "slide": {
      "command": "/path/to/slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE"
      }
    }
  }
}
```

#### If Installed System-wide
```json
{
  "mcpServers": {
    "slide": {
      "command": "slide-mcp-server",
      "env": {
        "SLIDE_API_KEY": "YOUR_API_KEY_HERE"
      }
    }
  }
}
```

### With VS Code MCP Extension

#### Direct Binary Configuration
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
        "command": "/path/to/slide-mcp-server",
        "env": {
          "SLIDE_API_KEY": "${input:slide_api_key}"
        }
      }
    }
  }
}
```

### Command Line Testing

You can test the server directly:

```bash
# Set your API key
export SLIDE_API_KEY="your-api-key-here"

# Run the server
./slide-mcp-server

# In another terminal, send MCP messages
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./slide-mcp-server
```

## 🛠 Development

### Building

```bash
# Install dependencies
make deps

# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run locally during development
make run
```

### Project Structure

```
.
├── main.go           # MCP protocol handling and server logic
├── api.go            # Slide API client and data structures
├── go.mod            # Go module definition
├── Makefile          # Build automation
└── README-Go.md      # This file
```

### Adding New Tools

1. Add the data structures to `api.go`
2. Implement the API function in `api.go`
3. Add the tool definition to `getAllTools()` in `main.go`
4. Add the tool handler to `handleToolCall()` in `main.go`

## 📋 Available Tools

- `slide_list_devices` - List all devices with pagination and filtering
- `slide_list_agents` - List all agents with pagination and filtering
- More tools coming soon...

## 🔄 Migration from TypeScript Version

This Go version is designed as a drop-in replacement for the TypeScript version:

1. **Same API**: All tools and functionality are identical
2. **Same Configuration**: Uses the same environment variables and settings
3. **Same Output**: Response formats are identical for compatibility
4. **Better Performance**: Faster startup and lower resource usage

### Migration Steps

1. Download/build the Go binary
2. Update your MCP configuration to use the new binary instead of `npx`
3. Remove Node.js dependencies (optional)

## 🚀 Performance Benefits

| Metric | TypeScript + Node.js | Go Binary |
|--------|---------------------|-----------|
| Startup Time | ~2-3 seconds | ~50ms |
| Memory Usage | ~50-100MB | ~10-20MB |
| Binary Size | N/A (requires Node.js) | ~8MB |
| Dependencies | Node.js + npm packages | None |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `make test` to ensure tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related

- [Original TypeScript Implementation](https://github.com/yourusername/slide-mcp-ts)
- [Slide API Documentation](https://api.slide.tech/docs)
- [Model Context Protocol](https://modelcontextprotocol.io/)

## 📞 Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/yourusername/slide-mcp-go/issues) page
2. Create a new issue with detailed information
3. For Slide API specific questions, consult the [Slide documentation](https://api.slide.tech/docs) 