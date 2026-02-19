package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// APITelemetryEvents emits structured event logs for key MVP funnel endpoints.
func APITelemetryEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		eventName, inputSource, tracked := eventMetadata(c.Request.Method, c.FullPath(), c.Request.URL.Path)
		if !tracked {
			return
		}

		status := "success"
		if c.Writer.Status() >= 400 {
			status = "error"
		}

		slog.Info("api event",
			"event_name", eventName,
			"event_time", time.Now().UTC().Format(time.RFC3339),
			"request_id", GetRequestID(c),
			"status", status,
			"http_status", c.Writer.Status(),
			"duration_ms", time.Since(startedAt).Milliseconds(),
			"input_source", inputSource,
			"path", c.Request.URL.Path,
		)

		event := TelemetryEvent{
			EventName:   eventName,
			EventTime:   time.Now().UTC(),
			Status:      status,
			InputSource: inputSource,
			DurationMS:  time.Since(startedAt).Milliseconds(),
			HTTPStatus:  c.Writer.Status(),
			Path:        c.Request.URL.Path,
			RequestID:   GetRequestID(c),
			Source:      "backend",
		}

		// Attach AI metadata if set by handler (e.g. convert_handler.recordTokenUsage)
		if model, ok := c.Get("ai_model"); ok {
			event.AIModel, _ = model.(string)
		}
		if cost, ok := c.Get("ai_estimated_cost_usd"); ok {
			event.AIEstimatedCostUSD, _ = cost.(float64)
		}
		if tokens, ok := c.Get("ai_input_tokens"); ok {
			event.AIInputTokens, _ = tokens.(int64)
		}
		if tokens, ok := c.Get("ai_output_tokens"); ok {
			event.AIOutputTokens, _ = tokens.(int64)
		}

		RecordTelemetryEvent(event)
	}
}

func eventMetadata(method, fullPath, rawPath string) (eventName string, inputSource string, tracked bool) {
	if method != "POST" {
		return "", "", false
	}

	path := fullPath
	if path == "" {
		path = rawPath
	}

	switch path {
	case "/api/mdflow/preview", "/api/mdflow/tsv/preview", "/api/mdflow/xlsx/preview",
		"/api/v1/mdflow/preview", "/api/v1/mdflow/tsv/preview", "/api/v1/mdflow/xlsx/preview":
		return "api_preview_completed", previewInputSource(path), true
	case "/api/mdflow/paste", "/api/mdflow/tsv", "/api/mdflow/xlsx",
		"/api/v1/mdflow/paste", "/api/v1/mdflow/tsv", "/api/v1/mdflow/xlsx":
		return "api_convert_completed", convertInputSource(path), true
	case "/api/mdflow/ai/suggest", "/api/v1/mdflow/ai/suggest":
		return "api_ai_suggest_completed", "paste", true
	case "/api/share":
		return "api_share_created", "share", true
	default:
		return "", "", false
	}
}

func previewInputSource(path string) string {
	switch path {
	case "/api/mdflow/tsv/preview", "/api/v1/mdflow/tsv/preview":
		return "tsv"
	case "/api/mdflow/xlsx/preview", "/api/v1/mdflow/xlsx/preview":
		return "xlsx"
	default:
		return "paste"
	}
}

func convertInputSource(path string) string {
	switch path {
	case "/api/mdflow/tsv", "/api/v1/mdflow/tsv":
		return "tsv"
	case "/api/mdflow/xlsx", "/api/v1/mdflow/xlsx":
		return "xlsx"
	default:
		return "paste"
	}
}
