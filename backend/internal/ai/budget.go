package ai

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BudgetAlertLevel indicates severity of budget alert
type BudgetAlertLevel string

const (
	BudgetAlertWarning  BudgetAlertLevel = "warning"
	BudgetAlertHardStop BudgetAlertLevel = "hard_stop"
)

// BudgetAlert is fired when spending crosses a threshold
type BudgetAlert struct {
	Level      BudgetAlertLevel `json:"level"`
	Message    string           `json:"message"`
	Spent      float64          `json:"spent"`
	Budget     float64          `json:"budget"`
	Percentage float64          `json:"percentage"`
	Timestamp  time.Time        `json:"timestamp"`
}

// BudgetConfig configures the budget manager
type BudgetConfig struct {
	DailyBudget       float64           // Max USD per period (0 = unlimited)
	WarningThreshold  float64           // Fraction at which to warn (e.g., 0.80)
	HardStopThreshold float64           // Fraction at which to stop (e.g., 1.00)
	ResetInterval     time.Duration     // How often to reset (default: 24h)
	PersistPath       string            // File path to persist budget state (empty = in-memory only)
	OnAlert           func(BudgetAlert) // Alert callback
}

// DefaultBudgetConfig returns sensible defaults
func DefaultBudgetConfig() BudgetConfig {
	return BudgetConfig{
		DailyBudget:       10.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
		ResetInterval:     24 * time.Hour,
		PersistPath:       ".cache/ai_budget.json",
	}
}

// BudgetStatus is a point-in-time view of budget state
type BudgetStatus struct {
	DailyBudget  float64   `json:"daily_budget"`
	Spent        float64   `json:"spent"`
	Remaining    float64   `json:"remaining"`
	Percentage   float64   `json:"percentage"` // % of budget used
	IsOverBudget bool      `json:"is_over_budget"`
	PeriodStart  time.Time `json:"period_start"`
}

// budgetState is the persisted portion of budget state
type budgetState struct {
	Spent       float64   `json:"spent"`
	PeriodStart time.Time `json:"period_start"`
}

// BudgetManager tracks spending against a daily budget
type BudgetManager struct {
	mu            sync.RWMutex
	config        BudgetConfig
	spent         float64
	periodStart   time.Time
	warningFired  bool
	hardStopFired bool
}

// NewBudgetManager creates a new budget manager, restoring persisted state if available
func NewBudgetManager(cfg BudgetConfig) *BudgetManager {
	if cfg.ResetInterval == 0 {
		cfg.ResetInterval = 24 * time.Hour
	}
	bm := &BudgetManager{
		config:      cfg,
		periodStart: time.Now(),
	}
	bm.loadState()
	return bm
}

// RecordSpend adds spending and checks thresholds
func (bm *BudgetManager) RecordSpend(amount float64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.maybeResetLocked()
	bm.spent += amount
	bm.saveStateLocked()

	if bm.config.DailyBudget <= 0 {
		return // unlimited
	}

	percentage := bm.spent / bm.config.DailyBudget

	// Check hard stop first
	if !bm.hardStopFired && percentage >= bm.config.HardStopThreshold {
		bm.hardStopFired = true
		bm.warningFired = true // skip warning if we jump straight to hard stop
		bm.fireAlert(BudgetAlertHardStop, percentage)
	} else if !bm.warningFired && percentage >= bm.config.WarningThreshold {
		bm.warningFired = true
		bm.fireAlert(BudgetAlertWarning, percentage)
	}
}

// CheckBudget returns whether requests are allowed and remaining budget
func (bm *BudgetManager) CheckBudget() (bool, float64) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bm.config.DailyBudget <= 0 {
		return true, 0 // unlimited
	}

	// If the period has expired, treat budget as replenished
	if time.Since(bm.periodStart) > bm.config.ResetInterval {
		return true, bm.config.DailyBudget
	}

	remaining := bm.config.DailyBudget - bm.spent
	if remaining < 0 {
		remaining = 0
	}

	percentage := bm.spent / bm.config.DailyBudget
	isOverBudget := percentage >= bm.config.HardStopThreshold

	return !isOverBudget, remaining
}

// GetStatus returns the current budget status
func (bm *BudgetManager) GetStatus() BudgetStatus {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	remaining := bm.config.DailyBudget - bm.spent
	if remaining < 0 {
		remaining = 0
	}

	var percentage float64
	if bm.config.DailyBudget > 0 {
		percentage = (bm.spent / bm.config.DailyBudget) * 100
	}

	return BudgetStatus{
		DailyBudget:  bm.config.DailyBudget,
		Spent:        bm.spent,
		Remaining:    remaining,
		Percentage:   percentage,
		IsOverBudget: bm.config.DailyBudget > 0 && bm.spent >= bm.config.DailyBudget*bm.config.HardStopThreshold,
		PeriodStart:  bm.periodStart,
	}
}

// Reset clears spent amount and resets the period
func (bm *BudgetManager) Reset() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.spent = 0
	bm.periodStart = time.Now()
	bm.warningFired = false
	bm.hardStopFired = false
	bm.saveStateLocked()
}

// maybeResetLocked checks if the period has expired and resets. Must be called with lock held.
func (bm *BudgetManager) maybeResetLocked() {
	if time.Since(bm.periodStart) > bm.config.ResetInterval {
		bm.spent = 0
		bm.periodStart = time.Now()
		bm.warningFired = false
		bm.hardStopFired = false
	}
}

// fireAlert sends an alert via callback and logs it. Must be called with write lock held.
func (bm *BudgetManager) fireAlert(level BudgetAlertLevel, percentage float64) {
	alert := BudgetAlert{
		Level:      level,
		Spent:      bm.spent,
		Budget:     bm.config.DailyBudget,
		Percentage: percentage * 100,
		Timestamp:  time.Now(),
	}

	switch level {
	case BudgetAlertWarning:
		alert.Message = "AI spending approaching daily budget limit"
		slog.Warn("budget_warning",
			"spent", bm.spent,
			"budget", bm.config.DailyBudget,
			"percentage", percentage*100,
		)
	case BudgetAlertHardStop:
		alert.Message = "AI spending exceeded daily budget - requests will be rejected"
		slog.Error("budget_hard_stop",
			"spent", bm.spent,
			"budget", bm.config.DailyBudget,
			"percentage", percentage*100,
		)
	}

	if bm.config.OnAlert != nil {
		bm.config.OnAlert(alert)
	}
}

// loadState restores budget state from disk. If the persisted period has expired,
// the state is discarded and the budget starts fresh.
func (bm *BudgetManager) loadState() {
	if bm.config.PersistPath == "" {
		return
	}
	data, err := os.ReadFile(bm.config.PersistPath)
	if err != nil {
		return // file doesn't exist yet — first run
	}
	var state budgetState
	if err := json.Unmarshal(data, &state); err != nil {
		slog.Warn("budget: failed to parse persisted state, starting fresh", "error", err)
		return
	}
	// Discard if the persisted period has already expired
	if time.Since(state.PeriodStart) > bm.config.ResetInterval {
		return
	}
	bm.spent = state.Spent
	bm.periodStart = state.PeriodStart
}

// saveStateLocked persists the current budget state to disk. Must be called with lock held.
func (bm *BudgetManager) saveStateLocked() {
	if bm.config.PersistPath == "" {
		return
	}
	state := budgetState{
		Spent:       bm.spent,
		PeriodStart: bm.periodStart,
	}
	data, err := json.Marshal(state)
	if err != nil {
		slog.Warn("budget: failed to marshal state", "error", err)
		return
	}
	dir := filepath.Dir(bm.config.PersistPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Warn("budget: failed to create dir", "error", err)
		return
	}
	if err := os.WriteFile(bm.config.PersistPath, data, 0o644); err != nil {
		slog.Warn("budget: failed to persist state", "error", err)
	}
}
