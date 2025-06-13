# Slide MCP Server Makefile

BINARY_NAME=slide-mcp-server
VERSION=v1.14
BUILD_DIR=build

# Code signing variables (set these via environment or command line)
DEVELOPER_ID ?= ""
KEYCHAIN_PROFILE ?= ""

# Build flags for signed binaries
LDFLAGS=-ldflags "-s -w"

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for all platforms
.PHONY: build-all
build-all: clean
	mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Build for all platforms with macOS code signing
.PHONY: build-all-signed
build-all-signed: clean
	@if [ -z "$(DEVELOPER_ID)" ]; then \
		echo "Error: DEVELOPER_ID must be set for code signing"; \
		echo "Usage: make build-all-signed DEVELOPER_ID='Developer ID Application: Your Name (TEAMID)'"; \
		exit 1; \
	fi
	mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	$(MAKE) sign-macos BINARY=$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	$(MAKE) sign-macos BINARY=$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Sign macOS binary
.PHONY: sign-macos
sign-macos:
	@if [ -z "$(BINARY)" ]; then \
		echo "Error: BINARY must be specified"; \
		exit 1; \
	fi
	@if [ -z "$(DEVELOPER_ID)" ]; then \
		echo "Error: DEVELOPER_ID must be set"; \
		exit 1; \
	fi
	@echo "Signing $(BINARY) with identity: $(DEVELOPER_ID)"
	codesign --sign "$(DEVELOPER_ID)" \
		--options runtime \
		--timestamp \
		--verbose=2 \
		"$(BINARY)"
	@echo "Verifying signature for $(BINARY)"
	codesign --verify --verbose=2 "$(BINARY)"
	spctl --assess --type execute --verbose=2 "$(BINARY)"

# Notarize macOS binaries (requires Apple Developer account)
.PHONY: notarize-macos
notarize-macos:
	@if [ -z "$(KEYCHAIN_PROFILE)" ]; then \
		echo "Error: KEYCHAIN_PROFILE must be set for notarization"; \
		echo "Create a keychain profile with: xcrun notarytool store-credentials"; \
		exit 1; \
	fi
	@echo "Creating ZIP archives for notarization..."
	cd $(BUILD_DIR) && \
	zip $(BINARY_NAME)-darwin-amd64.zip $(BINARY_NAME)-darwin-amd64 && \
	zip $(BINARY_NAME)-darwin-arm64.zip $(BINARY_NAME)-darwin-arm64
	@echo "Submitting for notarization..."
	xcrun notarytool submit $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64.zip \
		--keychain-profile "$(KEYCHAIN_PROFILE)" \
		--wait
	xcrun notarytool submit $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64.zip \
		--keychain-profile "$(KEYCHAIN_PROFILE)" \
		--wait
	@echo "Stapling notarization..."
	xcrun stapler staple $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	xcrun stapler staple $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64

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

# Run the server locally
.PHONY: run
run:
	go run .

# Install to system PATH
.PHONY: install
install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Create release packages
.PHONY: release
release: build-all
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

# Create signed release packages (macOS binaries will be signed and notarized)
.PHONY: release-signed
release-signed: build-all-signed notarize-macos
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

# Show help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build           - Build for current platform"
	@echo "  build-all       - Build for all supported platforms"
	@echo "  build-all-signed - Build for all platforms with macOS code signing"
	@echo "  sign-macos      - Sign a specific macOS binary"
	@echo "  notarize-macos  - Notarize macOS binaries with Apple"
	@echo "  deps            - Install/update dependencies"
	@echo "  test            - Run tests"
	@echo "  clean           - Remove build artifacts"
	@echo "  run             - Run the server locally"
	@echo "  install         - Install to system PATH"
	@echo "  release         - Create release packages"
	@echo "  release-signed  - Create signed release packages"
	@echo "  help            - Show this help"
	@echo ""
	@echo "Code Signing Usage:"
	@echo "  make build-all-signed DEVELOPER_ID='Developer ID Application: Your Name (TEAMID)'"
	@echo "  make notarize-macos KEYCHAIN_PROFILE='YourKeychainProfile'" 