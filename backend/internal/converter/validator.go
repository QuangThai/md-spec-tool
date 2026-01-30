package converter

import (
	"regexp"
	"strings"
)

// ValidationRules holds user-configurable validation rules
type ValidationRules struct {
	RequiredFields []string          `json:"required_fields"` // canonical field names, e.g. "id", "feature", "expected"
	FormatRules    *FormatRules      `json:"format_rules,omitempty"`
	CrossField     []CrossFieldRule  `json:"cross_field,omitempty"`
}

// FormatRules define format validation per field
type FormatRules struct {
	IDPattern   string `json:"id_pattern,omitempty"`   // regex for ID field
	DateFormat  string `json:"date_format,omitempty"` // e.g. "2006-01-02"
	EmailFields []string `json:"email_fields,omitempty"` // field names to validate as email
	URLFields   []string `json:"url_fields,omitempty"`   // field names to validate as URL
}

// CrossFieldRule defines a rule like "if field A present then field B required"
type CrossFieldRule struct {
	IfField   string `json:"if_field"`   // when this field is non-empty
	ThenField string `json:"then_field"` // this field must be non-empty
	Message   string `json:"message,omitempty"`
}

// ValidationResult holds validation errors/warnings
type ValidationResult struct {
	Valid    bool      `json:"valid"`
	Warnings []Warning `json:"warnings"`
}

// Validate runs custom validation rules against a SpecDoc and returns validation warnings
func Validate(doc *SpecDoc, rules *ValidationRules) ValidationResult {
	if doc == nil || rules == nil {
		return ValidationResult{Valid: true, Warnings: nil}
	}

	var warnings []Warning
	requiredSet := make(map[string]bool)
	for _, f := range rules.RequiredFields {
		requiredSet[strings.TrimSpace(strings.ToLower(f))] = true
	}

	for i, row := range doc.Rows {
		rowNum := i + 1
		// Required fields
		for field := range requiredSet {
			val := getFieldValue(&row, field)
			if strings.TrimSpace(val) == "" {
				warnings = append(warnings, newWarning(
					"VALIDATION_REQUIRED",
					SeverityWarn,
					CatRows,
					"Required field \""+field+"\" is empty",
					"Fill in the required field or disable this rule.",
					map[string]any{"row": rowNum, "field": field},
				))
			}
		}

		// Format rules
		if rules.FormatRules != nil {
			if rules.FormatRules.IDPattern != "" {
				idVal := getFieldValue(&row, "id")
				if idVal != "" {
					re, err := regexp.Compile(rules.FormatRules.IDPattern)
					if err == nil && !re.MatchString(idVal) {
						warnings = append(warnings, newWarning(
							"VALIDATION_ID_FORMAT",
							SeverityWarn,
							CatRows,
							"ID does not match pattern: "+idVal,
							"Use a valid ID format (regex: "+rules.FormatRules.IDPattern+").",
							map[string]any{"row": rowNum, "value": idVal},
						))
					}
				}
			}
			for _, f := range rules.FormatRules.EmailFields {
				val := getFieldValue(&row, f)
				if val != "" && !matchEmail(val) {
					warnings = append(warnings, newWarning(
						"VALIDATION_EMAIL",
						SeverityWarn,
						CatRows,
							"Field \""+f+"\" is not a valid email: "+val,
						"Enter a valid email address.",
						map[string]any{"row": rowNum, "field": f, "value": val},
					))
				}
			}
			for _, f := range rules.FormatRules.URLFields {
				val := getFieldValue(&row, f)
				if val != "" && !matchURL(val) {
					warnings = append(warnings, newWarning(
						"VALIDATION_URL",
						SeverityWarn,
						CatRows,
						"Field \""+f+"\" is not a valid URL: "+val,
						"Enter a valid URL.",
						map[string]any{"row": rowNum, "field": f, "value": val},
					))
				}
			}
		}
	}

	// Cross-field rules: if A is set then B required
	for _, cf := range rules.CrossField {
		for ri, row := range doc.Rows {
			rowNum := ri + 1
			ifV := getFieldValue(&row, cf.IfField)
			thenV := getFieldValue(&row, cf.ThenField)
			if strings.TrimSpace(ifV) != "" && strings.TrimSpace(thenV) == "" {
				msg := cf.Message
				if msg == "" {
					msg = "When \"" + cf.IfField + "\" is set, \"" + cf.ThenField + "\" is required"
				}
				warnings = append(warnings, newWarning(
					"VALIDATION_CROSS_FIELD",
					SeverityWarn,
					CatRows,
					msg,
					"Fill in \""+cf.ThenField+"\" or clear \""+cf.IfField+"\".",
					map[string]any{"row": rowNum, "if_field": cf.IfField, "then_field": cf.ThenField},
				))
			}
		}
	}

	return ValidationResult{
		Valid:    len(warnings) == 0,
		Warnings: warnings,
	}
}

func getFieldValue(row *SpecRow, field string) string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "id":
		return row.ID
	case "feature":
		return row.Feature
	case "scenario":
		return row.Scenario
	case "instructions":
		return row.Instructions
	case "inputs":
		return row.Inputs
	case "expected":
		return row.Expected
	case "precondition":
		return row.Precondition
	case "priority":
		return row.Priority
	case "type":
		return row.Type
	case "status":
		return row.Status
	case "endpoint":
		return row.Endpoint
	case "notes":
		return row.Notes
	case "no":
		return row.No
	case "item_name":
		return row.ItemName
	case "item_type":
		return row.ItemType
	case "required_optional":
		return row.RequiredOptional
	case "input_restrictions":
		return row.InputRestrictions
	case "display_conditions":
		return row.DisplayConditions
	case "action":
		return row.Action
	case "navigation_destination":
		return row.NavigationDest
	default:
		if row.Metadata != nil {
			return row.Metadata[field]
		}
		return ""
	}
}

func matchEmail(s string) bool {
	// Simple email pattern: something@something.something
	re := regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)
	return re.MatchString(strings.TrimSpace(s))
}

func matchURL(s string) bool {
	trimmed := strings.TrimSpace(s)
	return strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://")
}
