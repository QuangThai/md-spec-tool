package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func TestConvertPaste(t *testing.T) {
	tests := []struct {
		name           string
		pasteText      string
		detectOnly     bool
		expectedStatus int
		shouldHaveData bool
	}{
		{
			name:           "valid paste conversion",
			pasteText:      "Header1\tHeader2\nValue1\tValue2",
			detectOnly:     false,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "detect only request",
			pasteText:      "Header1\tHeader2\nValue1\tValue2",
			detectOnly:     true,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "empty paste text",
			pasteText:      "   ",
			detectOnly:     false,
			expectedStatus: http.StatusBadRequest,
			shouldHaveData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoadConfig()
			h := handlers.NewConvertHandler(converter.NewConverter(), cfg, handlers.NewAIServiceProvider(cfg))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			reqBody := handlers.PasteConvertRequest{
				PasteText: tt.pasteText,
				Template:  "spec",
				Format:    "spec",
			}
			bodyJSON, _ := json.Marshal(reqBody)
			c.Request, _ = http.NewRequest("POST", "/api/mdflow/paste", bytes.NewReader(bodyJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.detectOnly {
				c.Request.URL.RawQuery = "detect_only=true"
			}

			h.ConvertPaste(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.shouldHaveData && w.Code == http.StatusOK {
				var resp interface{}
				if tt.detectOnly {
					resp = &handlers.InputAnalysisResponse{}
				} else {
					resp = &handlers.MDFlowConvertResponse{}
				}
				if err := json.Unmarshal(w.Body.Bytes(), resp); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
			}
		})
	}
}

func TestConvertPaste_RespectsConvertOptions(t *testing.T) {
	cfg := config.LoadConfig()
	h := handlers.NewConvertHandler(converter.NewConverter(), cfg, handlers.NewAIServiceProvider(cfg))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	includeMetadata := false
	numberRows := true
	reqBody := handlers.PasteConvertRequest{
		PasteText:       "Feature\tScenario\tExpected\nAuth\tLogin works\tDashboard shown\nBilling\tPay invoice\tPayment success",
		Template:        "table",
		Format:          "table",
		IncludeMetadata: &includeMetadata,
		NumberRows:      &numberRows,
	}
	bodyJSON, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/mdflow/paste", bytes.NewReader(bodyJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	h.ConvertPaste(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp handlers.MDFlowConvertResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	trimmed := strings.TrimSpace(resp.MDFlow)
	if strings.HasPrefix(trimmed, "---") || strings.Contains(resp.MDFlow, "name: \"Specification\"") {
		t.Fatal("expected output without metadata front matter when include_metadata=false")
	}
	if !bytes.Contains([]byte(resp.MDFlow), []byte("| # |")) {
		t.Fatal("expected numbered rows header when number_rows=true")
	}
}

func TestConvertXLSX(t *testing.T) {
	tests := []struct {
		name           string
		hasFile        bool
		expectedStatus int
	}{
		{
			name:           "no file uploaded",
			hasFile:        false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoadConfig()
			h := handlers.NewConvertHandler(converter.NewConverter(), cfg, handlers.NewAIServiceProvider(cfg))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/api/mdflow/xlsx", nil)

			if tt.hasFile {
				// Would need to create a proper multipart form with XLSX file
				// Skipping for now as it requires actual file setup
			}

			h.ConvertXLSX(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetXLSXSheets(t *testing.T) {
	t.Run("no file uploaded", func(t *testing.T) {
		cfg := config.LoadConfig()
		h := handlers.NewConvertHandler(converter.NewConverter(), cfg, handlers.NewAIServiceProvider(cfg))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/api/mdflow/xlsx/sheets", nil)

		h.GetXLSXSheets(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// createMultipartForm is a helper to create multipart form with file (used by other tests if needed)
func createMultipartForm(filename string, fileContent []byte) (*bytes.Buffer, string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(part, bytes.NewReader(fileContent))
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
