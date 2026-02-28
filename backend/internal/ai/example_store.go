package ai

import (
	"fmt"
	"strings"
	"sync"
)

// Example represents a few-shot example for an AI operation
type Example struct {
	Operation  string                  // e.g., "column_mapping", "paste_analysis"
	SchemaType string                  // e.g., "test_case", "ui_spec", "api_spec"
	Language   string                  // e.g., "en", "ja", "vi"
	Headers    []string                // Input headers
	Mappings   []CanonicalFieldMapping // Expected output mappings (for column_mapping)
	// For paste analysis examples
	PasteInput    string         // Input text
	PasteExpected *PasteAnalysis // Expected paste analysis
}

// ExampleFilter controls which examples are returned
type ExampleFilter struct {
	SchemaType string // Filter by schema type
	Language   string // Filter by language
	MaxResults int    // Maximum examples to return (0 = all)
}

// ExampleStore manages few-shot examples for AI operations
type ExampleStore struct {
	mu       sync.RWMutex
	examples map[string][]Example // operation → examples
}

// NewExampleStore creates a new empty example store
func NewExampleStore() *ExampleStore {
	return &ExampleStore{
		examples: make(map[string][]Example),
	}
}

// Register adds an example to the store
func (s *ExampleStore) Register(example Example) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.examples[example.Operation] = append(s.examples[example.Operation], example)
}

// GetExamples retrieves examples matching the filter
func (s *ExampleStore) GetExamples(operation string, filter ExampleFilter) []Example {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all, ok := s.examples[operation]
	if !ok {
		return nil
	}

	var result []Example
	for _, ex := range all {
		if filter.SchemaType != "" && ex.SchemaType != filter.SchemaType {
			continue
		}
		if filter.Language != "" && ex.Language != filter.Language {
			continue
		}
		result = append(result, ex)
		if filter.MaxResults > 0 && len(result) >= filter.MaxResults {
			break
		}
	}
	return result
}

// FormatExamplesForPrompt converts examples into a text block for inclusion in prompts
func FormatExamplesForPrompt(examples []Example) string {
	if len(examples) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("FEW-SHOT EXAMPLES:\n")
	for i, ex := range examples {
		b.WriteString(fmt.Sprintf("\n--- Example %d (%s, %s) ---\n", i+1, ex.SchemaType, ex.Language))
		b.WriteString(fmt.Sprintf("Headers: %v\n", ex.Headers))
		if len(ex.Mappings) > 0 {
			b.WriteString("Expected mappings:\n")
			for _, m := range ex.Mappings {
				b.WriteString(fmt.Sprintf("  %s → %s (index=%d, confidence=%.1f, reason=%q)\n",
					m.SourceHeader, m.CanonicalName, m.ColumnIndex, m.Confidence, m.Reasoning))
			}
		}
	}
	return b.String()
}

// DefaultExampleStore creates a store pre-loaded with all current examples from prompts.go
func DefaultExampleStore() *ExampleStore {
	store := NewExampleStore()

	// Test Case Style - English
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "test_case", Language: "en",
		Headers: []string{"TC ID", "Test Case Name", "Precondition", "Steps", "Expected", "Status"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "TC ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Test case identifier"},
			{CanonicalName: "feature", SourceHeader: "Test Case Name", ColumnIndex: 1, Confidence: 0.95, Reasoning: "Title of test case"},
			{CanonicalName: "precondition", SourceHeader: "Precondition", ColumnIndex: 2, Confidence: 1.0, Reasoning: "Exact match"},
			{CanonicalName: "instructions", SourceHeader: "Steps", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Exact match"},
			{CanonicalName: "expected", SourceHeader: "Expected", ColumnIndex: 4, Confidence: 0.9, Reasoning: "Expected result"},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 5, Confidence: 1.0, Reasoning: "Exact match"},
		},
	})

	// Issue Tracker Style - Japanese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "issue_tracker", Language: "ja",
		Headers: []string{"Issue #", "概要", "優先度", "担当者", "備考"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Issue #", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Issue identifier"},
			{CanonicalName: "feature", SourceHeader: "概要", ColumnIndex: 1, Confidence: 0.95, Reasoning: "JP: summary/overview"},
			{CanonicalName: "priority", SourceHeader: "優先度", ColumnIndex: 2, Confidence: 1.0, Reasoning: "JP: priority"},
			{CanonicalName: "assignee", SourceHeader: "担当者", ColumnIndex: 3, Confidence: 1.0, Reasoning: "JP: assignee"},
			{CanonicalName: "notes", SourceHeader: "備考", ColumnIndex: 4, Confidence: 0.9, Reasoning: "JP: remarks"},
		},
	})

	// UI Spec Table Style
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "ui_spec", Language: "en",
		Headers: []string{"No", "Item Name", "Item Type", "Required/Optional", "Input Restrictions", "Display Conditions", "Action", "Navigation Destination"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "no", SourceHeader: "No", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Row number"},
			{CanonicalName: "item_name", SourceHeader: "Item Name", ColumnIndex: 1, Confidence: 1.0, Reasoning: "UI item name"},
			{CanonicalName: "item_type", SourceHeader: "Item Type", ColumnIndex: 2, Confidence: 1.0, Reasoning: "UI item type"},
			{CanonicalName: "required_optional", SourceHeader: "Required/Optional", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Required/optional flag"},
			{CanonicalName: "input_restrictions", SourceHeader: "Input Restrictions", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Input constraints"},
			{CanonicalName: "display_conditions", SourceHeader: "Display Conditions", ColumnIndex: 5, Confidence: 1.0, Reasoning: "Display rules"},
			{CanonicalName: "action", SourceHeader: "Action", ColumnIndex: 6, Confidence: 1.0, Reasoning: "User interaction"},
			{CanonicalName: "navigation_destination", SourceHeader: "Navigation Destination", ColumnIndex: 7, Confidence: 1.0, Reasoning: "Navigation target"},
		},
	})

	// Product Backlog Style
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "product_backlog", Language: "en",
		Headers: []string{"Story ID", "Title", "Description", "Acceptance Criteria", "Priority", "Story Points"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Story ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "User story identifier"},
			{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 1, Confidence: 1.0, Reasoning: "Story title"},
			{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 1.0, Reasoning: "Detailed description"},
			{CanonicalName: "acceptance_criteria", SourceHeader: "Acceptance Criteria", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Done definition"},
			{CanonicalName: "priority", SourceHeader: "Priority", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Story priority"},
			{CanonicalName: "notes", SourceHeader: "Story Points", ColumnIndex: 5, Confidence: 0.7, Reasoning: "Effort estimation"},
		},
	})

	// API Specification Style
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "api_spec", Language: "en",
		Headers: []string{"Endpoint", "Method", "Description", "Parameters", "Response", "Status Code"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "endpoint", SourceHeader: "Endpoint", ColumnIndex: 0, Confidence: 1.0, Reasoning: "API endpoint URL path"},
			{CanonicalName: "method", SourceHeader: "Method", ColumnIndex: 1, Confidence: 1.0, Reasoning: "HTTP method"},
			{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 0.95, Reasoning: "Endpoint description"},
			{CanonicalName: "parameters", SourceHeader: "Parameters", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Request parameters"},
			{CanonicalName: "response", SourceHeader: "Response", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Response structure"},
			{CanonicalName: "status_code", SourceHeader: "Status Code", ColumnIndex: 5, Confidence: 1.0, Reasoning: "HTTP status code"},
		},
	})

	// Mixed/Generic Style
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "generic", Language: "en",
		Headers: []string{"ID", "Name", "Type", "Status", "Owner", "Notes"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95, Reasoning: "Generic identifier"},
			{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.8, Reasoning: "Could be title or feature name"},
			{CanonicalName: "type", SourceHeader: "Type", ColumnIndex: 2, Confidence: 0.85, Reasoning: "Classification"},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Exact match"},
			{CanonicalName: "assignee", SourceHeader: "Owner", ColumnIndex: 4, Confidence: 0.95, Reasoning: "Owner/responsible person"},
			{CanonicalName: "notes", SourceHeader: "Notes", ColumnIndex: 5, Confidence: 1.0, Reasoning: "Exact match"},
		},
	})

	// Vietnamese Examples
	// Test Case - Vietnamese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "test_case", Language: "vi",
		Headers: []string{"Mã TC", "Tên trường hợp kiểm thử", "Điều kiện tiên quyết", "Các bước", "Kết quả mong đợi", "Trạng thái"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Mã TC", ColumnIndex: 0, Confidence: 1.0, Reasoning: "VN: test case ID"},
			{CanonicalName: "feature", SourceHeader: "Tên trường hợp kiểm thử", ColumnIndex: 1, Confidence: 0.95, Reasoning: "VN: test case name"},
			{CanonicalName: "precondition", SourceHeader: "Điều kiện tiên quyết", ColumnIndex: 2, Confidence: 1.0, Reasoning: "VN: precondition"},
			{CanonicalName: "instructions", SourceHeader: "Các bước", ColumnIndex: 3, Confidence: 1.0, Reasoning: "VN: steps"},
			{CanonicalName: "expected", SourceHeader: "Kết quả mong đợi", ColumnIndex: 4, Confidence: 0.95, Reasoning: "VN: expected result"},
			{CanonicalName: "status", SourceHeader: "Trạng thái", ColumnIndex: 5, Confidence: 1.0, Reasoning: "VN: status"},
		},
	})

	// API Spec - Vietnamese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "api_spec", Language: "vi",
		Headers: []string{"Điểm cuối", "Phương pháp", "Mô tả", "Thông số", "Phản hồi", "Mã trạng thái"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "endpoint", SourceHeader: "Điểm cuối", ColumnIndex: 0, Confidence: 1.0, Reasoning: "VN: API endpoint"},
			{CanonicalName: "method", SourceHeader: "Phương pháp", ColumnIndex: 1, Confidence: 1.0, Reasoning: "VN: HTTP method"},
			{CanonicalName: "description", SourceHeader: "Mô tả", ColumnIndex: 2, Confidence: 0.95, Reasoning: "VN: description"},
			{CanonicalName: "parameters", SourceHeader: "Thông số", ColumnIndex: 3, Confidence: 1.0, Reasoning: "VN: parameters"},
			{CanonicalName: "response", SourceHeader: "Phản hồi", ColumnIndex: 4, Confidence: 1.0, Reasoning: "VN: response"},
			{CanonicalName: "status_code", SourceHeader: "Mã trạng thái", ColumnIndex: 5, Confidence: 1.0, Reasoning: "VN: status code"},
		},
	})

	// UI Spec - Vietnamese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "ui_spec", Language: "vi",
		Headers: []string{"Số", "Tên mục", "Loại mục", "Bắt buộc/Tùy chọn", "Hạn chế nhập", "Điều kiện hiển thị", "Hành động", "Điểm đến điều hướng"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "no", SourceHeader: "Số", ColumnIndex: 0, Confidence: 1.0, Reasoning: "VN: row number"},
			{CanonicalName: "item_name", SourceHeader: "Tên mục", ColumnIndex: 1, Confidence: 1.0, Reasoning: "VN: item name"},
			{CanonicalName: "item_type", SourceHeader: "Loại mục", ColumnIndex: 2, Confidence: 1.0, Reasoning: "VN: item type"},
			{CanonicalName: "required_optional", SourceHeader: "Bắt buộc/Tùy chọn", ColumnIndex: 3, Confidence: 1.0, Reasoning: "VN: required/optional"},
			{CanonicalName: "input_restrictions", SourceHeader: "Hạn chế nhập", ColumnIndex: 4, Confidence: 1.0, Reasoning: "VN: input restrictions"},
			{CanonicalName: "display_conditions", SourceHeader: "Điều kiện hiển thị", ColumnIndex: 5, Confidence: 1.0, Reasoning: "VN: display conditions"},
			{CanonicalName: "action", SourceHeader: "Hành động", ColumnIndex: 6, Confidence: 1.0, Reasoning: "VN: action"},
			{CanonicalName: "navigation_destination", SourceHeader: "Điểm đến điều hướng", ColumnIndex: 7, Confidence: 1.0, Reasoning: "VN: navigation"},
		},
	})

	// Product Backlog - Vietnamese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "product_backlog", Language: "vi",
		Headers: []string{"Mã câu chuyện", "Tiêu đề", "Mô tả", "Tiêu chí chấp nhận", "Ưu tiên", "Điểm câu chuyện"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Mã câu chuyện", ColumnIndex: 0, Confidence: 1.0, Reasoning: "VN: story ID"},
			{CanonicalName: "title", SourceHeader: "Tiêu đề", ColumnIndex: 1, Confidence: 1.0, Reasoning: "VN: title"},
			{CanonicalName: "description", SourceHeader: "Mô tả", ColumnIndex: 2, Confidence: 1.0, Reasoning: "VN: description"},
			{CanonicalName: "acceptance_criteria", SourceHeader: "Tiêu chí chấp nhận", ColumnIndex: 3, Confidence: 1.0, Reasoning: "VN: acceptance criteria"},
			{CanonicalName: "priority", SourceHeader: "Ưu tiên", ColumnIndex: 4, Confidence: 1.0, Reasoning: "VN: priority"},
			{CanonicalName: "notes", SourceHeader: "Điểm câu chuyện", ColumnIndex: 5, Confidence: 0.7, Reasoning: "VN: story points"},
		},
	})

	// Japanese Examples
	// Test Case - Japanese (additional variant)
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "test_case", Language: "ja",
		Headers: []string{"テストID", "テスト名", "前提条件", "テスト手順", "期待結果", "実行結果"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "テストID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "JP: test ID"},
			{CanonicalName: "feature", SourceHeader: "テスト名", ColumnIndex: 1, Confidence: 0.95, Reasoning: "JP: test name"},
			{CanonicalName: "precondition", SourceHeader: "前提条件", ColumnIndex: 2, Confidence: 1.0, Reasoning: "JP: precondition"},
			{CanonicalName: "instructions", SourceHeader: "テスト手順", ColumnIndex: 3, Confidence: 1.0, Reasoning: "JP: test steps"},
			{CanonicalName: "expected", SourceHeader: "期待結果", ColumnIndex: 4, Confidence: 1.0, Reasoning: "JP: expected result"},
			{CanonicalName: "status", SourceHeader: "実行結果", ColumnIndex: 5, Confidence: 0.9, Reasoning: "JP: execution result"},
		},
	})

	// API Spec - Japanese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "api_spec", Language: "ja",
		Headers: []string{"エンドポイント", "メソッド", "説明", "パラメータ", "レスポンス", "ステータスコード"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "endpoint", SourceHeader: "エンドポイント", ColumnIndex: 0, Confidence: 1.0, Reasoning: "JP: endpoint"},
			{CanonicalName: "method", SourceHeader: "メソッド", ColumnIndex: 1, Confidence: 1.0, Reasoning: "JP: HTTP method"},
			{CanonicalName: "description", SourceHeader: "説明", ColumnIndex: 2, Confidence: 0.95, Reasoning: "JP: description"},
			{CanonicalName: "parameters", SourceHeader: "パラメータ", ColumnIndex: 3, Confidence: 1.0, Reasoning: "JP: parameters"},
			{CanonicalName: "response", SourceHeader: "レスポンス", ColumnIndex: 4, Confidence: 1.0, Reasoning: "JP: response"},
			{CanonicalName: "status_code", SourceHeader: "ステータスコード", ColumnIndex: 5, Confidence: 1.0, Reasoning: "JP: status code"},
		},
	})

	// UI Spec - Japanese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "ui_spec", Language: "ja",
		Headers: []string{"番号", "項目名", "項目型", "必須/任意", "入力制限", "表示条件", "アクション", "遷移先"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "no", SourceHeader: "番号", ColumnIndex: 0, Confidence: 1.0, Reasoning: "JP: row number"},
			{CanonicalName: "item_name", SourceHeader: "項目名", ColumnIndex: 1, Confidence: 1.0, Reasoning: "JP: item name"},
			{CanonicalName: "item_type", SourceHeader: "項目型", ColumnIndex: 2, Confidence: 1.0, Reasoning: "JP: item type"},
			{CanonicalName: "required_optional", SourceHeader: "必須/任意", ColumnIndex: 3, Confidence: 1.0, Reasoning: "JP: required/optional"},
			{CanonicalName: "input_restrictions", SourceHeader: "入力制限", ColumnIndex: 4, Confidence: 1.0, Reasoning: "JP: input restrictions"},
			{CanonicalName: "display_conditions", SourceHeader: "表示条件", ColumnIndex: 5, Confidence: 1.0, Reasoning: "JP: display conditions"},
			{CanonicalName: "action", SourceHeader: "アクション", ColumnIndex: 6, Confidence: 1.0, Reasoning: "JP: action"},
			{CanonicalName: "navigation_destination", SourceHeader: "遷移先", ColumnIndex: 7, Confidence: 1.0, Reasoning: "JP: navigation"},
		},
	})

	// Product Backlog - Japanese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "product_backlog", Language: "ja",
		Headers: []string{"ストーリーID", "タイトル", "説明", "受け入れ基準", "優先度", "ストーリーポイント"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "ストーリーID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "JP: story ID"},
			{CanonicalName: "title", SourceHeader: "タイトル", ColumnIndex: 1, Confidence: 1.0, Reasoning: "JP: title"},
			{CanonicalName: "description", SourceHeader: "説明", ColumnIndex: 2, Confidence: 1.0, Reasoning: "JP: description"},
			{CanonicalName: "acceptance_criteria", SourceHeader: "受け入れ基準", ColumnIndex: 3, Confidence: 1.0, Reasoning: "JP: acceptance criteria"},
			{CanonicalName: "priority", SourceHeader: "優先度", ColumnIndex: 4, Confidence: 1.0, Reasoning: "JP: priority"},
			{CanonicalName: "notes", SourceHeader: "ストーリーポイント", ColumnIndex: 5, Confidence: 0.7, Reasoning: "JP: story points"},
		},
	})

	// Issue Tracker - English (variant)
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "issue_tracker", Language: "en",
		Headers: []string{"Issue ID", "Summary", "Priority", "Status", "Assignee"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Issue ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Issue identifier"},
			{CanonicalName: "feature", SourceHeader: "Summary", ColumnIndex: 1, Confidence: 0.95, Reasoning: "Issue summary"},
			{CanonicalName: "priority", SourceHeader: "Priority", ColumnIndex: 2, Confidence: 1.0, Reasoning: "Priority level"},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Current status"},
			{CanonicalName: "assignee", SourceHeader: "Assignee", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Assigned to"},
		},
	})

	// Issue Tracker - Vietnamese
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "issue_tracker", Language: "vi",
		Headers: []string{"Số vấn đề", "Tóm tắt", "Ưu tiên", "Người được giao", "Ghi chú"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Số vấn đề", ColumnIndex: 0, Confidence: 1.0, Reasoning: "VN: issue number"},
			{CanonicalName: "feature", SourceHeader: "Tóm tắt", ColumnIndex: 1, Confidence: 0.95, Reasoning: "VN: summary"},
			{CanonicalName: "priority", SourceHeader: "Ưu tiên", ColumnIndex: 2, Confidence: 1.0, Reasoning: "VN: priority"},
			{CanonicalName: "assignee", SourceHeader: "Người được giao", ColumnIndex: 3, Confidence: 1.0, Reasoning: "VN: assignee"},
			{CanonicalName: "notes", SourceHeader: "Ghi chú", ColumnIndex: 4, Confidence: 0.9, Reasoning: "VN: notes"},
		},
	})

	// Database Schema - English
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "db_schema", Language: "en",
		Headers: []string{"Column Name", "Data Type", "Nullable", "Primary Key", "Foreign Key", "Index", "Default Value"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "title", SourceHeader: "Column Name", ColumnIndex: 0, Confidence: 0.9, Reasoning: "Column identifier"},
			{CanonicalName: "type", SourceHeader: "Data Type", ColumnIndex: 1, Confidence: 1.0, Reasoning: "Column data type"},
			{CanonicalName: "notes", SourceHeader: "Nullable", ColumnIndex: 2, Confidence: 0.8, Reasoning: "Nullable flag"},
			{CanonicalName: "description", SourceHeader: "Primary Key", ColumnIndex: 3, Confidence: 0.85, Reasoning: "Primary key indicator"},
			{CanonicalName: "status", SourceHeader: "Foreign Key", ColumnIndex: 4, Confidence: 0.85, Reasoning: "Foreign key reference"},
			{CanonicalName: "priority", SourceHeader: "Index", ColumnIndex: 5, Confidence: 0.8, Reasoning: "Index flag"},
			{CanonicalName: "assignee", SourceHeader: "Default Value", ColumnIndex: 6, Confidence: 0.8, Reasoning: "Default value"},
		},
	})

	// Software Requirements Spec - English
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "requirements", Language: "en",
		Headers: []string{"REQ-ID", "Requirement", "Priority", "Module", "Status", "Owner"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "REQ-ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Requirement ID"},
			{CanonicalName: "description", SourceHeader: "Requirement", ColumnIndex: 1, Confidence: 1.0, Reasoning: "Requirement description"},
			{CanonicalName: "priority", SourceHeader: "Priority", ColumnIndex: 2, Confidence: 1.0, Reasoning: "Priority level"},
			{CanonicalName: "component", SourceHeader: "Module", ColumnIndex: 3, Confidence: 0.95, Reasoning: "Module/component"},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 4, Confidence: 1.0, Reasoning: "Status"},
			{CanonicalName: "assignee", SourceHeader: "Owner", ColumnIndex: 5, Confidence: 0.95, Reasoning: "Owner/responsible"},
		},
	})

	// Defect/Bug Report - English
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "defect_report", Language: "en",
		Headers: []string{"Bug ID", "Title", "Severity", "Component", "Steps to Reproduce", "Expected Behavior", "Actual Behavior", "Status"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Bug ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Bug identifier"},
			{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 1, Confidence: 1.0, Reasoning: "Bug title"},
			{CanonicalName: "priority", SourceHeader: "Severity", ColumnIndex: 2, Confidence: 0.95, Reasoning: "Severity/priority"},
			{CanonicalName: "component", SourceHeader: "Component", ColumnIndex: 3, Confidence: 1.0, Reasoning: "Affected component"},
			{CanonicalName: "instructions", SourceHeader: "Steps to Reproduce", ColumnIndex: 4, Confidence: 0.95, Reasoning: "Reproduction steps"},
			{CanonicalName: "expected", SourceHeader: "Expected Behavior", ColumnIndex: 5, Confidence: 1.0, Reasoning: "Expected outcome"},
			{CanonicalName: "description", SourceHeader: "Actual Behavior", ColumnIndex: 6, Confidence: 1.0, Reasoning: "Actual behavior"},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 7, Confidence: 1.0, Reasoning: "Bug status"},
		},
	})

	// Security Requirements - English
	store.Register(Example{
		Operation: "column_mapping", SchemaType: "security_req", Language: "en",
		Headers: []string{"Control ID", "Requirement", "Category", "Risk Level", "Implementation Status", "Owner"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "Control ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Control identifier"},
			{CanonicalName: "description", SourceHeader: "Requirement", ColumnIndex: 1, Confidence: 1.0, Reasoning: "Security requirement"},
			{CanonicalName: "type", SourceHeader: "Category", ColumnIndex: 2, Confidence: 0.95, Reasoning: "Control category"},
			{CanonicalName: "priority", SourceHeader: "Risk Level", ColumnIndex: 3, Confidence: 0.9, Reasoning: "Risk severity"},
			{CanonicalName: "status", SourceHeader: "Implementation Status", ColumnIndex: 4, Confidence: 0.95, Reasoning: "Implementation status"},
			{CanonicalName: "assignee", SourceHeader: "Owner", ColumnIndex: 5, Confidence: 0.9, Reasoning: "Responsible team/owner"},
		},
	})

	return store
}

// SelectionContext provides context for dynamic example selection
type SelectionContext struct {
	SchemaHint  string // Expected schema type
	Language    string // Content language
	ColumnCount int    // Number of columns in input
	MaxResults  int    // Max examples to return (default: 3)
}

const DefaultMaxExamples = 3

// SelectExamples dynamically selects the most relevant examples using scoring.
func (s *ExampleStore) SelectExamples(operation string, ctx SelectionContext) []Example {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all, ok := s.examples[operation]
	if !ok || len(all) == 0 {
		return nil
	}

	maxResults := ctx.MaxResults
	if maxResults <= 0 {
		maxResults = DefaultMaxExamples
	}

	// Score all examples
	type scored struct {
		example Example
		score   int
	}
	scoredExamples := make([]scored, len(all))
	for i, ex := range all {
		scoredExamples[i] = scored{
			example: ex,
			score:   calculateExampleScore(ex, ctx),
		}
	}

	// Sort by score descending (insertion sort — stable, fine for small N)
	for i := 1; i < len(scoredExamples); i++ {
		key := scoredExamples[i]
		j := i - 1
		for j >= 0 && scoredExamples[j].score < key.score {
			scoredExamples[j+1] = scoredExamples[j]
			j--
		}
		scoredExamples[j+1] = key
	}

	// Take top N
	if len(scoredExamples) > maxResults {
		scoredExamples = scoredExamples[:maxResults]
	}

	result := make([]Example, len(scoredExamples))
	for i, sc := range scoredExamples {
		result[i] = sc.example
	}
	return result
}

// calculateExampleScore computes a relevance score for an example given context.
//
// Scoring breakdown:
//   - Exact schema type match: +100
//   - Language match:          +50
//   - Column count similarity: 0–30 (30 − 5×|diff|, floored at 0)
//   - Generic schema type:     +10 base (ensures it surfaces as fallback)
func calculateExampleScore(ex Example, ctx SelectionContext) int {
	score := 0

	// Schema type match: +100 for exact match
	if ctx.SchemaHint != "" && ex.SchemaType == ctx.SchemaHint {
		score += 100
	}

	// Language match: +50
	if ctx.Language != "" && ex.Language == ctx.Language {
		score += 50
	}

	// Column count similarity: 0-30 points
	if ctx.ColumnCount > 0 && len(ex.Headers) > 0 {
		diff := ctx.ColumnCount - len(ex.Headers)
		if diff < 0 {
			diff = -diff
		}
		columnScore := 30 - (diff * 5)
		if columnScore < 0 {
			columnScore = 0
		}
		score += columnScore
	}

	// Generic examples get a small base score so they always surface as fallback
	if ex.SchemaType == "generic" {
		score += 10
	}

	return score
}
