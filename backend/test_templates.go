//go:build manualtest
// +build manualtest

package main

import (
	"fmt"
	"strings"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func main() {
	markdownInput := `> ## Background
> Test background text
> 
> ## Scope
> Test scope text`

	tableInput := `No	Item Name	Type
1	Test Item	text
2	Another Item	button`

	conv := converter.NewConverter()

	fmt.Println("\n========== TEST 1: MARKDOWN INPUT (empty template) ==========")
	result1, _ := conv.ConvertPaste(markdownInput, "")
	hasInputs1 := strings.Contains(result1.MDFlow, "_inputs")
	hasSpecType := strings.Contains(result1.MDFlow, `type: "specification"`)
	fmt.Printf("Has _inputs? %v (should be false)\n", hasInputs1)
	fmt.Printf("Has type: 'specification'? %v (should be true)\n", hasSpecType)
	fmt.Printf("Result: %s\n", result1.MDFlow[:min(400, len(result1.MDFlow))])

	fmt.Println("\n========== TEST 2: TABLE INPUT (empty template) ==========")
	result2, _ := conv.ConvertPaste(tableInput, "")
	hasInputs2 := strings.Contains(result2.MDFlow, "_inputs")
	hasDefault := strings.Contains(result2.MDFlow, "This specification contains")
	fmt.Printf("Has _inputs? %v (should be true for default template)\n", hasInputs2)
	fmt.Printf("Has test cases message? %v\n", hasDefault)
	fmt.Printf("Result: %s\n", result2.MDFlow[:min(400, len(result2.MDFlow))])

	fmt.Println("\n========== TEST 3: TABLE INPUT (spec-table template) ==========")
	result3, _ := conv.ConvertPaste(tableInput, "spec-table")
	hasSpecTable := strings.Contains(result3.MDFlow, `type: "spec-table"`)
	hasSummary := strings.Contains(result3.MDFlow, "Summary Table")
	fmt.Printf("Has type: 'spec-table'? %v (should be true)\n", hasSpecTable)
	fmt.Printf("Has Summary Table? %v\n", hasSummary)
	fmt.Printf("Result: %s\n", result3.MDFlow[:min(400, len(result3.MDFlow))])

	fmt.Println("\n========== SUMMARY ==========")
	test1OK := !hasInputs1 && hasSpecType
	test2OK := hasInputs2 && hasDefault
	test3OK := hasSpecTable && hasSummary
	
	fmt.Printf("TEST 1 (Markdown): %s\n", pass(test1OK))
	fmt.Printf("TEST 2 (Table default): %s\n", pass(test2OK))
	fmt.Printf("TEST 3 (Table spec-table): %s\n", pass(test3OK))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func pass(ok bool) string {
	if ok {
		return "✓ PASS"
	}
	return "✗ FAIL"
}
