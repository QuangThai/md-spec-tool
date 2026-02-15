package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func TestNewAIServiceProvider(t *testing.T) {
	cfg := config.LoadConfig()
	provider := handlers.NewAIServiceProvider(cfg)

	if provider == nil {
		t.Error("expected non-nil provider")
	}
	if !handlers.AIServiceProviderHasBYOKCacheForTest(provider) {
		t.Error("expected non-nil byokCache")
	}
}

func TestGetAIServiceForRequestWithoutBYOK(t *testing.T) {
	cfg := config.LoadConfig()
	provider := handlers.NewAIServiceProvider(cfg)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/test", nil)

	// Without BYOK header, should return default (nil)
	service := provider.GetAIServiceForRequest(c)
	if service != nil && !handlers.AIServiceProviderHasDefaultAIForTest(provider) {
		t.Error("expected nil service when no BYOK header and no default AI")
	}
}

func TestHasAIForRequestWithoutBYOK(t *testing.T) {
	cfg := config.LoadConfig()
	provider := handlers.NewAIServiceProvider(cfg)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/test", nil)

	// Without BYOK header and without default AI, should return false
	hasAI := provider.HasAIForRequest(c)
	if hasAI && !handlers.AIServiceProviderHasDefaultAIForTest(provider) {
		t.Error("expected false when no BYOK header and no default AI")
	}
}

func TestHasAIForRequestWithBYOKHeader(t *testing.T) {
	cfg := config.LoadConfig()
	provider := handlers.NewAIServiceProvider(cfg)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/test", nil)
	c.Request.Header.Set("X-OpenAI-API-Key", "test-key")

	// With BYOK header, should return true
	hasAI := provider.HasAIForRequest(c)
	if !hasAI {
		t.Error("expected true when BYOK header is present")
	}
}

func TestGetConverterForRequest(t *testing.T) {
	cfg := config.LoadConfig()
	provider := handlers.NewAIServiceProvider(cfg)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/test", nil)

	// Create a base converter
	baseConverter := handlers.ConvertHandlerConverterForTest(handlers.NewConvertHandler(nil, cfg, provider))

	// Should return the base converter without modification
	conv := provider.GetConverterForRequest(c, baseConverter)
	if conv == nil {
		t.Error("expected non-nil converter")
	}
}

func TestAIServiceProviderClose(t *testing.T) {
	cfg := config.LoadConfig()
	provider := handlers.NewAIServiceProvider(cfg)

	// Should not panic
	provider.Close()
}
