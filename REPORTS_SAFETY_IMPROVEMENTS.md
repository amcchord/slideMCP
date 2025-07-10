# Reports Tool Safety Improvements

## Problem
The reports tool was hanging due to infinite loops in pagination when processing snapshots and backups. This occurred when the API returned the same offset repeatedly or when individual agent reports got stuck.

## Solutions Implemented

### 1. Pagination Safety Checks

All pagination loops now include:

```go
// Safety check: ensure we're making progress
newOffset := *pagination.NextOffset
if newOffset <= offset {
    // We're not making progress, break to avoid infinite loop
    if verbose {
        fmt.Fprintf(os.Stderr, "[WARNING] Pagination not progressing, stopping at offset %d\n", offset)
    }
    break
}
```

**Applied to:**
- `calculateBackupStats` - backup pagination loop
- `calculateSnapshotStats` - active snapshots loop
- `calculateSnapshotStats` - deleted snapshots loop

### 2. Maximum Offset Limit

Added a hard limit to prevent runaway pagination:

```go
// Additional safety: limit maximum iterations
if offset > 10000 {
    if verbose {
        fmt.Fprintf(os.Stderr, "[WARNING] Reached maximum offset limit\n")
    }
    break
}
```

### 3. Per-Agent Timeouts

Individual agent reports now have a 30-second timeout:

```go
select {
case report := <-reportChan:
    resultChan <- report
case <-time.After(30 * time.Second):
    if verbose {
        fmt.Fprintf(os.Stderr, "[Concurrent] Timeout processing agent after 30 seconds\n")
    }
}
```

### 4. Overall Operation Timeout

The entire concurrent processing operation has a 5-minute timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
```

### 5. Context Cancellation

Proper context handling for graceful shutdown:
- Check context before starting work
- Check context when acquiring semaphore
- Check context while processing

## Testing

Run the safety test script:
```bash
./test_scripts/test_reports_no_hang.sh
```

This script:
- Uses a 60-second timeout to ensure the command doesn't hang
- Runs in verbose mode to show any warning messages
- Verifies all safety mechanisms are working

## Key Benefits

1. **No More Infinite Loops**: Pagination loops detect when they're not making progress
2. **Bounded Execution Time**: Timeouts ensure reports complete or fail within reasonable time
3. **Graceful Degradation**: Individual agent failures don't block the entire report
4. **Visibility**: Verbose mode shows exactly where issues occur
5. **Defensive Programming**: Multiple layers of protection against hanging

## Monitoring

When running reports with `verbose: true`, watch for these warning messages:
- `"Pagination not progressing"` - Indicates API returning same offset
- `"Reached maximum offset limit"` - Indicates runaway pagination
- `"Timeout processing agent"` - Indicates slow individual agent processing
- `"Context cancelled"` - Indicates overall timeout reached 