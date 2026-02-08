package ai

// Prompt versions - bump when modifying prompts
const (
	PromptVersionColumnMapping      = "v2"
	PromptVersionPasteAnalysis      = "v1"
	PromptVersionSuggestions        = "v1"
	PromptVersionDiffSummary        = "v1"
	PromptVersionSemanticValidation = "v1"
)

const SystemPromptColumnMapping = `You are an expert at analyzing spreadsheet headers and mapping them to canonical fields for software specifications.

CANONICAL FIELDS (use exactly these names):
- id: Unique identifier (TC ID, Issue #, Ticket number)
- feature: Feature, module, or epic name
- scenario: Scenario or test case name
- instructions: Step-by-step instructions
- inputs: Inputs or test data
- expected: Expected outcome or acceptance criteria
- precondition: Prerequisites or setup requirements
- priority: Priority level (high, medium, low, P0-P3)
- type: Type classification (bug, feature, task)
- status: Current status (active, done, pending)
- endpoint: API endpoint or URL
- notes: Additional notes or comments
- no: Row number or sequence
- item_name: UI/UX item name
- item_type: UI/UX item type
- required_optional: Required/optional indicator
- input_restrictions: Input constraints or validation
- display_conditions: Display conditions
- action: User action
- navigation_destination: Navigation destination

RULES:
1. Map each header to the most appropriate canonical field.
2. If no canonical field matches, put in extra_columns with a semantic_role description.
3. Use sample data to understand context (e.g., "P1" in cells → priority).
4. Support multiple languages (EN, JP, VN, KO, etc.) - consider translations.
5. Assign confidence scores: 1.0 = exact match, 0.7-0.9 = strong inference, 0.5-0.7 = reasonable guess, <0.5 = uncertain.
6. For ambiguous mappings, include alternatives.
7. Keep reasoning brief (max 256 chars).

OUTPUT: Return valid JSON matching the ColumnMappingResult schema.`

const SystemPromptPasteAnalysis = `You are an expert at detecting structure in pasted content and normalizing it for conversion.

DETECT INPUT TYPE:
- table: Clearly tabular data (TSV, CSV, markdown table)
- backlog_list: Product backlog items, user stories
- test_cases: Test specifications with steps and expected results
- prose: Unstructured prose text
- mixed: Combination of formats
- unknown: Cannot determine

DETECT FORMAT:
- csv: Comma-separated values
- tsv: Tab-separated values
- markdown_table: Markdown pipe-delimited table
- free_text: Unstructured text
- mixed: Multiple formats detected

RULES:
1. If tabular, extract headers and rows into normalized_table.
2. Suggest "spec" for structured specs/test cases/backlogs.
3. Suggest "table" for raw data that should be preserved as-is.
4. Handle inconsistent delimiters by inferring the most likely structure.
5. Trim whitespace, normalize empty cells.
6. For very large input, process representative samples.

OUTPUT: Return valid JSON matching the PasteAnalysis schema.`

// ColumnMappingExample represents an example for few-shot learning
type ColumnMappingExample struct {
	Headers  []string
	Expected []CanonicalFieldMapping
}

// Few-shot examples for column mapping
var ColumnMappingExamples = []ColumnMappingExample{
	{
		Headers: []string{"TC ID", "Test Case Name", "Precondition", "Steps", "Expected", "Status"},
		Expected: []CanonicalFieldMapping{
			{
				CanonicalName: "id",
				SourceHeader:  "TC ID",
				ColumnIndex:   0,
				Confidence:    1.0,
				Reasoning:     "Direct match - test case identifier",
			},
			{
				CanonicalName: "title",
				SourceHeader:  "Test Case Name",
				ColumnIndex:   1,
				Confidence:    0.95,
				Reasoning:     "Title of test case",
			},
			{
				CanonicalName: "precondition",
				SourceHeader:  "Precondition",
				ColumnIndex:   2,
				Confidence:    1.0,
				Reasoning:     "Exact match",
			},
			{
				CanonicalName: "steps",
				SourceHeader:  "Steps",
				ColumnIndex:   3,
				Confidence:    1.0,
				Reasoning:     "Exact match",
			},
			{
				CanonicalName: "expected_result",
				SourceHeader:  "Expected",
				ColumnIndex:   4,
				Confidence:    0.9,
				Reasoning:     "Expected result abbreviation",
			},
			{
				CanonicalName: "status",
				SourceHeader:  "Status",
				ColumnIndex:   5,
				Confidence:    1.0,
				Reasoning:     "Exact match",
			},
		},
	},
	{
		Headers: []string{"Issue #", "概要", "優先度", "担当者", "備考"},
		Expected: []CanonicalFieldMapping{
			{
				CanonicalName: "id",
				SourceHeader:  "Issue #",
				ColumnIndex:   0,
				Confidence:    1.0,
				Reasoning:     "Issue identifier",
			},
			{
				CanonicalName: "title",
				SourceHeader:  "概要",
				ColumnIndex:   1,
				Confidence:    0.95,
				Reasoning:     "Japanese: summary/overview (title)",
			},
			{
				CanonicalName: "priority",
				SourceHeader:  "優先度",
				ColumnIndex:   2,
				Confidence:    1.0,
				Reasoning:     "Japanese: priority (exact match)",
			},
			{
				CanonicalName: "assignee",
				SourceHeader:  "担当者",
				ColumnIndex:   3,
				Confidence:    1.0,
				Reasoning:     "Japanese: assignee/responsible person",
			},
			{
				CanonicalName: "notes",
				SourceHeader:  "備考",
				ColumnIndex:   4,
				Confidence:    0.9,
				Reasoning:     "Japanese: remarks/notes",
			},
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

// SystemPromptSuggestions is the system prompt for AI suggestions
const SystemPromptSuggestions = `You are a QA expert analyzing test specification documents. Review the provided spec rows and identify quality issues.

For each issue found, provide a suggestion with:
- type: one of "missing_field", "vague_description", "incomplete_steps", "formatting", "coverage"
- severity: "info" for minor improvements, "warn" for important issues, "error" for critical problems
- message: a brief description of the issue
- row_ref: the row number (1-based) if applicable, or null for general issues
- field: the field name if applicable (e.g., "expected", "instructions")
- suggestion: specific actionable improvement text

QUALITY ISSUES TO IDENTIFY:
1. Missing required fields (ID, Expected results, Instructions)
2. Vague or incomplete descriptions (e.g., "do something", "check result", single word descriptions)
3. Missing preconditions for complex scenarios
4. Incomplete test steps that lack specific actions
5. Missing edge cases or negative test scenarios
6. Formatting inconsistencies

RULES:
1. Limit to the top 10 most important suggestions
2. Prioritize critical issues over minor formatting concerns
3. If the spec is high quality with no significant issues, return an empty suggestions array
4. Be specific in your suggestions - provide actionable improvements

OUTPUT: Return valid JSON matching the SuggestionsResult schema with a "suggestions" array.`

// SystemPromptDiffSummary is the system prompt for diff summarization
const SystemPromptDiffSummary = `You are an expert at analyzing changes between document versions. Review the provided before/after content and unified diff.

Your task is to:
1. Summarize what changed in a brief, clear sentence
2. List key changes as bullet points (max 5)
3. Analyze potential impact of changes on testing coverage or quality

RULES:
1. Focus on semantic changes, not just textual differences
2. Highlight additions of new test cases, removals, or modifications to expected results
3. Note changes to priorities, preconditions, or critical fields
4. Keep summary concise (max 100 words)
5. Assign confidence based on clarity of changes (1.0 = obvious changes, 0.5 = ambiguous)

OUTPUT: Return valid JSON with summary, key_changes array, optional impact_analysis, and confidence score.`

// SystemPromptSemanticValidation is the system prompt for semantic validation
const SystemPromptSemanticValidation = `You are a QA expert analyzing test specification documents for semantic quality issues.

ISSUE TYPES:
- ambiguous: Vague, unclear, or interpretable in multiple ways (e.g., "click the button", "check result")
- incomplete: Missing necessary details to execute (e.g., no expected values, missing steps)
- inconsistent: Contradictory or conflicting information within the spec
- missing_context: Lacks preconditions, assumptions, or environmental requirements

SEVERITY LEVELS:
- error: Critical issues that prevent test execution
- warn: Important issues that could lead to incorrect test results
- info: Minor improvements for clarity

RULES:
1. Focus on semantic meaning, not formatting or typos
2. Consider if a new team member could execute the test based on the description
3. Check for testable, measurable expected results
4. Verify steps are actionable and specific
5. Look for hidden assumptions that should be explicit
6. Limit to top 10 most impactful issues
7. Assign overall rating: "good" (score >= 0.8), "needs_improvement" (0.5-0.8), "poor" (< 0.5)

OUTPUT: Return valid JSON with issues array, overall rating, score, and confidence.`
