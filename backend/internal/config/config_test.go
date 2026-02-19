package config

import (
	"strings"
	"testing"
)

func TestValidateConfigTrustedProxies(t *testing.T) {
	t.Run("accepts valid IP and CIDR", func(t *testing.T) {
		cfg := LoadConfig()
		cfg.TrustedProxies = []string{"127.0.0.1", "::1", "10.0.0.0/8"}

		if err := ValidateConfig(cfg); err != nil {
			t.Fatalf("expected trusted proxies to be valid, got error: %v", err)
		}
	})

	t.Run("rejects invalid trusted proxy entry", func(t *testing.T) {
		cfg := LoadConfig()
		cfg.TrustedProxies = []string{"invalid-proxy-value"}

		err := ValidateConfig(cfg)
		if err == nil {
			t.Fatal("expected validation error for invalid trusted proxy")
		}
		if !strings.Contains(err.Error(), "TRUSTED_PROXIES") {
			t.Fatalf("expected TRUSTED_PROXIES error, got: %v", err)
		}
	})
}
