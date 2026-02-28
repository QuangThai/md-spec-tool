package ai

// Prompt versions - bump when modifying prompts
const (
	PromptVersionColumnMapping       = "v3"
	PromptVersionColumnMappingLegacy = "v2"
	PromptVersionPasteAnalysis       = "v1"
	PromptVersionSuggestions         = "v2"
	PromptVersionSuggestionsLegacy   = "v1"
	PromptVersionDiffSummary         = "v1"
	PromptVersionSemanticValidation  = "v1"
)

// PromptDef ties a prompt name/version to its system message
// Use this registry when constructing prompts to ensure version consistency
type PromptDef struct {
	Name         string
	Version      string
	SystemPrompt string
}

// GetPromptDef returns the prompt definition for a given prompt name
// This ensures version consistency and makes it easy to audit which version is in use
func GetPromptDef(promptName string) *PromptDef {
	defs := map[string]*PromptDef{
		"column_mapping":       {Name: "column_mapping", Version: PromptVersionColumnMapping, SystemPrompt: BuildSystemPromptColumnMapping()},
		"paste_analysis":       {Name: "paste_analysis", Version: PromptVersionPasteAnalysis, SystemPrompt: BuildSystemPromptPasteAnalysis()},
		"suggestions":          {Name: "suggestions", Version: PromptVersionSuggestions, SystemPrompt: BuildSystemPromptSuggestions()},
		"diff_summary":         {Name: "diff_summary", Version: PromptVersionDiffSummary, SystemPrompt: BuildSystemPromptDiffSummary()},
		"semantic_validation":  {Name: "semantic_validation", Version: PromptVersionSemanticValidation, SystemPrompt: BuildSystemPromptSemanticValidation()},
	}
	return defs[promptName]
}
