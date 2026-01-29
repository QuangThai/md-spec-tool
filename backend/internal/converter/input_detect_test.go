package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func readTestFile(t *testing.T, relativePath string) string {
	// Navigate from backend/internal/converter to project root
	absPath, err := filepath.Abs(filepath.Join("..", "..", "..", relativePath))
	if err != nil {
		t.Fatalf("Failed to resolve path: %v", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", absPath, err)
	}
	return string(data)
}

func TestDetectInputType_Markdown(t *testing.T) {
	input := readTestFile(t, "example-1.md")
	result := DetectInputType(input)

	if result.Type != InputTypeMarkdown {
		t.Errorf("Expected InputTypeMarkdown, got %s (reason: %s)", result.Type, result.Reason)
	}
	if result.Confidence < 70 {
		t.Errorf("Expected confidence >= 70, got %d (reason: %s)", result.Confidence, result.Reason)
	}
}

func TestDetectInputType_Table(t *testing.T) {
	input := readTestFile(t, "example-2.md")
	result := DetectInputType(input)

	if result.Type != InputTypeTable {
		t.Errorf("Expected InputTypeTable, got %s (reason: %s)", result.Type, result.Reason)
	}
	if result.Confidence < 80 {
		t.Errorf("Expected confidence >= 80, got %d (reason: %s)", result.Confidence, result.Reason)
	}
}

func TestDetectInputType_EmptyInput(t *testing.T) {
	result := DetectInputType("")

	if result.Type != InputTypeUnknown {
		t.Errorf("Expected InputTypeUnknown, got %s", result.Type)
	}
	if result.Confidence != 0 {
		t.Errorf("Expected confidence 0, got %d", result.Confidence)
	}
}

func TestDetectInputType_PureMarkdownHeading(t *testing.T) {
	input := `# Title
## Section 1
Some content here
## Section 2
More content
`
	result := DetectInputType(input)

	if result.Type != InputTypeMarkdown {
		t.Errorf("Expected InputTypeMarkdown, got %s", result.Type)
	}
}

func TestDetectInputType_PureTable(t *testing.T) {
	input := `Col1	Col2	Col3	Col4
Row1	Data	Data	Data
Row2	Data	Data	Data
Row3	Data	Data	Data
`
	result := DetectInputType(input)

	if result.Type != InputTypeTable {
		t.Errorf("Expected InputTypeTable, got %s (reason: %s)", result.Type, result.Reason)
	}
	if result.Confidence < 80 {
		t.Errorf("Expected confidence >= 80, got %d", result.Confidence)
	}
}

func TestDetectInputType_CodeFences(t *testing.T) {
	input := "```html\n<div>Hello</div>\n```\n"
	result := DetectInputType(input)

	if result.Type != InputTypeMarkdown {
		t.Errorf("Expected InputTypeMarkdown, got %s", result.Type)
	}
}

func TestDetectInputType_BlockquoteMarkdown(t *testing.T) {
	input := `> ## Background
> Some background info
> 
> ## Scope
> Some scope info
`
	result := DetectInputType(input)

	if result.Type != InputTypeMarkdown {
		t.Errorf("Expected InputTypeMarkdown, got %s (reason: %s)", result.Type, result.Reason)
	}
}

func TestDetectInputType_BulletList(t *testing.T) {
	input := `# Requirements
- First requirement
- Second requirement
* Third requirement
`
	result := DetectInputType(input)

	if result.Type != InputTypeMarkdown {
		t.Errorf("Expected InputTypeMarkdown, got %s", result.Type)
	}
}
