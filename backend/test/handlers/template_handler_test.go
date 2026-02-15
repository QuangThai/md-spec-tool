package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func TestGetTemplates(t *testing.T) {
	cfg := config.LoadConfig()
	h := handlers.NewTemplateHandler(converter.NewMDFlowRenderer(), cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/mdflow/templates", nil)

	h.GetTemplates(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp handlers.TemplateListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if len(resp.Templates) == 0 {
		t.Error("expected at least one template")
	}

	// Check for spec and table templates
	hasSpec := false
	hasTable := false
	for _, tmpl := range resp.Templates {
		if tmpl.Name == "spec" {
			hasSpec = true
		}
		if tmpl.Name == "table" {
			hasTable = true
		}
	}

	if !hasSpec {
		t.Error("expected spec template")
	}
	if !hasTable {
		t.Error("expected table template")
	}
}

func TestPreviewTemplate(t *testing.T) {
	tests := []struct {
		name            string
		templateContent string
		sampleData      string
		expectedStatus  int
	}{
		{
			name:            "valid template preview",
			templateContent: "{{.Feature}}: {{.Scenario}}",
			sampleData:      "",
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "empty template content",
			templateContent: "",
			sampleData:      "",
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoadConfig()
			h := handlers.NewTemplateHandler(converter.NewMDFlowRenderer(), cfg)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			reqBody := handlers.TemplatePreviewRequest{
				TemplateContent: tt.templateContent,
				SampleData:      tt.sampleData,
			}
			bodyJSON, _ := json.Marshal(reqBody)
			c.Request, _ = http.NewRequest("POST", "/api/mdflow/templates/preview", bytes.NewReader(bodyJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			h.PreviewTemplate(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetTemplateInfo(t *testing.T) {
	cfg := config.LoadConfig()
	h := handlers.NewTemplateHandler(converter.NewMDFlowRenderer(), cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/mdflow/templates/info", nil)

	h.GetTemplateInfo(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestGetTemplateContent(t *testing.T) {
	tests := []struct {
		name           string
		templateName   string
		expectedStatus int
	}{
		{
			name:           "get spec template",
			templateName:   "spec",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get table template",
			templateName:   "table",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unknown template",
			templateName:   "unknown",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoadConfig()
			h := handlers.NewTemplateHandler(converter.NewMDFlowRenderer(), cfg)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/api/mdflow/templates/"+tt.templateName, nil)
			c.Params = append(c.Params, gin.Param{Key: "name", Value: tt.templateName})

			h.GetTemplateContent(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
