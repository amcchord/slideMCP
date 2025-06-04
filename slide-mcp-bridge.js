#!/usr/bin/env node
/**
 * Slide MCP Bridge for Claude Desktop
 * 
 * This script bridges Claude Desktop to the hosted Slide MCP server at www.slide.recipes/mcp
 * No local installation required - just download this file and configure Claude Desktop to use it.
 * 
 * Usage:
 * 1. Download this file
 * 2. Make it executable: chmod +x slide-mcp-bridge.js  
 * 3. Set SLIDE_API_KEY environment variable
 * 4. Configure Claude Desktop to use this script
 */

const https = require('https');
const readline = require('readline');

// Configuration
const API_KEY = process.env.SLIDE_API_KEY;
const SERVER_HOST = 'www.slide.recipes';
const SERVER_PATH = '/mcp';

if (!API_KEY) {
    console.error(JSON.stringify({
        jsonrpc: "2.0",
        id: null,
        error: {
            code: -32602,
            message: "SLIDE_API_KEY environment variable not set"
        }
    }));
    process.exit(1);
}

// Setup readline interface for MCP protocol
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: false
});

// Process each MCP request
rl.on('line', (line) => {
    if (!line.trim()) return;
    
    try {
        // Parse the MCP request
        const mcpRequest = JSON.parse(line);
        
        // Forward to hosted server
        const postData = JSON.stringify(mcpRequest);
        
        const options = {
            hostname: SERVER_HOST,
            path: SERVER_PATH,
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Content-Length': Buffer.byteLength(postData),
                'X-API-Key': API_KEY,
                'User-Agent': 'slide-mcp-bridge/1.0'
            }
        };
        
        const req = https.request(options, (res) => {
            let responseData = '';
            
            res.on('data', (chunk) => {
                responseData += chunk;
            });
            
            res.on('end', () => {
                try {
                    // Validate and forward the response
                    const response = JSON.parse(responseData);
                    console.log(JSON.stringify(response));
                } catch (parseError) {
                    // Send error response if server response is invalid
                    console.log(JSON.stringify({
                        jsonrpc: "2.0",
                        id: mcpRequest.id,
                        error: {
                            code: -32603,
                            message: "Invalid response from server"
                        }
                    }));
                }
            });
        });
        
        req.on('error', (error) => {
            // Send error response if network request fails
            console.log(JSON.stringify({
                jsonrpc: "2.0",
                id: mcpRequest.id,
                error: {
                    code: -32603,
                    message: `Network error: ${error.message}`
                }
            }));
        });
        
        // Send the request
        req.write(postData);
        req.end();
        
    } catch (parseError) {
        // Send error response if MCP request is invalid JSON
        console.log(JSON.stringify({
            jsonrpc: "2.0",
            id: null,
            error: {
                code: -32700,
                message: "Parse error"
            }
        }));
    }
});

// Handle cleanup
process.on('SIGINT', () => {
    rl.close();
    process.exit(0);
});

process.on('SIGTERM', () => {
    rl.close();
    process.exit(0);
}); 