package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yourorg/md-spec-tool/internal/converter"
)

type result struct {
	file     string
	template string
	ok       bool
	details  string
}

func main() {
	useCasesDir := filepath.Clean(filepath.Join("..", "use-cases"))
	outputRoot := filepath.Join(useCasesDir, "output")

	files, err := filepath.Glob(filepath.Join(useCasesDir, "*.md"))
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

	conv := converter.NewConverter()
	results := make([]result, 0)

	for _, filePath := range files {
		contentBytes, readErr := os.ReadFile(filePath)
		if readErr != nil {
			results = append(results, result{
				file:     filepath.Base(filePath),
				template: "(all)",
				ok:       false,
				details:  fmt.Sprintf("read error: %v", readErr),
			})
			continue
		}
		content := string(contentBytes)
		specDoc, specErr := converter.BuildSpecDocFromPaste(content)
		if specErr != nil {
			results = append(results, result{
				file:     filepath.Base(filePath),
				template: "(all)",
				ok:       false,
				details:  fmt.Sprintf("spec doc error: %v", specErr),
			})
			continue
		}

		analysis := converter.DetectInputType(content)

		for _, tmpl := range templates {
			outputDir := filepath.Join(outputRoot, tmpl)
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				results = append(results, result{
					file:     filepath.Base(filePath),
					template: tmpl,
					ok:       false,
					details:  fmt.Sprintf("mkdir error: %v", err),
				})
				continue
			}

			res, convErr := conv.ConvertPaste(content, tmpl)
			if convErr != nil {
				results = append(results, result{
					file:     filepath.Base(filePath),
					template: tmpl,
					ok:       false,
					details:  fmt.Sprintf("convert error: %v", convErr),
				})
				continue
			}

			outputPath := filepath.Join(outputDir, filepath.Base(filePath))
			if err := os.WriteFile(outputPath, []byte(res.MDFlow), 0o644); err != nil {
				results = append(results, result{
					file:     filepath.Base(filePath),
					template: tmpl,
					ok:       false,
					details:  fmt.Sprintf("write error: %v", err),
				})
				continue
			}

			missing := validateOutput(specDoc, analysis, tmpl, res.MDFlow)
			if len(missing) > 0 {
				results = append(results, result{
					file:     filepath.Base(filePath),
					template: tmpl,
					ok:       false,
					details:  "missing content: " + strings.Join(missing, ", "),
				})
				continue
			}

			results = append(results, result{
				file:     filepath.Base(filePath),
				template: tmpl,
				ok:       true,
				details:  "ok",
			})
		}
	}

	report(results, outputRoot)
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
		if !strings.Contains(normalizedOutput, normalize(trimmed)) {
			missing = append(missing, preview(trimmed))
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
	case "spec-table":
		check(row.No)
		check(row.ItemName)
		check(row.ItemType)
		check(row.RequiredOptional)
		check(row.DisplayConditions)
		check(row.InputRestrictions)
		check(row.Action)
		check(row.NavigationDest)
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
	if len(fields) <= 6 {
		return strings.Join(fields, " ")
	}
	return strings.Join(fields[:6], " ") + "..."
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

func report(results []result, outputRoot string) {
	total := len(results)
	failed := 0
	for _, r := range results {
		if !r.ok {
			failed++
		}
	}

	fmt.Printf("Use-cases converted: %d checks, %d failures\n", total, failed)
	fmt.Printf("Output directory: %s\n", outputRoot)
	for _, r := range results {
		if !r.ok {
			fmt.Printf("FAIL %s [%s] - %s\n", r.file, r.template, r.details)
		}
	}
	if failed == 0 {
		fmt.Println("All conversions succeeded with content coverage checks.")
	}
}
