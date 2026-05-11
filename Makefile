# Slide MCP Server Makefile
#
# The headline deliverable is build/slide-mcp-server.mcpb, a single Claude
# Desktop Extension bundle that contains every supported platform binary
# (signed + notarized macOS universal binary, linux amd64+arm64, windows
# amd64) and is selected at runtime via the manifest's platform_overrides.

BINARY_NAME = slide-mcp-server
VERSION     = v4.0.0
BUILD_DIR   = build
DXT_STAGE   = $(BUILD_DIR)/dxt-stage
MCPB_FILE   = $(BUILD_DIR)/$(BINARY_NAME).mcpb

# Code signing variables (set these via environment or command line)
DEVELOPER_ID     ?= ""
KEYCHAIN_PROFILE ?= ""

# Build flags
LDFLAGS = -ldflags "-s -w"

# Per-platform binary paths
BIN_DARWIN_AMD64    = $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
BIN_DARWIN_ARM64    = $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
BIN_DARWIN_UNIVERSAL = $(BUILD_DIR)/$(BINARY_NAME)-darwin-universal
BIN_LINUX_AMD64     = $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
BIN_LINUX_ARM64     = $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
BIN_WINDOWS_AMD64   = $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe

# Default target
.PHONY: all
all: build

# -----------------------------------------------------------------------------
# Per-platform builds (file targets so they are only rebuilt when sources change)
# -----------------------------------------------------------------------------

GO_SOURCES := $(shell find . -name '*.go' -not -path './build/*' -not -path './Installer/*')

$(BIN_DARWIN_AMD64): $(GO_SOURCES)
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $@ .

$(BIN_DARWIN_ARM64): $(GO_SOURCES)
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $@ .

# macOS universal binary (lipo'd Intel + Apple Silicon). One binary that
# runs on every Mac the .mcpb might land on.
$(BIN_DARWIN_UNIVERSAL): $(BIN_DARWIN_AMD64) $(BIN_DARWIN_ARM64)
	@if ! command -v lipo >/dev/null 2>&1; then \
		echo "ERROR: lipo not found. Universal binary can only be built on macOS."; \
		exit 1; \
	fi
	lipo -create -output $@ $(BIN_DARWIN_AMD64) $(BIN_DARWIN_ARM64)
	@echo "Built universal binary: $@"
	@lipo -info $@

$(BIN_LINUX_AMD64): $(GO_SOURCES)
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $@ .

$(BIN_LINUX_ARM64): $(GO_SOURCES)
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $@ .

$(BIN_WINDOWS_AMD64): $(GO_SOURCES)
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $@ .

# Convenience phony aggregates
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

.PHONY: build-universal-darwin
build-universal-darwin: $(BIN_DARWIN_UNIVERSAL)

.PHONY: build-all
build-all: $(BIN_LINUX_AMD64) $(BIN_LINUX_ARM64) $(BIN_DARWIN_AMD64) $(BIN_DARWIN_ARM64) $(BIN_WINDOWS_AMD64)

.PHONY: build-all-universal
build-all-universal: $(BIN_LINUX_AMD64) $(BIN_LINUX_ARM64) $(BIN_DARWIN_UNIVERSAL) $(BIN_WINDOWS_AMD64)

# -----------------------------------------------------------------------------
# Code signing + notarization (macOS)
# -----------------------------------------------------------------------------

# Sign an arbitrary macOS binary (BINARY=...). Used by sign-darwin-universal.
.PHONY: sign-macos
sign-macos:
	@if [ -z "$(BINARY)" ]; then echo "Error: BINARY must be specified"; exit 1; fi
	@if [ -z "$(DEVELOPER_ID)" ] || [ "$(DEVELOPER_ID)" = "" ] || [ "$(DEVELOPER_ID)" = "\"\"" ]; then \
		echo "Error: DEVELOPER_ID must be set"; \
		echo ""; \
		echo "Quickest fix on this machine:"; \
		echo "    source scripts/release-env.sh"; \
		echo ""; \
		echo "If that file is missing or out of date, run:"; \
		echo "    ./scripts/setup-signing.sh"; \
		echo ""; \
		echo "Or pass it on the command line:"; \
		echo "    make ... DEVELOPER_ID='Developer ID Application: Your Name (TEAMID)'"; \
		exit 1; \
	fi
	@echo "Signing $(BINARY) with identity: $(DEVELOPER_ID)"
	codesign --sign "$(DEVELOPER_ID)" \
		--options runtime \
		--timestamp \
		--force \
		--verbose=2 \
		"$(BINARY)"
	codesign --verify --verbose=2 "$(BINARY)"

.PHONY: sign-darwin-universal
sign-darwin-universal: $(BIN_DARWIN_UNIVERSAL)
	$(MAKE) sign-macos BINARY=$(BIN_DARWIN_UNIVERSAL)

# Notarize the universal binary that ships inside the .mcpb. We submit a
# zip of just that binary, wait for Apple to approve, then staple the
# ticket to the binary so it works offline. Stapling can fail for raw
# binaries on some setups; that's not fatal because the notarization
# record is fetched online by Gatekeeper as a fallback.
.PHONY: notarize-darwin-universal
notarize-darwin-universal: sign-darwin-universal
	@if [ -z "$(KEYCHAIN_PROFILE)" ] || [ "$(KEYCHAIN_PROFILE)" = "" ] || [ "$(KEYCHAIN_PROFILE)" = "\"\"" ]; then \
		echo "Error: KEYCHAIN_PROFILE must be set for notarization"; \
		echo ""; \
		echo "Quickest fix: run 'source scripts/release-env.sh', or"; \
		echo "set up signing from scratch with: ./scripts/setup-signing.sh"; \
		echo ""; \
		echo "Create one with: xcrun notarytool store-credentials"; \
		exit 1; \
	fi
	@echo "Zipping universal binary for notarization..."
	cd $(BUILD_DIR) && zip -q $(BINARY_NAME)-darwin-universal.zip $(BINARY_NAME)-darwin-universal
	@echo "Submitting to Apple notary service (this may take a few minutes)..."
	xcrun notarytool submit $(BUILD_DIR)/$(BINARY_NAME)-darwin-universal.zip \
		--keychain-profile "$(KEYCHAIN_PROFILE)" \
		--wait
	@echo "Stapling notarization ticket to the binary..."
	-xcrun stapler staple $(BIN_DARWIN_UNIVERSAL)
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)-darwin-universal.zip

# -----------------------------------------------------------------------------
# .mcpb (Claude Desktop Extension) bundle
# -----------------------------------------------------------------------------

# Stage one .mcpb that contains every platform binary plus the manifest.
# Layout:
#   manifest.json
#   icon.png
#   server/slide-mcp-server-darwin-universal
#   server/slide-mcp-server-linux-amd64
#   server/slide-mcp-server-linux-arm64
#   server/slide-mcp-server-windows-amd64.exe
.PHONY: stage-dxt
stage-dxt: $(BIN_DARWIN_UNIVERSAL) $(BIN_LINUX_AMD64) $(BIN_LINUX_ARM64) $(BIN_WINDOWS_AMD64)
	@rm -rf $(DXT_STAGE)
	@mkdir -p $(DXT_STAGE)/server
	cp dxt/manifest.json $(DXT_STAGE)/manifest.json
	cp $(BIN_DARWIN_UNIVERSAL) $(DXT_STAGE)/server/$(BINARY_NAME)-darwin-universal
	cp $(BIN_LINUX_AMD64)     $(DXT_STAGE)/server/$(BINARY_NAME)-linux-amd64
	cp $(BIN_LINUX_ARM64)     $(DXT_STAGE)/server/$(BINARY_NAME)-linux-arm64
	cp $(BIN_WINDOWS_AMD64)   $(DXT_STAGE)/server/$(BINARY_NAME)-windows-amd64.exe
	@if [ -f dxt/icon.png ]; then cp dxt/icon.png $(DXT_STAGE)/icon.png; fi
	@if [ -d dxt/icons ]; then cp -R dxt/icons $(DXT_STAGE)/icons; fi

# Unsigned .mcpb (for dev iteration). Macs will need quarantine cleared
# manually for the inner binary; ship pack-dxt-signed for releases.
.PHONY: pack-dxt
pack-dxt: stage-dxt
	@rm -f $(MCPB_FILE)
	cd $(DXT_STAGE) && zip -qr ../$(BINARY_NAME).mcpb .
	@echo "Built $(MCPB_FILE)"
	@du -h $(MCPB_FILE) | awk '{print "Size: " $$1}'

# Signed + notarized .mcpb. Requires DEVELOPER_ID and KEYCHAIN_PROFILE.
# The macOS universal binary is signed and notarized BEFORE being copied
# into the staging directory, so the .mcpb that ships to users contains a
# binary Claude Desktop can spawn without Gatekeeper rejection.
.PHONY: pack-dxt-signed
pack-dxt-signed: notarize-darwin-universal $(BIN_LINUX_AMD64) $(BIN_LINUX_ARM64) $(BIN_WINDOWS_AMD64)
	@rm -rf $(DXT_STAGE)
	@mkdir -p $(DXT_STAGE)/server
	cp dxt/manifest.json $(DXT_STAGE)/manifest.json
	cp $(BIN_DARWIN_UNIVERSAL) $(DXT_STAGE)/server/$(BINARY_NAME)-darwin-universal
	cp $(BIN_LINUX_AMD64)     $(DXT_STAGE)/server/$(BINARY_NAME)-linux-amd64
	cp $(BIN_LINUX_ARM64)     $(DXT_STAGE)/server/$(BINARY_NAME)-linux-arm64
	cp $(BIN_WINDOWS_AMD64)   $(DXT_STAGE)/server/$(BINARY_NAME)-windows-amd64.exe
	@if [ -f dxt/icon.png ]; then cp dxt/icon.png $(DXT_STAGE)/icon.png; fi
	@if [ -d dxt/icons ]; then cp -R dxt/icons $(DXT_STAGE)/icons; fi
	@rm -f $(MCPB_FILE)
	cd $(DXT_STAGE) && zip -qr ../$(BINARY_NAME).mcpb .
	@echo "Built signed $(MCPB_FILE)"

# Lightweight verifier: the .mcpb unzips, the manifest parses, and the
# binary for THIS host actually runs. Catches "I broke pack-dxt" before
# we ship.
.PHONY: verify-dxt
verify-dxt: $(MCPB_FILE)
	@./scripts/verify-dxt.sh $(MCPB_FILE)

# -----------------------------------------------------------------------------
# Dev environment
# -----------------------------------------------------------------------------

.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: setup-dev
setup-dev:
	./scripts/setup-dev.sh

.PHONY: doctor
doctor:
	@./scripts/setup-dev.sh
	@echo "--- go vet ---"
	go vet ./...
	@echo "--- go build (current platform) ---"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "--- code signing readiness (informational; non-fatal) ---"
	@./scripts/setup-signing.sh --check || echo "(signing not set up; OK for dev iteration)"
	@echo "doctor: ok"

# Read-only check that this Mac is set up to sign + notarize the .mcpb.
# Run before invoking pack-dxt-signed for the first time on a new machine.
.PHONY: doctor-signing
doctor-signing:
	@./scripts/setup-signing.sh --check

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: run
run:
	go run .

# -----------------------------------------------------------------------------
# Local install helpers (for users not using the .mcpb path)
# -----------------------------------------------------------------------------

.PHONY: install
install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

.PHONY: install-claude-code
install-claude-code: build
	./scripts/install-claude-code.sh $(BUILD_DIR)/$(BINARY_NAME)

.PHONY: install-claude-desktop
install-claude-desktop: build
	./scripts/install-claude-desktop.sh $(BUILD_DIR)/$(BINARY_NAME)

# -----------------------------------------------------------------------------
# Release packaging
# -----------------------------------------------------------------------------

# Per-platform tarballs/zip for power users. The .mcpb is the headline
# release asset; these are kept for `claude mcp add` users and CI.
.PHONY: release
release: build-all $(BIN_DARWIN_UNIVERSAL) pack-dxt
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-universal.tar.gz $(BINARY_NAME)-darwin-universal && \
	zip     $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release artifacts in $(BUILD_DIR)/:"
	@ls -lh $(BUILD_DIR)/*.tar.gz $(BUILD_DIR)/*.zip $(MCPB_FILE) 2>/dev/null

.PHONY: release-signed
release-signed: build-all pack-dxt-signed
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-$(VERSION)-darwin-universal.tar.gz $(BINARY_NAME)-darwin-universal && \
	zip     $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Signed release artifacts in $(BUILD_DIR)/:"
	@ls -lh $(BUILD_DIR)/*.tar.gz $(BUILD_DIR)/*.zip $(MCPB_FILE) 2>/dev/null

# -----------------------------------------------------------------------------
# Help
# -----------------------------------------------------------------------------

.PHONY: help
help:
	@echo "Slide MCP Server build targets ($(VERSION))"
	@echo ""
	@echo "Drag-and-drop install bundle (the headline deliverable):"
	@echo "  pack-dxt              Build build/$(BINARY_NAME).mcpb (unsigned, dev)"
	@echo "  pack-dxt-signed       Build signed + notarized .mcpb (release)"
	@echo "  verify-dxt            Validate the .mcpb on this host"
	@echo ""
	@echo "Builds:"
	@echo "  build                 Build the server for the current platform"
	@echo "  build-all             Build per-arch binaries for every platform"
	@echo "  build-universal-darwin  Build the macOS lipo'd universal binary"
	@echo "  build-all-universal   Build everything used inside the .mcpb"
	@echo ""
	@echo "Code signing (macOS):"
	@echo "  sign-darwin-universal       Sign the universal binary"
	@echo "  notarize-darwin-universal   Sign + notarize + staple the universal binary"
	@echo "  doctor-signing              Read-only check: keychain identity + notarytool profile"
	@echo "  (first-time setup)          ./scripts/setup-signing.sh"
	@echo ""
	@echo "Local install helpers (non-.mcpb users):"
	@echo "  install                       Copy binary to /usr/local/bin"
	@echo "  install-claude-code           claude mcp add slide ..."
	@echo "  install-claude-desktop        Merge into claude_desktop_config.json"
	@echo ""
	@echo "Release:"
	@echo "  release                Build all artifacts + unsigned .mcpb"
	@echo "  release-signed         Build all artifacts + signed .mcpb"
	@echo ""
	@echo "Dev:"
	@echo "  setup-dev / doctor     Bootstrap a fresh dev box, sanity check"
	@echo "  deps / test / clean    Standard targets"
	@echo "  run                    Run the server locally (go run .)"
	@echo ""
	@echo "Code signing usage:"
	@echo "  make pack-dxt-signed \\"
	@echo "       DEVELOPER_ID='Developer ID Application: Your Name (TEAMID)' \\"
	@echo "       KEYCHAIN_PROFILE='YourKeychainProfile'"
