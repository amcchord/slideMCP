package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/google/go-github/v57/github"
)

const (
	repoOwner  = "amcchord"
	repoName   = "slideMCP"
	binaryName = "slide-mcp-server"
	appTitle   = "Slide MCP Installer"
)

// ClaudeConfig represents the Claude Desktop configuration file
type ClaudeConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents an MCP server configuration
// Version field is used to track the installed version since the binary
// doesn't currently support --version flag
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env"`
	Version string            `json:"version,omitempty"` // Track installed version
}

type Installer struct {
	app              fyne.App
	window           fyne.Window
	statusLabel      *widget.Label
	progressBar      *widget.ProgressBar
	apiKeyEntry      *widget.Entry
	apiKeyStatus     *widget.Label
	installBtn       *widget.Button
	uninstallBtn     *widget.Button
	updateBtn        *widget.Button
	installPathEntry *widget.Entry
	browseBtn        *widget.Button
	versionLabel     *widget.Label
	latestVersion    string
	installedVersion string
}

func main() {
	a := app.New()
	a.SetIcon(resourceMCPInstallerPng)

	w := a.NewWindow(appTitle)
	w.SetIcon(resourceMCPInstallerPng)
	w.Resize(fyne.NewSize(500, 400))
	w.CenterOnScreen()

	installer := &Installer{
		app:    a,
		window: w,
	}

	installer.setupUI()
	w.ShowAndRun()
}

func (i *Installer) setupUI() {
	i.statusLabel = widget.NewLabel("Ready to install Slide MCP Server")
	i.progressBar = widget.NewProgressBar()
	i.progressBar.Hide()

	// Check Claude Desktop installation status
	claudeStatus := i.checkClaudeDesktop()
	claudeStatusLabel := widget.NewLabel(claudeStatus)

	// API Key input
	i.apiKeyEntry = widget.NewPasswordEntry()
	i.apiKeyEntry.SetPlaceHolder("Enter your Slide API key")
	i.apiKeyStatus = widget.NewLabel("")

	// Buttons
	i.installBtn = widget.NewButton("Install Slide MCP Server", i.install)
	i.uninstallBtn = widget.NewButton("Uninstall", i.uninstall)
	i.updateBtn = widget.NewButton("Update", i.update)

	// Install path input
	i.installPathEntry = widget.NewEntry()
	i.installPathEntry.SetPlaceHolder("Enter the install path")
	i.browseBtn = widget.NewButton("Browse", i.browse)

	// Version information
	i.versionLabel = widget.NewLabel("Installed Version: Unknown")
	i.latestVersion = "Unknown"
	i.installedVersion = "Unknown"

	// Initialize UI state
	i.refreshUI()

	// Get version information
	go i.getVersionInfo()

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Slide MCP Server Installer"),
		widget.NewSeparator(),
		claudeStatusLabel,
		widget.NewSeparator(),
		widget.NewLabel("Slide API Key:"),
		i.apiKeyEntry,
		i.apiKeyStatus,
		widget.NewSeparator(),
		widget.NewLabel("Install Path:"),
		container.NewBorder(nil, nil, nil, i.browseBtn, i.installPathEntry),
		widget.NewSeparator(),
		i.versionLabel,
		widget.NewSeparator(),
		i.statusLabel,
		i.progressBar,
		widget.NewSeparator(),
		container.NewHBox(
			layout.NewSpacer(),
			i.installBtn,
			i.uninstallBtn,
			i.updateBtn,
			layout.NewSpacer(),
		),
	)

	i.window.SetContent(container.NewPadded(content))
}

func (i *Installer) checkClaudeDesktop() string {
	configPath := i.getClaudeConfigPath()
	if configPath == "" {
		return "❌ Claude Desktop not found"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "⚠️  Claude Desktop found, but no configuration file yet"
	}

	return "✅ Claude Desktop found and configured"
}

func (i *Installer) getClaudeConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		return filepath.Join(homeDir, "AppData", "Roaming", "Claude", "claude_desktop_config.json")
	case "linux":
		return filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")
	default:
		return ""
	}
}

func (i *Installer) getInstallPath() string {
	// Check if custom path is set
	if i.installPathEntry != nil {
		customPath := strings.TrimSpace(i.installPathEntry.Text)
		if customPath != "" {
			return customPath
		}
	}

	// Default path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		return filepath.Join(homeDir, ".local", "bin", binaryName)
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "slide-mcp", binaryName+".exe")
	default:
		return ""
	}
}

func (i *Installer) isInstalled() bool {
	installPath := i.getInstallPath()
	if installPath == "" {
		return false
	}

	_, err := os.Stat(installPath)
	return err == nil
}

func (i *Installer) isConfigured() bool {
	configPath := i.getClaudeConfigPath()
	if configPath == "" {
		return false
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}

	_, exists := config.MCPServers["slide"]
	return exists
}

func (i *Installer) getCurrentAPIKey() string {
	configPath := i.getClaudeConfigPath()
	if configPath == "" {
		return ""
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}

	slideServer, exists := config.MCPServers["slide"]
	if !exists {
		return ""
	}

	return slideServer.Env["SLIDE_API_KEY"]
}

func (i *Installer) refreshUI() {
	i.installBtn.Enable()
	isInstalled := i.isInstalled()
	isConfigured := i.isConfigured()
	currentAPIKey := i.getCurrentAPIKey()

	// Handle version comparison
	hasUpdate := false
	if i.installedVersion != "Not installed" && i.installedVersion != "Unknown" &&
		i.latestVersion != "Unknown" && i.compareVersions(i.installedVersion, i.latestVersion) < 0 {
		hasUpdate = true
	}

	if isConfigured {
		if currentAPIKey != "" {
			i.apiKeyEntry.SetText(currentAPIKey)
			i.apiKeyStatus.SetText("(Currently configured API key)")
			if hasUpdate {
				i.statusLabel.SetText("Slide MCP Server is configured - Update available!")
			} else {
				i.statusLabel.SetText("Slide MCP Server is configured and ready")
			}
		} else {
			i.apiKeyStatus.SetText("(No API key configured)")
			i.statusLabel.SetText("Slide MCP Server is configured but missing API key")
		}
		i.installBtn.SetText("Update Configuration")
		i.uninstallBtn.Enable()
		i.updateBtn.SetText(fmt.Sprintf("Update to %s", i.latestVersion))
		if hasUpdate {
			i.updateBtn.Enable()
		} else {
			i.updateBtn.Disable()
		}
	} else if isInstalled {
		i.apiKeyStatus.SetText("(Binary installed but not configured)")
		i.statusLabel.SetText("Slide MCP Server binary found but not configured")
		i.installBtn.SetText("Configure Slide MCP Server")
		i.uninstallBtn.Enable()
		if hasUpdate {
			i.updateBtn.Enable()
			i.updateBtn.SetText(fmt.Sprintf("Update to %s", i.latestVersion))
		} else {
			i.updateBtn.Disable()
		}
	} else {
		i.apiKeyStatus.SetText("")
		i.statusLabel.SetText("Ready to install Slide MCP Server")
		i.installBtn.SetText("Install Slide MCP Server")
		i.uninstallBtn.Disable()
		i.updateBtn.Disable()
		i.apiKeyEntry.SetText("") // Clear API key field when not configured
	}
}

func (i *Installer) install() {
	apiKey := strings.TrimSpace(i.apiKeyEntry.Text)
	if apiKey == "" {
		dialog.ShowError(fmt.Errorf("Please enter your Slide API key"), i.window)
		return
	}

	// Disable UI during installation
	i.installBtn.Disable()
	i.uninstallBtn.Disable()
	i.progressBar.Show()
	i.statusLabel.SetText("Starting installation...")

	go func() {
		defer func() {
			i.refreshUI()
			i.progressBar.Hide()
		}()

		isUpdate := i.isConfigured()
		action := "Installation"
		if isUpdate {
			action = "Update"
		}

		if err := i.performInstall(apiKey); err != nil {
			dialog.ShowError(err, i.window)
			i.statusLabel.SetText(action + " failed")
		} else {
			successMsg := fmt.Sprintf("Slide MCP Server %s completed successfully!\n\nRestart Claude Desktop to use the updated server.", strings.ToLower(action))
			dialog.ShowInformation("Success", successMsg, i.window)
			i.statusLabel.SetText(action + " completed successfully")
		}
	}()
}

func (i *Installer) performInstall(apiKey string) error {
	// Step 1: Download binary
	i.statusLabel.SetText("Downloading Slide MCP Server...")
	i.progressBar.SetValue(0.1)

	binaryData, err := i.downloadBinary()
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Step 2: Install binary
	i.statusLabel.SetText("Installing binary...")
	i.progressBar.SetValue(0.5)

	// Use custom install path if provided
	installPath := strings.TrimSpace(i.installPathEntry.Text)
	if installPath == "" {
		installPath = i.getInstallPath()
	}

	if err := i.installBinary(binaryData, installPath); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Step 3: Update Claude config
	i.statusLabel.SetText("Updating Claude Desktop configuration...")
	i.progressBar.SetValue(0.8)

	if err := i.updateClaudeConfig(apiKey, installPath); err != nil {
		return fmt.Errorf("failed to update Claude config: %w", err)
	}

	i.progressBar.SetValue(1.0)
	return nil
}

func (i *Installer) downloadBinary() ([]byte, error) {
	client := github.NewClient(nil)
	ctx := context.Background()

	// Get the latest release
	release, _, err := client.Repositories.GetLatestRelease(ctx, repoOwner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	// Update our stored latest version
	if release.TagName != nil {
		i.latestVersion = *release.TagName
	}

	// Find the appropriate asset for current platform
	assetName := i.getAssetName(i.latestVersion)
	var assetURL string

	for _, asset := range release.Assets {
		if *asset.Name == assetName {
			assetURL = *asset.BrowserDownloadURL
			break
		}
	}

	if assetURL == "" {
		return nil, fmt.Errorf("no asset found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the asset
	resp, err := http.Get(assetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download asset: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset data: %w", err)
	}

	// Extract binary from archive
	return i.extractBinary(data, assetName)
}

func (i *Installer) getAssetName(version string) string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return fmt.Sprintf("%s-%s-macos-arm64.tar.gz", binaryName, version)
		}
		return fmt.Sprintf("%s-%s-macos-x64.tar.gz", binaryName, version)
	case "linux":
		if runtime.GOARCH == "arm64" {
			return fmt.Sprintf("%s-%s-linux-arm64.tar.gz", binaryName, version)
		}
		return fmt.Sprintf("%s-%s-linux-x64.tar.gz", binaryName, version)
	case "windows":
		// Windows releases use x64 naming
		return fmt.Sprintf("%s-%s-windows-x64.zip", binaryName, version)
	default:
		return ""
	}
}

func (i *Installer) extractBinary(data []byte, assetName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".tar.gz") {
		return i.extractFromTarGzCLI(data)
	} else if strings.HasSuffix(assetName, ".zip") {
		return i.extractFromZipCLI(data)
	}
	return nil, fmt.Errorf("unsupported archive format")
}

func (i *Installer) extractFromTarGzCLI(data []byte) ([]byte, error) {
	// Create a temporary file for the archive
	tmpArchive, err := os.CreateTemp("", "slide-mcp-*.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpArchive.Name())
	defer tmpArchive.Close()

	// Write archive data to temp file
	if _, err := tmpArchive.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpArchive.Close()

	// Create temp directory for extraction
	tmpDir, err := os.MkdirTemp("", "slide-mcp-extract-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract using tar command
	cmd := exec.Command("tar", "-xzf", tmpArchive.Name(), "-C", tmpDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	// Find the binary file
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read extracted dir: %w", err)
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), binaryName) && !entry.IsDir() {
			binaryPath := filepath.Join(tmpDir, entry.Name())
			return os.ReadFile(binaryPath)
		}
	}

	return nil, fmt.Errorf("binary not found in archive")
}

func (i *Installer) extractFromZipCLI(data []byte) ([]byte, error) {
	// Use Go's built-in zip package for better Windows compatibility
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip archive: %w", err)
	}

	// Find the binary file in the archive
	for _, file := range reader.File {
		// Skip directories (they end with /)
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		if strings.Contains(file.Name, binaryName) {
			// Open the file in the archive
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open file in archive: %w", err)
			}
			defer rc.Close()

			// Read the file contents
			fileData, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read file from archive: %w", err)
			}

			return fileData, nil
		}
	}

	return nil, fmt.Errorf("binary not found in archive")
}

func (i *Installer) installBinary(data []byte, installPath string) error {
	// Ensure install directory exists
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Write binary
	if err := os.WriteFile(installPath, data, 0755); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	return nil
}

func (i *Installer) updateClaudeConfig(apiKey, binaryPath string) error {
	configPath := i.getClaudeConfigPath()
	if configPath == "" {
		return fmt.Errorf("unsupported platform")
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing config or create new one
	var config ClaudeConfig
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing config: %w", err)
		}
	}

	// Initialize MCPServers if nil
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServer)
	}

	// Add/update slide server with version tracking
	config.MCPServers["slide"] = MCPServer{
		Command: binaryPath,
		Env: map[string]string{
			"SLIDE_API_KEY": apiKey,
		},
		Version: i.latestVersion, // Store the installed version
	}

	// Write updated config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (i *Installer) uninstall() {
	dialog.ShowConfirm("Confirm Uninstall",
		"Are you sure you want to uninstall Slide MCP Server?\n\nThis will remove the binary and update your Claude Desktop configuration.",
		func(confirmed bool) {
			if confirmed {
				i.performUninstall()
			}
		}, i.window)
}

func (i *Installer) performUninstall() {
	i.installBtn.Disable()
	i.uninstallBtn.Disable()
	i.statusLabel.SetText("Uninstalling...")

	go func() {
		defer func() {
			i.refreshUI()
		}()

		if err := i.performUninstallSteps(); err != nil {
			dialog.ShowError(err, i.window)
			i.statusLabel.SetText("Uninstall failed")
		} else {
			dialog.ShowInformation("Success", "Slide MCP Server uninstalled successfully!\n\nRestart Claude Desktop for changes to take effect.", i.window)
			i.statusLabel.SetText("Uninstalled successfully")
		}
	}()
}

func (i *Installer) performUninstallSteps() error {
	// Remove binary - use the current install path (which may be custom)
	installPath := i.getInstallPath()
	if installPath != "" {
		if err := os.Remove(installPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
	}

	// Update Claude config
	configPath := i.getClaudeConfigPath()
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			var config ClaudeConfig
			if err := json.Unmarshal(data, &config); err == nil {
				delete(config.MCPServers, "slide")

				if updatedData, err := json.MarshalIndent(config, "", "  "); err == nil {
					os.WriteFile(configPath, updatedData, 0644)
				}
			}
		}
	}

	return nil
}

func (i *Installer) browse() {
	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil || dir == nil {
			return
		}
		newPath := filepath.Join(dir.Path(), binaryName)
		if runtime.GOOS == "windows" {
			newPath += ".exe"
		}
		i.installPathEntry.SetText(newPath)
	}, i.window)
}

func (i *Installer) update() {
	i.install() // Use the same install logic for updates
}

func (i *Installer) getVersionInfo() {
	// Get latest version from GitHub
	client := github.NewClient(nil)
	ctx := context.Background()

	release, _, err := client.Repositories.GetLatestRelease(ctx, repoOwner, repoName)
	if err == nil && release.TagName != nil {
		i.latestVersion = *release.TagName
	}

	// Get installed version
	i.installedVersion = i.getInstalledVersion()

	// Update UI on main thread
	i.refreshVersionInfo()
}

func (i *Installer) refreshVersionInfo() {
	versionText := fmt.Sprintf("Installed: %s | Latest: %s", i.installedVersion, i.latestVersion)
	i.versionLabel.SetText(versionText)

	// Set default install path if empty
	if i.installPathEntry.Text == "" {
		i.installPathEntry.SetText(i.getInstallPath())
	}
}

func (i *Installer) getInstalledVersion() string {
	installPath := i.getInstallPath()
	if installPath == "" {
		return "Not installed"
	}

	// Check if binary exists
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return "Not installed"
	}

	// First, try to get version from Claude config
	configPath := i.getClaudeConfigPath()
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			var config ClaudeConfig
			if err := json.Unmarshal(data, &config); err == nil {
				if slideServer, exists := config.MCPServers["slide"]; exists && slideServer.Version != "" {
					return slideServer.Version
				}
			}
		}
	}

	// Fall back to trying to get version from the binary
	cmd := exec.Command(installPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "Unknown"
	}

	return version
}

func (i *Installer) compareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}

		if p1 < p2 {
			return -1
		} else if p1 > p2 {
			return 1
		}
	}

	return 0
}
