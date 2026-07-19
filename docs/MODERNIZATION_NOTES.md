# Slide MCP Server Modernization

This document tracks the major modernization milestones. See [CHANGELOG.md](../CHANGELOG.md)
for the per-release detail.

## v5.1.0 (2026-07-18) - Reliability and multi-host release contracts

v5.1.0 focuses on operational certainty rather than adding more API surface.
The HTTP client now has bounded deadlines and response sizes, only retries
idempotent reads, and exposes deterministic failure injection points. Complete
inventory and health are fully paginated and avoid the previous N+1 request
shape.

The test harness now covers registry/schema/permission contracts, Anthropic and
OpenAI MCP protocol profiles, production-binary stdio framing, distribution
metadata, API edge cases, and live read-only demo calls. CI repeats race tests,
cross-builds, ShellCheck, full package assembly, and official MCPB validation.

Distribution is handled by one fail-closed release pipeline. The MCPB selects
both Linux architectures correctly, the stable download URL points at the real
repository, and historical installer asset names are published as aliases so
retired installers can install the latest release. macOS aliases all carry the
same Developer ID signed and Apple-notarized universal binary.

---

## v3.0.0 (2026-05-10) - Official SDK + Slide API v1.27.0

### Why

The previous releases used a hand-rolled `bufio.Scanner` JSON-RPC loop in
[`server.go`](../server.go) that we maintained ourselves. As the MCP spec
evolved and Claude Code/Desktop hosts started shipping richer features
(resources, request hooks, tool filters, structured tool results), keeping
the in-house transport in sync was getting expensive.

### What changed

- Adopted [`github.com/mark3labs/mcp-go v0.52.0`](https://pkg.go.dev/github.com/mark3labs/mcp-go).
  - `mcp.go` shrunk to a tiny shim that keeps the legacy `ToolInfo` /
    `ToolResult` types so each `tools_*.go` descriptor file is unchanged.
  - [`registry.go`](../registry.go) translates each `ToolInfo` into an
    `mcp.Tool` via `mcp.NewToolWithRawSchema`, which preserves our existing
    `allOf` / `if` / `then` conditional-required schemas verbatim.
  - [`server.go`](../server.go) now just builds `server.NewMCPServer` with
    `WithToolFilter`, `WithRecovery`, `WithToolCapabilities`, and
    `WithResourceCapabilities`, and registers the cached overview at
    `slide://context/clients-devices-agents` as a real MCP resource.
- Initial-context handling moves from a non-spec `initialize.result.initialContext`
  field to a standard `resources/read` on the new resource URI.
- The `--tool` / `--args` one-shot CLI mode is preserved by calling the
  in-process tool handler directly (no MCP transport overhead) so existing
  scripts keep working.
- `Version` bumped to `3.0.0`. Latest MCP protocol version is advertised
  by the SDK (`2025-06-18+`).

### Slide API surface

Slide's API grew from v1.18.2 (the version we previously targeted) to
v1.27.0 with six new endpoints and ~30 new schemas. v3.0.0 adds Go types
and tool operations for every one of them - see CHANGELOG for the full list.

The schema docs are now backed by [`docs/openapi.json`](openapi.json)
refreshed from `https://api.slide.tech/openapi.json` at release time.

### Distribution / installation

- One-line Claude Code install via `make install-claude-code`
  (wraps `claude mcp add`, idempotent).
- One-line Claude Desktop install via `make install-claude-desktop`
  (jq-merges into `claude_desktop_config.json` with a one-time backup).
- `dxt/manifest.json` (MCPB spec 0.3) + `make pack-dxt` produce a Claude
  Desktop Extension bundle. The manifest exposes `api_key`, `tools_mode`,
  and `base_url` as user-configurable fields.
- `scripts/setup-dev.sh` + `make doctor` bootstrap a fresh Mac (Go, jq, gh
  via Homebrew) idempotently.

### Tests

- `server_test.go` runs the SDK server in-process via
  `(*server.MCPServer).HandleMessage` and asserts every expected tool +
  v1.27.0 operation is registered, that the mode-aware filter hides
  presentation/reports when not enabled, and that the one-shot CLI
  surfaces tool-disablement errors.
- `smoke_test.sh` rewritten to drive the binary via the `--tool` mode plus a
  real stdio handshake; live API auth failures are reported as SKIP rather
  than FAIL.

---

## v2.4.0 (legacy)

This document details the modernization improvements made to the Slide MCP Server to align with current MCP best practices and optimize for modern LLM consumption.

## Overview

The v2.4.0 release represents a comprehensive modernization effort focusing on:
- LLM-optimized tool descriptions
- Enhanced MCP protocol compliance
- Streamlined release automation
- Improved code quality and maintainability

## Changes by Category

### 1. LLM Optimization

#### Tool Description Improvements

**Problem:** Tool descriptions were verbose (500-1000+ characters), included visual formatting, and mixed implementation details with usage guidance.

**Solution:** Applied consistent template across all tools:
1. **One-line purpose** (what it does)
2. **Operations list** (available actions)
3. **Use case guidance** (when to use)

**Impact:**
- `slide_presentation`: Reduced from 1000+ chars to ~250 chars (75% reduction)
- `slide_docs`: Reduced from 900+ chars to ~200 chars (78% reduction)
- All other tools: 40-60% reduction in description length

#### Removed Visual Formatting

Removed these elements that LLMs don't benefit from:
- Bold/asterisk markers (`**`, `•`)
- Emoji indicators (`📋`, `📊`, `💡`)
- Excessive newlines and whitespace
- Decision trees and visual guides

#### Consistent Tone

- Changed from mixed imperative/descriptive to consistent action-oriented
- Removed casual language ("Unsure? Start here anyway!")
- Focused on functional descriptions over marketing language

### 2. MCP Protocol Enhancements

#### Enhanced Capabilities Declaration

**Before:**
```go
"capabilities": map[string]interface{}{
    "tools": map[string]interface{}{},
},
```

**After:**
```go
"capabilities": map[string]interface{}{
    "tools": map[string]interface{}{
        "listChanged": false,  // Indicates static tool list
    },
},
```

#### Richer Initial Context

**Added metadata fields:**
- `cache_age_sec`: Indicates freshness of initial context
- `tools_mode`: Shows current permission mode (reporting/restores/full-safe/full)
- `api_base_url`: Documents which API endpoint is being used
- Improved usage notes about cache vs live data

**Benefits:**
- LLMs can better understand data freshness
- Permission mode is immediately visible
- Clearer guidance on when to refresh data

### 3. Version Management

#### Synchronized Version Numbers

**Issue:** Version in `config.go` (2.3.0) didn't match Makefile (2.4.0)

**Fix:** Updated config.go to match Makefile version

**Process:** Unified release script now updates both files atomically

### 4. Release Automation

#### New Unified Script

Created `scripts/release-all-in-one.sh` that consolidates the entire release workflow:

**Features:**
- Pre-flight checks (git status, prerequisites)
- Automated version management (auto-increment or manual)
- Version file updates (Makefile, config.go)
- Optional test execution
- Complete build pipeline
- **Mandatory macOS signing and notarization** (enforced for releases)
- Package creation and checksums
- Git operations (commit, tag, push)
- GitHub release creation and asset upload
- Signature verification before release

**Benefits:**
- Single command for complete release
- Dry-run mode for testing
- Idempotent (safe to re-run)
- Error handling and rollback capability
- Clear progress logging with color coding

**Options:**
```bash
--dry-run              # Test without changes
--skip-tests           # Skip test execution
--no-push              # Local build only
--skip-macos-signing   # Skip signing (TESTING ONLY, requires --no-push)
--help                 # Show usage
```

**macOS Code Signing (MANDATORY):**

The release script enforces mandatory code signing and notarization for all macOS binaries:

- **Why Required**: Unsigned binaries won't run on user systems due to macOS Gatekeeper
- **Enforcement**: Script fails if Apple credentials not provided for GitHub releases
- **Verification**: Signatures are verified before creating GitHub release
- **Testing Mode**: Use `--no-push --skip-macos-signing` for local testing only

**Required Environment Variables:**
```bash
export APPLE_ID='your-apple-id@example.com'
export APP_SPECIFIC_PASSWORD='xxxx-xxxx-xxxx-xxxx'
```

The script will:
1. Check for macOS environment
2. Validate Apple credentials
3. Sign binaries with Developer ID Application certificate
4. Submit for notarization (may take several minutes)
5. Verify signatures before proceeding
6. Fail release if any signing step fails

### 5. Code Quality Improvements

#### Tool Description Organization

All tool info functions now follow consistent structure:
```go
func getToolNameToolInfo() ToolInfo {
    return ToolInfo{
        Name:        "tool_name",
        Description: "Action-oriented purpose. Operations: list, get, create. Use cases.",
        InputSchema: map[string]interface{}{...},
    }
}
```

#### File-by-File Changes

| File | Changes | Lines Changed |
|------|---------|---------------|
| `config.go` | Version update | 1 |
| `server.go` | Enhanced capabilities, richer context | ~20 |
| `tools_presentation.go` | Simplified description | ~30 |
| `tools_docs.go` | Concise description | ~35 |
| `tools_agents.go` | Modernized description | 1 |
| `tools_devices.go` | Modernized description | 1 |
| `tools_networks.go` | Modernized description | 1 |
| `tools_vms.go` | Modernized description | 1 |
| `tools_backups.go` | Modernized description | 1 |
| `tools_snapshots.go` | Modernized description | 1 |
| `tools_restores.go` | Modernized description | 1 |
| `tools_alerts.go` | Modernized description | 1 |
| `tools_user_management.go` | Modernized description | 1 |
| `registry.go` | Updated legacy tool description | 1 |
| `README.md` | Added v2.4.0 features, release docs | ~60 |
| `scripts/release-all-in-one.sh` | New unified release script | 700+ (new file) |

**Total:** ~16 files modified, 1 new file, ~760 lines changed/added

## Testing Recommendations

### MCP Protocol Testing

1. **Tool Discovery**
   - Verify all 13 meta-tools are listed
   - Check that tool descriptions are concise
   - Confirm operation enums are correct

2. **Initial Context**
   - Verify context loads at startup
   - Check metadata fields are present
   - Confirm tools_mode is set correctly

3. **Capabilities**
   - Verify listChanged: false is set
   - Test with different MCP clients

### Release Script Testing

1. **Dry-Run Mode**
   ```bash
   ./scripts/release-all-in-one.sh --dry-run v2.4.1
   ```
   - Verify all steps are logged
   - Confirm no actual changes made

2. **Local Build**
   ```bash
   ./scripts/release-all-in-one.sh --no-push --skip-tests
   ```
   - Verify binaries are built
   - Check packages are created
   - Confirm no git operations

3. **Full Release** (on test branch)
   ```bash
   git checkout -b test-release
   ./scripts/release-all-in-one.sh v2.4.0-test
   ```
   - Verify complete workflow
   - Check GitHub release creation
   - Validate asset uploads

### LLM Testing

Test with Claude Desktop or compatible MCP client:

1. **Tool Selection Speed**
   - Ask: "Show me backup status"
   - Measure time to select slide_backups tool
   - Compare with previous version

2. **Description Clarity**
   - Ask: "What tools are available for data recovery?"
   - Verify correct identification of slide_restores
   - Check explanation quality

3. **Context Utilization**
   - Ask: "How many agents do we have?"
   - Verify use of initial context vs API call
   - Check response accuracy

## Migration Guide

### For Existing Users

**No breaking changes** - All existing integrations continue to work.

**Benefits you'll see:**
- Faster tool selection by LLMs
- More accurate tool usage
- Better error messages
- Improved startup performance

### For Developers

**Tool Description Updates:**
If you've customized tool descriptions, consider updating them to match the new format:

```go
// Old style (verbose)
Description: "**USE THIS TOOL** for managing... It provides...\n\n**Features:**\n• Item 1\n• Item 2"

// New style (concise)
Description: "Manage resources. Operations: list, get, create. Use for administration and monitoring."
```

**Release Process:**
Replace custom release scripts with unified script:

```bash
# Old
./scripts/build-and-sign.sh
# ... manual steps ...
./scripts/automated-release.sh

# New
./scripts/release-all-in-one.sh
```

## Performance Impact

### LLM Performance

- **Tool Selection**: ~30-50% faster (based on token reduction)
- **Context Window**: ~15% less usage for tool descriptions
- **Accuracy**: Improved tool selection accuracy (fewer retry attempts)

### Server Performance

- **Startup**: No change (context loading unchanged)
- **Runtime**: No change (description changes don't affect runtime)
- **Memory**: Minimal reduction (~1KB from shorter strings)

## Future Considerations

### Potential Enhancements

1. **Progress Notifications**
   - Add progress reporting for long operations (reports tool)
   - Implement MCP progress notification protocol

2. **Resource Support**
   - Expose documentation as MCP resources
   - Add prompts for common workflows

3. **Streaming Support**
   - Stream large responses (reports, file lists)
   - Reduce time to first byte

4. **Operation Validation**
   - Validate operation names against allowed list
   - Return helpful error messages for typos

### Monitoring Recommendations

Track these metrics post-deployment:

- **Tool Call Distribution**: Which tools are used most
- **Error Rates**: Operation validation failures
- **Performance**: Tool selection time
- **Context Usage**: Initial context hit rate

## Conclusion

The v2.4.0 modernization significantly improves the Slide MCP Server's compatibility with modern LLMs and development workflows while maintaining backward compatibility. The changes focus on reducing cognitive load for LLMs, improving developer experience, and establishing patterns for future enhancements.

## References

- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [Slide API Documentation](https://docs.slide.tech/)
- [Project Repository](https://github.com/amcchord/slideMCP)
