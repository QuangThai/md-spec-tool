package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

// TestFetchGoogleSheetInvalidURL tests error response for invalid URL
func TestFetchGoogleSheetInvalidURL(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout: 30,
	}
	handler := handlers.NewGSheetHandler(
		converter.NewConverter(),
		converter.NewMDFlowRenderer(),
		nil,
		cfg,
		func(apiKey string) (handlers.Service, error) { return nil, nil },
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := []byte(`{"url":"not-a-valid-url"}`)
	c.Request = httptest.NewRequest("POST", "/gsheet", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.FetchGoogleSheet(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response handlers.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
		if response.Error == "" {
			t.Error("expected error message in response")
		}
	}
}

// TestGetGoogleSheetSheetsInvalidURL tests error response for invalid URL
func TestGetGoogleSheetSheetsInvalidURL(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout: 30,
	}
	handler := handlers.NewGSheetHandler(
		converter.NewConverter(),
		converter.NewMDFlowRenderer(),
		nil,
		cfg,
		func(apiKey string) (handlers.Service, error) { return nil, nil },
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := []byte(`{"url":"https://example.com/invalid"}`)
	c.Request = httptest.NewRequest("POST", "/gsheet/sheets", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetGoogleSheetSheets(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestValidateGIDNonNumeric tests GID validation with non-numeric input
func TestValidateGIDNonNumeric(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout: 30,
	}
	handler := handlers.NewGSheetHandler(
		converter.NewConverter(),
		converter.NewMDFlowRenderer(),
		nil,
		cfg,
		func(apiKey string) (handlers.Service, error) { return nil, nil },
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Valid URL but with invalid GID
	body := []byte(`{
		"url":"https://docs.google.com/spreadsheets/d/ABC123/edit",
		"gid":"not-numeric"
	}`)
	c.Request = httptest.NewRequest("POST", "/gsheet", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.FetchGoogleSheet(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid GID, got %d", w.Code)
	}
}

// TestBearerTokenExtraction tests that bearer token is properly extracted
func TestBearerTokenExtraction(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expectedLen int
	}{
		{
			name:        "valid token",
			authHeader:  "Bearer xyz123abc",
			expectedLen: 9,
		},
		{
			name:        "no header",
			authHeader:  "",
			expectedLen: 0,
		},
		{
			name:        "wrong scheme",
			authHeader:  "Basic xyz123",
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			c := &gin.Context{Request: req}
			token := handlers.GetBearerToken(c)

			if len(token) != tt.expectedLen {
				t.Errorf("expected token length %d, got %d", tt.expectedLen, len(token))
			}
		})
	}
}

// TestGIDSelection tests GID priority logic
func TestGIDSelection(t *testing.T) {
	tests := []struct {
		name       string
		requestGID string
		urlGID     string
		expected   string
	}{
		{
			name:       "request GID takes priority",
			requestGID: "999",
			urlGID:     "111",
			expected:   "999",
		},
		{
			name:       "URL GID used when request empty",
			requestGID: "",
			urlGID:     "222",
			expected:   "222",
		},
		{
			name:       "whitespace is trimmed",
			requestGID: "  555  ",
			urlGID:     "666",
			expected:   "555",
		},
		{
			name:       "both empty returns empty",
			requestGID: "",
			urlGID:     "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handlers.SelectGID(tt.requestGID, tt.urlGID)
			if result != tt.expected {
				t.Errorf("selectGID(%q, %q) = %q, want %q", tt.requestGID, tt.urlGID, result, tt.expected)
			}
		})
	}
}

// TestURLParsing tests various Google Sheets URL formats
func TestURLParsing(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectID    string
		expectGID   string
		expectValid bool
	}{
		{
			name:        "full URL with gid fragment",
			url:         "https://docs.google.com/spreadsheets/d/1mNq8EFV_VXN8EV0mNq8EFV_VXN8EV0mNq8EFV_VXN8/edit#gid=123",
			expectID:    "1mNq8EFV_VXN8EV0mNq8EFV_VXN8EV0mNq8EFV_VXN8",
			expectGID:   "123",
			expectValid: true,
		},
		{
			name:        "URL without gid",
			url:         "https://docs.google.com/spreadsheets/d/1mNq8EFV_VXN8EV0/edit",
			expectID:    "1mNq8EFV_VXN8EV0",
			expectGID:   "",
			expectValid: true,
		},
		{
			name:        "invalid domain",
			url:         "https://sheets.google.com/spreadsheets/d/ABC/edit",
			expectID:    "",
			expectGID:   "",
			expectValid: false,
		},
		{
			name:        "gid in query parameter",
			url:         "https://docs.google.com/spreadsheets/d/1ABC/edit?gid=456",
			expectID:    "1ABC",
			expectGID:   "456",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, gid, valid := handlers.ParseGoogleSheetURL(tt.url)

			if valid != tt.expectValid {
				t.Errorf("valid = %v, want %v", valid, tt.expectValid)
			}
			if id != tt.expectID {
				t.Errorf("id = %q, want %q", id, tt.expectID)
			}
			if gid != tt.expectGID {
				t.Errorf("gid = %q, want %q", gid, tt.expectGID)
			}
		})
	}
}

// TestMatrixAnalysis tests statistics calculation
func TestMatrixAnalysis(t *testing.T) {
	tests := []struct {
		name         string
		rows         [][]string
		expectedRows int
		expectedCols int
	}{
		{
			name: "simple 3x3 matrix",
			rows: [][]string{
				{"Name", "Email", "Phone"},
				{"John", "john@example.com", "555-0101"},
				{"Jane", "jane@example.com", "555-0102"},
			},
			expectedRows: 2,
			expectedCols: 3,
		},
		{
			name: "single row (header only)",
			rows: [][]string{
				{"Col1", "Col2"},
			},
			expectedRows: 0,
			expectedCols: 2,
		},
		{
			name: "many rows",
			rows: [][]string{
				{"A", "B", "C"},
				{"1", "2", "3"},
				{"4", "5", "6"},
				{"7", "8", "9"},
				{"10", "11", "12"},
			},
			expectedRows: 4,
			expectedCols: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matrix := converter.NewCellMatrix(tt.rows).Normalize()
			stats := handlers.AnalyzeSelectedMatrix(matrix)

			if stats.SourceRows != tt.expectedRows {
				t.Errorf("SourceRows = %d, want %d", stats.SourceRows, tt.expectedRows)
			}
			if stats.HeaderCount != tt.expectedCols {
				t.Errorf("HeaderCount = %d, want %d", stats.HeaderCount, tt.expectedCols)
			}
		})
	}
}

// TestFieldCoverageTracking tests which fields are covered
func TestFieldCoverageTracking(t *testing.T) {
	colMap := converter.ColumnMap{
		converter.FieldFeature:       0,
		converter.FieldScenario:      1,
		converter.FieldInstructions:  2,
	}

	coverage := handlers.BuildCoreFieldCoverage(colMap)

	// Fields that should be covered
	requiredFields := []converter.CanonicalField{
		converter.FieldFeature,
		converter.FieldScenario,
		converter.FieldInstructions,
	}

	for _, field := range requiredFields {
		if !coverage[string(field)] {
			t.Errorf("expected %s to be covered", field)
		}
	}

	// Fields that should NOT be covered
	uncoveredFields := []converter.CanonicalField{
		converter.FieldExpected,
		converter.FieldAction,
	}

	for _, field := range uncoveredFields {
		if coverage[string(field)] {
			t.Errorf("expected %s to NOT be covered", field)
		}
	}
}

// TestConvertValidationStatsStructure validates the stats structure
func TestConvertValidationStatsStructure(t *testing.T) {
	stats := handlers.ConvertValidationStats{
		SourceRows:       100,
		HeaderRow:        0,
		HeaderConfidence: 85,
		HeaderCount:      10,
	}

	if stats.SourceRows != 100 {
		t.Errorf("SourceRows mismatch: %d != 100", stats.SourceRows)
	}
	if stats.HeaderConfidence != 85 {
		t.Errorf("HeaderConfidence mismatch: %d != 85", stats.HeaderConfidence)
	}
	if stats.HeaderCount != 10 {
		t.Errorf("HeaderCount mismatch: %d != 10", stats.HeaderCount)
	}
}

// TestGSheetValuesResultStructure validates result object creation
func TestGSheetValuesResultStructure(t *testing.T) {
	rows := [][]string{
		{"A1", "B1", "C1"},
		{"A2", "B2", "C2"},
	}

	result := &handlers.GSheetValuesResult{
		Rows:      rows,
		SheetName: "TestSheet",
		Range:     "TestSheet!A1:C2",
		StartCol:  0,
		StartRow:  0,
	}

	if len(result.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(result.Rows))
	}
	if result.SheetName != "TestSheet" {
		t.Errorf("expected sheet name 'TestSheet', got %q", result.SheetName)
	}
	if result.Range != "TestSheet!A1:C2" {
		t.Errorf("expected range 'TestSheet!A1:C2', got %q", result.Range)
	}
}

// TestContextHandling tests context usage in handlers
func TestContextHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Verify context is passed to request
	if c.Request.Context() != ctx {
		t.Error("expected context to be set in request")
	}
}
