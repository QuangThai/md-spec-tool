package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/services"
)

type TemplateHandler struct {
	templateService *services.TemplateService
}

func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

type TemplateResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateTemplateRequest struct {
	Name    string `json:"name" binding:"required,min=1"`
	Content string `json:"content" binding:"required"`
}

// ListTemplates returns all templates (built-in + user's)
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	userID := c.GetString("user_id")

	// Get user templates
	userTemplates, err := h.templateService.ListTemplates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Always include default template
	defaultTmpl := &models.Template{
		ID:   "default",
		Name: "Default",
		Content: `# {{.SheetName}}

{{if gt (len .Rows) 0}}
| {{join " | " .Headers}} |
| {{range .Headers}}--- |{{end}}
{{range .Rows}}| {{range .Headers}}{{index . . }} | {{end}}
{{end}}

**Total Records**: {{.Count}}
{{else}}
No data rows found.
{{end}}`,
	}

	var responses []TemplateResponse
	responses = append(responses, TemplateResponse{
		ID:      defaultTmpl.ID,
		Name:    defaultTmpl.Name,
		Content: defaultTmpl.Content,
	})

	for _, tmpl := range userTemplates {
		responses = append(responses, TemplateResponse{
			ID:        tmpl.ID,
			Name:      tmpl.Name,
			Content:   tmpl.Content,
			CreatedAt: tmpl.CreatedAt.String(),
			UpdatedAt: tmpl.UpdatedAt.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": responses,
		"count":     len(responses),
	})
}

// GetTemplate returns a specific template
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")

	// Check if it's the default template
	if templateID == "default" {
		defaultTmpl := &models.Template{
			ID:   "default",
			Name: "Default",
			Content: `# {{.SheetName}}

{{if gt (len .Rows) 0}}
| {{join " | " .Headers}} |
| {{range .Headers}}--- |{{end}}
{{range .Rows}}| {{range .Headers}}{{index . . }} | {{end}}
{{end}}

**Total Records**: {{.Count}}
{{else}}
No data rows found.
{{end}}`,
		}

		c.JSON(http.StatusOK, TemplateResponse{
			ID:      defaultTmpl.ID,
			Name:    defaultTmpl.Name,
			Content: defaultTmpl.Content,
		})
		return
	}

	// Get user template
	tmpl, err := h.templateService.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "template not found"})
		return
	}

	userID := c.GetString("user_id")
	if tmpl.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		return
	}

	c.JSON(http.StatusOK, TemplateResponse{
		ID:        tmpl.ID,
		Name:      tmpl.Name,
		Content:   tmpl.Content,
		CreatedAt: tmpl.CreatedAt.String(),
		UpdatedAt: tmpl.UpdatedAt.String(),
	})
}

// CreateTemplate creates a new user template
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user_id not found"})
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	tmpl := &models.Template{
		UserID:  userID,
		Name:    req.Name,
		Content: req.Content,
	}

	if err := h.templateService.CreateTemplate(c.Request.Context(), tmpl); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, TemplateResponse{
		ID:        tmpl.ID,
		Name:      tmpl.Name,
		Content:   tmpl.Content,
		CreatedAt: tmpl.CreatedAt.String(),
		UpdatedAt: tmpl.UpdatedAt.String(),
	})
}
