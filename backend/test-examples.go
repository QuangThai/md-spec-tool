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
	example1, _ := ioutil.ReadFile("example-1.md")
	example2, _ := ioutil.ReadFile("example-2.md")

	fmt.Println("=== EXAMPLE 1 ===")
	fmt.Println(string(example1[:100]))
	fmt.Println()
	
	analysis1 := converter.DetectInputType(string(example1))
	fmt.Printf("Type: %v, Confidence: %d, Reason: %s\n\n", analysis1.Type, analysis1.Confidence, analysis1.Reason)

	result1, err := converter.NewConverter().ConvertPaste(string(example1), "")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Output length: %d\n", len(result1.MDFlow))
		fmt.Println(result1.MDFlow[:200])
	}

	fmt.Println("\n=== EXAMPLE 2 ===")
	fmt.Println(string(example2[:100]))
	fmt.Println()
	
	analysis2 := converter.DetectInputType(string(example2))
	fmt.Printf("Type: %v, Confidence: %d, Reason: %s\n\n", analysis2.Type, analysis2.Confidence, analysis2.Reason)

	result2, err := converter.NewConverter().ConvertPaste(string(example2), "")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Output length: %d\n", len(result2.MDFlow))
		fmt.Println(result2.MDFlow[:200])
	}
}
