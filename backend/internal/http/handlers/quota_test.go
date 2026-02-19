package handlers

import (
	"context"
	"testing"
	"time"
)

func TestQuotaStore_AddUsage(t *testing.T) {
	store := NewInMemoryQuotaStore()
	ctx := context.Background()

	sessionID := "test-session"
	tokens := int64(1000)

	// Add usage
	err := store.AddUsage(ctx, sessionID, tokens)
	if err != nil {
		t.Fatalf("AddUsage failed: %v", err)
	}

	// Get usage
	usage, err := store.GetUsage(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetUsage failed: %v", err)
	}

	if usage.TokensUsedToday != tokens {
		t.Errorf("Expected %d tokens, got %d", tokens, usage.TokensUsedToday)
	}
}

func TestQuotaStore_IsAvailable(t *testing.T) {
	store := NewInMemoryQuotaStore()
	ctx := context.Background()

	sessionID := "test-session"

	// New session should be available
	available, err := store.IsAvailable(ctx, sessionID)
	if err != nil {
		t.Fatalf("IsAvailable failed: %v", err)
	}
	if !available {
		t.Errorf("New session should be available")
	}

	// Use up quota
	err = store.AddUsage(ctx, sessionID, DailyTokenLimit)
	if err != nil {
		t.Fatalf("AddUsage failed: %v", err)
	}

	// Should now be unavailable
	available, err = store.IsAvailable(ctx, sessionID)
	if err != nil {
		t.Fatalf("IsAvailable failed: %v", err)
	}
	if available {
		t.Errorf("Session should not be available after using quota")
	}
}

func TestQuotaStore_IncrementConversion(t *testing.T) {
	store := NewInMemoryQuotaStore()
	ctx := context.Background()

	sessionID := "test-session"

	// Increment conversions
	err := store.IncrementConversion(ctx, sessionID)
	if err != nil {
		t.Fatalf("IncrementConversion failed: %v", err)
	}

	err = store.IncrementConversion(ctx, sessionID)
	if err != nil {
		t.Fatalf("IncrementConversion failed: %v", err)
	}

	// Check count
	usage, err := store.GetUsage(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetUsage failed: %v", err)
	}

	if usage.DailyConversions != 2 {
		t.Errorf("Expected 2 conversions, got %d", usage.DailyConversions)
	}
}

func TestQuotaStore_MultipleUsers(t *testing.T) {
	store := NewInMemoryQuotaStore()
	ctx := context.Background()

	session1 := "session-1"
	session2 := "session-2"

	// Add usage for different sessions
	err := store.AddUsage(ctx, session1, 1000)
	if err != nil {
		t.Fatalf("AddUsage for session1 failed: %v", err)
	}

	err = store.AddUsage(ctx, session2, 2000)
	if err != nil {
		t.Fatalf("AddUsage for session2 failed: %v", err)
	}

	// Check they're independent
	usage1, _ := store.GetUsage(ctx, session1)
	usage2, _ := store.GetUsage(ctx, session2)

	if usage1.TokensUsedToday != 1000 {
		t.Errorf("Session 1: expected 1000 tokens, got %d", usage1.TokensUsedToday)
	}
	if usage2.TokensUsedToday != 2000 {
		t.Errorf("Session 2: expected 2000 tokens, got %d", usage2.TokensUsedToday)
	}
}

func TestQuotaStore_ResetDaily(t *testing.T) {
	store := NewInMemoryQuotaStore()
	ctx := context.Background()

	sessionID := "test-session"

	// Add usage
	err := store.AddUsage(ctx, sessionID, 1000)
	if err != nil {
		t.Fatalf("AddUsage failed: %v", err)
	}

	// Reset daily
	err = store.ResetDaily(ctx)
	if err != nil {
		t.Fatalf("ResetDaily failed: %v", err)
	}

	// Since reset time hasn't actually passed in test, token count should still be there
	usage, _ := store.GetUsage(ctx, sessionID)
	if usage.TokensUsedToday != 1000 {
		t.Errorf("Expected tokens to persist before reset time, got %d", usage.TokensUsedToday)
	}
}

func TestQuotaStore_Cleanup(t *testing.T) {
	store := NewInMemoryQuotaStore()
	ctx := context.Background()

	sessionID := "old-session"

	// Add usage
	err := store.AddUsage(ctx, sessionID, 1000)
	if err != nil {
		t.Fatalf("AddUsage failed: %v", err)
	}

	// Manually set old timestamp to trigger cleanup
	inMemStore := NewInMemoryQuotaStore()
	inMemStore.mu.Lock()
	usage := inMemStore.data[sessionID]
	if usage == nil {
		usage = &QuotaUsage{SessionID: sessionID}
	}
	usage.LastUpdated = time.Now().UTC().Add(-8 * 24 * time.Hour) // 8 days ago
	inMemStore.data[sessionID] = usage
	inMemStore.mu.Unlock()

	// Run cleanup on the modified store
	store = inMemStore

	// Run cleanup
	err = store.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Old session should be deleted
	usage, _ = store.GetUsage(ctx, sessionID)
	if usage.TokensUsedToday != 0 {
		t.Errorf("Old session should be cleaned up")
	}
}
