#!/bin/bash

# Slide MCP Installer - macOS Signed Release Builder
# This script builds, signs, and notarizes the macOS installer for distribution

set -e  # Exit on any error

# Configuration
BINARY_NAME="slide-mcp-installer"
VERSION="v2.2.0"
BUILD_DIR="build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}INFO:${NC} $1"
}

log_success() {
    echo -e "${GREEN}SUCCESS:${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

log_error() {
    echo -e "${RED}ERROR:${NC} $1"
}

# Check if running on macOS
check_platform() {
    if [[ "$(uname)" != "Darwin" ]]; then
        log_error "This script must be run on macOS"
        exit 1
    fi
}

# Check required environment variables
check_environment() {
    local missing_vars=()
    
    if [[ -z "$DEVELOPER_ID_APPLICATION" ]]; then
        missing_vars+=("DEVELOPER_ID_APPLICATION")
    fi
    
    if [[ -z "$APPLE_ID" ]]; then
        missing_vars+=("APPLE_ID")
    fi
    
    if [[ -z "$APPLE_PASSWORD" ]]; then
        missing_vars+=("APPLE_PASSWORD")
    fi
    
    if [[ -z "$TEAM_ID" ]]; then
        missing_vars+=("TEAM_ID")
    fi
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        echo ""
        echo "Required environment variables:"
        echo "  DEVELOPER_ID_APPLICATION - Your 'Developer ID Application' certificate name"
        echo "  APPLE_ID                 - Your Apple ID email"
        echo "  APPLE_PASSWORD           - App-specific password or @keychain:AC_PASSWORD"
        echo "  TEAM_ID                  - Your Apple Developer Team ID"
        echo ""
        echo "Example setup:"
        echo "  export DEVELOPER_ID_APPLICATION=\"Developer ID Application: Your Name (TEAM_ID)\""
        echo "  export APPLE_ID=\"your-email@example.com\""
        echo "  export APPLE_PASSWORD=\"@keychain:AC_PASSWORD\""
        echo "  export TEAM_ID=\"YOUR_TEAM_ID\""
        exit 1
    fi
}

# Check required tools
check_tools() {
    local missing_tools=()
    
    if ! command -v fyne &> /dev/null; then
        missing_tools+=("fyne")
    fi
    
    if ! command -v codesign &> /dev/null; then
        missing_tools+=("codesign")
    fi
    
    if ! command -v xcrun &> /dev/null; then
        missing_tools+=("xcrun")
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log_error "Missing required tools:"
        for tool in "${missing_tools[@]}"; do
            echo "  - $tool"
        done
        echo ""
        echo "Install missing tools:"
        echo "  - fyne: go install fyne.io/fyne/v2/cmd/fyne@latest"
        echo "  - codesign and xcrun: Install Xcode Command Line Tools"
        exit 1
    fi
}

# Verify signing certificate
verify_certificate() {
    log_info "Verifying code signing certificate..."
    
    if ! security find-identity -v -p codesigning | grep -q "$DEVELOPER_ID_APPLICATION"; then
        log_error "Code signing certificate not found: $DEVELOPER_ID_APPLICATION"
        echo ""
        echo "Available certificates:"
        security find-identity -v -p codesigning
        exit 1
    fi
    
    log_success "Code signing certificate verified"
}

# Build the application
build_app() {
    log_info "Building application..."
    
    # Clean previous builds
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    
    # Generate icon resource
    if [[ -f "MCP-Installer.png" ]]; then
        log_info "Generating icon resource..."
        fyne bundle -o icon.go MCP-Installer.png
    else
        log_warning "Icon file MCP-Installer.png not found"
    fi
    
    # Build the app
    fyne package --target darwin --icon MCP-Installer.png --name "$BINARY_NAME" --source-dir .
    mv "${BINARY_NAME}.app" "$BUILD_DIR/"
    
    log_success "Application built successfully"
}

# Sign the application
sign_app() {
    local app_path="$BUILD_DIR/${BINARY_NAME}.app"
    
    log_info "Signing application..."
    
    codesign --force --deep --sign "$DEVELOPER_ID_APPLICATION" --options runtime "$app_path"
    
    log_info "Verifying code signature..."
    codesign --verify --verbose "$app_path"
    
    log_success "Application signed successfully"
}

# Notarize the application
notarize_app() {
    local app_path="$BUILD_DIR/${BINARY_NAME}.app"
    local zip_path="$BUILD_DIR/${BINARY_NAME}.zip"
    
    log_info "Creating zip archive for notarization..."
    cd "$BUILD_DIR" && zip -r "${BINARY_NAME}.zip" "${BINARY_NAME}.app"
    cd - > /dev/null
    
    log_info "Submitting application for notarization..."
    log_info "This may take several minutes..."
    
    xcrun notarytool submit "$zip_path" \
        --apple-id "$APPLE_ID" \
        --password "$APPLE_PASSWORD" \
        --team-id "$TEAM_ID" \
        --wait
    
    log_info "Stapling notarization ticket to application..."
    xcrun stapler staple "$app_path"
    
    log_info "Verifying notarization..."
    xcrun stapler validate "$app_path"
    
    # Clean up zip file
    rm "$zip_path"
    
    log_success "Application notarized successfully"
}

# Create release package
create_release_package() {
    local app_path="$BUILD_DIR/${BINARY_NAME}.app"
    local archive_name="${BINARY_NAME}-${VERSION}-darwin-$(uname -m)-signed.tar.gz"
    
    log_info "Creating release package..."
    
    cd "$BUILD_DIR"
    tar -czf "$archive_name" "${BINARY_NAME}.app"
    cd - > /dev/null
    
    log_success "Release package created: $BUILD_DIR/$archive_name"
}

# Main execution
main() {
    log_info "Starting macOS signed release build process..."
    
    check_platform
    check_environment
    check_tools
    verify_certificate
    
    build_app
    sign_app
    notarize_app
    create_release_package
    
    log_success "Signed release build completed successfully!"
    echo ""
    echo "Output:"
    echo "  App Bundle: $BUILD_DIR/${BINARY_NAME}.app"
    echo "  Release Package: $BUILD_DIR/${BINARY_NAME}-${VERSION}-darwin-$(uname -m)-signed.tar.gz"
}

# Run main function
main "$@" 