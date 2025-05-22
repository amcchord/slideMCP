#!/usr/bin/env node

/**
 * Claude Desktop Tool Handler: slide_update_account
 * This script is executed by Claude Desktop when the slide_update_account tool is called
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
async function updateAccount() {
  try {
    // Required parameters check
    if (!args.account_id) {
      throw new Error('account_id is required');
    }
    
    // Create payload
    const payload = {};
    if (args.alert_emails) payload.alert_emails = args.alert_emails;
    
    if (Object.keys(payload).length === 0) {
      throw new Error('At least one of alert_emails is required');
    }

    const result = await slideClient.updateAccount(args.account_id, payload);
    
    // Format the result for better readability in Claude's response
    const formattedResult = {
      id: result.account_id,
      alert_emails: result.alert_emails,
      updated_at: result.updated_at
    };
    
    // Claude Desktop expects JSON output
    console.log(JSON.stringify(formattedResult, null, 2));
  } catch (error) {
    console.error(JSON.stringify({ error: error.message }));
    process.exit(1);
  }
}

// Run the handler
updateAccount(); 