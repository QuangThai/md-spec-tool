package ai

// Refactored: All prompt logic has been migrated to separate files:
//   - prompts_policy.go: Security notices and canonical field definitions
//   - prompts_templates.go: Builder functions for system prompts
//   - prompts_versions.go: Version constants and prompt registry
//   - prompts_examples.go: Few-shot examples and test data
//   - prompts_validation_test.go: Comprehensive validation tests
//
// Legacy constants are maintained via init() for backwards compatibility

var (
	// These will be populated by init() below
	SystemPromptColumnMapping = ""
	SystemPromptPasteAnalysis = ""
	SystemPromptSuggestions   = ""
	SystemPromptDiffSummary   = ""
	SystemPromptSemanticValidation = ""
)

func init() {
	// Initialize prompts from builder functions
	SystemPromptColumnMapping = BuildSystemPromptColumnMapping()
	SystemPromptPasteAnalysis = BuildSystemPromptPasteAnalysis()
	SystemPromptSuggestions = BuildSystemPromptSuggestions()
	SystemPromptDiffSummary = BuildSystemPromptDiffSummary()
	SystemPromptSemanticValidation = BuildSystemPromptSemanticValidation()
}
