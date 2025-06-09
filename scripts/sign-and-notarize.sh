#!/bin/bash

# Slide MCP Server - macOS Code Signing and Notarization Script
# This script automates the process of signing and notarizing macOS binaries

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_DIR/build"
BINARY_NAME="slide-mcp-server"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
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

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    log_error "This script must be run on macOS for code signing and notarization"
    exit 1
fi

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check for Xcode command line tools
    if ! command -v codesign &> /dev/null; then
        log_error "Xcode command line tools not found. Install with: xcode-select --install"
        exit 1
    fi
    
    # Check for notarytool
    if ! command -v xcrun &> /dev/null; then
        log_error "xcrun not found. Please install Xcode command line tools"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to find signing identity
find_signing_identity() {
    log_info "Looking for signing identities..."
    
    # List all Developer ID Application certificates
    IDENTITIES=$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1)
    
    if [ -z "$IDENTITIES" ]; then
        log_error "No Developer ID Application certificate found"
        log_info "Please install a Developer ID Application certificate from your Apple Developer account"
        exit 1
    fi
    
    # Extract the identity name
    DEVELOPER_ID=$(echo "$IDENTITIES" | sed -n 's/.*"\(Developer ID Application: [^"]*\)".*/\1/p')
    
    if [ -z "$DEVELOPER_ID" ]; then
        log_error "Could not parse Developer ID from certificate"
        exit 1
    fi
    
    log_success "Found signing identity: $DEVELOPER_ID"
    echo "$DEVELOPER_ID"
}

# Function to check keychain profiles
check_keychain_profiles() {
    log_info "Checking for notarization keychain profiles..."
    
    # Try to list profiles (this will fail if none exist, but that's ok)
    PROFILES=$(xcrun notarytool list-submissions --keychain-profile "slide-mcp-profile" 2>/dev/null | head -1 || true)
    
    if [ -z "$PROFILES" ]; then
        log_warning "No 'slide-mcp-profile' keychain profile found"
        log_info "You can create one with:"
        log_info "xcrun notarytool store-credentials 'slide-mcp-profile' \\"
        log_info "  --apple-id 'your-apple-id@example.com' \\"
        log_info "  --team-id 'YOUR_TEAM_ID' \\"
        log_info "  --password 'your-app-specific-password'"
        return 1
    fi
    
    log_success "Found keychain profile: slide-mcp-profile"
    return 0
}

# Function to build binaries
build_binaries() {
    log_info "Building binaries..."
    
    cd "$PROJECT_DIR"
    
    if [ ! -f "Makefile" ]; then
        log_error "Makefile not found in project directory"
        exit 1
    fi
    
    # Build all platforms
    make build-all
    
    log_success "Binaries built successfully"
}

# Function to sign macOS binaries
sign_binaries() {
    local developer_id="$1"
    
    log_info "Signing macOS binaries..."
    
    local signed_count=0
    
    for binary in "$BUILD_DIR/$BINARY_NAME-darwin-"*; do
        if [ -f "$binary" ]; then
            log_info "Signing $(basename "$binary")..."
            
            codesign --sign "$developer_id" \
                --options runtime \
                --timestamp \
                --verbose=2 \
                "$binary"
            
            # Verify the signature
            codesign --verify --verbose=2 "$binary"
            spctl --assess --type execute --verbose=2 "$binary"
            
            log_success "Successfully signed $(basename "$binary")"
            ((signed_count++))
        fi
    done
    
    if [ $signed_count -eq 0 ]; then
        log_error "No macOS binaries found to sign"
        exit 1
    fi
    
    log_success "Signed $signed_count macOS binaries"
}

# Function to notarize binaries
notarize_binaries() {
    local keychain_profile="$1"
    
    log_info "Creating ZIP archives for notarization..."
    
    cd "$BUILD_DIR"
    
    local archives=()
    
    for binary in "$BINARY_NAME-darwin-"*; do
        if [ -f "$binary" ] && [[ "$binary" != *.zip ]]; then
            local archive="${binary}.zip"
            zip "$archive" "$binary"
            archives+=("$archive")
            log_success "Created $archive"
        fi
    done
    
    if [ ${#archives[@]} -eq 0 ]; then
        log_error "No macOS binaries found to notarize"
        exit 1
    fi
    
    log_info "Submitting for notarization..."
    
    for archive in "${archives[@]}"; do
        log_info "Submitting $archive..."
        
        xcrun notarytool submit "$archive" \
            --keychain-profile "$keychain_profile" \
            --wait
        
        log_success "Notarization completed for $archive"
    done
    
    log_info "Stapling notarization tickets..."
    
    for binary in "$BINARY_NAME-darwin-"*; do
        if [ -f "$binary" ] && [[ "$binary" != *.zip ]]; then
            xcrun stapler staple "$binary"
            log_success "Stapled notarization for $binary"
        fi
    done
    
    # Clean up ZIP files
    for archive in "${archives[@]}"; do
        rm "$archive"
    done
    
    log_success "Notarization process completed"
}

# Function to verify final binaries
verify_binaries() {
    log_info "Verifying final signed and notarized binaries..."
    
    local verified_count=0
    
    for binary in "$BUILD_DIR/$BINARY_NAME-darwin-"*; do
        if [ -f "$binary" ]; then
            log_info "Verifying $(basename "$binary")..."
            
            # Check code signature
            codesign --verify --verbose=2 "$binary"
            
            # Check Gatekeeper assessment
            spctl --assess --type execute --verbose=2 "$binary"
            
            # Check notarization
            xcrun stapler validate "$binary"
            
            log_success "$(basename "$binary") is properly signed and notarized"
            ((verified_count++))
        fi
    done
    
    log_success "Verified $verified_count binaries"
}

# Main function
main() {
    echo "🔐 Slide MCP Server - macOS Code Signing and Notarization"
    echo "========================================================="
    
    # Check prerequisites
    check_prerequisites
    
    # Find signing identity
    DEVELOPER_ID=$(find_signing_identity)
    
    # Check for keychain profiles
    HAS_KEYCHAIN_PROFILE=false
    if check_keychain_profiles; then
        HAS_KEYCHAIN_PROFILE=true
    fi
    
    # Build binaries
    build_binaries
    
    # Sign binaries
    sign_binaries "$DEVELOPER_ID"
    
    # Notarize if we have keychain profile
    if [ "$HAS_KEYCHAIN_PROFILE" = true ]; then
        notarize_binaries "slide-mcp-profile"
        verify_binaries
    else
        log_warning "Skipping notarization (no keychain profile found)"
        log_info "Binaries are signed but not notarized"
        log_info "Users may still see Gatekeeper warnings"
    fi
    
    echo ""
    log_success "🎉 Process completed successfully!"
    
    if [ "$HAS_KEYCHAIN_PROFILE" = true ]; then
        log_info "Your macOS binaries are now signed and notarized"
        log_info "They should pass Gatekeeper checks without warnings"
    else
        log_info "Your macOS binaries are signed but not notarized"
        log_info "Set up notarization credentials to complete the process"
    fi
    
    echo ""
    log_info "Signed binaries are located in: $BUILD_DIR"
    ls -la "$BUILD_DIR"/$BINARY_NAME-darwin-*
}

# Run main function
main "$@" 