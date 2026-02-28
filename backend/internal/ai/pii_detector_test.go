package ai

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// ── Email ─────────────────────────────────────────────────────────────────────

func TestDetectPII_Email(t *testing.T) {
	tests := []struct {
		input string
		want  string // expected redaction token in the match
	}{
		{"contact user@example.com for details", "[REDACTED_EMAIL]"},
		{"Send to alice.bob+tag@sub.domain.co.uk", "[REDACTED_EMAIL]"},
		{"noreply@company.io", "[REDACTED_EMAIL]"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			dets := DetectPII(tc.input)
			if len(dets) == 0 {
				t.Fatalf("expected email detection, got none")
			}
			found := false
			for _, d := range dets {
				if d.Type == PIITypeEmail && d.Redacted == tc.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("email not detected in %q, detections: %+v", tc.input, dets)
			}
		})
	}
}

// ── Phone ─────────────────────────────────────────────────────────────────────

func TestDetectPII_Phone(t *testing.T) {
	tests := []string{
		"+1-234-567-8900", // international with country code
		"(123) 456-7890",  // US with parentheses
		"090-1234-5678",   // Japanese mobile (3-4-4)
		"123-456-7890",    // US standard (3-3-4)
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			dets := DetectPII(input)
			found := false
			for _, d := range dets {
				if d.Type == PIITypePhone {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("phone not detected in %q", input)
			}
		})
	}
}

// ── Credit card ───────────────────────────────────────────────────────────────

func TestDetectPII_CreditCard(t *testing.T) {
	t.Run("luhn_valid_with_dashes", func(t *testing.T) {
		// 4111-1111-1111-1111 is the canonical Visa test card (Luhn valid)
		input := "Card: 4111-1111-1111-1111"
		dets := DetectPII(input)
		found := false
		for _, d := range dets {
			if d.Type == PIITypeCreditCard {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("credit card not detected in %q", input)
		}
	})

	t.Run("luhn_valid_spaces", func(t *testing.T) {
		input := "4111 1111 1111 1111"
		dets := DetectPII(input)
		found := false
		for _, d := range dets {
			if d.Type == PIITypeCreditCard {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("credit card not detected in %q", input)
		}
	})

	t.Run("luhn_invalid_not_detected", func(t *testing.T) {
		// Deliberately invalid: last digit changed so Luhn fails
		input := "4111-1111-1111-1112"
		dets := DetectPII(input)
		for _, d := range dets {
			if d.Type == PIITypeCreditCard {
				t.Errorf("Luhn-invalid number should not be detected as credit card")
			}
		}
	})
}

// ── SSN ───────────────────────────────────────────────────────────────────────

func TestDetectPII_SSN(t *testing.T) {
	tests := []string{
		"123-45-6789",
		"SSN: 987-65-4320",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			dets := DetectPII(input)
			found := false
			for _, d := range dets {
				if d.Type == PIITypeSSN {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SSN not detected in %q", input)
			}
		})
	}
}

// ── No PII ────────────────────────────────────────────────────────────────────

func TestDetectPII_NoPII(t *testing.T) {
	clean := []string{
		"hello world",
		"The product has 500 items in stock",
		"Step 3 of 12",
		"前提条件", // Japanese text
		"Test Case Name",
		"Expected Result",
		"version 1.2.3",
		"Score: 100/200",
	}
	for _, input := range clean {
		t.Run(input, func(t *testing.T) {
			dets := DetectPII(input)
			if len(dets) != 0 {
				t.Errorf("false positive on %q: %+v", input, dets)
			}
		})
	}
}

// ── RedactPII replaces all PII ────────────────────────────────────────────────

func TestRedactPII_ReplacesAll(t *testing.T) {
	input := "email: user@example.com, ssn: 123-45-6789"
	out := RedactPII(input)

	if strings.Contains(out, "user@example.com") {
		t.Error("email not redacted")
	}
	if strings.Contains(out, "123-45-6789") {
		t.Error("SSN not redacted")
	}
	if !strings.Contains(out, "[REDACTED_EMAIL]") {
		t.Error("expected [REDACTED_EMAIL] token")
	}
	if !strings.Contains(out, "[REDACTED_SSN]") {
		t.Error("expected [REDACTED_SSN] token")
	}
}

// ── RedactPII preserves non-PII text ─────────────────────────────────────────

func TestRedactPII_PreservesNonPII(t *testing.T) {
	prefix := "contact: "
	suffix := " for info"
	input := prefix + "user@example.com" + suffix
	out := RedactPII(input)

	if !strings.Contains(out, prefix) {
		t.Errorf("prefix %q not preserved in %q", prefix, out)
	}
	if !strings.Contains(out, suffix) {
		t.Errorf("suffix %q not preserved in %q", suffix, out)
	}
	if strings.Contains(out, "user@example.com") {
		t.Error("PII should have been redacted")
	}
}

// ── Multiple PII types in one string ─────────────────────────────────────────

func TestDetectPII_MultiplePIITypes(t *testing.T) {
	input := "Email user@example.com or call (123) 456-7890, SSN 123-45-6789"
	dets := DetectPII(input)

	typesSeen := make(map[PIIType]bool)
	for _, d := range dets {
		typesSeen[d.Type] = true
	}

	for _, want := range []PIIType{PIITypeEmail, PIITypePhone, PIITypeSSN} {
		if !typesSeen[want] {
			t.Errorf("expected PII type %q not detected; got %+v", want, dets)
		}
	}
}

// ── Concurrent safety ─────────────────────────────────────────────────────────

func TestDetectPII_ConcurrentSafety(t *testing.T) {
	input := "user@example.com and 123-45-6789"
	const goroutines = 50

	var wg sync.WaitGroup
	errs := make(chan string, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			dets := DetectPII(input)
			if len(dets) < 2 {
				errs <- fmt.Sprintf("goroutine %d: expected ≥2 detections, got %d", id, len(dets))
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// (no helpers needed — using strings.Contains directly)
