package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// TemplateListResponse represents available templates
type TemplateListResponse struct {
	Templates []TemplateInfo `json:"templates"`
}

// TemplateInfo represents metadata about a template
type TemplateInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Format      string `json:"format"`
}

// TemplatePreviewRequest represents the request for custom template preview
type TemplatePreviewRequest struct {
	TemplateContent string `json:"template_content" binding:"required"`
	SampleData      string `json:"sample_data"`
}

// TemplatePreviewResponse represents the response for custom template preview
type TemplatePreviewResponse struct {
	Output   string              `json:"output"`
	Error    string              `json:"error,omitempty"`
	Warnings []converter.Warning `json:"warnings"`
}

// GetTemplates handles GET /api/mdflow/templates
// Returns available MDFlow templates with metadata
func (h *MDFlowHandler) GetTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, TemplateListResponse{
		Templates: []TemplateInfo{
			{Name: "spec", Description: "Structured specification output", Format: "spec"},
			{Name: "table", Description: "Simple markdown table output", Format: "table"},
		},
	})
}

// PreviewTemplate handles POST /api/mdflow/templates/preview
// Renders sample data using a custom template
func (h *MDFlowHandler) PreviewTemplate(c *gin.Context) {
	var req TemplatePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "template_content is required"})
		return
	}

	// Limit template size
	if len(req.TemplateContent) > 50000 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "template_content exceeds 50KB limit"})
		return
	}

	// Use sample data or default sample
	sampleData := req.SampleData
	if sampleData == "" {
		sampleData = defaultSampleData
	}

	// Parse sample data to SpecDoc
	specDoc, err := converter.BuildSpecDocFromPaste(sampleData)
	if err != nil {
		c.JSON(http.StatusOK, TemplatePreviewResponse{
			Output: "",
			Error:  "Failed to parse sample data: " + err.Error(),
		})
		return
	}

	// Render with custom template
	output, err := h.renderer.RenderCustom(specDoc, req.TemplateContent)
	if err != nil {
		c.JSON(http.StatusOK, TemplatePreviewResponse{
			Output: "",
			Error:  "Template error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TemplatePreviewResponse{
		Output:   output,
		Warnings: specDoc.Warnings,
	})
}

// GetTemplateInfo handles GET /api/mdflow/templates/info
// Returns available template variables and functions
func (h *MDFlowHandler) GetTemplateInfo(c *gin.Context) {
	info := h.renderer.GetTemplateInfo()
	c.JSON(http.StatusOK, info)
}

// GetTemplateContent handles GET /api/mdflow/templates/:name
// Returns the content of a built-in template
func (h *MDFlowHandler) GetTemplateContent(c *gin.Context) {
	name := strings.ToLower(strings.TrimSpace(c.Param("name")))
	if name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "template name is required"})
		return
	}

	canonicalName, ok := resolveTemplateContentName(name)
	if !ok {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "template not found"})
		return
	}

	content := h.renderer.GetTemplateContent(canonicalName)
	if content == "" {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    canonicalName,
		"content": content,
	})
}

func resolveTemplateContentName(name string) (canonicalName string, ok bool) {
	switch name {
	case "spec", "table":
		return name, true
	default:
		return "", false
	}
}

// Default sample data for template preview
const defaultSampleData = `Feature	Scenario	Instructions	Expected	Priority	Type	Notes
User Authentication	Valid Login	1. Enter username
2. Enter password
3. Click login button	Dashboard should display with user name	High	Positive	Core feature
User Authentication	Invalid Password	1. Enter valid username
2. Enter wrong password
3. Click login button	Error message: "Invalid credentials"	High	Negative	Security test
Profile Management	Update Profile	1. Go to settings
2. Change display name
3. Click save	Profile updated successfully message shown	Medium	Positive	
Profile Management	Upload Avatar	1. Click avatar
2. Select image file
3. Confirm upload	New avatar displayed	Low	Positive	Max 5MB`
