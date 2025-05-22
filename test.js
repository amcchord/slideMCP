/**
 * Test script for Slide API client
 * This will test basic functionality using the provided API key
 */

require('dotenv').config();
const { createClient } = require('./index');

// Get API key from environment variables
const apiKey = process.env.SLIDE_API_KEY;

if (!apiKey) {
  console.error('ERROR: No API key found. Please create a .env file with SLIDE_API_KEY');
  process.exit(1);
}

// Create Slide client
const slideClient = createClient(apiKey);

// Function to test the API
async function testApi() {
  try {
    console.log('Testing Slide API connection...');
    
    // Test devices endpoint
    console.log('\n--- Testing Devices API ---');
    const devices = await slideClient.getDevices({ limit: 5 });
    console.log(`Found ${devices.pagination.total} devices`);
    
    if (devices.data.length > 0) {
      const deviceId = devices.data[0].device_id;
      console.log(`Getting details for device: ${deviceId}`);
      const device = await slideClient.getDevice(deviceId);
      console.log(`Device name: ${device.display_name}`);
    }

    // Test agents endpoint if there are devices
    if (devices.data.length > 0) {
      console.log('\n--- Testing Agents API ---');
      const agents = await slideClient.getAgents({ limit: 5 });
      console.log(`Found ${agents.pagination.total} agents`);
      
      if (agents.data.length > 0) {
        const agentId = agents.data[0].agent_id;
        console.log(`Getting details for agent: ${agentId}`);
        const agent = await slideClient.getAgent(agentId);
        console.log(`Agent hostname: ${agent.hostname}`);
      }
    }

    console.log('\nAPI test completed successfully!');
  } catch (error) {
    console.error('API test failed:', error.message);
    process.exit(1);
  }
}

// Run the test
testApi(); 