package converter

import (
	"strings"
)

type mappingCandidate struct {
	field CanonicalField
	score float64
}

// enhanceColumnMapping fills obvious gaps for unknown schemas by combining
// header and sample-value heuristics. It only maps high-confidence columns.
func enhanceColumnMapping(headers []string, dataRows [][]string, current ColumnMap) (ColumnMap, []string, []Warning) {
	if len(headers) == 0 {
		return current, nil, nil
	}

	enhanced := cloneColumnMap(current)
	usedColumns := make(map[int]bool, len(enhanced))
	for _, idx := range enhanced {
		usedColumns[idx] = true
	}

	availableFields := make(map[CanonicalField]bool)
	for _, field := range allCanonicalFields() {
		if _, exists := enhanced[field]; !exists {
			availableFields[field] = true
		}
	}

	inferred := 0
	for idx, header := range headers {
		if usedColumns[idx] {
			continue
		}

		best, second := inferBestField(header, idx, dataRows, availableFields)
		if best.field == "" {
			continue
		}

		if best.score < 0.62 {
			continue
		}
		if second.score > 0 && best.score-second.score < 0.12 {
			continue
		}

		enhanced[best.field] = idx
		usedColumns[idx] = true
		delete(availableFields, best.field)
		inferred++
	}

	unmapped := collectUnmappedHeaders(headers, enhanced)
	if inferred == 0 {
		return enhanced, unmapped, nil
	}

	return enhanced, unmapped, []Warning{newWarning(
		"MAPPING_DYNAMIC_INFERENCE",
		SeverityInfo,
		CatMapping,
		"Applied dynamic schema inference for non-standard columns.",
		"Review inferred mappings in preview if output looks unusual.",
		map[string]any{"inferred_columns": inferred},
	)}
}

func inferBestField(header string, colIdx int, dataRows [][]string, available map[CanonicalField]bool) (mappingCandidate, mappingCandidate) {
	stats := collectColumnStats(dataRows, colIdx)
	normalized := normalizeHeader(header)

	best := mappingCandidate{}
	second := mappingCandidate{}
	for _, field := range allCanonicalFields() {
		if !available[field] {
			continue
		}

		score := scoreCandidate(field, normalized, stats)
		candidate := mappingCandidate{field: field, score: score}
		if candidate.score > best.score {
			second = best
			best = candidate
		} else if candidate.score > second.score {
			second = candidate
		}
	}

	return best, second
}

type columnStats struct {
	samples         int
	urlRatio        float64
	numericRatio    float64
	longTextRatio   float64
	statusRatio     float64
	priorityRatio   float64
	requiredRatio   float64
	actionRatio     float64
	noteRatio       float64
	methodRatio     float64
	statusCodeRatio float64
}

func collectColumnStats(rows [][]string, colIdx int) columnStats {
	stats := columnStats{}
	for _, row := range rows {
		if colIdx >= len(row) {
			continue
		}
		value := strings.TrimSpace(row[colIdx])
		if value == "" || value == "-" {
			continue
		}

		stats.samples++
		lower := strings.ToLower(value)

		if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "/") {
			stats.urlRatio += 1
		}
		if isLikelyNumeric(lower) {
			stats.numericRatio += 1
		}
		if len(value) >= 24 {
			stats.longTextRatio += 1
		}
		if isStatusLike(lower) {
			stats.statusRatio += 1
		}
		if isPriorityLike(lower) {
			stats.priorityRatio += 1
		}
		if isRequiredLike(lower) {
			stats.requiredRatio += 1
		}
		if isActionLike(lower) {
			stats.actionRatio += 1
		}
		if strings.Contains(lower, "note") || strings.Contains(lower, "remark") || strings.Contains(lower, "comment") {
			stats.noteRatio += 1
		}
		if isHTTPMethodLike(lower) {
			stats.methodRatio += 1
		}
		if isStatusCodeLike(lower) {
			stats.statusCodeRatio += 1
		}
	}

	if stats.samples == 0 {
		return stats
	}

	denom := float64(stats.samples)
	stats.urlRatio /= denom
	stats.numericRatio /= denom
	stats.longTextRatio /= denom
	stats.statusRatio /= denom
	stats.priorityRatio /= denom
	stats.requiredRatio /= denom
	stats.actionRatio /= denom
	stats.noteRatio /= denom
	stats.methodRatio /= denom
	stats.statusCodeRatio /= denom
	return stats
}

func scoreCandidate(field CanonicalField, header string, stats columnStats) float64 {
	score := 0.0

	if hasHeaderKeyword(header, fieldHeaderKeywords(field)) {
		score += 0.62
	}

	switch field {
	case FieldEndpoint:
		score += stats.urlRatio * 0.45
	case FieldMethod:
		score += stats.methodRatio * 0.45
	case FieldStatusCode:
		score += stats.statusCodeRatio * 0.45
	case FieldStatus:
		score += stats.statusRatio * 0.35
	case FieldPriority:
		score += stats.priorityRatio * 0.35
	case FieldRequiredOptional:
		score += stats.requiredRatio * 0.40
	case FieldAction:
		score += stats.actionRatio * 0.35
	case FieldNo, FieldID:
		score += stats.numericRatio * 0.35
	case FieldInstructions, FieldExpected, FieldDisplayConditions, FieldInputRestrictions:
		score += stats.longTextRatio * 0.20
	case FieldDescription, FieldAcceptance, FieldParameters, FieldResponse:
		score += stats.longTextRatio * 0.20
	case FieldNotes:
		score += stats.noteRatio * 0.40
	}

	if score > 1 {
		return 1
	}
	return score
}

func fieldHeaderKeywords(field CanonicalField) []string {
	switch field {
	case FieldID:
		return []string{"id", "case id", "ticket", "ref", "reference"}
	case FieldTitle:
		return []string{"title", "summary", "name"}
	case FieldDescription:
		return []string{"description", "details", "overview"}
	case FieldAcceptance:
		return []string{"acceptance", "criteria", "definition of done"}
	case FieldFeature:
		return []string{"feature", "module", "story", "requirement"}
	case FieldScenario:
		return []string{"scenario", "case", "flow", "usecase", "test"}
	case FieldInstructions:
		return []string{"step", "instruction", "procedure", "how to", "action detail"}
	case FieldInputs:
		return []string{"input", "data", "parameter", "param", "payload"}
	case FieldExpected:
		return []string{"expected", "result", "outcome", "success", "acceptance", "criteria"}
	case FieldPrecondition:
		return []string{"precondition", "pre-condition", "prerequisite", "setup", "given"}
	case FieldPriority:
		return []string{"priority", "severity", "impact", "urgency"}
	case FieldType:
		return []string{"type", "kind", "class"}
	case FieldStatus:
		return []string{"status", "state", "progress"}
	case FieldEndpoint:
		return []string{"endpoint", "api", "url", "route", "path", "uri"}
	case FieldMethod:
		return []string{"method", "http method", "verb"}
	case FieldParameters:
		return []string{"parameters", "params", "request", "payload"}
	case FieldResponse:
		return []string{"response", "response body", "response schema"}
	case FieldStatusCode:
		return []string{"status code", "http status"}
	case FieldNotes:
		return []string{"note", "remark", "comment", "memo"}
	case FieldComponent:
		return []string{"component", "module", "area"}
	case FieldAssignee:
		return []string{"assignee", "owner", "assigned"}
	case FieldCategory:
		return []string{"category", "type", "kind"}
	case FieldNo:
		return []string{"no", "number", "seq", "index"}
	case FieldItemName:
		return []string{"item", "field", "label", "name"}
	case FieldItemType:
		return []string{"item type", "field type", "control", "widget"}
	case FieldRequiredOptional:
		return []string{"required", "optional", "mandatory", "required/optional"}
	case FieldInputRestrictions:
		return []string{"restriction", "constraint", "validation", "rule"}
	case FieldDisplayConditions:
		return []string{"display", "visibility", "condition", "show when"}
	case FieldAction:
		return []string{"action", "trigger", "event", "on click"}
	case FieldNavigationDest:
		return []string{"navigation", "destination", "target", "redirect", "next screen"}
	default:
		return nil
	}
}

func hasHeaderKeyword(header string, keywords []string) bool {
	if len(keywords) == 0 {
		return false
	}
	for _, keyword := range keywords {
		if strings.Contains(header, keyword) {
			return true
		}
	}
	return false
}

func isLikelyNumeric(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return value != ""
}

func isHTTPMethodLike(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD":
		return true
	default:
		return false
	}
}

func isStatusCodeLike(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return value >= "100" && value <= "599"
}

func isStatusLike(value string) bool {
	statusValues := []string{"new", "open", "wip", "in progress", "done", "closed", "passed", "failed", "draft", "active"}
	for _, v := range statusValues {
		if value == v {
			return true
		}
	}
	return false
}

func isPriorityLike(value string) bool {
	priorityValues := []string{"p0", "p1", "p2", "p3", "high", "medium", "low", "critical", "blocker", "minor"}
	for _, v := range priorityValues {
		if value == v {
			return true
		}
	}
	return false
}

func isRequiredLike(value string) bool {
	requiredValues := []string{"required", "optional", "must", "yes", "no", "mandatory"}
	for _, v := range requiredValues {
		if value == v {
			return true
		}
	}
	return false
}

func isActionLike(value string) bool {
	actionTokens := []string{"click", "tap", "select", "enter", "submit", "open", "navigate", "press"}
	for _, token := range actionTokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func collectUnmappedHeaders(headers []string, colMap ColumnMap) []string {
	mapped := make(map[int]bool, len(colMap))
	for _, idx := range colMap {
		mapped[idx] = true
	}

	unmapped := make([]string, 0)
	for idx, header := range headers {
		if !mapped[idx] {
			unmapped = append(unmapped, header)
		}
	}
	return unmapped
}

func cloneColumnMap(colMap ColumnMap) ColumnMap {
	if len(colMap) == 0 {
		return make(ColumnMap)
	}
	cloned := make(ColumnMap, len(colMap))
	for field, idx := range colMap {
		cloned[field] = idx
	}
	return cloned
}

func allCanonicalFields() []CanonicalField {
	return []CanonicalField{
		FieldID,
		FieldTitle,
		FieldDescription,
		FieldAcceptance,
		FieldFeature,
		FieldScenario,
		FieldInstructions,
		FieldInputs,
		FieldExpected,
		FieldPrecondition,
		FieldPriority,
		FieldType,
		FieldStatus,
		FieldEndpoint,
		FieldMethod,
		FieldParameters,
		FieldResponse,
		FieldStatusCode,
		FieldNotes,
		FieldComponent,
		FieldAssignee,
		FieldCategory,
		FieldNo,
		FieldItemName,
		FieldItemType,
		FieldRequiredOptional,
		FieldInputRestrictions,
		FieldDisplayConditions,
		FieldAction,
		FieldNavigationDest,
	}
}
