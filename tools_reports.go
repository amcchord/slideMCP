package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Report data structures
type BackupStats struct {
	Total           int            `json:"total"`
	Successful      int            `json:"successful"`
	Failed          int            `json:"failed"`
	InProgress      int            `json:"in_progress"`
	SuccessRate     float64        `json:"success_rate"`
	FailuresByError map[string]int `json:"failures_by_error,omitempty"`
}

type SnapshotStats struct {
	Total        int `json:"total"`
	Active       int `json:"active"`
	Deleted      int `json:"deleted"`
	LocalStorage int `json:"local_storage"`
	CloudStorage int `json:"cloud_storage"`
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

// handleReportsTool handles all report-related operations
func handleReportsTool(args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required")
	}

	// Check if operation is allowed in current tools mode
	if !isOperationAllowed("slide_reports", operation) {
		return "", fmt.Errorf("operation '%s' not available for slide_reports in '%s' mode", operation, toolsMode)
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

// generateDailyBackupSnapshotReport generates daily statistics
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

	// Get all agents based on scope
	agents, err := getAgentsForScope(args)
	if err != nil {
		return "", fmt.Errorf("failed to get agents: %w", err)
	}

	// Generate reports for each agent
	var reports []DailyReport
	for _, agent := range agents {
		report, err := generateAgentDailyReport(agent, targetDate)
		if err != nil {
			// Skip agents with errors
			continue
		}
		reports = append(reports, *report)
	}

	// Format output
	if format == "markdown" {
		return formatDailyReportsAsMarkdown(reports, targetDate), nil
	}

	// Return as JSON
	result := map[string]interface{}{
		"date":    targetDate.Format("2006-01-02"),
		"reports": reports,
		"_metadata": map[string]interface{}{
			"description": "Daily backup and snapshot statistics",
			"guidance":    "Use this data to identify backup failures, storage trends, and system health",
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonData), nil
}

// generateWeeklyBackupSnapshotReport generates weekly statistics
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

	// Get output format
	format, _ := args["format"].(string)
	if format == "" {
		format = "json"
	}

	// Get all agents based on scope
	agents, err := getAgentsForScope(args)
	if err != nil {
		return "", fmt.Errorf("failed to get agents: %w", err)
	}

	// Collect reports for each day of the week
	dayLabels := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	weekReports := make([][]DailyReport, 7)

	for i := 0; i < 7; i++ {
		currentDate := startOfWeek.AddDate(0, 0, i)
		var dayReports []DailyReport

		for _, agent := range agents {
			report, err := generateAgentDailyReport(agent, currentDate)
			if err != nil {
				continue
			}
			dayReports = append(dayReports, *report)
		}

		weekReports[i] = dayReports
	}

	// Format output
	if format == "markdown" {
		return formatWeeklyReportsAsMarkdown(weekReports, startOfWeek, dayLabels), nil
	}

	// Return as JSON
	result := map[string]interface{}{
		"week_start":    startOfWeek.Format("2006-01-02"),
		"week_end":      startOfWeek.AddDate(0, 0, 6).Format("2006-01-02"),
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

// generateMonthlyBackupSnapshotReport generates monthly statistics
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

	// Get all agents based on scope
	agents, err := getAgentsForScope(args)
	if err != nil {
		return "", fmt.Errorf("failed to get agents: %w", err)
	}

	// Collect reports for each day of the month
	monthReports := make(map[int][]DailyReport)

	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		var dayReports []DailyReport

		for _, agent := range agents {
			report, err := generateAgentDailyReport(agent, d)
			if err != nil {
				continue
			}
			dayReports = append(dayReports, *report)
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

// getAgentsForScope returns the list of agents based on the provided scope
func getAgentsForScope(args map[string]interface{}) ([]Agent, error) {
	var agents []Agent

	// Check if specific agent requested
	if agentID, ok := args["agent_id"].(string); ok && agentID != "" {
		agentData, err := getAgent(map[string]interface{}{"agent_id": agentID})
		if err != nil {
			return nil, err
		}

		var agent Agent
		if err := json.Unmarshal([]byte(agentData), &agent); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	} else if deviceID, ok := args["device_id"].(string); ok && deviceID != "" {
		// Get all agents on device
		agentsData, err := listAgents(map[string]interface{}{
			"device_id": deviceID,
			"limit":     50,
		})
		if err != nil {
			return nil, err
		}

		var agentsList struct {
			Data []Agent `json:"data"`
		}
		if err := json.Unmarshal([]byte(agentsData), &agentsList); err != nil {
			return nil, err
		}
		agents = agentsList.Data
	} else if clientID, ok := args["client_id"].(string); ok && clientID != "" {
		// Get all agents for client
		agentsData, err := listAgents(map[string]interface{}{
			"client_id": clientID,
			"limit":     50,
		})
		if err != nil {
			return nil, err
		}

		var agentsList struct {
			Data []Agent `json:"data"`
		}
		if err := json.Unmarshal([]byte(agentsData), &agentsList); err != nil {
			return nil, err
		}
		agents = agentsList.Data
	} else {
		// Get all agents
		offset := 0
		limit := 50
		maxIterations := 100 // Safety limit

		for iterations := 0; iterations < maxIterations; iterations++ {
			agentsData, err := listAgents(map[string]interface{}{
				"limit":  limit,
				"offset": offset,
			})
			if err != nil {
				return nil, err
			}

			var agentsList struct {
				Pagination Pagination `json:"pagination"`
				Data       []Agent    `json:"data"`
			}
			if err := json.Unmarshal([]byte(agentsData), &agentsList); err != nil {
				return nil, err
			}

			agents = append(agents, agentsList.Data...)

			// Check if we have more data
			if agentsList.Pagination.NextOffset == nil || len(agentsList.Data) == 0 {
				break
			}

			// Safety check: ensure offset is progressing
			newOffset := *agentsList.Pagination.NextOffset
			if newOffset <= offset {
				// Not making progress, break to avoid infinite loop
				break
			}
			offset = newOffset

			// Safety check: reasonable limit
			if offset > 5000 {
				// Reached a reasonable maximum
				break
			}
		}
	}

	return agents, nil
}

// generateAgentDailyReport generates a report for a specific agent on a specific date
func generateAgentDailyReport(agent Agent, date time.Time) (*DailyReport, error) {
	report := &DailyReport{
		Date:       date.Format("2006-01-02"),
		AgentID:    agent.AgentID,
		AgentName:  agent.DisplayName,
		DeviceID:   agent.DeviceID,
		ClientID:   "",
		ClientName: "",
	}

	// Use hostname if display name is empty
	if report.AgentName == "" {
		report.AgentName = agent.Hostname
	}

	// Set client info if available
	if agent.ClientID != nil {
		report.ClientID = *agent.ClientID
		report.ClientName = getClientName(report.ClientID)
	}

	// Get device name
	if report.DeviceID != "" {
		deviceData, err := getDevice(map[string]interface{}{"device_id": report.DeviceID})
		if err == nil {
			var device Device
			if err := json.Unmarshal([]byte(deviceData), &device); err == nil {
				report.DeviceName = device.DisplayName
				if report.DeviceName == "" {
					report.DeviceName = device.Hostname
				}
			}
		}
	}

	// Calculate backup stats
	backupStats, err := calculateBackupStatsForDate(agent.AgentID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate backup stats: %w", err)
	}
	report.Backups = *backupStats

	// Calculate snapshot stats
	snapshotStats, err := calculateSnapshotStatsForDate(agent.AgentID, agent.DeviceID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate snapshot stats: %w", err)
	}
	report.Snapshots = *snapshotStats

	return report, nil
}

// calculateBackupStatsForDate counts backups for a specific date
func calculateBackupStatsForDate(agentID string, targetDate time.Time) (*BackupStats, error) {
	stats := &BackupStats{
		FailuresByError: make(map[string]int),
	}

	// Define the date range for the target day
	startOfDay := targetDate.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Fetch backups starting from newest
	offset := 0
	limit := 50
	foundOlderBackup := false
	maxIterations := 100 // Safety limit
	consecutiveEmptyPages := 0

	for iterations := 0; iterations < maxIterations; iterations++ {
		// Get backups in descending order (newest first)
		backupsData, err := listBackups(map[string]interface{}{
			"agent_id": agentID,
			"limit":    limit,
			"offset":   offset,
			"sort_by":  "start_time",
			"sort_asc": false,
		})
		if err != nil {
			return nil, err
		}

		var backupsList struct {
			Pagination Pagination `json:"pagination"`
			Data       []Backup   `json:"data"`
		}
		if err := json.Unmarshal([]byte(backupsData), &backupsList); err != nil {
			return nil, err
		}

		// Check for empty pages
		if len(backupsList.Data) == 0 {
			consecutiveEmptyPages++
			if consecutiveEmptyPages > 2 {
				// Multiple empty pages, stop
				break
			}
		} else {
			consecutiveEmptyPages = 0
		}

		// Process each backup
		for _, backup := range backupsList.Data {
			// Parse backup start time
			backupTime, err := time.Parse(time.RFC3339, backup.StartedAt)
			if err != nil {
				continue
			}

			// Check if backup is within our target date
			if backupTime.Before(startOfDay) {
				// We've gone past our target date
				foundOlderBackup = true
				break
			}

			if backupTime.After(endOfDay) || backupTime.Equal(endOfDay) {
				// This backup is after our target date, skip it
				continue
			}

			// This backup is within our target date
			stats.Total++

			switch backup.Status {
			case "succeeded":
				stats.Successful++
			case "failed":
				stats.Failed++
				if backup.ErrorMessage != nil {
					stats.FailuresByError[*backup.ErrorMessage]++
				}
			case "running", "pending":
				stats.InProgress++
			}
		}

		// If we found a backup older than our target date, we can stop
		if foundOlderBackup {
			break
		}

		// Check if there are more backups
		if backupsList.Pagination.NextOffset == nil {
			break
		}

		// Safety check: ensure offset is progressing
		newOffset := *backupsList.Pagination.NextOffset
		if newOffset <= offset {
			// Not making progress, break to avoid infinite loop
			break
		}
		offset = newOffset

		// Safety check: reasonable limit on offset
		if offset > 10000 {
			// Reached a reasonable maximum
			break
		}
	}

	// Calculate success rate
	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Successful) / float64(stats.Total) * 100
	}

	return stats, nil
}

// calculateSnapshotStatsForDate counts snapshots created on a specific date
func calculateSnapshotStatsForDate(agentID string, deviceID string, targetDate time.Time) (*SnapshotStats, error) {
	stats := &SnapshotStats{}

	// Define the date range for the target day
	startOfDay := targetDate.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Fetch snapshots starting from newest
	offset := 0
	limit := 50
	foundOlderSnapshot := false
	maxIterations := 100 // Safety limit
	consecutiveEmptyPages := 0

	for iterations := 0; iterations < maxIterations; iterations++ {
		// Get snapshots in descending order (newest first)
		snapshotsData, err := listSnapshots(map[string]interface{}{
			"agent_id": agentID,
			"limit":    limit,
			"offset":   offset,
			"sort_by":  "created",
			"sort_asc": false,
		})
		if err != nil {
			return nil, err
		}

		var snapshotsList struct {
			Pagination Pagination `json:"pagination"`
			Data       []Snapshot `json:"data"`
		}
		if err := json.Unmarshal([]byte(snapshotsData), &snapshotsList); err != nil {
			return nil, err
		}

		// Check for empty pages
		if len(snapshotsList.Data) == 0 {
			consecutiveEmptyPages++
			if consecutiveEmptyPages > 2 {
				// Multiple empty pages, stop
				break
			}
		} else {
			consecutiveEmptyPages = 0
		}

		// Process each snapshot
		for _, snapshot := range snapshotsList.Data {
			// Parse snapshot creation time (using backup_ended_at as creation time)
			snapshotTime, err := time.Parse(time.RFC3339, snapshot.BackupEndedAt)
			if err != nil {
				continue
			}

			// Check if snapshot was created on our target date
			if snapshotTime.Before(startOfDay) {
				// We've gone past our target date
				foundOlderSnapshot = true
				break
			}

			if snapshotTime.After(endOfDay) || snapshotTime.Equal(endOfDay) {
				// This snapshot is after our target date, skip it
				continue
			}

			// This snapshot was created on our target date
			stats.Total++

			// Check if snapshot is deleted
			if snapshot.Deleted != nil && *snapshot.Deleted != "" {
				stats.Deleted++
			} else {
				stats.Active++
			}

			// Check storage locations
			hasLocal := false
			hasCloud := false
			for _, location := range snapshot.Locations {
				if location.DeviceID == deviceID {
					hasLocal = true
				} else {
					hasCloud = true
				}
			}

			if hasLocal {
				stats.LocalStorage++
			}
			if hasCloud {
				stats.CloudStorage++
			}
		}

		// If we found a snapshot older than our target date, we can stop
		if foundOlderSnapshot {
			break
		}

		// Check if there are more snapshots
		if snapshotsList.Pagination.NextOffset == nil {
			break
		}

		// Safety check: ensure offset is progressing
		newOffset := *snapshotsList.Pagination.NextOffset
		if newOffset <= offset {
			// Not making progress, break to avoid infinite loop
			break
		}
		offset = newOffset

		// Safety check: reasonable limit on offset
		if offset > 10000 {
			// Reached a reasonable maximum
			break
		}
	}

	return stats, nil
}

// formatDailyReportsAsMarkdown formats daily reports as markdown
func formatDailyReportsAsMarkdown(reports []DailyReport, date time.Time) string {
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

	for _, report := range reports {
		totalBackups += report.Backups.Total
		totalSuccessful += report.Backups.Successful
		totalSnapshots += report.Snapshots.Total
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Agents Reporting**: %d\n", len(reports)))
	sb.WriteString(fmt.Sprintf("- **Total Backups**: %d\n", totalBackups))
	if totalBackups > 0 {
		overallSuccessRate := float64(totalSuccessful) / float64(totalBackups) * 100
		sb.WriteString(fmt.Sprintf("- **Overall Success Rate**: %.1f%%\n", overallSuccessRate))
	}
	sb.WriteString(fmt.Sprintf("- **Total Snapshots Created**: %d\n\n", totalSnapshots))

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
		sb.WriteString("\n**Snapshots Created:**\n")
		sb.WriteString(fmt.Sprintf("- Total: %d\n", report.Snapshots.Total))
		sb.WriteString(fmt.Sprintf("- Active: %d\n", report.Snapshots.Active))
		sb.WriteString(fmt.Sprintf("- Deleted: %d\n", report.Snapshots.Deleted))
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
	agentSet := make(map[string]bool)

	for _, dayReports := range weekReports {
		for _, report := range dayReports {
			weeklyBackupTotal += report.Backups.Total
			weeklySuccessTotal += report.Backups.Successful
			weeklySnapshotTotal += report.Snapshots.Total
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
	sb.WriteString(fmt.Sprintf("- **Total Snapshots Created**: %d\n\n", weeklySnapshotTotal))

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
		sb.WriteString(fmt.Sprintf("- Snapshots Created: %d\n\n", daySnapshotTotal))
	}

	return sb.String()
}

// formatMonthlyReportsAsMarkdown formats monthly reports as markdown
func formatMonthlyReportsAsMarkdown(monthReports map[int][]DailyReport, firstDay time.Time) string {
	var sb strings.Builder

	year, month, _ := firstDay.Date()
	monthName := firstDay.Format("January 2006")

	sb.WriteString(fmt.Sprintf("# Monthly Backup & Snapshot Report - %s\n\n", monthName))

	// Calculate monthly totals
	monthlyBackupTotal := 0
	monthlySuccessTotal := 0
	monthlySnapshotTotal := 0
	agentSet := make(map[string]bool)

	for _, dayReports := range monthReports {
		for _, report := range dayReports {
			monthlyBackupTotal += report.Backups.Total
			monthlySuccessTotal += report.Backups.Successful
			monthlySnapshotTotal += report.Snapshots.Total
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
	sb.WriteString(fmt.Sprintf("- **Total Snapshots Created**: %d\n\n", monthlySnapshotTotal))

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
		sb.WriteString(fmt.Sprintf("- **Snapshots Created**: %d\n\n", daySnapshotTotal))
	}

	return sb.String()
}

// getReportsToolInfo returns the tool definition for the reports meta-tool
func getReportsToolInfo() ToolInfo {
	return ToolInfo{
		Name:        "slide_reports",
		Description: "Generate statistical reports about backups and snapshots. Counts items created on specific dates by iterating through data from newest to oldest.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The type of report to generate",
					"enum":        []string{"daily_backup_snapshot", "weekly_backup_snapshot", "monthly_backup_snapshot"},
				},
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
			},
			"required": []string{"operation"},
		},
	}
}
