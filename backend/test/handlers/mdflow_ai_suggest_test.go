package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func TestGetAISuggestions_WithProviderAndBYOK_NoPanic(t *testing.T) {
	cfg := config.LoadConfig()
	provider := handlers.NewAIServiceProvider(cfg)
	h := handlers.NewMDFlowHandlerWithProvider(
		converter.NewConverter(),
		converter.NewMDFlowRenderer(),
		nil,
		cfg,
		provider,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/mdflow/ai/suggest", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set(handlers.BYOKHeader, "sk-test-byok")

	h.GetAISuggestions(c)

	if w.Code == http.StatusInternalServerError {
		t.Fatalf("expected no panic path, got status %d with body: %s", w.Code, w.Body.String())
	}
}
