package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestNewGSheetHandler tests handler initialization
func TestNewGSheetHandler(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout: 30,
	}
	getAIService := func(apiKey string) (handlers.Service, error) {
		return nil, nil
	}

	handler := handlers.NewGSheetHandler(nil, nil, nil, cfg, getAIService)

	if handler == nil {
		t.Fatal("expected handler to be non-nil")
	}
	if handlers.GSheetHandlerConfigForTest(handler) != cfg {
		t.Error("expected config to be set")
	}
	if !handlers.GSheetHandlerHasGetAIServiceForTest(handler) {
		t.Error("expected getAIService to be set")
	}
}

// TestParseGoogleSheetURL tests URL parsing
func TestParseGoogleSheetURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  string
		wantGID string
		wantOK  bool
	}{
		{
			name:    "standard format with gid",
			url:     "https://docs.google.com/spreadsheets/d/SHEET_ID_123/edit#gid=456",
			wantID:  "SHEET_ID_123",
			wantGID: "456",
			wantOK:  true,
		},
		{
			name:    "format without gid",
			url:     "https://docs.google.com/spreadsheets/d/SHEET_ID_789/edit",
			wantID:  "SHEET_ID_789",
			wantGID: "",
			wantOK:  true,
		},
		{
			name:    "invalid url",
			url:     "https://example.com/invalid",
			wantID:  "",
			wantGID: "",
			wantOK:  false,
		},
		{
			name:    "empty url",
			url:     "",
			wantID:  "",
			wantGID: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, gid, ok := handlers.ParseGoogleSheetURL(tt.url)
			if ok != tt.wantOK {
				t.Errorf("parseGoogleSheetURL(%q) ok = %v, want %v", tt.url, ok, tt.wantOK)
			}
			if id != tt.wantID {
				t.Errorf("parseGoogleSheetURL(%q) id = %q, want %q", tt.url, id, tt.wantID)
			}
			if gid != tt.wantGID {
				t.Errorf("parseGoogleSheetURL(%q) gid = %q, want %q", tt.url, gid, tt.wantGID)
			}
		})
	}
}

// TestSelectGID tests GID selection logic
func TestSelectGID(t *testing.T) {
	tests := []struct {
		name       string
		requestGID string
		urlGID     string
		want       string
	}{
		{
			name:       "request gid takes precedence",
			requestGID: "100",
			urlGID:     "200",
			want:       "100",
		},
		{
			name:       "use url gid when request empty",
			requestGID: "",
			urlGID:     "200",
			want:       "200",
		},
		{
			name:       "whitespace trimmed",
			requestGID: "  100  ",
			urlGID:     "200",
			want:       "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.SelectGID(tt.requestGID, tt.urlGID)
			if got != tt.want {
				t.Errorf("selectGID(%q, %q) = %q, want %q", tt.requestGID, tt.urlGID, got, tt.want)
			}
		})
	}
}

// TestValidateGID tests GID validation
func TestValidateGID(t *testing.T) {
	tests := []struct {
		name    string
		gid     string
		wantErr bool
	}{
		{
			name:    "valid numeric gid",
			gid:     "123",
			wantErr: false,
		},
		{
			name:    "empty gid is valid",
			gid:     "",
			wantErr: false,
		},
		{
			name:    "invalid non-numeric",
			gid:     "abc",
			wantErr: true,
		},
		{
			name:    "zero is valid",
			gid:     "0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidateGID(tt.gid)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGID(%q) error = %v, wantErr %v", tt.gid, err, tt.wantErr)
			}
		})
	}
}

// TestGetBearerToken tests bearer token extraction
func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer my-token-123",
			expectedToken: "my-token-123",
		},
		{
			name:          "no header",
			authHeader:    "",
			expectedToken: "",
		},
		{
			name:          "invalid format",
			authHeader:    "Basic xyz",
			expectedToken: "",
		},
		{
			name:          "case insensitive bearer",
			authHeader:    "bearer token-abc",
			expectedToken: "token-abc",
		},
		{
			name:          "whitespace trimmed",
			authHeader:    "Bearer   spaced-token  ",
			expectedToken: "spaced-token",
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

			if token != tt.expectedToken {
				t.Errorf("getBearerToken() = %q, want %q", token, tt.expectedToken)
			}
		})
	}
}

// TestIsAuthError tests auth error detection
func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantAuth bool
	}{
		{
			name:     "nil error",
			err:      nil,
			wantAuth: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			wantAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.IsAuthError(tt.err)
			if got != tt.wantAuth {
				t.Errorf("isAuthError() = %v, want %v", got, tt.wantAuth)
			}
		})
	}
}

// TestFetchGoogleSheetURL tests HTTP endpoint for fetching sheet
func TestFetchGoogleSheetURL(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout:       30,
		SpecMinHeaderConfidence: 70,
		SpecStrictMode:          true,
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

	// Test missing URL
	c.Request = httptest.NewRequest("POST", "/gsheet", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Body = nil

	handler.FetchGoogleSheet(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestPreviewGoogleSheetRequest tests preview request parsing
func TestPreviewGoogleSheetRequest(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout:       30,
		SpecMinHeaderConfidence: 70,
		SpecStrictMode:          true,
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

	// Test invalid URL format
	body := strings.NewReader(`{"url":"invalid-url"}`)
	c.Request = httptest.NewRequest("POST", "/gsheet/preview", body)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.PreviewGoogleSheet(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid URL, got %d", w.Code)
	}
}

// TestConvertMatrixSelection tests block selection logic
func TestConvertMatrixSelection(t *testing.T) {
	matrix := converter.NewCellMatrix([][]string{
		{"Header1", "Header2"},
		{"Value1", "Value2"},
		{"Value3", "Value4"},
	})

	ctx := context.Background()
	conv := converter.NewConverter()

	// Test with no selected block ID
	selected := handlers.SelectMatrixForConvert(ctx, conv, matrix, "spec", "", "")
	if selected == nil {
		t.Error("expected selected matrix to be non-nil")
	}
	if selected.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", selected.RowCount())
	}
}

// TestAnalyzeSelectedMatrix tests matrix analysis
func TestAnalyzeSelectedMatrix(t *testing.T) {
	matrix := converter.NewCellMatrix([][]string{
		{"Name", "Email", "Phone"},
		{"John", "john@example.com", "123-456"},
		{"Jane", "jane@example.com", "789-012"},
	}).Normalize()

	stats := handlers.AnalyzeSelectedMatrix(matrix)

	if stats.SourceRows != 2 {
		t.Errorf("expected 2 source rows, got %d", stats.SourceRows)
	}
	if stats.HeaderCount != 3 {
		t.Errorf("expected 3 header columns, got %d", stats.HeaderCount)
	}
}

// TestBuildCoreFieldCoverage tests coverage calculation
func TestBuildCoreFieldCoverage(t *testing.T) {
	colMap := converter.ColumnMap{
		converter.FieldFeature:     0,
		converter.FieldScenario:    1,
		converter.FieldDescription: 2,
	}

	coverage := handlers.BuildCoreFieldCoverage(colMap)

	if !coverage[string(converter.FieldFeature)] {
		t.Error("expected Feature to be covered")
	}
	if !coverage[string(converter.FieldScenario)] {
		t.Error("expected Scenario to be covered")
	}
	if coverage[string(converter.FieldExpected)] {
		t.Error("expected Expected to not be covered")
	}
}

// TestConvertValidationStats validates stats structure
func TestConvertValidationStats(t *testing.T) {
	stats := handlers.ConvertValidationStats{
		SourceRows:       10,
		HeaderRow:        0,
		HeaderConfidence: 95,
		HeaderCount:      5,
	}

	if stats.SourceRows != 10 {
		t.Error("expected SourceRows to be 10")
	}
	if stats.HeaderConfidence != 95 {
		t.Error("expected HeaderConfidence to be 95")
	}
}

// TestGSheetValuesResult validates result structure
func TestGSheetValuesResult(t *testing.T) {
	result := &handlers.GSheetValuesResult{
		Rows:      [][]string{{"a", "b"}},
		SheetName: "Sheet1",
		Range:     "Sheet1!A1:B2",
		StartCol:  0,
		StartRow:  0,
	}

	if len(result.Rows) != 1 {
		t.Error("expected 1 row")
	}
	if result.SheetName != "Sheet1" {
		t.Error("expected sheet name to be Sheet1")
	}
}

func TestFetchGoogleSheet_ReturnsLegacyResponse(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout: 30,
		MaxUploadBytes:    1 << 20,
	}
	handler := handlers.NewGSheetHandler(
		converter.NewConverter(),
		converter.NewMDFlowRenderer(),
		nil,
		cfg,
		func(apiKey string) (handlers.Service, error) { return nil, nil },
	)
	handlers.SetGSheetHTTPClientForTest(handler, &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("Feature,Scenario\nLogin,Valid")),
				Header:     make(http.Header),
			}, nil
		}),
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"url":"https://docs.google.com/spreadsheets/d/abc123/edit#gid=0"}`)
	c.Request = httptest.NewRequest("POST", "/gsheet", body)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.FetchGoogleSheet(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp handlers.GoogleSheetResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.SheetID != "abc123" {
		t.Fatalf("expected sheet_id abc123, got %q", resp.SheetID)
	}
	if !strings.Contains(resp.Data, "Feature,Scenario") {
		t.Fatalf("expected legacy data payload, got %q", resp.Data)
	}
}

func TestPreviewGoogleSheet_FallbackIncludesBlocks(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout: 30,
		MaxUploadBytes:    1 << 20,
	}
	handler := handlers.NewGSheetHandler(
		converter.NewConverter(),
		converter.NewMDFlowRenderer(),
		nil,
		cfg,
		func(apiKey string) (handlers.Service, error) { return nil, nil },
	)
	handlers.SetGSheetHTTPClientForTest(handler, &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			csv := "Feature,Scenario,Instructions,Expected\nAuth,Login,Open app,Success"
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(csv)),
				Header:     make(http.Header),
			}, nil
		}),
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"url":"https://docs.google.com/spreadsheets/d/abc123/edit#gid=0","template":"spec","format":"spec"}`)
	c.Request = httptest.NewRequest("POST", "/gsheet/preview", body)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.PreviewGoogleSheet(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp handlers.PreviewResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Blocks) == 0 {
		t.Fatal("expected blocks to be present")
	}
	if strings.TrimSpace(resp.SelectedBlockID) == "" {
		t.Fatal("expected selected_block_id to be present")
	}
	if strings.TrimSpace(resp.SelectedBlockRange) == "" {
		t.Fatal("expected selected_block_range to be present")
	}
}

// BenchmarkParseGoogleSheetURL benchmarks URL parsing
func BenchmarkParseGoogleSheetURL(b *testing.B) {
	url := "https://docs.google.com/spreadsheets/d/SHEET_ID_123/edit#gid=456"

	for i := 0; i < b.N; i++ {
		handlers.ParseGoogleSheetURL(url)
	}
}

// BenchmarkSelectGID benchmarks GID selection
func BenchmarkSelectGID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		handlers.SelectGID("100", "200")
	}
}

// BenchmarkValidateGID benchmarks GID validation
func BenchmarkValidateGID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = handlers.ValidateGID("12345")
	}
}

// BenchmarkAnalyzeSelectedMatrix benchmarks matrix analysis
func BenchmarkAnalyzeSelectedMatrix(b *testing.B) {
	rows := make([][]string, 100)
	for i := 0; i < 100; i++ {
		rows[i] = []string{"col1", "col2", "col3", "col4", "col5"}
	}
	matrix := converter.NewCellMatrix(rows).Normalize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handlers.AnalyzeSelectedMatrix(matrix)
	}
}
