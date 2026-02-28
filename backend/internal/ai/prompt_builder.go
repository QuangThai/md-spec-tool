package ai

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// PromptContext provides context for dynamic prompt building
type PromptContext struct {
	SchemaHint        string // Expected schema type hint (e.g., "test_case", "api_spec")
	Language          string // Content language (e.g., "en", "ja")
	ColumnCount       int    // Number of columns in the input
	RefinementContext string // Additional refinement instructions to append
}

// BuiltPrompt is the result of building a complete system prompt
type BuiltPrompt struct {
	Content      string // Full assembled prompt text
	Hash         string // SHA256 hex hash of Content (for cache key purposes)
	CacheVersion string // Version string for cache keys: "baseVersion:hash_prefix"
	BaseVersion  string // Version from the registry entry
	OperationID  string // Which operation this prompt is for
}

// PromptBuilder assembles complete system prompts from a registry, an example
// store, and per-request context.  It is the single place responsible for
// combining:
//
//  1. Base system prompt (from PromptRegistry)
//  2. Context hints (schema type, language, column count)
//  3. Refinement instructions
//  4. JSON output schema reminder (for structured-output operations)
//  5. Dynamically selected few-shot examples (from ExampleStore)
type PromptBuilder struct {
	registry     *PromptRegistry
	exampleStore *ExampleStore
}

// NewPromptBuilder creates a new PromptBuilder backed by the given registry
// and example store.
func NewPromptBuilder(registry *PromptRegistry, exampleStore *ExampleStore) *PromptBuilder {
	return &PromptBuilder{
		registry:     registry,
		exampleStore: exampleStore,
	}
}

// BuildPrompt assembles a complete system prompt for operationID using ctx.
// It returns a BuiltPrompt that contains the full text, a SHA-256 hash of that
// text, and a cache-version string suitable for use in cache keys.
//
// Returns an error if operationID is not registered.
func (b *PromptBuilder) BuildPrompt(operationID string, ctx PromptContext) (*BuiltPrompt, error) {
	// 1. Retrieve base prompt from registry.
	entry, ok := b.registry.Get(operationID)
	if !ok {
		return nil, fmt.Errorf("prompt not found for operation: %s", operationID)
	}

	var parts []string

	// 2. Start with base system prompt content.
	parts = append(parts, entry.Content)

	// 3. Append context hints (schema, language, column count).
	if section := b.buildContextSection(ctx); section != "" {
		parts = append(parts, section)
	}

	// 4. Append refinement context when provided.
	if ctx.RefinementContext != "" {
		parts = append(parts, fmt.Sprintf("\nADDITIONAL CONTEXT:\n%s", ctx.RefinementContext))
	}

	// 5. For structured-output operations, inject a brief JSON schema reminder
	//    so the model has an explicit field-name reference (e.g. "canonical_name").
	if hint := b.buildJSONSchemaHint(operationID); hint != "" {
		parts = append(parts, hint)
	}

	// 6. Dynamically select and format the most relevant few-shot examples.
	examples := b.exampleStore.SelectExamples(operationID, SelectionContext{
		SchemaHint:  ctx.SchemaHint,
		Language:    ctx.Language,
		ColumnCount: ctx.ColumnCount,
	})
	if len(examples) > 0 {
		parts = append(parts, "\n"+FormatExamplesForPrompt(examples))
	}

	// 7. Join all parts into the final prompt text.
	content := strings.Join(parts, "\n")

	// 8. Compute SHA-256 hash of the assembled content.
	h := sha256.Sum256([]byte(content))
	hash := fmt.Sprintf("%x", h[:])
	hashPrefix := hash
	if len(hashPrefix) > 8 {
		hashPrefix = hashPrefix[:8]
	}

	return &BuiltPrompt{
		Content:      content,
		Hash:         hash,
		CacheVersion: fmt.Sprintf("%s:%s", entry.Version, hashPrefix),
		BaseVersion:  entry.Version,
		OperationID:  operationID,
	}, nil
}

// buildContextSection returns a CONTEXT HINTS block when any hint is present,
// or an empty string when there is nothing to add.
func (b *PromptBuilder) buildContextSection(ctx PromptContext) string {
	var hints []string

	if ctx.SchemaHint != "" {
		hints = append(hints, fmt.Sprintf("schema_type: %s", ctx.SchemaHint))
	}
	if ctx.Language != "" {
		hints = append(hints, fmt.Sprintf("content_language: %s", ctx.Language))
	}
	if ctx.ColumnCount > 0 {
		hints = append(hints, fmt.Sprintf("column_count: %d", ctx.ColumnCount))
	}

	if len(hints) == 0 {
		return ""
	}

	return fmt.Sprintf("\nCONTEXT HINTS:\n%s", strings.Join(hints, "\n"))
}

// buildJSONSchemaHint returns a brief JSON field-name reminder for operations
// that produce structured JSON output.  For other operations it returns "".
//
// This ensures that field names like "canonical_name" appear in the assembled
// prompt text (which is important for hash-based cache keys and for model
// clarity), without requiring every base prompt to repeat the schema inline.
func (b *PromptBuilder) buildJSONSchemaHint(operationID string) string {
	switch operationID {
	case PromptIDColumnMapping, PromptIDRefineMapping:
		return "\nJSON OUTPUT REMINDER:\n" +
			"Return JSON with a \"canonical_fields\" array where each element has:\n" +
			"  canonical_name, source_header, column_index, confidence, reasoning, alternatives\n" +
			"Also include \"extra_columns\" for unmapped headers and a \"meta\" object."
	default:
		return ""
	}
}
