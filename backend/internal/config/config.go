package config

import (
	"os"
)

type Config struct {
	Host         string
	Port         string
	CORSOrigins  []string
	OpenAIAPIKey string
	OpenAIModel  string
	ShareStorePath string
}

func LoadConfig() *Config {
	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080")
	parsedCORSOrigins := splitCSV(corsOrigins)
	if len(parsedCORSOrigins) == 0 {
		parsedCORSOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	}

	return &Config{
		Host:         getEnv("HOST", "0.0.0.0"),
		Port:         getEnv("PORT", "8080"),
		CORSOrigins:  parsedCORSOrigins,
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		ShareStorePath: getEnv("SHARE_STORE_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	var items []string
	current := ""
	for _, ch := range value {
		if ch == ',' {
			item := trimSpaces(current)
			if item != "" {
				items = append(items, item)
			}
			current = ""
			continue
		}
		current += string(ch)
	}
	item := trimSpaces(current)
	if item != "" {
		items = append(items, item)
	}
	return items
}

func trimSpaces(value string) string {
	start := 0
	end := len(value)
	for start < end && (value[start] == ' ' || value[start] == '\n' || value[start] == '\t' || value[start] == '\r') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\n' || value[end-1] == '\t' || value[end-1] == '\r') {
		end--
	}
	return value[start:end]
}
