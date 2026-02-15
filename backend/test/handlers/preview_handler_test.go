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

func TestPreviewPaste(t *testing.T) {
	tests := []struct {
		name           string
		pasteText      string
		expectedStatus int
	}{
		{
			name:           "valid table paste",
			pasteText:      "Header1\tHeader2\nValue1\tValue2\nValue3\tValue4",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty paste",
			pasteText:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "markdown content",
			pasteText:      "# Title\n\nSome markdown content here",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoadConfig()
			h := handlers.NewPreviewHandler(converter.NewConverter(), cfg, handlers.NewAIServiceProvider(cfg))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			reqBody := handlers.PasteConvertRequest{
				PasteText: tt.pasteText,
				Template:  "spec",
			}
			bodyJSON, _ := json.Marshal(reqBody)
			c.Request, _ = http.NewRequest("POST", "/api/mdflow/preview", bytes.NewReader(bodyJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			h.PreviewPaste(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp handlers.PreviewResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				// Check that response has expected structure
				if resp.ColumnMapping == nil {
					t.Error("expected ColumnMapping to be initialized")
				}
				if resp.UnmappedCols == nil {
					t.Error("expected UnmappedCols to be initialized")
				}
			}
		})
	}
}

func TestPreviewHandler_emptyTablePreview(t *testing.T) {
	cfg := config.LoadConfig()
	h := handlers.NewPreviewHandler(converter.NewConverter(), cfg, handlers.NewAIServiceProvider(cfg))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/test", nil)

	resp := handlers.EmptyTablePreviewForTest(h, c, 0, "table")

	if resp.InputType != "table" {
		t.Errorf("expected input_type 'table', got %s", resp.InputType)
	}
	if len(resp.Headers) != 0 {
		t.Errorf("expected empty headers, got %d", len(resp.Headers))
	}
	if len(resp.Rows) != 0 {
		t.Errorf("expected empty rows, got %d", len(resp.Rows))
	}
}
