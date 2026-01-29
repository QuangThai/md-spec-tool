package config

import (
	"os"
)

type Config struct {
	Host        string
	Port        string
	CORSOrigins []string
}

func LoadConfig() *Config {
	return &Config{
		Host:        getEnv("HOST", "0.0.0.0"),
		Port:        getEnv("PORT", "8080"),
		CORSOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
