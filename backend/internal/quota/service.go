package quota

import (
	"context"
	"errors"
	"time"
)

// QuotaStore interface for quota persistence
type QuotaStore interface {
	AddUsage(ctx context.Context, sessionID string, tokens int64) error
	GetUsage(ctx context.Context, sessionID string) (*QuotaUsage, error)
	IsAvailable(ctx context.Context, sessionID string) (bool, error)
	IncrementConversion(ctx context.Context, sessionID string) error
	RecordConversion(ctx context.Context, sessionID string, tokens int64) error
	ResetDaily(ctx context.Context) error
	Cleanup(ctx context.Context) error
}

// QuotaUsage represents quota usage for a session
type QuotaUsage struct {
	SessionID        string
	UserID           string // Track user (new in bead-4)
	TokensUsedToday  int64
	DailyConversions int
	ResetTime        time.Time
	LastUpdated      time.Time
}

// Service provides quota enforcement operations
type Service struct {
	store  QuotaStore
	engine *PolicyEngine
}

// NewService creates a new quota service
func NewService(store QuotaStore) *Service {
	return &Service{
		store:  store,
		engine: NewPolicyEngine(store),
	}
}

// CheckQuota verifies if request is allowed
// Returns (allowed, remaining_tokens, error)
func (s *Service) CheckQuota(ctx context.Context, userID, sessionID string) (bool, int64, error) {
	if sessionID == "" {
		return false, 0, errors.New("session_id required")
	}

	return s.engine.Enforce(ctx, userID, sessionID)
}

// RecordUsage records token usage for a session
func (s *Service) RecordUsage(ctx context.Context, sessionID string, tokens int64) error {
	if sessionID == "" {
		return errors.New("session_id required")
	}

	return s.store.AddUsage(ctx, sessionID, tokens)
}

// RecordConversion increments conversion count
func (s *Service) RecordConversion(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session_id required")
	}

	return s.store.IncrementConversion(ctx, sessionID)
}

// GetUsage returns current quota usage
func (s *Service) GetUsage(ctx context.Context, sessionID string) (*QuotaUsage, error) {
	if sessionID == "" {
		return nil, errors.New("session_id required")
	}

	return s.store.GetUsage(ctx, sessionID)
}

// GetPolicyEngine returns the policy engine for configuration
func (s *Service) GetPolicyEngine() *PolicyEngine {
	return s.engine
}

// GetStore returns the underlying quota store
func (s *Service) GetStore() QuotaStore {
	return s.store
}
