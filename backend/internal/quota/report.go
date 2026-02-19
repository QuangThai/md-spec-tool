package quota

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DailyReport represents usage for a single day
type DailyReport struct {
	Date             time.Time `json:"date"`
	SessionID        string    `json:"session_id,omitempty"`
	UserID           string    `json:"user_id,omitempty"`
	TokensUsed       int64     `json:"tokens_used"`
	ConversionsCount int       `json:"conversions_count"`
	RequestCount     int       `json:"request_count"`
}

// UsageReportRequest filters for report generation
type UsageReportRequest struct {
	SessionID string     // If set, return only for this session
	UserID    string     // If set, return for all sessions of this user
	Days      int        // Number of past days (default 7)
	AggregateBy string    // "session" or "user"
}

// SimpleQuotaUsage represents quota usage (simplified version from handlers)
type SimpleQuotaUsage struct {
	SessionID        string
	UserID           string
	TokensUsedToday  int64
	DailyConversions int
	ResetTime        time.Time
	LastUpdated      time.Time
}

// ReportGenerator creates usage reports
type ReportGenerator struct {
	mu    sync.RWMutex
	// Historical data: map[date_string]map[sessionID]*DailySnapshot
	history map[string]map[string]*DailySnapshot
}

// DailySnapshot captures daily quota data
type DailySnapshot struct {
	SessionID        string
	UserID           string
	TokensUsedToday  int64
	DailyConversions int
	RequestCount     int
	Timestamp        time.Time
}

// NewReportGenerator creates a new report generator
func NewReportGenerator() *ReportGenerator {
	rg := &ReportGenerator{
		history: make(map[string]map[string]*DailySnapshot),
	}

	// Start daily snapshot routine
	go rg.snapshotLoop()

	return rg
}

// RecordSnapshot saves current quotas to history (called daily)
func (rg *ReportGenerator) RecordSnapshot(ctx context.Context, sessionID string, usage *SimpleQuotaUsage) error {
	if sessionID == "" {
		return fmt.Errorf("session_id required")
	}

	if usage == nil {
		return fmt.Errorf("usage required")
	}

	rg.mu.Lock()
	defer rg.mu.Unlock()

	dateStr := time.Now().UTC().Format("2006-01-02")
	if rg.history[dateStr] == nil {
		rg.history[dateStr] = make(map[string]*DailySnapshot)
	}

	rg.history[dateStr][sessionID] = &DailySnapshot{
		SessionID:        sessionID,
		UserID:           usage.UserID,
		TokensUsedToday:  usage.TokensUsedToday,
		DailyConversions: usage.DailyConversions,
		RequestCount:     usage.DailyConversions, // Approximate
		Timestamp:        time.Now().UTC(),
	}

	return nil
}

// GetDailyReport returns usage report for specified days
// aggregateBy: "session" or "user"
func (rg *ReportGenerator) GetDailyReport(ctx context.Context, req *UsageReportRequest) ([]*DailyReport, error) {
	if req == nil {
		req = &UsageReportRequest{
			Days:        7,
			AggregateBy: "session",
		}
	}

	if req.Days == 0 {
		req.Days = 7
	}

	if req.AggregateBy == "" {
		req.AggregateBy = "session"
	}

	rg.mu.RLock()
	defer rg.mu.RUnlock()

	var reports []*DailyReport
	now := time.Now().UTC()

	// Iterate last N days
	for i := 0; i < req.Days; i++ {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		dayData := rg.history[dateStr]
		if dayData == nil {
			continue
		}

		// Aggregate by session or user
		if req.AggregateBy == "user" {
			userMap := make(map[string]*DailyReport)

			for _, snapshot := range dayData {
				// Skip if filtering by session
				if req.SessionID != "" && snapshot.SessionID != req.SessionID {
					continue
				}

				// Skip if filtering by user
				if req.UserID != "" && snapshot.UserID != req.UserID {
					continue
				}

				userID := snapshot.UserID
				if userID == "" {
					userID = "anonymous"
				}

				if _, exists := userMap[userID]; !exists {
					userMap[userID] = &DailyReport{
						Date:   date,
						UserID: userID,
					}
				}

				userMap[userID].TokensUsed += snapshot.TokensUsedToday
				userMap[userID].ConversionsCount += snapshot.DailyConversions
				userMap[userID].RequestCount += snapshot.RequestCount
			}

			for _, report := range userMap {
				reports = append(reports, report)
			}
		} else {
			// Aggregate by session (default)
			for sessionID, snapshot := range dayData {
				// Skip if filtering by session
				if req.SessionID != "" && sessionID != req.SessionID {
					continue
				}

				// Skip if filtering by user
				if req.UserID != "" && snapshot.UserID != req.UserID {
					continue
				}

				reports = append(reports, &DailyReport{
					Date:             date,
					SessionID:        sessionID,
					UserID:           snapshot.UserID,
					TokensUsed:       snapshot.TokensUsedToday,
					ConversionsCount: snapshot.DailyConversions,
					RequestCount:     snapshot.RequestCount,
				})
			}
		}
	}

	return reports, nil
}

// snapshotLoop captures daily snapshots at midnight UTC
func (rg *ReportGenerator) snapshotLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// Check if it's time to snapshot (simplified: every hour, check if past midnight)
		now := time.Now().UTC()
		if now.Hour() == 0 && now.Minute() < 5 { // Within first 5 minutes of day
			// Snapshot would be called here with actual quota data
			// In production, iterate all sessions and call RecordSnapshot
		}
	}
}

// ClearHistory removes old snapshots (for testing/cleanup)
func (rg *ReportGenerator) ClearHistory() {
	rg.mu.Lock()
	defer rg.mu.Unlock()

	rg.history = make(map[string]map[string]*DailySnapshot)
}

// GetHistorySize returns number of stored snapshots (for testing)
func (rg *ReportGenerator) GetHistorySize() int {
	rg.mu.RLock()
	defer rg.mu.RUnlock()

	count := 0
	for _, dayData := range rg.history {
		count += len(dayData)
	}

	return count
}
