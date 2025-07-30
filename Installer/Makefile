# Slide MCP Installer Makefile
# 
# This builds native platform packages using Fyne:
# - macOS: Creates .app bundle with proper icon and no terminal window
# - Windows: Creates .exe with embedded icon
# - Linux: Creates standard binary

BINARY_NAME=slide-mcp-installer
VERSION=v2.3.0
BUILD_DIR=build

# Build flags
LDFLAGS=-ldflags "-s -w"

# macOS Code Signing and Notarization (set these environment variables)
# DEVELOPER_ID_APPLICATION - Your "Developer ID Application" certificate name
# APPLE_ID - Your Apple ID email
# APPLE_PASSWORD - App-specific password or @keychain:AC_PASSWORD 
# TEAM_ID - Your Apple Developer Team ID

# Default target
.PHONY: all
all: generate-icon build

# Build for current platform (unsigned)
.PHONY: build
build: generate-icon
	mkdir -p $(BUILD_DIR)
	fyne package --target $$(go env GOOS) --icon MCP-Installer.png --name $(BINARY_NAME) --source-dir .
	@if [ "$$(go env GOOS)" = "darwin" ]; then \
		mv $(BINARY_NAME).app $(BUILD_DIR)/ 2>/dev/null || true; \
	elif [ "$$(go env GOOS)" = "windows" ]; then \
		mv $(BINARY_NAME).exe $(BUILD_DIR)/ 2>/dev/null || true; \
	else \
		mv $(BINARY_NAME) $(BUILD_DIR)/ 2>/dev/null || true; \
	fi

# Build signed and notarized app for macOS
.PHONY: build-signed
build-signed: generate-icon
	@if [ "$$(go env GOOS)" != "darwin" ]; then \
		echo "Signed builds are only supported on macOS"; \
		exit 1; \
	fi
	@if [ -z "$$DEVELOPER_ID_APPLICATION" ]; then \
		echo "Error: DEVELOPER_ID_APPLICATION environment variable not set"; \
		echo "Set it to your 'Developer ID Application' certificate name"; \
		exit 1; \
	fi
	mkdir -p $(BUILD_DIR)
	fyne package --target darwin --icon MCP-Installer.png --name $(BINARY_NAME) --source-dir .
	mv $(BINARY_NAME).app $(BUILD_DIR)/
	@echo "Signing macOS app bundle..."
	codesign --force --deep --sign "$$DEVELOPER_ID_APPLICATION" --options runtime $(BUILD_DIR)/$(BINARY_NAME).app
	@echo "Verifying code signature..."
	codesign --verify --verbose $(BUILD_DIR)/$(BINARY_NAME).app

# Notarize the signed macOS app
.PHONY: notarize
notarize: check-notarization-env
	@if [ ! -d "$(BUILD_DIR)/$(BINARY_NAME).app" ]; then \
		echo "Error: App bundle not found. Run 'make build-signed' first."; \
		exit 1; \
	fi
	@echo "Creating zip archive for notarization..."
	cd $(BUILD_DIR) && zip -r $(BINARY_NAME).zip $(BINARY_NAME).app
	@echo "Submitting app for notarization..."
	xcrun notarytool submit $(BUILD_DIR)/$(BINARY_NAME).zip \
		--apple-id "$$APPLE_ID" \
		--password "$$APPLE_PASSWORD" \
		--team-id "$$TEAM_ID" \
		--wait
	@echo "Stapling notarization ticket to app..."
	xcrun stapler staple $(BUILD_DIR)/$(BINARY_NAME).app
	@echo "Verifying notarization..."
	xcrun stapler validate $(BUILD_DIR)/$(BINARY_NAME).app
	rm $(BUILD_DIR)/$(BINARY_NAME).zip

# Check if notarization environment variables are set
.PHONY: check-notarization-env
check-notarization-env:
	@if [ -z "$$APPLE_ID" ]; then \
		echo "Error: APPLE_ID environment variable not set"; \
		exit 1; \
	fi
	@if [ -z "$$APPLE_PASSWORD" ]; then \
		echo "Error: APPLE_PASSWORD environment variable not set"; \
		echo "This should be an app-specific password or @keychain:AC_PASSWORD"; \
		exit 1; \
	fi
	@if [ -z "$$TEAM_ID" ]; then \
		echo "Error: TEAM_ID environment variable not set"; \
		exit 1; \
	fi

# Build, sign, and notarize in one step
.PHONY: build-release
build-release: build-signed notarize
	@echo "macOS app successfully built, signed, and notarized!"

# Build for all platforms (requires CGO setup for each target)
.PHONY: build-all
build-all: clean
	@echo "Note: Cross-compilation of GUI apps with Fyne requires CGO setup for each target platform."
	@echo "Building for current platform only. Use native builds on each target OS for best results."
	mkdir -p $(BUILD_DIR)
	fyne package --target $$(go env GOOS) --icon MCP-Installer.png --name $(BINARY_NAME) --source-dir .
	@if [ "$$(go env GOOS)" = "darwin" ]; then \
		mv $(BINARY_NAME).app $(BUILD_DIR)/ 2>/dev/null || true; \
	elif [ "$$(go env GOOS)" = "windows" ]; then \
		mv $(BINARY_NAME).exe $(BUILD_DIR)/ 2>/dev/null || true; \
	else \
		mv $(BINARY_NAME) $(BUILD_DIR)/ 2>/dev/null || true; \
	fi

# Install dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Run tests
.PHONY: test
test:
	go test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# Generate icon resource (run this if icon changes)
.PHONY: generate-icon
generate-icon:
	fyne bundle -o icon.go MCP-Installer.png

# Run the installer locally
.PHONY: run
run:
	go run .

# Build and open the app (macOS only)
.PHONY: open
open: build
	@if [ "$$(go env GOOS)" = "darwin" ]; then \
		open $(BUILD_DIR)/$(BINARY_NAME).app; \
	else \
		echo "open target is only supported on macOS"; \
	fi

# Create release packages
.PHONY: release
release: build
	@echo "Creating release package for current platform..."
	cd $(BUILD_DIR) && \
	if [ "$$(go env GOOS)" = "darwin" ]; then \
		tar -czf $(BINARY_NAME)-$(VERSION)-$$(go env GOOS)-$$(go env GOARCH).tar.gz $(BINARY_NAME).app; \
	elif [ "$$(go env GOOS)" = "windows" ]; then \
		zip -r $(BINARY_NAME)-$(VERSION)-$$(go env GOOS)-$$(go env GOARCH).zip $(BINARY_NAME).exe; \
	else \
		tar -czf $(BINARY_NAME)-$(VERSION)-$$(go env GOOS)-$$(go env GOARCH).tar.gz $(BINARY_NAME); \
	fi

# Create signed release package for macOS
.PHONY: release-signed
release-signed: build-release
	@echo "Creating signed release package for macOS..."
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-$$(go env GOARCH)-signed.tar.gz $(BINARY_NAME).app

# Build for all platforms
.PHONY: build-all-platforms
build-all-platforms:
	./build-all-platforms.sh

# Build for all platforms with macOS signing
.PHONY: build-all-platforms-signed
build-all-platforms-signed:
	./build-all-platforms.sh --sign-macos

# Build for current platform only (all architectures)
.PHONY: build-current-platform
build-current-platform:
	./build-all-platforms.sh --current-only

# Show help
.PHONY: help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development builds:"
	@echo "  build          - Build native app package for current platform (unsigned)"
	@echo "  build-all      - Build native app package for current platform (GUI apps don't cross-compile easily)"
	@echo "  run            - Run the installer locally"
	@echo "  open           - Build and open the app (macOS only)"
	@echo ""
	@echo "macOS Release builds (requires Apple Developer account):"
	@echo "  build-signed   - Build and code sign the macOS app"
	@echo "  notarize       - Notarize the signed macOS app (must run build-signed first)"
	@echo "  build-release  - Build, sign, and notarize in one step"
	@echo "  release-signed - Create signed release package for macOS"
	@echo ""
	@echo "Multi-platform builds:"
	@echo "  build-all-platforms        - Build for all supported platforms"
	@echo "  build-all-platforms-signed - Build for all platforms, sign macOS versions"
	@echo "  build-current-platform     - Build for current platform only (all architectures)"
	@echo ""
	@echo "General:"
	@echo "  generate-icon  - Generate icon resource from MCP-Installer.png"
	@echo "  deps           - Install/update dependencies"
	@echo "  test           - Run tests"
	@echo "  clean          - Remove build artifacts"
	@echo "  release        - Create release package for current platform (unsigned)"
	@echo "  help           - Show this help"
	@echo ""
	@echo "macOS Code Signing Environment Variables:"
	@echo "  DEVELOPER_ID_APPLICATION - Your 'Developer ID Application' certificate name"
	@echo "  APPLE_ID                 - Your Apple ID email"
	@echo "  APPLE_PASSWORD           - App-specific password or @keychain:AC_PASSWORD"
	@echo "  TEAM_ID                  - Your Apple Developer Team ID"
	@echo ""
	@echo "Note: For multi-platform releases, build natively on each target OS" 