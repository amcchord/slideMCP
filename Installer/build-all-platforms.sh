#!/bin/bash

# Slide MCP Installer - Multi-Platform Release Builder
# This script builds the installer for all supported operating systems
# and signs/notarizes the macOS version if running on macOS

set -e  # Exit on any error

# Configuration
BINARY_NAME="slide-mcp-installer"
VERSION="v2.3.2"
BUILD_DIR="build"
RELEASE_DIR="releases"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

log_step() {
    echo -e "${PURPLE}STEP:${NC} $1"
}

# Get current platform
get_platform() {
    case "$(uname)" in
        Darwin) echo "darwin" ;;
        Linux) echo "linux" ;;
        MINGW*|CYGWIN*|MSYS*) echo "windows" ;;
        *) echo "unknown" ;;
    esac
}

# Check required tools
check_tools() {
    local missing_tools=()
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if ! command -v fyne &> /dev/null; then
        missing_tools+=("fyne")
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log_error "Missing required tools:"
        for tool in "${missing_tools[@]}"; do
            echo "  - $tool"
        done
        echo ""
        echo "Install missing tools:"
        echo "  - go: https://golang.org/dl/"
        echo "  - fyne: go install fyne.io/fyne/v2/cmd/fyne@latest"
        exit 1
    fi
}

# Check if macOS signing is available
check_macos_signing() {
    if [[ "$(get_platform)" != "darwin" ]]; then
        return 1
    fi
    
    if [[ -z "$DEVELOPER_ID_APPLICATION" || -z "$APPLE_ID" || -z "$APPLE_PASSWORD" || -z "$TEAM_ID" ]]; then
        return 1
    fi
    
    if ! command -v codesign &> /dev/null || ! command -v xcrun &> /dev/null; then
        return 1
    fi
    
    return 0
}

# Generate icon resource
generate_icon() {
    if [[ -f "MCP-Installer.png" ]]; then
        log_info "Generating icon resource..."
        fyne bundle -o icon.go MCP-Installer.png
    else
        log_warning "Icon file MCP-Installer.png not found"
    fi
}

# Clean previous builds
clean_builds() {
    log_info "Cleaning previous builds..."
    rm -rf "$BUILD_DIR" "$RELEASE_DIR"
    mkdir -p "$BUILD_DIR" "$RELEASE_DIR"
}

# Build for a specific platform
build_platform() {
    local target_os="$1"
    local target_arch="$2"
    local sign_macos="${3:-false}"
    
    log_step "Building for $target_os/$target_arch..."
    
    # Set environment variables for cross-compilation
    export GOOS="$target_os"
    export GOARCH="$target_arch"
    
    case "$target_os" in
        "darwin")
            if [[ "$sign_macos" == "true" && "$(get_platform)" == "darwin" ]]; then
                # Build signed version on macOS
                build_macos_signed "$target_arch"
            else
                # Build unsigned version
                build_macos_unsigned "$target_arch"
            fi
            ;;
        "windows")
            build_windows "$target_arch"
            ;;
        "linux")
            build_linux "$target_arch"
            ;;
        *)
            log_error "Unsupported target OS: $target_os"
            return 1
            ;;
    esac
}

# Build unsigned macOS version
build_macos_unsigned() {
    local arch="$1"
    local output_dir="$BUILD_DIR/darwin-$arch"
    
    mkdir -p "$output_dir"
    
    if [[ "$(get_platform)" == "darwin" ]]; then
        # Native build on macOS
        fyne package --target darwin --icon MCP-Installer.png --name "$BINARY_NAME" --source-dir .
        mv "${BINARY_NAME}.app" "$output_dir/"
    else
        log_warning "Cross-compiling macOS GUI apps from $(get_platform) is not supported"
        log_warning "Please build macOS version natively on macOS for best results"
        return 1
    fi
    
    # Create release package
    cd "$output_dir"
    tar -czf "../../$RELEASE_DIR/${BINARY_NAME}-${VERSION}-darwin-${arch}-unsigned.tar.gz" "${BINARY_NAME}.app"
    cd - > /dev/null
    
    log_success "macOS $arch build completed (unsigned)"
}

# Build signed macOS version
build_macos_signed() {
    local arch="$1"
    local output_dir="$BUILD_DIR/darwin-$arch-signed"
    
    mkdir -p "$output_dir"
    
    log_info "Building signed macOS app for $arch..."
    
    # Build the app
    fyne package --target darwin --icon MCP-Installer.png --name "$BINARY_NAME" --source-dir .
    mv "${BINARY_NAME}.app" "$output_dir/"
    
    # Sign the app
    log_info "Signing macOS app..."
    codesign --force --deep --sign "$DEVELOPER_ID_APPLICATION" --options runtime "$output_dir/${BINARY_NAME}.app"
    
    # Verify signature
    codesign --verify --verbose "$output_dir/${BINARY_NAME}.app"
    
    # Notarize if credentials are available
    if [[ -n "$APPLE_ID" && -n "$APPLE_PASSWORD" && -n "$TEAM_ID" ]]; then
        log_info "Notarizing macOS app..."
        
        # Create zip for notarization
        cd "$output_dir"
        zip -r "${BINARY_NAME}.zip" "${BINARY_NAME}.app"
        
        # Submit for notarization
        xcrun notarytool submit "${BINARY_NAME}.zip" \
            --apple-id "$APPLE_ID" \
            --password "$APPLE_PASSWORD" \
            --team-id "$TEAM_ID" \
            --wait
        
        # Staple the notarization
        xcrun stapler staple "${BINARY_NAME}.app"
        xcrun stapler validate "${BINARY_NAME}.app"
        
        # Clean up zip
        rm "${BINARY_NAME}.zip"
        cd - > /dev/null
        
        log_success "macOS app signed and notarized"
    else
        log_warning "Notarization skipped - missing credentials"
    fi
    
    # Create release package
    cd "$output_dir"
    tar -czf "../../$RELEASE_DIR/${BINARY_NAME}-${VERSION}-darwin-${arch}-signed.tar.gz" "${BINARY_NAME}.app"
    cd - > /dev/null
    
    log_success "macOS $arch build completed (signed)"
}

# Build Windows version
build_windows() {
    local arch="$1"
    local output_dir="$BUILD_DIR/windows-$arch"
    
    mkdir -p "$output_dir"
    
    log_info "Building Windows app for $arch..."
    
    if [[ "$(get_platform)" == "windows" ]] || command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        # Native Windows build or cross-compilation with MinGW
        export CGO_ENABLED=1
        if [[ "$arch" == "amd64" && "$(get_platform)" != "windows" ]]; then
            export CC=x86_64-w64-mingw32-gcc
        elif [[ "$arch" == "386" && "$(get_platform)" != "windows" ]]; then
            export CC=i686-w64-mingw32-gcc
        fi
        
        fyne package --target windows --icon MCP-Installer.png --name "$BINARY_NAME" --source-dir .
        mv "${BINARY_NAME}.exe" "$output_dir/"
        
        # Create release package
        cd "$output_dir"
        zip -r "../../$RELEASE_DIR/${BINARY_NAME}-${VERSION}-windows-${arch}.zip" "${BINARY_NAME}.exe"
        cd - > /dev/null
        
        log_success "Windows $arch build completed"
    else
        log_warning "Cross-compiling Windows GUI apps requires MinGW-w64"
        log_warning "Install mingw-w64 or build natively on Windows"
        return 1
    fi
}

# Build Linux version
build_linux() {
    local arch="$1"
    local output_dir="$BUILD_DIR/linux-$arch"
    
    mkdir -p "$output_dir"
    
    log_info "Building Linux app for $arch..."
    
    export CGO_ENABLED=1
    
    if [[ "$(get_platform)" == "linux" ]] || command -v gcc &> /dev/null; then
        fyne package --target linux --icon MCP-Installer.png --name "$BINARY_NAME" --source-dir .
        mv "$BINARY_NAME" "$output_dir/"
        
        # Create release package
        cd "$output_dir"
        tar -czf "../../$RELEASE_DIR/${BINARY_NAME}-${VERSION}-linux-${arch}.tar.gz" "$BINARY_NAME"
        cd - > /dev/null
        
        log_success "Linux $arch build completed"
    else
        log_warning "Cross-compiling Linux GUI apps requires appropriate C compiler"
        return 1
    fi
}

# Show build summary
show_summary() {
    echo ""
    log_success "Multi-platform build completed!"
    echo ""
    echo "Release packages created in $RELEASE_DIR/:"
    if [[ -d "$RELEASE_DIR" ]]; then
        ls -la "$RELEASE_DIR/"
    fi
    echo ""
    
    # Show platform-specific notes
    local current_platform=$(get_platform)
    echo "Platform-specific notes:"
    if [[ "$current_platform" != "darwin" ]]; then
        echo "  • macOS builds require native compilation on macOS"
    fi
    if [[ "$current_platform" != "windows" ]] && ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        echo "  • Windows builds require MinGW-w64 for cross-compilation"
    fi
    if [[ "$current_platform" != "linux" ]] && ! command -v gcc &> /dev/null; then
        echo "  • Linux builds require a C compiler for cross-compilation"
    fi
}

# Main execution
main() {
    local sign_macos=false
    local build_current_only=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --sign-macos)
                sign_macos=true
                shift
                ;;
            --current-only)
                build_current_only=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --sign-macos     Sign and notarize macOS builds (requires signing setup)"
                echo "  --current-only   Only build for the current platform"
                echo "  --help, -h       Show this help message"
                echo ""
                echo "Environment variables for macOS signing:"
                echo "  DEVELOPER_ID_APPLICATION - Code signing certificate name"
                echo "  APPLE_ID                 - Apple ID email"
                echo "  APPLE_PASSWORD           - App-specific password"
                echo "  TEAM_ID                  - Apple Developer Team ID"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    log_info "Starting multi-platform build process..."
    
    # Check prerequisites
    check_tools
    
    # Check macOS signing availability
    if [[ "$sign_macos" == "true" ]]; then
        if check_macos_signing; then
            log_info "macOS signing is available"
        else
            log_warning "macOS signing requested but not available"
            log_warning "Missing signing credentials or not running on macOS"
            sign_macos=false
        fi
    fi
    
    # Prepare build environment
    generate_icon
    clean_builds
    
    # Determine platforms to build
    local current_platform=$(get_platform)
    local platforms_to_build=()
    
    if [[ "$build_current_only" == "true" ]]; then
        case "$current_platform" in
            "darwin") platforms_to_build=("darwin:arm64" "darwin:amd64") ;;
            "linux") platforms_to_build=("linux:amd64" "linux:arm64") ;;
            "windows") platforms_to_build=("windows:amd64" "windows:386") ;;
        esac
    else
        # Build for all platforms (where possible)
        platforms_to_build=(
            "darwin:arm64"
            "darwin:amd64" 
            "windows:amd64"
            "windows:386"
            "linux:amd64"
            "linux:arm64"
        )
    fi
    
    # Build for each platform
    local successful_builds=0
    local total_builds=${#platforms_to_build[@]}
    
    for platform_arch in "${platforms_to_build[@]}"; do
        IFS=':' read -r platform arch <<< "$platform_arch"
        
        if build_platform "$platform" "$arch" "$sign_macos"; then
            successful_builds=$((successful_builds + 1))
        else
            log_warning "Build failed for $platform/$arch"
        fi
    done
    
    # Show summary
    show_summary
    
    echo ""
    log_info "Build summary: $successful_builds/$total_builds platforms built successfully"
    
    if [[ $successful_builds -lt $total_builds ]]; then
        log_warning "Some builds failed. Check the output above for details."
        log_warning "Consider building natively on target platforms for GUI applications."
    fi
}

# Run main function
main "$@" 