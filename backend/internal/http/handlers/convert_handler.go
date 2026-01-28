package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/converters"
	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/services"
)

type ConvertHandler struct {
	templateService *services.TemplateService
	converter       *converters.MarkdownConverter
}

func NewConvertHandler(templateService *services.TemplateService) *ConvertHandler {
	return &ConvertHandler{
		templateService: templateService,
		converter:       converters.NewMarkdownConverter(),
	}
}

type ConvertRequest struct {
	TableData  *models.TableData `json:"table_data" binding:"required"`
	TemplateID *string           `json:"template_id"`
}

type ConvertResponse struct {
	Markdown string `json:"markdown"`
}

// ConvertToMarkdown converts table data to markdown using a template
func (h *ConvertHandler) ConvertToMarkdown(c *gin.Context) {
	var req ConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var tmpl *models.Template
	var err error

	if req.TemplateID != nil {
		// Use specified template
		tmpl, err = h.templateService.GetTemplate(c.Request.Context(), *req.TemplateID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "template not found"})
			return
		}
	} else {
		// Use default template
		tmpl = h.converter.GetDefaultTemplate()
	}

	// Convert to markdown
	markdown, err := h.converter.Convert(tmpl, req.TableData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ConvertResponse{Markdown: markdown})
}
