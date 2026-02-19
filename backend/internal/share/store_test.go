package share

import (
	"os"
	"testing"
	"time"
)

func TestUpdateCommentResolution(t *testing.T) {
	// Setup
	store := NewStore("")

	share, err := store.CreateShare(CreateShareInput{
		Title:         "Test Share",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "test-share",
		IsPublic:      true,
		AllowComments: true,
		Permission:    PermissionComment,
	})
	if err != nil {
		t.Fatalf("Failed to create share: %v", err)
	}

	// Add comment
	comment, err := store.AddComment(share.Token, CommentInput{
		Author:  "TestAuthor",
		Message: "Test comment",
	})
	if err != nil {
		t.Fatalf("Failed to add comment: %v", err)
	}

	// Test: resolve comment
	resolved, err := store.UpdateComment(share.Token, comment.ID, true)
	if err != nil {
		t.Fatalf("Failed to resolve comment: %v", err)
	}

	if !resolved.Resolved {
		t.Errorf("Expected comment.Resolved=true, got false")
	}

	// Verify resolution event was emitted
	updated, err := store.GetShare(share.Token)
	if err != nil {
		t.Fatalf("Failed to fetch updated share: %v", err)
	}

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
		if event.Data != "TestAuthor" {
			t.Errorf("Expected data='TestAuthor', got '%s'", event.Data)
		}
	}
}

func TestResolutionEventPersistence(t *testing.T) {
	// Create temp file
	tmpfile, err := os.CreateTemp("", "test_share_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create store and add share with resolved comment
	store := NewStore(tmpfile.Name())

	share, err := store.CreateShare(CreateShareInput{
		Title:         "Persist Test",
		Template:      "test.md",
		MDFlow:        "# Persist",
		Slug:          "persist-test",
		IsPublic:      true,
		AllowComments: true,
		Permission:    PermissionComment,
	})
	if err != nil {
		t.Fatalf("Failed to create share: %v", err)
	}

	comment, err := store.AddComment(share.Token, CommentInput{
		Author:  "Author1",
		Message: "Message1",
	})
	if err != nil {
		t.Fatalf("Failed to add comment: %v", err)
	}

	_, err = store.UpdateComment(share.Token, comment.ID, true)
	if err != nil {
		t.Fatalf("Failed to resolve comment: %v", err)
	}

	// Load from disk
	store2 := NewStore(tmpfile.Name())
	loaded, err := store2.GetShare(share.Token)
	if err != nil {
		t.Fatalf("Failed to load share: %v", err)
	}

	if len(loaded.ResolutionEvents) != 1 {
		t.Errorf("Expected 1 resolution event after reload, got %d", len(loaded.ResolutionEvents))
	}

	if len(loaded.ResolutionEvents) > 0 {
		event := loaded.ResolutionEvents[0]
		if event.EventType != "comment_resolved" {
			t.Errorf("Expected event_type='comment_resolved', got '%s'", event.EventType)
		}
	}
}

func TestMultipleResolutionEvents(t *testing.T) {
	store := NewStore("")

	share, err := store.CreateShare(CreateShareInput{
		Title:         "Multi-Comment Test",
		Template:      "test.md",
		MDFlow:        "# Multi",
		Slug:          "multi-test",
		IsPublic:      true,
		AllowComments: true,
		Permission:    PermissionComment,
	})
	if err != nil {
		t.Fatalf("Failed to create share: %v", err)
	}

	// Add and resolve 3 comments
	for i := 1; i <= 3; i++ {
		comment, err := store.AddComment(share.Token, CommentInput{
			Author:  "Author" + string(rune(48+i)),
			Message: "Message" + string(rune(48+i)),
		})
		if err != nil {
			t.Fatalf("Failed to add comment %d: %v", i, err)
		}

		_, err = store.UpdateComment(share.Token, comment.ID, true)
		if err != nil {
			t.Fatalf("Failed to resolve comment %d: %v", i, err)
		}

		time.Sleep(10 * time.Millisecond) // Slight delay to ensure different timestamps
	}

	// Verify all events
	final, err := store.GetShare(share.Token)
	if err != nil {
		t.Fatalf("Failed to fetch final share: %v", err)
	}

	if len(final.ResolutionEvents) != 3 {
		t.Errorf("Expected 3 resolution events, got %d", len(final.ResolutionEvents))
	}

	// Check ordering (should be chronological)
	for i := 1; i <= 3; i++ {
		if i > len(final.ResolutionEvents) {
			break
		}
		event := final.ResolutionEvents[i-1]
		if event.EventType != "comment_resolved" {
			t.Errorf("Event %d: Expected event_type='comment_resolved', got '%s'", i, event.EventType)
		}
	}
}

func TestCommentResolutionBackwardCompat(t *testing.T) {
	// Simulate old share without ResolutionEvents (backward compatibility)
	store := NewStore("")

	share, err := store.CreateShare(CreateShareInput{
		Title:         "Backward Compat Test",
		Template:      "test.md",
		MDFlow:        "# Test",
		Slug:          "backcompat-test",
		IsPublic:      true,
		AllowComments: true,
		Permission:    PermissionComment,
	})
	if err != nil {
		t.Fatalf("Failed to create share: %v", err)
	}

	// Verify ResolutionEvents initialized (even if empty)
	if share.ResolutionEvents == nil {
		t.Errorf("Expected ResolutionEvents to be initialized (not nil)")
	}

	if len(share.ResolutionEvents) != 0 {
		t.Errorf("Expected empty ResolutionEvents, got %d", len(share.ResolutionEvents))
	}
}
