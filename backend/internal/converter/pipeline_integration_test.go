package converter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

// ---------------------------------------------------------------------------
// Test 1: Full paste-to-markdown pipeline with mock AI (happy path)
// ---------------------------------------------------------------------------

// TestPipeline_PasteToMarkdownWithAI verifies the end-to-end conversion flow when
// AI is configured and returns high-confidence mappings.
//
// Flow: TSV text → pasteParser → MockAIService.MapColumns → columnMap → MDFlow renderer → markdown
//
// Asserts:
//   - No error returned
//   - Output markdown is non-empty
//   - meta.AIUsed == true (AI contributed to the mapping)
//   - MockAIService.MapColumns called exactly once
func TestPipeline_PasteToMarkdownWithAI(t *testing.T) {
	input := "ID\tTitle\tDescription\n1\tLogin Feature\tUser can log in\n2\tLogout Feature\tUser can log out"

	mock := ai.NewMockAIService()
	mock.MapColumnsFunc = func(_ context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
		// High-confidence mapping: all 3 columns identified
		return &ai.ColumnMappingResult{
			SchemaVersion: ai.SchemaVersionColumnMapping,
			CanonicalFields: []ai.CanonicalFieldMapping{
				{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.92, Reasoning: "ID is a unique identifier"},
				{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 1, Confidence: 0.92, Reasoning: "Title maps to feature title"},
				{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 0.91, Reasoning: "Description is the requirement description"},
			},
			Meta: ai.MappingMeta{
				DetectedType:   "product_backlog",
				SourceLanguage: "en",
				TotalColumns:   3,
				MappedColumns:  3,
				AvgConfidence:  0.917,
			},
		}, nil
	}

	conv := NewConverter().WithAIService(mock)
	resp, err := conv.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")

	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if resp.MDFlow == "" {
		t.Error("expected non-empty markdown output")
	}
	if !resp.Meta.AIUsed {
		t.Errorf("expected meta.AIUsed=true, got false (mode=%q, degraded=%v, fallback_reason=%q)",
			resp.Meta.AIMode, resp.Meta.AIDegraded, resp.Meta.AIFallbackReason)
	}
	if n := mock.CallCountFor("MapColumns"); n != 1 {
		t.Errorf("expected exactly 1 MapColumns call, got %d", n)
	}

	// Spot-check output contains field data
	if !strings.Contains(resp.MDFlow, "Login Feature") && !strings.Contains(resp.MDFlow, "Logout Feature") {
		t.Errorf("output should contain row data; got:\n%s", resp.MDFlow)
	}
}

// ---------------------------------------------------------------------------
// Test 2: Low-confidence AI mapping → degraded path (converter level)
// ---------------------------------------------------------------------------

// TestPipeline_LowConfidenceTriggersRefinement verifies the converter degrades
// gracefully when AI returns mappings below the confidence threshold.
//
// NOTE: The *actual* refinement call chain (ServiceImpl.GetMappingWithFallback →
// ServiceImpl.RefineMapping → client.RefineMapping) requires a live OpenAI client
// and is exercised in the internal/ai package unit tests. This test covers the
// converter's visible behaviour on a low-confidence result: warning + degraded meta.
//
// Asserts:
//   - No error returned
//   - Output still produced (heuristic fallback)
//   - MAPPING_AI_LOW_CONFIDENCE warning present
//   - meta.AIDegraded == true
//   - AI was called once (mock invocation recorded)
func TestPipeline_LowConfidenceTriggersRefinement(t *testing.T) {
	// 5 columns: only 1 mapped with very low confidence
	input := "ID\tTitle\tDescription\tFoo\tBar\n1\tTest\tDesc\tA\tB"

	mock := ai.NewMockAIService()
	mock.MapColumnsFunc = func(_ context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
		// Average confidence 0.4 < aiMinAvgConfidence (0.75)
		// Mapped ratio 1/5 = 0.20 < aiMinMappedRatio (0.60)
		return &ai.ColumnMappingResult{
			SchemaVersion: ai.SchemaVersionColumnMapping,
			CanonicalFields: []ai.CanonicalFieldMapping{
				{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.40},
			},
			Meta: ai.MappingMeta{
				DetectedType:   "generic",
				SourceLanguage: "en",
				TotalColumns:   5,
				MappedColumns:  1,
				AvgConfidence:  0.40,
			},
		}, nil
	}

	conv := NewConverter().WithAIService(mock)
	resp, err := conv.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")

	if err != nil {
		t.Fatalf("pipeline should not fail on low confidence: %v", err)
	}
	if resp.MDFlow == "" {
		t.Error("expected non-empty output even with low confidence (heuristic should take over)")
	}

	// AI was called once
	if n := mock.CallCountFor("MapColumns"); n != 1 {
		t.Errorf("expected 1 MapColumns call, got %d", n)
	}

	// Mapping should be marked degraded
	if !resp.Meta.AIDegraded {
		t.Errorf("expected meta.AIDegraded=true for low-confidence mapping")
	}

	// Warning must be present
	hasWarning := false
	for _, w := range resp.Warnings {
		if w.Code == "MAPPING_AI_LOW_CONFIDENCE" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Errorf("expected MAPPING_AI_LOW_CONFIDENCE warning; got warnings: %+v", resp.Warnings)
	}
}

// ---------------------------------------------------------------------------
// Test 3: AI returns ErrAIUnavailable → heuristic fallback
// ---------------------------------------------------------------------------

// TestPipeline_AIFailureFallsBackToHeuristic verifies the converter recovers
// gracefully when the AI service is unavailable (circuit breaker open, network
// failure, etc.).
//
// Asserts:
//   - No error returned
//   - Output produced via heuristic column mapping
//   - meta.AIUsed == false
//   - AI_UNAVAILABLE warning present
//   - meta.AIFallbackReason == "ai_unavailable"
func TestPipeline_AIFailureFallsBackToHeuristic(t *testing.T) {
	// Use recognisable headers so the heuristic mapper can produce output
	input := "ID\tTitle\tDescription\n1\tOrder Checkout\tUser completes purchase\n2\tCart Update\tUser updates quantity"

	mock := ai.NewMockAIService()
	mock.MapColumnsFunc = func(_ context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
		return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "circuit breaker open"}
	}

	conv := NewConverter().WithAIService(mock)
	resp, err := conv.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")

	if err != nil {
		t.Fatalf("pipeline should not error on AI unavailability: %v", err)
	}
	if resp.MDFlow == "" {
		t.Error("expected heuristic output even when AI is unavailable")
	}

	// AI was called but result not used
	if resp.Meta.AIUsed {
		t.Error("expected meta.AIUsed=false when AI call failed")
	}
	if resp.Meta.AIFallbackReason != "ai_unavailable" {
		t.Errorf("expected AIFallbackReason=ai_unavailable, got %q", resp.Meta.AIFallbackReason)
	}

	hasWarning := false
	for _, w := range resp.Warnings {
		if w.Code == "AI_UNAVAILABLE" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Errorf("expected AI_UNAVAILABLE warning; got: %+v", resp.Warnings)
	}
}

// ---------------------------------------------------------------------------
// Test 4: "spec" format calls AI; "table" format bypasses AI
// ---------------------------------------------------------------------------

// TestPipeline_DifferentFormatsProduceDifferentOutput verifies that output
// format choice affects the AI call path:
//   - "spec" → column mapping AI call is made, spec-style markdown produced
//   - "table" → AI skipped entirely, simple markdown table produced
//
// Asserts:
//   - Both produce non-empty valid markdown
//   - spec format: mock.CallCountFor("MapColumns") >= 1 (AI invoked)
//   - table format: mock.CallCountFor("MapColumns") == 0 (AI not invoked)
func TestPipeline_DifferentFormatsProduceDifferentOutput(t *testing.T) {
	input := "ID\tTitle\tDescription\n1\tFeature A\tDoes something\n2\tFeature B\tDoes something else"

	// --- spec format ---
	specMock := ai.NewMockAIService()
	specMock.MapColumnsFunc = func(_ context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
		return &ai.ColumnMappingResult{
			SchemaVersion: ai.SchemaVersionColumnMapping,
			CanonicalFields: []ai.CanonicalFieldMapping{
				{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.90},
				{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 1, Confidence: 0.90},
				{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 0.90},
			},
			Meta: ai.MappingMeta{TotalColumns: 3, MappedColumns: 3, AvgConfidence: 0.90},
		}, nil
	}
	specConv := NewConverter().WithAIService(specMock)
	specResp, err := specConv.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")
	if err != nil {
		t.Fatalf("spec format error: %v", err)
	}
	if specResp.MDFlow == "" {
		t.Error("spec format: expected non-empty output")
	}
	if n := specMock.CallCountFor("MapColumns"); n == 0 {
		t.Error("spec format: expected AI MapColumns to be called at least once")
	}

	// --- table format ---
	tableMock := ai.NewMockAIService()
	tableMock.MapColumnsFunc = func(_ context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
		t.Error("table format: AI MapColumns should NOT be called")
		return nil, nil
	}
	tableConv := NewConverter().WithAIService(tableMock)
	tableResp, err := tableConv.ConvertPasteWithFormatContext(context.Background(), input, "", "table")
	if err != nil {
		t.Fatalf("table format error: %v", err)
	}
	if tableResp.MDFlow == "" {
		t.Error("table format: expected non-empty output")
	}
	if n := tableMock.CallCountFor("MapColumns"); n != 0 {
		t.Errorf("table format: expected 0 AI calls, got %d", n)
	}

	// Sanity: outputs should differ (spec is structured, table is a markdown table)
	if specResp.MDFlow == tableResp.MDFlow {
		t.Error("spec and table format outputs should differ")
	}
}

// ---------------------------------------------------------------------------
// Test 5: Large input is sanitized before reaching AI
// ---------------------------------------------------------------------------

// TestPipeline_LargeInputSanitized verifies the sanitizer limits the payload
// sent to AI (MaxColumnCount=50 headers, MaxSampleRows=5 sample rows).
//
// Uses 60-column, 200-row input to exceed both limits.
//
// Asserts:
//   - AI receives exactly MaxColumnCount (50) headers (trimmed from 60)
//   - AI receives at most MaxSampleRows (5) sample rows
//   - Output markdown is still produced (heuristic fallback after partial mapping)
func TestPipeline_LargeInputSanitized(t *testing.T) {
	// Build 60-column header row; first 3 are recognisable canonical names
	headers := make([]string, 60)
	headers[0] = "ID"
	headers[1] = "Title"
	headers[2] = "Description"
	for i := 3; i < 60; i++ {
		headers[i] = fmt.Sprintf("col_%d", i)
	}

	// Build 200 data rows
	rowLines := make([]string, 0, 201)
	rowLines = append(rowLines, strings.Join(headers, "\t"))
	for r := 0; r < 200; r++ {
		cells := make([]string, 60)
		for c := range cells {
			cells[c] = fmt.Sprintf("v%d_%d", r, c)
		}
		rowLines = append(rowLines, strings.Join(cells, "\t"))
	}
	input := strings.Join(rowLines, "\n")

	// Capture what the AI actually receives
	var capturedReq ai.MapColumnsRequest
	var capturedOnce sync.Once

	mock := ai.NewMockAIService()
	mock.MapColumnsFunc = func(_ context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
		capturedOnce.Do(func() { capturedReq = req })
		// Return minimal valid mapping for the 3 well-known columns
		return &ai.ColumnMappingResult{
			SchemaVersion: ai.SchemaVersionColumnMapping,
			CanonicalFields: []ai.CanonicalFieldMapping{
				{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.90},
				{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 1, Confidence: 0.90},
				{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 0.90},
			},
			Meta: ai.MappingMeta{
				TotalColumns:  len(req.Headers),
				MappedColumns: 3,
				AvgConfidence: 0.90,
			},
		}, nil
	}

	conv := NewConverter().WithAIService(mock)
	resp, err := conv.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")

	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if resp.MDFlow == "" {
		t.Error("expected non-empty output for large input")
	}
	if mock.CallCountFor("MapColumns") == 0 {
		t.Fatal("AI mock was never called — check test setup")
	}

	// Assert sanitizer applied column limit
	if len(capturedReq.Headers) > MaxColumnCount {
		t.Errorf("AI received %d headers (limit %d): sanitizer should have capped this",
			len(capturedReq.Headers), MaxColumnCount)
	}
	if len(capturedReq.Headers) != MaxColumnCount {
		t.Errorf("expected AI to receive exactly %d headers (max cap), got %d",
			MaxColumnCount, len(capturedReq.Headers))
	}

	// Assert sanitizer applied row limit
	if len(capturedReq.SampleRows) > MaxSampleRows {
		t.Errorf("AI received %d sample rows (limit %d): sanitizer should have capped this",
			len(capturedReq.SampleRows), MaxSampleRows)
	}
}

// ---------------------------------------------------------------------------
// Test 6: BYOK isolation — different API keys get separate service instances
// ---------------------------------------------------------------------------

// TestPipeline_BYOKIsolation verifies that the BYOKServiceCache returns:
//   - Distinct service instances for different API keys (isolation)
//   - The same cached service instance for repeated calls with the same key
//
// This guards against per-user AI state (prompts, caches, cost trackers)
// leaking between BYOK users.
func TestPipeline_BYOKIsolation(t *testing.T) {
	type callEntry struct{ apiKey string }
	var (
		mu      sync.Mutex
		factory []callEntry
	)

	// Factory creates a new MockAIService per unique API key
	serviceFactory := func(apiKey string) (ai.Service, error) {
		mu.Lock()
		factory = append(factory, callEntry{apiKey: apiKey})
		mu.Unlock()

		svc := ai.NewMockAIService()
		svc.Model = "gpt-4o-mini-byok-" + apiKey // unique model tag per key for traceability
		return svc, nil
	}

	cache := ai.NewBYOKServiceCache(ai.BYOKServiceCacheConfig{
		TTL:           5 * time.Minute,
		CleanupTicker: 1 * time.Minute,
		MaxEntries:    100,
	}, serviceFactory)
	defer cache.Close()

	// First call for key-a: creates a new service
	svcA1, err := cache.GetOrCreate("key-user-a")
	if err != nil {
		t.Fatalf("GetOrCreate(key-user-a) error: %v", err)
	}

	// First call for key-b: creates a different service
	svcB1, err := cache.GetOrCreate("key-user-b")
	if err != nil {
		t.Fatalf("GetOrCreate(key-user-b) error: %v", err)
	}

	// Second call for key-a: should return the CACHED instance
	svcA2, err := cache.GetOrCreate("key-user-a")
	if err != nil {
		t.Fatalf("GetOrCreate(key-user-a) second call error: %v", err)
	}

	// Different keys → different service instances (isolation enforced)
	if svcA1 == svcB1 {
		t.Error("expected distinct service instances for different API keys (BYOK isolation)")
	}

	// Same key → same cached instance (no duplicate creation)
	if svcA1 != svcA2 {
		t.Error("expected same service instance for repeated calls with the same API key (cache hit)")
	}

	// Factory was called exactly twice (once per unique key)
	mu.Lock()
	factoryCallCount := len(factory)
	mu.Unlock()
	if factoryCallCount != 2 {
		t.Errorf("expected factory called 2 times (once per unique key), got %d", factoryCallCount)
	}

	// Cache holds 2 entries
	if n := cache.Size(); n != 2 {
		t.Errorf("expected cache size 2, got %d", n)
	}

	// Both services work correctly when plugged into the converter
	input := "ID\tTitle\n1\tTest"
	convA := NewConverter().WithAIService(svcA1)
	respA, err := convA.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")
	if err != nil || respA.MDFlow == "" {
		t.Errorf("converter with BYOK service A failed: err=%v, mdflow=%q", err, respA.MDFlow)
	}

	convB := NewConverter().WithAIService(svcB1)
	respB, err := convB.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")
	if err != nil || respB.MDFlow == "" {
		t.Errorf("converter with BYOK service B failed: err=%v, mdflow=%q", err, respB.MDFlow)
	}
}
