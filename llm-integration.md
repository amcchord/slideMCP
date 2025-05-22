# Slide MCP LLM Integration Guide

This guide explains how to integrate the Slide API Model Configuration Profile (MCP) with various Large Language Models (LLMs).

## Overview

The Slide MCP provides a structured way for LLMs to interact with the Slide API, enabling them to:

1. Query information about devices, agents, backups, and snapshots
2. Initiate operations like backups and restores
3. Manage virtual machines and disaster recovery environments

## Understanding the Slide Entity Relationships

When working with the Slide API, it's important to understand the relationships between different entities:

```
Device → Agent → Backup → Snapshot
```

- **Device**: A physical Slide Box hardware appliance that provides backup storage and disaster recovery capabilities.
- **Agent**: Software installed on a client system (server, workstation, etc.) to be protected. An Agent connects to a Device (Slide Box) for backup.
- **Backup**: A point-in-time operation that transfers data from an Agent to a Device.
- **Snapshot**: Created when a Backup completes successfully. It represents a recoverable point-in-time for the Agent's data.

Key relationships:
- A Device can have multiple Agents paired to it
- An Agent belongs to one Device and can have multiple Backups
- A Backup belongs to one Agent and can create one Snapshot if successful
- A Snapshot belongs to one Backup and one Agent

This hierarchical structure is important for understanding how to navigate the API and perform operations like:
- Listing all Agents connected to a specific Device
- Starting a Backup for a particular Agent
- Browsing Snapshots available for an Agent
- Creating a virtual machine from a specific Snapshot

### Examples for Traversing Entity Relationships

Here are examples of how to navigate between related entities:

#### From Device to Agents
```javascript
// Get all agents for a specific device
async function getAgentsForDevice(deviceId) {
  const agents = await slideClient.getAgents({ device_id: deviceId });
  return agents.data;
}
```

#### From Agent to Backups
```javascript
// Get all backups for a specific agent
async function getBackupsForAgent(agentId) {
  const backups = await slideClient.getBackups({ agent_id: agentId });
  return backups.data;
}
```

#### From Backup to Snapshot
```javascript
// Get the snapshot created by a backup
async function getSnapshotFromBackup(backup) {
  if (backup.snapshot_id) {
    const snapshot = await slideClient.getSnapshot(backup.snapshot_id);
    return snapshot;
  }
  return null; // No snapshot was created (backup might have failed)
}
```

#### From Agent to Latest Snapshot
```javascript
// Get the latest snapshot for an agent
async function getLatestSnapshotForAgent(agentId) {
  const snapshots = await slideClient.getSnapshots({ 
    agent_id: agentId,
    sort_by: 'backup_end_time',
    sort_asc: false,
    limit: 1
  });
  
  return snapshots.data.length > 0 ? snapshots.data[0] : null;
}
```

#### Complete Workflow Example
```javascript
// A complete workflow: finding the latest snapshot for a device's agent and creating a VM
async function createVMFromLatestSnapshot(deviceId, agentName) {
  // Find the specified agent on the device
  const agents = await slideClient.getAgents({ device_id: deviceId });
  const agent = agents.data.find(a => a.display_name === agentName);
  
  if (!agent) {
    throw new Error(`Agent with name "${agentName}" not found on device ${deviceId}`);
  }
  
  // Get the latest snapshot for this agent
  const snapshots = await slideClient.getSnapshots({ 
    agent_id: agent.agent_id,
    sort_by: 'backup_end_time',
    sort_asc: false,
    limit: 1
  });
  
  if (snapshots.data.length === 0) {
    throw new Error(`No snapshots found for agent ${agent.agent_id}`);
  }
  
  const latestSnapshot = snapshots.data[0];
  
  // Create a virtual machine from this snapshot
  const vm = await slideClient.createVirtualMachine({
    device_id: deviceId,
    snapshot_id: latestSnapshot.snapshot_id,
    cpu_count: 2,
    memory_in_mb: 4096
  });
  
  return vm;
}
```

## Snapshot Recovery Options

Snapshots in the Slide platform are versatile recovery points that can be used in multiple ways. Each recovery method serves different use cases and provides different levels of granularity for recovery.

### Virtualization

Virtualization creates a complete virtual machine from a snapshot, allowing you to run the entire system in a virtualized environment on the Slide device.

**Use Cases:**
- Disaster recovery testing
- Running a production workload after a disaster
- Testing application updates in a sandbox
- Development environments from production snapshots

**Example Workflow:**
```javascript
// Create a VM from a snapshot
async function createVMFromSnapshot(snapshotId, deviceId, config = {}) {
  // Default VM configuration
  const vmConfig = {
    snapshot_id: snapshotId,
    device_id: deviceId,
    cpu_count: config.cpu_count || 2,
    memory_in_mb: config.memory_in_mb || 4096,
    disk_bus: config.disk_bus || 'sata',
    network_model: config.network_model || 'e1000'
  };
  
  // Add network configuration if provided
  if (config.network_type) {
    vmConfig.network_type = config.network_type;
    
    if (config.network_type === 'network-id' && config.network_id) {
      vmConfig.network_source = config.network_id;
    }
  }
  
  // Create the VM
  const vm = await slideClient.createVirtualMachine(vmConfig);
  
  return vm;
}

// Start a stopped VM
async function startVM(virtId) {
  return slideClient.updateVirtualMachine(virtId, {
    state: 'running'
  });
}

// Stop a running VM
async function stopVM(virtId) {
  return slideClient.updateVirtualMachine(virtId, {
    state: 'stopped'
  });
}

// Delete a VM
async function deleteVM(virtId) {
  return slideClient.deleteVirtualMachine(virtId);
}
```

### Image Export

Image exports allow you to export a snapshot as a disk image in various formats (VHDX, VHD, Raw) for use with other virtualization platforms or for offline recovery.

**Use Cases:**
- Migration to other virtualization platforms
- Offline recovery without the Slide device
- Long-term archival of system state
- Transferring systems between environments

**Example Workflow:**
```javascript
// Export a snapshot as a disk image
async function exportSnapshotAsImage(snapshotId, deviceId, imageType = 'vhdx') {
  // Create the image export
  const imageExport = await slideClient.createImageExport({
    snapshot_id: snapshotId,
    device_id: deviceId,
    image_type: imageType
  });
  
  // Wait for the export to complete (this might take some time)
  // In a real implementation, you would poll until completion
  
  // Get the export entries (disk images)
  const exportEntries = await slideClient.browseImageExport(imageExport.image_export_id);
  
  return {
    imageExport,
    entries: exportEntries.data
  };
}

// Get download URLs for exported images
async function getImageDownloadURLs(imageExportId) {
  const exportEntries = await slideClient.browseImageExport(imageExportId);
  
  return exportEntries.data.map(entry => ({
    name: entry.name,
    size: entry.size,
    downloadURIs: entry.download_uris
  }));
}

// Delete an image export when no longer needed
async function deleteImageExport(imageExportId) {
  return slideClient.deleteImageExport(imageExportId);
}
```

### File Restore

File restores mount a snapshot to allow browsing and restoring individual files and folders without recovering the entire system.

**Use Cases:**
- Recovering specific deleted files
- Accessing previous versions of documents
- Extracting configuration files from a backup
- Retrieving data from a corrupted system

**Example Workflow:**
```javascript
// Create a file restore from a snapshot
async function createFileRestoreFromSnapshot(snapshotId, deviceId) {
  return slideClient.createFileRestore({
    snapshot_id: snapshotId,
    device_id: deviceId
  });
}

// Browse files in a file restore
async function browseFileRestore(fileRestoreId, path = '') {
  return slideClient.browseFileRestore(fileRestoreId, {
    path: path
  });
}

// Get file download URL
async function getFileDownloadInfo(fileRestoreId, path) {
  const browseResult = await slideClient.browseFileRestore(fileRestoreId, {
    path: path
  });
  
  // Find the specific file in the results
  const fileName = path.split('/').pop();
  const file = browseResult.data.find(item => 
    item.name === fileName && item.type === 'file'
  );
  
  if (!file) {
    throw new Error(`File not found: ${path}`);
  }
  
  return {
    name: file.name,
    size: file.size,
    modified_at: file.modified_at,
    download_uris: file.download_uris
  };
}

// Delete a file restore when no longer needed
async function deleteFileRestore(fileRestoreId) {
  return slideClient.deleteFileRestore(fileRestoreId);
}
```

## Integration Methods

### OpenAI Function Calling

For models that support function calling (like GPT-4), you can define functions that map to the Slide MCP:

```javascript
const functions = [
  {
    name: "getDevices",
    description: "Get a list of Slide devices",
    parameters: {
      type: "object",
      properties: {
        limit: {
          type: "integer",
          description: "Maximum number of devices to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  {
    name: "getDevice",
    description: "Get details about a specific Slide device",
    parameters: {
      type: "object",
      properties: {
        deviceId: {
          type: "string",
          description: "ID of the device to retrieve"
        }
      },
      required: ["deviceId"]
    }
  },
  // Add more functions here
];

// When the LLM calls a function, execute the corresponding Slide MCP method
async function executeFunction(functionName, args) {
  const slideMCP = require('./index');
  const slideClient = slideMCP.createClient(process.env.SLIDE_API_KEY);
  
  switch(functionName) {
    case "getDevices":
      return await slideClient.getDevices(args);
    case "getDevice":
      return await slideClient.getDevice(args.deviceId);
    // Add more cases here
  }
}
```

### Anthropic Claude / Function Calling

For Anthropic Claude models with tool use:

```javascript
const tools = [
  {
    name: "slide_get_devices",
    description: "Get a list of Slide devices",
    input_schema: {
      type: "object",
      properties: {
        limit: {
          type: "integer",
          description: "Maximum number of devices to return"
        },
        offset: {
          type: "integer",
          description: "Starting index for pagination"
        }
      }
    }
  },
  // Add more tools here
];

// Tool execution handler
async function handleToolCall(toolName, args) {
  const slideMCP = require('./index');
  const slideClient = slideMCP.createClient(process.env.SLIDE_API_KEY);
  
  if (toolName === "slide_get_devices") {
    return await slideClient.getDevices(args);
  }
  // Add more handlers here
}
```

### Langchain Integration

You can integrate the Slide MCP with Langchain:

```javascript
const { Tool } = require('langchain/tools');
const slideMCP = require('./index');

class SlideDevicesTool extends Tool {
  name = "slide_devices";
  description = "Get information about Slide devices";
  
  constructor(apiKey) {
    super();
    this.client = slideMCP.createClient(apiKey);
  }
  
  async _call(arg) {
    const args = JSON.parse(arg || "{}");
    return JSON.stringify(await this.client.getDevices(args));
  }
}

// Add more tool classes for other endpoints
```

## Error Handling

When integrating with LLMs, it's important to handle errors gracefully:

1. API errors should be presented in a human-readable format
2. Authentication failures should guide the user to check their API key
3. Rate limiting should be handled with appropriate retry logic

## Example Prompts

Here are some example prompts to help LLMs use the Slide MCP effectively:

1. "List all Slide devices and their status"
2. "Start a backup for agent with ID a_0123456789ab"
3. "Create a virtual machine from the most recent snapshot of agent a_0123456789ab"
4. "Check if there are any unresolved alerts"

## Security Considerations

1. Always use environment variables for API keys
2. Never expose API keys in client-side code
3. Consider implementing rate limiting in your application
4. Set appropriate timeouts for API calls
5. Validate user input before passing to the API 