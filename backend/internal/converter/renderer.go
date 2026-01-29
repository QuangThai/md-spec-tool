package converter

import (
	"bytes"
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
	}
	tmpl := template.Must(template.New("markdown").Funcs(funcMap).Parse(markdownTemplate))

	data := map[string]interface{}{
		"Title":             doc.Title,
		"Sections":          doc.Prose.Sections,
		"OriginalMarkdown":  doc.Prose.OriginalMarkdown,
		"RawMessage":        doc.Prose.RawMessage,
		"GeneratedAt":       time.Now().Format("2006-01-02"),
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
		"formatSteps":    formatSteps,
		"formatBullets":  formatBullets,
		"notEmpty":       notEmpty,
		"displayTitle":   displayTitle,
		"escapeYAML":     escapeYAML,
		"trimPrefix":     strings.TrimPrefix,
		"lower":          strings.ToLower,
		"upper":          strings.ToUpper,
		"replace":        strings.ReplaceAll,
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
	return strings.TrimSpace(s) != ""
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

{{range .FeatureGroups}}
## {{.Feature}}

{{range .Rows}}
{{- $title := displayTitle .Feature .Scenario -}}
{{if ne (lower $title) (lower .Feature)}}
### {{ $title }}
{{end}}

{{if notEmpty .Instructions}}
{{.Instructions}}
{{end}}

{{if notEmpty .Expected}}
**Acceptance Criteria:**
{{formatBullets .Expected}}
{{end}}

{{end}}
{{end}}
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

{{range .FeatureGroups}}
### {{.Feature}}

{{range .Rows}}
#### {{.ID}}: {{.Scenario}}

| Field | Value |
|---|---|
| Priority | {{.Priority}} |
| Type | {{.Type}} |
| Status | {{.Status}} |
{{if notEmpty .Endpoint}}| Endpoint | ` + "`{{.Endpoint}}`" + ` |{{end}}

{{if notEmpty .Precondition}}
**Preconditions:**
{{formatBullets .Precondition}}
{{end}}

{{if notEmpty .Instructions}}
**Test Steps:**
{{formatSteps .Instructions}}
{{end}}

{{if notEmpty .Inputs}}
**Test Data:**
` + "```" + `
{{.Inputs}}
` + "```" + `
{{end}}

{{if notEmpty .Expected}}
**Expected Result:**
{{.Expected}}
{{end}}

---

{{end}}
{{end}}
`

// API endpoint template
const apiEndpointTemplate = `---
name: "{{.Title}} - API Specification"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
---

# {{.Title}} - API Specification

{{range .FeatureGroups}}
## {{.Feature}}

{{range .Rows}}
### {{if .Endpoint}}{{.Endpoint}}{{else}}{{.Scenario}}{{end}}

{{if notEmpty .Scenario}}
**Description:** {{.Scenario}}
{{end}}

{{if notEmpty .Instructions}}
**Flow:**
{{formatSteps .Instructions}}
{{end}}

{{if notEmpty .Inputs}}
**Request Parameters:**
` + "```" + `
{{.Inputs}}
` + "```" + `
{{end}}

{{if notEmpty .Expected}}
**Response:**
{{.Expected}}
{{end}}

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

{{range .Sections}}
## {{.Heading}}

{{.Content}}

{{end}}
{{if .RawMessage}}
## Raw Message

{{.RawMessage}}

{{end}}
---

<details>
<summary>Original Content</summary>

{{ replace .OriginalMarkdown "##" "####" -1 }}

</details>
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

| No | Item Name | Type | Required | Display Conditions | Action | Navigation |
|----|-----------|------|----------|-------------------|--------|-----------|
{{range .Rows}}| {{.No}} | {{.ItemName}} | {{.ItemType}} | {{.RequiredOptional}} | {{.DisplayConditions}} | {{.Action}} | {{.NavigationDest}} |
{{end}}

---

## Item Details

{{range .Rows}}
### {{if .No}}{{.No}}. {{end}}{{.ItemName}}

{{if notEmpty .ItemType}}**Type:** {{.ItemType}}
{{end}}
{{if notEmpty .RequiredOptional}}**Required:** {{.RequiredOptional}}
{{end}}
{{if notEmpty .DisplayConditions}}
**Display Conditions:**
{{.DisplayConditions}}

{{end}}
**Input Restrictions:**
{{if notEmpty .InputRestrictions}}{{.InputRestrictions}}{{else}}-{{end}}

{{if notEmpty .Action}}
**Action:** {{.Action}}
{{end}}
{{if notEmpty .NavigationDest}}
**Navigation Destination:** {{.NavigationDest}}
{{end}}

---

{{end}}`
