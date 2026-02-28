package feedback

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// ErrInvalidRating is returned when a feedback rating is not 1 or 5.
var ErrInvalidRating = errors.New("feedback: rating must be 1 (thumbs down) or 5 (thumbs up)")

// ErrEmptyRequestHash is returned when the request hash is blank.
var ErrEmptyRequestHash = errors.New("feedback: request_hash must not be empty")

// Feedback represents a single feedback entry.
type Feedback struct {
	ID          int64     `json:"id"`
	RequestHash string    `json:"request_hash"`           // SHA256 of the original request
	Rating      int       `json:"rating"`                 // 1 (thumbs down) or 5 (thumbs up)
	Corrections string    `json:"corrections,omitempty"`  // User's corrections/notes
	ColumnFixes string    `json:"column_fixes,omitempty"` // JSON: corrected column mappings
	SessionID   string    `json:"session_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// FeedbackStats contains aggregated feedback statistics.
type FeedbackStats struct {
	TotalCount    int     `json:"total_count"`
	PositiveCount int     `json:"positive_count"`
	NegativeCount int     `json:"negative_count"`
	PositiveRate  float64 `json:"positive_rate"` // 0.0–1.0
	RecentTrend   string  `json:"recent_trend"`  // "improving", "stable", "declining"
}

// StoreInterface is satisfied by *Store and by test doubles.
type StoreInterface interface {
	Submit(f *Feedback) error
	GetStats() (*FeedbackStats, error)
	GetByRequestHash(hash string) ([]Feedback, error)
	Close() error
}

// Store manages feedback persistence in a SQLite database.
type Store struct {
	db *sql.DB
	mu sync.Mutex // serialises writes
}

// NewStore opens (or creates) a SQLite feedback database at dbPath.
// Parent directories are created automatically.
// If dbPath is empty, ":memory:" is used (useful for tests).
func NewStore(dbPath string) (*Store, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}

	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("feedback: create dir %q: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("feedback: open db: %w", err)
	}
	// Single-writer connection keeps WAL-mode safe, mirroring cache_persistent.go.
	db.SetMaxOpenConns(1)

	if err := initFeedbackSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// initFeedbackSchema creates the feedback table and index if they do not exist.
func initFeedbackSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS feedback (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		request_hash TEXT    NOT NULL,
		rating       INTEGER NOT NULL CHECK(rating IN (1, 5)),
		corrections  TEXT    NOT NULL DEFAULT '',
		column_fixes TEXT    NOT NULL DEFAULT '',
		session_id   TEXT    NOT NULL DEFAULT '',
		created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("feedback: create table: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_feedback_request_hash ON feedback(request_hash)`)
	if err != nil {
		return fmt.Errorf("feedback: create index: %w", err)
	}

	return nil
}

// Submit validates and persists a feedback entry.
// It populates f.ID and f.CreatedAt after a successful insert.
func (s *Store) Submit(f *Feedback) error {
	if f.Rating != 1 && f.Rating != 5 {
		return ErrInvalidRating
	}
	if f.RequestHash == "" {
		return ErrEmptyRequestHash
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var id int64
	var createdAt time.Time

	err := s.db.QueryRow(
		`INSERT INTO feedback (request_hash, rating, corrections, column_fixes, session_id)
		 VALUES (?, ?, ?, ?, ?)
		 RETURNING id, created_at`,
		f.RequestHash, f.Rating, f.Corrections, f.ColumnFixes, f.SessionID,
	).Scan(&id, &createdAt)
	if err != nil {
		return fmt.Errorf("feedback: submit: %w", err)
	}

	f.ID = id
	f.CreatedAt = createdAt
	return nil
}

// GetStats returns aggregated statistics across all feedback entries.
func (s *Store) GetStats() (*FeedbackStats, error) {
	var total, positive, negative int
	err := s.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0)
		FROM feedback
	`).Scan(&total, &positive, &negative)
	if err != nil {
		return nil, fmt.Errorf("feedback: get stats: %w", err)
	}

	var positiveRate float64
	if total > 0 {
		positiveRate = float64(positive) / float64(total)
	}

	trend, err := s.computeRecentTrend()
	if err != nil {
		// Non-fatal: log and fall back to "stable"
		slog.Warn("feedback: could not compute trend", "error", err)
		trend = "stable"
	}

	return &FeedbackStats{
		TotalCount:    total,
		PositiveCount: positive,
		NegativeCount: negative,
		PositiveRate:  positiveRate,
		RecentTrend:   trend,
	}, nil
}

// computeRecentTrend compares the positive rate of the most recent 10 entries
// against the previous 10. Returns "improving", "declining", or "stable".
func (s *Store) computeRecentTrend() (string, error) {
	rows, err := s.db.Query(`SELECT rating FROM feedback ORDER BY id DESC LIMIT 20`)
	if err != nil {
		return "stable", fmt.Errorf("feedback: trend query: %w", err)
	}
	defer rows.Close()

	var ratings []int
	for rows.Next() {
		var r int
		if err := rows.Scan(&r); err != nil {
			return "stable", fmt.Errorf("feedback: trend scan: %w", err)
		}
		ratings = append(ratings, r)
	}
	if err := rows.Err(); err != nil {
		return "stable", fmt.Errorf("feedback: trend rows: %w", err)
	}

	return recentTrend(ratings), nil
}

// recentTrend returns "improving", "declining", or "stable" by comparing
// the most-recent half of ratings against the older half.
// ratings must be ordered newest-first.
func recentTrend(ratings []int) string {
	if len(ratings) < 4 {
		return "stable"
	}

	half := len(ratings) / 2
	recent := ratings[:half]
	older := ratings[half:]

	recentRate := positiveRate(recent)
	olderRate := positiveRate(older)

	const threshold = 0.10
	switch {
	case recentRate-olderRate > threshold:
		return "improving"
	case olderRate-recentRate > threshold:
		return "declining"
	default:
		return "stable"
	}
}

func positiveRate(ratings []int) float64 {
	if len(ratings) == 0 {
		return 0
	}
	pos := 0
	for _, r := range ratings {
		if r == 5 {
			pos++
		}
	}
	return float64(pos) / float64(len(ratings))
}

// GetByRequestHash returns all feedback entries for the given request hash.
func (s *Store) GetByRequestHash(hash string) ([]Feedback, error) {
	rows, err := s.db.Query(
		`SELECT id, request_hash, rating, corrections, column_fixes, session_id, created_at
		 FROM feedback WHERE request_hash = ? ORDER BY id DESC`,
		hash,
	)
	if err != nil {
		return nil, fmt.Errorf("feedback: get by hash: %w", err)
	}
	defer rows.Close()

	var results []Feedback
	for rows.Next() {
		var f Feedback
		if err := rows.Scan(&f.ID, &f.RequestHash, &f.Rating, &f.Corrections, &f.ColumnFixes, &f.SessionID, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("feedback: scan by hash: %w", err)
		}
		results = append(results, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("feedback: rows by hash: %w", err)
	}

	return results, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
