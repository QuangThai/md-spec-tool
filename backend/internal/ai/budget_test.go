package ai

import (
	"testing"
	"time"
)

func TestBudgetManager_CheckBudget_UnderLimit(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00, // $1.00/day
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})

	ok, remaining := bm.CheckBudget()
	if !ok {
		t.Error("expected budget OK when no spending")
	}
	if remaining != 1.00 {
		t.Errorf("expected $1.00 remaining, got $%.2f", remaining)
	}
}

func TestBudgetManager_RecordSpend(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})

	bm.RecordSpend(0.50)
	ok, remaining := bm.CheckBudget()
	if !ok {
		t.Error("expected budget OK at 50%")
	}
	tolerance := 0.001
	if diff := remaining - 0.50; diff > tolerance || diff < -tolerance {
		t.Errorf("expected $0.50 remaining, got $%.4f", remaining)
	}
}

func TestBudgetManager_WarningThreshold(t *testing.T) {
	alerts := make([]BudgetAlert, 0)
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
		OnAlert: func(alert BudgetAlert) {
			alerts = append(alerts, alert)
		},
	})

	bm.RecordSpend(0.85) // 85% of budget

	if len(alerts) != 1 {
		t.Fatalf("expected 1 warning alert, got %d", len(alerts))
	}
	if alerts[0].Level != BudgetAlertWarning {
		t.Errorf("expected warning alert, got %s", alerts[0].Level)
	}
}

func TestBudgetManager_HardStopThreshold(t *testing.T) {
	alerts := make([]BudgetAlert, 0)
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
		OnAlert: func(alert BudgetAlert) {
			alerts = append(alerts, alert)
		},
	})

	bm.RecordSpend(1.05) // Over budget

	ok, _ := bm.CheckBudget()
	if ok {
		t.Error("expected budget NOT ok when over hard stop")
	}

	hasHardStop := false
	for _, a := range alerts {
		if a.Level == BudgetAlertHardStop {
			hasHardStop = true
		}
	}
	if !hasHardStop {
		t.Error("expected hard stop alert")
	}
}

func TestBudgetManager_RejectWhenOverBudget(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       0.10,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})

	bm.RecordSpend(0.15) // Over budget

	ok, _ := bm.CheckBudget()
	if ok {
		t.Error("should reject when over budget")
	}
}

func TestBudgetManager_GetStatus(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})

	bm.RecordSpend(0.30)
	status := bm.GetStatus()

	if status.DailyBudget != 1.00 {
		t.Errorf("expected daily budget $1.00, got $%.2f", status.DailyBudget)
	}
	if status.Spent != 0.30 {
		t.Errorf("expected $0.30 spent, got $%.4f", status.Spent)
	}
	if status.Remaining != 0.70 {
		t.Errorf("expected $0.70 remaining, got $%.4f", status.Remaining)
	}
	if status.Percentage < 29 || status.Percentage > 31 {
		t.Errorf("expected ~30%% usage, got %.0f%%", status.Percentage)
	}
}

func TestBudgetManager_Reset(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})

	bm.RecordSpend(0.90)
	bm.Reset()

	ok, remaining := bm.CheckBudget()
	if !ok {
		t.Error("expected budget OK after reset")
	}
	if remaining != 1.00 {
		t.Errorf("expected $1.00 remaining after reset, got $%.2f", remaining)
	}
}

func TestBudgetManager_AutoResetAfterPeriod(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
		ResetInterval:     50 * time.Millisecond, // very short for test
	})

	bm.RecordSpend(0.90) // 90%

	// Wait for auto-reset
	time.Sleep(60 * time.Millisecond)

	// After reset, budget should be replenished
	ok, _ := bm.CheckBudget()
	if !ok {
		t.Error("expected budget OK after auto-reset period")
	}
}

func TestBudgetManager_WarningOnlyFiresOnce(t *testing.T) {
	alertCount := 0
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       1.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
		OnAlert: func(alert BudgetAlert) {
			alertCount++
		},
	})

	bm.RecordSpend(0.81) // triggers warning
	bm.RecordSpend(0.01) // should NOT trigger again

	if alertCount != 1 {
		t.Errorf("expected warning to fire only once, got %d alerts", alertCount)
	}
}

func TestBudgetManager_ZeroBudget_AlwaysOK(t *testing.T) {
	// Zero budget means disabled (no limit)
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget: 0,
	})

	bm.RecordSpend(100.00) // huge spend
	ok, _ := bm.CheckBudget()
	if !ok {
		t.Error("zero budget should mean unlimited")
	}
}

func TestBudgetManager_ConcurrentSafety(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       100.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			bm.RecordSpend(0.01)
			bm.CheckBudget()
			bm.GetStatus()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	status := bm.GetStatus()
	tolerance := 0.001
	expectedSpent := 0.50
	if diff := status.Spent - expectedSpent; diff > tolerance || diff < -tolerance {
		t.Errorf("expected $%.2f spent, got $%.4f", expectedSpent, status.Spent)
	}
}

func TestDefaultBudgetConfig(t *testing.T) {
	cfg := DefaultBudgetConfig()
	if cfg.DailyBudget != 10.00 {
		t.Errorf("expected $10.00 default budget, got $%.2f", cfg.DailyBudget)
	}
	if cfg.WarningThreshold != 0.80 {
		t.Errorf("expected 0.80 warning threshold, got %.2f", cfg.WarningThreshold)
	}
	if cfg.HardStopThreshold != 1.00 {
		t.Errorf("expected 1.00 hard stop threshold, got %.2f", cfg.HardStopThreshold)
	}
}
