package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/hibiken/asynq"
)

// Alert analysis task handler
func (app *App) handleAlertAnalysis(ctx context.Context, t *asynq.Task) error {
	var alert Alert
	if err := json.Unmarshal(t.Payload(), &alert); err != nil {
		return fmt.Errorf("failed to unmarshal alert: %v", err)
	}

	log.Printf("Processing alert analysis for alert ID: %s", alert.ID)
	startTime := time.Now()

	// Query logs around the alert time (±15 minutes)
	lokiLogs, err := app.LokiClient.QueryLogsAroundTime(alert.ProjectID, alert.Timestamp, 15)
	if err != nil {
		log.Printf("Failed to query Loki logs: %v", err)
		// Continue with empty logs for demo purposes
		lokiLogs = []LokiLog{}
	}

	// Normalize logs
	var normalizedLogs []NormalizedLog
	for _, lokiLog := range lokiLogs {
		normalized, err := app.Normalizer.NormalizeLog(lokiLog)
		if err != nil {
			log.Printf("Failed to normalize log: %v", err)
			continue
		}
		normalizedLogs = append(normalizedLogs, *normalized)
	}

	// Perform correlation analysis
	correlationResult, err := app.Correlator.CorrelateLogsForAlert(alert, normalizedLogs)
	if err != nil {
		return fmt.Errorf("failed to correlate logs: %v", err)
	}

	// Build enrichment data
	enrichmentData := app.buildEnrichmentData(alert, correlationResult)

	// Create analysis result
	analysisResult := AnalysisResult{
		AlertID:           alert.ID,
		ProjectID:         alert.ProjectID,
		CorrelatedLogs:    normalizedLogs,
		UserCorrelations:  correlationResult.UserCorrelations,
		EnrichmentData:    enrichmentData,
		AnalysisTimestamp: time.Now(),
		ProcessingTimeMs:  time.Since(startTime).Milliseconds(),
	}

	// Store analysis result
	if err := app.storeAnalysisResult(analysisResult); err != nil {
		return fmt.Errorf("failed to store analysis result: %v", err)
	}

	log.Printf("Completed analysis for alert %s in %dms", alert.ID, analysisResult.ProcessingTimeMs)
	return nil
}

func (app *App) buildEnrichmentData(alert Alert, correlationResult *CorrelationResult) map[string]interface{} {
	enrichment := make(map[string]interface{})

	// Basic alert information
	enrichment["alert_source"] = alert.Source
	enrichment["alert_severity"] = alert.Severity
	enrichment["analysis_window"] = map[string]interface{}{
		"start": correlationResult.TimeWindow.Start,
		"end":   correlationResult.TimeWindow.End,
	}

	// Correlation statistics
	enrichment["correlation_stats"] = map[string]interface{}{
		"total_logs_analyzed":     len(correlationResult.RelatedLogs),
		"user_correlations_found": len(correlationResult.UserCorrelations),
		"correlation_score":       correlationResult.CorrelationScore,
	}

	// Source system breakdown
	sourceCounts := make(map[string]int)
	for _, log := range correlationResult.RelatedLogs {
		sourceCounts[log.Source]++
	}
	enrichment["source_breakdown"] = sourceCounts

	// High-confidence correlations
	var highConfidenceCorrelations []UserCorrelation
	for _, correlation := range correlationResult.UserCorrelations {
		if correlation.ConfidenceScore > 0.7 {
			highConfidenceCorrelations = append(highConfidenceCorrelations, correlation)
		}
	}
	enrichment["high_confidence_correlations"] = highConfidenceCorrelations

	// Unique users and IPs involved
	uniqueUsers := make(map[string]bool)
	uniqueIPs := make(map[string]bool)
	for _, log := range correlationResult.RelatedLogs {
		for _, email := range log.UserEmails {
			uniqueUsers[email] = true
		}
		for _, ip := range log.IPAddresses {
			uniqueIPs[ip] = true
		}
	}

	userList := make([]string, 0, len(uniqueUsers))
	for user := range uniqueUsers {
		userList = append(userList, user)
	}

	ipList := make([]string, 0, len(uniqueIPs))
	for ip := range uniqueIPs {
		ipList = append(ipList, ip)
	}

	enrichment["involved_users"] = userList
	enrichment["involved_ips"] = ipList

	return enrichment
}

func (app *App) storeAnalysisResult(result AnalysisResult) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis result: %v", err)
	}

	query := `
		INSERT INTO analysis_results (alert_id, project_id, result_data)
		VALUES ($1, $2, $3)
		ON CONFLICT (alert_id) 
		DO UPDATE SET result_data = $3, created_at = NOW()
	`

	_, err = app.DB.Exec(query, result.AlertID, result.ProjectID, resultJSON)
	return err
}

func (app *App) getStoredAnalysisResult(alertID string) (*AnalysisResult, error) {
	var resultJSON []byte
	query := `SELECT result_data FROM analysis_results WHERE alert_id = $1`

	err := app.DB.QueryRow(query, alertID).Scan(&resultJSON)
	if err != nil {
		return nil, err
	}

	var result AnalysisResult
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis result: %v", err)
	}

	return &result, nil
}

// Mock data generator for testing
func (app *App) startMockDataGenerator() {
	log.Println("Starting mock data generator...")

	// Generate mock alerts every 30 seconds for demo
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			app.generateMockAlert()
		}
	}
}

func (app *App) generateMockAlert() {
	alerts := []Alert{
		{
			Source:    "aws_waf",
			Severity:  "high",
			Message:   "Suspicious SQL injection attempt detected",
			ProjectID: "demo-project-1",
			RawData: map[string]interface{}{
				"clientIP": "192.168.1.100",
				"uri":      "/api/users",
				"method":   "POST",
			},
		},
		{
			Source:    "deep_security",
			Severity:  "medium",
			Message:   "File modification detected in system directory",
			ProjectID: "demo-project-1",
			RawData: map[string]interface{}{
				"host":     "web-server-01",
				"filePath": "/etc/passwd",
			},
		},
		{
			Source:    "azure_waf",
			Severity:  "high",
			Message:   "Cross-site scripting attempt blocked",
			ProjectID: "demo-project-1",
			RawData: map[string]interface{}{
				"clientIP":  "10.0.0.50",
				"uri":       "/dashboard",
				"userAgent": "Mozilla/5.0",
			},
		},
	}

	// Randomly select an alert
	alert := alerts[rand.Intn(len(alerts))]
	alert.ID = generateID()
	alert.Timestamp = time.Now()

	// Queue the alert for analysis
	task := asynq.NewTask("alert:analyze", mustMarshal(alert))
	if _, err := app.TaskClient.Enqueue(task); err != nil {
		log.Printf("Failed to queue mock alert: %v", err)
	} else {
		log.Printf("Generated mock alert: %s (Source: %s, Severity: %s)", alert.ID, alert.Source, alert.Severity)
	}
}

// Mock Loki data generator - this would normally be handled by your log ingestion pipeline
func (app *App) generateMockLokiData() {
	// This function would typically push mock log data to Loki
	// For the PoC, we'll simulate this by having the Loki client return mock data
	log.Println("Mock Loki data generation would happen here in a real implementation")
}
