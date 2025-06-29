#!/bin/bash

# Automated Release Script for Slide MCP Server
# This script automates the entire release process including building, signing, 
# notarizing, packaging, and uploading to GitHub

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_DIR/build"
BINARY_NAME="slide-mcp-server"

# Version will be passed as argument or auto-incremented
NEW_VERSION="${1:-}"

# Apple credentials for notarization (must be set via environment variables)
APPLE_ID="${APPLE_ID:-}"
APP_SPECIFIC_PASSWORD="${APP_SPECIFIC_PASSWORD:-}"
KEYCHAIN_PROFILE="slide-mcp-release"

# GitHub configuration
GITHUB_REPO="amcchord/slideMCP"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if on macOS for code signing
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "This script must be run on macOS for code signing and notarization"
        exit 1
    fi
    
    # Check for required tools
    local missing_tools=()
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if ! command -v codesign &> /dev/null; then
        missing_tools+=("codesign (Xcode Command Line Tools)")
    fi
    
    if ! command -v xcrun &> /dev/null; then
        missing_tools+=("xcrun (Xcode Command Line Tools)")
    fi
    
    if ! command -v gh &> /dev/null; then
        missing_tools+=("gh (GitHub CLI)")
    fi
    
    if ! command -v jq &> /dev/null; then
        missing_tools+=("jq")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools:"
        for tool in "${missing_tools[@]}"; do
            echo "  - $tool"
        done
        log_info "Install missing tools and try again"
        exit 1
    fi
    
    # Check GitHub CLI authentication
    if ! gh auth status &> /dev/null; then
        log_error "GitHub CLI not authenticated. Run 'gh auth login' first"
        exit 1
    fi
    
    # Check Apple credentials for notarization
    if [ -z "$APPLE_ID" ]; then
        log_error "APPLE_ID environment variable is required for notarization"
        log_info "Set it with: export APPLE_ID='your-apple-id@example.com'"
        exit 1
    fi
    
    if [ -z "$APP_SPECIFIC_PASSWORD" ]; then
        log_error "APP_SPECIFIC_PASSWORD environment variable is required for notarization"
        log_info "Set it with: export APP_SPECIFIC_PASSWORD='your-app-specific-password'"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to determine next version
determine_version() {
    if [ -n "$NEW_VERSION" ]; then
        log_info "Using provided version: $NEW_VERSION"
        return
    fi
    
    # Get current version from Makefile
    local current_version
    current_version=$(grep "^VERSION=" "$PROJECT_DIR/Makefile" | cut -d'=' -f2)
    
    log_info "Current version: $current_version"
    
    # Extract version number and increment
    local version_num
    version_num=$(echo "$current_version" | sed 's/v//')
    local major minor patch
    IFS='.' read -r major minor patch <<< "$version_num"
    
    # Increment patch version
    patch=$((patch + 1))
    NEW_VERSION="v$major.$minor.$patch"
    
    log_info "Auto-incremented version: $NEW_VERSION"
}

# Function to update version in files
update_version_files() {
    log_info "Updating version to $NEW_VERSION in project files..."
    
    # Update Makefile
    sed -i.bak "s/^VERSION=.*/VERSION=$NEW_VERSION/" "$PROJECT_DIR/Makefile"
    
    # Update build-and-sign.sh
    sed -i.bak "s/^VERSION=.*/VERSION=\"$NEW_VERSION\"/" "$PROJECT_DIR/scripts/build-and-sign.sh"
    
    # Clean up backup files
    find "$PROJECT_DIR" -name "*.bak" -delete
    
    log_success "Version updated to $NEW_VERSION"
}

# Function to setup notarization keychain profile
setup_notarization() {
    log_info "Setting up notarization keychain profile..."
    
    # Check if profile already exists
    if xcrun notarytool history --keychain-profile "$KEYCHAIN_PROFILE" &> /dev/null; then
        log_info "Keychain profile '$KEYCHAIN_PROFILE' already exists"
        return
    fi
    
    # Create new keychain profile
    log_info "Creating new keychain profile for notarization..."
    xcrun notarytool store-credentials "$KEYCHAIN_PROFILE" \
        --apple-id "$APPLE_ID" \
        --password "$APP_SPECIFIC_PASSWORD" \
        --team-id "$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1 | sed -n 's/.*(\([^)]*\)).*/\1/p')"
    
    log_success "Notarization keychain profile created"
}

# Function to find signing identity
find_signing_identity() {
    log_info "Finding code signing identity..."
    
    local identities
    identities=$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1)
    
    if [ -z "$identities" ]; then
        log_error "No Developer ID Application certificate found"
        log_info "Please install a Developer ID Application certificate from your Apple Developer account"
        exit 1
    fi
    
    # Extract the identity name
    DEVELOPER_ID=$(echo "$identities" | sed -n 's/.*"\(Developer ID Application: [^"]*\)".*/\1/p')
    
    if [ -z "$DEVELOPER_ID" ]; then
        log_error "Could not parse Developer ID from certificate"
        exit 1
    fi
    
    log_success "Found signing identity: $DEVELOPER_ID"
}

# Function to clean build directory
clean_build() {
    log_info "Cleaning build directory..."
    if [ -d "$BUILD_DIR" ]; then
        rm -rf "$BUILD_DIR"
    fi
    mkdir -p "$BUILD_DIR"
    log_success "Build directory cleaned"
}

# Function to build all platforms
build_all_platforms() {
    log_info "Building for all platforms..."
    
    cd "$PROJECT_DIR"
    
    # Use make to build all platforms
    make clean
    make build-all
    
    log_success "All platforms built successfully"
}

# Function to sign macOS binaries
sign_macos_binaries() {
    log_info "Signing macOS binaries..."
    
    cd "$PROJECT_DIR"
    
    # Sign using Makefile
    DEVELOPER_ID="$DEVELOPER_ID" make sign-macos BINARY="$BUILD_DIR/$BINARY_NAME-darwin-amd64"
    DEVELOPER_ID="$DEVELOPER_ID" make sign-macos BINARY="$BUILD_DIR/$BINARY_NAME-darwin-arm64"
    
    log_success "macOS binaries signed successfully"
}

# Function to notarize macOS binaries
notarize_macos_binaries() {
    log_info "Notarizing macOS binaries..."
    
    cd "$PROJECT_DIR"
    
    # Notarize using our keychain profile
    KEYCHAIN_PROFILE="$KEYCHAIN_PROFILE" make notarize-macos
    
    log_success "macOS binaries notarized successfully"
}

# Function to create release packages
create_release_packages() {
    log_info "Creating release packages..."
    
    cd "$BUILD_DIR"
    
    # Create .tar.gz for Unix-like systems (Linux and macOS)
    if [ -f "$BINARY_NAME-linux-amd64" ]; then
        tar -czf "$BINARY_NAME-$NEW_VERSION-linux-x64.tar.gz" "$BINARY_NAME-linux-amd64"
        log_success "Created Linux x64 package"
    fi
    
    if [ -f "$BINARY_NAME-linux-arm64" ]; then
        tar -czf "$BINARY_NAME-$NEW_VERSION-linux-arm64.tar.gz" "$BINARY_NAME-linux-arm64"
        log_success "Created Linux ARM64 package"
    fi
    
    if [ -f "$BINARY_NAME-darwin-amd64" ]; then
        tar -czf "$BINARY_NAME-$NEW_VERSION-macos-x64.tar.gz" "$BINARY_NAME-darwin-amd64"
        log_success "Created macOS x64 package"
    fi
    
    if [ -f "$BINARY_NAME-darwin-arm64" ]; then
        tar -czf "$BINARY_NAME-$NEW_VERSION-macos-arm64.tar.gz" "$BINARY_NAME-darwin-arm64"
        log_success "Created macOS ARM64 package"
    fi
    
    # Create .zip for Windows
    if [ -f "$BINARY_NAME-windows-amd64.exe" ]; then
        zip "$BINARY_NAME-$NEW_VERSION-windows-x64.zip" "$BINARY_NAME-windows-amd64.exe"
        log_success "Created Windows x64 package"
    fi
    
    cd "$PROJECT_DIR"
}

# Function to generate checksums
generate_checksums() {
    log_info "Generating checksums..."
    
    cd "$BUILD_DIR"
    
    # Generate SHA256 checksums for release packages
    find . -name "*.tar.gz" -o -name "*.zip" | sort | while read -r file; do
        shasum -a 256 "$file" >> checksums.sha256
    done
    
    log_success "Checksums generated"
    cd "$PROJECT_DIR"
}

# Function to create release notes
create_release_notes() {
    log_info "Creating release notes..."
    
    local release_notes_file="$BUILD_DIR/RELEASE_NOTES.md"
    
    # Get recent commits since last version tag
    local last_version
    last_version=$(git tag --sort=-version:refname | grep -E '^v[0-9]+\.[0-9]+' | head -1 || echo "")
    
    local commits
    if [ -n "$last_version" ]; then
        commits=$(git log "$last_version..HEAD" --oneline --no-merges | head -10) || commits=""
    else
        commits=$(git log --oneline --no-merges | head -10) || commits=""
    fi
    
    cat > "$release_notes_file" << EOF
# Slide MCP Server $NEW_VERSION

## Changes

EOF
    
    if [ -n "$commits" ]; then
        echo "$commits" | while IFS= read -r line; do
            if [ -n "$line" ]; then
                echo "- $line" >> "$release_notes_file"
            fi
        done
    else
        echo "- Various improvements and bug fixes" >> "$release_notes_file"
    fi
    
    cat >> "$release_notes_file" << EOF

## Installation

Download the appropriate binary for your platform:

- **Linux x64**: slide-mcp-server-$NEW_VERSION-linux-x64.tar.gz
- **Linux ARM64**: slide-mcp-server-$NEW_VERSION-linux-arm64.tar.gz  
- **macOS x64**: slide-mcp-server-$NEW_VERSION-macos-x64.tar.gz
- **macOS ARM64**: slide-mcp-server-$NEW_VERSION-macos-arm64.tar.gz
- **Windows x64**: slide-mcp-server-$NEW_VERSION-windows-x64.zip

## Verification

Verify the integrity of your download using the checksums.sha256 file:

\`\`\`bash
shasum -a 256 -c checksums.sha256
\`\`\`

## macOS Security

The macOS binaries are signed and notarized by Apple. They should run without security warnings on macOS 10.15+ systems.

For older macOS versions or if you encounter security warnings, you may need to run:

\`\`\`bash
xattr -d com.apple.quarantine slide-mcp-server
\`\`\`
EOF
    
    log_success "Release notes created"
}

# Function to create release directory structure
create_release_directory() {
    log_info "Creating release directory structure..."
    
    local release_dir="$PROJECT_DIR/release/$NEW_VERSION"
    mkdir -p "$release_dir"
    
    # Copy release artifacts
    cp "$BUILD_DIR"/*.tar.gz "$release_dir/" 2>/dev/null || true
    cp "$BUILD_DIR"/*.zip "$release_dir/" 2>/dev/null || true
    cp "$BUILD_DIR/checksums.sha256" "$release_dir/" 2>/dev/null || true
    cp "$BUILD_DIR/RELEASE_NOTES.md" "$release_dir/" 2>/dev/null || true
    
    # Also copy individual binaries for compatibility
    cp "$BUILD_DIR/$BINARY_NAME-linux-amd64" "$release_dir/$BINARY_NAME-$NEW_VERSION-linux-x64" 2>/dev/null || true
    cp "$BUILD_DIR/$BINARY_NAME-linux-arm64" "$release_dir/$BINARY_NAME-$NEW_VERSION-linux-arm64" 2>/dev/null || true
    cp "$BUILD_DIR/$BINARY_NAME-darwin-amd64" "$release_dir/$BINARY_NAME-$NEW_VERSION-macos-x64" 2>/dev/null || true
    cp "$BUILD_DIR/$BINARY_NAME-darwin-arm64" "$release_dir/$BINARY_NAME-$NEW_VERSION-macos-arm64" 2>/dev/null || true
    cp "$BUILD_DIR/$BINARY_NAME-windows-amd64.exe" "$release_dir/$BINARY_NAME-$NEW_VERSION-windows-x64.exe" 2>/dev/null || true
    
    log_success "Release directory created: $release_dir"
}

# Function to create GitHub release
create_github_release() {
    log_info "Creating GitHub release..."
    
    local release_dir="$PROJECT_DIR/release/$NEW_VERSION"
    
    # Create the release
    gh release create "$NEW_VERSION" \
        --title "Slide MCP Server $NEW_VERSION" \
        --notes-file "$BUILD_DIR/RELEASE_NOTES.md" \
        --draft=false \
        --prerelease=false
    
    # Upload assets
    local assets=()
    
    # Add packaged files
    if [ -f "$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-linux-x64.tar.gz" ]; then
        assets+=("$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-linux-x64.tar.gz")
    fi
    
    if [ -f "$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-linux-arm64.tar.gz" ]; then
        assets+=("$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-linux-arm64.tar.gz")
    fi
    
    if [ -f "$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-macos-x64.tar.gz" ]; then
        assets+=("$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-macos-x64.tar.gz")
    fi
    
    if [ -f "$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-macos-arm64.tar.gz" ]; then
        assets+=("$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-macos-arm64.tar.gz")
    fi
    
    if [ -f "$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-windows-x64.zip" ]; then
        assets+=("$BUILD_DIR/$BINARY_NAME-$NEW_VERSION-windows-x64.zip")
    fi
    
    # Add checksums
    if [ -f "$BUILD_DIR/checksums.sha256" ]; then
        assets+=("$BUILD_DIR/checksums.sha256")
    fi
    
    # Upload all assets
    if [ ${#assets[@]} -gt 0 ]; then
        gh release upload "$NEW_VERSION" "${assets[@]}"
    fi
    
    log_success "GitHub release created: https://github.com/$GITHUB_REPO/releases/tag/$NEW_VERSION"
}

# Function to commit and tag
commit_and_tag() {
    log_info "Committing version changes and creating tag..."
    
    cd "$PROJECT_DIR"
    
    # Add changed files
    git add Makefile scripts/build-and-sign.sh "release/$NEW_VERSION/"
    
    # Commit
    git commit -m "Release $NEW_VERSION"
    
    # Create tag
    git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
    
    # Push changes and tag
    git push origin main
    git push origin "$NEW_VERSION"
    
    log_success "Changes committed and tagged"
}

# Function to show summary
show_summary() {
    echo
    log_success "Release $NEW_VERSION completed successfully!"
    echo "=============================================="
    echo "Version: $NEW_VERSION"
    echo "Build Directory: $BUILD_DIR"
    echo "Release Directory: $PROJECT_DIR/release/$NEW_VERSION"
    echo "GitHub Release: https://github.com/$GITHUB_REPO/releases/tag/$NEW_VERSION"
    echo
    echo "Release artifacts:"
    ls -la "$BUILD_DIR"/*.tar.gz "$BUILD_DIR"/*.zip "$BUILD_DIR/checksums.sha256" 2>/dev/null || echo "No artifacts found"
    echo
    echo "Next steps:"
    echo "1. Test the release binaries on different platforms"
    echo "2. Update documentation if needed"
    echo "3. Announce the release"
}

# Main execution function
main() {
    echo "Starting automated release process..."
    echo "====================================="
    
    # Check prerequisites
    check_prerequisites
    
    # Determine version
    determine_version
    
    # Update version files
    update_version_files
    
    # Setup notarization
    setup_notarization
    
    # Find signing identity
    find_signing_identity
    
    # Clean and build
    clean_build
    build_all_platforms
    
    # Sign and notarize macOS binaries
    sign_macos_binaries
    notarize_macos_binaries
    
    # Create packages and checksums
    create_release_packages
    generate_checksums
    
    # Create release notes
    create_release_notes
    
    # Create local release directory
    create_release_directory
    
    # Commit and tag
    commit_and_tag
    
    # Create GitHub release
    create_github_release
    
    # Show summary
    show_summary
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Slide MCP Server Automated Release Script"
        echo
        echo "Usage: $0 [version]"
        echo
        echo "Arguments:"
        echo "  version    Optional. Specify version (e.g., v1.18.0)"
        echo "             If not provided, patch version will be auto-incremented"
        echo
        echo "Environment Variables (Required):"
        echo "  APPLE_ID               Apple ID for notarization"
        echo "  APP_SPECIFIC_PASSWORD  App-specific password for notarization"
        echo
        echo "Examples:"
        echo "  export APPLE_ID='your-apple-id@example.com'"
        echo "  export APP_SPECIFIC_PASSWORD='your-app-specific-password'"
        echo "  $0              # Auto-increment patch version"
        echo "  $0 v1.18.0      # Use specific version"
        echo
        echo "Prerequisites:"
        echo "  - macOS (for code signing and notarization)"
        echo "  - Xcode Command Line Tools"
        echo "  - GitHub CLI (gh) authenticated"
        echo "  - Developer ID Application certificate installed"
        echo "  - Go development environment"
        exit 0
        ;;
    *)
        main
        ;;
esac 