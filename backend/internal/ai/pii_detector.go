package ai

import (
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// PIIType identifies the category of detected personally identifiable information.
type PIIType string

const (
	PIITypeEmail      PIIType = "email"
	PIITypePhone      PIIType = "phone"
	PIITypeCreditCard PIIType = "credit_card"
	PIITypeSSN        PIIType = "ssn"
)

// PIIDetection describes a single PII finding within an input string.
type PIIDetection struct {
	Type     PIIType
	Start    int    // byte offset, inclusive
	End      int    // byte offset, exclusive
	Redacted string // replacement token, e.g. "[REDACTED_EMAIL]"
}

// piiRule pairs a compiled regexp with its category, redaction token, and an
// optional secondary validation function (e.g. Luhn for credit cards).
// All regexps are compiled once at init time — goroutine-safe.
type piiRule struct {
	piiType  PIIType
	re       *regexp.Regexp
	redacted string
	validate func(s string) bool
}

var piiRules []piiRule

func init() {
	piiRules = []piiRule{
		// Email — specific, run first
		{
			piiType:  PIITypeEmail,
			re:       regexp.MustCompile(`\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}\b`),
			redacted: "[REDACTED_EMAIL]",
		},
		// SSN — strict 3-2-4 digit format, run before phone to prevent partial overlap
		{
			piiType:  PIITypeSSN,
			re:       regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			redacted: "[REDACTED_SSN]",
		},
		// Credit card — 13–19 digits in 4-digit groups with optional separators, Luhn-validated
		{
			piiType:  PIITypeCreditCard,
			re:       regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{1,7}\b`),
			redacted: "[REDACTED_CREDIT_CARD]",
			validate: luhnCheck,
		},
		// Phone — covers international (+1-234-567-8900), US with parens ((123) 456-7890),
		// Japanese mobile (090-1234-5678), and standard US (123-456-7890).
		// Requires explicit separators to avoid matching plain digit sequences.
		{
			piiType: PIITypePhone,
			re: regexp.MustCompile(
				`\+\d{1,3}[-.\s]\d{3}[-.\s]\d{3,4}[-.\s]\d{4}` + // +1-234-567-8900
					`|\(\d{3}\)[-.\s]\d{3}[-.\s]\d{4}` + // (123) 456-7890
					`|\b\d{3}[-.\s]\d{4}[-.\s]\d{4}\b` + // 090-1234-5678 (JP)
					`|\b\d{3}[-.\s]\d{3}[-.\s]\d{4}\b`, // 123-456-7890
			),
			redacted: "[REDACTED_PHONE]",
		},
	}
}

// luhnCheck strips non-digit characters from s and verifies the Luhn checksum.
// Returns false if the digit count is outside the valid 13–19 range.
func luhnCheck(s string) bool {
	digits := make([]int, 0, len(s))
	for _, ch := range s {
		if unicode.IsDigit(ch) {
			digits = append(digits, int(ch-'0'))
		}
	}
	n := len(digits)
	if n < 13 || n > 19 {
		return false
	}

	sum := 0
	for i := n - 1; i >= 0; i-- {
		d := digits[i]
		// Double every second digit counting from the right (0-indexed).
		if (n-1-i)%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

// DetectPII scans input for PII patterns and returns all findings sorted by
// start position. Overlapping matches are resolved greedily: the match with
// the earlier start (or longer length on ties) is kept.
// Safe for concurrent use.
func DetectPII(input string) []PIIDetection {
	type candidate struct {
		start, end int
		det        PIIDetection
	}

	var candidates []candidate

	for _, rule := range piiRules {
		for _, loc := range rule.re.FindAllStringIndex(input, -1) {
			match := input[loc[0]:loc[1]]
			if rule.validate != nil && !rule.validate(match) {
				continue
			}
			candidates = append(candidates, candidate{
				start: loc[0],
				end:   loc[1],
				det: PIIDetection{
					Type:     rule.piiType,
					Start:    loc[0],
					End:      loc[1],
					Redacted: rule.redacted,
				},
			})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort: earlier start first; on ties prefer the longer match.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].start != candidates[j].start {
			return candidates[i].start < candidates[j].start
		}
		return (candidates[i].end - candidates[i].start) > (candidates[j].end - candidates[j].start)
	})

	// Remove overlapping matches (greedy, non-overlapping).
	result := make([]PIIDetection, 0, len(candidates))
	lastEnd := -1
	for _, c := range candidates {
		if c.start >= lastEnd {
			result = append(result, c.det)
			lastEnd = c.end
		}
	}

	slog.Info("PII detected",
		"count", len(result),
		"types", piiTypeNames(result),
		"input_len", len(input),
	)

	return result
}

// RedactPII returns input with every detected PII span replaced by its
// redaction token (e.g. "[REDACTED_EMAIL]"). Non-PII text is preserved as-is.
// Safe for concurrent use.
func RedactPII(input string) string {
	detections := DetectPII(input)
	if len(detections) == 0 {
		return input
	}

	var sb strings.Builder
	prev := 0
	for _, d := range detections {
		sb.WriteString(input[prev:d.Start])
		sb.WriteString(d.Redacted)
		prev = d.End
	}
	sb.WriteString(input[prev:])
	return sb.String()
}

// piiTypeNames returns the unique PII type names present in a detection list.
// Used only for structured logging — never logs the actual PII values.
func piiTypeNames(detections []PIIDetection) []string {
	seen := make(map[PIIType]bool, len(detections))
	names := make([]string, 0, len(detections))
	for _, d := range detections {
		if !seen[d.Type] {
			seen[d.Type] = true
			names = append(names, string(d.Type))
		}
	}
	return names
}
