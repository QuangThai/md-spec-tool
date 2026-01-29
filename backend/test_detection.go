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
	example1, err := ioutil.ReadFile("../example-1.md")
	if err != nil {
		fmt.Printf("Error reading example-1: %v\n", err)
		return
	}
	
	example2, err := ioutil.ReadFile("../example-2.md")
	if err != nil {
		fmt.Printf("Error reading example-2: %v\n", err)
		return
	}

	fmt.Println("=== EXAMPLE 1 (Blockquote Markdown) ===")
	analysis1 := converter.DetectInputType(string(example1))
	fmt.Printf("Type: %v, Confidence: %d%%\n", analysis1.Type, analysis1.Confidence)
	if analysis1.Reason != "" {
		fmt.Printf("Reason: %s\n", analysis1.Reason)
	}

	fmt.Println("\n=== EXAMPLE 2 (Table/TSV) ===")
	analysis2 := converter.DetectInputType(string(example2))
	fmt.Printf("Type: %v, Confidence: %d%%\n", analysis2.Type, analysis2.Confidence)
	if analysis2.Reason != "" {
		fmt.Printf("Reason: %s\n", analysis2.Reason)
	}
}
