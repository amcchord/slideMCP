#!/bin/bash

# Change to the script directory
cd "$(dirname "$0")"

# Load environment variables from .env file
if [ -f .env ]; then
    source .env
else
    echo "Error: .env file not found"
    exit 1
fi

# Run the Node.js script with stdin/stdout connected
exec node mcp-test.js 