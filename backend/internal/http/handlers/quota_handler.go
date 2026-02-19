package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
	"github.com/yourorg/md-spec-tool/internal/quota"
)

type QuotaHandler struct {
	store  QuotaStore
	report *quota.ReportGenerator // For daily usage reports
}

func NewQuotaHandler(store QuotaStore) *QuotaHandler {
	return &QuotaHandler{
		store:  store,
		report: quota.NewReportGenerator(),
	}
}

type QuotaUsageResponse struct {
	SessionID        string    `json:"session_id"`
	UsedTokens       int64     `json:"used_tokens"`
	LimitTokens      int64     `json:"limit_tokens"`
	RemainingTokens  int64     `json:"remaining_tokens"`
	ResetAt          time.Time `json:"reset_at"`
	Status           string    `json:"status"` // "ok" or "exceeded"
	DailyConversions int       `json:"daily_conversions"`
}

// DailyReportRequest filters for report generation
type DailyReportRequest struct {
	SessionID   string `json:"session_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Days        int    `json:"days,omitempty"`
	AggregateBy string `json:"aggregate_by,omitempty"` // "session" or "user"
}

// DailyReportResponse wraps daily usage reports
type DailyReportResponse struct {
	Reports []*quota.DailyReport `json:"reports"`
	Period  string               `json:"period"`
	Count   int                  `json:"count"`
}

// GetQuotaStatus returns current usage for a session
// GET /api/quota/status
func (h *QuotaHandler) GetQuotaStatus(c *gin.Context) {
	sessionID := c.GetString("session_id")
	slog.Info("GetQuotaStatus called", "session_id", sessionID, "header_session_id", c.GetHeader("X-Session-ID"))
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "session_id not found in context",
			Code:  "MISSING_SESSION_ID",
		})
		return
	}

	usage, err := h.store.GetUsage(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "failed to get quota",
			Code:  "QUOTA_FETCH_ERROR",
		})
		return
	}

	const DailyTokenLimit = 100000
	remaining := DailyTokenLimit - usage.TokensUsedToday
	if remaining < 0 {
		remaining = 0
	}

	status := "ok"
	if remaining == 0 {
		status = "exceeded"
	}

	c.JSON(http.StatusOK, QuotaUsageResponse{
		SessionID:       sessionID,
		UsedTokens:      usage.TokensUsedToday,
		LimitTokens:     DailyTokenLimit,
		RemainingTokens: remaining,
		ResetAt:         usage.ResetTime,
		Status:          status,
		DailyConversions: usage.DailyConversions,
	})
}

// AddQuotaUsage is called by AI service after completion
// Internal method, not exposed as HTTP endpoint
func (h *QuotaHandler) AddQuotaUsage(ctx context.Context, sessionID string, tokens int64) error {
	if sessionID == "" {
		return errors.New("session_id required")
	}

	if err := h.store.AddUsage(ctx, sessionID, tokens); err != nil {
		return err
	}

	// Emit telemetry event for quota tracking
	middleware.RecordTelemetryEvent(middleware.TelemetryEvent{
		EventName:   "quota_used",
		EventTime:   time.Now().UTC(),
		SessionID:   sessionID,
		Status:      "success",
		DurationMS:  0,
		ConfidenceScore: float64(tokens) / 100000, // Simple ratio
		Source:      "backend",
	})

	return nil
}

// IncrementConversion increments the daily conversion count for a session
func (h *QuotaHandler) IncrementConversion(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session_id required")
	}

	return h.store.IncrementConversion(ctx, sessionID)
}

// RecordConversion atomically increments conversion count and adds token usage
func (h *QuotaHandler) RecordConversion(ctx context.Context, sessionID string, tokens int64) error {
	if sessionID == "" {
		return errors.New("session_id required")
	}

	if err := h.store.RecordConversion(ctx, sessionID, tokens); err != nil {
		return err
	}

	if tokens > 0 {
		middleware.RecordTelemetryEvent(middleware.TelemetryEvent{
			EventName:       "quota_used",
			EventTime:       time.Now().UTC(),
			SessionID:       sessionID,
			Status:          "success",
			DurationMS:      0,
			ConfidenceScore: float64(tokens) / 100000,
			Source:          "backend",
		})
	}

	return nil
}

// ValidateQuotaAvailable checks if session can make another request
func (h *QuotaHandler) ValidateQuotaAvailable(ctx context.Context, sessionID string) (bool, error) {
	return h.store.IsAvailable(ctx, sessionID)
}

// GetUsageDetails returns quota usage for display/headers
func (h *QuotaHandler) GetUsageDetails(ctx context.Context, sessionID string) (*middleware.QuotaUsage, error) {
	usage, err := h.store.GetUsage(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return &middleware.QuotaUsage{
		TokensUsedToday:  usage.TokensUsedToday,
		DailyConversions: usage.DailyConversions,
	}, nil
}

// GetDailyReport returns daily usage report
// GET /api/quota/daily-report
func (h *QuotaHandler) GetDailyReport(c *gin.Context) {
	var req DailyReportRequest

	// Parse query params
	req.SessionID = c.Query("session_id")
	req.UserID = c.Query("user_id")
	req.AggregateBy = c.DefaultQuery("aggregate_by", "session")

	// Parse days
	daysStr := c.DefaultQuery("days", "7")
	fmt.Sscanf(daysStr, "%d", &req.Days)
	if req.Days == 0 {
		req.Days = 7
	}

	// Convert to quota package request
	quotaReq := &quota.UsageReportRequest{
		SessionID:   req.SessionID,
		UserID:      req.UserID,
		Days:        req.Days,
		AggregateBy: req.AggregateBy,
	}

	reports, err := h.report.GetDailyReport(c.Request.Context(), quotaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "failed to generate report",
			Code:  "REPORT_GENERATION_ERROR",
		})
		return
	}

	period := fmt.Sprintf("last_%d_days", req.Days)
	if req.SessionID != "" {
		period += "_session_" + req.SessionID
	}

	c.JSON(http.StatusOK, DailyReportResponse{
		Reports: reports,
		Period:  period,
		Count:   len(reports),
	})
}
