package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// StreamHandler handles the SSE streaming conversion endpoint.
type StreamHandler struct {
	converter *converter.Converter
	cfg       *config.Config
	byokCache *AIServiceProvider
}

// NewStreamHandler creates a StreamHandler.  All parameters are optional
// (nil falls back to defaults), matching the pattern of ConvertHandler.
func NewStreamHandler(conv *converter.Converter, cfg *config.Config, byokCache *AIServiceProvider) *StreamHandler {
	if conv == nil {
		conv = converter.NewConverter()
	}
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if byokCache == nil {
		byokCache = NewAIServiceProvider(cfg)
	}
	return &StreamHandler{
		converter: conv,
		cfg:       cfg,
		byokCache: byokCache,
	}
}

// ConvertStream handles POST /api/v1/mdflow/convert/stream.
// It accepts the same JSON body as POST /api/v1/mdflow/paste and emits
// Server-Sent Events as the pipeline progresses.
//
// SSE event sequence:
//
//	event: progress  data: {"phase":"parsing",   "percent":20}
//	event: progress  data: {"phase":"mapping",   "percent":50}
//	event: progress  data: {"phase":"rendering", "percent":80}
//	event: complete  data: {"phase":"complete",  "percent":100}
//	event: result    data: <MDFlowConvertResponse JSON>
//
// On error:
//
//	event: error     data: {"error":"<message>"}
func (h *StreamHandler) ConvertStream(c *gin.Context) {
	// ── SSE headers ──────────────────────────────────────────────────────────
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no") // disable nginx buffering

	// ── Parse request ────────────────────────────────────────────────────────
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxPasteBytes+4<<10)

	var req PasteConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeSSEEvent(c, "error", map[string]string{"error": "request body exceeds limit"})
			return
		}
		writeSSEEvent(c, "error", map[string]string{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.PasteText) == "" {
		writeSSEEvent(c, "error", map[string]string{"error": "paste_text is required"})
		return
	}

	if int64(len(req.PasteText)) > h.cfg.MaxPasteBytes {
		writeSSEEvent(c, "error", map[string]string{
			"error": fmt.Sprintf("paste_text exceeds %s limit", humanSize(h.cfg.MaxPasteBytes)),
		})
		return
	}

	normalizedTemplate, normalizedFormat, err := normalizeTemplateAndFormat(req.Template, req.Format)
	if err != nil {
		writeSSEEvent(c, "error", map[string]string{"error": err.Error()})
		return
	}
	req.Template = normalizedTemplate
	req.Format = normalizedFormat

	slog.Info("mdflow.ConvertStream",
		"template", req.Template,
		"format", req.Format,
		"pasteBytes", len(req.PasteText),
	)

	// ── Flusher check ────────────────────────────────────────────────────────
	flusher, canFlush := c.Writer.(http.Flusher)

	// ── Context with timeout ─────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(c.Request.Context(), 150*time.Second)
	defer cancel()

	// ── BYOK-aware converter ─────────────────────────────────────────────────
	conv := h.byokCache.GetConverterForRequest(c, h.converter)

	// Build callback: write SSE event and flush to the client.
	callback := func(event converter.StreamEvent) {
		select {
		case <-ctx.Done():
			// Client disconnected or timed out; stop sending.
			return
		default:
		}
		writeSSEEvent(c, event.Event, event.Data)
		if canFlush {
			flusher.Flush()
		}
	}

	// ── Run streaming pipeline ───────────────────────────────────────────────
	result, err := conv.ConvertPasteStreaming(ctx, req.PasteText, req.Template, req.Format, callback)
	if err != nil {
		if ctx.Err() != nil {
			slog.Info("mdflow.ConvertStream: context done", "cause", ctx.Err())
			return
		}
		slog.Error("mdflow.ConvertStream failed", "error", err)
		writeSSEEvent(c, "error", map[string]string{"error": "conversion failed"})
		if canFlush {
			flusher.Flush()
		}
		return
	}

	// ── Send final result event ───────────────────────────────────────────────
	finalResponse := MDFlowConvertResponse{
		MDFlow:      result.MDFlow,
		Warnings:    result.Warnings,
		Meta:        result.Meta,
		Format:      req.Format,
		Template:    req.Template,
		NeedsReview: RequiresReview(result.Meta, result.Warnings),
	}
	writeSSEEvent(c, "result", finalResponse)
	if canFlush {
		flusher.Flush()
	}
}

// writeSSEEvent marshals data to JSON and writes a single SSE event.
// Format:
//
//	event: <eventType>\n
//	data: <json>\n
//	\n
func writeSSEEvent(c *gin.Context, eventType string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		slog.Warn("writeSSEEvent: marshal failed", "event", eventType, "error", err)
		return
	}
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", eventType, payload)
}
