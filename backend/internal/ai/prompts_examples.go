package ai

// ColumnMappingExample represents an example for few-shot learning
type ColumnMappingExample struct {
	Headers  []string
	Expected []CanonicalFieldMapping
}

// Few-shot examples for column mapping
// These examples teach the model how to map common header styles to canonical fields
var ColumnMappingExamples = []ColumnMappingExample{
	// Test Case Style - English
	{
		Headers: []string{"TC ID", "Test Case Name", "Precondition", "Steps", "Expected", "Status"},
		Expected: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "TC ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Test case identifier", Alternatives: nil},
			{CanonicalName: "title", SourceHeader: "Test Case Name", ColumnIndex: 1, Confidence: 0.95, Reasoning: "Test case name is title, not feature", Alternatives: []AlternativeColumn{{SourceHeader: "scenario", ColumnIndex: 1, Confidence: 0.85}}},
			{CanonicalName: "precondition", SourceHeader: "Precondition", ColumnIndex: 2, Confidence: 1.0, Reasoning: "Exact match", Alternatives: nil},
			{CanonicalName: "instructions", SourceHeader: "Steps", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Exact match", Alternatives: nil},
			{CanonicalName: "expected", SourceHeader: "Expected", ColumnIndex: 4, Confidence: 0.9, Reasoning: "Expected result abbreviation", Alternatives: nil},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 5, Confidence: 1.0, Reasoning: "Exact match", Alternatives: nil},
		},
	},
	// Issue Tracker Style - Japanese
	{
		Headers: []string{"Issue #", "概要", "優先度", "担当者", "備考"},
		Expected: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Issue #", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Issue identifier", Alternatives: nil},
			{CanonicalName: "feature", SourceHeader: "概要", ColumnIndex: 1, Confidence: 0.95, Reasoning: "JP: summary/overview", Alternatives: []AlternativeColumn{{SourceHeader: "title", ColumnIndex: 1, Confidence: 0.85}, {SourceHeader: "description", ColumnIndex: 1, Confidence: 0.80}}},
			{CanonicalName: "priority", SourceHeader: "優先度", ColumnIndex: 2, Confidence: 1.0, Reasoning: "JP: priority (exact match)", Alternatives: nil},
			{CanonicalName: "assignee", SourceHeader: "担当者", ColumnIndex: 3, Confidence: 1.0, Reasoning: "JP: assignee/owner", Alternatives: nil},
			{CanonicalName: "notes", SourceHeader: "備考", ColumnIndex: 4, Confidence: 0.9, Reasoning: "JP: remarks/notes", Alternatives: nil},
		},
	},
	// UI Spec Table Style
	{
		Headers: []string{"No", "Item Name", "Item Type", "Required/Optional", "Input Restrictions", "Display Conditions", "Action", "Navigation Destination"},
		Expected: []CanonicalFieldMapping{
			{CanonicalName: "no", SourceHeader: "No", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Row number", Alternatives: nil},
			{CanonicalName: "item_name", SourceHeader: "Item Name", ColumnIndex: 1, Confidence: 1.0, Reasoning: "UI item name", Alternatives: nil},
			{CanonicalName: "item_type", SourceHeader: "Item Type", ColumnIndex: 2, Confidence: 1.0, Reasoning: "UI item type", Alternatives: nil},
			{CanonicalName: "required_optional", SourceHeader: "Required/Optional", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Required/optional flag", Alternatives: nil},
			{CanonicalName: "input_restrictions", SourceHeader: "Input Restrictions", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Input constraints", Alternatives: nil},
			{CanonicalName: "display_conditions", SourceHeader: "Display Conditions", ColumnIndex: 5, Confidence: 1.0, Reasoning: "Display/visibility rules", Alternatives: nil},
			{CanonicalName: "action", SourceHeader: "Action", ColumnIndex: 6, Confidence: 1.0, Reasoning: "User interaction/behavior", Alternatives: nil},
			{CanonicalName: "navigation_destination", SourceHeader: "Navigation Destination", ColumnIndex: 7, Confidence: 1.0, Reasoning: "Navigation target", Alternatives: nil},
		},
	},
	// Product Backlog Style
	{
		Headers: []string{"Story ID", "Title", "Description", "Acceptance Criteria", "Priority", "Story Points"},
		Expected: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Story ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "User story identifier", Alternatives: nil},
			{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 1, Confidence: 1.0, Reasoning: "Story title/summary", Alternatives: nil},
			{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 1.0, Reasoning: "Detailed description", Alternatives: nil},
			{CanonicalName: "acceptance_criteria", SourceHeader: "Acceptance Criteria", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Done definition criteria", Alternatives: nil},
			{CanonicalName: "priority", SourceHeader: "Priority", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Story priority", Alternatives: nil},
		},
	},
	// Extra Columns Example - Unknown/Non-Standard Fields (demonstrates extra_columns behavior)
	{
		Headers: []string{"ID", "Task", "Effort Hours", "Blocker"},
		Expected: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95, Reasoning: "Generic identifier", Alternatives: nil},
			{CanonicalName: "title", SourceHeader: "Task", ColumnIndex: 1, Confidence: 0.9, Reasoning: "Task name → title", Alternatives: nil},
		},
	},
	// API Specification Style
	{
		Headers: []string{"Endpoint", "Method", "Description", "Parameters", "Response", "Status Code"},
		Expected: []CanonicalFieldMapping{
			{CanonicalName: "endpoint", SourceHeader: "Endpoint", ColumnIndex: 0, Confidence: 1.0, Reasoning: "API endpoint URL path", Alternatives: nil},
			{CanonicalName: "method", SourceHeader: "Method", ColumnIndex: 1, Confidence: 1.0, Reasoning: "HTTP method (GET, POST, etc.)", Alternatives: nil},
			{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 0.95, Reasoning: "Endpoint description", Alternatives: nil},
			{CanonicalName: "parameters", SourceHeader: "Parameters", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Request parameters/body", Alternatives: nil},
			{CanonicalName: "response", SourceHeader: "Response", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Response structure", Alternatives: nil},
			{CanonicalName: "status_code", SourceHeader: "Status Code", ColumnIndex: 5, Confidence: 1.0, Reasoning: "HTTP status code", Alternatives: nil},
		},
	},
	// Mixed/Generic Headers with Context Clues
	{
		Headers: []string{"ID", "Name", "Type", "Status", "Owner", "Notes"},
		Expected: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95, Reasoning: "Generic identifier", Alternatives: nil},
			{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.8, Reasoning: "Could be title or feature name (analyze sample data)", Alternatives: []AlternativeColumn{{SourceHeader: "feature", ColumnIndex: 1, Confidence: 0.75}, {SourceHeader: "description", ColumnIndex: 1, Confidence: 0.70}}},
			{CanonicalName: "type", SourceHeader: "Type", ColumnIndex: 2, Confidence: 0.85, Reasoning: "Classification (bug/feature/task)", Alternatives: []AlternativeColumn{{SourceHeader: "item_type", ColumnIndex: 2, Confidence: 0.75}, {SourceHeader: "category", ColumnIndex: 2, Confidence: 0.70}}},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Exact match", Alternatives: nil},
			{CanonicalName: "assignee", SourceHeader: "Owner", ColumnIndex: 4, Confidence: 0.95, Reasoning: "Owner/responsible person", Alternatives: nil},
			{CanonicalName: "notes", SourceHeader: "Notes", ColumnIndex: 5, Confidence: 1.0, Reasoning: "Exact match", Alternatives: nil},
		},
	},
}

// PasteAnalysisExample represents an example for paste analysis
type PasteAnalysisExample struct {
	Input    string
	Expected PasteAnalysis
}

// Few-shot examples for paste analysis
var PasteAnalysisExamples = []PasteAnalysisExample{
	{
		Input: "TC ID\tTest Case Name\tExpected\nTC001\tLogin with valid credentials\tUser logged in",
		Expected: PasteAnalysis{
			InputType:      "test_cases",
			DetectedFormat: "tsv",
			NormalizedTable: [][]string{
				{"TC ID", "Test Case Name", "Expected"},
				{"TC001", "Login with valid credentials", "User logged in"},
			},
			SuggestedOutput: "spec",
			Confidence:      0.95,
		},
	},
}
