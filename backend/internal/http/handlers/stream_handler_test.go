package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// ─── SSE parsing helpers ─────────────────────────────────────────────────────

// sseEvent is a parsed Server-Sent Event.
type sseEvent struct {
	Type string
	Data string
}

// parseSSEResponse parses a raw SSE response body into a slice of sseEvents.
// It handles the standard "event: …\ndata: …\n\n" format.
func parseSSEResponse(body string) []sseEvent {
	var events []sseEvent
	var cur sseEvent

	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			cur.Type = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			cur.Data = strings.TrimPrefix(line, "data: ")
		case line == "":
			if cur.Type != "" {
				events = append(events, cur)
				cur = sseEvent{}
			}
		}
	}
	// Flush any trailing event without a final blank line
	if cur.Type != "" {
		events = append(events, cur)
	}
	return events
}

// ─── Test setup ───────────────────────────────────────────────────────────────

func setupStreamHandler() *StreamHandler {
	cfg := config.LoadConfig()
	cfg.MaxPasteBytes = 1 << 20 // 1 MB
	conv := converter.NewConverter()
	provider := NewAIServiceProvider(cfg)
	return NewStreamHandler(conv, cfg, provider)
}

func makeStreamRequest(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mdflow/convert/stream",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

// ─── SSE header tests ─────────────────────────────────────────────────────────

func TestStreamHandler_SSEHeaders(t *testing.T) {
	h := setupStreamHandler()
	c, w := makeStreamRequest(`{"paste_text":"Feature\tScenario\nLogin\tHappy path"}`)

	h.ConvertStream(c)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", w.Header().Get("Cache-Control"))
	}
	if w.Header().Get("Connection") != "keep-alive" {
		t.Errorf("Connection = %q, want keep-alive", w.Header().Get("Connection"))
	}
	if w.Header().Get("X-Accel-Buffering") != "no" {
		t.Errorf("X-Accel-Buffering = %q, want no", w.Header().Get("X-Accel-Buffering"))
	}
}

// ─── Validation error tests ───────────────────────────────────────────────────

func TestStreamHandler_MissingPasteText_ReturnsErrorEvent(t *testing.T) {
	h := setupStreamHandler()
	c, w := makeStreamRequest(`{}`)

	h.ConvertStream(c)

	body := w.Body.String()
	events := parseSSEResponse(body)

	if len(events) == 0 {
		t.Fatal("expected at least one SSE event")
	}
	last := events[len(events)-1]
	if last.Type != "error" {
		t.Errorf("last event type = %q, want error", last.Type)
	}
	// When paste_text binding fails, the handler returns "invalid request body"
	// (Gin reports a required-field violation as a binding error, not a field error).
	if !strings.Contains(last.Data, "invalid") && !strings.Contains(last.Data, "paste_text") {
		t.Errorf("error data %q should contain 'invalid' or 'paste_text'", last.Data)
	}
}

func TestStreamHandler_InvalidJSON_ReturnsErrorEvent(t *testing.T) {
	h := setupStreamHandler()
	c, w := makeStreamRequest(`not-json`)

	h.ConvertStream(c)

	body := w.Body.String()
	events := parseSSEResponse(body)

	if len(events) == 0 {
		t.Fatal("expected at least one SSE event")
	}
	found := false
	for _, e := range events {
		if e.Type == "error" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected an error event, got events: %v", events)
	}
}

// ─── Successful conversion tests ──────────────────────────────────────────────

func TestStreamHandler_SuccessfulConversion_EventOrder(t *testing.T) {
	h := setupStreamHandler()
	body := `{"paste_text":"Feature\tScenario\tExpected\nLogin\tHappy path\tUser redirected","format":"spec"}`
	c, w := makeStreamRequest(body)

	h.ConvertStream(c)

	events := parseSSEResponse(w.Body.String())
	if len(events) == 0 {
		t.Fatal("expected SSE events, got none")
	}

	// Collect ordered event types
	var types []string
	for _, e := range events {
		types = append(types, e.Type)
	}

	// Must contain at least: progress, complete, result
	hasProgress := false
	hasComplete := false
	hasResult := false
	for _, et := range types {
		switch et {
		case "progress":
			hasProgress = true
		case "complete":
			hasComplete = true
		case "result":
			hasResult = true
		}
	}

	if !hasProgress {
		t.Error("expected at least one progress event")
	}
	if !hasComplete {
		t.Error("expected a complete event")
	}
	if !hasResult {
		t.Error("expected a result event with the final conversion")
	}

	// result must be last
	if len(types) == 0 || types[len(types)-1] != "result" {
		t.Errorf("last event = %q, want result; full sequence: %v", types[len(types)-1], types)
	}
}

func TestStreamHandler_SuccessfulConversion_PhaseSequence(t *testing.T) {
	h := setupStreamHandler()
	body := `{"paste_text":"Feature\tScenario\tExpected\nLogin\tHappy path\tUser redirected","format":"spec"}`
	c, w := makeStreamRequest(body)

	h.ConvertStream(c)

	events := parseSSEResponse(w.Body.String())
	var phases []string
	for _, e := range events {
		if e.Type != "progress" && e.Type != "complete" {
			continue
		}
		var pd map[string]interface{}
		if err := json.Unmarshal([]byte(e.Data), &pd); err != nil {
			t.Errorf("failed to parse event data %q: %v", e.Data, err)
			continue
		}
		if phase, ok := pd["phase"].(string); ok {
			phases = append(phases, phase)
		}
	}

	want := []string{"parsing", "mapping", "rendering", "complete"}
	if len(phases) != len(want) {
		t.Fatalf("phases = %v, want %v", phases, want)
	}
	for i, p := range want {
		if phases[i] != p {
			t.Errorf("phases[%d] = %q, want %q", i, phases[i], p)
		}
	}
}

func TestStreamHandler_SuccessfulConversion_ResultContainsMDFlow(t *testing.T) {
	h := setupStreamHandler()
	body := `{"paste_text":"Feature\tScenario\tExpected\nLogin\tHappy path\tUser redirected","format":"spec"}`
	c, w := makeStreamRequest(body)

	h.ConvertStream(c)

	events := parseSSEResponse(w.Body.String())
	for _, e := range events {
		if e.Type != "result" {
			continue
		}
		var resp map[string]interface{}
		if err := json.Unmarshal([]byte(e.Data), &resp); err != nil {
			t.Fatalf("failed to parse result data: %v", err)
		}
		mdflow, ok := resp["mdflow"].(string)
		if !ok || mdflow == "" {
			t.Error("result event mdflow field is empty or missing")
		}
		return
	}
	t.Error("no result event found")
}

// TestStreamHandler_SSEFormat verifies the raw SSE wire format.
func TestStreamHandler_SSEFormat(t *testing.T) {
	h := setupStreamHandler()
	body := `{"paste_text":"Feature\tScenario\nLogin\tHappy path","format":"spec"}`
	c, w := makeStreamRequest(body)

	h.ConvertStream(c)

	raw := w.Body.String()

	// Every event block must have both "event: X" and "data: Y" lines
	if !strings.Contains(raw, "event: ") {
		t.Error("response does not contain any 'event: ' lines")
	}
	if !strings.Contains(raw, "data: ") {
		t.Error("response does not contain any 'data: ' lines")
	}
	// Blocks must be separated by blank lines
	if !strings.Contains(raw, "\n\n") {
		t.Error("SSE blocks are not separated by blank lines")
	}
}

// TestStreamHandler_InvalidFormat_ReturnsErrorEvent ensures an unsupported
// format returns an error SSE event rather than a 500.
func TestStreamHandler_InvalidFormat_ReturnsErrorEvent(t *testing.T) {
	h := setupStreamHandler()
	body := `{"paste_text":"Feature\tScenario\nLogin\tHappy path","format":"xml"}`
	c, w := makeStreamRequest(body)

	h.ConvertStream(c)

	events := parseSSEResponse(w.Body.String())
	found := false
	for _, e := range events {
		if e.Type == "error" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error event for invalid format, got events: %v", events)
	}
}

// TestNewStreamHandler_Defaults verifies that nil parameters are handled.
func TestNewStreamHandler_Defaults(t *testing.T) {
	h := NewStreamHandler(nil, nil, nil)
	if h == nil {
		t.Fatal("NewStreamHandler returned nil")
	}
	if h.converter == nil {
		t.Error("converter should not be nil")
	}
	if h.cfg == nil {
		t.Error("cfg should not be nil")
	}
	if h.byokCache == nil {
		t.Error("byokCache should not be nil")
	}
}
