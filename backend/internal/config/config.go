package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Default values
const (
	DefaultHost        = "0.0.0.0"
	DefaultPort        = "8080"
	DefaultOpenAIModel = "gpt-4o-mini"

	// Upload/Paste limits
	DefaultMaxUploadBytes      = 10 << 20 // 10MB
	DefaultMaxPasteBytes       = 1 << 20  // 1MB
	DefaultMaxAudioUploadBytes = 30 << 20 // 30MB

	// HTTP client timeout for external calls (Google Sheets, etc.)
	DefaultHTTPClientTimeout = 30 * time.Second
	DefaultGSheetMaxRetries  = 2

	// Rate limiting defaults
	DefaultShareCreateRateLimit  = 10
	DefaultShareUpdateRateLimit  = 20
	DefaultShareCommentRateLimit = 20
	DefaultRateLimitWindow       = time.Minute
	DefaultPreviewRateLimit      = 60
	DefaultConvertRateLimit      = 60
	DefaultAISuggestRateLimit    = 30
	DefaultTrustedProxies        = "127.0.0.1,::1"

	// AI defaults
	DefaultAIRequestTimeout   = 30 * time.Second
	DefaultAIMaxRetries       = 3
	DefaultAICacheTTL         = 1 * time.Hour
	DefaultAIMaxCacheSize     = 1000
	DefaultAIRetryBaseDelay   = 1 * time.Second
	DefaultAISuggestTimeout   = 45 * time.Second
	DefaultAIPreviewModel     = ""
	DefaultAIConvertModel     = ""
	DefaultAISuggestModel     = ""
	DefaultAIPromptProfile    = "static_v3"
	DefaultAIPreviewMaxTokens = 600
	DefaultAIConvertMaxTokens = 1200
	DefaultAISuggestMaxTokens = 900

	// AI preview defaults (reduced for fast response when skip_ai=false)
	DefaultAIPreviewTimeout    = 10 * time.Second
	DefaultAIPreviewMaxRetries = 1

	// BYOK cache defaults
	DefaultBYOKCacheTTL      = 5 * time.Minute
	DefaultBYOKCleanupTicker = 1 * time.Minute
	DefaultBYOKMaxEntries    = 1000

	// Spec validation defaults
	DefaultSpecStrictMode          = true
	DefaultSpecMinHeaderConfidence = 60
	DefaultSpecMaxRowLossRatio     = 0.4
)

type Config struct {
	// Server
	Host        string
	Port        string
	CORSOrigins []string

	// Upload/Paste limits
	MaxUploadBytes      int64
	MaxPasteBytes       int64
	MaxAudioUploadBytes int64

	// HTTP client
	HTTPClientTimeout time.Duration

	// Google Sheets (optional; defaults to HTTPClientTimeout if not set)
	GSheetHTTPTimeout time.Duration
	GSheetMaxRetries  int

	// Rate limiting
	ShareCreateRateLimit  int
	ShareUpdateRateLimit  int
	ShareCommentRateLimit int
	PreviewRateLimit      int
	ConvertRateLimit      int
	AISuggestRateLimit    int
	RateLimitWindow       time.Duration
	TrustedProxies        []string

	// AI configuration
	OpenAIAPIKey       string
	OpenAIModel        string
	OpenAIPreviewModel string
	OpenAIConvertModel string
	OpenAISuggestModel string
	AIPromptProfile    string
	AIEnabled          bool // Auto-enabled when OPENAI_API_KEY is set
	AIRequestTimeout   time.Duration
	AISuggestTimeout   time.Duration
	AIMaxRetries       int
	AICacheTTL         time.Duration
	AIMaxCacheSize     int
	AIRetryBaseDelay   time.Duration
	AIPreviewMaxTokens int
	AIConvertMaxTokens int
	AISuggestMaxTokens int

	// AI preview configuration (reduced timeout/retries for when skip_ai=false on preview)
	AIPreviewTimeout    time.Duration
	AIPreviewMaxRetries int

	// BYOK cache configuration
	BYOKCacheTTL      time.Duration
	BYOKCleanupTicker time.Duration
	BYOKMaxEntries    int

	// Telemetry
	TelemetryMaxEvents int

	// Storage
	ShareStorePath string
	FeedbackDBPath string

	// Spec validation
	SpecStrictMode          bool
	SpecMinHeaderConfidence int
	SpecMaxRowLossRatio     float64
}

func LoadConfig() *Config {
	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080")
	parsedCORSOrigins := splitCSV(corsOrigins)
	if len(parsedCORSOrigins) == 0 {
		parsedCORSOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	}

	openAIAPIKey := getEnv("OPENAI_API_KEY", "")
	aiEnabled := openAIAPIKey != ""

	if aiEnabled {
		slog.Info("AI features enabled (OPENAI_API_KEY is set)")
	} else {
		slog.Info("AI features disabled (OPENAI_API_KEY not set)")
	}

	return &Config{
		// Server
		Host:        getEnv("HOST", DefaultHost),
		Port:        getEnv("PORT", DefaultPort),
		CORSOrigins: parsedCORSOrigins,

		// Upload/Paste limits
		MaxUploadBytes:      getEnvInt64("MAX_UPLOAD_BYTES", DefaultMaxUploadBytes),
		MaxPasteBytes:       getEnvInt64("MAX_PASTE_BYTES", DefaultMaxPasteBytes),
		MaxAudioUploadBytes: getEnvInt64("MAX_AUDIO_UPLOAD_BYTES", DefaultMaxAudioUploadBytes),

		// HTTP client
		HTTPClientTimeout: getEnvDuration("HTTP_CLIENT_TIMEOUT", DefaultHTTPClientTimeout),

		// Google Sheets
		GSheetHTTPTimeout: getEnvDuration("GSHEET_HTTP_TIMEOUT", DefaultHTTPClientTimeout+15*time.Second),
		GSheetMaxRetries:  getEnvInt("GSHEET_MAX_RETRIES", DefaultGSheetMaxRetries),

		// Rate limiting
		ShareCreateRateLimit:  getEnvInt("SHARE_CREATE_RATE_LIMIT", DefaultShareCreateRateLimit),
		ShareUpdateRateLimit:  getEnvInt("SHARE_UPDATE_RATE_LIMIT", DefaultShareUpdateRateLimit),
		ShareCommentRateLimit: getEnvInt("SHARE_COMMENT_RATE_LIMIT", DefaultShareCommentRateLimit),
		PreviewRateLimit:      getEnvInt("PREVIEW_RATE_LIMIT", DefaultPreviewRateLimit),
		ConvertRateLimit:      getEnvInt("CONVERT_RATE_LIMIT", DefaultConvertRateLimit),
		AISuggestRateLimit:    getEnvInt("AI_SUGGEST_RATE_LIMIT", DefaultAISuggestRateLimit),
		RateLimitWindow:       getEnvDuration("RATE_LIMIT_WINDOW", DefaultRateLimitWindow),
		TrustedProxies:        splitCSV(getEnv("TRUSTED_PROXIES", DefaultTrustedProxies)),

		// AI configuration
		OpenAIAPIKey:       openAIAPIKey,
		OpenAIModel:        getEnv("OPENAI_MODEL", DefaultOpenAIModel),
		OpenAIPreviewModel: getEnv("OPENAI_MODEL_PREVIEW", DefaultAIPreviewModel),
		OpenAIConvertModel: getEnv("OPENAI_MODEL_CONVERT", DefaultAIConvertModel),
		OpenAISuggestModel: getEnv("OPENAI_MODEL_SUGGEST", DefaultAISuggestModel),
		AIPromptProfile:    getEnv("AI_PROMPT_PROFILE", DefaultAIPromptProfile),
		AIEnabled:          aiEnabled,
		AIRequestTimeout:   getEnvDuration("AI_REQUEST_TIMEOUT", DefaultAIRequestTimeout),
		AISuggestTimeout:   getEnvDuration("AI_SUGGEST_TIMEOUT", DefaultAISuggestTimeout),
		AIMaxRetries:       getEnvInt("AI_MAX_RETRIES", DefaultAIMaxRetries),
		AICacheTTL:         getEnvDuration("AI_CACHE_TTL", DefaultAICacheTTL),
		AIMaxCacheSize:     getEnvInt("AI_MAX_CACHE_SIZE", DefaultAIMaxCacheSize),
		AIRetryBaseDelay:   getEnvDuration("AI_RETRY_BASE_DELAY", DefaultAIRetryBaseDelay),
		AIPreviewMaxTokens: getEnvInt("AI_PREVIEW_MAX_TOKENS", DefaultAIPreviewMaxTokens),
		AIConvertMaxTokens: getEnvInt("AI_CONVERT_MAX_TOKENS", DefaultAIConvertMaxTokens),
		AISuggestMaxTokens: getEnvInt("AI_SUGGEST_MAX_TOKENS", DefaultAISuggestMaxTokens),

		// AI preview configuration
		AIPreviewTimeout:    getEnvDuration("AI_PREVIEW_TIMEOUT", DefaultAIPreviewTimeout),
		AIPreviewMaxRetries: getEnvInt("AI_PREVIEW_MAX_RETRIES", DefaultAIPreviewMaxRetries),

		// BYOK cache configuration
		BYOKCacheTTL:      getEnvDuration("BYOK_CACHE_TTL", DefaultBYOKCacheTTL),
		BYOKCleanupTicker: getEnvDuration("BYOK_CLEANUP_TICKER", DefaultBYOKCleanupTicker),
		BYOKMaxEntries:    getEnvInt("BYOK_MAX_ENTRIES", DefaultBYOKMaxEntries),

		// Telemetry
		TelemetryMaxEvents: getEnvInt("TELEMETRY_MAX_EVENTS", 10000),

		// Storage
		ShareStorePath: getEnv("SHARE_STORE_PATH", ""),
		FeedbackDBPath: getEnv("FEEDBACK_DB_PATH", ".cache/feedback.db"),

		// Spec validation
		SpecStrictMode:          getEnvBool("SPEC_STRICT_MODE", DefaultSpecStrictMode),
		SpecMinHeaderConfidence: getEnvInt("SPEC_MIN_HEADER_CONFIDENCE", DefaultSpecMinHeaderConfidence),
		SpecMaxRowLossRatio:     getEnvFloat64("SPEC_MAX_ROW_LOSS_RATIO", DefaultSpecMaxRowLossRatio),
	}
}

// ValidateConfig checks config values and returns an error on failure.
// Call after LoadConfig to fail fast on invalid configuration.
func ValidateConfig(cfg *Config) error {
	if cfg.MaxPasteBytes > cfg.MaxUploadBytes {
		return fmt.Errorf("MAX_PASTE_BYTES (%d) must not exceed MAX_UPLOAD_BYTES (%d)", cfg.MaxPasteBytes, cfg.MaxUploadBytes)
	}
	if cfg.MaxUploadBytes <= 0 || cfg.MaxPasteBytes <= 0 {
		return fmt.Errorf("MAX_UPLOAD_BYTES and MAX_PASTE_BYTES must be positive")
	}
	if cfg.MaxAudioUploadBytes <= 0 {
		return fmt.Errorf("MAX_AUDIO_UPLOAD_BYTES must be positive")
	}
	if cfg.Port != "" {
		if _, err := strconv.Atoi(cfg.Port); err != nil {
			return fmt.Errorf("PORT must be numeric, got %q", cfg.Port)
		}
	}
	if len(cfg.CORSOrigins) == 0 {
		return fmt.Errorf("CORS_ORIGINS must have at least one origin")
	}
	for _, origin := range cfg.CORSOrigins {
		if origin == "" || !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return fmt.Errorf("CORS_ORIGINS entry %q must be a valid http(s) URL", origin)
		}
	}
	if cfg.ShareCreateRateLimit <= 0 || cfg.ShareUpdateRateLimit <= 0 || cfg.ShareCommentRateLimit <= 0 {
		return fmt.Errorf("share rate limits must be positive")
	}
	if cfg.PreviewRateLimit <= 0 || cfg.ConvertRateLimit <= 0 || cfg.AISuggestRateLimit <= 0 {
		return fmt.Errorf("preview/convert/ai_suggest rate limits must be positive")
	}
	if cfg.SpecMinHeaderConfidence < 0 || cfg.SpecMinHeaderConfidence > 100 {
		return fmt.Errorf("SPEC_MIN_HEADER_CONFIDENCE must be in range 0..100")
	}
	if cfg.SpecMaxRowLossRatio < 0 || cfg.SpecMaxRowLossRatio > 1 {
		return fmt.Errorf("SPEC_MAX_ROW_LOSS_RATIO must be in range 0..1")
	}
	if cfg.AIPreviewMaxTokens <= 0 || cfg.AIConvertMaxTokens <= 0 || cfg.AISuggestMaxTokens <= 0 {
		return fmt.Errorf("AI_*_MAX_TOKENS values must be positive")
	}
	if cfg.AISuggestTimeout <= 0 {
		return fmt.Errorf("AI_SUGGEST_TIMEOUT must be positive")
	}
	if len(cfg.TrustedProxies) == 0 {
		return fmt.Errorf("TRUSTED_PROXIES must have at least one entry")
	}
	for _, proxy := range cfg.TrustedProxies {
		if proxy == "" {
			return fmt.Errorf("TRUSTED_PROXIES must not contain empty entries")
		}
		if net.ParseIP(proxy) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(proxy); err == nil {
			continue
		}
		return fmt.Errorf("TRUSTED_PROXIES entry %q must be a valid IP or CIDR", proxy)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt64(key string, fallback int64) int64 {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvFloat64(key string, fallback float64) float64 {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	var items []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}
