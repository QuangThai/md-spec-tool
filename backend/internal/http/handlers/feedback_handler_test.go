package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/feedback"
)

// mockFeedbackStore is a test double for feedback.StoreInterface.
type mockFeedbackStore struct {
	submitted []feedback.Feedback
	submitErr error
	stats     *feedback.FeedbackStats
	statsErr  error
}

func (m *mockFeedbackStore) Submit(f *feedback.Feedback) error {
	if m.submitErr != nil {
		return m.submitErr
	}
	f.ID = int64(len(m.submitted) + 1)
	m.submitted = append(m.submitted, *f)
	return nil
}

func (m *mockFeedbackStore) GetStats() (*feedback.FeedbackStats, error) {
	return m.stats, m.statsErr
}

func (m *mockFeedbackStore) GetByRequestHash(hash string) ([]feedback.Feedback, error) {
	var results []feedback.Feedback
	for _, f := range m.submitted {
		if f.RequestHash == hash {
			results = append(results, f)
		}
	}
	return results, nil
}

func (m *mockFeedbackStore) Close() error { return nil }

// Verify interface compliance at compile time.
var _ feedback.StoreInterface = (*mockFeedbackStore)(nil)

// setupFeedbackRouter creates a test Gin router with FeedbackHandler mounted.
func setupFeedbackRouter(store feedback.StoreInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewFeedbackHandler(store)
	r.POST("/api/v1/mdflow/feedback", h.SubmitFeedback)
	r.GET("/api/v1/mdflow/feedback/stats", h.GetFeedbackStats)
	return r
}

// TestFeedbackHandler_SubmitSuccess verifies a valid feedback submission returns 201.
func TestFeedbackHandler_SubmitSuccess(t *testing.T) {
	store := &mockFeedbackStore{}
	router := setupFeedbackRouter(store)

	body, _ := json.Marshal(SubmitFeedbackRequest{
		RequestHash: "sha256abcdef",
		Rating:      5,
		Corrections: "All columns correct",
		SessionID:   "sess-abc",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mdflow/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp SubmitFeedbackResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == 0 {
		t.Error("expected non-zero ID in response")
	}
	if resp.RequestHash != "sha256abcdef" {
		t.Errorf("unexpected request_hash: %s", resp.RequestHash)
	}
	if resp.Rating != 5 {
		t.Errorf("unexpected rating: %d", resp.Rating)
	}

	if len(store.submitted) != 1 {
		t.Errorf("expected 1 stored entry, got %d", len(store.submitted))
	}
}

// TestFeedbackHandler_SubmitInvalidBody verifies that a malformed JSON body returns 400.
func TestFeedbackHandler_SubmitInvalidBody(t *testing.T) {
	store := &mockFeedbackStore{}
	router := setupFeedbackRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mdflow/feedback", bytes.NewBufferString("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestFeedbackHandler_SubmitInvalidRating verifies that an invalid rating returns 400.
func TestFeedbackHandler_SubmitInvalidRating(t *testing.T) {
	store := &mockFeedbackStore{submitErr: feedback.ErrInvalidRating}
	router := setupFeedbackRouter(store)

	body, _ := json.Marshal(SubmitFeedbackRequest{
		RequestHash: "abc",
		Rating:      3,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mdflow/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestFeedbackHandler_SubmitMissingHash verifies that missing request_hash returns 400.
func TestFeedbackHandler_SubmitMissingHash(t *testing.T) {
	store := &mockFeedbackStore{}
	router := setupFeedbackRouter(store)

	body, _ := json.Marshal(map[string]interface{}{
		"rating": 5,
		// request_hash omitted
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mdflow/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestFeedbackHandler_SubmitStoreError verifies that a store error returns 500.
func TestFeedbackHandler_SubmitStoreError(t *testing.T) {
	store := &mockFeedbackStore{submitErr: errors.New("db locked")}
	router := setupFeedbackRouter(store)

	body, _ := json.Marshal(SubmitFeedbackRequest{
		RequestHash: "abc",
		Rating:      1,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/mdflow/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// TestFeedbackHandler_GetStats verifies the stats endpoint returns correct data.
func TestFeedbackHandler_GetStats(t *testing.T) {
	store := &mockFeedbackStore{
		stats: &feedback.FeedbackStats{
			TotalCount:    10,
			PositiveCount: 8,
			NegativeCount: 2,
			PositiveRate:  0.8,
			RecentTrend:   "improving",
		},
	}
	router := setupFeedbackRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mdflow/feedback/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp feedback.FeedbackStats
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TotalCount != 10 {
		t.Errorf("TotalCount: want 10, got %d", resp.TotalCount)
	}
	if resp.PositiveRate != 0.8 {
		t.Errorf("PositiveRate: want 0.8, got %f", resp.PositiveRate)
	}
	if resp.RecentTrend != "improving" {
		t.Errorf("RecentTrend: want 'improving', got %q", resp.RecentTrend)
	}
}

// TestFeedbackHandler_GetStatsError verifies stats store error returns 500.
func TestFeedbackHandler_GetStatsError(t *testing.T) {
	store := &mockFeedbackStore{
		statsErr: errors.New("db error"),
	}
	router := setupFeedbackRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mdflow/feedback/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}
