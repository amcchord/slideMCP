# Automated Release Process

This document describes how to use the automated release script for Slide MCP Server.

## Quick Start

First, set your Apple credentials as environment variables:

```bash
export APPLE_ID='your-apple-id@example.com'
export APP_SPECIFIC_PASSWORD='your-app-specific-password'
```

Or use the provided template:

```bash
# Copy and customize the template
cp scripts/release-env.template scripts/release-env.sh
# Edit scripts/release-env.sh with your credentials
# Source the environment
source scripts/release-env.sh
```

To create a new release with auto-incremented version:

```bash
./scripts/automated-release.sh
```

To create a release with a specific version:

```bash
./scripts/automated-release.sh v1.18.0
```

## What the Script Does

The automated release script performs the following steps:

1. **Version Management**: Auto-increments patch version or uses provided version
2. **Build**: Cross-compiles binaries for all supported platforms:
   - Linux x64 and ARM64
   - macOS x64 and ARM64 (Apple Silicon)
   - Windows x64
3. **Code Signing**: Signs macOS binaries with Developer ID Application certificate
4. **Notarization**: Submits macOS binaries to Apple for notarization
5. **Packaging**: Creates appropriate packages:
   - `.tar.gz` for Linux and macOS
   - `.zip` for Windows
6. **Checksums**: Generates SHA256 checksums for all packages
7. **Release Notes**: Auto-generates release notes from recent commits
8. **Git Operations**: Commits changes, creates tags, and pushes to GitHub
9. **GitHub Release**: Creates GitHub release and uploads all assets

## Prerequisites

- **macOS**: Required for code signing and notarization
- **Xcode Command Line Tools**: `xcode-select --install`
- **Developer ID Application Certificate**: Installed from Apple Developer account
- **GitHub CLI**: `brew install gh` and authenticated with `gh auth login`
- **Go**: Development environment
- **jq**: JSON processor (`brew install jq`)

## Apple Developer Setup

The script requires Apple ID credentials to be provided via environment variables for security:

```bash
export APPLE_ID='your-apple-id@example.com'
export APP_SPECIFIC_PASSWORD='your-app-specific-password'
```

These credentials will be used to create a keychain profile for notarization. The script will fail if these environment variables are not set.

## Output

The script creates:

### Local Files
- `build/` directory with all binaries and packages
- `release/vX.X.X/` directory with release artifacts
- Updated `Makefile` and `scripts/build-and-sign.sh` with new version

### GitHub Release
- Release page at `https://github.com/amcchord/slideMCP/releases/tag/vX.X.X`
- All packaged binaries as downloadable assets
- SHA256 checksums file
- Auto-generated release notes

### Git Changes
- Version bump commit
- Git tag for the new version
- Pushed to origin/main

## Package Formats

As requested:
- **Linux/macOS**: `.tar.gz` archives
- **Windows**: `.zip` archives
- **Individual binaries**: Also available in release directory
- **Checksums**: SHA256 hashes for verification

## Release Notes

The script automatically generates release notes with:
- Recent commits since the last version tag
- Installation instructions for each platform
- Verification instructions using checksums
- macOS security information for signed/notarized binaries

## Troubleshooting

### Common Issues

1. **Code Signing Fails**: Ensure Developer ID Application certificate is installed
2. **Notarization Fails**: Check Apple ID credentials and app-specific password
3. **GitHub Upload Fails**: Verify GitHub CLI authentication with `gh auth status`
4. **Build Fails**: Ensure Go environment is properly set up

### Logs

The script provides detailed colored logging:
- **Blue [INFO]**: General information
- **Green [SUCCESS]**: Successful operations
- **Yellow [WARNING]**: Warnings (non-fatal)
- **Red [ERROR]**: Errors that will cause failure

## Example Output

```
Starting automated release process...
=====================================
[INFO] Checking prerequisites...
[SUCCESS] Prerequisites check passed
[INFO] Current version: v1.17.2
[INFO] Auto-incremented version: v1.17.3
...
[SUCCESS] Release v1.17.3 completed successfully!
GitHub Release: https://github.com/amcchord/slideMCP/releases/tag/v1.17.3
```

## Script Location

The script is located at: `scripts/automated-release.sh`

It's designed to be idempotent and can be safely re-run if issues occur during the release process. 