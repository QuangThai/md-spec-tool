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

func TestValidate(t *testing.T) {
	tests := []struct {
		name           string
		pasteText      string
		expectedStatus int
	}{
		{
			name:           "valid spec doc",
			pasteText:      "Feature\tScenario\nLogin\tValid user",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty paste text",
			pasteText:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoadConfig()
			h := handlers.NewValidationHandler(cfg, handlers.NewAIServiceProvider(cfg))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			reqBody := handlers.ValidateRequest{
				PasteText: tt.pasteText,
				Template:  "spec",
			}
			bodyJSON, _ := json.Marshal(reqBody)
			c.Request, _ = http.NewRequest("POST", "/api/mdflow/validate", bytes.NewReader(bodyJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Validate(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp handlers.ValidateResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				// Just verify the response structure
				_ = resp
			}
		})
	}
}

func TestValidateWithRules(t *testing.T) {
	cfg := config.LoadConfig()
	h := handlers.NewValidationHandler(cfg, handlers.NewAIServiceProvider(cfg))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	rules := &converter.ValidationRules{}
	reqBody := handlers.ValidateRequest{
		PasteText:       "Feature\tScenario\nLogin\tValid user",
		ValidationRules: rules,
		Template:        "spec",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/mdflow/validate", bytes.NewReader(bodyJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Validate(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp handlers.ValidateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	// Just verify the response has expected structure
	_ = resp
}
