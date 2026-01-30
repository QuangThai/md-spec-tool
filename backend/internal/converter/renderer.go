package converter

import (
	"bytes"
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
		templateName = "default"
	}

	tmpl, ok := r.templates[templateName]
	if !ok {
		tmpl = r.templates["default"]
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

	// Default template - comprehensive test case format
	r.templates["default"] = template.Must(template.New("default").Funcs(funcMap).Parse(defaultTemplate))

	// Feature spec template - simpler user story format
	r.templates["feature-spec"] = template.Must(template.New("feature-spec").Funcs(funcMap).Parse(featureSpecTemplate))

	// Test plan template - QA focused
	r.templates["test-plan"] = template.Must(template.New("test-plan").Funcs(funcMap).Parse(testPlanTemplate))

	// API endpoint template
	r.templates["api-endpoint"] = template.Must(template.New("api-endpoint").Funcs(funcMap).Parse(apiEndpointTemplate))

	// Spec table template - for UI/UX spec tables with Phase 3 fields
	r.templates["spec-table"] = template.Must(template.New("spec-table").Funcs(funcMap).Parse(specTableTemplate))
}

// GetTemplateNames returns available template names
func (r *MDFlowRenderer) GetTemplateNames() []string {
	var names []string
	for name := range r.templates {
		names = append(names, name)
	}
	return names
}

// HasTemplate returns true when template name exists.
func (r *MDFlowRenderer) HasTemplate(name string) bool {
	_, ok := r.templates[name]
	return ok
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

// Feature spec template
const featureSpecTemplate = `---
name: "{{.Title}}"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
type: "feature-spec"
---

# {{.Title}}
{{- range .FeatureGroups}}

## {{.Feature}}
{{- range .Rows}}
{{- $title := displayTitle .Feature .Scenario -}}
{{- if ne (lower $title) (lower .Feature)}}

### {{ $title }}
{{- end}}
{{- if notEmpty .Instructions}}

{{.Instructions}}
{{- end}}
{{- if notEmpty .Expected}}

**Acceptance Criteria:**
{{formatBullets .Expected}}
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
{{- end}}
{{- end}}
`

// Test plan template
const testPlanTemplate = `---
name: "{{.Title}} - Test Plan"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
---

# {{.Title}} - Test Plan

**Total Test Cases:** {{.TotalCount}}

| ID | Scenario | Priority | Type | Status |
|---|---|---|---|---|
{{range .Rows}}| {{.ID}} | {{.Scenario}} | {{.Priority}} | {{.Type}} | {{.Status}} |
{{end}}

---

## Test Case Details
{{- range .FeatureGroups}}

### {{.Feature}}
{{- $feature := .Feature -}}
{{- range .Rows}}
{{- if or .ID (ne (lower .Scenario) (lower $feature))}}

#### {{if .ID}}{{.ID}}: {{end}}{{.Scenario}}
{{- end}}
{{- if or (notEmpty .Priority) (notEmpty .Type) (notEmpty .Status) (notEmpty .Endpoint)}}

| Field | Value |
|---|---|
{{if notEmpty .Priority}}| Priority | {{.Priority}} |{{end}}
{{if notEmpty .Type}}| Type | {{.Type}} |{{end}}
{{if notEmpty .Status}}| Status | {{.Status}} |{{end}}
{{if notEmpty .Endpoint}}| Endpoint | ` + "`{{.Endpoint}}`" + ` |{{end}}
{{- end}}
{{- if notEmpty .Precondition}}

**Preconditions:**
{{formatBullets .Precondition}}
{{- end}}
{{- if notEmpty .Instructions}}

**Test Steps:**
{{formatSteps .Instructions}}
{{- end}}
{{- if notEmpty .Inputs}}

**Test Data:**
` + "```" + `
{{.Inputs}}
` + "```" + `
{{- end}}
{{- if notEmpty .Expected}}

**Expected Result:**
{{.Expected}}
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
{{- end}}
{{- end}}
`

// API endpoint template
const apiEndpointTemplate = `---
name: "{{.Title}} - API Specification"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
---

# {{.Title}} - API Specification
{{- range .FeatureGroups}}

## {{.Feature}}
{{- range .Rows}}
{{- $title := displayTitle .Feature .Scenario -}}

{{- if or .Endpoint (ne (lower $title) (lower .Feature))}}
### {{if .Endpoint}}{{.Endpoint}}{{else}}{{ $title }}{{end}}
{{- end}}
{{- if and (notEmpty .Scenario) (or .Endpoint (ne (lower $title) (lower .Feature)))}}

**Description:** {{.Scenario}}
{{- end}}
{{- if notEmpty .Instructions}}

**Flow:**
{{formatSteps .Instructions}}
{{- end}}
{{- if notEmpty .Inputs}}

**Request Parameters:**
` + "```" + `
{{.Inputs}}
` + "```" + `
{{- end}}
{{- if notEmpty .Expected}}

**Response:**
{{.Expected}}
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
{{- end}}
{{- end}}
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
type: "spec-table"
---

# {{.Title}}

## Summary Table

{{- if .Headers}}
| {{range .Headers}}{{headerCell .}} |{{end}}
| {{range .Headers}}---|{{end}}
{{range .Rows}}| {{range $h := $.Headers}}{{cellValue . $h}} |{{end}}
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
