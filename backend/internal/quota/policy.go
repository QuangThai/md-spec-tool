package quota

import (
	"context"
	"fmt"
	"time"
)

// PolicyType defines the quota enforcement strategy
type PolicyType string

const (
	PolicyTypeUser    PolicyType = "user"    // Per-user daily limit
	PolicyTypeSession PolicyType = "session" // Per-session daily limit
	PolicyTypeGlobal  PolicyType = "global"  // Global rate limit (future)
)

// QuotaPolicy defines enforcement rules for quota checking
type QuotaPolicy struct {
	Type           PolicyType
	DailyTokens    int64         // Daily token limit
	DailyRequests  int           // Daily request count limit
	HourlyTokens   int64         // Hourly token limit (optional)
	WindowSize     time.Duration // Time window for enforcement
	ResetTime      time.Time     // When quota resets (UTC midnight)
	GracePeriod    time.Duration // Grace period before enforcement
	Enabled        bool          // Is this policy active
}

// PolicyEngine enforces quota policies
type PolicyEngine struct {
	userPolicy    *QuotaPolicy
	sessionPolicy *QuotaPolicy
	store         QuotaStore
}

// NewPolicyEngine creates a new quota policy engine
func NewPolicyEngine(store QuotaStore) *PolicyEngine {
	return &PolicyEngine{
		userPolicy: &QuotaPolicy{
			Type:          PolicyTypeUser,
			DailyTokens:   1000000, // 1M tokens/day per user
			DailyRequests: 10000,   // 10k requests/day per user
			WindowSize:    24 * time.Hour,
			Enabled:       true,
		},
		sessionPolicy: &QuotaPolicy{
			Type:          PolicyTypeSession,
			DailyTokens:   100000, // 100k tokens/day per session
			DailyRequests: 1000,   // 1k requests/day per session
			WindowSize:    24 * time.Hour,
			Enabled:       true,
		},
		store: store,
	}
}

// Enforce checks if request is allowed under all policies
// Returns (allowed, remaining_tokens, error)
func (e *PolicyEngine) Enforce(ctx context.Context, userID, sessionID string) (bool, int64, error) {
	if sessionID == "" {
		return false, 0, fmt.Errorf("session_id required")
	}

	// Check user policy
	if e.userPolicy.Enabled && userID != "" {
		allowed, remaining, err := e.enforcePolicy(ctx, userID, e.userPolicy)
		if err != nil {
			return false, 0, err
		}
		if !allowed {
			return false, remaining, nil
		}
	}

	// Check session policy
	if e.sessionPolicy.Enabled {
		allowed, remaining, err := e.enforcePolicy(ctx, sessionID, e.sessionPolicy)
		if err != nil {
			return false, 0, err
		}
		if !allowed {
			return false, remaining, nil
		}
	}

	// Get remaining quota from session policy
	usage, err := e.store.GetUsage(ctx, sessionID)
	if err != nil {
		return false, 0, err
	}

	remaining := e.sessionPolicy.DailyTokens - usage.TokensUsedToday
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, nil
}

// enforcePolicy checks quota for a specific policy
// For user policies, 'id' is userID; for session policies, 'id' is sessionID
func (e *PolicyEngine) enforcePolicy(ctx context.Context, id string, policy *QuotaPolicy) (bool, int64, error) {
	// For user policy, we sum across all sessions (simplified: check direct user record)
	// For session policy, we check the session record
	// This is a simplified approach - production may need aggregation
	usage, err := e.store.GetUsage(ctx, id)
	if err != nil {
		return false, 0, err
	}

	remaining := policy.DailyTokens - usage.TokensUsedToday
	if remaining < 0 {
		remaining = 0
	}

	// Allow if tokens available
	allowed := usage.TokensUsedToday < policy.DailyTokens

	return allowed, remaining, nil
}

// SetUserPolicy updates user-level quota policy
func (e *PolicyEngine) SetUserPolicy(policy *QuotaPolicy) error {
	if policy.Type != PolicyTypeUser {
		return fmt.Errorf("invalid policy type for user policy")
	}
	e.userPolicy = policy
	return nil
}

// SetSessionPolicy updates session-level quota policy
func (e *PolicyEngine) SetSessionPolicy(policy *QuotaPolicy) error {
	if policy.Type != PolicyTypeSession {
		return fmt.Errorf("invalid policy type for session policy")
	}
	e.sessionPolicy = policy
	return nil
}

// GetUserPolicy returns current user policy
func (e *PolicyEngine) GetUserPolicy() *QuotaPolicy {
	return e.userPolicy
}

// GetSessionPolicy returns current session policy
func (e *PolicyEngine) GetSessionPolicy() *QuotaPolicy {
	return e.sessionPolicy
}

// DisablePolicy disables enforcement for a specific policy type
func (e *PolicyEngine) DisablePolicy(policyType PolicyType) {
	switch policyType {
	case PolicyTypeUser:
		e.userPolicy.Enabled = false
	case PolicyTypeSession:
		e.sessionPolicy.Enabled = false
	}
}

// EnablePolicy enables enforcement for a specific policy type
func (e *PolicyEngine) EnablePolicy(policyType PolicyType) {
	switch policyType {
	case PolicyTypeUser:
		e.userPolicy.Enabled = true
	case PolicyTypeSession:
		e.sessionPolicy.Enabled = true
	}
}
