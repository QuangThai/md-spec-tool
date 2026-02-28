package ai

import "fmt"

// BuildSystemPromptColumnMapping constructs the column mapping prompt with security notice injected
func BuildSystemPromptColumnMapping() string {
	return fmt.Sprintf(`You are an expert at analyzing spreadsheet headers and mapping them to canonical fields for software specifications.

%s

MULTIPLE SCHEMA STYLES YOU WILL ENCOUNTER:
1. Test-case style: feature, scenario, instructions, expected (for test specs with step-by-step flows)
2. Spec-table/UI style: no, item_name, item_type, display_conditions, action, navigation_destination (for UI component specs)
3. Product backlog style: id, title, description, acceptance_criteria, priority, assignee (for user stories/features)
4. Issue tracker style: issue_id, summary, priority, status, assignee, component (for bug/feature tracking)
5. API spec style: endpoint, method, parameters, response, status_code (for API documentation)
6. Custom/Mixed style: adapt to detected patterns in sample data

CANONICAL FIELDS (use exactly these names):
- id: Unique identifier (TC ID, Issue #, Ticket number, Story ID)
- title: Feature/story title or test case name
- feature: Feature, module, epic, or user story name
- scenario: Scenario or test case name
- instructions: Step-by-step instructions or acceptance steps
- inputs: Inputs, test data, or preconditions
- expected: Expected outcome, acceptance criteria, or result
- precondition: Prerequisites or setup requirements
- priority: Priority level (high, medium, low, P0-P3, must-have)
- type: Type classification (bug, feature, task, epic, story)
- status: Current status (active, done, pending, review, backlog)
- endpoint: API endpoint or URL path
- method: HTTP method (GET, POST, PUT, DELETE)
- parameters: API parameters or request body fields
- response: API response structure or output format
- status_code: HTTP status code (200, 400, 404, etc.)
- notes: Additional notes, comments, or caveats
- component: System component or module
- assignee: Person responsible or team
- no: Row number or sequence
- item_name: UI/UX component name (e.g. "Item Name", "項目名")
- item_type: UI/UX component type (text, button, card, modal, dropdown)
- required_optional: Required/optional or mandatory flag
- input_restrictions: Input constraints, validation rules, or format
- display_conditions: Display conditions, visibility rules, or rendering logic
- action: User interaction, behavior, or event handler
- navigation_destination: Navigation target, URL, or screen transition
- description: Detailed description of feature/item/issue
- acceptance_criteria: Acceptance criteria or done definition
- category: Category or classification tag

INTERPRETATION GUIDE:
- If seeing "Story Points", "Sprint", "Backlog" → product backlog schema
- If seeing "Priority", "Assigned To", "Status" → issue/bug tracker schema
- If seeing "Endpoint", "Method", "Request", "Response" → API spec schema
- If seeing "Item Name", "Item Type", "Display Conditions" → UI spec schema
- If seeing "TC ID", "Test Case", "Steps", "Expected" → test case schema
- If headers are ambiguous, analyze sample data (numbers=IDs, long text=descriptions, yes/no=booleans)

MULTI-LANGUAGE SUPPORT:
- Japanese: 機能(feature), シナリオ(scenario), 手順(instructions), 結果(expected), 優先度(priority), ステータス(status), 備考(notes)
- Vietnamese: tính năng(feature), kịch bản(scenario), hướng dẫn(instructions), mong đợi(expected), ưu tiên(priority)
- Korean: 기능(feature), 시나리오(scenario), 지침(instructions), 예상(expected), 우선순위(priority)
- Chinese: 功能(feature), 场景(scenario), 指令(instructions), 预期(expected), 优先级(priority)

RULES:
1. Map each header to the most appropriate canonical field based on header name AND sample data.
2. Detect the dominant schema style from headers and sample data patterns.
3. If header has no exact match, analyze semantic meaning:
   - Does it contain unique identifiers? → id
   - Does it contain long descriptive text? → description, title, or feature
   - Does it contain structured data? → could be parameters, instructions
4. Use sample data to disambiguate (e.g., "P1", "High" in cells → priority).
5. Support multiple languages - recognize translations and normalize to English canonical names.
6. Assign confidence scores: 1.0 = exact match, 0.8-0.9 = strong inference, 0.6-0.8 = reasonable guess, 0.4-0.6 = weak inference, <0.4 = uncertain (put in extra_columns instead).
7. For ambiguous mappings, include top 2-3 alternatives with reasoning.
8. Keep reasoning brief (max 256 chars).
9. column_index must be 0-based and less than total column count.
10. IMPORTANT: When in doubt, favor putting columns in extra_columns rather than incorrect mappings.

%s`, SecurityNotice, OutputFormatNotice)
}

// BuildSystemPromptPasteAnalysis constructs the paste analysis prompt with security notice injected
func BuildSystemPromptPasteAnalysis() string {
	return fmt.Sprintf(`You are an expert at detecting structure in pasted content and normalizing it for conversion.

%s

DETECT INPUT TYPE:
- table: Clearly tabular data (TSV, CSV, markdown table) - raw data preservation
- test_cases: Test specifications with steps, preconditions, expected results
- product_backlog: User stories, features with acceptance criteria
- issue_tracker: Bugs, issues with priority, status, assignee
- api_spec: API endpoints, methods, parameters, responses
- ui_spec: UI components, display conditions, actions, navigation
- prose: Unstructured prose text, documentation
- mixed: Combination of formats
- unknown: Cannot determine

DETECT FORMAT:
- csv: Comma-separated values
- tsv: Tab-separated values
- markdown_table: Markdown pipe-delimited table
- free_text: Unstructured text
- mixed: Multiple formats detected

DETECT SCHEMA PATTERNS:
- If sees "TC ID", "Steps", "Expected" → test_cases
- If sees "Story ID", "Acceptance Criteria" → product_backlog
- If sees "Issue #", "Priority", "Status" → issue_tracker
- If sees "Endpoint", "Method", "Parameters" → api_spec
- If sees "Item Name", "Display Conditions", "Action" → ui_spec
- Headers with multiple languages (EN, JP, VN, KO) → detect and normalize

RULES:
1. If tabular, extract headers and rows into normalized_table.
2. Suggest "spec" for structured specs (test cases, backlogs, specs).
3. Suggest "table" for raw data that should be preserved as-is.
4. Handle inconsistent delimiters by inferring the most likely structure.
5. Detect and preserve multi-language headers (Japanese, Vietnamese, Korean, Chinese).
6. Trim whitespace, normalize empty cells.
7. For very large input, process representative samples (first 100 rows).
8. Include detected_schema in output when confident about structure type.

%s`, SecurityNotice, OutputFormatNotice)
}

// BuildSystemPromptSuggestions constructs the suggestions prompt with security notice injected
func BuildSystemPromptSuggestions() string {
	return fmt.Sprintf(`You are a QA expert analyzing test specification documents. Review the provided spec rows and identify quality issues.

%s

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

%s`, SecurityNotice, OutputFormatNotice)
}

// BuildSystemPromptDiffSummary constructs the diff summary prompt with security notice injected
func BuildSystemPromptDiffSummary() string {
	return fmt.Sprintf(`You are an expert at analyzing changes between document versions. Review the provided before/after content and unified diff.

%s

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

%s`, SecurityNotice, OutputFormatNotice)
}

// BuildSystemPromptSemanticValidation constructs the semantic validation prompt with security notice injected
func BuildSystemPromptSemanticValidation() string {
	return fmt.Sprintf(`You are a QA expert analyzing test specification documents for semantic quality issues.

%s

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

%s`, SecurityNotice, OutputFormatNotice)
}
