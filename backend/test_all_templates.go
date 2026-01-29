//go:build manualtest
// +build manualtest

package main

import (
	"fmt"
	"strings"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func repeat(s string, n int) string {
	var result string
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

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

	fmt.Println(repeat("=", 80))
	fmt.Println("TEST 1: MARKDOWN INPUT with DEFAULT template")
	fmt.Println("=" * 80)
	result1, _ := conv.ConvertPaste(markdownInput, "")
	fmt.Println("Has _inputs?", strings.Contains(result1.MDFlow, "_inputs"))
	fmt.Println("Has markdownTemplate header (type: specification)?", strings.Contains(result1.MDFlow, `type: "specification"`))
	fmt.Println("First 300 chars:")
	fmt.Println(result1.MDFlow[:min(300, len(result1.MDFlow))])

	fmt.Println("\n" + "=" * 80)
	fmt.Println("TEST 2: TABLE INPUT with DEFAULT template")
	fmt.Println("=" * 80)
	result2, _ := conv.ConvertPaste(tableInput, "")
	fmt.Println("Has _inputs?", strings.Contains(result2.MDFlow, "_inputs"))
	fmt.Println("Has defaultTemplate header (_inputs)?", strings.Contains(result2.MDFlow, "_inputs"))
	fmt.Println("First 300 chars:")
	fmt.Println(result2.MDFlow[:min(300, len(result2.MDFlow))])

	fmt.Println("\n" + "=" * 80)
	fmt.Println("TEST 3: TABLE INPUT with spec-table template")
	fmt.Println("=" * 80)
	result3, _ := conv.ConvertPaste(tableInput, "spec-table")
	fmt.Println("Has spec-table header?", strings.Contains(result3.MDFlow, `type: "spec-table"`))
	fmt.Println("First 300 chars:")
	fmt.Println(result3.MDFlow[:min(300, len(result3.MDFlow))])

	fmt.Println("\n" + "=" * 80)
	fmt.Println("SUMMARY")
	fmt.Println("=" * 80)
	fmt.Println("TEST 1 (Markdown):")
	fmt.Println("  ✓ Should have: type: \"specification\"")
	fmt.Println("  ✓ Should NOT have: _inputs")
	fmt.Println("  Status:", checkTest(result1.MDFlow, true, "_inputs"))
	
	fmt.Println("\nTEST 2 (Table, default):")
	fmt.Println("  ✓ Should have: _inputs (defaultTemplate)")
	fmt.Println("  Status:", checkTest(result2.MDFlow, false, "_inputs"))
	
	fmt.Println("\nTEST 3 (Table, spec-table):")
	fmt.Println("  ✓ Should have: type: \"spec-table\"")
	fmt.Println("  Status:", checkTest(result3.MDFlow, true, `type: "spec-table"`))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func checkTest(output string, shouldHave bool, marker string) string {
	has := strings.Contains(output, marker)
	if shouldHave && has {
		return "✓ PASS"
	}
	if !shouldHave && !has {
		return "✓ PASS"
	}
	return "✗ FAIL - contains: " + fmt.Sprintf("%v", has) + ", expected: " + fmt.Sprintf("%v", shouldHave)
}
