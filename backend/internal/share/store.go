package share

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Permission string

const (
	PermissionView    Permission = "view"
	PermissionComment Permission = "comment"
)

var (
	ErrShareNotFound     = errors.New("share not found")
	ErrSlugExists        = errors.New("slug already exists")
	ErrInvalidSlug       = errors.New("invalid slug")
	ErrCommentsDisabled  = errors.New("comments disabled")
	ErrInvalidPermission = errors.New("invalid permission")
	ErrStoreFull         = errors.New("share limit exceeded")
)

type Share struct {
	Token            string     `json:"token"`
	Slug             string     `json:"slug"`
	Title            string     `json:"title"`
	Template         string     `json:"template"`
	MDFlow           string     `json:"mdflow"`
	IsPublic         bool       `json:"is_public"`
	AllowComments    bool       `json:"allow_comments"`
	Permission       Permission `json:"permission"`
	CreatedAt        time.Time  `json:"created_at"`
	Comments         []Comment  `json:"comments"`
	ResolutionEvents []Event    `json:"resolution_events"`
}

type Comment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Resolved  bool      `json:"resolved"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateShareInput struct {
	Title         string
	Template      string
	MDFlow        string
	Slug          string
	IsPublic      bool
	AllowComments bool
	Permission    Permission
}

type CommentInput struct {
	Author  string
	Message string
}

type Store struct {
	mu        sync.RWMutex
	shares    map[string]*Share
	slugIndex map[string]string
	path      string
	maxShares int // Maximum number of shares to prevent OOM (0 = unlimited)
}

type storeSnapshot struct {
	Shares map[string]*Share `json:"shares"`
}

func NewStore(path string) *Store {
	return NewStoreWithLimit(path, 10000) // Default limit: 10,000 shares
}

func NewStoreWithLimit(path string, maxShares int) *Store {
	store := &Store{
		shares:    make(map[string]*Share),
		slugIndex: make(map[string]string),
		path:      strings.TrimSpace(path),
		maxShares: maxShares,
	}
	store.loadFromDisk()

	// Ensure maxShares is at least as large as existing data to avoid locking out users
	if len(store.shares) > store.maxShares {
		store.maxShares = len(store.shares)
		slog.Warn("maxShares adjusted to accommodate existing data", "loaded_shares", len(store.shares), "new_max", store.maxShares)
	}
	return store
}

func (s *Store) CreateShare(input CreateShareInput) (*Share, error) {
	if input.Permission != "" && input.Permission != PermissionView && input.Permission != PermissionComment {
		return nil, ErrInvalidPermission
	}

	permission := input.Permission
	if permission == "" {
		if input.AllowComments {
			permission = PermissionComment
		} else {
			permission = PermissionView
		}
	}

	token, err := generateToken(18)
	if err != nil {
		return nil, err
	}

	slug := strings.TrimSpace(input.Slug)
	if input.IsPublic {
		slug = normalizeSlug(slug, input.Title)
	} else {
		slug = ""
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check max shares limit
	if s.maxShares > 0 && len(s.shares) >= s.maxShares {
		return nil, ErrStoreFull
	}

	if input.IsPublic {
		if slug == "" {
			slug = "spec-" + randomSlug(6)
		}
		if !isValidSlug(slug) {
			return nil, ErrInvalidSlug
		}
		if _, exists := s.slugIndex[slug]; exists {
			return nil, ErrSlugExists
		}
		if _, exists := s.shares[slug]; exists {
			return nil, ErrSlugExists
		}
	}

	share := &Share{
		Token:            token,
		Slug:             slug,
		Title:            input.Title,
		Template:         input.Template,
		MDFlow:           input.MDFlow,
		IsPublic:         input.IsPublic,
		AllowComments:    input.AllowComments,
		Permission:       permission,
		CreatedAt:        time.Now().UTC(),
		Comments:         []Comment{},
		ResolutionEvents: []Event{},
	}

	s.shares[token] = share
	if slug != "" {
		s.slugIndex[slug] = token
	}
	if err := s.saveToDiskLocked(); err != nil {
		return nil, err
	}

	return share, nil
}

func (s *Store) GetShare(key string) (*Share, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	share, ok := s.shares[key]
	if ok {
		return s.cloneShare(share), nil
	}

	token, exists := s.slugIndex[key]
	if !exists {
		return nil, ErrShareNotFound
	}

	share, ok = s.shares[token]
	if !ok {
		return nil, ErrShareNotFound
	}

	return s.cloneShare(share), nil
}

func (s *Store) ListPublic() []*Share {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Share, 0)
	for _, share := range s.shares {
		if share.IsPublic && share.Slug != "" {
			result = append(result, s.cloneShare(share))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

func (s *Store) UpdateShare(key string, isPublic *bool, allowComments *bool) (*Share, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	share, err := s.getShareLocked(key)
	if err != nil {
		return nil, err
	}

	if isPublic != nil {
		if *isPublic && share.Slug == "" {
			share.Slug = "spec-" + randomSlug(6)
			if !isValidSlug(share.Slug) {
				return nil, ErrInvalidSlug
			}
			if _, exists := s.slugIndex[share.Slug]; exists {
				return nil, ErrSlugExists
			}
			s.slugIndex[share.Slug] = share.Token
		}
		if !*isPublic && share.Slug != "" {
			delete(s.slugIndex, share.Slug)
			share.Slug = ""
		}
		share.IsPublic = *isPublic
	}

	if allowComments != nil {
		share.AllowComments = *allowComments
		if *allowComments {
			share.Permission = PermissionComment
		} else {
			share.Permission = PermissionView
		}
	}

	if err := s.saveToDiskLocked(); err != nil {
		return nil, err
	}

	return share, nil
}

func (s *Store) ListComments(key string) ([]Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	share, err := s.getShareLocked(key)
	if err != nil {
		return nil, err
	}

	comments := make([]Comment, len(share.Comments))
	copy(comments, share.Comments)
	return comments, nil
}

func (s *Store) AddComment(key string, input CommentInput) (Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	share, err := s.getShareLocked(key)
	if err != nil {
		return Comment{}, err
	}

	if !share.AllowComments || share.Permission != PermissionComment {
		return Comment{}, ErrCommentsDisabled
	}

	comment := Comment{
		ID:        generateCommentID(),
		Author:    input.Author,
		Message:   input.Message,
		Resolved:  false,
		CreatedAt: time.Now().UTC(),
	}

	share.Comments = append(share.Comments, comment)
	if err := s.saveToDiskLocked(); err != nil {
		return Comment{}, err
	}
	return comment, nil
}

func (s *Store) UpdateComment(key, commentID string, resolved bool) (Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	share, err := s.getShareLocked(key)
	if err != nil {
		return Comment{}, err
	}

	for i, comment := range share.Comments {
		if comment.ID == commentID {
			share.Comments[i].Resolved = resolved

			// Emit resolution event if resolving
			if resolved {
				event := Event{
					EventType: "comment_resolved",
					Timestamp: time.Now().UTC(),
					CommentID: commentID,
					Data:      comment.Author, // Store author in data field
				}
				share.ResolutionEvents = append(share.ResolutionEvents, event)
			}

			if err := s.saveToDiskLocked(); err != nil {
				return Comment{}, err
			}
			return share.Comments[i], nil
		}
	}

	return Comment{}, ErrShareNotFound
}

func (s *Store) getShareLocked(key string) (*Share, error) {
	if share, ok := s.shares[key]; ok {
		return share, nil
	}

	token, exists := s.slugIndex[key]
	if !exists {
		return nil, ErrShareNotFound
	}

	share, ok := s.shares[token]
	if !ok {
		return nil, ErrShareNotFound
	}

	return share, nil
}

func (s *Store) loadFromDisk() {
	if s.path == "" {
		return
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}

	var snapshot storeSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return
	}

	for token, share := range snapshot.Shares {
		if token == "" || share == nil {
			continue
		}
		share.Token = token
		s.shares[token] = share
		if share.Slug != "" {
			s.slugIndex[share.Slug] = token
		}
	}
}

func (s *Store) saveToDiskLocked() error {
	if s.path == "" {
		return nil
	}

	dir := filepath.Dir(s.path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	snapshot := storeSnapshot{Shares: s.shares}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func isValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 48 {
		return false
	}
	return slugPattern.MatchString(slug)
}

func normalizeSlug(slug, title string) string {
	base := strings.TrimSpace(slug)
	if base == "" {
		base = title
	}
	base = strings.ToLower(base)

	var builder strings.Builder
	lastDash := false
	for _, r := range base {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}

	slugified := strings.Trim(builder.String(), "-")
	if len(slugified) > 48 {
		slugified = slugified[:48]
		slugified = strings.TrimRight(slugified, "-")
	}
	return slugified
}

func generateToken(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func generateCommentID() string {
	token, err := generateToken(6)
	if err != nil {
		return "cmt-" + time.Now().Format("20060102150405")
	}
	return "cmt-" + token
}

func (s *Store) cloneShare(share *Share) *Share {
	if share == nil {
		return nil
	}
	comments := make([]Comment, len(share.Comments))
	copy(comments, share.Comments)
	events := make([]Event, len(share.ResolutionEvents))
	copy(events, share.ResolutionEvents)
	return &Share{
		Token:            share.Token,
		Slug:             share.Slug,
		Title:            share.Title,
		Template:         share.Template,
		MDFlow:           share.MDFlow,
		IsPublic:         share.IsPublic,
		AllowComments:    share.AllowComments,
		Permission:       share.Permission,
		CreatedAt:        share.CreatedAt,
		Comments:         comments,
		ResolutionEvents: events,
	}
}

func randomSlug(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	if length <= 0 {
		return ""
	}

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return time.Now().Format("0601021504")
	}

	for i := 0; i < length; i++ {
		buf[i] = chars[int(buf[i])%len(chars)]
	}

	return string(buf)
}
