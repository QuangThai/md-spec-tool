package converter

import "testing"

func TestDetectLikelyDelimiter_Semicolon(t *testing.T) {
	lines := []string{
		"id;name;status",
		"1;Login;open",
		"2;Checkout;done",
	}

	if got := detectLikelyDelimiter(lines); got != ';' {
		t.Fatalf("expected semicolon delimiter, got %q", string(got))
	}
}

func TestParseSimple_UsesDetectedDelimiter(t *testing.T) {
	parser := NewPasteParser()
	matrix, err := parser.parseSimple("id;name;status\n1;Login;open")
	if err != nil {
		t.Fatalf("parseSimple returned error: %v", err)
	}

	if matrix.RowCount() != 2 {
		t.Fatalf("expected 2 rows, got %d", matrix.RowCount())
	}
	if matrix.ColCount() != 3 {
		t.Fatalf("expected 3 columns, got %d", matrix.ColCount())
	}
}
