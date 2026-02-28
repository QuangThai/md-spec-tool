package ai

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"
)

// PersistentCacheConfig holds configuration for the SQLite-backed L2 cache.
type PersistentCacheConfig struct {
	// DBPath is the file path for the SQLite database.
	// Defaults to ".cache/ai_cache.db" if empty.
	DBPath string
	// MaxSize is the maximum number of entries to keep. Default 10000.
	MaxSize int
	// TTL is the time-to-live for each entry. Default 24h.
	TTL time.Duration
}

func (c *PersistentCacheConfig) applyDefaults() {
	if c.DBPath == "" {
		c.DBPath = ".cache/ai_cache.db"
	}
	if c.MaxSize <= 0 {
		c.MaxSize = 10000
	}
	if c.TTL <= 0 {
		c.TTL = 24 * time.Hour
	}
}

// PersistentCache is a SQLite-backed L2 cache implementing CacheLayer.
// Values are stored as JSON blobs and returned as json.RawMessage on Get.
// Thread-safe via a Mutex for DB operations and atomic counters for stats.
type PersistentCache struct {
	db     *sql.DB
	config PersistentCacheConfig
	mu     sync.Mutex // serialises writes and evictions
	wg     sync.WaitGroup
	hits   atomic.Int64
	misses atomic.Int64
	closed atomic.Bool // guards against use after Close
}

// NewPersistentCache opens (or creates) the SQLite database at config.DBPath
// and ensures the schema is up to date. It creates parent directories as needed.
func NewPersistentCache(config PersistentCacheConfig) (*PersistentCache, error) {
	config.applyDefaults()

	// Ensure parent directories exist
	dir := filepath.Dir(config.DBPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cache_persistent: create dir %q: %w", dir, err)
	}

	db, err := sql.Open("sqlite", config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("cache_persistent: open db: %w", err)
	}

	// Single-writer connection pool keeps WAL-mode safe
	db.SetMaxOpenConns(1)

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &PersistentCache{db: db, config: config}, nil
}

// initSchema creates the table and index if they don't exist.
func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cache_entries (
			key          TEXT    PRIMARY KEY,
			value        BLOB    NOT NULL,
			expires_at   INTEGER NOT NULL,
			created_at   INTEGER NOT NULL,
			access_count INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("cache_persistent: create table: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_expires ON cache_entries(expires_at)`)
	if err != nil {
		return fmt.Errorf("cache_persistent: create index: %w", err)
	}

	return nil
}

// Get retrieves a value by key. Returns (json.RawMessage, true) on hit,
// or (nil, false) on miss/expiry. Access count is incremented asynchronously.
func (pc *PersistentCache) Get(key string) (interface{}, bool) {
	now := time.Now().UnixMilli()

	var value []byte
	err := pc.db.QueryRow(
		`SELECT value FROM cache_entries WHERE key = ? AND expires_at > ?`,
		key, now,
	).Scan(&value)

	if err != nil {
		// sql.ErrNoRows is the normal miss path; other errors are logged
		if err != sql.ErrNoRows {
			slog.Error("cache_persistent: get query failed", "key", key, "error", err)
		}
		pc.misses.Add(1)
		return nil, false
	}

	pc.hits.Add(1)

	// Increment access_count in background — best effort, non-blocking.
	// Skip silently if the cache has already been closed (e.g. in tests).
	if pc.closed.Load() {
		return json.RawMessage(value), true
	}
	pc.wg.Add(1)
	go func() {
		defer pc.wg.Done()
		pc.mu.Lock()
		defer pc.mu.Unlock()
		if pc.closed.Load() {
			return
		}
		if _, e := pc.db.Exec(
			`UPDATE cache_entries SET access_count = access_count + 1 WHERE key = ?`, key,
		); e != nil {
			slog.Warn("cache_persistent: failed to update access_count", "key", key, "error", e)
		}
	}()

	return json.RawMessage(value), true
}

// Set stores value under key, serialising it to JSON if it is not already
// a json.RawMessage or []byte. Enforces MaxSize via LFU/LRU eviction after insert.
func (pc *PersistentCache) Set(key string, value interface{}) {
	data, err := marshalValue(value)
	if err != nil {
		slog.Error("cache_persistent: failed to marshal value", "key", key, "error", err)
		return
	}

	now := time.Now()
	expiresAt := now.Add(pc.config.TTL).UnixMilli()
	createdAt := now.UnixMilli()

	pc.mu.Lock()
	defer pc.mu.Unlock()

	_, err = pc.db.Exec(
		`INSERT OR REPLACE INTO cache_entries (key, value, expires_at, created_at, access_count)
		 VALUES (?, ?, ?, ?, 0)`,
		key, data, expiresAt, createdAt,
	)
	if err != nil {
		slog.Error("cache_persistent: failed to set", "key", key, "error", err)
		return
	}

	pc.evictLocked()
}

// evictLocked enforces MaxSize. Must be called with pc.mu held.
// Strategy: delete expired first, then least-accessed / oldest entries.
func (pc *PersistentCache) evictLocked() {
	var count int
	if err := pc.db.QueryRow(`SELECT COUNT(*) FROM cache_entries`).Scan(&count); err != nil {
		return
	}
	if count <= pc.config.MaxSize {
		return
	}

	// First pass: remove expired entries
	now := time.Now().UnixMilli()
	if _, err := pc.db.Exec(`DELETE FROM cache_entries WHERE expires_at <= ?`, now); err != nil {
		slog.Warn("cache_persistent: evict expired failed", "error", err)
	}

	// Re-check after expiry purge
	if err := pc.db.QueryRow(`SELECT COUNT(*) FROM cache_entries`).Scan(&count); err != nil {
		return
	}
	if count <= pc.config.MaxSize {
		return
	}

	// Second pass: evict least-accessed, then oldest, until within limit
	excess := count - pc.config.MaxSize
	_, err := pc.db.Exec(
		`DELETE FROM cache_entries
		 WHERE key IN (
		   SELECT key FROM cache_entries
		   ORDER BY access_count ASC, created_at ASC
		   LIMIT ?
		 )`,
		excess,
	)
	if err != nil {
		slog.Warn("cache_persistent: lfu eviction failed", "error", err)
	}
}

// Clear deletes all entries from the cache.
func (pc *PersistentCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if _, err := pc.db.Exec(`DELETE FROM cache_entries`); err != nil {
		slog.Error("cache_persistent: clear failed", "error", err)
	}
}

// Stats returns current cache metrics. Size is read from the DB; hits/misses
// are maintained via atomic counters for zero-contention reads.
func (pc *PersistentCache) Stats() CacheStats {
	var size int
	// No lock needed for a SELECT COUNT(*) on SQLite with a single writer
	if err := pc.db.QueryRow(`SELECT COUNT(*) FROM cache_entries`).Scan(&size); err != nil {
		slog.Warn("cache_persistent: stats count failed", "error", err)
	}

	return CacheStats{
		Hits:    pc.hits.Load(),
		Misses:  pc.misses.Load(),
		Size:    size,
		MaxSize: pc.config.MaxSize,
		Level:   "L2",
	}
}

// Close releases the underlying database connection.
// It sets the closed flag first (preventing new background goroutines from
// starting), waits for all in-flight goroutines to finish, then closes the DB.
func (pc *PersistentCache) Close() error {
	pc.closed.Store(true)
	pc.wg.Wait()
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if pc.db != nil {
		return pc.db.Close()
	}
	return nil
}

// marshalValue converts a value to a JSON byte slice for SQLite storage.
// json.RawMessage and []byte are stored as-is (assumed to be valid JSON).
// All other types are serialised with json.Marshal.
func marshalValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case json.RawMessage:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return json.Marshal(value)
	}
}
