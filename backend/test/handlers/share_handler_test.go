package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
	"github.com/yourorg/md-spec-tool/internal/share"
)

type errorResponse struct {
	Error string `json:"error"`
}

type shareResponse struct {
	Token         string `json:"token"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Template      string `json:"template"`
	MDFlow        string `json:"mdflow"`
	IsPublic      bool   `json:"is_public"`
	AllowComments bool   `json:"allow_comments"`
	Permission    string `json:"permission"`
	CreatedAt     string `json:"created_at"`
}

type commentResponse struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Message   string `json:"message"`
	Resolved  bool   `json:"resolved"`
	CreatedAt string `json:"created_at"`
}

type listCommentsResponse struct {
	Items []commentResponse `json:"items"`
}

type listPublicResponse struct {
	Items []share.ShareSummary `json:"items"`
}

func setupShareRouter(t *testing.T) (*gin.Engine, *share.Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	storePath := filepath.Join(t.TempDir(), "share.json")
	store := share.NewStore(storePath)
	handler := handlers.NewShareHandler(store)

	shareRoutes := router.Group("/api/share")
	{
		shareRoutes.POST("", middleware.RateLimit(10, time.Minute), handler.CreateShare)
		shareRoutes.GET("/public", handler.ListPublic)
		shareRoutes.GET("/:key", handler.GetShare)
		shareRoutes.PATCH("/:key", middleware.RateLimit(20, time.Minute), handler.UpdateShare)
		shareRoutes.GET("/:key/comments", handler.ListComments)
		shareRoutes.POST("/:key/comments", middleware.RateLimit(20, time.Minute), handler.CreateComment)
		shareRoutes.PATCH("/:key/comments/:commentId", middleware.RateLimit(20, time.Minute), handler.UpdateComment)
	}

	return router, store
}

func performRequest(t *testing.T, router *gin.Engine, method, path, body, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func decodeJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	decoder := json.NewDecoder(recorder.Body)
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
}

func TestCreateShareSuccess(t *testing.T) {
	router, _ := setupShareRouter(t)

	payload := `{"title":"Spec","template":"tmpl","mdflow":"content","slug":"spec-doc","is_public":true,"allow_comments":true}`
	recorder := performRequest(t, router, http.MethodPost, "/api/share", payload, "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response shareResponse
	decodeJSON(t, recorder, &response)
	if response.Token == "" {
		t.Fatal("expected token to be set")
	}
	if response.Slug != "spec-doc" {
		t.Fatalf("expected slug spec-doc, got %q", response.Slug)
	}
	if !response.AllowComments || response.Permission != string(share.PermissionComment) {
		t.Fatalf("expected comments enabled with permission comment")
	}
}

func TestCreateShareInvalidSlug(t *testing.T) {
	router, _ := setupShareRouter(t)

	payload := `{"title":"Spec","mdflow":"content","slug":"aa","is_public":true}`
	recorder := performRequest(t, router, http.MethodPost, "/api/share", payload, "")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var response errorResponse
	decodeJSON(t, recorder, &response)
	if response.Error != "invalid slug" {
		t.Fatalf("expected error invalid slug, got %q", response.Error)
	}
}

func TestCreateShareSlugExists(t *testing.T) {
	router, _ := setupShareRouter(t)

	payload := `{"title":"Spec","mdflow":"content","slug":"spec-doc","is_public":true}`
	first := performRequest(t, router, http.MethodPost, "/api/share", payload, "")
	if first.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", first.Code)
	}

	recorder := performRequest(t, router, http.MethodPost, "/api/share", payload, "")
	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}

	var response errorResponse
	decodeJSON(t, recorder, &response)
	if response.Error != "slug already exists" {
		t.Fatalf("expected error slug already exists, got %q", response.Error)
	}
}

func TestCreateShareInvalidPermission(t *testing.T) {
	router, _ := setupShareRouter(t)

	payload := `{"title":"Spec","mdflow":"content","permission":"admin"}`
	recorder := performRequest(t, router, http.MethodPost, "/api/share", payload, "")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var response errorResponse
	decodeJSON(t, recorder, &response)
	if response.Error != "invalid permission" {
		t.Fatalf("expected error invalid permission, got %q", response.Error)
	}
}

func TestGetShareSuccess(t *testing.T) {
	router, _ := setupShareRouter(t)

	payload := `{"title":"Spec","mdflow":"content"}`
	create := performRequest(t, router, http.MethodPost, "/api/share", payload, "")
	if create.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", create.Code)
	}

	var created shareResponse
	decodeJSON(t, create, &created)

	recorder := performRequest(t, router, http.MethodGet, "/api/share/"+created.Token, "", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response shareResponse
	decodeJSON(t, recorder, &response)
	if response.Token != created.Token {
		t.Fatalf("expected token %q, got %q", created.Token, response.Token)
	}
}

func TestListPublicReturnsItems(t *testing.T) {
	router, _ := setupShareRouter(t)

	payload := `{"title":"Spec","mdflow":"content","is_public":true}`
	create := performRequest(t, router, http.MethodPost, "/api/share", payload, "")
	if create.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", create.Code)
	}

	recorder := performRequest(t, router, http.MethodGet, "/api/share/public", "", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response listPublicResponse
	decodeJSON(t, recorder, &response)
	if len(response.Items) == 0 {
		t.Fatal("expected public items to be returned")
	}
}

func TestUpdateShareUpdatesFlags(t *testing.T) {
	router, _ := setupShareRouter(t)

	payload := `{"title":"Spec","mdflow":"content"}`
	create := performRequest(t, router, http.MethodPost, "/api/share", payload, "")
	if create.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", create.Code)
	}

	var created shareResponse
	decodeJSON(t, create, &created)

	patchPayload := `{"is_public":true,"allow_comments":true}`
	recorder := performRequest(t, router, http.MethodPatch, "/api/share/"+created.Token, patchPayload, "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response shareResponse
	decodeJSON(t, recorder, &response)
	if !response.IsPublic || !response.AllowComments {
		t.Fatal("expected share to be public with comments enabled")
	}
	if response.Permission != string(share.PermissionComment) {
		t.Fatalf("expected permission comment, got %q", response.Permission)
	}
	if response.Slug == "" {
		t.Fatal("expected slug to be generated for public share")
	}
}

func TestCommentsFlowDisabledAndEnabled(t *testing.T) {
	router, _ := setupShareRouter(t)

	disabledPayload := `{"title":"Spec","mdflow":"content"}`
	disabledCreate := performRequest(t, router, http.MethodPost, "/api/share", disabledPayload, "")
	if disabledCreate.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", disabledCreate.Code)
	}

	var disabledShare shareResponse
	decodeJSON(t, disabledCreate, &disabledShare)

	disabledComment := performRequest(t, router, http.MethodPost, "/api/share/"+disabledShare.Token+"/comments", `{"message":"hi"}`, "")
	if disabledComment.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", disabledComment.Code)
	}

	var disabledError errorResponse
	decodeJSON(t, disabledComment, &disabledError)
	if disabledError.Error != "comments are disabled" {
		t.Fatalf("expected error comments are disabled, got %q", disabledError.Error)
	}

	enabledPayload := `{"title":"Spec","mdflow":"content","allow_comments":true}`
	enabledCreate := performRequest(t, router, http.MethodPost, "/api/share", enabledPayload, "")
	if enabledCreate.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", enabledCreate.Code)
	}

	var enabledShare shareResponse
	decodeJSON(t, enabledCreate, &enabledShare)

	commentCreate := performRequest(t, router, http.MethodPost, "/api/share/"+enabledShare.Token+"/comments", `{"message":"hello"}`, "")
	if commentCreate.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", commentCreate.Code)
	}

	var createdComment commentResponse
	decodeJSON(t, commentCreate, &createdComment)

	listRecorder := performRequest(t, router, http.MethodGet, "/api/share/"+enabledShare.Token+"/comments", "", "")
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listRecorder.Code)
	}

	var listResponse listCommentsResponse
	decodeJSON(t, listRecorder, &listResponse)
	if len(listResponse.Items) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(listResponse.Items))
	}

	patchPayload := `{"resolved":true}`
	patchRecorder := performRequest(t, router, http.MethodPatch, "/api/share/"+enabledShare.Token+"/comments/"+createdComment.ID, patchPayload, "")
	if patchRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", patchRecorder.Code)
	}

	var patchedComment commentResponse
	decodeJSON(t, patchRecorder, &patchedComment)
	if !patchedComment.Resolved {
		t.Fatal("expected comment to be resolved")
	}
}

func TestRateLimitCreateShare(t *testing.T) {
	router, _ := setupShareRouter(t)
	remoteAddr := "10.0.0.1:1234"

	payload := `{"title":"Spec","mdflow":"content"}`
	for i := 0; i < 10; i++ {
		recorder := performRequest(t, router, http.MethodPost, "/api/share", payload, remoteAddr)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", recorder.Code)
		}
	}

	recorder := performRequest(t, router, http.MethodPost, "/api/share", payload, remoteAddr)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", recorder.Code)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}

	var response errorResponse
	decodeJSON(t, recorder, &response)
	if !strings.Contains(response.Error, "rate limit exceeded") {
		t.Fatalf("expected error containing 'rate limit exceeded', got %q", response.Error)
	}
}

func TestRateLimitCreateComment(t *testing.T) {
	router, _ := setupShareRouter(t)
	remoteAddr := "10.0.0.1:1234"

	create := performRequest(t, router, http.MethodPost, "/api/share", `{"title":"Spec","mdflow":"content","allow_comments":true}`, remoteAddr)
	if create.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", create.Code)
	}

	var shareCreated shareResponse
	decodeJSON(t, create, &shareCreated)

	for i := 0; i < 20; i++ {
		recorder := performRequest(t, router, http.MethodPost, "/api/share/"+shareCreated.Token+"/comments", `{"message":"hello"}`, remoteAddr)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", recorder.Code)
		}
	}

	recorder := performRequest(t, router, http.MethodPost, "/api/share/"+shareCreated.Token+"/comments", `{"message":"hello"}`, remoteAddr)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", recorder.Code)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}
}

func TestRateLimitUpdateShare(t *testing.T) {
	router, _ := setupShareRouter(t)
	remoteAddr := "10.0.0.2:1234"

	create := performRequest(t, router, http.MethodPost, "/api/share", `{"title":"Spec","mdflow":"content"}`, remoteAddr)
	if create.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", create.Code)
	}

	var created shareResponse
	decodeJSON(t, create, &created)

	patchPayload := `{"allow_comments":true}`
	for i := 0; i < 20; i++ {
		recorder := performRequest(t, router, http.MethodPatch, "/api/share/"+created.Token, patchPayload, remoteAddr)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", recorder.Code)
		}
	}

	recorder := performRequest(t, router, http.MethodPatch, "/api/share/"+created.Token, patchPayload, remoteAddr)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", recorder.Code)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}

	var response errorResponse
	decodeJSON(t, recorder, &response)
	if !strings.Contains(response.Error, "rate limit exceeded") {
		t.Fatalf("expected error containing 'rate limit exceeded', got %q", response.Error)
	}
}

func TestCreateSharePayloadTooLarge(t *testing.T) {
	router, _ := setupShareRouter(t)

	mdflow := strings.Repeat("a", (1<<20)+1)
	payload := `{"title":"Spec","mdflow":"` + mdflow + `"}`
	recorder := performRequest(t, router, http.MethodPost, "/api/share", payload, "")

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", recorder.Code)
	}

	var response errorResponse
	decodeJSON(t, recorder, &response)
	if response.Error != "payload too large" {
		t.Fatalf("expected error payload too large, got %q", response.Error)
	}
}

func TestCreateCommentTooLong(t *testing.T) {
	router, _ := setupShareRouter(t)

	create := performRequest(t, router, http.MethodPost, "/api/share", `{"title":"Spec","mdflow":"content","allow_comments":true}`, "")
	if create.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", create.Code)
	}

	var shareCreated shareResponse
	decodeJSON(t, create, &shareCreated)

	message := strings.Repeat("b", (5*1024)+1)
	payload := `{"message":"` + message + `"}`
	recorder := performRequest(t, router, http.MethodPost, "/api/share/"+shareCreated.Token+"/comments", payload, "")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var response errorResponse
	decodeJSON(t, recorder, &response)
	if response.Error != "comment too long" {
		t.Fatalf("expected error comment too long, got %q", response.Error)
	}
}
