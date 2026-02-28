package ai

import (
	"strings"
	"testing"
)

func TestPromptBuilder_BuildColumnMappingPrompt(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	prompt, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain base system prompt content
	if !strings.Contains(prompt.Content, "canonical_name") {
		t.Error("expected prompt to contain canonical_name from base system prompt")
	}

	// Should contain few-shot examples
	if !strings.Contains(prompt.Content, "FEW-SHOT EXAMPLES") {
		t.Error("expected prompt to contain few-shot examples section")
	}

	// Should contain test_case example since schema_hint=test_case
	if !strings.Contains(prompt.Content, "TC ID") {
		t.Error("expected prompt to contain test_case example headers")
	}

	// Should have a hash
	if prompt.Hash == "" {
		t.Error("expected non-empty prompt hash")
	}
}

func TestPromptBuilder_BuildPromptWithLanguageContext(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	prompt, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "",
		Language:    "ja",
		ColumnCount: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include Japanese example
	if !strings.Contains(prompt.Content, "概要") {
		t.Error("expected prompt to include Japanese example when language=ja")
	}
}

func TestPromptBuilder_DifferentContextDifferentHash(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	prompt1, _ := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})
	prompt2, _ := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "api_spec",
		Language:    "en",
		ColumnCount: 6,
	})

	if prompt1.Hash == prompt2.Hash {
		t.Error("different schema hints should produce different hashes")
	}
}

func TestPromptBuilder_SameContextSameHash(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	ctx := PromptContext{SchemaHint: "test_case", Language: "en", ColumnCount: 6}
	prompt1, _ := builder.BuildPrompt(PromptIDColumnMapping, ctx)
	prompt2, _ := builder.BuildPrompt(PromptIDColumnMapping, ctx)

	if prompt1.Hash != prompt2.Hash {
		t.Error("same context should produce same hash")
	}
}

func TestPromptBuilder_BuildNonMappingPrompt(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	prompt, err := builder.BuildPrompt(PromptIDPasteAnalysis, PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain base paste analysis prompt
	if prompt.Content == "" {
		t.Error("expected non-empty prompt content")
	}
	if prompt.Hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestPromptBuilder_UnknownOperation(t *testing.T) {
	registry := NewPromptRegistry()
	exampleStore := NewExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	_, err := builder.BuildPrompt("nonexistent_operation", PromptContext{})
	if err == nil {
		t.Error("expected error for unknown operation")
	}
}

func TestPromptBuilder_CacheVersionIncludesBuiltHash(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	prompt, _ := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})

	cacheVersion := prompt.CacheVersion
	if cacheVersion == "" {
		t.Error("expected non-empty cache version")
	}
	// Cache version should be longer than just a version string
	if len(cacheVersion) < 10 {
		t.Error("cache version should include hash component")
	}
}

func TestPromptBuilder_RefinementContextAppended(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	prompt, err := builder.BuildPrompt(PromptIDRefineMapping, PromptContext{
		SchemaHint:        "test_case",
		Language:          "en",
		ColumnCount:       6,
		RefinementContext: "Fields X and Y were ambiguous",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(prompt.Content, "Fields X and Y were ambiguous") {
		t.Error("expected refinement context to be included in prompt")
	}
}

func TestPromptBuilder_SchemaHintAddedToPrompt(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	prompt, _ := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "api_spec",
		Language:    "en",
		ColumnCount: 6,
	})

	// Schema hint should be mentioned in context section
	if !strings.Contains(prompt.Content, "api_spec") {
		t.Error("expected schema hint to be mentioned in prompt context")
	}
}
