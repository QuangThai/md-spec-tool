package ai

// ModelRouterConfig configures model selection thresholds and model names.
type ModelRouterConfig struct {
	SimpleModel     string // Model for simple inputs (default: gpt-4o-mini)
	ComplexModel    string // Model for complex inputs (default: gpt-4o)
	ColumnThreshold int    // Column count threshold for complex (default: 20)
}

// RoutingContext provides input characteristics used to select a model.
type RoutingContext struct {
	ColumnCount int
	Headers     []string
	Language    string
	SchemaHint  string
}

// ModelRouter selects the appropriate model based on input complexity.
type ModelRouter struct {
	config ModelRouterConfig
}

// NewModelRouter creates a ModelRouter with sensible defaults for any zero-value fields.
func NewModelRouter(cfg ModelRouterConfig) *ModelRouter {
	if cfg.SimpleModel == "" {
		cfg.SimpleModel = "gpt-4o-mini"
	}
	if cfg.ComplexModel == "" {
		cfg.ComplexModel = "gpt-4o"
	}
	if cfg.ColumnThreshold <= 0 {
		cfg.ColumnThreshold = 20
	}
	return &ModelRouter{config: cfg}
}

// SelectModel returns the model name appropriate for the given routing context.
//
// Rules (in priority order):
//  1. ColumnCount > threshold  → complex model
//  2. Language is non-empty and not "en" → complex model
//  3. Any header contains a non-ASCII rune → complex model
//  4. Default → simple model
func (r *ModelRouter) SelectModel(ctx RoutingContext) string {
	// Rule 1: High column count → complex
	if ctx.ColumnCount > r.config.ColumnThreshold {
		return r.config.ComplexModel
	}

	// Rule 2: Non-English language tag → complex
	if ctx.Language != "" && ctx.Language != "en" {
		return r.config.ComplexModel
	}

	// Rule 3: Non-ASCII headers detected → complex
	if hasNonASCIIHeaders(ctx.Headers) {
		return r.config.ComplexModel
	}

	// Default: simple model
	return r.config.SimpleModel
}

// hasNonASCIIHeaders reports whether any header contains a rune outside ASCII range (> 127).
func hasNonASCIIHeaders(headers []string) bool {
	for _, h := range headers {
		for _, r := range h {
			if r > 127 {
				return true
			}
		}
	}
	return false
}
