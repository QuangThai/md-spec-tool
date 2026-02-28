package converter

import (
	"context"
	"testing"
	"time"
)

// ─── Helpers ────────────────────────────────────────────────────────────────

// collectStreamEvents runs ConvertPasteStreaming and returns all fired events.
func collectStreamEvents(t *testing.T, conv *Converter, content, template, format string) ([]StreamEvent, *ConvertResponse, error) {
	t.Helper()
	var events []StreamEvent
	callback := func(e StreamEvent) {
		events = append(events, e)
	}
	result, err := conv.ConvertPasteStreaming(context.Background(), content, template, format, callback)
	return events, result, err
}

// progressPhases returns the Phase values from all "progress"/"complete" events in order.
func progressPhases(events []StreamEvent) []string {
	var phases []string
	for _, e := range events {
		if e.Event != "progress" && e.Event != "complete" {
			continue
		}
		if pd, ok := e.Data.(ProgressData); ok {
			phases = append(phases, pd.Phase)
		}
	}
	return phases
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// TestConvertPasteStreaming_TypesAreCorrectlyFormed ensures the data fields
// carry the right concrete types.
func TestConvertPasteStreaming_TypesAreCorrectlyFormed(t *testing.T) {
	conv := NewConverter()
	input := "Feature\tScenario\tExpected\nLogin\tHappy path\tUser is redirected"

	events, result, err := collectStreamEvents(t, conv, input, "spec", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	for _, e := range events {
		if e.Event == "" {
			t.Error("event type must not be empty")
		}
		if e.Data == nil {
			t.Errorf("event %q has nil data", e.Event)
		}
	}

	// The complete event data must be a ProgressData
	for _, e := range events {
		if e.Event == "complete" {
			if _, ok := e.Data.(ProgressData); !ok {
				t.Errorf("complete event data type %T, want ProgressData", e.Data)
			}
		}
	}
}

// TestConvertPasteStreaming_TableInput_PhaseOrder verifies that table input
// produces phases in the expected order: parsing → mapping → rendering → complete.
func TestConvertPasteStreaming_TableInput_PhaseOrder(t *testing.T) {
	conv := NewConverter()
	input := "Feature\tScenario\tExpected\nLogin\tHappy path\tUser is redirected"

	events, _, err := collectStreamEvents(t, conv, input, "spec", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	phases := progressPhases(events)
	want := []string{"parsing", "mapping", "rendering", "complete"}
	if len(phases) != len(want) {
		t.Fatalf("got phases %v, want %v", phases, want)
	}
	for i, p := range want {
		if phases[i] != p {
			t.Errorf("phase[%d] = %q, want %q", i, phases[i], p)
		}
	}
}

// TestConvertPasteStreaming_MarkdownInput_PhaseOrder verifies that markdown
// input skips the mapping phase: parsing → rendering → complete.
func TestConvertPasteStreaming_MarkdownInput_PhaseOrder(t *testing.T) {
	conv := NewConverter()
	// Clear markdown text – recognised as InputTypeMarkdown
	input := "# My Spec\n\n## Feature\n\nSome details here."

	events, _, err := collectStreamEvents(t, conv, input, "spec", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	phases := progressPhases(events)
	want := []string{"parsing", "rendering", "complete"}
	if len(phases) != len(want) {
		t.Fatalf("markdown input phases %v, want %v", phases, want)
	}
	for i, p := range want {
		if phases[i] != p {
			t.Errorf("phase[%d] = %q, want %q", i, phases[i], p)
		}
	}
}

// TestConvertPasteStreaming_ProgressPercentOrder verifies that percent values
// are non-decreasing and the final one is 100.
func TestConvertPasteStreaming_ProgressPercentOrder(t *testing.T) {
	conv := NewConverter()
	input := "Feature\tScenario\nLogin\tHappy path"

	events, _, err := collectStreamEvents(t, conv, input, "spec", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prev := -1
	for _, e := range events {
		if e.Event != "progress" && e.Event != "complete" {
			continue
		}
		pd, ok := e.Data.(ProgressData)
		if !ok {
			continue
		}
		if pd.Percent < prev {
			t.Errorf("percent went backwards: %d after %d", pd.Percent, prev)
		}
		prev = pd.Percent
	}
	if prev != 100 {
		t.Errorf("last percent = %d, want 100", prev)
	}
}

// TestConvertPasteStreaming_EmptyInput_SendsComplete ensures an empty input
// still fires a complete event and returns an empty result.
func TestConvertPasteStreaming_EmptyInput_SendsComplete(t *testing.T) {
	conv := NewConverter()

	events, result, err := collectStreamEvents(t, conv, "", "spec", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for empty input")
	}

	hasComplete := false
	for _, e := range events {
		if e.Event == "complete" {
			hasComplete = true
		}
	}
	if !hasComplete {
		t.Error("expected complete event for empty input")
	}
}

// TestConvertPasteStreaming_ContextCancellation verifies that a cancelled
// context causes ConvertPasteStreaming to return a context error before the
// complete event.
func TestConvertPasteStreaming_ContextCancellation(t *testing.T) {
	conv := NewConverter()
	input := "Feature\tScenario\tExpected\nLogin\tHappy path\tUser is redirected"

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately so the first ctx.Err() check fires.
	cancel()

	var events []StreamEvent
	_, err := conv.ConvertPasteStreaming(ctx, input, "spec", "spec", func(e StreamEvent) {
		events = append(events, e)
	})

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	// Must NOT have fired a "complete" event after cancellation.
	for _, e := range events {
		if e.Event == "complete" {
			t.Error("complete event fired after context cancellation")
		}
	}
}

// TestConvertPasteStreaming_ContextTimeout tests that a tiny timeout causes
// an error (not a hang).
func TestConvertPasteStreaming_ContextTimeout(t *testing.T) {
	conv := NewConverter()
	input := "Feature\tScenario\nLogin\tHappy path"

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	// Let timeout expire.
	time.Sleep(1 * time.Millisecond)

	_, err := conv.ConvertPasteStreaming(ctx, input, "spec", "spec", func(e StreamEvent) {})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestConvertPasteStreaming_InvalidFormat verifies that an unsupported format
// returns an error without panicking.
func TestConvertPasteStreaming_InvalidFormat(t *testing.T) {
	conv := NewConverter()
	input := "Feature\tScenario\nLogin\tHappy path"

	_, err := conv.ConvertPasteStreaming(context.Background(), input, "spec", "invalid_format", func(e StreamEvent) {})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

// TestConvertPasteStreaming_ResultIsNonEmpty verifies that a valid table input
// produces non-empty markdown output.
func TestConvertPasteStreaming_ResultIsNonEmpty(t *testing.T) {
	conv := NewConverter()
	input := "Feature\tScenario\tExpected\nLogin\tHappy path\tUser redirected"

	_, result, err := collectStreamEvents(t, conv, input, "spec", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MDFlow == "" {
		t.Error("expected non-empty MDFlow output")
	}
}
