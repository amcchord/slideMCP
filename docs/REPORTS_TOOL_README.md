# Slide Reports Tool

## Overview

The `slide_reports` tool has been added to the Slide MCP Server to provide pre-calculated statistics and reports for LLMs. This tool helps LLMs by doing the heavy lifting of data aggregation and calculation, presenting the results in easily digestible formats.

## Features

### Daily Backup/Snapshot Report (`daily_backup_snapshot`)

Provides comprehensive daily statistics including:

- **Backup Statistics**
  - Total backup attempts
  - Successful backups
  - Failed backups
  - In-progress backups
  - Success rate percentage
  - Failure reasons with counts

- **Snapshot Statistics**
  - Total snapshots
  - Active snapshots
  - Storage locations:
    - Local storage count
    - Cloud storage count

### Weekly Backup/Snapshot Report (`weekly_backup_snapshot`)

Provides weekly statistics (Sunday to Saturday) including:

- **Weekly Summary**
  - Total unique agents reporting
  - Total backups for the week
  - Weekly success rate
  - Total snapshots

- **Daily Breakdown**
  - Statistics for each day of the week
  - Number of agents per day
  - Daily backup counts and success rates
  - Daily snapshot counts

### Monthly Backup/Snapshot Report (`monthly_backup_snapshot`)

Provides monthly statistics with a visual calendar including:

- **Monthly Summary**
  - Total unique agents reporting
  - Total backups for the month
  - Monthly success rate
  - Total snapshots

- **Calendar View** (Markdown format only)
  - Visual calendar table showing each day
  - Success indicators on each day:
    - ✓ = ≥90% success rate
    - ~ = 50-89% success rate
    - ✗ = <50% success rate
    - No indicator = no backups that day

- **Daily Details**
  - Detailed breakdown for each day with activity
  - Agent counts, backup statistics, and success rates

## Usage

### Tool Parameters

- `operation` (required): Type of report to generate
  - `"daily_backup_snapshot"` - Single day report
  - `"weekly_backup_snapshot"` - Weekly report (Sunday to Saturday)
  - `"monthly_backup_snapshot"` - Monthly report with calendar view
- `date` (optional): Date in YYYY-MM-DD format
  - For daily: defaults to today
  - For weekly: any date in the target week (defaults to current week)
  - For monthly: any date in the target month (defaults to current month)
- `agent_id` (optional): Filter by specific agent ID
- `device_id` (optional): Filter by device ID (includes all agents on device)
- `client_id` (optional): Filter by client ID (includes all agents for client)
- `format` (optional): Output format - `"json"` (default) or `"markdown"`
- `verbose` (optional): Enable verbose progress logging to stderr - `true` or `false`

### Example Usage

Daily report:
```json
{
  "name": "slide_reports",
  "arguments": {
    "operation": "daily_backup_snapshot",
    "agent_id": "a_mbepgxrb629h",
    "format": "markdown"
  }
}
```

Weekly report for current week:
```json
{
  "name": "slide_reports",
  "arguments": {
    "operation": "weekly_backup_snapshot",
    "format": "markdown"
  }
}
```

Monthly report with calendar for December 2024:
```json
{
  "name": "slide_reports",
  "arguments": {
    "operation": "monthly_backup_snapshot",
    "date": "2024-12-15",
    "format": "markdown"
  }
}
```

### Output Formats

#### JSON Format
Returns structured data with:
- Date
- Array of reports (one per agent)
- Each report includes agent details and statistics
- Metadata with guidance for interpretation

#### Markdown Format
Returns a human-readable report with:
- Summary statistics across all agents
- Individual agent sections with detailed metrics
- Clear formatting for easy reading

## Implementation Details

The tool is implemented in `tools_reports.go` and includes:

1. **Data Aggregation**: Fetches backups and snapshots from the API
2. **Statistical Calculation**: Computes success rates, counts, and categorizations
3. **Filtering**: Supports filtering by agent, device, or client
4. **Date Handling**: Processes data for specific dates (defaults to today)
5. **Format Conversion**: Outputs in JSON or Markdown format

## Benefits for LLMs

1. **Pre-calculated Metrics**: No need for LLMs to perform complex calculations
2. **Structured Data**: Well-organized data that's easy to interpret
3. **Flexible Filtering**: Can get reports at different levels of granularity
4. **Multiple Formats**: JSON for data processing, Markdown for presentation
5. **Error Categorization**: Failure reasons are pre-grouped for analysis

## Testing

Several test scripts have been created:

- `verify_reports_tool.sh`: Verifies the tool is properly integrated
- `test_single_agent_report.sh`: Tests report generation for a single agent
- `example_report.sh`: Demonstrates usage with a specific agent
- `test_weekly_monthly_reports.sh`: Tests weekly and monthly report operations

To test with the provided API key:
```bash
./slide-mcp-server --api-key "tk_d3m9fta2e6qq_rFGiifl8IcDMGekXNvnjfh3dpapr4RbL"
```

## Performance Considerations

Report generation can be time-consuming for large deployments:

- **Daily reports**: Process 1 day × N agents
- **Weekly reports**: Process 7 days × N agents  
- **Monthly reports**: Process ~30 days × N agents

### Performance Tips

1. **Use filters** - Always filter by agent_id, device_id, or client_id when possible
2. **Enable verbose mode** - Set `verbose: true` to see progress during long operations
3. **Start small** - Test with single agents before running system-wide reports
4. **Consider timing** - Monthly reports for hundreds of agents can take several minutes

### What Happens During Report Generation

For each day in the report period:
1. Fetch list of agents (if no filter specified)
2. For each agent:
   - Get agent details
   - Fetch all backups for that date
   - Fetch all snapshots (active + deleted)
   - Calculate statistics

## Test Examples

### Daily Report for Single Agent
```bash
./verify_reports_tool.sh daily_single
```

### Weekly Report for Device
```bash
./verify_reports_tool.sh weekly_device
```

### Monthly Report with Calendar View
```bash
./verify_reports_tool.sh monthly_markdown
```

### Full Test Suite
```bash
./test_weekly_monthly_reports.sh
```

## Expected Performance

The reports tool makes multiple API calls with pagination, resulting in these approximate timings:

- **Single agent daily report**: 5-30 seconds
- **Single agent weekly report**: 30-120 seconds  
- **All agents daily report**: 2-5 minutes (for ~10 agents)
- **All agents weekly report**: 10-20 minutes (for ~10 agents)
- **All agents monthly report**: 20-40 minutes (for ~10 agents)

### Verbose Logging

Use the `verbose` flag to see progress during long operations:

```json
{
  "operation": "weekly_backup_snapshot",
  "agent_id": "a_mbepgxrb629h",
  "verbose": true
}
```

This will show progress messages like:
```
[Weekly Report] Processing Sunday (2025-07-06)...
[Backup Stats] Fetching backups for agent a_mbepgxrb629h (offset: 0)...
[Agent Fetch] Retrieved 10 agents (total: 10)
```

### Best Practices

1. **Always use filters** (agent_id, device_id, or client_id) for faster results
2. **Start with daily reports** before attempting weekly/monthly
3. **Schedule large reports** during off-hours in production
4. **Monitor progress** with the verbose flag for long operations

## Future Enhancements

Potential additional report types could include:
- Trend analysis reports
- Storage utilization reports
- Performance metrics reports
- Alert correlation reports
- Cached/pre-computed reports for better performance 