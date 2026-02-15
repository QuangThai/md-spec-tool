package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func TestDiffMDFlowBackwardCompat_NoPanicWithBYOKHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`{"before":"A","after":"B"}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/mdflow/diff", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set(handlers.BYOKHeader, "user-key")

	h := handlers.DiffMDFlow()
	h(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
