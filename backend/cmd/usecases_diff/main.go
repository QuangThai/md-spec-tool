package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yourorg/md-spec-tool/internal/converter"
)

type diffResult struct {
	file     string
	template string
	missing  []string
	err      error
}

func main() {
	useCasesDir := filepath.Clean(filepath.Join("..", "use-cases"))
	outputRoot := filepath.Join(useCasesDir, "output")
	reportPath := filepath.Join(useCasesDir, "diff-report.txt")

	files, err := collectUseCaseFiles(useCasesDir)
	if err != nil {
		fmt.Printf("Failed to list use-cases: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("No use-case files found.")
		return
	}

	renderer := converter.NewMDFlowRenderer()
	templates := renderer.GetTemplateNames()
	sort.Strings(templates)

	results := make([]diffResult, 0)
	totalChecks := 0

	for _, filePath := range files {
		inputBytes, readErr := os.ReadFile(filePath)
		if readErr != nil {
			results = append(results, diffResult{
				file:     filepath.Base(filePath),
				template: "(all)",
				err:      readErr,
			})
			continue
		}
		content := string(inputBytes)
		specDoc, specErr := converter.BuildSpecDocFromPaste(content)
		if specErr != nil {
			results = append(results, diffResult{
				file:     filepath.Base(filePath),
				template: "(all)",
				err:      specErr,
			})
			continue
		}
		analysis := converter.DetectInputType(content)

		for _, tmpl := range templates {
			totalChecks++
			outputPath := filepath.Join(outputRoot, tmpl, filepath.Base(filePath))
			outputBytes, outErr := os.ReadFile(outputPath)
			if outErr != nil {
				results = append(results, diffResult{
					file:     filepath.Base(filePath),
					template: tmpl,
					err:      outErr,
				})
				continue
			}

			missing := validateOutput(specDoc, analysis, tmpl, string(outputBytes))
			if len(missing) > 0 {
				results = append(results, diffResult{
					file:     filepath.Base(filePath),
					template: tmpl,
					missing:  missing,
				})
			}
		}
	}

	if err := writeReport(reportPath, results, totalChecks); err != nil {
		fmt.Printf("Failed to write report: %v\n", err)
		os.Exit(1)
	}

	failed := 0
	for _, r := range results {
		if r.err != nil || len(r.missing) > 0 {
			failed++
		}
	}
	fmt.Printf("Use-cases diff: %d checks, %d flagged\n", totalChecks, failed)
	fmt.Printf("Diff report: %s\n", reportPath)
}

func collectUseCaseFiles(dir string) ([]string, error) {
	patterns := []string{"*.md", "*.tsv", "*.csv"}
	files := make([]string, 0)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}
	sort.Strings(files)
	return files, nil
}

func validateOutput(doc *converter.SpecDoc, analysis converter.InputAnalysis, template string, output string) []string {
	if analysis.Type == converter.InputTypeMarkdown {
		return validateMarkdown(doc, output)
	}
	return validateTable(doc, template, output)
}

func validateMarkdown(doc *converter.SpecDoc, output string) []string {
	missing := make([]string, 0)
	if doc.Prose == nil {
		return missing
	}
	normalizedOutput := normalize(output)
	for _, section := range doc.Prose.Sections {
		if strings.TrimSpace(section.Heading) != "" {
			if !strings.Contains(normalizedOutput, normalize(section.Heading)) {
				missing = append(missing, section.Heading)
			}
		}
		if strings.TrimSpace(section.Content) != "" {
			if !strings.Contains(normalizedOutput, normalize(section.Content)) {
				missing = append(missing, preview(section.Content))
			}
		}
	}
	if strings.TrimSpace(doc.Prose.RawMessage) != "" {
		if !strings.Contains(normalizedOutput, normalize(doc.Prose.RawMessage)) {
			missing = append(missing, preview(doc.Prose.RawMessage))
		}
	}
	return uniqueMissing(missing)
}

func validateTable(doc *converter.SpecDoc, template string, output string) []string {
	normalizedOutput := normalize(output)
	missing := make([]string, 0)
	for _, row := range doc.Rows {
		missing = append(missing, validateRow(row, template, normalizedOutput)...)
	}
	return uniqueMissing(missing)
}

func validateRow(row converter.SpecRow, template string, normalizedOutput string) []string {
	missing := make([]string, 0)
	check := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || trimmed == "-" {
			return
		}
		lines := strings.Split(trimmed, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || line == "-" {
				continue
			}
			if !strings.Contains(normalizedOutput, normalize(line)) {
				missing = append(missing, preview(line))
			}
		}
	}

	feature := strings.TrimSpace(row.Feature)
	scenario := strings.TrimSpace(row.Scenario)
	if feature != "" {
		check(feature)
	}
	if scenario != "" && !strings.EqualFold(feature, scenario) {
		check(scenario)
	}

	switch template {
	case "default":
		check(row.ID)
		check(row.Priority)
		check(row.Type)
		check(row.Precondition)
		check(row.Instructions)
		check(row.Inputs)
		check(row.Expected)
		check(row.Endpoint)
		check(row.Notes)
	case "feature-spec":
		check(row.Instructions)
		check(row.Expected)
		check(row.Notes)
	case "test-plan":
		check(row.ID)
		check(row.Priority)
		check(row.Type)
		check(row.Status)
		check(row.Endpoint)
		check(row.Precondition)
		check(row.Instructions)
		check(row.Inputs)
		check(row.Expected)
		check(row.Notes)
	case "api-endpoint":
		if strings.TrimSpace(row.Endpoint) != "" {
			check(row.Endpoint)
		} else {
			check(row.Scenario)
		}
		check(row.Scenario)
		check(row.Instructions)
		check(row.Inputs)
		check(row.Expected)
		check(row.Notes)
	case "spec-table":
		check(row.No)
		check(row.ItemName)
		check(row.ItemType)
		check(row.RequiredOptional)
		check(row.DisplayConditions)
		check(row.InputRestrictions)
		check(row.Action)
		check(row.NavigationDest)
		check(row.Notes)
	}

	return missing
}

func normalize(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.Join(strings.Fields(value), " ")
}

func preview(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.ReplaceAll(trimmed, "\n", " ")
	trimmed = strings.ReplaceAll(trimmed, "\t", " ")
	fields := strings.Fields(trimmed)
	if len(fields) <= 8 {
		return strings.Join(fields, " ")
	}
	return strings.Join(fields[:8], " ") + "..."
}

func uniqueMissing(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func writeReport(reportPath string, results []diffResult, totalChecks int) error {
	var builder strings.Builder
	builder.WriteString("Use-cases diff report\n")
	builder.WriteString(fmt.Sprintf("Total checks: %d\n", totalChecks))
	builder.WriteString("\n")

	flagged := 0
	for _, r := range results {
		if r.err == nil && len(r.missing) == 0 {
			continue
		}
		flagged++
		builder.WriteString(fmt.Sprintf("File: %s | Template: %s\n", r.file, r.template))
		if r.err != nil {
			builder.WriteString(fmt.Sprintf("Error: %v\n", r.err))
		} else {
			builder.WriteString("Missing lines:\n")
			for _, line := range r.missing {
				builder.WriteString("- ")
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}
		builder.WriteString("\n")
	}

	if flagged == 0 {
		builder.WriteString("No missing lines detected.\n")
	}

	return os.WriteFile(reportPath, []byte(builder.String()), 0o644)
}
