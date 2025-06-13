# macOS Code Signing and Notarization

This document explains how to set up and use code signing and notarization for the Slide MCP Installer on macOS.

## Prerequisites

### Apple Developer Account
You need an active Apple Developer account ($99/year) to:
- Obtain code signing certificates
- Submit apps for notarization

### Required Certificates
You need a "Developer ID Application" certificate:
1. Log into [Apple Developer Portal](https://developer.apple.com/account/)
2. Go to Certificates, Identifiers & Profiles
3. Create a new certificate: Developer ID Application
4. Download and install the certificate in your macOS Keychain

### App-Specific Password
Create an app-specific password for notarization:
1. Sign in to [appleid.apple.com](https://appleid.apple.com/)
2. Go to Sign-In and Security → App-Specific Passwords
3. Generate a new password (save it securely)

## Environment Setup

Set up the required environment variables:

```bash
# Your code signing certificate name (find it with: security find-identity -v -p codesigning)
export DEVELOPER_ID_APPLICATION="Developer ID Application: Your Name (TEAM_ID)"

# Your Apple ID email
export APPLE_ID="your-email@example.com"

# App-specific password (or keychain reference)
export APPLE_PASSWORD="abcd-efgh-ijkl-mnop"
# OR store in keychain and use:
# export APPLE_PASSWORD="@keychain:AC_PASSWORD"

# Your Apple Developer Team ID
export TEAM_ID="YOUR_TEAM_ID"
```

### Storing Credentials in Keychain (Recommended)

For security, store your app-specific password in keychain:

```bash
# Add password to keychain
security add-generic-password -a "your-email@example.com" -w "abcd-efgh-ijkl-mnop" -s "AC_PASSWORD"

# Use keychain reference in environment
export APPLE_PASSWORD="@keychain:AC_PASSWORD"
```

## Usage

### Method 1: Using Make Targets

Build, sign, and notarize in one step:
```bash
make build-release
```

Or run individual steps:
```bash
# Build and sign only
make build-signed

# Notarize previously built app
make notarize

# Create signed release package
make release-signed
```

### Method 2: Using the Build Script

The standalone script provides more detailed output and error checking:

```bash
# Run the complete build, sign, and notarization process
./build-signed-release.sh
```

## Build Targets

### Development Builds (Unsigned)
- `make build` - Build unsigned app for development/testing
- `make run` - Run the installer locally
- `make open` - Build and open the app

### Release Builds (Signed & Notarized)
- `make build-signed` - Build and code sign the app
- `make notarize` - Notarize the signed app (requires build-signed first)
- `make build-release` - Complete build, sign, and notarization process
- `make release-signed` - Create distributable release package

## Troubleshooting

### Common Issues

**Certificate not found:**
```
Error: Code signing certificate not found
```
- Check your certificate name with: `security find-identity -v -p codesigning`
- Ensure the certificate is installed in your login keychain
- Verify the certificate hasn't expired

**Notarization failed:**
```
Error: Notarization submission failed
```
- Verify your Apple ID and app-specific password
- Check your team ID is correct
- Ensure your Apple Developer account is active

**App won't run on other Macs:**
```
"slide-mcp-installer.app" cannot be opened because the developer cannot be verified
```
- The app needs to be properly signed and notarized
- Users may need to right-click → Open for the first run

### Checking Certificate Details

List available certificates:
```bash
security find-identity -v -p codesigning
```

Verify app signature:
```bash
codesign --verify --verbose build/slide-mcp-installer.app
```

Check notarization status:
```bash
xcrun stapler validate build/slide-mcp-installer.app
```

## Automation

### CI/CD Integration

For automated builds, set environment variables in your CI system:

```yaml
# GitHub Actions example
env:
  DEVELOPER_ID_APPLICATION: ${{ secrets.DEVELOPER_ID_APPLICATION }}
  APPLE_ID: ${{ secrets.APPLE_ID }}
  APPLE_PASSWORD: ${{ secrets.APPLE_PASSWORD }}
  TEAM_ID: ${{ secrets.TEAM_ID }}
```

### Security Best Practices

1. Never commit certificates or passwords to version control
2. Use app-specific passwords, not your main Apple ID password
3. Store sensitive values in CI secrets or keychain
4. Regularly rotate app-specific passwords
5. Monitor certificate expiration dates

## Resources

- [Apple Code Signing Guide](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [Notarization Documentation](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [Xcode Help - Code Signing](https://help.apple.com/xcode/mac/current/#/dev60b6fbbc7) 