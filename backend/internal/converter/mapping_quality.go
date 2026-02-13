package converter

type mappingQuality struct {
	Score        float64
	HeaderScore  float64
	MappedRatio  float64
	CoreCoverage float64
	CoreMapped   int
}

// Core fields for test-case schema (Feature, Scenario, Instructions, Expected)
var coreFieldsTestCase = []CanonicalField{
	FieldScenario, FieldFeature, FieldInstructions, FieldExpected,
}

// Core fields for spec-table schema (No, ItemName, ItemType, DisplayConditions, Action, NavigationDest)
var coreFieldsSpecTable = []CanonicalField{
	FieldNo, FieldItemName, FieldItemType,
	FieldDisplayConditions, FieldAction, FieldNavigationDest,
}

// Core fields for product backlog schema (ID, Title, Description, Acceptance Criteria)
var coreFieldsBacklog = []CanonicalField{
	FieldID, FieldTitle, FieldDescription, FieldAcceptance,
}

// Core fields for API schema (Endpoint, Method, Parameters, Response, Status Code)
var coreFieldsAPI = []CanonicalField{
	FieldEndpoint, FieldMethod, FieldParameters, FieldResponse, FieldStatusCode,
}

// Core fields for issue tracker schema (ID, Title, Priority, Status)
var coreFieldsIssue = []CanonicalField{
	FieldID, FieldTitle, FieldPriority, FieldStatus,
}

func evaluateMappingQuality(headerConfidence int, headers []string, colMap ColumnMap) mappingQuality {
	totalColumns := len(headers)
	if totalColumns <= 0 {
		return mappingQuality{}
	}

	countMapped := func(fields []CanonicalField) int {
		mapped := 0
		for _, field := range fields {
			if _, ok := colMap[field]; ok {
				mapped++
			}
		}
		return mapped
	}

	testCaseMapped := countMapped(coreFieldsTestCase)
	specTableMapped := countMapped(coreFieldsSpecTable)
	backlogMapped := countMapped(coreFieldsBacklog)
	apiMapped := countMapped(coreFieldsAPI)
	issueMapped := countMapped(coreFieldsIssue)

	coreMapped := testCaseMapped
	if specTableMapped > coreMapped {
		coreMapped = specTableMapped
	}
	if backlogMapped > coreMapped {
		coreMapped = backlogMapped
	}
	if apiMapped > coreMapped {
		coreMapped = apiMapped
	}
	if issueMapped > coreMapped {
		coreMapped = issueMapped
	}

	coreCoverage := float64(testCaseMapped) / float64(len(coreFieldsTestCase))
	specTableCoverage := float64(specTableMapped) / float64(len(coreFieldsSpecTable))
	backlogCoverage := float64(backlogMapped) / float64(len(coreFieldsBacklog))
	apiCoverage := float64(apiMapped) / float64(len(coreFieldsAPI))
	issueCoverage := float64(issueMapped) / float64(len(coreFieldsIssue))
	if specTableCoverage > coreCoverage {
		coreCoverage = specTableCoverage
	}
	if backlogCoverage > coreCoverage {
		coreCoverage = backlogCoverage
	}
	if apiCoverage > coreCoverage {
		coreCoverage = apiCoverage
	}
	if issueCoverage > coreCoverage {
		coreCoverage = issueCoverage
	}

	headerScore := float64(headerConfidence) / 100.0
	mappedRatio := float64(len(colMap)) / float64(totalColumns)
	score := (headerScore * 0.35) + (mappedRatio * 0.40) + (coreCoverage * 0.25)
	if score > 1 {
		score = 1
	}

	return mappingQuality{
		Score:        score,
		HeaderScore:  headerScore,
		MappedRatio:  mappedRatio,
		CoreCoverage: coreCoverage,
		CoreMapped:   coreMapped,
	}
}

func shouldFallbackToTable(format string, quality mappingQuality) bool {
	if format != string(OutputFormatSpec) {
		return false
	}
	// Fallback when no core fields mapped (either schema)
	if quality.CoreMapped == 0 {
		return true
	}
	// No fallback when we have core fields and reasonable mapped ratio (e.g. 2/4 columns)
	// This distinguishes "Case, Result Signal, Current State, Notes" with expected+status from
	// generic "A,B,C,D" with only Endpoint mapped
	if quality.MappedRatio >= 0.5 && quality.CoreMapped >= 1 {
		return false
	}
	// Fallback when core coverage is weak and overall score is low
	// (1/4 generic columns like "A,B,C,D" + Endpoint has MappedRatio 0.25, Score ~0.22)
	if quality.CoreCoverage < 0.34 && quality.Score < 0.45 {
		return true
	}
	// No fallback when score is good
	if quality.Score >= 0.45 {
		return false
	}
	// Fallback when mapped ratio is very low with weak score
	if quality.MappedRatio < 0.25 {
		return true
	}
	return false
}
