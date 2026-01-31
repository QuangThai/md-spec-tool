export const STORAGE_KEY = "mdflow-custom-templates";

// Built-in template names
export const BUILT_IN_TEMPLATES = [
  { id: "default", name: "Default", description: "Standard test case format" },
  {
    id: "feature-spec",
    name: "Feature Spec",
    description: "User story format",
  },
  { id: "test-plan", name: "Test Plan", description: "QA test plan format" },
  {
    id: "api-endpoint",
    name: "API Endpoint",
    description: "API documentation",
  },
  {
    id: "spec-table",
    name: "Spec Table",
    description: "UI specification table",
  },
];

// Default sample data for preview
export const DEFAULT_SAMPLE_DATA = `Feature\tScenario\tInstructions\tExpected\tPriority\tType\tNotes
User Authentication\tValid Login\t1. Enter username
2. Enter password
3. Click login button\tDashboard should display with user name\tHigh\tPositive\tCore feature
User Authentication\tInvalid Password\t1. Enter valid username
2. Enter wrong password
3. Click login button\tError message: "Invalid credentials"\tHigh\tNegative\tSecurity test
Profile Management\tUpdate Profile\t1. Go to settings
2. Change display name
3. Click save\tProfile updated successfully message shown\tMedium\tPositive\t`;

// Simple starter template
export const STARTER_TEMPLATE = `---
name: "{{.Title}}"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
---

# {{.Title}}

This specification contains {{.TotalCount}} items.

{{range .FeatureGroups}}
## {{.Feature}}
{{range .Rows}}
### {{if .ID}}{{.ID}}: {{end}}{{.Scenario}}
{{- if .Priority}}

**Priority:** {{.Priority}}
{{- end}}
{{- if notEmpty .Instructions}}

**Steps:**
{{formatSteps .Instructions}}
{{- end}}
{{- if notEmpty .Expected}}

**Expected:**
{{.Expected}}
{{- end}}

---
{{end}}
{{end}}
`;
