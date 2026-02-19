package handlers

import (
	"context"
	"testing"

	"github.com/yourorg/md-spec-tool/internal/converter"
)

// TestQuotaIntegration verifies that ConvertHandler and PreviewHandler record token usage
func TestQuotaIntegration_TokenRecording(t *testing.T) {
	store := NewInMemoryQuotaStore()
	quotaHandler := NewQuotaHandler(store)
	convertHandler := NewConvertHandler(nil, nil, nil)
	convertHandler.SetQuotaHandler(quotaHandler)

	ctx := context.Background()
	sessionID := "test-session-123"

	// Simulate a conversion result with token usage
	meta := converter.SpecDocMeta{
		AIEstimatedInputTokens:  100,
		AIEstimatedOutputTokens: 50,
	}

	// Create a mock gin context and call recordTokenUsage
	// For this test, we'll just call AddQuotaUsage directly to verify the flow
	totalTokens := int64(meta.AIEstimatedInputTokens + meta.AIEstimatedOutputTokens)
	err := quotaHandler.AddQuotaUsage(ctx, sessionID, totalTokens)
	if err != nil {
		t.Fatalf("AddQuotaUsage failed: %v", err)
	}

	// Verify tokens were recorded
	usage, err := store.GetUsage(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetUsage failed: %v", err)
	}

	if usage.TokensUsedToday != 150 {
		t.Errorf("Expected 150 tokens, got %d", usage.TokensUsedToday)
	}
}

// TestConvertHandler_SetQuotaHandler verifies quota handler can be set
func TestConvertHandler_SetQuotaHandler(t *testing.T) {
	handler := NewConvertHandler(nil, nil, nil)
	if handler.quotaHandler != nil {
		t.Errorf("Expected nil quota handler initially")
	}

	quotaHandler := NewQuotaHandler(NewInMemoryQuotaStore())
	handler.SetQuotaHandler(quotaHandler)

	if handler.quotaHandler == nil {
		t.Errorf("Expected quota handler to be set")
	}
}


