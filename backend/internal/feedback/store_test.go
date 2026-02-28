package feedback

import (
	"sync"
	"testing"
	"time"
)

// newTestStore creates an in-memory store for tests.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore("") // "" → :memory:
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestStore_Submit verifies that a valid feedback entry is persisted.
func TestStore_Submit(t *testing.T) {
	s := newTestStore(t)

	f := &Feedback{
		RequestHash: "abc123",
		Rating:      5,
		Corrections: "Looks great",
		SessionID:   "sess-1",
	}

	if err := s.Submit(f); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if f.ID == 0 {
		t.Error("expected non-zero ID after submit")
	}
	if f.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be populated")
	}

	// Verify stored
	entries, err := s.GetByRequestHash("abc123")
	if err != nil {
		t.Fatalf("GetByRequestHash: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Rating != 5 {
		t.Errorf("expected rating 5, got %d", entries[0].Rating)
	}
	if entries[0].SessionID != "sess-1" {
		t.Errorf("unexpected session_id: %s", entries[0].SessionID)
	}
}

// TestStore_GetStats verifies aggregated counts and positive rate.
func TestStore_GetStats(t *testing.T) {
	s := newTestStore(t)

	submit := func(hash string, rating int) {
		t.Helper()
		if err := s.Submit(&Feedback{RequestHash: hash, Rating: rating}); err != nil {
			t.Fatalf("Submit(%s, %d): %v", hash, rating, err)
		}
	}

	// 3 positive, 2 negative
	submit("h1", 5)
	submit("h2", 5)
	submit("h3", 5)
	submit("h4", 1)
	submit("h5", 1)

	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalCount != 5 {
		t.Errorf("TotalCount: want 5, got %d", stats.TotalCount)
	}
	if stats.PositiveCount != 3 {
		t.Errorf("PositiveCount: want 3, got %d", stats.PositiveCount)
	}
	if stats.NegativeCount != 2 {
		t.Errorf("NegativeCount: want 2, got %d", stats.NegativeCount)
	}
	want := 3.0 / 5.0
	if stats.PositiveRate != want {
		t.Errorf("PositiveRate: want %.4f, got %.4f", want, stats.PositiveRate)
	}
}

// TestStore_GetByRequestHash verifies filtering by request hash.
func TestStore_GetByRequestHash(t *testing.T) {
	s := newTestStore(t)

	_ = s.Submit(&Feedback{RequestHash: "target", Rating: 5})
	_ = s.Submit(&Feedback{RequestHash: "target", Rating: 1})
	_ = s.Submit(&Feedback{RequestHash: "other", Rating: 5})

	entries, err := s.GetByRequestHash("target")
	if err != nil {
		t.Fatalf("GetByRequestHash: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for 'target', got %d", len(entries))
	}
	for _, e := range entries {
		if e.RequestHash != "target" {
			t.Errorf("unexpected request_hash: %s", e.RequestHash)
		}
	}

	other, _ := s.GetByRequestHash("other")
	if len(other) != 1 {
		t.Errorf("expected 1 entry for 'other', got %d", len(other))
	}

	empty, _ := s.GetByRequestHash("nonexistent")
	if len(empty) != 0 {
		t.Errorf("expected 0 entries for 'nonexistent', got %d", len(empty))
	}
}

// TestStore_RecentTrend verifies trend detection logic.
func TestStore_RecentTrend(t *testing.T) {
	cases := []struct {
		name    string
		ratings []int // newest-first order
		want    string
	}{
		{"too_few_entries", []int{5, 1, 5}, "stable"},
		{"all_positive_stable", []int{5, 5, 5, 5}, "stable"},
		{"improving", []int{5, 5, 5, 5, 5, 5, 1, 1, 1, 1}, "improving"},
		{"declining", []int{1, 1, 1, 1, 1, 1, 5, 5, 5, 5}, "declining"},
		{"stable_mixed", []int{5, 1, 5, 1, 5, 1, 5, 1}, "stable"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := recentTrend(tc.ratings)
			if got != tc.want {
				t.Errorf("recentTrend(%v) = %q, want %q", tc.ratings, got, tc.want)
			}
		})
	}
}

// TestStore_RecentTrend_Integration submits entries and checks trend via store.
func TestStore_RecentTrend_Integration(t *testing.T) {
	s := newTestStore(t)

	// Submit 10 old (negative) entries and 10 recent (positive) entries.
	// The table inserts in-order so "old" entries have earlier created_at.
	// We rely on auto-increment ordering to simulate time ordering.
	for i := 0; i < 10; i++ {
		_ = s.Submit(&Feedback{RequestHash: "r", Rating: 1})
	}
	for i := 0; i < 10; i++ {
		_ = s.Submit(&Feedback{RequestHash: "r", Rating: 5})
	}

	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.RecentTrend != "improving" {
		t.Errorf("expected 'improving', got %q", stats.RecentTrend)
	}
}

// TestStore_ValidationRejectsInvalidRating verifies that rating != 1 or 5 is rejected.
func TestStore_ValidationRejectsInvalidRating(t *testing.T) {
	s := newTestStore(t)

	invalid := []int{0, 2, 3, 4, 6, -1, 100}
	for _, r := range invalid {
		err := s.Submit(&Feedback{RequestHash: "h", Rating: r})
		if err == nil {
			t.Errorf("expected error for rating=%d, got nil", r)
		}
		if err != ErrInvalidRating {
			t.Errorf("expected ErrInvalidRating for rating=%d, got %v", r, err)
		}
	}

	// Valid ratings must not error
	for _, r := range []int{1, 5} {
		if err := s.Submit(&Feedback{RequestHash: "h", Rating: r}); err != nil {
			t.Errorf("unexpected error for valid rating=%d: %v", r, err)
		}
	}
}

// TestStore_ValidationRejectsEmptyHash verifies that an empty request_hash is rejected.
func TestStore_ValidationRejectsEmptyHash(t *testing.T) {
	s := newTestStore(t)

	err := s.Submit(&Feedback{RequestHash: "", Rating: 5})
	if err != ErrEmptyRequestHash {
		t.Errorf("expected ErrEmptyRequestHash, got %v", err)
	}
}

// TestStore_ConcurrentSafety verifies that parallel writes do not corrupt the store.
func TestStore_ConcurrentSafety(t *testing.T) {
	s := newTestStore(t)

	const goroutines = 20
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			rating := 5
			if i%2 == 0 {
				rating = 1
			}
			if err := s.Submit(&Feedback{RequestHash: "concurrent", Rating: rating}); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent submit error: %v", err)
	}

	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("GetStats after concurrent writes: %v", err)
	}
	if stats.TotalCount != goroutines {
		t.Errorf("expected %d total entries, got %d", goroutines, stats.TotalCount)
	}
}

// TestStore_Close verifies that the store can be closed without error.
func TestStore_Close(t *testing.T) {
	s, err := NewStore("")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	_ = s.Submit(&Feedback{RequestHash: "x", Rating: 5})

	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

// TestStore_EmptyStats verifies behavior when there are no entries.
func TestStore_EmptyStats(t *testing.T) {
	s := newTestStore(t)

	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("GetStats on empty store: %v", err)
	}
	if stats.TotalCount != 0 {
		t.Errorf("expected 0 total, got %d", stats.TotalCount)
	}
	if stats.PositiveRate != 0 {
		t.Errorf("expected 0 positive rate, got %f", stats.PositiveRate)
	}
	if stats.RecentTrend != "stable" {
		t.Errorf("expected 'stable' for empty store, got %q", stats.RecentTrend)
	}
}

// TestPositiveRate_Unit is a pure unit test for the positiveRate helper.
func TestPositiveRate_Unit(t *testing.T) {
	if positiveRate(nil) != 0 {
		t.Error("empty slice should return 0")
	}
	if positiveRate([]int{5, 5, 1}) != 2.0/3.0 {
		t.Error("wrong positive rate")
	}
	if positiveRate([]int{1, 1, 1}) != 0 {
		t.Error("all negative should return 0")
	}
}

// Ensure Store implements StoreInterface at compile time.
var _ StoreInterface = (*Store)(nil)

// Ensure time is imported (used by Feedback.CreatedAt assertions).
var _ = time.Now
