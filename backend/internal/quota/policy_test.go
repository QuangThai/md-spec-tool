package quota

import (
	"context"
	"sync"
	"testing"
	"time"
)

// MockQuotaStore for testing
type MockQuotaStore struct {
	mu   sync.RWMutex
	data map[string]*QuotaUsage
}

func NewMockQuotaStore() *MockQuotaStore {
	return &MockQuotaStore{
		data: make(map[string]*QuotaUsage),
	}
}

func (m *MockQuotaStore) AddUsage(ctx context.Context, sessionID string, tokens int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage, exists := m.data[sessionID]
	if !exists {
		usage = &QuotaUsage{
			SessionID: sessionID,
			ResetTime: time.Now().UTC().Add(24 * time.Hour),
		}
	}

	usage.TokensUsedToday += tokens
	usage.LastUpdated = time.Now().UTC()
	m.data[sessionID] = usage

	return nil
}

func (m *MockQuotaStore) GetUsage(ctx context.Context, sessionID string) (*QuotaUsage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	usage, exists := m.data[sessionID]
	if !exists {
		return &QuotaUsage{
			SessionID: sessionID,
			ResetTime: time.Now().UTC().Add(24 * time.Hour),
		}, nil
	}

	return usage, nil
}

func (m *MockQuotaStore) IsAvailable(ctx context.Context, sessionID string) (bool, error) {
	usage, _ := m.GetUsage(ctx, sessionID)
	return usage.TokensUsedToday < 100000, nil
}

func (m *MockQuotaStore) IncrementConversion(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage, exists := m.data[sessionID]
	if !exists {
		usage = &QuotaUsage{
			SessionID: sessionID,
			ResetTime: time.Now().UTC().Add(24 * time.Hour),
		}
	}

	usage.DailyConversions++
	m.data[sessionID] = usage

	return nil
}

func (m *MockQuotaStore) RecordConversion(ctx context.Context, sessionID string, tokens int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage, exists := m.data[sessionID]
	if !exists {
		usage = &QuotaUsage{
			SessionID: sessionID,
			ResetTime: time.Now().UTC().Add(24 * time.Hour),
		}
	}

	usage.DailyConversions++
	usage.TokensUsedToday += tokens
	m.data[sessionID] = usage

	return nil
}

func (m *MockQuotaStore) ResetDaily(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for sessionID := range m.data {
		m.data[sessionID] = &QuotaUsage{
			SessionID: sessionID,
			ResetTime: time.Now().UTC().Add(24 * time.Hour),
		}
	}

	return nil
}

func (m *MockQuotaStore) Cleanup(ctx context.Context) error {
	return nil
}

// Tests

func TestPolicyType(t *testing.T) {
	tests := []struct {
		name       string
		policyType PolicyType
		expected   bool
	}{
		{"User Policy", PolicyTypeUser, true},
		{"Session Policy", PolicyTypeSession, true},
		{"Global Policy", PolicyTypeGlobal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.policyType) == 0 != !tt.expected {
				t.Errorf("policyType mismatch: got %q, expected non-empty", tt.policyType)
			}
		})
	}
}

func TestPolicyEngine_NewPolicyEngine(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)

	if engine == nil {
		t.Fatal("expected non-nil engine")
	}

	if engine.userPolicy == nil {
		t.Fatal("expected non-nil user policy")
	}

	if engine.sessionPolicy == nil {
		t.Fatal("expected non-nil session policy")
	}

	if engine.userPolicy.Type != PolicyTypeUser {
		t.Errorf("user policy type: got %q, want %q", engine.userPolicy.Type, PolicyTypeUser)
	}

	if engine.sessionPolicy.Type != PolicyTypeSession {
		t.Errorf("session policy type: got %q, want %q", engine.sessionPolicy.Type, PolicyTypeSession)
	}
}

func TestPolicyEngine_Enforce_SessionPolicy(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)
	ctx := context.Background()

	sessionID := "sess_123"

	// First check should pass (no usage)
	allowed, remaining, err := engine.Enforce(ctx, "", sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !allowed {
		t.Error("expected quota to be available for new session")
	}

	if remaining != engine.sessionPolicy.DailyTokens {
		t.Errorf("remaining tokens: got %d, want %d", remaining, engine.sessionPolicy.DailyTokens)
	}

	// Add usage and check again
	_ = store.AddUsage(ctx, sessionID, 50000)
	allowed, remaining, err = engine.Enforce(ctx, "", sessionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !allowed {
		t.Error("expected quota to be available with 50% usage")
	}

	expectedRemaining := int64(50000)
	if remaining != expectedRemaining {
		t.Errorf("remaining tokens: got %d, want %d", remaining, expectedRemaining)
	}

	// Exceed quota
	_ = store.AddUsage(ctx, sessionID, 50001)
	allowed, remaining, err = engine.Enforce(ctx, "", sessionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Error("expected quota to be exceeded")
	}

	if remaining != 0 {
		t.Errorf("remaining tokens: got %d, want 0", remaining)
	}
}

func TestPolicyEngine_Enforce_UserPolicy(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)
	ctx := context.Background()

	userID := "user_456"
	sessionID := "sess_456"

	// Mock user data with high usage
	store.data[userID] = &QuotaUsage{
		SessionID:       userID,
		TokensUsedToday: 900000, // 90% of 1M user limit
		ResetTime:       time.Now().UTC().Add(24 * time.Hour),
	}

	// Session has no usage
	store.data[sessionID] = &QuotaUsage{
		SessionID:       sessionID,
		TokensUsedToday: 0,
		ResetTime:       time.Now().UTC().Add(24 * time.Hour),
	}

	allowed, _, err := engine.Enforce(ctx, userID, sessionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still be allowed because user hasn't hit 1M limit (900k < 1M)
	if !allowed {
		t.Error("expected quota to be available (900k < 1M user limit)")
	}

	// Now test actual exceeded case
	store.data[userID].TokensUsedToday = 1000001 // Exceed 1M limit
	allowed, _, err = engine.Enforce(ctx, userID, sessionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Error("expected user quota to be exceeded (1M+ > 1M limit)")
	}
}

func TestPolicyEngine_Enforce_MissingSessionID(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)
	ctx := context.Background()

	_, _, err := engine.Enforce(ctx, "user_123", "")

	if err == nil {
		t.Error("expected error for missing session_id")
	}
}

func TestPolicyEngine_SetUserPolicy(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)

	newPolicy := &QuotaPolicy{
		Type:         PolicyTypeUser,
		DailyTokens:  500000,
		DailyRequests: 5000,
	}

	err := engine.SetUserPolicy(newPolicy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine.userPolicy.DailyTokens != 500000 {
		t.Errorf("daily tokens: got %d, want 500000", engine.userPolicy.DailyTokens)
	}
}

func TestPolicyEngine_SetSessionPolicy(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)

	newPolicy := &QuotaPolicy{
		Type:         PolicyTypeSession,
		DailyTokens:  50000,
		DailyRequests: 500,
	}

	err := engine.SetSessionPolicy(newPolicy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine.sessionPolicy.DailyTokens != 50000 {
		t.Errorf("daily tokens: got %d, want 50000", engine.sessionPolicy.DailyTokens)
	}
}

func TestPolicyEngine_SetInvalidPolicy(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)

	// Try to set session policy as user policy
	invalidPolicy := &QuotaPolicy{
		Type:        PolicyTypeSession,
		DailyTokens: 50000,
	}

	err := engine.SetUserPolicy(invalidPolicy)
	if err == nil {
		t.Error("expected error for invalid policy type")
	}
}

func TestPolicyEngine_DisableEnablePolicy(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)

	// Disable user policy
	engine.DisablePolicy(PolicyTypeUser)
	if engine.userPolicy.Enabled {
		t.Error("expected user policy to be disabled")
	}

	// Enable it back
	engine.EnablePolicy(PolicyTypeUser)
	if !engine.userPolicy.Enabled {
		t.Error("expected user policy to be enabled")
	}

	// Disable session policy
	engine.DisablePolicy(PolicyTypeSession)
	if engine.sessionPolicy.Enabled {
		t.Error("expected session policy to be disabled")
	}
}

func TestPolicyEngine_GetPolicies(t *testing.T) {
	store := NewMockQuotaStore()
	engine := NewPolicyEngine(store)

	userPolicy := engine.GetUserPolicy()
	if userPolicy == nil {
		t.Fatal("expected non-nil user policy")
	}

	sessionPolicy := engine.GetSessionPolicy()
	if sessionPolicy == nil {
		t.Fatal("expected non-nil session policy")
	}

	if userPolicy.Type != PolicyTypeUser {
		t.Errorf("user policy type: got %q, want %q", userPolicy.Type, PolicyTypeUser)
	}

	if sessionPolicy.Type != PolicyTypeSession {
		t.Errorf("session policy type: got %q, want %q", sessionPolicy.Type, PolicyTypeSession)
	}
}
