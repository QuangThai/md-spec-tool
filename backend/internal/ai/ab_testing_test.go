package ai

import (
	"math"
	"math/rand"
	"sync"
	"testing"
)

// ---- helpers ----

func newSeededABManager(seed int64) *ABTestManager {
	mgr := NewABTestManager(nil)
	mgr.rng = rand.New(rand.NewSource(seed))
	return mgr
}

func mustCreateABTest(t *testing.T, mgr *ABTestManager, test ABTest) {
	t.Helper()
	if err := mgr.CreateTest(test); err != nil {
		t.Fatalf("CreateTest: %v", err)
	}
}

// recordN records n identical results for the given variant.
func recordN(mgr *ABTestManager, testID, variant string, n int, confidence, latencyMs, cost float64, hasError bool) {
	for i := 0; i < n; i++ {
		mgr.RecordResult(testID, variant, confidence, latencyMs, cost, hasError)
	}
}

// ---- test cases ----

func TestABTest_CreateAndList(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "t1",
		OperationID: "column_mapping",
		VariantA:    "v3",
		VariantB:    "v4",
		TrafficPct:  0.3,
		MinSamples:  100,
		Status:      "running",
	})

	tests := mgr.ListTests()
	if len(tests) != 1 {
		t.Fatalf("expected 1 test, got %d", len(tests))
	}
	got := tests[0]
	if got.ID != "t1" {
		t.Errorf("ID: want t1, got %s", got.ID)
	}
	if got.OperationID != "column_mapping" {
		t.Errorf("OperationID: want column_mapping, got %s", got.OperationID)
	}
	if got.VariantA != "v3" || got.VariantB != "v4" {
		t.Errorf("variants: want A=v3 B=v4, got A=%s B=%s", got.VariantA, got.VariantB)
	}
	if got.Status != "running" {
		t.Errorf("Status: want running, got %s", got.Status)
	}
}

func TestABTest_CreateTest_DuplicateIDReturnsError(t *testing.T) {
	mgr := newSeededABManager(42)
	test := ABTest{ID: "dup", OperationID: "op", VariantA: "v1", VariantB: "v2", TrafficPct: 0.5, MinSamples: 10, Status: "running"}
	if err := mgr.CreateTest(test); err != nil {
		t.Fatalf("first CreateTest: %v", err)
	}
	if err := mgr.CreateTest(test); err == nil {
		t.Fatal("expected error for duplicate test ID, got nil")
	}
}

func TestABTest_SelectVariant_RespectTrafficSplit(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "split-test",
		OperationID: "column_mapping",
		VariantA:    "v3",
		VariantB:    "v4",
		TrafficPct:  0.30,
		MinSamples:  100,
		Status:      "running",
	})

	const trials = 1000
	bCount := 0
	for i := 0; i < trials; i++ {
		_, variant, _ := mgr.SelectVariant("column_mapping")
		if variant == "B" {
			bCount++
		}
	}

	bPct := float64(bCount) / trials
	// Allow ±5% tolerance around 30%
	if bPct < 0.25 || bPct > 0.35 {
		t.Errorf("expected ~30%% B traffic, got %.1f%% (%d/1000)", bPct*100, bCount)
	}
}

func TestABTest_SelectVariant_ReturnsCorrectVersionStrings(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "ver-test",
		OperationID: "paste_analysis",
		VariantA:    "v2",
		VariantB:    "v3",
		TrafficPct:  0.5,
		MinSamples:  10,
		Status:      "running",
	})

	aVersionSeen, bVersionSeen := false, false
	for i := 0; i < 200; i++ {
		testID, variant, version := mgr.SelectVariant("paste_analysis")
		if testID != "ver-test" {
			t.Errorf("unexpected testID: %s", testID)
		}
		switch variant {
		case "A":
			if version != "v2" {
				t.Errorf("A version: want v2, got %s", version)
			}
			aVersionSeen = true
		case "B":
			if version != "v3" {
				t.Errorf("B version: want v3, got %s", version)
			}
			bVersionSeen = true
		}
	}
	if !aVersionSeen || !bVersionSeen {
		t.Error("expected both variants to be selected in 200 trials")
	}
}

func TestABTest_SelectVariant_NoActiveTest(t *testing.T) {
	mgr := newSeededABManager(42)

	testID, variant, version := mgr.SelectVariant("column_mapping")
	if testID != "" || variant != "" || version != "" {
		t.Errorf("expected empty returns for unknown operation, got (%q, %q, %q)", testID, variant, version)
	}
}

func TestABTest_SelectVariant_SkipsNonRunningTests(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "completed-test",
		OperationID: "column_mapping",
		VariantA:    "v3",
		VariantB:    "v4",
		TrafficPct:  0.5,
		MinSamples:  10,
		Status:      "completed",
	})

	testID, variant, version := mgr.SelectVariant("column_mapping")
	if testID != "" || variant != "" || version != "" {
		t.Errorf("should not select from completed test, got (%q, %q, %q)", testID, variant, version)
	}
}

func TestABTest_RecordResult_TracksPerVariant(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "track-test",
		OperationID: "column_mapping",
		VariantA:    "v3",
		VariantB:    "v4",
		TrafficPct:  0.5,
		MinSamples:  10,
		Status:      "running",
	})

	mgr.RecordResult("track-test", "A", 0.8, 100.0, 0.001, false)
	mgr.RecordResult("track-test", "A", 0.9, 120.0, 0.002, false)
	mgr.RecordResult("track-test", "B", 0.95, 90.0, 0.0008, false)

	cmp, err := mgr.GetComparison("track-test")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}

	if cmp.VariantA.Samples != 2 {
		t.Errorf("A samples: want 2, got %d", cmp.VariantA.Samples)
	}
	if cmp.VariantB.Samples != 1 {
		t.Errorf("B samples: want 1, got %d", cmp.VariantB.Samples)
	}

	wantAConf := (0.8 + 0.9) / 2.0
	if math.Abs(cmp.VariantA.AvgConfidence-wantAConf) > 1e-9 {
		t.Errorf("A AvgConfidence: want %.4f, got %.4f", wantAConf, cmp.VariantA.AvgConfidence)
	}
	wantALatency := (100.0 + 120.0) / 2.0
	if math.Abs(cmp.VariantA.AvgLatencyMs-wantALatency) > 1e-9 {
		t.Errorf("A AvgLatencyMs: want %.2f, got %.2f", wantALatency, cmp.VariantA.AvgLatencyMs)
	}
}

func TestABTest_RecordResult_TracksErrorRate(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "err-test",
		OperationID: "op",
		VariantA:    "v1",
		VariantB:    "v2",
		TrafficPct:  0.5,
		MinSamples:  10,
		Status:      "running",
	})

	// Record 4 calls for A: 1 error
	mgr.RecordResult("err-test", "A", 0.8, 100, 0.001, false)
	mgr.RecordResult("err-test", "A", 0.8, 100, 0.001, false)
	mgr.RecordResult("err-test", "A", 0.8, 100, 0.001, false)
	mgr.RecordResult("err-test", "A", 0.0, 0, 0, true)

	cmp, err := mgr.GetComparison("err-test")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}
	wantErrRate := 0.25 // 1/4
	if math.Abs(cmp.VariantA.ErrorRate-wantErrRate) > 1e-9 {
		t.Errorf("A ErrorRate: want %.2f, got %.2f", wantErrRate, cmp.VariantA.ErrorRate)
	}
}

func TestABTest_Comparison_CalculatesDeltas(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "delta-test",
		OperationID: "op",
		VariantA:    "v1",
		VariantB:    "v2",
		TrafficPct:  0.5,
		MinSamples:  1,
		Status:      "running",
	})

	// A: confidence=0.80, latency=120ms, cost=0.002
	mgr.RecordResult("delta-test", "A", 0.80, 120.0, 0.002, false)
	// B: confidence=0.90, latency=100ms, cost=0.001
	mgr.RecordResult("delta-test", "B", 0.90, 100.0, 0.001, false)

	cmp, err := mgr.GetComparison("delta-test")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}

	// ConfidenceDelta = B - A = +0.10
	if math.Abs(cmp.ConfidenceDelta-0.10) > 1e-9 {
		t.Errorf("ConfidenceDelta: want 0.10, got %.4f", cmp.ConfidenceDelta)
	}
	// LatencyDelta = B - A = -20ms (B is faster)
	if math.Abs(cmp.LatencyDelta-(-20.0)) > 1e-9 {
		t.Errorf("LatencyDelta: want -20.0, got %.4f", cmp.LatencyDelta)
	}
	// CostDelta = B - A = -0.001 (B is cheaper)
	if math.Abs(cmp.CostDelta-(-0.001)) > 1e-9 {
		t.Errorf("CostDelta: want -0.001, got %.6f", cmp.CostDelta)
	}
}

func TestABTest_Comparison_NotSignificantBelowMinSamples(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "sig-test",
		OperationID: "op",
		VariantA:    "v1",
		VariantB:    "v2",
		TrafficPct:  0.5,
		MinSamples:  100,
		Status:      "running",
	})

	// Only 5 samples per variant — well below MinSamples=100
	recordN(mgr, "sig-test", "A", 5, 0.8, 100, 0.001, false)
	recordN(mgr, "sig-test", "B", 5, 0.9, 90, 0.0009, false)

	cmp, err := mgr.GetComparison("sig-test")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}
	if cmp.Significant {
		t.Error("expected Significant=false with only 5 samples (MinSamples=100)")
	}
	if cmp.ShouldPromote {
		t.Error("ShouldPromote must be false when not significant")
	}
}

func TestABTest_Comparison_SignificantAboveMinSamples(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "sig2-test",
		OperationID: "op",
		VariantA:    "v1",
		VariantB:    "v2",
		TrafficPct:  0.5,
		MinSamples:  100,
		Status:      "running",
	})

	recordN(mgr, "sig2-test", "A", 100, 0.8, 100, 0.001, false)
	recordN(mgr, "sig2-test", "B", 100, 0.9, 90, 0.0009, false)

	cmp, err := mgr.GetComparison("sig2-test")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}
	if !cmp.Significant {
		t.Error("expected Significant=true with 100 samples each (MinSamples=100)")
	}
}

func TestABTest_ShouldPromote_WhenBBetter(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "promo-good",
		OperationID: "op",
		VariantA:    "v1",
		VariantB:    "v2",
		TrafficPct:  0.5,
		MinSamples:  10,
		Status:      "running",
	})

	// B has higher confidence AND lower cost → should promote
	recordN(mgr, "promo-good", "A", 10, 0.80, 120.0, 0.002, false)
	recordN(mgr, "promo-good", "B", 10, 0.92, 100.0, 0.0015, false)

	cmp, err := mgr.GetComparison("promo-good")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}
	if !cmp.Significant {
		t.Fatal("test must be significant to evaluate ShouldPromote")
	}
	if !cmp.ShouldPromote {
		t.Errorf("expected ShouldPromote=true (B confidence=0.92 > A=0.80, B cost=0.0015 < A=0.002)")
	}
}

func TestABTest_ShouldNotPromote_WhenBWorse(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "promo-bad",
		OperationID: "op",
		VariantA:    "v1",
		VariantB:    "v2",
		TrafficPct:  0.5,
		MinSamples:  10,
		Status:      "running",
	})

	// B has LOWER confidence — should not promote even if cost is lower
	recordN(mgr, "promo-bad", "A", 10, 0.90, 120.0, 0.002, false)
	recordN(mgr, "promo-bad", "B", 10, 0.75, 100.0, 0.001, false)

	cmp, err := mgr.GetComparison("promo-bad")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}
	if cmp.ShouldPromote {
		t.Error("expected ShouldPromote=false when B confidence < A confidence")
	}
}

func TestABTest_ShouldNotPromote_WhenBHigherCost(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "promo-cost",
		OperationID: "op",
		VariantA:    "v1",
		VariantB:    "v2",
		TrafficPct:  0.5,
		MinSamples:  10,
		Status:      "running",
	})

	// B has higher confidence BUT higher cost → should not promote
	recordN(mgr, "promo-cost", "A", 10, 0.80, 120.0, 0.001, false)
	recordN(mgr, "promo-cost", "B", 10, 0.92, 100.0, 0.003, false)

	cmp, err := mgr.GetComparison("promo-cost")
	if err != nil {
		t.Fatalf("GetComparison: %v", err)
	}
	if cmp.ShouldPromote {
		t.Error("expected ShouldPromote=false when B cost > A cost")
	}
}

func TestABTest_PromoteVariant(t *testing.T) {
	reg := NewPromptRegistry()
	reg.Register(PromptEntry{ID: "column_mapping", Version: "v3", Content: "prompt v3"})
	reg.Register(PromptEntry{ID: "column_mapping", Version: "v4", Content: "prompt v4"})

	mgr := NewABTestManager(reg)
	mgr.rng = rand.New(rand.NewSource(42))
	mustCreateABTest(t, mgr, ABTest{
		ID:          "promote-test",
		OperationID: "column_mapping",
		VariantA:    "v3",
		VariantB:    "v4",
		TrafficPct:  0.5,
		MinSamples:  5,
		Status:      "running",
	})

	if err := mgr.PromoteVariant("promote-test"); err != nil {
		t.Fatalf("PromoteVariant: %v", err)
	}

	// Status should be "promoted"
	tests := mgr.ListTests()
	if len(tests) != 1 || tests[0].Status != "promoted" {
		t.Errorf("expected status=promoted, got %s", tests[0].Status)
	}

	// Registry should now serve v4 (variant B version)
	entry, ok := reg.Get("column_mapping")
	if !ok {
		t.Fatal("expected to find column_mapping in registry")
	}
	if entry.Version != "v4" {
		t.Errorf("expected registry to serve v4 after promote, got %s", entry.Version)
	}
}

func TestABTest_PromoteVariant_UnknownTestReturnsError(t *testing.T) {
	mgr := newSeededABManager(42)
	if err := mgr.PromoteVariant("nonexistent"); err == nil {
		t.Fatal("expected error for unknown test ID")
	}
}

func TestABTest_GetComparison_UnknownTestReturnsError(t *testing.T) {
	mgr := newSeededABManager(42)
	if _, err := mgr.GetComparison("nonexistent"); err == nil {
		t.Fatal("expected error for unknown test ID")
	}
}

func TestABTest_PromoteVariant_SelectVariantSkipsPromoted(t *testing.T) {
	mgr := newSeededABManager(42)
	mustCreateABTest(t, mgr, ABTest{
		ID:          "skip-promoted",
		OperationID: "column_mapping",
		VariantA:    "v3",
		VariantB:    "v4",
		TrafficPct:  0.5,
		MinSamples:  5,
		Status:      "running",
	})

	_ = mgr.PromoteVariant("skip-promoted")

	// After promotion, SelectVariant should return empty
	testID, variant, version := mgr.SelectVariant("column_mapping")
	if testID != "" || variant != "" || version != "" {
		t.Errorf("should not select promoted test, got (%q, %q, %q)", testID, variant, version)
	}
}

func TestABTest_ConcurrentSafety(t *testing.T) {
	mgr := NewABTestManager(nil)
	// Don't seed — use default random for concurrency test

	mustCreateABTest(t, mgr, ABTest{
		ID:          "concurrent",
		OperationID: "column_mapping",
		VariantA:    "v3",
		VariantB:    "v4",
		TrafficPct:  0.4,
		MinSamples:  50,
		Status:      "running",
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				testID, variant, _ := mgr.SelectVariant("column_mapping")
				if testID != "" && variant != "" {
					mgr.RecordResult(testID, variant, 0.85, 100.0, 0.001, false)
				}
				_, _ = mgr.GetComparison("concurrent")
				_ = mgr.ListTests()
			}
		}()
	}
	wg.Wait()

	// Just verify we can still read a valid comparison
	cmp, err := mgr.GetComparison("concurrent")
	if err != nil {
		t.Fatalf("GetComparison after concurrent ops: %v", err)
	}
	if cmp.VariantA.Samples+cmp.VariantB.Samples == 0 {
		t.Error("expected some samples recorded")
	}
}
