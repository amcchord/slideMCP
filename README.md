# slideMCP

A Model Configuration Profile (MCP) for Large Language Models to interact with the Slide API.

## Entity Relationships

The Slide API is built around the following key entities and their relationships:

```
Device → Agent → Backup → Snapshot
```

- **Device**: A physical Slide Box hardware appliance
- **Agent**: Software installed on systems to be protected
- **Backup**: A point-in-time operation copying data from an Agent
- **Snapshot**: A recoverable point-in-time created by a successful Backup

Understanding these relationships is crucial for effective API usage:
- A Device can have multiple Agents
- An Agent can have multiple Backups
- A successful Backup creates a Snapshot

## Snapshot Recovery Options

Snapshots are the core recovery point in the Slide platform and can be used in multiple ways:

1. **Virtualization**: Create a virtual machine from a snapshot
   - Run a complete system in a virtualized environment
   - Used for disaster recovery, testing, or development environments
   - Example: `slideClient.createVirtualMachine({ snapshot_id: 's_0123456789ab', device_id: 'd_0123456789ab' })`

2. **Image Export**: Export a snapshot as a disk image
   - Export in various formats (VHDX, VHD, Raw)
   - Used for migration to other platforms or offline recovery
   - Example: `slideClient.createImageExport({ snapshot_id: 's_0123456789ab', device_id: 'd_0123456789ab', image_type: 'vhdx' })`

3. **File Restore**: Mount a snapshot to browse and restore files
   - Access individual files and folders without full system restore
   - Used for targeted recovery of specific data
   - Example: `slideClient.createFileRestore({ snapshot_id: 's_0123456789ab', device_id: 'd_0123456789ab' })`

## Setup

1. Clone this repository
2. Create a `.env` file in the root directory with your API key:
   ```
   SLIDE_API_KEY=your_api_key_here
   ```
3. Install dependencies:
   ```bash
   npm install
   ```

## Usage

```javascript
const slideMCP = require('./index');
const slideClient = slideMCP.createClient(process.env.SLIDE_API_KEY);

// Example: List devices
slideClient.getDevices()
  .then(response => console.log(response))
  .catch(error => console.error(error));

// Example: Get all agents for a specific device
slideClient.getAgents({ device_id: 'd_0123456789ab' })
  .then(response => console.log(response))
  .catch(error => console.error(error));

// Example: Start a backup for an agent
slideClient.startBackup({ agent_id: 'a_0123456789ab' })
  .then(response => console.log(response))
  .catch(error => console.error(error));

// Example: Create a virtual machine from a snapshot
slideClient.createVirtualMachine({
  snapshot_id: 's_0123456789ab',
  device_id: 'd_0123456789ab',
  cpu_count: 2,
  memory_in_mb: 4096
})
  .then(response => console.log(response))
  .catch(error => console.error(error));
```

## Security

- Never commit your `.env` file to the repository
- The `.gitignore` file is configured to exclude `.env` files
- Use environment variables in production environments

## API Coverage

This MCP provides access to the following Slide API endpoints:

- Devices
- Agents
- Backups
- Snapshots
- File Restores
- Virtual Machines
- Image Exports
- Networks
- Users
- Alerts
- Account