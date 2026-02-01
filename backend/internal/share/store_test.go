package share

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateShare_PublicAndPrivate(t *testing.T) {
	store := NewStore("")

	publicShare, err := store.CreateShare(CreateShareInput{
		Title:         "Hello, World!!",
		Template:      "tmpl",
		MDFlow:        "md",
		IsPublic:      true,
		AllowComments: true,
	})
	if err != nil {
		t.Fatalf("CreateShare public: %v", err)
	}
	if publicShare.Token == "" {
		t.Fatal("public share token should not be empty")
	}
	if publicShare.Slug != "hello-world" {
		t.Fatalf("expected normalized slug 'hello-world', got %q", publicShare.Slug)
	}
	if publicShare.Permission != PermissionComment {
		t.Fatalf("expected permission %q, got %q", PermissionComment, publicShare.Permission)
	}
	if publicShare.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
	if publicShare.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected CreatedAt in UTC, got %v", publicShare.CreatedAt.Location())
	}

	privateShare, err := store.CreateShare(CreateShareInput{
		Title:         "Private",
		Template:      "tmpl",
		MDFlow:        "md",
		IsPublic:      false,
		AllowComments: false,
		Slug:          "ignored",
	})
	if err != nil {
		t.Fatalf("CreateShare private: %v", err)
	}
	if privateShare.Slug != "" {
		t.Fatalf("expected private share slug to be empty, got %q", privateShare.Slug)
	}
	if privateShare.Permission != PermissionView {
		t.Fatalf("expected permission %q, got %q", PermissionView, privateShare.Permission)
	}
}

func TestCreateShare_InvalidPermission(t *testing.T) {
	store := NewStore("")

	_, err := store.CreateShare(CreateShareInput{
		Title:      "Bad",
		IsPublic:   true,
		Permission: Permission("admin"),
	})
	if !errors.Is(err, ErrInvalidPermission) {
		t.Fatalf("expected ErrInvalidPermission, got %v", err)
	}
}

func TestCreateShare_SlugBehavior(t *testing.T) {
	store := NewStore("")

	_, err := store.CreateShare(CreateShareInput{
		Title:    "",
		IsPublic: true,
		Slug:     "a",
	})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for short slug, got %v", err)
	}

	share, err := store.CreateShare(CreateShareInput{
		Title:    "Duplicate",
		IsPublic: true,
		Slug:     "dup-slug",
	})
	if err != nil {
		t.Fatalf("CreateShare first slug: %v", err)
	}
	if share.Slug != "dup-slug" {
		t.Fatalf("expected slug dup-slug, got %q", share.Slug)
	}

	_, err = store.CreateShare(CreateShareInput{
		Title:    "Duplicate Two",
		IsPublic: true,
		Slug:     "dup-slug",
	})
	if !errors.Is(err, ErrSlugExists) {
		t.Fatalf("expected ErrSlugExists, got %v", err)
	}
}

func TestGetShare_ByTokenAndSlug(t *testing.T) {
	store := NewStore("")
	share, err := store.CreateShare(CreateShareInput{
		Title:    "Get Me",
		IsPublic: true,
	})
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}

	byToken, err := store.GetShare(share.Token)
	if err != nil {
		t.Fatalf("GetShare token: %v", err)
	}
	if byToken.Token != share.Token {
		t.Fatalf("expected token %q, got %q", share.Token, byToken.Token)
	}

	bySlug, err := store.GetShare(share.Slug)
	if err != nil {
		t.Fatalf("GetShare slug: %v", err)
	}
	if bySlug.Slug != share.Slug {
		t.Fatalf("expected slug %q, got %q", share.Slug, bySlug.Slug)
	}
}

func TestListPublic_SortsAndFilters(t *testing.T) {
	store := NewStore("")

	first, err := store.CreateShare(CreateShareInput{
		Title:    "First",
		IsPublic: true,
	})
	if err != nil {
		t.Fatalf("CreateShare first: %v", err)
	}
	second, err := store.CreateShare(CreateShareInput{
		Title:    "Second",
		IsPublic: true,
	})
	if err != nil {
		t.Fatalf("CreateShare second: %v", err)
	}
	third, err := store.CreateShare(CreateShareInput{
		Title:    "Third",
		IsPublic: true,
	})
	if err != nil {
		t.Fatalf("CreateShare third: %v", err)
	}
	_, err = store.CreateShare(CreateShareInput{
		Title:    "Private",
		IsPublic: false,
	})
	if err != nil {
		t.Fatalf("CreateShare private: %v", err)
	}

	store.mu.Lock()
	first.CreatedAt = time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	second.CreatedAt = time.Date(2024, 2, 10, 10, 0, 0, 0, time.UTC)
	third.CreatedAt = time.Date(2023, 12, 10, 10, 0, 0, 0, time.UTC)
	store.mu.Unlock()

	list := store.ListPublic()
	if len(list) != 3 {
		t.Fatalf("expected 3 public shares, got %d", len(list))
	}
	if list[0].Token != second.Token || list[1].Token != first.Token || list[2].Token != third.Token {
		t.Fatalf("unexpected order: %s, %s, %s", list[0].Token, list[1].Token, list[2].Token)
	}
}

func TestUpdateShare_TogglesAndPermissions(t *testing.T) {
	store := NewStore("")
	share, err := store.CreateShare(CreateShareInput{
		Title:         "Toggle",
		IsPublic:      false,
		AllowComments: false,
	})
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}

	makePublic := true
	updated, err := store.UpdateShare(share.Token, &makePublic, nil)
	if err != nil {
		t.Fatalf("UpdateShare public true: %v", err)
	}
	if !updated.IsPublic || updated.Slug == "" {
		t.Fatal("expected share to be public with slug")
	}
	if token, ok := store.slugIndex[updated.Slug]; !ok || token != updated.Token {
		t.Fatal("expected slug index to map slug to token")
	}

	makePrivate := false
	updated, err = store.UpdateShare(share.Token, &makePrivate, nil)
	if err != nil {
		t.Fatalf("UpdateShare public false: %v", err)
	}
	if updated.IsPublic || updated.Slug != "" {
		t.Fatal("expected share to be private with empty slug")
	}

	allow := true
	updated, err = store.UpdateShare(share.Token, nil, &allow)
	if err != nil {
		t.Fatalf("UpdateShare allow true: %v", err)
	}
	if updated.Permission != PermissionComment {
		t.Fatalf("expected permission %q, got %q", PermissionComment, updated.Permission)
	}

	deny := false
	updated, err = store.UpdateShare(share.Token, nil, &deny)
	if err != nil {
		t.Fatalf("UpdateShare allow false: %v", err)
	}
	if updated.Permission != PermissionView {
		t.Fatalf("expected permission %q, got %q", PermissionView, updated.Permission)
	}
}

func TestComments_AddListUpdate(t *testing.T) {
	store := NewStore("")

	disabled, err := store.CreateShare(CreateShareInput{
		Title:         "Disabled",
		IsPublic:      true,
		AllowComments: false,
	})
	if err != nil {
		t.Fatalf("CreateShare disabled: %v", err)
	}
	_, err = store.AddComment(disabled.Token, CommentInput{Author: "A", Message: "Hi"})
	if !errors.Is(err, ErrCommentsDisabled) {
		t.Fatalf("expected ErrCommentsDisabled, got %v", err)
	}

	enabled, err := store.CreateShare(CreateShareInput{
		Title:         "Enabled",
		IsPublic:      true,
		AllowComments: true,
	})
	if err != nil {
		t.Fatalf("CreateShare enabled: %v", err)
	}
	comment, err := store.AddComment(enabled.Token, CommentInput{Author: "Ana", Message: "First"})
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	if comment.ID == "" {
		t.Fatal("expected comment ID to be set")
	}

	comments, err := store.ListComments(enabled.Token)
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if len(comments) != 1 || comments[0].Message != "First" {
		t.Fatalf("unexpected comments: %+v", comments)
	}

	updated, err := store.UpdateComment(enabled.Token, comment.ID, true)
	if err != nil {
		t.Fatalf("UpdateComment: %v", err)
	}
	if !updated.Resolved {
		t.Fatal("expected comment to be resolved")
	}
}

func TestStorePersistence_RoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "store.json")

	store := NewStore(path)
	share, err := store.CreateShare(CreateShareInput{
		Title:         "Persisted",
		IsPublic:      true,
		AllowComments: true,
	})
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}
	_, err = store.AddComment(share.Token, CommentInput{Author: "Zoe", Message: "Saved"})
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	reloaded := NewStore(path)
	byToken, err := reloaded.GetShare(share.Token)
	if err != nil {
		t.Fatalf("GetShare token after reload: %v", err)
	}
	if byToken.Title != "Persisted" || !byToken.AllowComments {
		t.Fatalf("unexpected share after reload: %+v", byToken)
	}
	bySlug, err := reloaded.GetShare(share.Slug)
	if err != nil {
		t.Fatalf("GetShare slug after reload: %v", err)
	}
	if bySlug.Token != share.Token {
		t.Fatalf("expected slug lookup token %q, got %q", share.Token, bySlug.Token)
	}

	comments, err := reloaded.ListComments(share.Token)
	if err != nil {
		t.Fatalf("ListComments after reload: %v", err)
	}
	if len(comments) != 1 || comments[0].Message != "Saved" {
		t.Fatalf("unexpected persisted comments: %+v", comments)
	}
}
