package converter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"
)

// MDFlowRenderer renders SpecDoc to MDFlow format
type MDFlowRenderer struct {
	templates map[string]*template.Template
}

// NewMDFlowRenderer creates a new MDFlowRenderer
func NewMDFlowRenderer() *MDFlowRenderer {
	r := &MDFlowRenderer{
		templates: make(map[string]*template.Template),
	}
	r.registerDefaultTemplates()
	return r
}

// Render converts a SpecDoc to MDFlow markdown
func (r *MDFlowRenderer) Render(doc *SpecDoc, templateName string) (string, error) {
	if templateName == "" {
		templateName = "spec"
	}

	tmpl, ok := r.templates[templateName]
	if !ok {
		// Return detailed error message with available templates
		names := r.GetTemplateNames()
		return "", fmt.Errorf("unknown template: %s (available: %v)", templateName, names)
	}

	// Group rows by feature
	featureGroups := r.groupByFeature(doc.Rows)

	data := map[string]interface{}{
		"Title":         doc.Title,
		"Rows":          doc.Rows,
		"FeatureGroups": featureGroups,
		"TotalCount":    len(doc.Rows),
		"GeneratedAt":   time.Now().Format("2006-01-02"),
		"Meta":          doc.Meta,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderMarkdown converts a markdown SpecDoc to MDFlow format
func (r *MDFlowRenderer) RenderMarkdown(doc *SpecDoc, templateName string) (string, error) {
	// Default markdown template for prose content
	funcMap := template.FuncMap{
		"replace": strings.Replace,
		"add1": func(i int) int {
			return i + 1
		},
	}
	tmpl := template.Must(template.New("markdown").Funcs(funcMap).Parse(markdownTemplate))

	data := map[string]interface{}{
		"Title":            doc.Title,
		"Sections":         doc.Prose.Sections,
		"OriginalMarkdown": doc.Prose.OriginalMarkdown,
		"RawMessage":       doc.Prose.RawMessage,
		"GeneratedAt":      time.Now().Format("2006-01-02"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// FeatureGroup represents a group of rows for a single feature
type FeatureGroup struct {
	Feature string
	Rows    []SpecRow
}

// MetadataPair represents a key/value metadata item
type MetadataPair struct {
	Key   string
	Value string
}

// groupByFeature groups rows by their Feature field
func (r *MDFlowRenderer) groupByFeature(rows []SpecRow) []FeatureGroup {
	groupMap := make(map[string][]SpecRow)
	var order []string

	for _, row := range rows {
		feature := row.Feature
		if feature == "" {
			feature = "Uncategorized"
		}
		if _, exists := groupMap[feature]; !exists {
			order = append(order, feature)
		}
		groupMap[feature] = append(groupMap[feature], row)
	}

	var groups []FeatureGroup
	for _, feature := range order {
		groups = append(groups, FeatureGroup{
			Feature: feature,
			Rows:    groupMap[feature],
		})
	}

	return groups
}

// registerDefaultTemplates registers built-in templates
func (r *MDFlowRenderer) registerDefaultTemplates() {
	funcMap := template.FuncMap{
		"formatSteps":   formatSteps,
		"formatBullets": formatBullets,
		"notEmpty":      notEmpty,
		"displayTitle":  displayTitle,
		"escapeYAML":    escapeYAML,
		"cellValue":     cellValue,
		"headerCell":    headerCell,
		"metadataPairs": metadataPairs,
		"trimPrefix":    strings.TrimPrefix,
		"lower":         strings.ToLower,
		"upper":         strings.ToUpper,
		"replace":       strings.ReplaceAll,
	}

	r.templates["spec"] = template.Must(template.New("spec").Funcs(funcMap).Parse(defaultTemplate))
	r.templates["table"] = template.Must(template.New("table").Funcs(funcMap).Parse(specTableTemplate))
}

// GetTemplateNames returns available template names
func (r *MDFlowRenderer) GetTemplateNames() []string {
	names := []string{"spec", "table"}
	return names
}

// HasTemplate returns true when template name exists.
func (r *MDFlowRenderer) HasTemplate(name string) bool {
	_, ok := r.templates[name]
	return ok
}

// RenderCustom renders a SpecDoc using a custom template string
func (r *MDFlowRenderer) RenderCustom(doc *SpecDoc, templateContent string) (string, error) {
	funcMap := template.FuncMap{
		"formatSteps":   formatSteps,
		"formatBullets": formatBullets,
		"notEmpty":      notEmpty,
		"displayTitle":  displayTitle,
		"escapeYAML":    escapeYAML,
		"cellValue":     cellValue,
		"headerCell":    headerCell,
		"metadataPairs": metadataPairs,
		"trimPrefix":    strings.TrimPrefix,
		"lower":         strings.ToLower,
		"upper":         strings.ToUpper,
		"replace":       strings.ReplaceAll,
	}

	tmpl, err := template.New("custom").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return "", err
	}

	// Group rows by feature
	featureGroups := r.groupByFeature(doc.Rows)

	data := map[string]interface{}{
		"Title":         doc.Title,
		"Rows":          doc.Rows,
		"FeatureGroups": featureGroups,
		"TotalCount":    len(doc.Rows),
		"GeneratedAt":   time.Now().Format("2006-01-02"),
		"Meta":          doc.Meta,
		"Headers":       doc.Headers,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetTemplateContent returns the content of a built-in template
func (r *MDFlowRenderer) GetTemplateContent(name string) string {
	switch name {
	case "spec":
		return defaultTemplate
	case "table":
		return specTableTemplate
	default:
		return ""
	}
}

// TemplateInfo contains metadata about template variables and functions
type TemplateInfo struct {
	Variables []TemplateVariable `json:"variables"`
	Functions []TemplateFunction `json:"functions"`
}

// TemplateVariable describes a template variable
type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// TemplateFunction describes a template function
type TemplateFunction struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
}

// GetTemplateInfo returns information about available template variables and functions
func (r *MDFlowRenderer) GetTemplateInfo() TemplateInfo {
	return TemplateInfo{
		Variables: []TemplateVariable{
			{Name: ".Title", Type: "string", Description: "Document title (from sheet name or 'Converted Spec')"},
			{Name: ".Rows", Type: "[]SpecRow", Description: "All data rows from the input"},
			{Name: ".FeatureGroups", Type: "[]FeatureGroup", Description: "Rows grouped by Feature field"},
			{Name: ".TotalCount", Type: "int", Description: "Total number of rows"},
			{Name: ".GeneratedAt", Type: "string", Description: "Generation date (YYYY-MM-DD)"},
			{Name: ".Meta", Type: "SpecDocMeta", Description: "Metadata about the parsed document"},
			{Name: ".Headers", Type: "[]string", Description: "Original column headers from input"},
		},
		Functions: []TemplateFunction{
			{Name: "formatSteps", Signature: "formatSteps(text string) string", Description: "Format multi-line text as steps"},
			{Name: "formatBullets", Signature: "formatBullets(text string) string", Description: "Format text as bullet points"},
			{Name: "notEmpty", Signature: "notEmpty(s string) bool", Description: "Check if string is not empty or '-'"},
			{Name: "displayTitle", Signature: "displayTitle(feature, scenario string) string", Description: "Get display title from feature/scenario"},
			{Name: "escapeYAML", Signature: "escapeYAML(s string) string", Description: "Escape special YAML characters"},
			{Name: "cellValue", Signature: "cellValue(row SpecRow, header string) string", Description: "Get cell value by header name"},
			{Name: "headerCell", Signature: "headerCell(header string) string", Description: "Format header for table cell"},
			{Name: "metadataPairs", Signature: "metadataPairs(row SpecRow) []MetadataPair", Description: "Get unmapped metadata as key-value pairs"},
			{Name: "trimPrefix", Signature: "trimPrefix(s, prefix string) string", Description: "Remove prefix from string"},
			{Name: "lower", Signature: "lower(s string) string", Description: "Convert to lowercase"},
			{Name: "upper", Signature: "upper(s string) string", Description: "Convert to uppercase"},
			{Name: "replace", Signature: "replace(s, old, new string) string", Description: "Replace all occurrences"},
		},
	}
}

// Helper functions for templates
func formatSteps(text string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func formatBullets(text string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Remove existing bullet if present
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			result = append(result, "- "+line)
		}
	}
	return strings.Join(result, "\n")
}

func notEmpty(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return false
	}
	return trimmed != "-"
}

func displayTitle(feature string, scenario string) string {
	feature = strings.TrimSpace(feature)
	scenario = strings.TrimSpace(scenario)
	if scenario == "" {
		return feature
	}
	if feature != "" && strings.EqualFold(feature, scenario) {
		return feature
	}
	return scenario
}

func escapeYAML(s string) string {
	// Escape special YAML characters
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	if strings.Contains(s, "\n") || strings.Contains(s, ":") || strings.Contains(s, "#") {
		return `"` + s + `"`
	}
	return s
}

func metadataPairs(row SpecRow) []MetadataPair {
	if len(row.Metadata) == 0 {
		return nil
	}
	keys := make([]string, 0, len(row.Metadata))
	for key, value := range row.Metadata {
		key = strings.TrimSpace(key)
		value = normalizeCellValue(value)
		if key == "" || value == "" {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)
	items := make([]MetadataPair, 0, len(keys))
	for _, key := range keys {
		items = append(items, MetadataPair{Key: key, Value: normalizeCellValue(row.Metadata[key])})
	}
	return items
}

func headerCell(header string) string {
	return formatTableCell(header)
}

func cellValue(row SpecRow, header string) string {
	normalized := normalizeHeader(header)
	if field, ok := HeaderSynonyms[normalized]; ok {
		return formatTableCell(valueForField(row, field))
	}
	if value, ok := row.Metadata[header]; ok {
		return formatTableCell(value)
	}
	return ""
}

func normalizeHeader(header string) string {
	h := strings.ToLower(strings.TrimSpace(header))
	h = strings.Join(strings.Fields(h), " ")
	return h
}

func normalizeCellValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "-" {
		return ""
	}
	return trimmed
}

func formatTableCell(value string) string {
	value = normalizeCellValue(value)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\n", "<br>")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}

func valueForField(row SpecRow, field CanonicalField) string {
	switch field {
	case FieldID:
		return normalizeCellValue(row.ID)
	case FieldFeature:
		return normalizeCellValue(row.Feature)
	case FieldScenario:
		return normalizeCellValue(row.Scenario)
	case FieldInstructions:
		return normalizeCellValue(row.Instructions)
	case FieldInputs:
		return normalizeCellValue(row.Inputs)
	case FieldExpected:
		return normalizeCellValue(row.Expected)
	case FieldPrecondition:
		return normalizeCellValue(row.Precondition)
	case FieldPriority:
		return normalizeCellValue(row.Priority)
	case FieldType:
		return normalizeCellValue(row.Type)
	case FieldStatus:
		return normalizeCellValue(row.Status)
	case FieldEndpoint:
		return normalizeCellValue(row.Endpoint)
	case FieldNotes:
		return normalizeCellValue(row.Notes)
	case FieldNo:
		return normalizeCellValue(row.No)
	case FieldItemName:
		return normalizeCellValue(row.ItemName)
	case FieldItemType:
		return normalizeCellValue(row.ItemType)
	case FieldRequiredOptional:
		return normalizeCellValue(row.RequiredOptional)
	case FieldInputRestrictions:
		return normalizeCellValue(row.InputRestrictions)
	case FieldDisplayConditions:
		return normalizeCellValue(row.DisplayConditions)
	case FieldAction:
		return normalizeCellValue(row.Action)
	case FieldNavigationDest:
		return normalizeCellValue(row.NavigationDest)
	default:
		return ""
	}
}

// Default template
const defaultTemplate = `---
name: "{{.Title}}"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
_inputs:
  _feature_filter:
    type: text
    description: "Filter by feature name"
---

# {{.Title}}

This specification contains {{.TotalCount}} test cases.

{{range .FeatureGroups}}
## {{.Feature}}
{{range .Rows}}
{{- $title := displayTitle .Feature .Scenario -}}
{{if or .ID (ne (lower $title) (lower .Feature))}}
### {{if .ID}}{{.ID}}: {{end}}{{ $title }}
{{end}}
{{- if .Priority}}**Priority:** {{.Priority}}{{end}}{{if .Type}} | **Type:** {{.Type}}{{end}}
{{- if notEmpty .Precondition}}
**Preconditions:**
{{formatBullets .Precondition}}
{{- end}}
{{- if notEmpty .Instructions}}
**Steps:**
{{formatSteps .Instructions}}
{{- end}}
{{- if notEmpty .Inputs}}
**Test Data:**
- {{.Inputs}}
{{- end}}
{{- if notEmpty .Expected}}
**Expected Result:**
- {{.Expected}}
{{- end}}
{{- if notEmpty .Endpoint}}
**API/Endpoint:** ` + "`{{.Endpoint}}`" + `
{{- end}}
{{- if notEmpty .Notes}}
**Notes:** {{.Notes}}
{{- end}}
{{- $meta := metadataPairs .}}
{{- if $meta}}
**Additional Fields:**
{{- range $meta}}
- {{.Key}}: {{.Value}}
{{- end}}
{{- end}}

---
{{end}}
{{end}}
`

// Markdown template for prose/blockquote content
const markdownTemplate = `---
name: "{{.Title}}"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
type: "specification"
---

# {{.Title}}
{{- if .Sections}}
{{- range $index, $section := .Sections}}

## {{ add1 $index }}. {{ $section.Heading }}
{{- if $section.Content}}

{{ $section.Content }}
{{- end}}
{{- end}}
{{- else}}

{{ .OriginalMarkdown }}
{{- end}}
{{- if .RawMessage}}

## Raw Message

{{ .RawMessage }}
{{- end}}
`

// Spec table template for UI/UX specification tables with Phase 3 fields
const specTableTemplate = `---
name: "{{.Title}}"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
type: "table"
---

# {{.Title}}

## Summary Table

{{- if .Headers}}
| {{range .Headers}}{{headerCell .}} |{{end}}
| {{range .Headers}}---|{{end}}
{{range $row := .Rows}}| {{range $h := $.Headers}}{{cellValue $row $h}} |{{end}}
{{end}}
{{- else}}
| No | Item Name | Type | Required | Display Conditions | Action | Navigation | Notes |
|----|-----------|------|----------|-------------------|--------|-----------|-------|
{{range .Rows}}| {{.No}} | {{.ItemName}} | {{.ItemType}} | {{.RequiredOptional}} | {{.DisplayConditions}} | {{.Action}} | {{.NavigationDest}} | {{.Notes}} |
{{end}}
{{- end}}

---

## Item Details
{{- range .Rows}}

### {{if .No}}{{.No}}. {{end}}{{.ItemName}}
{{- if notEmpty .ItemType}}

**Type:** {{.ItemType}}
{{- end}}
{{- if notEmpty .RequiredOptional}}

**Required:** {{.RequiredOptional}}
{{- end}}
{{- if notEmpty .DisplayConditions}}

**Display Conditions:**
{{.DisplayConditions}}
{{- end}}
{{- if notEmpty .InputRestrictions}}

**Input Restrictions:**
{{.InputRestrictions}}
{{- end}}
{{- if notEmpty .Action}}

**Action:** {{.Action}}
{{- end}}
{{- if notEmpty .NavigationDest}}

**Navigation Destination:** {{.NavigationDest}}
{{- end}}
{{- if notEmpty .Notes}}

**Notes:** {{.Notes}}
{{- end}}
{{- $meta := metadataPairs .}}
{{- if $meta}}

**Additional Fields:**
{{- range $meta}}
- {{.Key}}: {{.Value}}
{{- end}}
{{- end}}

---
{{- end}}`
