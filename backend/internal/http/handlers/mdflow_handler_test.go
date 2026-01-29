package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

type pasteRequest struct {
	PasteText string `json:"paste_text"`
	Template  string `json:"template,omitempty"`
}

func performJSONRequest(router http.Handler, method string, path string, body interface{}) *httptest.ResponseRecorder {
	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(body)
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func setupMDFlowRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mdflowHandler := NewMDFlowHandler()
	mdflow := router.Group("/api/mdflow")
	{
		mdflow.POST("/paste", mdflowHandler.ConvertPaste)
	}
	return router
}

func readTestFile(t *testing.T, relativePath string) string {
	absPath, err := filepath.Abs(filepath.Join("..", "..", "..", "..", relativePath))
	if err != nil {
		t.Fatalf("Failed to resolve path: %v", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", absPath, err)
	}
	return string(data)
}

func TestConvertPasteAPI_Example1_Markdown(t *testing.T) {
	router := setupMDFlowRouter()
	input := readTestFile(t, "example-1.md")

	// Detect only
	wDetect := performJSONRequest(router, http.MethodPost, "/api/mdflow/paste?detect_only=true", pasteRequest{PasteText: input})
	if wDetect.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wDetect.Code, wDetect.Body.String())
	}
	var analysis InputAnalysisResponse
	if err := json.Unmarshal(wDetect.Body.Bytes(), &analysis); err != nil {
		t.Fatalf("invalid detection response: %v", err)
	}
	if analysis.Type != "markdown" {
		t.Errorf("expected markdown, got %s (reason: %s)", analysis.Type, analysis.Reason)
	}

	// Convert
	wConvert := performJSONRequest(router, http.MethodPost, "/api/mdflow/paste", pasteRequest{PasteText: input})
	if wConvert.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wConvert.Code, wConvert.Body.String())
	}
	var resp MDFlowConvertResponse
	if err := json.Unmarshal(wConvert.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid convert response: %v", err)
	}
	if resp.MDFlow == "" {
		t.Error("expected non-empty mdflow output")
	}
}

func TestConvertPasteAPI_Example2_Table(t *testing.T) {
	router := setupMDFlowRouter()
	input := readTestFile(t, "example-2.md")

	// Detect only
	wDetect := performJSONRequest(router, http.MethodPost, "/api/mdflow/paste?detect_only=true", pasteRequest{PasteText: input})
	if wDetect.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wDetect.Code, wDetect.Body.String())
	}
	var analysis InputAnalysisResponse
	if err := json.Unmarshal(wDetect.Body.Bytes(), &analysis); err != nil {
		t.Fatalf("invalid detection response: %v", err)
	}
	if analysis.Type != "table" {
		t.Errorf("expected table, got %s (reason: %s)", analysis.Type, analysis.Reason)
	}

	// Convert
	wConvert := performJSONRequest(router, http.MethodPost, "/api/mdflow/paste", pasteRequest{PasteText: input})
	if wConvert.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wConvert.Code, wConvert.Body.String())
	}
	var resp MDFlowConvertResponse
	if err := json.Unmarshal(wConvert.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid convert response: %v", err)
	}
	if resp.MDFlow == "" {
		t.Error("expected non-empty mdflow output")
	}
	if resp.Meta.TotalRows == 0 {
		t.Error("expected rows parsed from table input")
	}
}
