package main

import (
	"encoding/json"
	"net"
	"regexp"
	"strings"
	"time"
)

type LogNormalizer struct {
	emailRegex *regexp.Regexp
	ipRegex    *regexp.Regexp
}

type NormalizedLog struct {
	OriginalLog string                 `json:"original_log"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	IPAddresses []string               `json:"ip_addresses"`
	UserEmails  []string               `json:"user_emails"`
	UserNames   []string               `json:"user_names"`
	Action      string                 `json:"action"`
	Severity    string                 `json:"severity"`
	CompanyCode string                 `json:"company_code"`
	Host        string                 `json:"host"`
	URI         string                 `json:"uri"`
	Method      string                 `json:"method"`
	StatusCode  string                 `json:"status_code"`
	Country     string                 `json:"country"`
	RawData     map[string]interface{} `json:"raw_data"`
}

func NewLogNormalizer() *LogNormalizer {
	return &LogNormalizer{
		emailRegex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		ipRegex:    regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`),
	}
}

func (ln *LogNormalizer) NormalizeLog(lokiLog LokiLog) (*NormalizedLog, error) {
	normalized := &NormalizedLog{
		OriginalLog: lokiLog.Line,
		Timestamp:   lokiLog.Timestamp,
		RawData:     make(map[string]interface{}),
	}

	// Try to parse as JSON first
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(lokiLog.Line), &logData); err == nil {
		normalized.RawData = logData
		ln.extractFromJSON(normalized, logData)
	} else {
		// Fallback to text parsing
		ln.extractFromText(normalized, lokiLog.Line)
	}

	// Extract IPs and emails from the entire log line
	normalized.IPAddresses = ln.extractIPs(lokiLog.Line)
	normalized.UserEmails = ln.extractEmails(lokiLog.Line)

	// Determine source based on log content
	normalized.Source = ln.determineSource(lokiLog.Line, logData)

	return normalized, nil
}

func (ln *LogNormalizer) extractFromJSON(normalized *NormalizedLog, data map[string]interface{}) {
	// Common field mappings across different log sources
	fieldMappings := map[string][]string{
		"company_code": {"company_code", "companyCode"},
		"timestamp":    {"timestamp", "time", "reqTimeSec", "Event_date"},
		"action":       {"action", "terminatingRuleType", "operationName"},
		"severity":     {"severity", "Importance"},
		"host":         {"host", "reqHost", "Host", "Company_host"},
		"uri":          {"uri", "requestUri", "reqPath"},
		"method":       {"httpMethod", "reqMethod"},
		"status_code":  {"statusCode", "status"},
		"country":      {"country", "client_country_name", "client_country_code"},
	}

	for field, keys := range fieldMappings {
		for _, key := range keys {
			if value, exists := data[key]; exists && value != nil {
				switch field {
				case "company_code":
					normalized.CompanyCode = toString(value)
				case "action":
					normalized.Action = toString(value)
				case "severity":
					normalized.Severity = toString(value)
				case "host":
					normalized.Host = toString(value)
				case "uri":
					normalized.URI = toString(value)
				case "method":
					normalized.Method = toString(value)
				case "status_code":
					normalized.StatusCode = toString(value)
				case "country":
					normalized.Country = toString(value)
				}
				break
			}
		}
	}

	// Extract specific IP fields
	ipFields := []string{"clientIP", "cliIP", "client_ip", "clientIp"}
	for _, field := range ipFields {
		if value, exists := data[field]; exists {
			if ip := toString(value); ln.isValidIP(ip) {
				normalized.IPAddresses = append(normalized.IPAddresses, ip)
			}
		}
	}

	// Extract forwarded IPs
	forwardedFields := []string{"xForwardedFor", "x-forwarded-for"}
	for _, field := range forwardedFields {
		if value, exists := data[field]; exists {
			ips := strings.Split(toString(value), ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if ln.isValidIP(ip) {
					normalized.IPAddresses = append(normalized.IPAddresses, ip)
				}
			}
		}
	}
}

func (ln *LogNormalizer) extractFromText(normalized *NormalizedLog, logLine string) {
	// Extract common patterns from text logs
	if strings.Contains(logLine, "Deep Security") {
		normalized.Source = "deep_security"
		// Extract host from Deep Security logs
		if strings.Contains(logLine, "Host:") {
			parts := strings.Split(logLine, "Host:")
			if len(parts) > 1 {
				hostPart := strings.TrimSpace(parts[1])
				hostEnd := strings.Index(hostPart, ",")
				if hostEnd > 0 {
					normalized.Host = hostPart[:hostEnd]
				}
			}
		}
	}
}

func (ln *LogNormalizer) determineSource(logLine string, data map[string]interface{}) string {
	if data != nil {
		// Check for specific source indicators in JSON
		if _, exists := data["webaclId"]; exists {
			return "aws_waf"
		}
		if _, exists := data["operationName"]; exists && strings.Contains(toString(data["operationName"]), "Microsoft.Cdn") {
			return "azure_waf"
		}
		if _, exists := data["streamId"]; exists {
			return "akamai_waf"
		}
		if _, exists := data["Rule_name"]; exists {
			return "deep_security"
		}
		if _, exists := data["type"]; exists && strings.Contains(toString(data["type"]), "guardduty") {
			return "aws_guardduty"
		}
	}

	// Fallback to text-based detection
	if strings.Contains(logLine, "Deep Security") {
		return "deep_security"
	}
	if strings.Contains(logLine, "GuardDuty") {
		return "aws_guardduty"
	}
	if strings.Contains(logLine, "Akamai") {
		return "akamai_waf"
	}

	return "unknown"
}

func (ln *LogNormalizer) extractIPs(text string) []string {
	matches := ln.ipRegex.FindAllString(text, -1)
	var validIPs []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if ln.isValidIP(match) && !seen[match] {
			validIPs = append(validIPs, match)
			seen[match] = true
		}
	}

	return validIPs
}

func (ln *LogNormalizer) extractEmails(text string) []string {
	matches := ln.emailRegex.FindAllString(text, -1)
	var validEmails []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if !seen[match] {
			validEmails = append(validEmails, strings.ToLower(match))
			seen[match] = true
		}
	}

	return validEmails
}

func (ln *LogNormalizer) isValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && !parsed.IsLoopback() && !parsed.IsUnspecified()
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	if num, ok := value.(float64); ok {
		return string(rune(int(num)))
	}
	return ""
}
