package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/services"
)

type CommentHandler struct {
	commentService *services.CommentService
}

func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// AddComment adds a comment to a spec
// POST /spec/:id/comments
func (h *CommentHandler) AddComment(c *gin.Context) {
	specID := c.Param("id")
	userID := c.GetString("user_id")

	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	comment, err := h.commentService.AddComment(c.Request.Context(), specID, userID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.CommentResponse{
		Message: "comment added successfully",
		Comment: comment,
	})
}

// GetComments retrieves all comments for a spec with threaded replies
// GET /spec/:id/comments
func (h *CommentHandler) GetComments(c *gin.Context) {
	specID := c.Param("id")
	
	comments, err := h.commentService.GetCommentsWithReplies(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if comments == nil {
		comments = []*models.Comment{}
	}
	
	c.JSON(http.StatusOK, models.CommentResponse{
		Comments: comments,
	})
}

// UpdateComment updates a comment
// PUT /spec/:id/comments/:comment_id
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	userID := c.GetString("user_id")

	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	comment, err := h.commentService.UpdateComment(c.Request.Context(), commentID, userID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.CommentResponse{
		Message: "comment updated successfully",
		Comment: comment,
	})
}

// DeleteComment deletes a comment
// DELETE /spec/:id/comments/:comment_id
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	userID := c.GetString("user_id")
	
	err := h.commentService.DeleteComment(c.Request.Context(), commentID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})
}

// AddReply adds a reply to a comment
// POST /spec/:id/comments/:comment_id/reply
func (h *CommentHandler) AddReply(c *gin.Context) {
	specID := c.Param("id")
	parentCommentID := c.Param("comment_id")
	userID := c.GetString("user_id")
	
	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	
	reply, err := h.commentService.AddReply(c.Request.Context(), specID, parentCommentID, userID, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, models.CommentResponse{
		Message: "reply added successfully",
		Comment: reply,
	})
}

// PHASE 7: Comment edit endpoints

// EditComment updates a comment (only by author)
// PUT /spec/:id/comments/:comment_id
func (h *CommentHandler) EditComment(c *gin.Context) {
	_ = c.Param("id")       // specID - validated by middleware but not needed here
	commentID := c.Param("comment_id")
	userID := c.GetString("user_id")

	var req models.CommentEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	comment, err := h.commentService.EditComment(c.Request.Context(), commentID, userID, req.Content)
	if err != nil {
		// Check if it's an ownership error
		if err.Error() == "only comment author can edit" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "comment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load edit history
	comment, err = h.commentService.GetCommentWithEditHistory(c.Request.Context(), commentID, true)
	if err == nil {
		c.JSON(http.StatusOK, models.CommentResponse{
			Message: "comment updated successfully",
			Comment: comment,
		})
	} else {
		c.JSON(http.StatusOK, models.CommentResponse{
			Message: "comment updated successfully",
			Comment: comment,
		})
	}
}

// GetCommentEdits retrieves the edit history of a comment
// GET /comments/:comment_id/edits
func (h *CommentHandler) GetCommentEdits(c *gin.Context) {
	commentID := c.Param("comment_id")

	edits, err := h.commentService.GetCommentEdits(c.Request.Context(), commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comment_id": commentID,
		"edits":      edits,
	})
}
