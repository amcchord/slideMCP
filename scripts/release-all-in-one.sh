#!/bin/bash

# Unified Release Script for Slide MCP Server
# This script handles the complete release process from commit to GitHub release

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_DIR/build"
BINARY_NAME="slide-mcp-server"

# Command line options
DRY_RUN=false
NEW_VERSION=""
SKIP_TESTS=false
NO_PUSH=false
SKIP_MACOS_SIGNING=false
FORCE_OVERWRITE=false

# Apple credentials for notarization
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
NC='\033[0m'

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --no-push)
            NO_PUSH=true
            shift
            ;;
        --skip-macos-signing)
            SKIP_MACOS_SIGNING=true
            shift
            ;;
        --force|-f)
            FORCE_OVERWRITE=true
            shift
            ;;
        --help|-h)
            echo "Slide MCP Server Unified Release Script"
            echo
            echo "Usage: $0 [options] [version]"
            echo
            echo "Options:"
            echo "  --dry-run              Simulate release without making changes"
            echo "  --skip-tests           Skip running tests"
            echo "  --no-push              Don't push to GitHub (local build only)"
            echo "  --skip-macos-signing   Skip macOS signing/notarization (TESTING ONLY)"
            echo "                         WARNING: Unsigned binaries won't work on user systems"
            echo "                         Must be used with --no-push"
            echo "  -f, --force            Force overwrite existing release/tag"
            echo "                         Deletes existing git tag and GitHub release"
            echo "  -h, --help             Show this help message"
            echo
            echo "Arguments:"
            echo "  version         Version to release (e.g., v2.4.1)"
            echo "                  If not provided, patch version will be auto-incremented"
            echo
            echo "Environment Variables:"
            echo "  APPLE_ID               Apple ID for notarization (required for macOS)"
            echo "  APP_SPECIFIC_PASSWORD  App-specific password (required for macOS)"
            echo
            echo "Examples:"
            echo "  # Full release (requires macOS + Apple credentials)"
            echo "  export APPLE_ID='your-apple-id@example.com'"
            echo "  export APP_SPECIFIC_PASSWORD='xxxx-xxxx-xxxx-xxxx'"
            echo "  $0                               # Auto-increment patch version"
            echo "  $0 v2.5.0                        # Release specific version"
            echo ""
            echo "  # Testing/local builds"
            echo "  $0 --dry-run v2.5.0              # Test without making changes"
            echo "  $0 --no-push --skip-macos-signing   # Local unsigned build"
            echo ""
            echo "  # Re-release existing version (overwrite)"
            echo "  $0 --force v2.5.0                # Force overwrite existing release"
            exit 0
            ;;
        *)
            if [ -z "$NEW_VERSION" ]; then
                NEW_VERSION=$1
            else
                echo "Error: Unknown argument: $1"
                exit 1
            fi
            shift
            ;;
    esac
done

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

log_dry_run() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}[DRY-RUN]${NC} $1"
    fi
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if [ "$NO_PUSH" = false ]; then
        if ! command -v gh &> /dev/null; then
            missing_tools+=("gh (GitHub CLI)")
        fi
    fi
    
    if ! command -v jq &> /dev/null; then
        missing_tools+=("jq")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools:"
        for tool in "${missing_tools[@]}"; do
            echo "  - $tool"
        done
        exit 1
    fi
    
    # Check GitHub CLI authentication if pushing
    if [ "$NO_PUSH" = false ] && ! gh auth status &> /dev/null; then
        log_error "GitHub CLI not authenticated. Run 'gh auth login' first"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to check git status
check_git_status() {
    log_info "Checking git status..."
    
    cd "$PROJECT_DIR"
    
    # Check if on main branch
    current_branch=$(git branch --show-current)
    if [ "$current_branch" != "main" ]; then
        log_error "Not on main branch (currently on: $current_branch)"
        log_info "Switch to main branch with: git checkout main"
        exit 1
    fi
    
    # Check for uncommitted changes
    if [ -n "$(git status --porcelain)" ]; then
        log_warning "There are uncommitted changes in the working directory"
        git status --short
        echo
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Aborting release"
            exit 1
        fi
    fi
    
    log_success "Git status check passed"
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
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would update Makefile to VERSION=$NEW_VERSION"
        log_dry_run "Would update config.go to Version = \"${NEW_VERSION:1}\"" # Remove 'v' prefix
        return
    fi
    
    # Update Makefile
    sed -i.bak "s/^VERSION=.*/VERSION=$NEW_VERSION/" "$PROJECT_DIR/Makefile"
    
    # Update config.go (remove 'v' prefix for version constant)
    local version_no_v="${NEW_VERSION:1}"
    sed -i.bak "s/Version    = \".*\"/Version    = \"$version_no_v\"/" "$PROJECT_DIR/config.go"
    
    # Clean up backup files
    find "$PROJECT_DIR" -name "*.bak" -delete
    
    log_success "Version updated to $NEW_VERSION"
}

# Function to run tests
run_tests() {
    if [ "$SKIP_TESTS" = true ]; then
        log_warning "Skipping tests as requested"
        return
    fi
    
    log_info "Running tests..."
    
    cd "$PROJECT_DIR"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would run: go test -v ./..."
        return
    fi
    
    # Run Go tests if they exist
    if go test -v ./... 2>&1 | grep -q "no test files"; then
        log_warning "No tests found"
    else
        log_success "Tests passed"
    fi
}

# Function to commit version changes
commit_version_changes() {
    log_info "Committing version changes..."
    
    cd "$PROJECT_DIR"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would commit changes with message: Release $NEW_VERSION"
        return
    fi
    
    # Add changed files
    git add Makefile config.go
    
    # Check if there are changes to commit
    if git diff --cached --quiet; then
        log_info "No version changes to commit"
    else
        git commit -m "Release $NEW_VERSION"
        log_success "Version changes committed"
    fi
}

# Function to build all platforms
build_all_platforms() {
    log_info "Building for all platforms..."
    
    cd "$PROJECT_DIR"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would run: make clean && make build-all"
        return
    fi
    
    make clean
    make build-all
    # v3.1.0+: also produce the macOS universal binary (lipo'd amd64 + arm64).
    # The .mcpb ships only the universal binary on Mac, so this is required
    # for pack-dxt-signed below.
    if [[ "$OSTYPE" == "darwin"* ]]; then
        make build-universal-darwin
    else
        log_warning "Skipping macOS universal binary (only producible on macOS)"
    fi
    
    log_success "All platforms built successfully"
}

# Function to sign and notarize macOS binaries
sign_and_notarize_macos() {
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would sign and notarize macOS binaries"
        return
    fi
    
    # Allow skipping ONLY for local testing (not for GitHub releases)
    if [ "$SKIP_MACOS_SIGNING" = true ]; then
        if [ "$NO_PUSH" = false ]; then
            log_error "Cannot skip macOS signing for GitHub releases"
            log_info "macOS binaries MUST be signed and notarized for user distribution"
            log_info "--skip-macos-signing can only be used with --no-push"
            exit 1
        fi
        log_warning "Skipping macOS signing/notarization (TESTING ONLY)"
        log_warning "These binaries will NOT work on user systems without manual approval"
        return
    fi
    
    # Check if on macOS
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_error "macOS code signing REQUIRED but not on macOS"
        log_info "macOS binaries must be signed and notarized to work on users' systems"
        log_info "Options:"
        log_info "  1. Run this script on macOS"
        log_info "  2. Use --no-push --skip-macos-signing for local testing only"
        exit 1
    fi
    
    # Check for Apple credentials - these are REQUIRED for proper releases
    if [ -z "$APPLE_ID" ] || [ -z "$APP_SPECIFIC_PASSWORD" ]; then
        log_error "Apple credentials REQUIRED for macOS notarization"
        log_info "macOS binaries must be notarized to work without security warnings"
        log_info ""
        log_info "Set these environment variables:"
        log_info "  export APPLE_ID='your-apple-id@example.com'"
        log_info "  export APP_SPECIFIC_PASSWORD='your-app-specific-password'"
        log_info ""
        log_info "To get an app-specific password:"
        log_info "  1. Go to appleid.apple.com"
        log_info "  2. Sign in and go to Security section"
        log_info "  3. Generate an app-specific password"
        log_info ""
        log_info "For testing only: use --no-push --skip-macos-signing"
        exit 1
    fi
    
    log_info "Signing and notarizing macOS binaries..."
    
    cd "$PROJECT_DIR"
    
    # Setup notarization keychain profile
    if ! xcrun notarytool history --keychain-profile "$KEYCHAIN_PROFILE" &> /dev/null; then
        log_info "Creating keychain profile..."
        local team_id
        team_id=$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1 | sed -n 's/.*(\([^)]*\)).*/\1/p')
        xcrun notarytool store-credentials "$KEYCHAIN_PROFILE" \
            --apple-id "$APPLE_ID" \
            --password "$APP_SPECIFIC_PASSWORD" \
            --team-id "$team_id"
    fi
    
    # Find signing identity
    local identities
    identities=$(security find-identity -v -p codesigning | grep "Developer ID Application" | head -1)
    local developer_id
    developer_id=$(echo "$identities" | sed -n 's/.*"\(Developer ID Application: [^"]*\)".*/\1/p')
    
    # Sign + notarize the macOS UNIVERSAL binary that ships inside the .mcpb.
    # The standalone per-arch binaries (kept for power users) are signed too
    # so the per-arch tarballs are also distributable.
    log_info "Signing + notarizing macOS universal binary..."
    if ! DEVELOPER_ID="$developer_id" KEYCHAIN_PROFILE="$KEYCHAIN_PROFILE" \
            make notarize-darwin-universal; then
        log_error "Failed to sign+notarize universal binary"
        exit 1
    fi

    log_info "Signing per-arch macOS binaries (kept as standalone downloads)..."
    DEVELOPER_ID="$developer_id" make sign-macos BINARY="$BUILD_DIR/$BINARY_NAME-darwin-amd64"
    DEVELOPER_ID="$developer_id" make sign-macos BINARY="$BUILD_DIR/$BINARY_NAME-darwin-arm64"

    log_info "Verifying signatures..."
    codesign --verify --verbose=2 "$BUILD_DIR/$BINARY_NAME-darwin-universal"
    codesign --verify --verbose=2 "$BUILD_DIR/$BINARY_NAME-darwin-amd64"
    codesign --verify --verbose=2 "$BUILD_DIR/$BINARY_NAME-darwin-arm64"

    log_success "macOS binaries signed and notarized successfully"
    log_info "Binaries are ready for distribution"
}

# Function to assemble the .mcpb (Claude Desktop Extension) bundle - the
# headline release artifact for v3.1.0+. Must run AFTER signing so the
# universal binary inside the bundle is notarized.
build_mcpb() {
    log_info "Building .mcpb bundle..."
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would run: make pack-dxt (uses already-signed binaries)"
        return
    fi

    cd "$PROJECT_DIR"
    # We use pack-dxt (not pack-dxt-signed) here because we already signed
    # and notarized notarize-darwin-universal in sign_and_notarize_macos();
    # pack-dxt-signed would re-run the whole notarize pipeline and double
    # our wall-clock time.
    if ! make stage-dxt && cd "$BUILD_DIR/dxt-stage" && zip -qr "../$BINARY_NAME.mcpb" .; then
        log_error "Failed to assemble .mcpb"
        exit 1
    fi
    cd "$PROJECT_DIR"
    log_success "Built $BUILD_DIR/$BINARY_NAME.mcpb (drag-drop install for Claude Desktop)"
}

# Function to create release packages
create_release_packages() {
    log_info "Creating release packages..."
    
    cd "$BUILD_DIR"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would create release packages"
        return
    fi
    
    # Per-platform tarballs (kept for `claude mcp add` / Cursor / CI users)
    for arch in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 darwin-universal; do
        if [ -f "$BINARY_NAME-$arch" ]; then
            tar -czf "$BINARY_NAME-$NEW_VERSION-${arch/darwin-/macos-}.tar.gz" "$BINARY_NAME-$arch"
            # Also create darwin-named packages for backward compatibility
            if [[ $arch == darwin-* ]]; then
                tar -czf "$BINARY_NAME-$NEW_VERSION-$arch.tar.gz" "$BINARY_NAME-$arch"
            fi
        fi
    done

    # Windows zip
    if [ -f "$BINARY_NAME-windows-amd64.exe" ]; then
        zip "$BINARY_NAME-$NEW_VERSION-windows-x64.zip" "$BINARY_NAME-windows-amd64.exe"
    fi

    # Versioned copy of the .mcpb so users can pin to a specific version, in
    # addition to the unversioned $BINARY_NAME.mcpb that lives at the
    # `releases/latest/download/slide-mcp-server.mcpb` URL.
    if [ -f "$BINARY_NAME.mcpb" ]; then
        cp "$BINARY_NAME.mcpb" "$BINARY_NAME-$NEW_VERSION.mcpb"
    fi

    cd "$PROJECT_DIR"
    log_success "Release packages created"
}

# Function to generate checksums
generate_checksums() {
    log_info "Generating checksums..."
    
    cd "$BUILD_DIR"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would generate SHA256 checksums"
        return
    fi
    
    rm -f checksums.sha256
    find . \( -name "*.tar.gz" -o -name "*.zip" -o -name "*.mcpb" \) | sort | while read -r file; do
        shasum -a 256 "$file" >> checksums.sha256
    done
    
    cd "$PROJECT_DIR"
    log_success "Checksums generated"
}

# Function to create release notes
create_release_notes() {
    log_info "Creating release notes..."
    
    local release_notes_file="$BUILD_DIR/RELEASE_NOTES.md"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would create release notes"
        return
    fi
    
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
    
    cat >> "$release_notes_file" << 'EOF'

## Installation

### Claude Desktop (drag and drop)

Download **slide-mcp-server.mcpb** and drop it onto Claude Desktop's
Extensions screen. Claude Desktop will prompt for your Slide API token
and start the server automatically. The `.mcpb` contains signed binaries
for macOS (universal), Linux (amd64 + arm64), and Windows (amd64).

### Claude Code

```bash
claude mcp add slide --env SLIDE_API_KEY=tk_... -- /path/to/slide-mcp-server
```

### Standalone binaries (Cursor, CI, custom hosts)

- **macOS** (universal Apple Silicon + Intel): slide-mcp-server-$NEW_VERSION-darwin-universal.tar.gz
- **Linux x64**: slide-mcp-server-$NEW_VERSION-linux-amd64.tar.gz
- **Linux ARM64**: slide-mcp-server-$NEW_VERSION-linux-arm64.tar.gz
- **Windows x64**: slide-mcp-server-$NEW_VERSION-windows-x64.zip

(Per-arch macOS tarballs `darwin-amd64.tar.gz` / `darwin-arm64.tar.gz` are
also published for backward compatibility.)

## Verification

```bash
shasum -a 256 -c checksums.sha256
```

## macOS Security

The macOS binaries (universal + per-arch) are signed and notarized by
Apple. They run without Gatekeeper warnings on macOS 10.15+.
EOF
    
    log_success "Release notes created"
}

# Function to create release directory
create_release_directory() {
    log_info "Creating release directory structure..."
    
    local release_dir="$PROJECT_DIR/release/$NEW_VERSION"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would create release directory: $release_dir"
        return
    fi
    
    mkdir -p "$release_dir"
    
    # Copy release artifacts
    cp "$BUILD_DIR"/*.tar.gz "$release_dir/" 2>/dev/null || true
    cp "$BUILD_DIR"/*.zip "$release_dir/" 2>/dev/null || true
    cp "$BUILD_DIR/checksums.sha256" "$release_dir/" 2>/dev/null || true
    cp "$BUILD_DIR/RELEASE_NOTES.md" "$release_dir/" 2>/dev/null || true
    
    log_success "Release directory created: $release_dir"
}

# Function to clean up existing release
cleanup_existing_release() {
    log_info "Checking for existing release..."
    
    cd "$PROJECT_DIR"
    
    # Check if tag exists locally
    if git rev-parse "$NEW_VERSION" >/dev/null 2>&1; then
        if [ "$FORCE_OVERWRITE" = true ]; then
            log_warning "Deleting existing local tag: $NEW_VERSION"
            if [ "$DRY_RUN" = false ]; then
                git tag -d "$NEW_VERSION"
            else
                log_dry_run "Would delete local tag: $NEW_VERSION"
            fi
        else
            log_error "Tag $NEW_VERSION already exists locally"
            log_info "Use --force to overwrite existing release"
            exit 1
        fi
    fi
    
    # Check if tag exists on remote
    if [ "$NO_PUSH" = false ] && git ls-remote --tags origin | grep -q "refs/tags/$NEW_VERSION"; then
        if [ "$FORCE_OVERWRITE" = true ]; then
            log_warning "Deleting existing remote tag: $NEW_VERSION"
            if [ "$DRY_RUN" = false ]; then
                git push --delete origin "$NEW_VERSION" 2>/dev/null || true
            else
                log_dry_run "Would delete remote tag: $NEW_VERSION"
            fi
        else
            log_error "Tag $NEW_VERSION already exists on remote"
            log_info "Use --force to overwrite existing release"
            exit 1
        fi
    fi
    
    # Check if GitHub release exists
    if [ "$NO_PUSH" = false ] && command -v gh &> /dev/null; then
        if gh release view "$NEW_VERSION" &> /dev/null; then
            if [ "$FORCE_OVERWRITE" = true ]; then
                log_warning "Deleting existing GitHub release: $NEW_VERSION"
                if [ "$DRY_RUN" = false ]; then
                    gh release delete "$NEW_VERSION" --yes 2>/dev/null || true
                else
                    log_dry_run "Would delete GitHub release: $NEW_VERSION"
                fi
            else
                log_error "GitHub release $NEW_VERSION already exists"
                log_info "Use --force to overwrite existing release"
                exit 1
            fi
        fi
    fi
    
    if [ "$FORCE_OVERWRITE" = true ]; then
        log_success "Cleaned up existing release artifacts"
    fi
}

# Function to create git tag and push
create_tag_and_push() {
    log_info "Creating git tag..."
    
    cd "$PROJECT_DIR"
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would create tag: $NEW_VERSION"
        log_dry_run "Would push to origin"
        return
    fi
    
    # Create tag
    git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
    
    if [ "$NO_PUSH" = true ]; then
        log_warning "Skipping git push (--no-push flag set)"
        return
    fi
    
    # Push changes and tag
    git push origin main
    git push origin "$NEW_VERSION"
    
    log_success "Changes committed and tagged"
}

# Function to create GitHub release
create_github_release() {
    if [ "$NO_PUSH" = true ]; then
        log_warning "Skipping GitHub release creation (--no-push flag set)"
        return
    fi
    
    # Final check: ensure macOS binaries are signed (universal is what ships
    # in the .mcpb; per-arch are still kept for the standalone tarballs).
    if [[ "$OSTYPE" == "darwin"* ]]; then
        log_info "Verifying macOS binaries are properly signed..."
        for bin in "$BINARY_NAME-darwin-universal" "$BINARY_NAME-darwin-amd64" "$BINARY_NAME-darwin-arm64"; do
            if ! codesign --verify "$BUILD_DIR/$bin" 2>/dev/null; then
                log_error "$bin is not properly signed!"
                log_info "macOS binaries must be signed for user distribution"
                exit 1
            fi
        done
        # The .mcpb itself must also be present.
        if [ ! -f "$BUILD_DIR/$BINARY_NAME.mcpb" ]; then
            log_error ".mcpb bundle is missing - did build_mcpb run?"
            exit 1
        fi
        log_success "macOS binaries verified as signed; .mcpb present"
    fi
    
    log_info "Creating GitHub release..."
    
    if [ "$DRY_RUN" = true ]; then
        log_dry_run "Would create GitHub release: $NEW_VERSION"
        return
    fi
    
    # Create the release
    gh release create "$NEW_VERSION" \
        --title "Slide MCP Server $NEW_VERSION" \
        --notes-file "$BUILD_DIR/RELEASE_NOTES.md" \
        --draft=false \
        --prerelease=false
    
    # Upload assets
    local assets=()
    # The unversioned .mcpb is the headline asset - the README links it via
    # https://github.com/$GITHUB_REPO/releases/latest/download/slide-mcp-server.mcpb
    # so the link never has to change per release.
    if [ -f "$BUILD_DIR/$BINARY_NAME.mcpb" ]; then
        assets+=("$BUILD_DIR/$BINARY_NAME.mcpb")
    fi
    while IFS= read -r file; do
        assets+=("$file")
    done < <(find "$BUILD_DIR" -name "*$NEW_VERSION*.tar.gz" -o -name "*$NEW_VERSION*.zip" -o -name "*$NEW_VERSION*.mcpb")

    if [ -f "$BUILD_DIR/checksums.sha256" ]; then
        assets+=("$BUILD_DIR/checksums.sha256")
    fi

    if [ ${#assets[@]} -gt 0 ]; then
        gh release upload "$NEW_VERSION" "${assets[@]}"
    fi
    
    log_success "GitHub release created: https://github.com/$GITHUB_REPO/releases/tag/$NEW_VERSION"
}

# Function to show summary
show_summary() {
    echo
    log_success "Release $NEW_VERSION completed successfully!"
    echo "=============================================="
    echo "Version: $NEW_VERSION"
    echo "Build Directory: $BUILD_DIR"
    echo "Release Directory: $PROJECT_DIR/release/$NEW_VERSION"
    if [ "$NO_PUSH" = false ]; then
        echo "GitHub Release: https://github.com/$GITHUB_REPO/releases/tag/$NEW_VERSION"
    fi
    echo
    if [ "$DRY_RUN" = false ]; then
        echo "Release artifacts:"
        ls -lh "$BUILD_DIR"/*.tar.gz "$BUILD_DIR"/*.zip "$BUILD_DIR/checksums.sha256" 2>/dev/null || echo "No artifacts found"
    fi
    echo
}

# Main execution function
main() {
    echo "=============================================="
    echo "Slide MCP Server Unified Release Script"
    echo "=============================================="
    echo
    
    if [ "$DRY_RUN" = true ]; then
        log_warning "DRY-RUN MODE - No changes will be made"
        echo
    fi
    
    # Run release steps
    check_prerequisites
    check_git_status
    determine_version
    cleanup_existing_release
    update_version_files
    run_tests
    commit_version_changes
    build_all_platforms
    sign_and_notarize_macos
    build_mcpb
    create_release_packages
    generate_checksums
    create_release_notes
    create_release_directory
    create_tag_and_push
    create_github_release
    show_summary
}

# Run main
main

