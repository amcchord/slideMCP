# Slide MCP Server Quickstart Guide for Claude Desktop

This guide will help you set up the Slide MCP Server to use with Claude Desktop so you can manage your Slide devices and agents.

## Prerequisites

- [Node.js](https://nodejs.org/) (version 18 or higher)
- [Claude Desktop](https://claude.ai/desktop) installed on your computer
- A Slide API key (obtained from your Slide account)

## Step 1: Get Your Slide API Key

1. Log in to your [Slide account](https://console.slide.tech/)
2. Navigate to your account settings
3. Generate an API key from the API section

## Step 2: Download and Configure the MCP Server

### Option 1: Using NPX (Faster for Node peopl)

1. Make sure you have Node.js installed (version 18 or higher)

2. You can manually install the package globally (optional):
   ```bash
   npm install -g @modelcontextprotocol/server-slide
   ```
   
   Note: This step is optional since the configuration below uses NPX which will automatically download and run the package.

3. Create a file called `claude_desktop_config.json` in your home directory:
   - On macOS/Linux: `~/.claude_desktop_config.json`
   - On Windows: `C:\Users\YourUsername\.claude_desktop_config.json`

4. Add the following configuration to the file:

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

3. Replace `YOUR_API_KEY_HERE` with your actual Slide API key.

### Option 2: Using Docker (Austin's preffered method due to node alergies)

If you prefer using Docker:

1. Make sure Docker is installed on your system
2. Clone or download the Slide MCP Server code
3. Build the Docker image:
   ```bash
   docker build -t mcp/slide:latest -f Dockerfile .
   ```
4. Create the `claude_desktop_config.json` file as described above
5. Use the following configuration instead:

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

## Step 3: Start Claude Desktop

1. Launch Claude Desktop
2. The MCP server will automatically start when needed

## Step 4: Using Slide Tools with Claude

Claude can now access your Slide devices and agents. Try asking Claude questions like:

- "Show me a list of my Slide devices"
- "Help me create a backup of my agent"
- "List all my Slide snapshots"
- "Create a virtual machine from my latest snapshot"

Claude will use the appropriate Slide tools to fulfill your requests.

## Troubleshooting

- **Configuration File Not Found**: Make sure your `claude_desktop_config.json` file is in the correct location and properly formatted.
- **API Key Issues**: Verify that your API key is correct and has the necessary permissions.
- **Server Not Starting**: Check if Node.js is properly installed and is version 18 or higher.
- **Docker Issues**: Ensure Docker is running if you chose the Docker configuration.

For more detailed information, refer to the full [README.md](README.md) file in the project repository.
