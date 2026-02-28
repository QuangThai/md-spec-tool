package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PromptFile represents a YAML prompt file structure
type PromptFile struct {
	Version         string `yaml:"version"`
	SchemaVersion   string `yaml:"schema_version"`
	OperationID     string `yaml:"operation_id"`
	CreatedAt       string `yaml:"created_at"`
	UpdatedAt       string `yaml:"updated_at"`
	Description     string `yaml:"description"`
	SystemPrompt    string `yaml:"system_prompt"`
	BreakingChange  bool   `yaml:"breaking_change"`
	Notes           string `yaml:"notes"`
}

// LoadPromptFile loads a YAML prompt file from disk
func LoadPromptFile(filePath string) (*PromptFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file %s: %w", filePath, err)
	}

	var pf PromptFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("failed to parse prompt file %s: %w", filePath, err)
	}

	// Validate required fields
	if pf.Version == "" {
		return nil, fmt.Errorf("prompt file %s missing required field: version", filePath)
	}
	if pf.OperationID == "" {
		return nil, fmt.Errorf("prompt file %s missing required field: operation_id", filePath)
	}
	if pf.SystemPrompt == "" {
		return nil, fmt.Errorf("prompt file %s missing required field: system_prompt", filePath)
	}

	return &pf, nil
}

// LoadPromptsFromDirectory loads all YAML prompt files from a directory
// and registers them with the provided registry.
func LoadPromptsFromDirectory(dir string, registry *PromptRegistry) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read prompts directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		pf, err := LoadPromptFile(filePath)
		if err != nil {
			return fmt.Errorf("error loading prompt file %s: %w", filePath, err)
		}

		// Register the prompt
		registry.Register(PromptEntry{
			ID:      pf.OperationID,
			Version: pf.Version,
			Content: pf.SystemPrompt,
		})
	}

	return nil
}

// NewRegistryFromFiles creates a new PromptRegistry and loads all YAML files
// from the specified directory (default: internal/ai/prompts)
func NewRegistryFromFiles(dir string) (*PromptRegistry, error) {
	registry := NewPromptRegistry()

	// If dir is empty, use default
	if dir == "" {
		// Try to find prompts directory relative to this package
		exePath, err := os.Executable()
		if err == nil {
			baseDir := filepath.Dir(exePath)
			dir = filepath.Join(baseDir, "prompts")
		}
	}

	// Load YAML files if directory exists
	if _, err := os.Stat(dir); err == nil {
		if err := LoadPromptsFromDirectory(dir, registry); err != nil {
			// Log warning but don't fail - fallback to embedded prompts
			// In production, this might be a hard error
		}
	}

	// Fall back to embedded prompts in code
	registry.Register(PromptEntry{
		ID:      PromptIDColumnMapping,
		Version: PromptVersionColumnMapping,
		Content: SystemPromptColumnMapping,
	})
	registry.Register(PromptEntry{
		ID:      PromptIDPasteAnalysis,
		Version: PromptVersionPasteAnalysis,
		Content: SystemPromptPasteAnalysis,
	})

	return registry, nil
}
