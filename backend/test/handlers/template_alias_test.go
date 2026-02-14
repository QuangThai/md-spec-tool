package handlers_test

import (
	. "github.com/yourorg/md-spec-tool/internal/http/handlers"
	"testing"
)

func TestResolveTemplateContentName(t *testing.T) {
	tests := []struct {
		name         string
		expectedName string
		expectedOK   bool
	}{
		{name: "spec", expectedName: "spec", expectedOK: true},
		{name: "table", expectedName: "table", expectedOK: true},
		{name: "default", expectedName: "", expectedOK: false},
		{name: "spec-table", expectedName: "", expectedOK: false},
		{name: "unknown", expectedName: "", expectedOK: false},
	}

	for _, tt := range tests {
		gotName, gotOK := ResolveTemplateContentName(tt.name)
		if gotName != tt.expectedName || gotOK != tt.expectedOK {
			t.Fatalf("ResolveTemplateContentName(%q) = (%q,%v), want (%q,%v)",
				tt.name, gotName, gotOK, tt.expectedName, tt.expectedOK)
		}
	}
}
