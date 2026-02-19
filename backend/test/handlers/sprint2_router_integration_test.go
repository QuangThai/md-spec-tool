package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mdhttp "github.com/yourorg/md-spec-tool/internal/http"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func setupRouterWithSprint2Limits(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.LoadConfig()
	cfg.PreviewRateLimit = 1
	cfg.ConvertRateLimit = 1
	cfg.AISuggestRateLimit = 1
	cfg.RateLimitWindow = time.Minute
	cfg.ShareStorePath = filepath.Join(t.TempDir(), "share-store.json")

	router, cleanup := mdhttp.SetupRouterWithCleanup(cfg)
	t.Cleanup(cleanup)

	return router
}

func performJSONRequest(t *testing.T, router http.Handler, method, path, body, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestConvertPasteSetsNeedsReviewWhenColumnsUnmapped(t *testing.T) {
	router := setupRouterWithSprint2Limits(t)

	tsvContent := []byte("1\t2\t3\n4\t5\t6\n7\t8\t9")
	body, contentType, err := createMultipartForm("sample.tsv", tsvContent)
	if err != nil {
		t.Fatalf("failed to build multipart form: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/mdflow/tsv", body)
	req.RemoteAddr = "10.10.10.10:1234"
	req.Header.Set("Content-Type", contentType)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body: %s", recorder.Code, recorder.Body.String())
	}

	var resp handlers.MDFlowConvertResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.NeedsReview {
		t.Fatalf("expected needs_review=true, got false (warnings=%d, unmapped=%d)", len(resp.Warnings), len(resp.Meta.UnmappedColumns))
	}
}

func TestRateLimitPreviewEndpoint(t *testing.T) {
	router := setupRouterWithSprint2Limits(t)
	remoteAddr := "10.0.0.11:1234"
	payload := `{"paste_text":"Header1\tHeader2\nA\tB","template":"spec"}`

	first := performJSONRequest(t, router, http.MethodPost, "/api/mdflow/preview", payload, remoteAddr)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request status 200, got %d", first.Code)
	}

	second := performJSONRequest(t, router, http.MethodPost, "/api/mdflow/preview", payload, remoteAddr)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status 429, got %d", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}
}

func TestRateLimitConvertSharedAcrossLegacyAndV1(t *testing.T) {
	router := setupRouterWithSprint2Limits(t)
	remoteAddr := "10.0.0.12:1234"
	payload := `{"paste_text":"Header1\tHeader2\nA\tB","template":"spec","format":"spec"}`

	first := performJSONRequest(t, router, http.MethodPost, "/api/mdflow/paste", payload, remoteAddr)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request status 200, got %d", first.Code)
	}

	second := performJSONRequest(t, router, http.MethodPost, "/api/v1/mdflow/paste", payload, remoteAddr)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status 429, got %d", second.Code)
	}
}

func TestRateLimitAISuggestEndpoint(t *testing.T) {
	router := setupRouterWithSprint2Limits(t)
	remoteAddr := "10.0.0.13:1234"
	payload := `{"paste_text":"Feature\tScenario\nLogin\tSuccessful login","template":"spec"}`

	first := performJSONRequest(t, router, http.MethodPost, "/api/mdflow/ai/suggest", payload, remoteAddr)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request status 200, got %d", first.Code)
	}

	second := performJSONRequest(t, router, http.MethodPost, "/api/mdflow/ai/suggest", payload, remoteAddr)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status 429, got %d", second.Code)
	}
}

func TestTelemetryIngestEndpointUnaffectedByHeavyEndpointRateLimits(t *testing.T) {
	router := setupRouterWithSprint2Limits(t)
	remoteAddr := "10.0.0.14:1234"
	body := bytes.NewBufferString(`{
		"events":[
			{"event_name":"studio_opened","event_time":"2026-01-01T00:00:00Z","status":"success"}
		]
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/telemetry/events", body)
	req.RemoteAddr = remoteAddr
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}
}
