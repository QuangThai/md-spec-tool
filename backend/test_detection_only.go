//go:build manualtest
// +build manualtest

package main

import (
	"fmt"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func main() {
	// Test different variations
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Simple markdown",
			input: `> ## Background
> Text here`,
		},
		{
			name: "With HTML",
			input: `> ## Background
> <meta name="description" content="test">`,
		},
		{
			name: "With backticks",
			input: "```html\n<meta name=\"description\">",
		},
		{
			name: "User exact input",
			input: `> ## Background
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
> - Please update the <meta name="description"> as follows:<meta name="description" content="ウェビナー動画をアーカイブ配信でリード獲得資産に。掲載料無料・完全成果報酬型で、視聴リード単価5,000円〜。アーカイブ配信メディア「WebinarStock（ウェビナーストック）」">`,
		},
	}

	for _, tt := range tests {
		analysis := converter.DetectInputType(tt.input)
		fmt.Printf("%s:\n  Type: %v\n  Confidence: %d%%\n\n", tt.name, analysis.Type, analysis.Confidence)
	}
}
