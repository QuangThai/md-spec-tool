package handlers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
)

type TelemetryHandler struct{}

func NewTelemetryHandler() *TelemetryHandler {
	return &TelemetryHandler{}
}

type telemetryEventPayload struct {
	EventName          string  `json:"event_name"`
	EventTime          string  `json:"event_time"`
	SessionID          string  `json:"session_id"`
	Status             string  `json:"status"`
	InputSource        string  `json:"input_source"`
	TemplateType       string  `json:"template_type"`
	DurationMS         int64   `json:"duration_ms"`
	ErrorCode          string  `json:"error_code"`
	WarningCount       int     `json:"warning_count"`
	ConfidenceScore    float64 `json:"confidence_score"`
	NeedsReview        bool    `json:"needs_review"`
	TotalRows          int     `json:"total_rows"`
	AIModel            string  `json:"ai_model"`
	AIEstimatedCostUSD float64 `json:"ai_estimated_cost_usd"`
	AIInputTokens      int64   `json:"ai_input_tokens"`
	AIOutputTokens     int64   `json:"ai_output_tokens"`
}

type telemetryIngestRequest struct {
	Events []telemetryEventPayload `json:"events"`
}

type telemetryDashboardResponse struct {
	GeneratedAt string `json:"generated_at"`
	WindowHours int    `json:"window_hours"`
	Totals      struct {
		EventsTotal    int `json:"events_total"`
		FrontendEvents int `json:"frontend_events"`
		BackendEvents  int `json:"backend_events"`
	} `json:"totals"`
	Funnel struct {
		StudioOpened     int `json:"studio_opened"`
		InputProvided    int `json:"input_provided"`
		PreviewSucceeded int `json:"preview_succeeded"`
		ConvertSucceeded int `json:"convert_succeeded"`
		ShareCreated     int `json:"share_created"`
	} `json:"funnel"`
	KPIs struct {
		ActivationRate10m float64 `json:"activation_rate_10m"`
		TimeToValueMS     struct {
			Median int64 `json:"median"`
			P75    int64 `json:"p75"`
			P95    int64 `json:"p95"`
		} `json:"time_to_value_ms"`
		PreviewSuccessRate float64 `json:"preview_success_rate"`
		ConvertSuccessRate float64 `json:"convert_success_rate"`
	} `json:"kpis"`
	Reliability struct {
		API5xxRate          float64 `json:"api_5xx_rate"`
		P95PreviewLatencyMS int64   `json:"p95_preview_latency_ms"`
		P95ConvertLatencyMS int64   `json:"p95_convert_latency_ms"`
	} `json:"reliability"`
	Errors []struct {
		EventName string `json:"event_name"`
		Count     int    `json:"count"`
	} `json:"errors"`
	AICost struct {
		TotalCostUSD      float64 `json:"total_cost_usd"`
		AvgCostPerConvert float64 `json:"avg_cost_per_convert"`
		TotalInputTokens  int64   `json:"total_input_tokens"`
		TotalOutputTokens int64   `json:"total_output_tokens"`
		TotalAIRequests   int     `json:"total_ai_requests"`
		CostByModel       []struct {
			Model    string  `json:"model"`
			CostUSD  float64 `json:"cost_usd"`
			Requests int     `json:"requests"`
		} `json:"cost_by_model"`
	} `json:"ai_cost"`
}

// IngestEvents handles POST /api/telemetry/events.
func (h *TelemetryHandler) IngestEvents(c *gin.Context) {
	const maxBodyBytes = 256 << 10 // 256KB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)

	var batch telemetryIngestRequest
	if err := c.ShouldBindJSON(&batch); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "payload too large"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid telemetry payload"})
		return
	}

	recorded := 0
	for _, raw := range batch.Events {
		eventName := strings.TrimSpace(raw.EventName)
		if eventName == "" {
			continue
		}

		status := strings.TrimSpace(raw.Status)
		switch status {
		case "success", "error", "cancel":
		default:
			status = "success"
		}

		eventTime := time.Now().UTC()
		if strings.TrimSpace(raw.EventTime) != "" {
			if parsed, err := time.Parse(time.RFC3339, raw.EventTime); err == nil {
				eventTime = parsed.UTC()
			}
		}

		middleware.RecordTelemetryEvent(middleware.TelemetryEvent{
			EventName:          eventName,
			EventTime:          eventTime,
			SessionID:          strings.TrimSpace(raw.SessionID),
			Status:             status,
			InputSource:        strings.TrimSpace(raw.InputSource),
			TemplateType:       strings.TrimSpace(raw.TemplateType),
			DurationMS:         maxInt64(raw.DurationMS, 0),
			ErrorCode:          strings.TrimSpace(raw.ErrorCode),
			WarningCount:       maxInt(raw.WarningCount, 0),
			ConfidenceScore:    raw.ConfidenceScore,
			NeedsReview:        raw.NeedsReview,
			TotalRows:          maxInt(raw.TotalRows, 0),
			AIModel:            strings.TrimSpace(raw.AIModel),
			AIEstimatedCostUSD: raw.AIEstimatedCostUSD,
			AIInputTokens:      maxInt64(raw.AIInputTokens, 0),
			AIOutputTokens:     maxInt64(raw.AIOutputTokens, 0),
			Source:             "frontend",
		})
		recorded++
	}

	c.JSON(http.StatusAccepted, gin.H{"recorded": recorded})
}

// Dashboard handles GET /api/telemetry/dashboard.
func (h *TelemetryHandler) Dashboard(c *gin.Context) {
	hours := 24
	if raw := strings.TrimSpace(c.Query("hours")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 168 {
			hours = parsed
		}
	}

	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	events := middleware.SnapshotTelemetryEvents(since)

	resp := telemetryDashboardResponse{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		WindowHours: hours,
	}

	type sessionTimes struct {
		studio  time.Time
		convert time.Time
	}
	sessionMap := make(map[string]sessionTimes)

	previewDurations := make([]int64, 0)
	convertDurations := make([]int64, 0)
	ttvDurations := make([]int64, 0)
	errorCounts := map[string]int{}

	var previewSuccess, previewFailed, convertSuccess, convertFailed int
	var backendTotal, backend5xx, frontendTotal int

	for _, e := range events {
		resp.Totals.EventsTotal++
		if e.Source == "frontend" {
			frontendTotal++
		}
		if e.Source == "backend" {
			backendTotal++
			if e.HTTPStatus >= 500 {
				backend5xx++
			}
		}

		if e.Status == "error" {
			errorCounts[e.EventName]++
		}

		switch e.EventName {
		case "studio_opened":
			if e.Status == "success" {
				resp.Funnel.StudioOpened++
				if e.SessionID != "" {
					s := sessionMap[e.SessionID]
					if s.studio.IsZero() || e.EventTime.Before(s.studio) {
						s.studio = e.EventTime
						sessionMap[e.SessionID] = s
					}
				}
			}
		case "input_provided":
			if e.Status == "success" {
				resp.Funnel.InputProvided++
			}
		case "preview_succeeded":
			resp.Funnel.PreviewSucceeded++
			previewSuccess++
		case "preview_failed":
			previewFailed++
		case "convert_succeeded":
			resp.Funnel.ConvertSucceeded++
			convertSuccess++
			if e.SessionID != "" {
				s := sessionMap[e.SessionID]
				if s.convert.IsZero() || e.EventTime.Before(s.convert) {
					s.convert = e.EventTime
					sessionMap[e.SessionID] = s
				}
			}
		case "convert_failed":
			convertFailed++
		case "share_created_ui":
			if e.Status == "success" {
				resp.Funnel.ShareCreated++
			}
		}

		switch e.EventName {
		case "api_preview_completed":
			if e.DurationMS > 0 {
				previewDurations = append(previewDurations, e.DurationMS)
			}
		case "api_convert_completed":
			if e.DurationMS > 0 {
				convertDurations = append(convertDurations, e.DurationMS)
			}
		}
	}

	resp.Totals.FrontendEvents = frontendTotal
	resp.Totals.BackendEvents = backendTotal

	for _, s := range sessionMap {
		if s.studio.IsZero() || s.convert.IsZero() || s.convert.Before(s.studio) {
			continue
		}
		ttv := s.convert.Sub(s.studio).Milliseconds()
		ttvDurations = append(ttvDurations, ttv)
	}

	activated := 0
	for _, s := range sessionMap {
		if s.studio.IsZero() || s.convert.IsZero() || s.convert.Before(s.studio) {
			continue
		}
		if s.convert.Sub(s.studio) <= 10*time.Minute {
			activated++
		}
	}
	if resp.Funnel.StudioOpened > 0 {
		resp.KPIs.ActivationRate10m = ratio(activated, resp.Funnel.StudioOpened)
	}

	resp.KPIs.TimeToValueMS.Median = percentile(ttvDurations, 50)
	resp.KPIs.TimeToValueMS.P75 = percentile(ttvDurations, 75)
	resp.KPIs.TimeToValueMS.P95 = percentile(ttvDurations, 95)

	resp.KPIs.PreviewSuccessRate = ratio(previewSuccess, previewSuccess+previewFailed)
	resp.KPIs.ConvertSuccessRate = ratio(convertSuccess, convertSuccess+convertFailed)

	resp.Reliability.API5xxRate = ratio(backend5xx, backendTotal)
	resp.Reliability.P95PreviewLatencyMS = percentile(previewDurations, 95)
	resp.Reliability.P95ConvertLatencyMS = percentile(convertDurations, 95)

	type errorItem struct {
		name  string
		count int
	}
	items := make([]errorItem, 0, len(errorCounts))
	for name, count := range errorCounts {
		items = append(items, errorItem{name: name, count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].count == items[j].count {
			return items[i].name < items[j].name
		}
		return items[i].count > items[j].count
	})
	for i, item := range items {
		if i >= 10 {
			break
		}
		resp.Errors = append(resp.Errors, struct {
			EventName string `json:"event_name"`
			Count     int    `json:"count"`
		}{EventName: item.name, Count: item.count})
	}

	// AI cost aggregation
	var totalCostUSD float64
	var totalInputTokens, totalOutputTokens int64
	var totalAIRequests int
	costByModel := map[string]struct{ cost float64; count int }{}

	for _, e := range events {
		if e.AIEstimatedCostUSD > 0 {
			totalCostUSD += e.AIEstimatedCostUSD
			totalInputTokens += e.AIInputTokens
			totalOutputTokens += e.AIOutputTokens
			totalAIRequests++
			m := costByModel[e.AIModel]
			m.cost += e.AIEstimatedCostUSD
			m.count++
			costByModel[e.AIModel] = m
		}
	}

	resp.AICost.TotalCostUSD = totalCostUSD
	resp.AICost.TotalInputTokens = totalInputTokens
	resp.AICost.TotalOutputTokens = totalOutputTokens
	resp.AICost.TotalAIRequests = totalAIRequests
	if totalAIRequests > 0 {
		resp.AICost.AvgCostPerConvert = totalCostUSD / float64(totalAIRequests)
	}

	for model, data := range costByModel {
		modelName := model
		if modelName == "" {
			modelName = "unknown"
		}
		resp.AICost.CostByModel = append(resp.AICost.CostByModel, struct {
			Model    string  `json:"model"`
			CostUSD  float64 `json:"cost_usd"`
			Requests int     `json:"requests"`
		}{Model: modelName, CostUSD: data.cost, Requests: data.count})
	}
	sort.Slice(resp.AICost.CostByModel, func(i, j int) bool {
		return resp.AICost.CostByModel[i].CostUSD > resp.AICost.CostByModel[j].CostUSD
	})

	c.JSON(http.StatusOK, resp)
}

func ratio(part, total int) float64 {
	if total <= 0 || part <= 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func percentile(values []int64, p int) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	idx := int(float64(p) / 100.0 * float64(len(sorted)-1))
	return sorted[idx]
}

func maxInt(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func maxInt64(v, min int64) int64 {
	if v < min {
		return min
	}
	return v
}
