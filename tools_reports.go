package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Cache structures for reducing redundant API calls
type reportCache struct {
	agents  map[string]*Agent
	devices map[string]*Device
	clients map[string]string // client ID -> name
	mu      sync.RWMutex
}

var globalReportCache = &reportCache{
	agents:  make(map[string]*Agent),
	devices: make(map[string]*Device),
	clients: make(map[string]string),
}

// getCachedAgent retrieves agent info from cache or API
func getCachedAgent(agentID string) (*Agent, error) {
	globalReportCache.mu.RLock()
	if agent, exists := globalReportCache.agents[agentID]; exists {
		globalReportCache.mu.RUnlock()
		return agent, nil
	}
	globalReportCache.mu.RUnlock()

	// Not in cache, fetch from API
	agentData, err := getAgent(map[string]interface{}{"agent_id": agentID})
	if err != nil {
		return nil, err
	}

	var agent Agent
	if err := json.Unmarshal([]byte(agentData), &agent); err != nil {
		return nil, err
	}

	// Store in cache
	globalReportCache.mu.Lock()
	globalReportCache.agents[agentID] = &agent
	globalReportCache.mu.Unlock()

	return &agent, nil
}

// getCachedDevice retrieves device info from cache or API
func getCachedDevice(deviceID string) (*Device, error) {
	if deviceID == "" {
		return nil, nil
	}

	globalReportCache.mu.RLock()
	if device, exists := globalReportCache.devices[deviceID]; exists {
		globalReportCache.mu.RUnlock()
		return device, nil
	}
	globalReportCache.mu.RUnlock()

	// Not in cache, fetch from API
	deviceData, err := getDevice(map[string]interface{}{"device_id": deviceID})
	if err != nil {
		return nil, err
	}

	var device Device
	if err := json.Unmarshal([]byte(deviceData), &device); err != nil {
		return nil, err
	}

	// Store in cache
	globalReportCache.mu.Lock()
	globalReportCache.devices[deviceID] = &device
	globalReportCache.mu.Unlock()

	return &device, nil
}

// getCachedClientName retrieves client name from cache or API
func getCachedClientName(clientID string) string {
	if clientID == "" {
		return ""
	}

	globalReportCache.mu.RLock()
	if name, exists := globalReportCache.clients[clientID]; exists {
		globalReportCache.mu.RUnlock()
		return name
	}
	globalReportCache.mu.RUnlock()

	// This assumes getClientName function exists and fetches from API
	name := getClientName(clientID)

	// Store in cache
	globalReportCache.mu.Lock()
	globalReportCache.clients[clientID] = name
	globalReportCache.mu.Unlock()

	return name
}

// clearReportCache clears the global cache to ensure fresh data
func clearReportCache() {
	globalReportCache.mu.Lock()
	defer globalReportCache.mu.Unlock()

	globalReportCache.agents = make(map[string]*Agent)
	globalReportCache.devices = make(map[string]*Device)
	globalReportCache.clients = make(map[string]string)
}

// ReportData structures
type BackupStats struct {
	Total           int            `json:"total"`
	Successful      int            `json:"successful"`
	Failed          int            `json:"failed"`
	InProgress      int            `json:"in_progress"`
	SuccessRate     float64        `json:"success_rate"`
	FailuresByError map[string]int `json:"failures_by_error,omitempty"`
}

type SnapshotStats struct {
	Total              int `json:"total"`
	Active             int `json:"active"`
	Deleted            int `json:"deleted"`
	DeletedByRetention int `json:"deleted_by_retention"`
	DeletedManually    int `json:"deleted_manually"`
	DeletedOther       int `json:"deleted_other"`
	LocalStorage       int `json:"local_storage"`
	CloudStorage       int `json:"cloud_storage"`
}

type DailyReport struct {
	Date       string        `json:"date"`
	Backups    BackupStats   `json:"backups"`
	Snapshots  SnapshotStats `json:"snapshots"`
	AgentID    string        `json:"agent_id,omitempty"`
	AgentName  string        `json:"agent_name,omitempty"`
	DeviceID   string        `json:"device_id,omitempty"`
	DeviceName string        `json:"device_name,omitempty"`
	ClientID   string        `json:"client_id,omitempty"`
	ClientName string        `json:"client_name,omitempty"`
}

// handleReportsTool handles all report-related operations through a single meta-tool
func handleReportsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Clear cache at the start of each report generation to ensure fresh data
	clearReportCache()

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_reports", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_reports in '%s' mode", operation, toolsMode)
	}

	// Set verbose environment variable if verbose flag is true
	verbose, _ := args["verbose"].(bool)
	if verbose {
		os.Setenv("SLIDE_REPORTS_VERBOSE", "true")
		defer os.Unsetenv("SLIDE_REPORTS_VERBOSE")
	}

	switch operation {
	case "daily_backup_snapshot":
		return generateDailyBackupSnapshotReport(args)
	case "weekly_backup_snapshot":
		return generateWeeklyBackupSnapshotReport(args)
	case "monthly_backup_snapshot":
		return generateMonthlyBackupSnapshotReport(args)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// generateDailyBackupSnapshotReport generates daily statistics for backups and snapshots
func generateDailyBackupSnapshotReport(args map[string]interface{}) (string, error) {
	// Parse date parameter
	dateStr, _ := args["date"].(string)
	var targetDate time.Time
	var err error

	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return "", fmt.Errorf("invalid date format. Use YYYY-MM-DD: %w", err)
		}
	} else {
		// Default to today
		targetDate = time.Now().UTC().Truncate(24 * time.Hour)
	}

	// Get output format
	format, _ := args["format"].(string)
	if format == "" {
		format = "json"
	}

	// Build report based on scope
	var reports []DailyReport

	// Check if specific agent requested
	if agentID, ok := args["agent_id"].(string); ok && agentID != "" {
		report, err := generateAgentDailyReport(agentID, targetDate)
		if err != nil {
			return "", fmt.Errorf("failed to generate agent report: %w", err)
		}
		reports = append(reports, *report)
	} else if deviceID, ok := args["device_id"].(string); ok && deviceID != "" {
		// Generate report for all agents on a device
		agentReports, err := generateDeviceDailyReport(deviceID, targetDate)
		if err != nil {
			return "", fmt.Errorf("failed to generate device report: %w", err)
		}
		reports = agentReports
	} else if clientID, ok := args["client_id"].(string); ok && clientID != "" {
		// Generate report for all agents under a client
		clientReports, err := generateClientDailyReport(clientID, targetDate)
		if err != nil {
			return "", fmt.Errorf("failed to generate client report: %w", err)
		}
		reports = clientReports
	} else {
		// Generate summary for all agents
		allReports, err := generateAllAgentsDailyReport(targetDate)
		if err != nil {
			return "", fmt.Errorf("failed to generate all agents report: %w", err)
		}
		reports = allReports
	}

	// Format output
	if format == "markdown" {
		return formatReportsAsMarkdown(reports, targetDate), nil
	}

	// Return as JSON
	result := map[string]interface{}{
		"date":    targetDate.Format("2006-01-02"),
		"reports": reports,
		"_metadata": map[string]interface{}{
			"description": "Daily backup and snapshot statistics",
			"guidance":    "Use this data to identify backup failures, storage trends, and deletion patterns",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// generateWeeklyBackupSnapshotReport generates weekly statistics for backups and snapshots
func generateWeeklyBackupSnapshotReport(args map[string]interface{}) (string, error) {
	// Parse date parameter
	dateStr, _ := args["date"].(string)
	var targetDate time.Time
	var err error

	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return "", fmt.Errorf("invalid date format. Use YYYY-MM-DD: %w", err)
		}
	} else {
		// Default to current week
		targetDate = time.Now().UTC()
	}

	// Find the start of the week (Sunday)
	weekday := int(targetDate.Weekday())
	startOfWeek := targetDate.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	// Get output format
	format, _ := args["format"].(string)
	if format == "" {
		format = "json"
	}

	// Collect reports for each day of the week
	var weekReports [][]DailyReport
	dayLabels := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	// Add verbose logging
	verbose, _ := args["verbose"].(bool)

	for i := 0; i < 7; i++ {
		currentDate := startOfWeek.AddDate(0, 0, i)

		if verbose {
			fmt.Fprintf(os.Stderr, "[Weekly Report] Processing %s (%s)...\n", dayLabels[i], currentDate.Format("2006-01-02"))
		}

		// Build report based on scope
		var dayReports []DailyReport

		if agentID, ok := args["agent_id"].(string); ok && agentID != "" {
			report, err := generateAgentDailyReport(agentID, currentDate)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Weekly Report] Error processing agent %s on %s: %v\n", agentID, currentDate.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = append(dayReports, *report)
		} else if deviceID, ok := args["device_id"].(string); ok && deviceID != "" {
			reports, err := generateDeviceDailyReport(deviceID, currentDate)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Weekly Report] Error processing device %s on %s: %v\n", deviceID, currentDate.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = reports
		} else if clientID, ok := args["client_id"].(string); ok && clientID != "" {
			reports, err := generateClientDailyReport(clientID, currentDate)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Weekly Report] Error processing client %s on %s: %v\n", clientID, currentDate.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = reports
		} else {
			reports, err := generateAllAgentsDailyReport(currentDate)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Weekly Report] Error processing all agents on %s: %v\n", currentDate.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = reports
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Weekly Report] Found %d agent reports for %s\n", len(dayReports), currentDate.Format("2006-01-02"))
		}

		weekReports = append(weekReports, dayReports)
	}

	// Format output
	if format == "markdown" {
		return formatWeeklyReportsAsMarkdown(weekReports, startOfWeek, dayLabels), nil
	}

	// Return as JSON
	result := map[string]interface{}{
		"week_start":    startOfWeek.Format("2006-01-02"),
		"week_end":      endOfWeek.AddDate(0, 0, -1).Format("2006-01-02"),
		"daily_reports": weekReports,
		"_metadata": map[string]interface{}{
			"description": "Weekly backup and snapshot statistics (Sunday to Saturday)",
			"guidance":    "Use this data to identify weekly patterns and trends",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// generateMonthlyBackupSnapshotReport generates monthly statistics for backups and snapshots
func generateMonthlyBackupSnapshotReport(args map[string]interface{}) (string, error) {
	// Parse date parameter
	dateStr, _ := args["date"].(string)
	var targetDate time.Time
	var err error

	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return "", fmt.Errorf("invalid date format. Use YYYY-MM-DD: %w", err)
		}
	} else {
		// Default to current month
		targetDate = time.Now().UTC()
	}

	// Get the first and last day of the month
	year, month, _ := targetDate.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Get output format
	format, _ := args["format"].(string)
	if format == "" {
		format = "json"
	}

	// Collect reports for each day of the month
	monthReports := make(map[int][]DailyReport)

	// Add verbose logging
	verbose, _ := args["verbose"].(bool)
	totalDays := lastDay.Day()
	currentDay := 0

	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		currentDay++
		if verbose {
			fmt.Fprintf(os.Stderr, "[Monthly Report] Processing day %d/%d (%s)...\n", currentDay, totalDays, d.Format("2006-01-02"))
		}

		// Build report based on scope
		var dayReports []DailyReport

		if agentID, ok := args["agent_id"].(string); ok && agentID != "" {
			report, err := generateAgentDailyReport(agentID, d)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Monthly Report] Error processing agent %s on %s: %v\n", agentID, d.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = append(dayReports, *report)
		} else if deviceID, ok := args["device_id"].(string); ok && deviceID != "" {
			reports, err := generateDeviceDailyReport(deviceID, d)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Monthly Report] Error processing device %s on %s: %v\n", deviceID, d.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = reports
		} else if clientID, ok := args["client_id"].(string); ok && clientID != "" {
			reports, err := generateClientDailyReport(clientID, d)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Monthly Report] Error processing client %s on %s: %v\n", clientID, d.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = reports
		} else {
			reports, err := generateAllAgentsDailyReport(d)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Monthly Report] Error processing all agents on %s: %v\n", d.Format("2006-01-02"), err)
				}
				continue
			}
			dayReports = reports
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Monthly Report] Found %d agent reports for %s\n", len(dayReports), d.Format("2006-01-02"))
		}

		monthReports[d.Day()] = dayReports
	}

	// Format output
	if format == "markdown" {
		return formatMonthlyReportsAsMarkdown(monthReports, firstDay), nil
	}

	// Return as JSON
	result := map[string]interface{}{
		"month":         firstDay.Format("2006-01"),
		"month_name":    firstDay.Format("January 2006"),
		"daily_reports": monthReports,
		"_metadata": map[string]interface{}{
			"description": "Monthly backup and snapshot statistics",
			"guidance":    "Use this data to identify monthly patterns and long-term trends",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// generateAgentDailyReport generates report for a specific agent
func generateAgentDailyReport(agentID string, date time.Time) (*DailyReport, error) {
	report := &DailyReport{
		Date:    date.Format("2006-01-02"),
		AgentID: agentID,
	}

	// Get agent details from cache
	agent, err := getCachedAgent(agentID)
	if err == nil && agent != nil {
		report.AgentName = agent.DisplayName
		report.DeviceID = agent.DeviceID
		if agent.ClientID != nil {
			report.ClientID = *agent.ClientID
			report.ClientName = getCachedClientName(report.ClientID)
		}

		// Get device name from cache
		if report.DeviceID != "" {
			device, err := getCachedDevice(report.DeviceID)
			if err == nil && device != nil {
				report.DeviceName = device.DisplayName
			}
		}
	}

	// Calculate backup and snapshot stats in parallel
	var backupStats *BackupStats
	var snapshotStats *SnapshotStats
	var backupErr, snapshotErr error

	var wg sync.WaitGroup
	wg.Add(2)

	// Fetch backup stats concurrently
	go func() {
		defer wg.Done()
		backupStats, backupErr = calculateBackupStats(agentID, date)
	}()

	// Fetch snapshot stats concurrently
	go func() {
		defer wg.Done()
		snapshotStats, snapshotErr = calculateSnapshotStats(agentID, date)
	}()

	wg.Wait()

	if backupErr != nil {
		return nil, fmt.Errorf("failed to calculate backup stats: %w", backupErr)
	}
	report.Backups = *backupStats

	if snapshotErr != nil {
		return nil, fmt.Errorf("failed to calculate snapshot stats: %w", snapshotErr)
	}
	report.Snapshots = *snapshotStats

	return report, nil
}

// generateAgentReportsConcurrent generates reports for multiple agents in parallel
func generateAgentReportsConcurrent(agents []Agent, date time.Time, maxConcurrency int) []DailyReport {
	if maxConcurrency <= 0 {
		maxConcurrency = 10 // Default concurrency limit
	}

	verbose := os.Getenv("SLIDE_REPORTS_VERBOSE") == "true"

	// Create a buffered channel to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)

	// Channel to collect results
	resultChan := make(chan *DailyReport, len(agents))

	var wg sync.WaitGroup

	for i, agent := range agents {
		wg.Add(1)
		go func(idx int, ag Agent) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if verbose && idx%10 == 0 {
				fmt.Fprintf(os.Stderr, "[Concurrent] Processing agent %d/%d (%s)...\n",
					idx+1, len(agents), ag.DisplayName)
			}

			report, err := generateAgentDailyReport(ag.AgentID, date)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Concurrent] Error processing agent %s: %v\n",
						ag.DisplayName, err)
				}
				return
			}

			resultChan <- report
		}(i, agent)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var reports []DailyReport
	for report := range resultChan {
		if report != nil {
			reports = append(reports, *report)
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[Concurrent] Completed processing %d agents\n", len(reports))
	}

	return reports
}

// generateDeviceDailyReport generates reports for all agents on a device
func generateDeviceDailyReport(deviceID string, date time.Time) ([]DailyReport, error) {
	// Get all agents on the device
	agentsData, err := listAgents(map[string]interface{}{
		"device_id": deviceID,
		"limit":     100, // Increased limit for better batching
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	var agentsList struct {
		Data []Agent `json:"data"`
	}
	if err := json.Unmarshal([]byte(agentsData), &agentsList); err != nil {
		return nil, fmt.Errorf("failed to parse agents: %w", err)
	}

	// Process agents concurrently
	reports := generateAgentReportsConcurrent(agentsList.Data, date, 10)
	return reports, nil
}

// generateClientDailyReport generates reports for all agents under a client
func generateClientDailyReport(clientID string, date time.Time) ([]DailyReport, error) {
	// Get all agents for the client
	agentsData, err := listAgents(map[string]interface{}{
		"client_id": clientID,
		"limit":     100, // Increased limit for better batching
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	var agentsList struct {
		Data []Agent `json:"data"`
	}
	if err := json.Unmarshal([]byte(agentsData), &agentsList); err != nil {
		return nil, fmt.Errorf("failed to parse agents: %w", err)
	}

	// Process agents concurrently
	reports := generateAgentReportsConcurrent(agentsList.Data, date, 10)
	return reports, nil
}

// generateAllAgentsDailyReport generates reports for all agents
func generateAllAgentsDailyReport(date time.Time) ([]DailyReport, error) {
	var allAgents []Agent
	offset := 0
	limit := 100 // Increased for better batching

	// Check if verbose logging is needed
	verbose := os.Getenv("SLIDE_REPORTS_VERBOSE") == "true"

	// First, collect all agents
	for {
		if verbose {
			fmt.Fprintf(os.Stderr, "[Agent Fetch] Fetching agents batch (offset: %d, limit: %d)...\n", offset, limit)
		}

		agentsData, err := listAgents(map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list agents: %w", err)
		}

		var agentsList struct {
			Pagination Pagination `json:"pagination"`
			Data       []Agent    `json:"data"`
		}
		if err := json.Unmarshal([]byte(agentsData), &agentsList); err != nil {
			return nil, fmt.Errorf("failed to parse agents: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Agent Fetch] Retrieved %d agents (total: %d)\n", len(agentsList.Data), agentsList.Pagination.Total)
		}

		allAgents = append(allAgents, agentsList.Data...)

		// Check if there are more agents
		if agentsList.Pagination.NextOffset == nil {
			break
		}

		// Safety check: ensure we're making progress
		newOffset := *agentsList.Pagination.NextOffset
		if newOffset == offset || len(agentsList.Data) == 0 {
			if verbose {
				fmt.Fprintf(os.Stderr, "[Agent Fetch] Pagination complete at offset %d\n", offset)
			}
			break
		}

		offset = newOffset
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[Agent Fetch] Total agents collected: %d\n", len(allAgents))
	}

	// Process all agents concurrently with higher concurrency for larger datasets
	reports := generateAgentReportsConcurrent(allAgents, date, 20)
	return reports, nil
}

// calculateBackupStats calculates backup statistics for an agent on a specific date
func calculateBackupStats(agentID string, date time.Time) (*BackupStats, error) {
	stats := &BackupStats{
		FailuresByError: make(map[string]int),
	}

	// Set date range for the day
	startDate := date
	endDate := date.Add(24 * time.Hour)

	// Try to use date filters if the API supports them
	// Format dates as ISO 8601 strings
	startDateStr := startDate.Format(time.RFC3339)
	endDateStr := endDate.Format(time.RFC3339)

	// Fetch backups for the specific date range
	offset := 0
	limit := 50
	processedCount := 0

	verbose := os.Getenv("SLIDE_REPORTS_VERBOSE") == "true"

	for {
		if verbose && offset == 0 {
			fmt.Fprintf(os.Stderr, "[Backup Stats] Fetching backups for agent %s on %s...\n", agentID, date.Format("2006-01-02"))
		}

		// Build query parameters with date filters if possible
		queryParams := map[string]interface{}{
			"agent_id": agentID,
			"limit":    limit,
			"offset":   offset,
			"sort_by":  "start_time",
		}

		// Add date filters if the API supports them (you may need to check API docs)
		// These are common parameter names, adjust if your API uses different ones
		queryParams["start_date"] = startDateStr
		queryParams["end_date"] = endDateStr

		backupsData, err := listBackups(queryParams)
		if err != nil {
			// If date filtering fails, try without date filters
			delete(queryParams, "start_date")
			delete(queryParams, "end_date")
			backupsData, err = listBackups(queryParams)
			if err != nil {
				return nil, fmt.Errorf("failed to list backups: %w", err)
			}
		}

		var backupsList struct {
			Pagination Pagination `json:"pagination"`
			Data       []Backup   `json:"data"`
		}
		if err := json.Unmarshal([]byte(backupsData), &backupsList); err != nil {
			return nil, fmt.Errorf("failed to parse backups: %w", err)
		}

		// Process backups
		foundInRange := false
		for _, backup := range backupsList.Data {
			// Parse backup start time
			backupTime, err := time.Parse(time.RFC3339, backup.StartedAt)
			if err != nil {
				continue
			}

			// Check if backup is within the target date
			if backupTime.Before(startDate) {
				// If we're seeing backups before our date range and we've already processed some,
				// we can stop since results are sorted by start_time
				if processedCount > 0 {
					goto done
				}
				continue
			}
			if backupTime.After(endDate) {
				// We've gone past our date range
				goto done
			}

			foundInRange = true
			processedCount++
			stats.Total++

			switch backup.Status {
			case "success":
				stats.Successful++
			case "failed":
				stats.Failed++
				// Track failure reasons
				if backup.ErrorMessage != nil {
					stats.FailuresByError[*backup.ErrorMessage]++
				}
			case "running", "pending":
				stats.InProgress++
			}
		}

		// If we found no backups in range and have been processing,
		// we're likely past the date range
		if !foundInRange && processedCount > 0 {
			break
		}

		// Check if there are more backups
		if backupsList.Pagination.NextOffset == nil || len(backupsList.Data) == 0 {
			break
		}

		// Update offset
		offset = *backupsList.Pagination.NextOffset
	}

done:
	// Calculate success rate
	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Successful) / float64(stats.Total) * 100
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[Backup Stats] Found %d backups for agent %s on %s\n", stats.Total, agentID, date.Format("2006-01-02"))
	}

	return stats, nil
}

// calculateSnapshotStats calculates snapshot statistics for an agent
func calculateSnapshotStats(agentID string, date time.Time) (*SnapshotStats, error) {
	stats := &SnapshotStats{}

	verbose := os.Getenv("SLIDE_REPORTS_VERBOSE") == "true"

	// We'll fetch snapshots in smaller batches and count as we go
	// This is more efficient than fetching ALL snapshots

	// Count active snapshots
	activeCount := 0
	localCount := 0
	cloudCount := 0

	offset := 0
	limit := 50

	if verbose {
		fmt.Fprintf(os.Stderr, "[Snapshot Stats] Counting active snapshots for agent %s...\n", agentID)
	}

	for {
		snapshotsData, err := listSnapshots(map[string]interface{}{
			"agent_id": agentID,
			"limit":    limit,
			"offset":   offset,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list active snapshots: %w", err)
		}

		var snapshotsList struct {
			Pagination Pagination `json:"pagination"`
			Data       []Snapshot `json:"data"`
		}
		if err := json.Unmarshal([]byte(snapshotsData), &snapshotsList); err != nil {
			return nil, fmt.Errorf("failed to parse snapshots: %w", err)
		}

		// Count snapshots and their storage locations
		for _, snapshot := range snapshotsList.Data {
			activeCount++

			// Check storage location
			hasLocal := false
			hasCloud := false
			for _, location := range snapshot.Locations {
				if location.DeviceID == agentID {
					hasLocal = true
				} else {
					hasCloud = true
				}
			}

			if hasLocal {
				localCount++
			}
			if hasCloud {
				cloudCount++
			}
		}

		// Use pagination total if available (more efficient)
		if snapshotsList.Pagination.Total > 0 && offset == 0 {
			stats.Active = snapshotsList.Pagination.Total
			// We still need to count storage locations, so continue if we haven't seen all
			if activeCount >= stats.Active {
				break
			}
		}

		if snapshotsList.Pagination.NextOffset == nil || len(snapshotsList.Data) == 0 {
			if stats.Active == 0 {
				stats.Active = activeCount
			}
			break
		}

		offset = *snapshotsList.Pagination.NextOffset
	}

	stats.LocalStorage = localCount
	stats.CloudStorage = cloudCount

	// Count deleted snapshots
	deletedCount := 0
	deletedByRetention := 0
	deletedManually := 0
	deletedOther := 0

	offset = 0

	if verbose {
		fmt.Fprintf(os.Stderr, "[Snapshot Stats] Counting deleted snapshots for agent %s...\n", agentID)
	}

	for {
		snapshotsData, err := listSnapshots(map[string]interface{}{
			"agent_id":          agentID,
			"snapshot_location": "exists_deleted",
			"limit":             limit,
			"offset":            offset,
		})
		if err != nil {
			// Some APIs might not support deleted snapshots query
			if verbose {
				fmt.Fprintf(os.Stderr, "[Snapshot Stats] Could not fetch deleted snapshots: %v\n", err)
			}
			break
		}

		var snapshotsList struct {
			Pagination Pagination `json:"pagination"`
			Data       []Snapshot `json:"data"`
		}
		if err := json.Unmarshal([]byte(snapshotsData), &snapshotsList); err != nil {
			break
		}

		// Count deletion reasons
		for _, snapshot := range snapshotsList.Data {
			deletedCount++

			// Analyze deletion reason
			for _, deletion := range snapshot.Deletions {
				switch deletion.Type {
				case "retention":
					deletedByRetention++
				case "manual":
					deletedManually++
				default:
					deletedOther++
				}
				break // Count each snapshot only once
			}
		}

		// Use pagination total if available (more efficient)
		if snapshotsList.Pagination.Total > 0 && offset == 0 {
			stats.Deleted = snapshotsList.Pagination.Total
			// We still need to count deletion reasons, so continue if we haven't seen all
			if deletedCount >= stats.Deleted {
				break
			}
		}

		if snapshotsList.Pagination.NextOffset == nil || len(snapshotsList.Data) == 0 {
			if stats.Deleted == 0 {
				stats.Deleted = deletedCount
			}
			break
		}

		offset = *snapshotsList.Pagination.NextOffset
	}

	stats.DeletedByRetention = deletedByRetention
	stats.DeletedManually = deletedManually
	stats.DeletedOther = deletedOther

	// Calculate total
	stats.Total = stats.Active + stats.Deleted

	if verbose {
		fmt.Fprintf(os.Stderr, "[Snapshot Stats] Found %d total snapshots (%d active, %d deleted) for agent %s\n",
			stats.Total, stats.Active, stats.Deleted, agentID)
	}

	return stats, nil
}

// formatReportsAsMarkdown formats the reports as markdown
func formatReportsAsMarkdown(reports []DailyReport, date time.Time) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Daily Backup & Snapshot Report - %s\n\n", date.Format("2006-01-02")))

	if len(reports) == 0 {
		sb.WriteString("No data available for the specified criteria.\n")
		return sb.String()
	}

	// Summary stats
	totalBackups := 0
	totalSuccessful := 0
	totalSnapshots := 0
	totalDeleted := 0

	for _, report := range reports {
		totalBackups += report.Backups.Total
		totalSuccessful += report.Backups.Successful
		totalSnapshots += report.Snapshots.Total
		totalDeleted += report.Snapshots.Deleted
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Agents Reporting**: %d\n", len(reports)))
	sb.WriteString(fmt.Sprintf("- **Total Backups**: %d\n", totalBackups))
	if totalBackups > 0 {
		overallSuccessRate := float64(totalSuccessful) / float64(totalBackups) * 100
		sb.WriteString(fmt.Sprintf("- **Overall Success Rate**: %.1f%%\n", overallSuccessRate))
	}
	sb.WriteString(fmt.Sprintf("- **Total Snapshots**: %d\n", totalSnapshots))
	sb.WriteString(fmt.Sprintf("- **Snapshots Deleted**: %d\n\n", totalDeleted))

	// Individual agent reports
	sb.WriteString("## Agent Details\n\n")

	for _, report := range reports {
		sb.WriteString(fmt.Sprintf("### %s\n", report.AgentName))
		if report.ClientName != "" {
			sb.WriteString(fmt.Sprintf("**Client**: %s\n", report.ClientName))
		}
		if report.DeviceName != "" {
			sb.WriteString(fmt.Sprintf("**Device**: %s\n", report.DeviceName))
		}
		sb.WriteString("\n")

		// Backup stats
		sb.WriteString("**Backups:**\n")
		sb.WriteString(fmt.Sprintf("- Total: %d\n", report.Backups.Total))
		sb.WriteString(fmt.Sprintf("- Successful: %d\n", report.Backups.Successful))
		sb.WriteString(fmt.Sprintf("- Failed: %d\n", report.Backups.Failed))
		if report.Backups.InProgress > 0 {
			sb.WriteString(fmt.Sprintf("- In Progress: %d\n", report.Backups.InProgress))
		}
		sb.WriteString(fmt.Sprintf("- Success Rate: %.1f%%\n", report.Backups.SuccessRate))

		if len(report.Backups.FailuresByError) > 0 {
			sb.WriteString("\n**Failure Reasons:**\n")
			for errMsg, count := range report.Backups.FailuresByError {
				sb.WriteString(fmt.Sprintf("- %s: %d\n", errMsg, count))
			}
		}

		// Snapshot stats
		sb.WriteString("\n**Snapshots:**\n")
		sb.WriteString(fmt.Sprintf("- Total: %d\n", report.Snapshots.Total))
		sb.WriteString(fmt.Sprintf("- Active: %d\n", report.Snapshots.Active))
		sb.WriteString(fmt.Sprintf("- Deleted: %d\n", report.Snapshots.Deleted))
		if report.Snapshots.Deleted > 0 {
			sb.WriteString(fmt.Sprintf("  - By Retention Policy: %d\n", report.Snapshots.DeletedByRetention))
			sb.WriteString(fmt.Sprintf("  - Manually Deleted: %d\n", report.Snapshots.DeletedManually))
			sb.WriteString(fmt.Sprintf("  - Other Reasons: %d\n", report.Snapshots.DeletedOther))
		}
		sb.WriteString(fmt.Sprintf("- Local Storage: %d\n", report.Snapshots.LocalStorage))
		sb.WriteString(fmt.Sprintf("- Cloud Storage: %d\n\n", report.Snapshots.CloudStorage))

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// formatWeeklyReportsAsMarkdown formats weekly reports as markdown
func formatWeeklyReportsAsMarkdown(weekReports [][]DailyReport, startOfWeek time.Time, dayLabels []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Weekly Backup & Snapshot Report\n"))
	sb.WriteString(fmt.Sprintf("## Week of %s to %s\n\n",
		startOfWeek.Format("January 2, 2006"),
		startOfWeek.AddDate(0, 0, 6).Format("January 2, 2006")))

	// Calculate weekly totals
	weeklyBackupTotal := 0
	weeklySuccessTotal := 0
	weeklySnapshotTotal := 0
	weeklyDeletedTotal := 0
	agentSet := make(map[string]bool)

	for _, dayReports := range weekReports {
		for _, report := range dayReports {
			weeklyBackupTotal += report.Backups.Total
			weeklySuccessTotal += report.Backups.Successful
			weeklySnapshotTotal += report.Snapshots.Total
			weeklyDeletedTotal += report.Snapshots.Deleted
			agentSet[report.AgentID] = true
		}
	}

	// Weekly summary
	sb.WriteString("## Weekly Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Unique Agents**: %d\n", len(agentSet)))
	sb.WriteString(fmt.Sprintf("- **Total Backups**: %d\n", weeklyBackupTotal))
	if weeklyBackupTotal > 0 {
		weeklySuccessRate := float64(weeklySuccessTotal) / float64(weeklyBackupTotal) * 100
		sb.WriteString(fmt.Sprintf("- **Weekly Success Rate**: %.1f%%\n", weeklySuccessRate))
	}
	sb.WriteString(fmt.Sprintf("- **Total Snapshots**: %d\n", weeklySnapshotTotal))
	sb.WriteString(fmt.Sprintf("- **Snapshots Deleted**: %d\n\n", weeklyDeletedTotal))

	// Daily breakdown
	sb.WriteString("## Daily Breakdown\n\n")

	for i, dayReports := range weekReports {
		currentDate := startOfWeek.AddDate(0, 0, i)
		sb.WriteString(fmt.Sprintf("### %s - %s\n\n", dayLabels[i], currentDate.Format("Jan 2")))

		if len(dayReports) == 0 {
			sb.WriteString("No data available for this day.\n\n")
			continue
		}

		// Daily totals
		dayBackupTotal := 0
		daySuccessTotal := 0
		daySnapshotTotal := 0

		for _, report := range dayReports {
			dayBackupTotal += report.Backups.Total
			daySuccessTotal += report.Backups.Successful
			daySnapshotTotal += report.Snapshots.Total
		}

		sb.WriteString(fmt.Sprintf("- Agents: %d\n", len(dayReports)))
		sb.WriteString(fmt.Sprintf("- Backups: %d (", dayBackupTotal))
		if dayBackupTotal > 0 {
			daySuccessRate := float64(daySuccessTotal) / float64(dayBackupTotal) * 100
			sb.WriteString(fmt.Sprintf("%.1f%% success", daySuccessRate))
		} else {
			sb.WriteString("no backups")
		}
		sb.WriteString(")\n")
		sb.WriteString(fmt.Sprintf("- Snapshots: %d\n\n", daySnapshotTotal))
	}

	return sb.String()
}

// formatMonthlyReportsAsMarkdown formats monthly reports as markdown with calendar table
func formatMonthlyReportsAsMarkdown(monthReports map[int][]DailyReport, firstDay time.Time) string {
	var sb strings.Builder

	year, month, _ := firstDay.Date()
	monthName := firstDay.Format("January 2006")

	sb.WriteString(fmt.Sprintf("# Monthly Backup & Snapshot Report - %s\n\n", monthName))

	// Calculate monthly totals
	monthlyBackupTotal := 0
	monthlySuccessTotal := 0
	monthlySnapshotTotal := 0
	monthlyDeletedTotal := 0
	agentSet := make(map[string]bool)

	for _, dayReports := range monthReports {
		for _, report := range dayReports {
			monthlyBackupTotal += report.Backups.Total
			monthlySuccessTotal += report.Backups.Successful
			monthlySnapshotTotal += report.Snapshots.Total
			monthlyDeletedTotal += report.Snapshots.Deleted
			agentSet[report.AgentID] = true
		}
	}

	// Monthly summary
	sb.WriteString("## Monthly Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Unique Agents**: %d\n", len(agentSet)))
	sb.WriteString(fmt.Sprintf("- **Total Backups**: %d\n", monthlyBackupTotal))
	if monthlyBackupTotal > 0 {
		monthlySuccessRate := float64(monthlySuccessTotal) / float64(monthlyBackupTotal) * 100
		sb.WriteString(fmt.Sprintf("- **Monthly Success Rate**: %.1f%%\n", monthlySuccessRate))
	}
	sb.WriteString(fmt.Sprintf("- **Total Snapshots**: %d\n", monthlySnapshotTotal))
	sb.WriteString(fmt.Sprintf("- **Snapshots Deleted**: %d\n\n", monthlyDeletedTotal))

	// Calendar view
	sb.WriteString("## Calendar View\n\n")
	sb.WriteString("| Sun | Mon | Tue | Wed | Thu | Fri | Sat |\n")
	sb.WriteString("|-----|-----|-----|-----|-----|-----|-----|\n")

	// Get the first day of the month and its weekday
	firstWeekday := int(firstDay.Weekday())
	lastDay := firstDay.AddDate(0, 1, -1).Day()

	// Create calendar grid
	currentDay := 1
	for week := 0; week < 6; week++ {
		if currentDay > lastDay {
			break
		}

		sb.WriteString("|")
		for weekday := 0; weekday < 7; weekday++ {
			if (week == 0 && weekday < firstWeekday) || currentDay > lastDay {
				sb.WriteString("     |")
			} else {
				// Get stats for this day
				dayReports := monthReports[currentDay]

				if len(dayReports) == 0 {
					sb.WriteString(fmt.Sprintf(" %2d  |", currentDay))
				} else {
					// Calculate day stats
					dayBackupTotal := 0
					daySuccessTotal := 0

					for _, report := range dayReports {
						dayBackupTotal += report.Backups.Total
						daySuccessTotal += report.Backups.Successful
					}

					// Show day with backup count and success indicator
					if dayBackupTotal == 0 {
						sb.WriteString(fmt.Sprintf(" %2d  |", currentDay))
					} else {
						successRate := float64(daySuccessTotal) / float64(dayBackupTotal) * 100
						var indicator string
						if successRate >= 90 {
							indicator = "✓" // High success
						} else if successRate >= 50 {
							indicator = "~" // Partial success
						} else {
							indicator = "✗" // Low success
						}
						sb.WriteString(fmt.Sprintf(" %2d%s |", currentDay, indicator))
					}
				}
				currentDay++
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n**Legend:** ✓ = ≥90% success, ~ = 50-89% success, ✗ = <50% success\n\n")

	// Detailed daily breakdown
	sb.WriteString("## Daily Details\n\n")

	for day := 1; day <= lastDay; day++ {
		dayReports := monthReports[day]
		if len(dayReports) == 0 {
			continue
		}

		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		sb.WriteString(fmt.Sprintf("### %s\n", date.Format("January 2 (Monday)")))

		// Daily totals
		dayBackupTotal := 0
		daySuccessTotal := 0
		dayFailedTotal := 0
		daySnapshotTotal := 0

		for _, report := range dayReports {
			dayBackupTotal += report.Backups.Total
			daySuccessTotal += report.Backups.Successful
			dayFailedTotal += report.Backups.Failed
			daySnapshotTotal += report.Snapshots.Total
		}

		sb.WriteString(fmt.Sprintf("- **Agents**: %d\n", len(dayReports)))
		sb.WriteString(fmt.Sprintf("- **Backups**: %d total (%d successful, %d failed)\n",
			dayBackupTotal, daySuccessTotal, dayFailedTotal))
		if dayBackupTotal > 0 {
			daySuccessRate := float64(daySuccessTotal) / float64(dayBackupTotal) * 100
			sb.WriteString(fmt.Sprintf("- **Success Rate**: %.1f%%\n", daySuccessRate))
		}
		sb.WriteString(fmt.Sprintf("- **Snapshots**: %d\n\n", daySnapshotTotal))
	}

	return sb.String()
}

// getReportsToolInfo returns the tool definition for the reports meta-tool
func getReportsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_reports",
		Description: "Generate statistical reports about backups, snapshots, and system health. Provides pre-calculated metrics to help LLMs analyze data without complex calculations.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The type of report to generate",
					"enum":        []string{"daily_backup_snapshot", "weekly_backup_snapshot", "monthly_backup_snapshot"},
				},
				// Parameters for all report types
				"date": map[string]interface{}{
					"type":        "string",
					"description": "Date for the report in YYYY-MM-DD format (defaults to today/current week/current month)",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter report by specific agent ID",
				},
				"device_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter report by device ID (includes all agents on device)",
				},
				"client_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter report by client ID (includes all agents for client)",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format for the report",
					"enum":        []string{"json", "markdown"},
				},
				"verbose": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable verbose progress logging to stderr (useful for long operations)",
				},
			},
			"required": []string{"operation"},
		},
	}
}
