//go:build manualtest
// +build manualtest

package main

import (
	"fmt"
	"io/ioutil"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func main() {
	// Read examples
	example1, _ := ioutil.ReadFile("../example-1.md")
	example2, _ := ioutil.ReadFile("../example-2.md")

	conv := converter.NewConverter()

	fmt.Println("=== EXAMPLE 1 CONVERSION ===")
	result1, err := conv.ConvertPaste(string(example1), "")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("MDFlow length: %d bytes\n", len(result1.MDFlow))
		fmt.Printf("Warnings: %d\n", len(result1.Warnings))
		for _, w := range result1.Warnings {
			fmt.Printf("  - %s\n", w)
		}
		fmt.Printf("\nFirst 500 chars:\n%s\n", result1.MDFlow[:min(500, len(result1.MDFlow))])
	}

	fmt.Println("\n=== EXAMPLE 2 CONVERSION ===")
	result2, err := conv.ConvertPaste(string(example2), "")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("MDFlow length: %d bytes\n", len(result2.MDFlow))
		fmt.Printf("Warnings: %d\n", len(result2.Warnings))
		for _, w := range result2.Warnings {
			fmt.Printf("  - %s\n", w)
		}
		fmt.Printf("Total rows: %d\n", result2.Meta.TotalRows)
		fmt.Printf("\nFirst 500 chars:\n%s\n", result2.MDFlow[:min(500, len(result2.MDFlow))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
