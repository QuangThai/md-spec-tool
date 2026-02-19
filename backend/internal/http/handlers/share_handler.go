package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/share"
)

type ShareHandler struct {
	store share.StoreInterface
}

// NewShareHandler creates a new ShareHandler with a store
func NewShareHandler(store share.StoreInterface) *ShareHandler {
	return &ShareHandler{store: store}
}

// NewShareHandlerWithStore creates a new ShareHandler with a concrete Store (backward compatibility)
func NewShareHandlerWithStore(store *share.Store) *ShareHandler {
	return NewShareHandler(store)
}

type CreateShareRequest struct {
	Title         string `json:"title"`
	Template      string `json:"template"`
	MDFlow        string `json:"mdflow" binding:"required"`
	Slug          string `json:"slug"`
	IsPublic      bool   `json:"is_public"`
	AllowComments bool   `json:"allow_comments"`
	Permission    string `json:"permission"`
}

type ShareResponse struct {
	Token            string          `json:"token"`
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Template         string          `json:"template"`
	MDFlow           string          `json:"mdflow"`
	IsPublic         bool            `json:"is_public"`
	AllowComments    bool            `json:"allow_comments"`
	Permission       string          `json:"permission"`
	CreatedAt        string          `json:"created_at"`
	ResolutionEvents []EventResponse `json:"resolution_events"`
}

type EventResponse struct {
	EventType string `json:"event_type"`
	Timestamp string `json:"timestamp"`
	CommentID string `json:"comment_id"`
	Data      string `json:"data"`
}

type CommentResponse struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Message   string `json:"message"`
	Resolved  bool   `json:"resolved"`
	CreatedAt string `json:"created_at"`
}

type CreateCommentRequest struct {
	Author  string `json:"author"`
	Message string `json:"message" binding:"required"`
}

type UpdateCommentRequest struct {
	Resolved bool   `json:"resolved"`
	Token    string `json:"token"` // Optional: share token for permission validation
}

type UpdateShareRequest struct {
	IsPublic      *bool `json:"is_public"`
	AllowComments *bool `json:"allow_comments"`
}

type CloneTemplateRequest struct {
	SourceShareSlug string `json:"source_share_slug"`
	SourceTemplate  string `json:"source_template"`
	NewTitle        string `json:"new_title"`
	TemplateContent string `json:"template_content"`
}

type CloneTemplateResponse struct {
	Token           string `json:"token"`
	Slug            string `json:"slug"`
	RedirectURL     string `json:"redirect_url"`
	SourceShareSlug string `json:"source_share_slug"`
}

func (h *ShareHandler) CreateShare(c *gin.Context) {
	const maxMDFlowBytes = 1 << 20
	const maxCreateShareBodyBytes = maxMDFlowBytes + (10 << 10)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateShareBodyBytes)
	var req CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "payload too large"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "mdflow is required"})
		return
	}

	if len([]byte(req.MDFlow)) > maxMDFlowBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "payload too large"})
		return
	}

	permission := share.Permission(strings.TrimSpace(req.Permission))
	if permission == "" && req.AllowComments {
		permission = share.PermissionComment
	}
	if permission == "" {
		permission = share.PermissionView
	}

	created, err := h.store.CreateShare(share.CreateShareInput{
		Title:         strings.TrimSpace(req.Title),
		Template:      strings.TrimSpace(req.Template),
		MDFlow:        req.MDFlow,
		Slug:          strings.TrimSpace(req.Slug),
		IsPublic:      req.IsPublic,
		AllowComments: req.AllowComments,
		Permission:    permission,
	})
	if err != nil {
		switch err {
		case share.ErrSlugExists:
			c.JSON(http.StatusConflict, ErrorResponse{Error: "slug already exists"})
		case share.ErrInvalidSlug:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid slug"})
		case share.ErrInvalidPermission:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid permission"})
		case share.ErrStoreFull:
			c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "share limit exceeded"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create share"})
		}
		return
	}

	c.JSON(http.StatusOK, toShareResponse(created))
}

func (h *ShareHandler) GetShare(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "share key is required"})
		return
	}

	result, err := h.store.GetShare(key)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "share not found"})
		return
	}

	c.JSON(http.StatusOK, toShareResponse(result))
}

func (h *ShareHandler) ListPublic(c *gin.Context) {
	items := h.store.ListPublic()
	response := make([]share.ShareSummary, 0, len(items))
	for _, item := range items {
		response = append(response, share.ShareSummary{
			Slug:      item.Slug,
			Title:     item.Title,
			Template:  item.Template,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": response})
}

func (h *ShareHandler) ListComments(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "share key is required"})
		return
	}

	comments, err := h.store.ListComments(key)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "share not found"})
		return
	}

	result, err := h.store.GetShare(key)
	if err == nil && !result.AllowComments {
		c.JSON(http.StatusOK, gin.H{"items": []CommentResponse{}})
		return
	}

	response := make([]CommentResponse, 0, len(comments))
	for _, comment := range comments {
		response = append(response, CommentResponse{
			ID:        comment.ID,
			Author:    comment.Author,
			Message:   comment.Message,
			Resolved:  comment.Resolved,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": response})
}

func (h *ShareHandler) CreateComment(c *gin.Context) {
	const maxCommentBytes = 5 * 1024
	const maxCreateCommentBodyBytes = 8 * 1024

	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "share key is required"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateCommentBodyBytes)
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "payload too large"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "message is required"})
		return
	}

	if len([]byte(req.Message)) > maxCommentBytes {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "comment too long"})
		return
	}

	author := strings.TrimSpace(req.Author)
	if author == "" {
		author = "Anonymous"
	}

	comment, err := h.store.AddComment(key, share.CommentInput{
		Author:  author,
		Message: strings.TrimSpace(req.Message),
	})
	if err != nil {
		switch err {
		case share.ErrCommentsDisabled:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "comments are disabled"})
		default:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "share not found"})
		}
		return
	}

	c.JSON(http.StatusOK, CommentResponse{
		ID:        comment.ID,
		Author:    comment.Author,
		Message:   comment.Message,
		Resolved:  comment.Resolved,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	})
}

func (h *ShareHandler) UpdateComment(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	commentID := strings.TrimSpace(c.Param("commentId"))
	if key == "" || commentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "share key and comment id are required"})
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "resolved is required"})
		return
	}

	// Permission validation: only share creator (with token) can resolve comments
	shareData, err := h.store.GetShare(key)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "share not found"})
		return
	}

	// Verify token if provided (allows share creator to resolve comments)
	// If no token provided, allow resolution (backward compatible)
	if req.Token != "" && req.Token != shareData.Token {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "insufficient permissions to resolve comment"})
		return
	}

	comment, err := h.store.UpdateComment(key, commentID, req.Resolved)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "comment not found"})
		return
	}

	c.JSON(http.StatusOK, CommentResponse{
		ID:        comment.ID,
		Author:    comment.Author,
		Message:   comment.Message,
		Resolved:  comment.Resolved,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	})
}

func (h *ShareHandler) GetShareEvents(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "share key is required"})
		return
	}

	share, err := h.store.GetShare(key)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "share not found"})
		return
	}

	// Map internal Event structs to EventResponse
	events := make([]EventResponse, 0)
	if share.ResolutionEvents != nil {
		for _, e := range share.ResolutionEvents {
			events = append(events, EventResponse{
				EventType: e.EventType,
				Timestamp: e.Timestamp.Format(time.RFC3339),
				CommentID: e.CommentID,
				Data:      e.Data,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": events})
}

func (h *ShareHandler) UpdateShare(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "share key is required"})
		return
	}

	var req UpdateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid payload"})
		return
	}

	updated, err := h.store.UpdateShare(key, req.IsPublic, req.AllowComments)
	if err != nil {
		switch err {
		case share.ErrInvalidSlug:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid slug"})
		case share.ErrSlugExists:
			c.JSON(http.StatusConflict, ErrorResponse{Error: "slug already exists"})
		default:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "share not found"})
		}
		return
	}

	c.JSON(http.StatusOK, toShareResponse(updated))
}

// CloneTemplate handles POST /api/mdflow/clone-template
// Creates a new share from a template or source share
func (h *ShareHandler) CloneTemplate(c *gin.Context) {
	var req CloneTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid payload"})
		return
	}

	// Determine MDFlow content to clone
	var mdflowContent string
	var sourceSlug string

	if req.SourceShareSlug != "" {
		// Clone from existing share
		sourceSh, err := h.store.GetShare(req.SourceShareSlug)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "source share not found"})
			return
		}
		mdflowContent = sourceSh.MDFlow
		sourceSlug = sourceSh.Slug
	} else if req.TemplateContent != "" {
		// Clone from direct template content
		mdflowContent = req.TemplateContent
	} else {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "source_share_slug or template_content required"})
		return
	}

	// Create new share from template
	title := strings.TrimSpace(req.NewTitle)
	if title == "" {
		title = "Cloned Spec"
	}

	created, err := h.store.CreateShare(share.CreateShareInput{
		Title:         title,
		Template:      req.SourceTemplate,
		MDFlow:        mdflowContent,
		IsPublic:      false, // Clones are private by default
		AllowComments: false,
		Permission:    share.PermissionView,
	})
	if err != nil {
		switch err {
		case share.ErrStoreFull:
			c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "share limit exceeded"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to clone template"})
		}
		return
	}

	// TODO: Emit event "template_cloned" with source reference

	redirectURL := "/s/" + created.Slug
	c.JSON(http.StatusOK, CloneTemplateResponse{
		Token:           created.Token,
		Slug:            created.Slug,
		RedirectURL:     redirectURL,
		SourceShareSlug: sourceSlug,
	})
}

func isRequestBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}

	return strings.Contains(err.Error(), "request body too large")
}

// toShareResponse converts a Share model to ShareResponse, mapping resolution events
func toShareResponse(s *share.Share) ShareResponse {
	events := make([]EventResponse, 0)
	if s.ResolutionEvents != nil {
		for _, e := range s.ResolutionEvents {
			events = append(events, EventResponse{
				EventType: e.EventType,
				Timestamp: e.Timestamp.Format(time.RFC3339),
				CommentID: e.CommentID,
				Data:      e.Data,
			})
		}
	}

	return ShareResponse{
		Token:            s.Token,
		Slug:             s.Slug,
		Title:            s.Title,
		Template:         s.Template,
		MDFlow:           s.MDFlow,
		IsPublic:         s.IsPublic,
		AllowComments:    s.AllowComments,
		Permission:       string(s.Permission),
		CreatedAt:        s.CreatedAt.Format(time.RFC3339),
		ResolutionEvents: events,
	}
}
