package quota

import (
	"context"
	"testing"
	"time"
)

func TestReportGenerator_NewReportGenerator(t *testing.T) {
	gen := NewReportGenerator()

	if gen == nil {
		t.Fatal("expected non-nil generator")
	}

	if gen.history == nil {
		t.Fatal("expected non-nil history")
	}
}

func TestReportGenerator_RecordSnapshot(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	usage := &SimpleQuotaUsage{
		SessionID:        "sess_123",
		UserID:           "user_456",
		TokensUsedToday:  50000,
		DailyConversions: 10,
		ResetTime:        time.Now().UTC().Add(24 * time.Hour),
	}

	err := gen.RecordSnapshot(ctx, "sess_123", usage)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	size := gen.GetHistorySize()
	if size != 1 {
		t.Errorf("expected 1 snapshot, got %d", size)
	}
}

func TestReportGenerator_RecordSnapshot_MissingSessionID(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	usage := &SimpleQuotaUsage{
		TokensUsedToday:  50000,
		DailyConversions: 10,
	}

	err := gen.RecordSnapshot(ctx, "", usage)
	if err == nil {
		t.Error("expected error for missing session_id")
	}
}

func TestReportGenerator_RecordSnapshot_MissingUsage(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	err := gen.RecordSnapshot(ctx, "sess_123", nil)
	if err == nil {
		t.Error("expected error for nil usage")
	}
}

func TestReportGenerator_GetDailyReport_NoData(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	reports, err := gen.GetDailyReport(ctx, &UsageReportRequest{Days: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 0 {
		t.Errorf("expected 0 reports, got %d", len(reports))
	}
}

func TestReportGenerator_GetDailyReport_SingleSnapshot(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	usage := &SimpleQuotaUsage{
		SessionID:        "sess_123",
		UserID:           "user_456",
		TokensUsedToday:  50000,
		DailyConversions: 10,
		ResetTime:        time.Now().UTC().Add(24 * time.Hour),
	}

	_ = gen.RecordSnapshot(ctx, "sess_123", usage)

	reports, err := gen.GetDailyReport(ctx, &UsageReportRequest{
		Days:        7,
		AggregateBy: "session",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	report := reports[0]
	if report.SessionID != "sess_123" {
		t.Errorf("session_id: got %q, want %q", report.SessionID, "sess_123")
	}

	if report.TokensUsed != 50000 {
		t.Errorf("tokens_used: got %d, want 50000", report.TokensUsed)
	}

	if report.ConversionsCount != 10 {
		t.Errorf("conversions_count: got %d, want 10", report.ConversionsCount)
	}
}

func TestReportGenerator_GetDailyReport_FilterBySession(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	// Record multiple sessions
	usage1 := &SimpleQuotaUsage{
		SessionID:        "sess_1",
		UserID:           "user_1",
		TokensUsedToday:  50000,
		DailyConversions: 10,
	}

	usage2 := &SimpleQuotaUsage{
		SessionID:        "sess_2",
		UserID:           "user_2",
		TokensUsedToday:  30000,
		DailyConversions: 5,
	}

	_ = gen.RecordSnapshot(ctx, "sess_1", usage1)
	_ = gen.RecordSnapshot(ctx, "sess_2", usage2)

	// Query for specific session
	reports, err := gen.GetDailyReport(ctx, &UsageReportRequest{
		SessionID:   "sess_1",
		Days:        7,
		AggregateBy: "session",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	if reports[0].SessionID != "sess_1" {
		t.Errorf("expected sess_1, got %q", reports[0].SessionID)
	}
}

func TestReportGenerator_GetDailyReport_FilterByUser(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	// Record multiple sessions for same user
	usage1 := &SimpleQuotaUsage{
		SessionID:        "sess_1",
		UserID:           "user_1",
		TokensUsedToday:  50000,
		DailyConversions: 10,
	}

	usage2 := &SimpleQuotaUsage{
		SessionID:        "sess_2",
		UserID:           "user_1",
		TokensUsedToday:  30000,
		DailyConversions: 5,
	}

	usage3 := &SimpleQuotaUsage{
		SessionID:        "sess_3",
		UserID:           "user_2",
		TokensUsedToday:  20000,
		DailyConversions: 3,
	}

	_ = gen.RecordSnapshot(ctx, "sess_1", usage1)
	_ = gen.RecordSnapshot(ctx, "sess_2", usage2)
	_ = gen.RecordSnapshot(ctx, "sess_3", usage3)

	// Query for specific user aggregated
	reports, err := gen.GetDailyReport(ctx, &UsageReportRequest{
		UserID:      "user_1",
		Days:        7,
		AggregateBy: "user",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	// Should be aggregated: 50k + 30k = 80k tokens, 10 + 5 = 15 conversions
	if reports[0].TokensUsed != 80000 {
		t.Errorf("tokens_used: got %d, want 80000", reports[0].TokensUsed)
	}

	if reports[0].ConversionsCount != 15 {
		t.Errorf("conversions_count: got %d, want 15", reports[0].ConversionsCount)
	}
}

func TestReportGenerator_GetDailyReport_AggregateByUser(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	usage1 := &SimpleQuotaUsage{
		SessionID:        "sess_1",
		UserID:           "user_1",
		TokensUsedToday:  25000,
		DailyConversions: 5,
	}

	usage2 := &SimpleQuotaUsage{
		SessionID:        "sess_2",
		UserID:           "user_1",
		TokensUsedToday:  25000,
		DailyConversions: 5,
	}

	_ = gen.RecordSnapshot(ctx, "sess_1", usage1)
	_ = gen.RecordSnapshot(ctx, "sess_2", usage2)

	reports, err := gen.GetDailyReport(ctx, &UsageReportRequest{
		Days:        7,
		AggregateBy: "user",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 aggregated report, got %d", len(reports))
	}

	// Total should be 50k tokens, 10 conversions
	if reports[0].TokensUsed != 50000 {
		t.Errorf("tokens_used: got %d, want 50000", reports[0].TokensUsed)
	}

	if reports[0].ConversionsCount != 10 {
		t.Errorf("conversions_count: got %d, want 10", reports[0].ConversionsCount)
	}
}

func TestReportGenerator_ClearHistory(t *testing.T) {
	gen := NewReportGenerator()
	ctx := context.Background()

	usage := &SimpleQuotaUsage{
		SessionID:        "sess_123",
		UserID:           "user_456",
		TokensUsedToday:  50000,
		DailyConversions: 10,
	}

	_ = gen.RecordSnapshot(ctx, "sess_123", usage)

	if gen.GetHistorySize() != 1 {
		t.Error("expected 1 snapshot before clear")
	}

	gen.ClearHistory()

	if gen.GetHistorySize() != 0 {
		t.Error("expected 0 snapshots after clear")
	}
}
