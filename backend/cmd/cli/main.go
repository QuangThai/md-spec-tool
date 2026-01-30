package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/diff"
)

const (
	version = "1.0.0"
	usage   = `MDFlow CLI - Convert spreadsheets to Markdown specifications

Usage:
  mdflow <command> [options]

Commands:
  convert     Convert a file (TSV/CSV/XLSX) to MDFlow markdown
  diff        Compare two MDFlow files
  templates   List available templates
  version     Print version information

Run 'mdflow <command> --help' for more information on a command.

Examples:
  mdflow convert --input spec.tsv --output spec.mdflow.md
  mdflow convert --input data.xlsx --sheet "Sheet1" --template test-plan
  mdflow diff before.md after.md
  mdflow templates
`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "convert":
		runConvert(os.Args[2:])
	case "diff":
		runDiff(os.Args[2:])
	case "templates":
		runTemplates()
	case "version", "-v", "--version":
		fmt.Printf("mdflow version %s\n", version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Print(usage)
		os.Exit(1)
	}
}

func runConvert(args []string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	input := fs.String("input", "", "Input file path (required)")
	output := fs.String("output", "", "Output file path (default: stdout)")
	template := fs.String("template", "default", "Template name")
	sheet := fs.String("sheet", "", "Sheet name (for XLSX files)")
	jsonOutput := fs.Bool("json", false, "Output as JSON with metadata")
	
	fs.Usage = func() {
		fmt.Println(`Convert a file to MDFlow markdown

Usage:
  mdflow convert --input <file> [options]

Options:
  --input     Input file path (TSV, CSV, or XLSX) (required)
  --output    Output file path (default: stdout)
  --template  Template name (default: "default")
  --sheet     Sheet name for XLSX files
  --json      Output as JSON with metadata

Examples:
  mdflow convert --input spec.tsv
  mdflow convert --input spec.tsv --output spec.mdflow.md
  mdflow convert --input data.xlsx --sheet "Requirements" --template feature-spec
  mdflow convert --input test.csv --json`)
	}
	
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *input == "" {
		fmt.Fprintln(os.Stderr, "Error: --input is required")
		fs.Usage()
		os.Exit(1)
	}

	// Read input file
	content, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	conv := converter.NewConverter()
	var result *converter.ConvertResponse

	ext := strings.ToLower(filepath.Ext(*input))
	switch ext {
	case ".xlsx", ".xls":
		result, err = conv.ConvertXLSX(*input, *sheet, *template)
	case ".tsv", ".csv", ".txt", ".md":
		result, err = conv.ConvertPaste(string(content), *template)
	default:
		// Try as text/paste
		result, err = conv.ConvertPaste(string(content), *template)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting file: %v\n", err)
		os.Exit(1)
	}

	// Prepare output
	var outputContent string
	if *jsonOutput {
		jsonBytes, jsonErr := json.MarshalIndent(map[string]interface{}{
			"mdflow":   result.MDFlow,
			"warnings": result.Warnings,
			"meta":     result.Meta,
		}, "", "  ")
		if jsonErr != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", jsonErr)
			os.Exit(1)
		}
		outputContent = string(jsonBytes)
	} else {
		outputContent = result.MDFlow
	}

	// Write output
	if *output == "" {
		fmt.Print(outputContent)
	} else {
		if err := os.WriteFile(*output, []byte(outputContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Written to %s\n", *output)
		
		// Print warnings to stderr
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "[%s] %s\n", w.Severity, w.Message)
		}
	}
}

func runDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	output := fs.String("output", "", "Output file path (default: stdout)")
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	
	fs.Usage = func() {
		fmt.Println(`Compare two MDFlow files

Usage:
  mdflow diff <before-file> <after-file> [options]

Options:
  --output    Output file path (default: stdout)
  --json      Output as JSON

Examples:
  mdflow diff old.md new.md
  mdflow diff spec-v1.mdflow.md spec-v2.mdflow.md --json`)
	}
	
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	remainingArgs := fs.Args()
	if len(remainingArgs) < 2 {
		fmt.Fprintln(os.Stderr, "Error: two files are required")
		fs.Usage()
		os.Exit(1)
	}

	beforePath := remainingArgs[0]
	afterPath := remainingArgs[1]

	before, err := os.ReadFile(beforePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading before file: %v\n", err)
		os.Exit(1)
	}

	after, err := os.ReadFile(afterPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading after file: %v\n", err)
		os.Exit(1)
	}

	result := diff.Diff(string(before), string(after))

	var outputContent string
	if *jsonOutput {
		jsonBytes, jsonErr := json.MarshalIndent(result, "", "  ")
		if jsonErr != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", jsonErr)
			os.Exit(1)
		}
		outputContent = string(jsonBytes)
	} else {
		outputContent = diff.FormatUnified(result)
	}

	// Write output
	if *output == "" {
		fmt.Print(outputContent)
	} else {
		if err := os.WriteFile(*output, []byte(outputContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Written to %s\n", *output)
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "Changes: +%d -%d lines\n", result.Added, result.Removed)
}

func runTemplates() {
	renderer := converter.NewMDFlowRenderer()
	templates := renderer.GetTemplateNames()

	fmt.Println("Available templates:")
	for _, t := range templates {
		desc := getTemplateDescription(t)
		fmt.Printf("  %-15s %s\n", t, desc)
	}
}

func getTemplateDescription(name string) string {
	descriptions := map[string]string{
		"default":      "Standard test case format",
		"feature-spec": "User story and acceptance criteria format",
		"test-plan":    "QA test plan format with test cases",
		"api-endpoint": "API documentation format",
		"spec-table":   "UI specification table format",
	}
	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return ""
}
