#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_get_alerts
 * This script is executed by Claude Desktop when the slide_get_alerts tool is called
 */

require('dotenv').config({ path: require('path').resolve(__dirname, '../.env') });
const { createClient } = require('../index');

// Parse input arguments from Claude Desktop
const args = process.argv.length > 2 ? JSON.parse(process.argv[2]) : {};

// Get API key from environment variables
const apiKey = process.env.SLIDE_API_KEY;

if (!apiKey) {
  console.error('ERROR: No API key found. Please create a .env file with SLIDE_API_KEY');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Execute the API call
async function getAlerts() {
  try {
    // Optional parameters
    const params = {};
    if (args.device_id) params.device_id = args.device_id;
    if (args.agent_id) params.agent_id = args.agent_id;
    if (args.resolved !== undefined) params.resolved = args.resolved;
    if (args.limit) params.limit = args.limit;
    if (args.offset) params.offset = args.offset;
    if (args.sort_by) params.sort_by = args.sort_by;
    if (args.sort_asc) params.sort_asc = args.sort_asc;

    const result = await slideClient.getAlerts(params);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      alerts: result.data.map(alert => ({
        id: alert.alert_id,
        type: alert.type,
        severity: alert.severity,
        message: alert.message,
        created_at: alert.created_at,
        resolved: alert.resolved,
        resolved_at: alert.resolved_at,
        device_id: alert.device_id,
        agent_id: alert.agent_id
      })),
      total: result.pagination.total
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
getAlerts(); 