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
	converter      *converter.Converter
	renderer       *converter.MDFlowRenderer
	aiSuggester    *suggest.Suggester
	aiService      ai.Service
	httpClient     *http.Client
	cfg            *config.Config
	sheetsService  *sheets.Service
	sheetsInitOnce sync.Once
	sheetsInitErr  error
}

// NewMDFlowHandler creates a new MDFlowHandler with injected dependencies
// httpClient is optional (can be nil); if nil, a default client will be used internally
func NewMDFlowHandler(conv *converter.Converter, rend *converter.MDFlowRenderer, httpClient *http.Client, cfg *config.Config) *MDFlowHandler {
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

	return &MDFlowHandler{
		converter:   conv,
		renderer:    rend,
		aiSuggester: nil,
		httpClient:  httpClient,
		cfg:         cfg,
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
// =============================================================================

const BYOKHeader = "X-OpenAI-API-Key"

// getUserAPIKey extracts the user-provided OpenAI API key from the request header
func getUserAPIKey(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader(BYOKHeader))
}

// getAIServiceForRequest returns an AI service for the current request.
// If X-OpenAI-API-Key header is present, creates a per-request service with the user's key.
// Otherwise falls back to the server-configured service (may be nil).
func (h *MDFlowHandler) getAIServiceForRequest(c *gin.Context) ai.Service {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return h.aiService
	}

	aiCfg := ai.DefaultConfig()
	aiCfg.APIKey = userKey
	if h.cfg != nil {
		aiCfg.Model = h.cfg.OpenAIModel
		aiCfg.RequestTimeout = h.cfg.AIRequestTimeout
		aiCfg.MaxRetries = h.cfg.AIMaxRetries
		aiCfg.CacheTTL = h.cfg.AICacheTTL
		aiCfg.MaxCacheSize = h.cfg.AIMaxCacheSize
		aiCfg.RetryBaseDelay = h.cfg.AIRetryBaseDelay
	}

	service, err := ai.NewService(aiCfg)
	if err != nil {
		slog.Warn("BYOK: failed to create AI service with user key", "error", err)
		return h.aiService
	}
	return service
}

// getConverterForRequest returns a converter with the appropriate AI service.
// If X-OpenAI-API-Key header is present, creates a new converter with the user's AI service.
// Otherwise returns the server-configured converter.
func (h *MDFlowHandler) getConverterForRequest(c *gin.Context) *converter.Converter {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return h.converter
	}

	aiService := h.getAIServiceForRequest(c)
	if aiService == nil {
		return h.converter
	}

	conv := converter.NewConverter()
	if h.cfg != nil && h.cfg.UseNewConverterPipeline {
		conv.WithNewPipeline()
	}
	conv.WithAIService(aiService)
	return conv
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
		return h.aiSuggester
	}

	return suggest.NewSuggester(aiService)
}

// hasAIForRequest checks if AI is available for this request (server-configured or BYOK)
func (h *MDFlowHandler) hasAIForRequest(c *gin.Context) bool {
	return h.converter.HasAIService() || getUserAPIKey(c) != ""
}
