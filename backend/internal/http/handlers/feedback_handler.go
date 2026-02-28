package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/feedback"
)

// FeedbackHandler handles feedback submission and statistics endpoints.
type FeedbackHandler struct {
	store feedback.StoreInterface
}

// NewFeedbackHandler creates a FeedbackHandler backed by the given store.
func NewFeedbackHandler(store feedback.StoreInterface) *FeedbackHandler {
	return &FeedbackHandler{store: store}
}

// SubmitFeedbackRequest is the request body for POST /api/v1/mdflow/feedback.
type SubmitFeedbackRequest struct {
	RequestHash string `json:"request_hash" binding:"required"`
	Rating      int    `json:"rating"       binding:"required"`
	Corrections string `json:"corrections"`
	ColumnFixes string `json:"column_fixes"`
	SessionID   string `json:"session_id"`
}

// SubmitFeedbackResponse is the response body for a successful feedback submission.
type SubmitFeedbackResponse struct {
	ID          int64  `json:"id"`
	RequestHash string `json:"request_hash"`
	Rating      int    `json:"rating"`
}

// SubmitFeedback handles POST /api/v1/mdflow/feedback.
// Body: { "request_hash": "...", "rating": 5, "corrections": "...", ... }
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	const maxBodyBytes = 64 * 1024 // 64 KB is more than enough for feedback
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var req SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "payload too large"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "request_hash and rating are required"})
		return
	}

	req.RequestHash = strings.TrimSpace(req.RequestHash)
	if req.RequestHash == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "request_hash must not be empty"})
		return
	}

	f := &feedback.Feedback{
		RequestHash: req.RequestHash,
		Rating:      req.Rating,
		Corrections: req.Corrections,
		ColumnFixes: req.ColumnFixes,
		SessionID:   strings.TrimSpace(req.SessionID),
	}

	if err := h.store.Submit(f); err != nil {
		switch err {
		case feedback.ErrInvalidRating:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "rating must be 1 (thumbs down) or 5 (thumbs up)",
			})
		case feedback.ErrEmptyRequestHash:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "request_hash must not be empty"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to submit feedback"})
		}
		return
	}

	c.JSON(http.StatusCreated, SubmitFeedbackResponse{
		ID:          f.ID,
		RequestHash: f.RequestHash,
		Rating:      f.Rating,
	})
}

// GetFeedbackStats handles GET /api/v1/mdflow/feedback/stats.
// Response: { "total_count": 100, "positive_rate": 0.85, ... }
func (h *FeedbackHandler) GetFeedbackStats(c *gin.Context) {
	stats, err := h.store.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve feedback stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
