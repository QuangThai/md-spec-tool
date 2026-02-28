package ai

import (
	"context"
	"testing"
)

func TestLLMRequest_Structure(t *testing.T) {
	req := LLMRequest{
		SystemPrompt: "You are a mapper",
		UserContent:  "Map these headers",
		Schema:       map[string]interface{}{"type": "object"},
		MaxTokens:    1200,
		Temperature:  0,
	}
	if req.SystemPrompt == "" {
		t.Error("expected system prompt")
	}
}

func TestOpenAIProvider_ImplementsInterface(t *testing.T) {
	var _ LLMProvider = (*OpenAIProvider)(nil)
}

func TestOpenAIProvider_Name(t *testing.T) {
	p := &OpenAIProvider{model: "gpt-4o-mini"}
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}

func TestOpenAIProvider_ModelID(t *testing.T) {
	p := &OpenAIProvider{model: "gpt-4o-mini"}
	if p.ModelID() != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini, got %s", p.ModelID())
	}
}

func TestOpenAIProvider_CallStructured_ReturnsModel(t *testing.T) {
	p := &OpenAIProvider{model: "gpt-4o-mini"}
	resp, err := p.CallStructured(context.Background(), LLMRequest{
		SystemPrompt: "test",
		UserContent:  "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", resp.Model)
	}
}

func TestNewOpenAIProvider(t *testing.T) {
	p := NewOpenAIProvider(nil, "gpt-4o")
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.ModelID() != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", p.ModelID())
	}
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}
