package main

import (
	"database/sql"
	"fmt"
	"time"
)

type CorrelationEngine struct {
	db *sql.DB
}

type UserCorrelation struct {
	UserIdentifier  string    `json:"user_identifier"`
	IPAddress       string    `json:"ip_address"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	ConfidenceScore float64   `json:"confidence_score"`
	SourceSystems   []string  `json:"source_systems"`
	CorrelationType string    `json:"correlation_type"`
}

type CorrelationResult struct {
	PrimaryLog       *NormalizedLog    `json:"primary_log"`
	RelatedLogs      []NormalizedLog   `json:"related_logs"`
	UserCorrelations []UserCorrelation `json:"user_correlations"`
	TimeWindow       TimeWindow        `json:"time_window"`
	CorrelationScore float64           `json:"correlation_score"`
}

type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func NewCorrelationEngine(db *sql.DB) *CorrelationEngine {
	return &CorrelationEngine{db: db}
}

func (ce *CorrelationEngine) CorrelateLogsForAlert(alert Alert, logs []NormalizedLog) (*CorrelationResult, error) {
	result := &CorrelationResult{
		TimeWindow: TimeWindow{
			Start: alert.Timestamp.Add(-15 * time.Minute),
			End:   alert.Timestamp.Add(15 * time.Minute),
		},
		RelatedLogs: logs,
	}

	// Build user-to-IP correlations from the logs
	userCorrelations := ce.buildUserIPCorrelations(logs)

	// Store correlations in database for future use
	for _, correlation := range userCorrelations {
		ce.storeUserCorrelation(correlation)
	}

	// Find existing correlations from database
	existingCorrelations, err := ce.getExistingCorrelations(logs)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing correlations: %v", err)
	}

	// Merge and deduplicate correlations
	allCorrelations := ce.mergeCorrelations(userCorrelations, existingCorrelations)
	result.UserCorrelations = allCorrelations

	// Calculate correlation score
	result.CorrelationScore = ce.calculateCorrelationScore(logs, allCorrelations)

	return result, nil
}

func (ce *CorrelationEngine) buildUserIPCorrelations(logs []NormalizedLog) []UserCorrelation {
	var correlations []UserCorrelation

	// Group logs by time proximity (within 5 minutes)
	timeGroups := ce.groupLogsByTime(logs, 5*time.Minute)

	for _, group := range timeGroups {
		// Find logs with emails and logs with IPs in the same time group
		emailLogs := ce.filterLogsByEmails(group)
		ipLogs := ce.filterLogsByIPs(group)

		// Create correlations between users and IPs in the same time window
		for _, emailLog := range emailLogs {
			for _, email := range emailLog.UserEmails {
				for _, ipLog := range ipLogs {
					for _, ip := range ipLog.IPAddresses {
						correlation := UserCorrelation{
							UserIdentifier:  email,
							IPAddress:       ip,
							FirstSeen:       emailLog.Timestamp,
							LastSeen:        ipLog.Timestamp,
							ConfidenceScore: ce.calculateConfidenceScore(emailLog, ipLog),
							SourceSystems:   []string{emailLog.Source, ipLog.Source},
							CorrelationType: "time_proximity",
						}
						correlations = append(correlations, correlation)
					}
				}
			}
		}
	}

	// Also look for direct correlations (same log contains both email and IP)
	for _, log := range logs {
		if len(log.UserEmails) > 0 && len(log.IPAddresses) > 0 {
			for _, email := range log.UserEmails {
				for _, ip := range log.IPAddresses {
					correlation := UserCorrelation{
						UserIdentifier:  email,
						IPAddress:       ip,
						FirstSeen:       log.Timestamp,
						LastSeen:        log.Timestamp,
						ConfidenceScore: 0.9, // High confidence for direct correlation
						SourceSystems:   []string{log.Source},
						CorrelationType: "direct",
					}
					correlations = append(correlations, correlation)
				}
			}
		}
	}

	return ce.deduplicateCorrelations(correlations)
}

func (ce *CorrelationEngine) groupLogsByTime(logs []NormalizedLog, window time.Duration) [][]NormalizedLog {
	if len(logs) == 0 {
		return nil
	}

	var groups [][]NormalizedLog
	var currentGroup []NormalizedLog

	// Sort logs by timestamp first
	sortedLogs := make([]NormalizedLog, len(logs))
	copy(sortedLogs, logs)

	// Simple bubble sort for timestamp
	for i := 0; i < len(sortedLogs)-1; i++ {
		for j := 0; j < len(sortedLogs)-i-1; j++ {
			if sortedLogs[j].Timestamp.After(sortedLogs[j+1].Timestamp) {
				sortedLogs[j], sortedLogs[j+1] = sortedLogs[j+1], sortedLogs[j]
			}
		}
	}

	currentGroup = append(currentGroup, sortedLogs[0])
	groupStart := sortedLogs[0].Timestamp

	for i := 1; i < len(sortedLogs); i++ {
		if sortedLogs[i].Timestamp.Sub(groupStart) <= window {
			currentGroup = append(currentGroup, sortedLogs[i])
		} else {
			groups = append(groups, currentGroup)
			currentGroup = []NormalizedLog{sortedLogs[i]}
			groupStart = sortedLogs[i].Timestamp
		}
	}

	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

func (ce *CorrelationEngine) filterLogsByEmails(logs []NormalizedLog) []NormalizedLog {
	var filtered []NormalizedLog
	for _, log := range logs {
		if len(log.UserEmails) > 0 {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

func (ce *CorrelationEngine) filterLogsByIPs(logs []NormalizedLog) []NormalizedLog {
	var filtered []NormalizedLog
	for _, log := range logs {
		if len(log.IPAddresses) > 0 {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

func (ce *CorrelationEngine) calculateConfidenceScore(emailLog, ipLog NormalizedLog) float64 {
	score := 0.5 // Base score

	// Time proximity increases confidence
	timeDiff := emailLog.Timestamp.Sub(ipLog.Timestamp)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	if timeDiff <= 1*time.Minute {
		score += 0.3
	} else if timeDiff <= 5*time.Minute {
		score += 0.2
	} else if timeDiff <= 15*time.Minute {
		score += 0.1
	}

	// Same company/host increases confidence
	if emailLog.CompanyCode != "" && emailLog.CompanyCode == ipLog.CompanyCode {
		score += 0.2
	}

	if emailLog.Host != "" && emailLog.Host == ipLog.Host {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (ce *CorrelationEngine) storeUserCorrelation(correlation UserCorrelation) error {
	query := `
		INSERT INTO user_correlations (user_identifier, ip_address, first_seen, last_seen, confidence_score, source_systems)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_identifier, ip_address) 
		DO UPDATE SET 
			last_seen = GREATEST(user_correlations.last_seen, $4),
			confidence_score = GREATEST(user_correlations.confidence_score, $5),
			source_systems = array(SELECT DISTINCT unnest(user_correlations.source_systems || $6))
	`

	_, err := ce.db.Exec(query,
		correlation.UserIdentifier,
		correlation.IPAddress,
		correlation.FirstSeen,
		correlation.LastSeen,
		correlation.ConfidenceScore,
		correlation.SourceSystems)

	return err
}

func (ce *CorrelationEngine) getExistingCorrelations(logs []NormalizedLog) ([]UserCorrelation, error) {
	var correlations []UserCorrelation

	// Collect all unique IPs and emails from logs
	ips := make(map[string]bool)
	emails := make(map[string]bool)

	for _, log := range logs {
		for _, ip := range log.IPAddresses {
			ips[ip] = true
		}
		for _, email := range log.UserEmails {
			emails[email] = true
		}
	}

	// Query for existing correlations
	if len(ips) > 0 || len(emails) > 0 {
		query := `
			SELECT user_identifier, ip_address, first_seen, last_seen, confidence_score, source_systems
			FROM user_correlations 
			WHERE user_identifier = ANY($1) OR ip_address = ANY($2)
		`

		emailList := make([]string, 0, len(emails))
		for email := range emails {
			emailList = append(emailList, email)
		}

		ipList := make([]string, 0, len(ips))
		for ip := range ips {
			ipList = append(ipList, ip)
		}

		rows, err := ce.db.Query(query, emailList, ipList)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var correlation UserCorrelation
			var sourceSystems []string

			err := rows.Scan(
				&correlation.UserIdentifier,
				&correlation.IPAddress,
				&correlation.FirstSeen,
				&correlation.LastSeen,
				&correlation.ConfidenceScore,
				&sourceSystems,
			)
			if err != nil {
				continue
			}

			correlation.SourceSystems = sourceSystems
			correlation.CorrelationType = "historical"
			correlations = append(correlations, correlation)
		}
	}

	return correlations, nil
}

func (ce *CorrelationEngine) mergeCorrelations(new, existing []UserCorrelation) []UserCorrelation {
	correlationMap := make(map[string]UserCorrelation)

	// Add existing correlations
	for _, correlation := range existing {
		key := correlation.UserIdentifier + "|" + correlation.IPAddress
		correlationMap[key] = correlation
	}

	// Add or update with new correlations
	for _, correlation := range new {
		key := correlation.UserIdentifier + "|" + correlation.IPAddress
		if existing, exists := correlationMap[key]; exists {
			// Merge: take higher confidence score and combine source systems
			if correlation.ConfidenceScore > existing.ConfidenceScore {
				existing.ConfidenceScore = correlation.ConfidenceScore
			}
			existing.SourceSystems = ce.mergeSources(existing.SourceSystems, correlation.SourceSystems)
			correlationMap[key] = existing
		} else {
			correlationMap[key] = correlation
		}
	}

	// Convert back to slice
	var result []UserCorrelation
	for _, correlation := range correlationMap {
		result = append(result, correlation)
	}

	return result
}

func (ce *CorrelationEngine) mergeSources(sources1, sources2 []string) []string {
	sourceMap := make(map[string]bool)
	for _, source := range sources1 {
		sourceMap[source] = true
	}
	for _, source := range sources2 {
		sourceMap[source] = true
	}

	var result []string
	for source := range sourceMap {
		result = append(result, source)
	}
	return result
}

func (ce *CorrelationEngine) deduplicateCorrelations(correlations []UserCorrelation) []UserCorrelation {
	correlationMap := make(map[string]UserCorrelation)

	for _, correlation := range correlations {
		key := correlation.UserIdentifier + "|" + correlation.IPAddress
		if existing, exists := correlationMap[key]; exists {
			// Keep the one with higher confidence
			if correlation.ConfidenceScore > existing.ConfidenceScore {
				correlationMap[key] = correlation
			}
		} else {
			correlationMap[key] = correlation
		}
	}

	var result []UserCorrelation
	for _, correlation := range correlationMap {
		result = append(result, correlation)
	}

	return result
}

func (ce *CorrelationEngine) calculateCorrelationScore(logs []NormalizedLog, correlations []UserCorrelation) float64 {
	if len(logs) == 0 {
		return 0.0
	}

	score := 0.0

	// Base score from number of correlations found
	score += float64(len(correlations)) * 0.1

	// Bonus for high-confidence correlations
	for _, correlation := range correlations {
		if correlation.ConfidenceScore > 0.8 {
			score += 0.2
		} else if correlation.ConfidenceScore > 0.6 {
			score += 0.1
		}
	}

	// Bonus for multiple source systems involved
	sourceMap := make(map[string]bool)
	for _, log := range logs {
		sourceMap[log.Source] = true
	}
	if len(sourceMap) > 2 {
		score += 0.3
	} else if len(sourceMap) > 1 {
		score += 0.2
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}
