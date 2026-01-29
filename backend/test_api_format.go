//go:build manualtest
// +build manualtest

package main

import (
	"encoding/json"
	"fmt"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func main() {
	// Test what the actual API handler returns
	markdownText := `> ## Background
> Test content`

	analysis := converter.DetectInputType(markdownText)

	// Simulate what handler does
	typeStr := "unknown"
	switch analysis.Type {
	case converter.InputTypeMarkdown:
		typeStr = "markdown"
	case converter.InputTypeTable:
		typeStr = "table"
	}

	response := handlers.InputAnalysisResponse{
		Type:       typeStr,
		Confidence: float64(analysis.Confidence),
		Reason:     analysis.Reason,
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println("API Response JSON:")
	fmt.Println(string(jsonData))

	// Parse it back to see if FE can understand
	var parsed handlers.InputAnalysisResponse
	json.Unmarshal(jsonData, &parsed)
	fmt.Printf("\nParsed back:\n  Type: %s\n  Confidence: %f\n  Reason: %s\n", parsed.Type, parsed.Confidence, parsed.Reason)

	// Check what FE interface expects
	fmt.Println("\nFE expects (from TypeScript interface):")
	fmt.Println("  type: 'markdown' | 'table' | 'unknown'")
	fmt.Println("  confidence: number")
	fmt.Println("  reason?: string")
}
