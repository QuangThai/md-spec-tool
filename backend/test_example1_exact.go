//go:build manualtest
// +build manualtest

package main

import (
	"fmt"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func main() {
	// Exact content from example-1.md with line breaks
	example1 := `
> ## Background
> As part of SEO-related maintenance and operations work, the meta description on the Top page needs to be updated to reflect the latest marketing message for WebinarStock.
> 
> ## Scope
> Update the `  + "`" + `<meta name="description">` + "`" + ` tag on the Top page only.
> 
> ## Requirements
> - Replace the existing meta description with the following content:
> 
> ` + "```html\n" + `
> <meta name="description" content="ウェビナー動画をアーカイブ配信でリード獲得資産に。掲載料無料・完全成果報酬型で、視聴リード単価5,000円〜。アーカイブ配信メディア「WebinarStock（ウェビナーストック）」">
> Raw message:
> [Top page]
> - Please update the <meta name="description"> as follows:<meta name="description" content="ウェビナー動画をアーカイブ配信でリード獲得資産に。掲載料無料・完全成果報酬型で、視聴リード単価5,000円〜。アーカイブ配信メディア「WebinarStock（ウェビナーストック）」">`

	conv := converter.NewConverter()
	
	fmt.Println("=== DETECTION ===")
	analysis := converter.DetectInputType(example1)
	fmt.Printf("Type: %v, Confidence: %d%%\n", analysis.Type, analysis.Confidence)
	fmt.Printf("Reason: %s\n\n", analysis.Reason)

	fmt.Println("=== CONVERSION ===")
	result, err := conv.ConvertPaste(example1, "")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("Output length: %d\n", len(result.MDFlow))
	fmt.Printf("Warnings: %v\n\n", result.Warnings)
	fmt.Println(result.MDFlow[:min(600, len(result.MDFlow))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
