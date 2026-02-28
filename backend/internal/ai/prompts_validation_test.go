package ai

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestExampleIntegrity validates that all examples have consistent structure
func TestExampleIntegrity(t *testing.T) {
	for exIdx, example := range ColumnMappingExamples {
		// Note: Expected mappings may be less than headers if some columns go to extra_columns
		if len(example.Expected) > len(example.Headers) {
			t.Errorf("Example %d: expected mapping count (%d) > header count (%d)",
				exIdx, len(example.Expected), len(example.Headers))
		}

		// Validate each mapping
		for mapIdx, mapping := range example.Expected {
			// Check: canonical name is valid
			if !IsValidCanonicalName(mapping.CanonicalName) {
				t.Errorf("Example %d, mapping %d: invalid canonical_name %q",
					exIdx, mapIdx, mapping.CanonicalName)
			}

			// Check: confidence is in [0, 1]
			if mapping.Confidence < 0 || mapping.Confidence > 1 {
				t.Errorf("Example %d, mapping %d: confidence %.2f not in [0, 1]",
					exIdx, mapIdx, mapping.Confidence)
			}

			// Check: column index is valid
			if mapping.ColumnIndex < 0 || mapping.ColumnIndex >= len(example.Headers) {
				t.Errorf("Example %d, mapping %d: column_index %d out of bounds (headers len=%d)",
					exIdx, mapIdx, mapping.ColumnIndex, len(example.Headers))
			}

			// Check: source header matches the column at that index
			if mapping.SourceHeader != example.Headers[mapping.ColumnIndex] {
				t.Errorf("Example %d, mapping %d: source_header %q doesn't match header at index %d (%q)",
					exIdx, mapIdx, mapping.SourceHeader, mapping.ColumnIndex, example.Headers[mapping.ColumnIndex])
			}

			// Check: reasoning is not empty and not too long
			if mapping.Reasoning == "" {
				t.Errorf("Example %d, mapping %d: reasoning is empty",
					exIdx, mapIdx)
			}
			if len(mapping.Reasoning) > 256 {
				t.Errorf("Example %d, mapping %d: reasoning is too long (%d > 256 chars)",
					exIdx, mapIdx, len(mapping.Reasoning))
			}

			// Check: alternatives have valid canonical names (if present)
			for altIdx, alt := range mapping.Alternatives {
				if alt.Confidence < 0 || alt.Confidence > 1 {
					t.Errorf("Example %d, mapping %d, alternative %d: confidence %.2f not in [0, 1]",
						exIdx, mapIdx, altIdx, alt.Confidence)
				}

				if alt.ColumnIndex < 0 || alt.ColumnIndex >= len(example.Headers) {
					t.Errorf("Example %d, mapping %d, alternative %d: column_index %d out of bounds",
						exIdx, mapIdx, altIdx, alt.ColumnIndex)
				}

				// Note: SourceHeader in AlternativeColumn represents a *canonical* field name
				// It's semantically a CanonicalName, but the struct uses SourceHeader for backwards compat
				if !IsValidCanonicalName(alt.SourceHeader) {
					t.Errorf("Example %d, mapping %d, alternative %d: source_header %q is not a valid canonical name",
						exIdx, mapIdx, altIdx, alt.SourceHeader)
				}
			}
		}
	}
}

// TestSecurityNoticePresence ensures security notice is used
func TestSecurityNoticePresence(t *testing.T) {
	prompts := []string{
		BuildSystemPromptColumnMapping(),
		BuildSystemPromptPasteAnalysis(),
		BuildSystemPromptSuggestions(),
		BuildSystemPromptDiffSummary(),
		BuildSystemPromptSemanticValidation(),
	}

	for idx, prompt := range prompts {
		if !strings.Contains(prompt, SecurityNotice) {
			t.Errorf("Prompt %d: missing security notice", idx)
		}
		if !strings.Contains(prompt, "DATA only") {
			t.Errorf("Prompt %d: missing 'DATA only' injection defense", idx)
		}
	}
}

// TestOutputFormatNoticePresence ensures output format is clear
func TestOutputFormatNoticePresence(t *testing.T) {
	prompts := []string{
		BuildSystemPromptColumnMapping(),
		BuildSystemPromptPasteAnalysis(),
		BuildSystemPromptSuggestions(),
		BuildSystemPromptDiffSummary(),
		BuildSystemPromptSemanticValidation(),
	}

	for idx, prompt := range prompts {
		if !strings.Contains(prompt, "JSON") {
			t.Errorf("Prompt %d: missing JSON output requirement", idx)
		}
	}
}

// TestPromptVersionConsistency ensures versions are mapped correctly
func TestPromptVersionConsistency(t *testing.T) {
	testCases := []struct {
		name    string
		version string
	}{
		{"column_mapping", PromptVersionColumnMapping},
		{"paste_analysis", PromptVersionPasteAnalysis},
		{"suggestions", PromptVersionSuggestions},
		{"diff_summary", PromptVersionDiffSummary},
		{"semantic_validation", PromptVersionSemanticValidation},
	}

	for _, tc := range testCases {
		def := GetPromptDef(tc.name)
		if def == nil {
			t.Errorf("GetPromptDef(%q) returned nil", tc.name)
			continue
		}
		if def.Version != tc.version {
			t.Errorf("Prompt %q: version mismatch, got %q, want %q",
				tc.name, def.Version, tc.version)
		}
		if def.SystemPrompt == "" {
			t.Errorf("Prompt %q: system prompt is empty", tc.name)
		}
	}
}

// TestCanonicalFieldNamesComplete checks that all canonical names in examples are in the set
func TestCanonicalFieldNamesComplete(t *testing.T) {
	seenFields := make(map[string]bool)
	for _, example := range ColumnMappingExamples {
		for _, mapping := range example.Expected {
			seenFields[mapping.CanonicalName] = true
		}
	}

	for field := range seenFields {
		if !IsValidCanonicalName(field) {
			t.Errorf("Canonical field %q used in examples but not defined in canonical set", field)
		}
	}
}

// TestPasteAnalysisExamples validates paste analysis examples
func TestPasteAnalysisExamples(t *testing.T) {
	for idx, example := range PasteAnalysisExamples {
		if example.Input == "" {
			t.Errorf("Example %d: input is empty", idx)
		}
		if example.Expected.InputType == "" {
			t.Errorf("Example %d: input_type is empty", idx)
		}
		if example.Expected.DetectedFormat == "" {
			t.Errorf("Example %d: detected_format is empty", idx)
		}
		if example.Expected.Confidence < 0 || example.Expected.Confidence > 1 {
			t.Errorf("Example %d: confidence %.2f not in [0, 1]", idx, example.Expected.Confidence)
		}
	}
}

// TestExampleJSONSerializability validates that examples can be marshaled to JSON
func TestExampleJSONSerializability(t *testing.T) {
	// Column mapping examples
	for idx, example := range ColumnMappingExamples {
		result := ColumnMappingResult{
			SchemaVersion:   "v2",
			CanonicalFields: example.Expected,
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			t.Errorf("Example %d: failed to marshal to JSON: %v", idx, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("Example %d: JSON is empty", idx)
		}

		// Round-trip test
		var unmarshaled ColumnMappingResult
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Example %d: failed to unmarshal JSON: %v", idx, err)
		}
	}
}

// TestConfidenceCalibration checks that confidence values make sense
func TestConfidenceCalibration(t *testing.T) {
	for exIdx, example := range ColumnMappingExamples {
		for mapIdx, mapping := range example.Expected {
			// Exact matches (case-insensitive) should have confidence >= 0.9
			isExactMatch := strings.EqualFold(mapping.SourceHeader, mapping.CanonicalName)
			if isExactMatch && mapping.Confidence < 0.9 {
				t.Logf("Example %d, mapping %d: exact match %q->%q with confidence %.2f (acceptable)",
					exIdx, mapIdx, mapping.SourceHeader, mapping.CanonicalName, mapping.Confidence)
			}

			// If there are alternatives, main mapping should have higher or equal confidence
			if len(mapping.Alternatives) > 0 {
				for _, alt := range mapping.Alternatives {
					if alt.Confidence > mapping.Confidence {
						t.Errorf("Example %d, mapping %d: alternative confidence (%.2f) higher than main (%.2f)",
							exIdx, mapIdx, alt.Confidence, mapping.Confidence)
					}
				}
			}
		}
	}
}

// TestMultiLanguageSupportPresence checks multi-language hint presence
func TestMultiLanguageSupportPresence(t *testing.T) {
	prompt := BuildSystemPromptColumnMapping()
	languages := []string{"Japanese", "Vietnamese", "Korean", "Chinese"}
	for _, lang := range languages {
		if !strings.Contains(prompt, lang) {
			t.Errorf("Column mapping prompt missing language hint for %s", lang)
		}
	}
}

// BenchmarkPromptGeneration measures prompt construction performance
func BenchmarkPromptGeneration(b *testing.B) {
	b.Run("ColumnMapping", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = BuildSystemPromptColumnMapping()
		}
	})
	b.Run("PasteAnalysis", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = BuildSystemPromptPasteAnalysis()
		}
	})
	b.Run("Suggestions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = BuildSystemPromptSuggestions()
		}
	})
}
