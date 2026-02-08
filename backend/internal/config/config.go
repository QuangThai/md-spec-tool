package config

import (
	"log/slog"
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
	DefaultMaxUploadBytes = 10 << 20 // 10MB
	DefaultMaxPasteBytes  = 1 << 20  // 1MB

	// HTTP client timeout for external calls (Google Sheets, etc.)
	DefaultHTTPClientTimeout = 30 * time.Second

	// Rate limiting defaults
	DefaultShareCreateRateLimit  = 10
	DefaultShareUpdateRateLimit  = 20
	DefaultShareCommentRateLimit = 20
	DefaultRateLimitWindow       = time.Minute

	// AI defaults
	DefaultAIRequestTimeout = 30 * time.Second
	DefaultAIMaxRetries     = 3
	DefaultAICacheTTL       = 1 * time.Hour
	DefaultAIMaxCacheSize   = 1000
	DefaultAIRetryBaseDelay = 1 * time.Second

	// AI preview defaults (reduced for fast response when skip_ai=false)
	DefaultAIPreviewTimeout    = 10 * time.Second
	DefaultAIPreviewMaxRetries = 1
)

type Config struct {
	// Server
	Host        string
	Port        string
	CORSOrigins []string

	// Upload/Paste limits
	MaxUploadBytes int64
	MaxPasteBytes  int64

	// HTTP client
	HTTPClientTimeout time.Duration

	// Rate limiting
	ShareCreateRateLimit  int
	ShareUpdateRateLimit  int
	ShareCommentRateLimit int
	RateLimitWindow       time.Duration

	// AI configuration
	OpenAIAPIKey     string
	OpenAIModel      string
	AIEnabled        bool // Auto-enabled when OPENAI_API_KEY is set
	AIRequestTimeout time.Duration
	AIMaxRetries     int
	AICacheTTL       time.Duration
	AIMaxCacheSize   int
	AIRetryBaseDelay time.Duration

	// AI preview configuration (reduced timeout/retries for when skip_ai=false on preview)
	AIPreviewTimeout    time.Duration
	AIPreviewMaxRetries int

	// Feature flags
	UseNewConverterPipeline bool

	// Storage
	ShareStorePath string
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
		MaxUploadBytes: getEnvInt64("MAX_UPLOAD_BYTES", DefaultMaxUploadBytes),
		MaxPasteBytes:  getEnvInt64("MAX_PASTE_BYTES", DefaultMaxPasteBytes),

		// HTTP client
		HTTPClientTimeout: getEnvDuration("HTTP_CLIENT_TIMEOUT", DefaultHTTPClientTimeout),

		// Rate limiting
		ShareCreateRateLimit:  getEnvInt("SHARE_CREATE_RATE_LIMIT", DefaultShareCreateRateLimit),
		ShareUpdateRateLimit:  getEnvInt("SHARE_UPDATE_RATE_LIMIT", DefaultShareUpdateRateLimit),
		ShareCommentRateLimit: getEnvInt("SHARE_COMMENT_RATE_LIMIT", DefaultShareCommentRateLimit),
		RateLimitWindow:       getEnvDuration("RATE_LIMIT_WINDOW", DefaultRateLimitWindow),

		// AI configuration
		OpenAIAPIKey:     openAIAPIKey,
		OpenAIModel:      getEnv("OPENAI_MODEL", DefaultOpenAIModel),
		AIEnabled:        aiEnabled,
		AIRequestTimeout: getEnvDuration("AI_REQUEST_TIMEOUT", DefaultAIRequestTimeout),
		AIMaxRetries:     getEnvInt("AI_MAX_RETRIES", DefaultAIMaxRetries),
		AICacheTTL:       getEnvDuration("AI_CACHE_TTL", DefaultAICacheTTL),
		AIMaxCacheSize:   getEnvInt("AI_MAX_CACHE_SIZE", DefaultAIMaxCacheSize),
		AIRetryBaseDelay: getEnvDuration("AI_RETRY_BASE_DELAY", DefaultAIRetryBaseDelay),

		// AI preview configuration
		AIPreviewTimeout:    getEnvDuration("AI_PREVIEW_TIMEOUT", DefaultAIPreviewTimeout),
		AIPreviewMaxRetries: getEnvInt("AI_PREVIEW_MAX_RETRIES", DefaultAIPreviewMaxRetries),

		// Feature flags
		UseNewConverterPipeline: getEnvBool("USE_NEW_CONVERTER_PIPELINE", false),

		// Storage
		ShareStorePath: getEnv("SHARE_STORE_PATH", ""),
	}
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
