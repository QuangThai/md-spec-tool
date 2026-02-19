package middleware

import (
	"sync"
	"time"
)

// MaxTelemetryEvents is the maximum number of events kept in the rolling buffer.
// Override via SetMaxTelemetryEvents() before server startup.
var maxTelemetryEvents = 10000

// SetMaxTelemetryEvents allows configuring the buffer size at startup.
func SetMaxTelemetryEvents(n int) {
	if n > 0 {
		maxTelemetryEvents = n
	}
}

// TelemetryEvent is a normalized event payload used by the MVP dashboard.
type TelemetryEvent struct {
	EventName       string    `json:"event_name"`
	EventTime       time.Time `json:"event_time"`
	SessionID       string    `json:"session_id,omitempty"`
	Status          string    `json:"status"`
	InputSource     string    `json:"input_source,omitempty"`
	TemplateType    string    `json:"template_type,omitempty"`
	DurationMS      int64     `json:"duration_ms,omitempty"`
	ErrorCode       string    `json:"error_code,omitempty"`
	WarningCount    int       `json:"warning_count,omitempty"`
	ConfidenceScore float64   `json:"confidence_score,omitempty"`
	NeedsReview     bool      `json:"needs_review,omitempty"`
	TotalRows            int       `json:"total_rows,omitempty"`
	AIModel              string    `json:"ai_model,omitempty"`
	AIEstimatedCostUSD   float64   `json:"ai_estimated_cost_usd,omitempty"`
	AIInputTokens        int64     `json:"ai_input_tokens,omitempty"`
	AIOutputTokens       int64     `json:"ai_output_tokens,omitempty"`
	HTTPStatus           int       `json:"http_status,omitempty"`
	Path            string    `json:"path,omitempty"`
	RequestID       string    `json:"request_id,omitempty"`
	Source          string    `json:"source"` // frontend | backend
}

type telemetryStore struct {
	mu     sync.RWMutex
	events []TelemetryEvent
}

var defaultTelemetryStore = &telemetryStore{
	events: make([]TelemetryEvent, 0, maxTelemetryEvents),
}

// RecordTelemetryEvent appends an event to the in-memory rolling store.
func RecordTelemetryEvent(event TelemetryEvent) {
	defaultTelemetryStore.record(event)
}

// SnapshotTelemetryEvents returns events after `since`.
func SnapshotTelemetryEvents(since time.Time) []TelemetryEvent {
	return defaultTelemetryStore.snapshot(since)
}

func (s *telemetryStore) record(event TelemetryEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.EventTime.IsZero() {
		event.EventTime = time.Now().UTC()
	}

	if len(s.events) >= maxTelemetryEvents {
		copy(s.events, s.events[1:])
		s.events[len(s.events)-1] = event
		return
	}

	s.events = append(s.events, event)
}

func (s *telemetryStore) snapshot(since time.Time) []TelemetryEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.events) == 0 {
		return nil
	}

	result := make([]TelemetryEvent, 0, len(s.events))
	for _, e := range s.events {
		if !since.IsZero() && e.EventTime.Before(since) {
			continue
		}
		result = append(result, e)
	}
	return result
}
