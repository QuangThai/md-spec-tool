package ai

import (
	"context"
	"errors"
	"testing"
)

// MockProvider for testing
type mockProvider struct {
	name      string
	model     string
	callFunc  func(ctx context.Context, req LLMRequest) (*LLMResponse, error)
	callCount int
}

func (m *mockProvider) CallStructured(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	m.callCount++
	if m.callFunc != nil {
		return m.callFunc(ctx, req)
	}
	return &LLMResponse{Model: m.model, Content: "{}"}, nil
}
func (m *mockProvider) Name() string    { return m.name }
func (m *mockProvider) ModelID() string { return m.model }

func TestFallbackChain_UsePrimaryOnSuccess(t *testing.T) {
	primary := &mockProvider{name: "primary", model: "gpt-4o-mini"}
	secondary := &mockProvider{name: "secondary", model: "gpt-4o"}

	chain := NewFallbackChain(primary, secondary)
	resp, err := chain.Call(context.Background(), LLMRequest{SystemPrompt: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Model != "gpt-4o-mini" {
		t.Errorf("expected primary model, got %s", resp.Model)
	}
	if primary.callCount != 1 {
		t.Errorf("expected 1 primary call, got %d", primary.callCount)
	}
	if secondary.callCount != 0 {
		t.Errorf("expected 0 secondary calls, got %d", secondary.callCount)
	}
}

func TestFallbackChain_FallsBackOnTransientError(t *testing.T) {
	primary := &mockProvider{
		name:  "primary",
		model: "gpt-4o-mini",
		callFunc: func(_ context.Context, _ LLMRequest) (*LLMResponse, error) {
			return nil, ErrAIUnavailable
		},
	}
	secondary := &mockProvider{name: "secondary", model: "gpt-4o"}

	chain := NewFallbackChain(primary, secondary)
	resp, err := chain.Call(context.Background(), LLMRequest{SystemPrompt: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Model != "gpt-4o" {
		t.Errorf("expected secondary model, got %s", resp.Model)
	}
	if primary.callCount != 1 {
		t.Error("expected primary to be tried")
	}
	if secondary.callCount != 1 {
		t.Error("expected secondary to be called")
	}
}

func TestFallbackChain_AllFailReturnsError(t *testing.T) {
	failing := func(_ context.Context, _ LLMRequest) (*LLMResponse, error) {
		return nil, ErrAIUnavailable
	}
	primary := &mockProvider{name: "primary", model: "m1", callFunc: failing}
	secondary := &mockProvider{name: "secondary", model: "m2", callFunc: failing}

	chain := NewFallbackChain(primary, secondary)
	_, err := chain.Call(context.Background(), LLMRequest{SystemPrompt: "test"})
	if err == nil {
		t.Error("expected error when all providers fail")
	}
	if !errors.Is(err, ErrAIUnavailable) {
		t.Errorf("expected ErrAIUnavailable, got %v", err)
	}
}

func TestFallbackChain_PermanentErrorNoFallback(t *testing.T) {
	permanentErr := errors.New("bad request")
	primary := &mockProvider{
		name:  "primary",
		model: "m1",
		callFunc: func(_ context.Context, _ LLMRequest) (*LLMResponse, error) {
			return nil, &ClassifiedError{
				Category:    ErrorCategoryPermanent,
				ShouldRetry: false,
				Original:    permanentErr,
				Message:     "bad request",
			}
		},
	}
	secondary := &mockProvider{name: "secondary", model: "m2"}

	chain := NewFallbackChain(primary, secondary)
	_, err := chain.Call(context.Background(), LLMRequest{})
	if err == nil {
		t.Error("expected error on permanent failure")
	}
	// Should NOT fall back to secondary for permanent errors
	if secondary.callCount != 0 {
		t.Error("should not fallback for permanent errors")
	}
}

func TestFallbackChain_SingleProviderNoFallback(t *testing.T) {
	primary := &mockProvider{name: "only", model: "m1"}
	chain := NewFallbackChain(primary)

	resp, err := chain.Call(context.Background(), LLMRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Model != "m1" {
		t.Errorf("expected m1, got %s", resp.Model)
	}
}

func TestFallbackChain_TracksAttempts(t *testing.T) {
	failing := func(_ context.Context, _ LLMRequest) (*LLMResponse, error) {
		return nil, ErrAIUnavailable
	}
	primary := &mockProvider{name: "p", model: "m1", callFunc: failing}
	secondary := &mockProvider{name: "s", model: "m2"}

	chain := NewFallbackChain(primary, secondary)
	resp, err := chain.Call(context.Background(), LLMRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", resp.Attempts)
	}
	if resp.FallbackUsed != true {
		t.Error("expected fallback_used=true")
	}
}

func TestFallbackChain_ConcurrentSafety(t *testing.T) {
	primary := &mockProvider{name: "p", model: "m1"}
	secondary := &mockProvider{name: "s", model: "m2"}
	chain := NewFallbackChain(primary, secondary)

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			chain.Call(context.Background(), LLMRequest{}) //nolint:errcheck
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}
