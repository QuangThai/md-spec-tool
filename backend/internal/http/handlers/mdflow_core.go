package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/suggest"
	"google.golang.org/api/sheets/v4"
)

type MDFlowHandler struct {
	converter        *converter.Converter
	renderer         *converter.MDFlowRenderer
	aiSuggester      *suggest.Suggester
	aiService        ai.Service
	aiProvider       *AIServiceProvider // Shared AI service provider for BYOK support
	httpClient       *http.Client
	gsheetHTTPClient *http.Client
	gsheetClientOnce sync.Once
	cfg              *config.Config
	sheetsService    *sheets.Service
	sheetsInitOnce   sync.Once
	sheetsInitErr    error
	byokCache        *ai.BYOKServiceCache // Consolidated BYOK cache
}

// NewMDFlowHandler creates a new MDFlowHandler with injected dependencies
// httpClient is optional (can be nil); if nil, a default client will be used internally
func NewMDFlowHandler(conv *converter.Converter, rend *converter.MDFlowRenderer, httpClient *http.Client, cfg *config.Config) *MDFlowHandler {
	return NewMDFlowHandlerWithProvider(conv, rend, httpClient, cfg, nil)
}

// NewMDFlowHandlerWithProvider creates a new MDFlowHandler with a shared AIServiceProvider
// This avoids duplicate caches when the provider is available
func NewMDFlowHandlerWithProvider(conv *converter.Converter, rend *converter.MDFlowRenderer, httpClient *http.Client, cfg *config.Config, aiProvider *AIServiceProvider) *MDFlowHandler {
	if conv == nil {
		conv = converter.NewConverter()
	}
	if rend == nil {
		rend = converter.NewMDFlowRenderer()
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	h := &MDFlowHandler{
		converter:   conv,
		renderer:    rend,
		aiSuggester: nil,
		aiProvider:  aiProvider,
		httpClient:  httpClient,
		cfg:         cfg,
	}

	// Only initialize local cache if no shared provider is available
	// Otherwise, the handler will use the shared provider's cache
	if aiProvider == nil {
		h.byokCache = ai.NewBYOKServiceCache(
			ai.BYOKServiceCacheConfig{
				TTL:           cfg.BYOKCacheTTL,
				CleanupTicker: cfg.BYOKCleanupTicker,
				MaxEntries:    cfg.BYOKMaxEntries,
			},
			h.newAIService,
		)
	}

	return h
}

// newAIService creates a new AI service with the given API key
func (h *MDFlowHandler) newAIService(apiKey string) (ai.Service, error) {
	aiCfg := ai.DefaultConfig()
	aiCfg.APIKey = apiKey
	aiCfg.DisableCache = true // BYOK: isolate per-user
	if h.cfg != nil {
		aiCfg.Model = h.cfg.OpenAIModel
		aiCfg.RequestTimeout = h.cfg.AIRequestTimeout
		aiCfg.MaxRetries = h.cfg.AIMaxRetries
		aiCfg.CacheTTL = h.cfg.AICacheTTL
		aiCfg.MaxCacheSize = h.cfg.AIMaxCacheSize
		aiCfg.RetryBaseDelay = h.cfg.AIRetryBaseDelay
	}
	return ai.NewService(aiCfg)
}

// Close gracefully shuts down the BYOK cache
func (h *MDFlowHandler) Close() {
	if h.byokCache != nil {
		h.byokCache.Close()
	}
}

// SetAISuggester sets the AI suggester instance
func (h *MDFlowHandler) SetAISuggester(suggester *suggest.Suggester) {
	h.aiSuggester = suggester
}

// SetAIService sets the AI service instance for semantic validation
func (h *MDFlowHandler) SetAIService(service ai.Service) {
	h.aiService = service
}

const (
	maxTemplateLen  = 64
	maxSheetNameLen = 128
)

const (
	googleSheetsCredsEnv     = "GOOGLE_APPLICATION_CREDENTIALS"
	googleSheetsExportURLFmt = "https://docs.google.com/spreadsheets/d/%s/export?format=csv"
)

func humanSize(b int64) string {
	if b >= 1<<20 {
		return fmt.Sprintf("%dMB", b>>20)
	}
	if b >= 1<<10 {
		return fmt.Sprintf("%dKB", b>>10)
	}
	return fmt.Sprintf("%d bytes", b)
}

var errSheetsNotConfigured = errors.New("google sheets credentials not configured")

var xlsxMagic = []byte{0x50, 0x4B, 0x03, 0x04}

func (h *MDFlowHandler) validateTemplate(template string) error {
	_, err := normalizeTemplate(template)
	return err
}

func (h *MDFlowHandler) validateFormat(format string) error {
	_, err := normalizeFormat(format)
	return err
}

func normalizeTemplate(template string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(template))
	if normalized == "" {
		return "", nil
	}
	if len(normalized) > maxTemplateLen {
		return "", fmt.Errorf("template exceeds %d characters", maxTemplateLen)
	}
	switch normalized {
	case "spec", "table":
		return normalized, nil
	default:
		return "", fmt.Errorf("unknown template: %s (supported: spec, table)", template)
	}
}

func normalizeFormat(format string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(format))
	if normalized == "" {
		return "", nil
	}
	switch normalized {
	case "spec", "table":
		return normalized, nil
	default:
		return "", fmt.Errorf("unknown format: %s (supported: spec, table)", format)
	}
}

func normalizeTemplateAndFormat(template string, format string) (string, string, error) {
	normalizedTemplate, err := normalizeTemplate(template)
	if err != nil {
		return "", "", err
	}
	normalizedFormat, err := normalizeFormat(format)
	if err != nil {
		return "", "", err
	}

	if normalizedTemplate == "" && normalizedFormat != "" {
		normalizedTemplate = normalizedFormat
	}
	if normalizedFormat == "" && normalizedTemplate != "" {
		normalizedFormat = normalizedTemplate
	}
	if normalizedTemplate == "" && normalizedFormat == "" {
		normalizedTemplate = "spec"
		normalizedFormat = "spec"
	}

	return normalizedTemplate, normalizedFormat, nil
}

func validateSheetName(sheetName string) error {
	if sheetName == "" {
		return nil
	}
	if len(sheetName) > maxSheetNameLen {
		return fmt.Errorf("sheet_name exceeds %d characters", maxSheetNameLen)
	}
	if strings.IndexFunc(sheetName, unicode.IsControl) != -1 {
		return fmt.Errorf("sheet_name contains invalid characters")
	}
	return nil
}

// =============================================================================
// BYOK (Bring Your Own Key) support
// Allows users to provide their own OpenAI API key via request header.
// When present, a per-request AI service is created with the user's key.
// Note: BYOKHeader and getUserAPIKey are defined in ai_service_provider.go
// =============================================================================

// getAIServiceForRequest returns an AI service for the current request.
// If X-OpenAI-API-Key header is present, uses/caches per-key service (TTL 5min).
// Otherwise falls back to the server-configured service (may be nil).
func (h *MDFlowHandler) getAIServiceForRequest(c *gin.Context) ai.Service {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return h.aiService
	}

	if h.aiProvider != nil {
		return h.aiProvider.GetAIServiceForRequest(c)
	}

	if h.byokCache == nil {
		slog.Warn("BYOK: cache not initialized")
		return nil
	}

	service, err := h.byokCache.GetOrCreate(userKey)
	if err != nil {
		slog.Warn("BYOK: failed to get/create AI service", "error", err)
		return nil
	}
	return service
}

// getConverterForRequest returns a converter with the appropriate AI service.
// If X-OpenAI-API-Key header is present, clones the server converter with the user's AI service.
// Otherwise returns the server-configured converter.
// This avoids re-initializing the TemplateRegistry and other expensive components.
func (h *MDFlowHandler) getConverterForRequest(c *gin.Context) *converter.Converter {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return h.converter
	}

	aiService := h.getAIServiceForRequest(c)
	if aiService == nil {
		return h.converter
	}

	return h.converter.CloneWithAIService(aiService)
}

// getSuggesterForRequest returns an AI suggester for the current request.
// If X-OpenAI-API-Key header is present, creates a per-request suggester.
// Otherwise falls back to the server-configured suggester (may be nil).
func (h *MDFlowHandler) getSuggesterForRequest(c *gin.Context) *suggest.Suggester {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return h.aiSuggester
	}

	aiService := h.getAIServiceForRequest(c)
	if aiService == nil {
		return nil
	}

	return suggest.NewSuggester(aiService)
}

// hasAIForRequest checks if AI is available for this request (server-configured or BYOK)
func (h *MDFlowHandler) hasAIForRequest(c *gin.Context) bool {
	return h.converter.HasAIService() || getUserAPIKey(c) != ""
}
