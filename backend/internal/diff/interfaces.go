package diff

// Differ interface defines text comparison operations
type Differ interface {
	// Diff computes the differences between two texts and returns a unified diff format
	Diff(before, after string) (string, error)
}

// Implementation note: This interface allows for swapping diff algorithms
// Current implementation uses a map-based approach; future improvements could use:
// - Myers diff algorithm (most accurate)
// - github.com/sergi/go-diff (proven library)
// - github.com/pmezard/go-difflib (Python-like diff)
