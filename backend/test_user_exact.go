//go:build manualtest
// +build manualtest

package main

import (
	"fmt"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func main() {
	// Exact user input from screenshot
	userInput := `> ## Background
> As part of SEO-related maintenance and operations work, the meta description on the Top page needs to be updated to reflect the latest marketing message for WebinarStock.
> 
> ## Scope
> Update the ` + "`" + `<meta name="description">` + "`" + ` tag on the Top page only.
> 
> ## Requirements
> - Replace the existing meta description with the following content:
> 
> ` + "```html\n" + `> <meta name="description" content="ウェビナー動画をアーカイブ配信でリード獲得資産に。掲載料無料・完全成果報酬型で、視聴リード単価5,000円〜。アーカイブ配信メディア「WebinarStock（ウェビナーストック）」">
> Raw message:
> [Top page]
> - Please update the <meta name="description"> as follows:<meta name="description" content="ウェビナー動画をアーカイブ配信でリード獲得資産に。掲載料無料・完全成果報酬型で、視聴リード単価5,000円〜。アーカイブ配信メディア「WebinarStock（ウェビナーストック）」">`

	fmt.Println("=== INPUT DETECTION ===")
	analysis := converter.DetectInputType(userInput)
	fmt.Printf("Type: %v\n", analysis.Type)
	fmt.Printf("Confidence: %d%%\n", analysis.Confidence)
	fmt.Printf("Reason: %s\n\n", analysis.Reason)

	if analysis.Type != converter.InputTypeMarkdown {
		fmt.Println("❌ PROBLEM: Input detected as " + string(analysis.Type) + ", not markdown!")
	}

	fmt.Println("\n=== CONVERSION RESULT ===")
	conv := converter.NewConverter()
	result, _ := conv.ConvertPaste(userInput, "")
	
	fmt.Printf("Output type in YAML: %v\n", "type: \"specification\"")
	fmt.Printf("First 500 chars:\n%s\n", result.MDFlow[:min(500, len(result.MDFlow))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
