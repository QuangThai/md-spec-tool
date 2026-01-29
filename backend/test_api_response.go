//go:build manualtest
// +build manualtest

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func main() {
	// Read example-1
	example1, _ := ioutil.ReadFile("../example-1.md")

	conv := converter.NewConverter()
	result, err := conv.ConvertPaste(string(example1), "")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	// Simulate API response
	type Response struct {
		MDFlow   string                  `json:"mdflow"`
		Warnings []string                `json:"warnings"`
		Meta     converter.SpecDocMeta   `json:"meta"`
	}

	response := Response{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
	}

	// JSON marshal like API would
	jsonResp, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println("=== API Response JSON ===")
	fmt.Println(string(jsonResp))

	fmt.Println("\n=== First 500 chars of MDFlow ===")
	fmt.Println(result.MDFlow[:500])
}
