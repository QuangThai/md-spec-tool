package ai

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
)

// PromptEntry represents a registered prompt with metadata
type PromptEntry struct {
	ID      string // Unique operation identifier: "column_mapping", "paste_analysis", etc.
	Version string // Semantic version: "v1", "v2", "v3"
	Content string // Full prompt text
	Hash    string // SHA256 of Content (computed on register)
}

// CacheVersion returns a version string suitable for cache keys.
// Format: "version:hash_prefix" to ensure cache invalidation on content change.
func (e PromptEntry) CacheVersion() string {
	if e.Hash == "" {
		return e.Version
	}
	// Use first 8 chars of hash for brevity while still ensuring uniqueness
	hashPrefix := e.Hash
	if len(hashPrefix) > 8 {
		hashPrefix = hashPrefix[:8]
	}
	return fmt.Sprintf("%s:%s", e.Version, hashPrefix)
}

// PromptRegistry manages all prompts centrally with versioning and hashing
type PromptRegistry struct {
	mu        sync.RWMutex
	prompts   map[string][]PromptEntry // ID → versions (latest last)
	overrides map[string]string        // ID → forced version
}

// NewPromptRegistry creates a new empty prompt registry
func NewPromptRegistry() *PromptRegistry {
	return &PromptRegistry{
		prompts:   make(map[string][]PromptEntry),
		overrides: make(map[string]string),
	}
}

// Register adds or updates a prompt entry. Hash is computed automatically.
func (r *PromptRegistry) Register(entry PromptEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Compute content hash
	h := sha256.Sum256([]byte(entry.Content))
	entry.Hash = fmt.Sprintf("%x", h[:])

	versions := r.prompts[entry.ID]
	// Replace existing version or append
	found := false
	for i, existing := range versions {
		if existing.Version == entry.Version {
			versions[i] = entry
			found = true
			break
		}
	}
	if !found {
		versions = append(versions, entry)
	}
	r.prompts[entry.ID] = versions
}

// Get retrieves the active prompt entry for an operation.
// If a version override is set, returns that version.
// Otherwise returns the latest registered version.
func (r *PromptRegistry) Get(id string) (PromptEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.prompts[id]
	if !ok || len(versions) == 0 {
		return PromptEntry{}, false
	}

	// Check programmatic override (highest priority)
	if override, hasOverride := r.overrides[id]; hasOverride {
		for _, v := range versions {
			if v.Version == override {
				return v, true
			}
		}
	}

	// Check environment variable override: AI_PROMPT_VERSION_{UPPERCASE_ID}
	envKey := fmt.Sprintf("AI_PROMPT_VERSION_%s", toEnvKey(id))
	if envVersion := os.Getenv(envKey); envVersion != "" {
		for _, v := range versions {
			if v.Version == envVersion {
				return v, true
			}
		}
	}

	// Return latest version (last appended)
	return versions[len(versions)-1], true
}

// SetVersionOverride forces a specific version for an operation.
func (r *PromptRegistry) SetVersionOverride(id, version string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.overrides[id] = version
}

// ClearVersionOverride removes a version override.
func (r *PromptRegistry) ClearVersionOverride(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.overrides, id)
}

// List returns the latest version of every registered prompt.
func (r *PromptRegistry) List() []PromptEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]PromptEntry, 0, len(r.prompts))
	for _, versions := range r.prompts {
		if len(versions) > 0 {
			result = append(result, versions[len(versions)-1])
		}
	}
	return result
}

// toEnvKey converts a prompt ID to an environment variable suffix.
// e.g., "column_mapping" → "COLUMN_MAPPING"
func toEnvKey(id string) string {
	result := make([]byte, len(id))
	for i, c := range id {
		if c >= 'a' && c <= 'z' {
			result[i] = byte(c - 32) // to uppercase
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

// ---- Operation ID constants ----

const (
	PromptIDColumnMapping      = "column_mapping"
	PromptIDRefineMapping      = "refine_mapping"
	PromptIDPasteAnalysis      = "paste_analysis"
	PromptIDSuggestions        = "suggestions"
	PromptIDDiffSummary        = "diff_summary"
	PromptIDSemanticValidation = "semantic_validation"
)

// DefaultPromptRegistry creates a registry pre-loaded with all current prompts.
func DefaultPromptRegistry() *PromptRegistry {
	reg := NewPromptRegistry()

	reg.Register(PromptEntry{
		ID:      PromptIDColumnMapping,
		Version: PromptVersionColumnMapping,
		Content: SystemPromptColumnMapping,
	})
	reg.Register(PromptEntry{
		ID:      PromptIDRefineMapping,
		Version: PromptVersionColumnMapping,
		Content: SystemPromptColumnMapping + "\n\nREFINEMENT CONTEXT:\nThis is a refinement pass on a previous mapping. Be extra critical of low-confidence mappings. If a mapping is truly ambiguous, move it to extra_columns rather than force an incorrect mapping.",
	})
	reg.Register(PromptEntry{
		ID:      PromptIDPasteAnalysis,
		Version: PromptVersionPasteAnalysis,
		Content: SystemPromptPasteAnalysis,
	})
	reg.Register(PromptEntry{
		ID:      PromptIDSuggestions,
		Version: PromptVersionSuggestions,
		Content: SystemPromptSuggestions,
	})
	reg.Register(PromptEntry{
		ID:      PromptIDDiffSummary,
		Version: PromptVersionDiffSummary,
		Content: SystemPromptDiffSummary,
	})
	reg.Register(PromptEntry{
		ID:      PromptIDSemanticValidation,
		Version: PromptVersionSemanticValidation,
		Content: SystemPromptSemanticValidation,
	})

	return reg
}
