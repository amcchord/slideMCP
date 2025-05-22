/**
 * Example script for using Slide MCP
 * This demonstrates how to use the MCP in your own code
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

// Example function to list all devices
async function listDevices() {
  try {
    console.log('Getting devices from Slide API...');
    const devices = await slideClient.getDevices();
    
    console.log(`\nFound ${devices.pagination.total} devices:`);
    devices.data.forEach(device => {
      console.log(`- ${device.device_id}: ${device.display_name || 'No name'} (Status: ${device.status})`);
    });
    
    return devices;
  } catch (error) {
    console.error('Error getting devices:', error.message);
  }
}

// Run the example
listDevices(); 