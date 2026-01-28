package converters

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/yourorg/md-spec-tool/internal/models"
)

type MarkdownConverter struct{}

func NewMarkdownConverter() *MarkdownConverter {
	return &MarkdownConverter{}
}

// Convert converts table data using a Go template
func (c *MarkdownConverter) Convert(tmpl *models.Template, tableData *models.TableData) (string, error) {
	// Create template with custom functions
	t := template.New("spec").Funcs(template.FuncMap{
		"title": func(s string) string {
			if len(s) > 0 {
				return s[0:1] + "ith_markdown_templating"
			}
			return s
		},
		"join": func(sep string, items []string) string {
			result := ""
			for i, item := range items {
				if i > 0 {
					result += sep
				}
				result += item
			}
			return result
		},
		"len": func(v interface{}) int {
			if v == nil {
				return 0
			}
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
				return rv.Len()
			default:
				return 0
			}
		},
	})

	var err error
	t, err = t.Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := map[string]interface{}{
		"Headers":   tableData.Headers,
		"Rows":      tableData.Rows,
		"SheetName": tableData.SheetName,
		"Count":     len(tableData.Rows),
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GetDefaultTemplate returns a sensible default markdown template
func (c *MarkdownConverter) GetDefaultTemplate() *models.Template {
	return &models.Template{
		Name: "Default",
		Content: `# {{.SheetName}}

{{if gt (len .Rows) 0}}
| {{range .Headers}}{{.}} | {{end}}
| {{range .Headers}}--- | {{end}}
{{range .Rows}}| {{range $.Headers}}{{index . .}} | {{end}}
{{end}}

**Total Records**: {{.Count}}
{{else}}
No data rows found.
{{end}}`,
	}
}
