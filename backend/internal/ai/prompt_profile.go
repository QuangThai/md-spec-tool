package ai

import "strings"

const (
	PromptProfileStaticV3 = "static_v3"
	PromptProfileLegacyV2 = "legacy_v2"
)

func NormalizePromptProfile(profile string) string {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "legacy", "legacy_v2", "v2":
		return PromptProfileLegacyV2
	case "static", "static_v3", "v3":
		return PromptProfileStaticV3
	default:
		return PromptProfileStaticV3
	}
}

func ColumnMappingPromptVersion(profile string) string {
	if NormalizePromptProfile(profile) == PromptProfileLegacyV2 {
		return PromptVersionColumnMappingLegacy
	}
	return PromptVersionColumnMapping
}

func SuggestionsPromptVersion(profile string) string {
	if NormalizePromptProfile(profile) == PromptProfileLegacyV2 {
		return PromptVersionSuggestionsLegacy
	}
	return PromptVersionSuggestions
}
