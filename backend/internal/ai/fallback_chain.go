package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// FallbackChain tries providers in order, falling back on transient failures.
// It is safe for concurrent use.
type FallbackChain struct {
	providers []LLMProvider
}

// NewFallbackChain creates a chain. The first provider is primary; the rest are
// fallbacks tried in order when the previous one returns a transient error.
func NewFallbackChain(providers ...LLMProvider) *FallbackChain {
	return &FallbackChain{providers: providers}
}

// Call tries each provider in order until one succeeds or a permanent error occurs.
// It returns ErrAIUnavailable (wrapping the last error) when every provider fails.
func (c *FallbackChain) Call(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	var lastErr error

	for i, provider := range c.providers {
		attempt := i + 1

		resp, err := provider.CallStructured(ctx, req)
		if err == nil {
			resp.Attempts = attempt
			resp.FallbackUsed = i > 0
			return resp, nil
		}

		lastErr = err

		// Permanent errors must not trigger fallback — fail fast.
		var classified *ClassifiedError
		if errors.As(err, &classified) && classified.Category == ErrorCategoryPermanent {
			slog.Warn("fallback_chain_permanent_error",
				"provider", provider.Name(),
				"model", provider.ModelID(),
				"error", err,
			)
			return nil, err
		}

		// Transient error — log and try the next provider.
		slog.Warn("fallback_chain_provider_failed",
			"provider", provider.Name(),
			"model", provider.ModelID(),
			"attempt", attempt,
			"error", err,
		)
	}

	return nil, fmt.Errorf("%w: all %d providers failed, last error: %v",
		ErrAIUnavailable, len(c.providers), lastErr)
}
