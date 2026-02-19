package handlers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// AIServiceProvider manages AI service creation and caching for BYOK (Bring Your Own Key)
type AIServiceProvider struct {
	cfg       *config.Config
	byokCache *ai.BYOKServiceCache
	defaultAI ai.Service
}

// NewAIServiceProvider creates a new AI service provider
func NewAIServiceProvider(cfg *config.Config) *AIServiceProvider {
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	provider := &AIServiceProvider{
		cfg: cfg,
	}

	// Initialize consolidated BYOK cache
	provider.byokCache = ai.NewBYOKServiceCache(
		ai.BYOKServiceCacheConfig{
			TTL:           cfg.BYOKCacheTTL,
			CleanupTicker: cfg.BYOKCleanupTicker,
			MaxEntries:    cfg.BYOKMaxEntries,
		},
		provider.newAIService,
	)

	return provider
}

// newAIService creates a new AI service with the given API key
func (p *AIServiceProvider) newAIService(apiKey string) (ai.Service, error) {
	aiCfg := ai.DefaultConfig()
	aiCfg.APIKey = apiKey
	aiCfg.DisableCache = true // BYOK: isolate per-user
	if p.cfg != nil {
		model := p.cfg.OpenAIConvertModel
		if model == "" {
			model = p.cfg.OpenAIModel
		}
		aiCfg.Model = model
		aiCfg.RequestTimeout = p.cfg.AIRequestTimeout
		aiCfg.MaxRetries = p.cfg.AIMaxRetries
		aiCfg.CacheTTL = p.cfg.AICacheTTL
		aiCfg.MaxCacheSize = p.cfg.AIMaxCacheSize
		aiCfg.RetryBaseDelay = p.cfg.AIRetryBaseDelay
		aiCfg.MaxCompletionTokens = p.cfg.AIConvertMaxTokens
		aiCfg.PromptProfile = p.cfg.AIPromptProfile
	}
	return ai.NewService(aiCfg)
}

// SetDefaultAIService sets the default (server-configured) AI service
func (p *AIServiceProvider) SetDefaultAIService(service ai.Service) {
	p.defaultAI = service
}

// GetAIServiceForRequest returns an AI service for the current request
// If X-OpenAI-API-Key header is present, uses/caches per-key service (TTL 5min)
// Otherwise falls back to the default (server-configured) service (may be nil)
func (p *AIServiceProvider) GetAIServiceForRequest(c *gin.Context) ai.Service {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return p.defaultAI
	}

	service, err := p.GetAIServiceForAPIKey(userKey)
	if err != nil {
		return nil
	}
	return service
}

// GetAIServiceForAPIKey returns a cached AI service for a specific BYOK API key.
func (p *AIServiceProvider) GetAIServiceForAPIKey(apiKey string) (ai.Service, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return p.defaultAI, nil
	}
	if p.byokCache == nil {
		return nil, fmt.Errorf("byok service cache unavailable")
	}

	service, err := p.byokCache.GetOrCreate(apiKey)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetConverterForRequest returns a converter with the appropriate AI service
// If X-OpenAI-API-Key header is present, clones the base converter with the user's AI service
// Otherwise returns the provided converter as-is
// This avoids re-initializing the TemplateRegistry and other expensive components.
func (p *AIServiceProvider) GetConverterForRequest(c *gin.Context, baseConverter *converter.Converter) *converter.Converter {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return baseConverter
	}

	// User provided BYOK key: must create a converter with that service (or no AI if unavailable)
	// Do NOT fall back to base converter's AI service
	aiService := p.GetAIServiceForRequest(c)
	if aiService == nil {
		// Clone without AI to ensure isolation from base converter's AI service
		return baseConverter.CloneWithAIService(nil)
	}

	return baseConverter.CloneWithAIService(aiService)
}

// HasAIForRequest checks if AI is available for this request (default or BYOK)
func (p *AIServiceProvider) HasAIForRequest(c *gin.Context) bool {
	return p.defaultAI != nil || getUserAPIKey(c) != ""
}

// Close gracefully shuts down the BYOK cache
func (p *AIServiceProvider) Close() {
	if p.byokCache != nil {
		p.byokCache.Close()
	}
}

// getUserAPIKey extracts the user-provided OpenAI API key from the request header
func getUserAPIKey(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader(BYOKHeader))
}

const BYOKHeader = "X-OpenAI-API-Key"
