package ai

import (
	"fmt"
	"log/slog"
	"regexp"
)

// InjectionResult holds the outcome of an injection detection check.
type InjectionResult struct {
	Detected bool
	Pattern  string // named category, e.g. "instruction_override"
	Matched  string // the sub-string that triggered detection
}

// injectionRule pairs a compiled regexp with its category name.
type injectionRule struct {
	re       *regexp.Regexp
	category string
}

// rules is the ordered set of injection patterns. Compiled once at init time,
// making DetectInjection and SanitizeForPrompt safe for concurrent use.
var rules []injectionRule

func init() {
	defs := []struct {
		pattern  string
		category string
	}{
		// Direct instruction override — must come before role_manipulation to
		// avoid mis-categorising "ignore" sentences as role-play.
		{`(?i)ignore\s+(all\s+)?previous\s+instructions`, "instruction_override"},
		{`(?i)disregard\s+(the\s+)?system\s+prompt`, "instruction_override"},
		{`(?i)forget\s+everything\s+(above|before)`, "instruction_override"},

		// Role / persona manipulation
		{`(?i)(you\s+are\s+now|act\s+as\s+if|pretend\s+you\s+are)`, "role_manipulation"},

		// Delimiter / token escape — raw string to avoid backslash noise.
		{"\"\"\"", "delimiter_escape"},
		{`<\|system\|>`, "delimiter_escape"},
		{`<\|end\|>`, "delimiter_escape"},
		// Triple-dash separator on its own line (common markdown/YAML delimiter abuse).
		{`(?m)^---\s*$`, "delimiter_escape"},

		// Output / response hijacking
		{`(?i)(output|return|respond\s+with)\s+the\s+following`, "output_manipulation"},
		// "Return this exact response:" / "output this exact response:" variants.
		{`(?i)(output|return)\s+this\s+(exact\s+)?(response|output)`, "output_manipulation"},
	}

	rules = make([]injectionRule, 0, len(defs))
	for _, d := range defs {
		rules = append(rules, injectionRule{
			re:       regexp.MustCompile(d.pattern),
			category: d.category,
		})
	}
}

// DetectInjection scans s against known prompt-injection patterns.
// It returns on the first match; callers that need all matches should call
// it iteratively on sub-strings (not required by current consumers).
func DetectInjection(s string) InjectionResult {
	for _, rule := range rules {
		if loc := rule.re.FindString(s); loc != "" {
			return InjectionResult{
				Detected: true,
				Pattern:  rule.category,
				Matched:  loc,
			}
		}
	}
	return InjectionResult{}
}

// SanitizeForPrompt returns s unchanged when no injection is found.
// When injection is detected it wraps the value in a clearly-labelled
// [USER_INPUT: …] bracket so the LLM treats it as opaque data, not as a
// directive.  It also logs a warning so operators can monitor attempts.
func SanitizeForPrompt(s string) string {
	result := DetectInjection(s)
	if !result.Detected {
		return s
	}

	slog.Warn("prompt injection attempt detected",
		"pattern", result.Pattern,
		"matched", result.Matched,
		"input_len", len(s),
	)

	return fmt.Sprintf("[USER_INPUT: %s]", s)
}

// SanitizeHeadersForPrompt applies SanitizeForPrompt to every header in the
// slice and returns a new slice, leaving the original untouched.
func SanitizeHeadersForPrompt(headers []string) []string {
	out := make([]string, len(headers))
	for i, h := range headers {
		out[i] = SanitizeForPrompt(h)
	}
	return out
}
