package handlers

import (
	"context"
	"sync"
	"time"

	_ "github.com/google/uuid"
)

const DailyTokenLimit int64 = 100000

type QuotaUsage struct {
	SessionID        string
	TokensUsedToday  int64
	DailyConversions int
	ResetTime        time.Time
	LastUpdated      time.Time
}

// QuotaStore interface defines quota storage operations
type QuotaStore interface {
	// AddUsage increments token count for a session
	AddUsage(ctx context.Context, sessionID string, tokens int64) error

	// GetUsage returns current usage stats for a session
	GetUsage(ctx context.Context, sessionID string) (*QuotaUsage, error)

	// IsAvailable returns true if session has remaining quota
	IsAvailable(ctx context.Context, sessionID string) (bool, error)

	// IncrementConversion increments daily conversion count
	IncrementConversion(ctx context.Context, sessionID string) error

	// RecordConversion atomically increments conversion count and adds token usage
	RecordConversion(ctx context.Context, sessionID string, tokens int64) error

	// ResetDaily resets all counts at midnight UTC (runs in background)
	ResetDaily(ctx context.Context) error

	// Cleanup removes old entries (runs periodically)
	Cleanup(ctx context.Context) error
}

// InMemoryQuotaStore is a simple in-memory implementation for MVP
// Thread-safe using sync.RWMutex
type InMemoryQuotaStore struct {
	mu   sync.RWMutex
	data map[string]*QuotaUsage
}

func NewInMemoryQuotaStore() *InMemoryQuotaStore {
	store := &InMemoryQuotaStore{
		data: make(map[string]*QuotaUsage),
	}

	// Start daily reset goroutine
	go store.dailyResetLoop()

	// Start cleanup goroutine
	go store.cleanupLoop()

	return store
}

// AddUsage increments token count for a session
func (s *InMemoryQuotaStore) AddUsage(ctx context.Context, sessionID string, tokens int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	usage, exists := s.data[sessionID]
	if !exists {
		usage = &QuotaUsage{
			SessionID: sessionID,
			ResetTime: s.getNextReset(),
		}
	}

	// Check if reset time has passed
	if time.Now().UTC().After(usage.ResetTime) {
		usage.TokensUsedToday = 0
		usage.DailyConversions = 0
		usage.ResetTime = s.getNextReset()
	}

	usage.TokensUsedToday += tokens
	usage.LastUpdated = time.Now().UTC()
	s.data[sessionID] = usage

	return nil
}

// GetUsage returns current usage stats
func (s *InMemoryQuotaStore) GetUsage(ctx context.Context, sessionID string) (*QuotaUsage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	usage, exists := s.data[sessionID]
	if !exists {
		return &QuotaUsage{
			SessionID: sessionID,
			ResetTime: s.getNextReset(),
		}, nil
	}

	// Check if reset time has passed
	if time.Now().UTC().After(usage.ResetTime) {
		resetUsage := &QuotaUsage{
			SessionID:   sessionID,
			ResetTime:   s.getNextReset(),
			LastUpdated: time.Now().UTC(),
		}
		s.data[sessionID] = resetUsage
		return resetUsage, nil
	}

	return usage, nil
}

// IsAvailable returns true if session has remaining quota
func (s *InMemoryQuotaStore) IsAvailable(ctx context.Context, sessionID string) (bool, error) {
	usage, err := s.GetUsage(ctx, sessionID)
	if err != nil {
		return false, err
	}

	return usage.TokensUsedToday < DailyTokenLimit, nil
}

// IncrementConversion increments daily conversion count
func (s *InMemoryQuotaStore) IncrementConversion(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	usage, exists := s.data[sessionID]
	if !exists {
		usage = &QuotaUsage{
			SessionID: sessionID,
			ResetTime: s.getNextReset(),
		}
	}

	// Check if reset time has passed
	if time.Now().UTC().After(usage.ResetTime) {
		usage.TokensUsedToday = 0
		usage.DailyConversions = 0
		usage.ResetTime = s.getNextReset()
	}

	usage.DailyConversions++
	usage.LastUpdated = time.Now().UTC()
	s.data[sessionID] = usage

	return nil
}

// RecordConversion atomically increments conversion count and adds token usage
func (s *InMemoryQuotaStore) RecordConversion(ctx context.Context, sessionID string, tokens int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	usage, exists := s.data[sessionID]
	if !exists {
		usage = &QuotaUsage{
			SessionID: sessionID,
			ResetTime: s.getNextReset(),
		}
	}

	// Single reset check for both counters
	if time.Now().UTC().After(usage.ResetTime) {
		usage.TokensUsedToday = 0
		usage.DailyConversions = 0
		usage.ResetTime = s.getNextReset()
	}

	usage.DailyConversions++
	usage.TokensUsedToday += tokens
	usage.LastUpdated = time.Now().UTC()
	s.data[sessionID] = usage

	return nil
}

// ResetDaily resets all counts (called by background goroutine)
func (s *InMemoryQuotaStore) ResetDaily(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for sessionID, usage := range s.data {
		if now.After(usage.ResetTime) {
			s.data[sessionID] = &QuotaUsage{
				SessionID:   sessionID,
				ResetTime:   s.getNextReset(),
				LastUpdated: now,
			}
		}
	}

	return nil
}

// Cleanup removes entries older than 7 days
func (s *InMemoryQuotaStore) Cleanup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)

	for sessionID, usage := range s.data {
		if usage.LastUpdated.Before(sevenDaysAgo) {
			delete(s.data, sessionID)
		}
	}

	return nil
}

// Helper functions

func (s *InMemoryQuotaStore) getNextReset() time.Time {
	now := time.Now().UTC()
	return now.AddDate(0, 0, 1).Truncate(24 * time.Hour)
}

// dailyResetLoop runs ResetDaily every minute (check if reset time passed)
func (s *InMemoryQuotaStore) dailyResetLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.ResetDaily(context.Background())
	}
}

// cleanupLoop runs Cleanup every 6 hours
func (s *InMemoryQuotaStore) cleanupLoop() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.Cleanup(context.Background())
	}
}
