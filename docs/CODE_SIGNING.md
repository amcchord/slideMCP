# macOS Code Signing Guide for Slide MCP Server

This guide explains how to set up code signing for macOS binaries to avoid Gatekeeper warnings.

## Prerequisites

1. **Apple Developer Account** - Required for Developer ID certificates and notarization
2. **Xcode Command Line Tools** - Install with `xcode-select --install`
3. **macOS development machine** - Code signing must be done on macOS

## Step 1: Obtain Developer ID Certificate

### Option A: Using Xcode (Recommended)
1. Open Xcode
2. Go to **Xcode > Preferences > Accounts**
3. Add your Apple ID associated with your Developer Account
4. Select your team and click **Download Manual Profiles**
5. Go to **Manage Certificates** and click the **+** button
6. Select **Developer ID Application**

### Option B: Using Developer Portal
1. Visit [Apple Developer Portal](https://developer.apple.com/account/resources/certificates)
2. Create a new **Developer ID Application** certificate
3. Download and install it by double-clicking the `.cer` file

## Step 2: Verify Certificate Installation

```bash
# List available signing identities
security find-identity -v -p codesigning

# You should see something like:
# 1) ABCDEF1234567890... "Developer ID Application: Your Name (TEAM_ID)"
```

## Step 3: Set Up Notarization Credentials

### Create App-Specific Password
1. Go to [Apple ID Account Page](https://appleid.apple.com/)
2. Sign in and go to **Security > App-Specific Passwords**
3. Generate a new password for "slide-mcp-notarization"
4. Save this password securely

### Store Credentials in Keychain
```bash
# Store notarization credentials
xcrun notarytool store-credentials "slide-mcp-profile" \
  --apple-id "your-apple-id@example.com" \
  --team-id "YOUR_TEAM_ID" \
  --password "your-app-specific-password"
```

## Step 4: Build and Sign

### Basic Signed Build
```bash
# Build with code signing
make build-all-signed DEVELOPER_ID="Developer ID Application: Your Name (TEAM_ID)"
```

### Full Signed and Notarized Build
```bash
# Build, sign, and notarize
make build-all-signed DEVELOPER_ID="Developer ID Application: Your Name (TEAM_ID)"
make notarize-macos KEYCHAIN_PROFILE="slide-mcp-profile"
```

### Complete Release Process
```bash
# Create fully signed and notarized release packages
make release-signed \
  DEVELOPER_ID="Developer ID Application: Your Name (TEAM_ID)" \
  KEYCHAIN_PROFILE="slide-mcp-profile"
```

## Step 5: Verify Signing

### Check Code Signature
```bash
# Verify signature
codesign --verify --verbose=2 build/slide-mcp-server-darwin-arm64

# Check signing details
codesign --display --verbose=4 build/slide-mcp-server-darwin-arm64

# Test Gatekeeper assessment
spctl --assess --type execute --verbose=2 build/slide-mcp-server-darwin-arm64
```

### Test Notarization
```bash
# Check notarization status
xcrun stapler validate build/slide-mcp-server-darwin-arm64
```

## Environment Variables

You can set these environment variables to avoid repeating parameters:

```bash
# Add to your ~/.zshrc or ~/.bash_profile
export DEVELOPER_ID="Developer ID Application: Your Name (TEAM_ID)"
export KEYCHAIN_PROFILE="slide-mcp-profile"

# Then just run:
make build-all-signed
make notarize-macos
```

## Troubleshooting

### Common Issues

#### 1. Certificate Not Found
```
Error: No signing identity found
```
**Solution**: Ensure your Developer ID certificate is installed in your Keychain Access.

#### 2. Notarization Failed
```
Error: The software asset has already been uploaded
```
**Solution**: Each binary version must be unique. Update the version or rebuild.

#### 3. Gatekeeper Still Blocks
```
"slide-mcp-server" cannot be opened because the developer cannot be verified
```
**Solution**: Ensure both signing AND notarization were successful. Check with `spctl --assess`.

### Debug Commands

```bash
# Check keychain for certificates
security find-identity -v -p codesigning

# List keychain profiles
xcrun notarytool list-submissions --keychain-profile "slide-mcp-profile"

# Get detailed notarization info
xcrun notarytool info SUBMISSION_ID --keychain-profile "slide-mcp-profile"
```

## Automated CI/CD

For GitHub Actions or other CI/CD systems, you'll need to:

1. **Store certificates as secrets** - Export p12 certificates
2. **Import certificates in CI** - Use security commands
3. **Set environment variables** - For signing identity and credentials
4. **Install certificates** - Before building

Example secrets needed:
- `DEVELOPER_ID_P12_BASE64` - Base64 encoded p12 certificate
- `DEVELOPER_ID_P12_PASSWORD` - Certificate password
- `APPLE_ID` - Your Apple ID
- `APPLE_TEAM_ID` - Your team ID
- `APPLE_APP_PASSWORD` - App-specific password

## Security Best Practices

1. **Protect your certificates** - Store p12 files securely
2. **Use app-specific passwords** - Never use your main Apple ID password
3. **Rotate credentials regularly** - Update app-specific passwords periodically
4. **Limit certificate access** - Only install on necessary machines
5. **Monitor notarization** - Check Apple's notarization logs regularly

## Cost Considerations

- **Apple Developer Program**: $99/year
- **Code signing**: Free with Developer account
- **Notarization**: Free with Developer account
- **Certificate renewal**: Automatic with active Developer account

## Additional Resources

- [Apple Code Signing Guide](https://developer.apple.com/library/archive/documentation/Security/Conceptual/CodeSigningGuide/)
- [Notarization Documentation](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [Gatekeeper and XProtect](https://support.apple.com/guide/security/gatekeeper-and-runtime-protection-sec5599b66df/web) 