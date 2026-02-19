package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/share"
)

func setupTestShareHandler() (*ShareHandler, *share.Store) {
	store := share.NewStore("")
	handler := NewShareHandler(store)
	return handler, store
}

func TestUpdateComment_PermissionWithToken(t *testing.T) {
	handler, store := setupTestShareHandler()

	// Create share with token
	created, _ := store.CreateShare(share.CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share",
		IsPublic:      true,
		AllowComments: true,
		Permission:    share.PermissionComment,
	})

	// Add comment
	comment, _ := store.AddComment(created.Token, share.CommentInput{
		Author:  "TestAuthor",
		Message: "Test comment",
	})

	// Test 1: Resolve with correct token - should succeed
	reqBody, _ := json.Marshal(UpdateCommentRequest{
		Resolved: true,
		Token:    created.Token,
	})

	req := httptest.NewRequest("PATCH", "/api/share/"+created.Token+"/comments/"+comment.ID, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "key", Value: created.Token},
		{Key: "commentId", Value: comment.ID},
	}

	handler.UpdateComment(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CommentResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Resolved {
		t.Errorf("Expected resolved=true, got false")
	}
}

func TestUpdateComment_PermissionWithWrongToken(t *testing.T) {
	handler, store := setupTestShareHandler()

	// Create share
	created, _ := store.CreateShare(share.CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share-2",
		IsPublic:      true,
		AllowComments: true,
		Permission:    share.PermissionComment,
	})

	// Add comment
	comment, _ := store.AddComment(created.Token, share.CommentInput{
		Author:  "TestAuthor",
		Message: "Test comment",
	})

	// Test: Resolve with wrong token - should fail with 403
	reqBody, _ := json.Marshal(UpdateCommentRequest{
		Resolved: true,
		Token:    "wrong-token-xyz",
	})

	req := httptest.NewRequest("PATCH", "/api/share/"+created.Token+"/comments/"+comment.ID, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "key", Value: created.Token},
		{Key: "commentId", Value: comment.ID},
	}

	handler.UpdateComment(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestUpdateComment_PermissionNoToken(t *testing.T) {
	handler, store := setupTestShareHandler()

	// Create share
	created, _ := store.CreateShare(share.CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share-3",
		IsPublic:      true,
		AllowComments: true,
		Permission:    share.PermissionComment,
	})

	// Add comment
	comment, _ := store.AddComment(created.Token, share.CommentInput{
		Author:  "TestAuthor",
		Message: "Test comment",
	})

	// Test: Resolve without token - should succeed (backward compatible)
	reqBody, _ := json.Marshal(UpdateCommentRequest{
		Resolved: true,
		Token:    "",
	})

	req := httptest.NewRequest("PATCH", "/api/share/"+created.Token+"/comments/"+comment.ID, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "key", Value: created.Token},
		{Key: "commentId", Value: comment.ID},
	}

	handler.UpdateComment(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestUpdateComment_ResolutionEventEmitted(t *testing.T) {
	handler, store := setupTestShareHandler()

	// Create share
	created, _ := store.CreateShare(share.CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share-4",
		IsPublic:      true,
		AllowComments: true,
		Permission:    share.PermissionComment,
	})

	// Add comment
	comment, _ := store.AddComment(created.Token, share.CommentInput{
		Author:  "TestAuthor",
		Message: "Test comment",
	})

	// Resolve comment
	reqBody, _ := json.Marshal(UpdateCommentRequest{
		Resolved: true,
		Token:    created.Token,
	})

	req := httptest.NewRequest("PATCH", "/api/share/"+created.Token+"/comments/"+comment.ID, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "key", Value: created.Token},
		{Key: "commentId", Value: comment.ID},
	}

	handler.UpdateComment(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify event was emitted
	updated, _ := store.GetShare(created.Token)
	if len(updated.ResolutionEvents) != 1 {
		t.Errorf("Expected 1 resolution event, got %d", len(updated.ResolutionEvents))
	}

	if len(updated.ResolutionEvents) > 0 {
		event := updated.ResolutionEvents[0]
		if event.EventType != "comment_resolved" {
			t.Errorf("Expected event_type='comment_resolved', got '%s'", event.EventType)
		}
		if event.CommentID != comment.ID {
			t.Errorf("Expected comment_id='%s', got '%s'", comment.ID, event.CommentID)
		}
	}
}

func TestUpdateComment_MultipleResolutions(t *testing.T) {
	handler, store := setupTestShareHandler()

	// Create share
	created, _ := store.CreateShare(share.CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share-5",
		IsPublic:      true,
		AllowComments: true,
		Permission:    share.PermissionComment,
	})

	// Add 3 comments
	var comments []share.Comment
	for i := 0; i < 3; i++ {
		c, _ := store.AddComment(created.Token, share.CommentInput{
			Author:  "Author" + string(rune(49+i)), // '1', '2', '3'
			Message: "Message" + string(rune(49+i)),
		})
		comments = append(comments, c)
	}

	// Resolve all 3 comments
	for _, comment := range comments {
		reqBody, _ := json.Marshal(UpdateCommentRequest{
			Resolved: true,
			Token:    created.Token,
		})

		req := httptest.NewRequest("PATCH", "/api/share/"+created.Token+"/comments/"+comment.ID, bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{
			{Key: "key", Value: created.Token},
			{Key: "commentId", Value: comment.ID},
		}

		handler.UpdateComment(c)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	}

	// Verify 3 events
	updated, _ := store.GetShare(created.Token)
	if len(updated.ResolutionEvents) != 3 {
		t.Errorf("Expected 3 resolution events, got %d", len(updated.ResolutionEvents))
	}
}

func TestGetShareEvents_EmptyEvents(t *testing.T) {
	handler, store := setupTestShareHandler()

	// Create share (no events yet)
	created, _ := store.CreateShare(share.CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share-events-1",
		IsPublic:      true,
		AllowComments: true,
		Permission:    share.PermissionComment,
	})

	// Get events (should be empty)
	req := httptest.NewRequest("GET", "/api/share/"+created.Token+"/events", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "key", Value: created.Token},
	}

	handler.GetShareEvents(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string][]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result["items"]) != 0 {
		t.Errorf("Expected 0 events, got %d", len(result["items"]))
	}
}

func TestGetShareEvents_WithResolutions(t *testing.T) {
	handler, store := setupTestShareHandler()

	// Create share
	created, _ := store.CreateShare(share.CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share-events-2",
		IsPublic:      true,
		AllowComments: true,
		Permission:    share.PermissionComment,
	})

	// Add 2 comments and resolve them
	for i := 0; i < 2; i++ {
		comment, _ := store.AddComment(created.Token, share.CommentInput{
			Author:  "Author" + string(rune(49+i)),
			Message: "Message" + string(rune(49+i)),
		})

		store.UpdateComment(created.Token, comment.ID, true)
	}

	// Get events
	req := httptest.NewRequest("GET", "/api/share/"+created.Token+"/events", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "key", Value: created.Token},
	}

	handler.GetShareEvents(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string][]map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result["items"]) != 2 {
		t.Errorf("Expected 2 events, got %d", len(result["items"]))
	}

	// Verify event structure
	for _, event := range result["items"] {
		if event["event_type"] != "comment_resolved" {
			t.Errorf("Expected event_type='comment_resolved', got '%v'", event["event_type"])
		}
		if event["comment_id"] == "" {
			t.Errorf("Expected comment_id to be non-empty")
		}
	}
}

func TestGetShareEvents_NotFound(t *testing.T) {
	handler, _ := setupTestShareHandler()

	// Get events for non-existent share
	req := httptest.NewRequest("GET", "/api/share/nonexistent/events", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "key", Value: "nonexistent"},
	}

	handler.GetShareEvents(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
