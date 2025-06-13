package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type LokiClient struct {
	BaseURL string
	Client  *http.Client
}

type LokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type LokiLog struct {
	Timestamp time.Time         `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels"`
}

func NewLokiClient(baseURL string) *LokiClient {
	return &LokiClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (lc *LokiClient) QueryRange(query string, start, end time.Time, projectID string) ([]LokiLog, error) {
	// Build LogQL query with project filter
	fullQuery := fmt.Sprintf(`{project_id="%s"} |= "%s"`, projectID, query)

	params := url.Values{}
	params.Set("query", fullQuery)
	params.Set("start", strconv.FormatInt(start.UnixNano(), 10))
	params.Set("end", strconv.FormatInt(end.UnixNano(), 10))
	params.Set("limit", "1000")

	queryURL := fmt.Sprintf("%s/loki/api/v1/query_range?%s", lc.BaseURL, params.Encode())

	resp, err := lc.Client.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query Loki: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Loki query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var lokiResp LokiQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&lokiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Loki response: %v", err)
	}

	var logs []LokiLog
	for _, result := range lokiResp.Data.Result {
		for _, value := range result.Values {
			if len(value) < 2 {
				continue
			}

			timestamp, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				continue
			}

			logs = append(logs, LokiLog{
				Timestamp: time.Unix(0, timestamp),
				Line:      value[1],
				Labels:    result.Stream,
			})
		}
	}

	return logs, nil
}

func (lc *LokiClient) QueryLogsAroundTime(projectID string, alertTime time.Time, windowMinutes int) ([]LokiLog, error) {
	start := alertTime.Add(-time.Duration(windowMinutes) * time.Minute)
	end := alertTime.Add(time.Duration(windowMinutes) * time.Minute)

	// Query for all logs in the time window
	query := "" // Empty query to get all logs
	return lc.QueryRange(query, start, end, projectID)
}

func (lc *LokiClient) QueryLogsByIP(projectID string, ipAddress string, start, end time.Time) ([]LokiLog, error) {
	query := ipAddress
	return lc.QueryRange(query, start, end, projectID)
}

func (lc *LokiClient) QueryLogsByUser(projectID string, userIdentifier string, start, end time.Time) ([]LokiLog, error) {
	query := userIdentifier
	return lc.QueryRange(query, start, end, projectID)
}
