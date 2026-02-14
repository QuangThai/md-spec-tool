package converter_test

import (
	. "github.com/yourorg/md-spec-tool/internal/converter"
	"testing"
)

func TestMDFlowRenderer_GetTemplateNames_CanonicalOnly(t *testing.T) {
	r := NewMDFlowRenderer()
	names := r.GetTemplateNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 template names, got %d", len(names))
	}
	if names[0] != "spec" || names[1] != "table" {
		t.Fatalf("expected [spec table], got %v", names)
	}
}

func TestMDFlowRenderer_Render_RejectsLegacyAliases(t *testing.T) {
	r := NewMDFlowRenderer()
	doc := &SpecDoc{
		Title: "T",
		Rows: []SpecRow{{
			Feature:  "Auth",
			Scenario: "Login",
		}},
	}

	if _, err := r.Render(doc, "spec"); err != nil {
		t.Fatalf("render spec failed: %v", err)
	}
	if _, err := r.Render(doc, "table"); err != nil {
		t.Fatalf("render table failed: %v", err)
	}
	if _, err := r.Render(doc, "default"); err == nil {
		t.Fatalf("expected legacy template alias default to be rejected")
	}
	if _, err := r.Render(doc, "spec-table"); err == nil {
		t.Fatalf("expected legacy template alias spec-table to be rejected")
	}
}
