package converter

import (
	"strings"
)

// HeaderSynonyms maps common header names to canonical fields
var HeaderSynonyms = map[string]CanonicalField{
	// ID
	"id":        FieldID,
	"tc_id":     FieldID,
	"tc id":     FieldID,
	"test_id":   FieldID,
	"test id":   FieldID,
	"case_id":   FieldID,
	"case id":   FieldID,
	"ref":       FieldID,
	"reference": FieldID,

	// Feature
	"feature":     FieldFeature,
	"req":         FieldFeature,
	"requirement": FieldFeature,
	"story":       FieldFeature,
	"user story":  FieldFeature,
	"task":        FieldFeature,
	"title":       FieldFeature,
	"name":        FieldFeature,
	"module":      FieldFeature,

	// Scenario
	"scenario":  FieldScenario,
	"test case": FieldScenario,
	"tc":        FieldScenario,
	"test_case": FieldScenario,
	"test name": FieldScenario,
	"test_name": FieldScenario,
	"case":      FieldScenario,
	"case_name": FieldScenario,
	"case name": FieldScenario,

	// Instructions
	"instructions": FieldInstructions,
	"description":  FieldInstructions,
	"steps":        FieldInstructions,
	"test steps":   FieldInstructions,
	"test_steps":   FieldInstructions,
	"action":       FieldInstructions,
	"actions":      FieldInstructions,
	"procedure":    FieldInstructions,

	// Inputs
	"inputs":     FieldInputs,
	"input":      FieldInputs,
	"test data":  FieldInputs,
	"test_data":  FieldInputs,
	"testdata":   FieldInputs,
	"data":       FieldInputs,
	"parameters": FieldInputs,
	"params":     FieldInputs,

	// Expected
	"expected":            FieldExpected,
	"expected output":     FieldExpected,
	"expected_output":     FieldExpected,
	"expected result":     FieldExpected,
	"expected_result":     FieldExpected,
	"acceptance":          FieldExpected,
	"acceptance criteria": FieldExpected,
	"acceptance_criteria": FieldExpected,
	"result":              FieldExpected,
	"outcome":             FieldExpected,

	// Precondition
	"precondition":  FieldPrecondition,
	"preconditions": FieldPrecondition,
	"pre-condition": FieldPrecondition,
	"pre":           FieldPrecondition,
	"given":         FieldPrecondition,
	"prerequisites": FieldPrecondition,
	"setup":         FieldPrecondition,

	// Priority
	"priority": FieldPriority,
	"prio":     FieldPriority,
	"p":        FieldPriority,
	"severity": FieldPriority,
	"sev":      FieldPriority,

	// Type
	"type":      FieldType,
	"category":  FieldType,
	"test type": FieldType,
	"test_type": FieldType,
	"kind":      FieldType,

	// Status
	"status": FieldStatus,
	"state":  FieldStatus,

	// Endpoint
	"endpoint":     FieldEndpoint,
	"api":          FieldEndpoint,
	"api/endpoint": FieldEndpoint,
	"url":          FieldEndpoint,
	"route":        FieldEndpoint,
	"path":         FieldEndpoint,

	// Notes
	"notes":    FieldNotes,
	"note":     FieldNotes,
	"comments": FieldNotes,
	"comment":  FieldNotes,
	"remarks":  FieldNotes,
	"remark":   FieldNotes,

	// === No / Number (Phase 3) ===
	"no":  FieldNo,
	"no.": FieldNo,
	"#":   FieldNo,
	"番号":  FieldNo,
	"連番":  FieldNo,

	// === Item Name (Phase 3) ===
	"item name": FieldItemName,
	"item_name": FieldItemName,
	"項目名":       FieldItemName,
	"項目名称":      FieldItemName,
	"名称":        FieldItemName,

	// === Item Type (Phase 3) ===
	"item type": FieldItemType,
	"item_type": FieldItemType,
	"種別":        FieldItemType,
	"タイプ":       FieldItemType,
	"項目種別":      FieldItemType,

	// === Required/Optional (Phase 3) ===
	"required/optional": FieldRequiredOptional,
	"required":          FieldRequiredOptional,
	"optional":          FieldRequiredOptional,
	"必須":                FieldRequiredOptional,
	"任意":                FieldRequiredOptional,
	"必須/任意":             FieldRequiredOptional,
	"要否":                FieldRequiredOptional,

	// === Input Restrictions (Phase 3) ===
	"input constraint":   FieldInputRestrictions,
	"input constraints":  FieldInputRestrictions,
	"input restrictions": FieldInputRestrictions,
	"input_restrictions": FieldInputRestrictions,
	"入力制限":               FieldInputRestrictions,
	"制約":                 FieldInputRestrictions,
	"バリデーション":            FieldInputRestrictions,
	"validation":         FieldInputRestrictions,

	// === Display Conditions (Phase 3) ===
	"display condition":  FieldDisplayConditions,
	"display conditions": FieldDisplayConditions,
	"display_conditions": FieldDisplayConditions,
	"表示条件":               FieldDisplayConditions,
	"表示制御":               FieldDisplayConditions,
	"条件":                 FieldDisplayConditions,

	// === Action (Phase 3) ===
	"操作":       FieldAction,
	"アクション":    FieldAction,
	"イベント":     FieldAction,
	"on click": FieldAction,

	// === Navigation Destination (Phase 3) ===
	"navigation destination": FieldNavigationDest,
	"navigation_destination": FieldNavigationDest,
	"遷移先":                    FieldNavigationDest,
	"画面遷移先":                  FieldNavigationDest,
	"遷移先画面":                  FieldNavigationDest,
	"destination":            FieldNavigationDest,
	"link":                   FieldNavigationDest,
}

// ColumnMapper maps headers to canonical fields
type ColumnMapper struct{}

// NewColumnMapper creates a new ColumnMapper
func NewColumnMapper() *ColumnMapper {
	return &ColumnMapper{}
}

// MapColumns analyzes headers and returns column mapping
func (m *ColumnMapper) MapColumns(headers []string) (ColumnMap, []string) {
	colMap := make(ColumnMap)
	var unmapped []string

	for i, header := range headers {
		normalized := m.normalizeHeader(header)
		if field, ok := HeaderSynonyms[normalized]; ok {
			// Only map if not already mapped (first occurrence wins)
			if _, exists := colMap[field]; !exists {
				colMap[field] = i
			}
		} else {
			unmapped = append(unmapped, header)
		}
	}

	return colMap, unmapped
}

// normalizeHeader converts a header to lowercase and trims whitespace
func (m *ColumnMapper) normalizeHeader(header string) string {
	// Convert to lowercase
	h := strings.ToLower(header)
	// Remove extra whitespace
	h = strings.TrimSpace(h)
	// Replace multiple spaces with single space
	h = strings.Join(strings.Fields(h), " ")
	return h
}

// GetFieldValue extracts a field value from a row using the column map
func GetFieldValue(row []string, colMap ColumnMap, field CanonicalField) string {
	if idx, ok := colMap[field]; ok && idx < len(row) {
		return row[idx]
	}
	return ""
}
