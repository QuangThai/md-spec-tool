package ai

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ABTest represents an A/B test configuration.
type ABTest struct {
	ID          string  // Unique test identifier
	OperationID string  // Prompt operation under test (e.g. "column_mapping")
	VariantA    string  // Version string for the control  (e.g. "v3")
	VariantB    string  // Version string for the experiment (e.g. "v4")
	TrafficPct  float64 // Fraction of traffic routed to B (0.0 – 1.0)
	MinSamples  int     // Minimum samples per variant before the test is considered significant
	Status      string  // "running" | "completed" | "promoted"
}

// ABTestResult is the public, per-variant metric snapshot derived from
// accumulated raw observations.
type ABTestResult struct {
	TestID        string
	Variant       string // "A" or "B"
	Samples       int
	AvgConfidence float64
	AvgLatencyMs  float64
	AvgCost       float64
	ErrorRate     float64
}

// ABTestComparison shows how variant B compares to variant A at a point in time.
type ABTestComparison struct {
	TestID          string
	Status          string
	VariantA        ABTestResult
	VariantB        ABTestResult
	ConfidenceDelta float64 // B − A; positive means B is better
	LatencyDelta    float64 // B − A; negative means B is faster
	CostDelta       float64 // B − A; negative means B is cheaper
	Significant     bool    // true once both variants have ≥ MinSamples
	ShouldPromote   bool    // true when significant AND B beats A on confidence AND cost
}

// variantAccumulator holds running totals for a single variant.
// Averages are computed lazily by toResult.
type variantAccumulator struct {
	samples       int
	sumConfidence float64
	sumLatencyMs  float64
	sumCost       float64
	errorCount    int
}

// toResult converts raw running totals into the public ABTestResult view.
func (a *variantAccumulator) toResult(testID, variant string) ABTestResult {
	res := ABTestResult{
		TestID:  testID,
		Variant: variant,
		Samples: a.samples,
	}
	if a.samples > 0 {
		res.AvgConfidence = a.sumConfidence / float64(a.samples)
		res.AvgLatencyMs = a.sumLatencyMs / float64(a.samples)
		res.AvgCost = a.sumCost / float64(a.samples)
		res.ErrorRate = float64(a.errorCount) / float64(a.samples)
	}
	return res
}

// ABTestManager manages prompt A/B tests.
// All exported methods are safe for concurrent use.
type ABTestManager struct {
	mu       sync.RWMutex
	tests    map[string]*ABTest
	results  map[string]map[string]*variantAccumulator // testID → variant → accumulator
	rng      *rand.Rand
	registry *PromptRegistry // optional; used by PromoteVariant
}

// NewABTestManager creates a ready-to-use ABTestManager.
// Pass a *PromptRegistry if you intend to call PromoteVariant; nil is safe
// for unit tests that don't exercise promotion.
func NewABTestManager(registry *PromptRegistry) *ABTestManager {
	return &ABTestManager{
		tests:    make(map[string]*ABTest),
		results:  make(map[string]map[string]*variantAccumulator),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		registry: registry,
	}
}

// CreateTest registers a new A/B test.
// Returns an error if a test with the same ID already exists or if
// TrafficPct is out of the [0, 1] range.
func (m *ABTestManager) CreateTest(test ABTest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tests[test.ID]; exists {
		return fmt.Errorf("ab_test: test %q already exists", test.ID)
	}
	if test.TrafficPct < 0 || test.TrafficPct > 1 {
		return fmt.Errorf("ab_test: TrafficPct must be in [0, 1], got %.4f", test.TrafficPct)
	}
	if test.Status == "" {
		test.Status = "running"
	}

	copied := test
	m.tests[test.ID] = &copied
	m.results[test.ID] = map[string]*variantAccumulator{
		"A": {},
		"B": {},
	}
	return nil
}

// SelectVariant picks a variant for the given operationID.
//
// It scans all tests in "running" status for the requested operation and
// returns (testID, variant, version).  variant is either "A" or "B"; version
// is the corresponding VersionA or VersionB string.
//
// Returns ("", "", "") when no running test exists for the operation.
func (m *ABTestManager) SelectVariant(operationID string) (testID, variant, version string) {
	// Write-lock because rand.Float64 mutates the RNG state.
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, test := range m.tests {
		if test.OperationID != operationID || test.Status != "running" {
			continue
		}
		if m.rng.Float64() < test.TrafficPct {
			return test.ID, "B", test.VariantB
		}
		return test.ID, "A", test.VariantA
	}
	return "", "", ""
}

// RecordResult appends a single observation for the given test and variant.
// Unknown testIDs or variant values are silently ignored to keep hot-path
// callers error-free.
func (m *ABTestManager) RecordResult(testID, variant string, confidence, latencyMs, cost float64, hasError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	variantMap, ok := m.results[testID]
	if !ok {
		return
	}
	acc, ok := variantMap[variant]
	if !ok {
		acc = &variantAccumulator{}
		variantMap[variant] = acc
	}

	acc.samples++
	acc.sumConfidence += confidence
	acc.sumLatencyMs += latencyMs
	acc.sumCost += cost
	if hasError {
		acc.errorCount++
	}
}

// GetComparison returns a snapshot comparison between variants A and B for
// the given test.  Returns an error if the test does not exist.
func (m *ABTestManager) GetComparison(testID string) (*ABTestComparison, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	test, ok := m.tests[testID]
	if !ok {
		return nil, fmt.Errorf("ab_test: test %q not found", testID)
	}

	variantMap := m.results[testID]
	accA := variantMap["A"]
	accB := variantMap["B"]
	if accA == nil {
		accA = &variantAccumulator{}
	}
	if accB == nil {
		accB = &variantAccumulator{}
	}

	resA := accA.toResult(testID, "A")
	resB := accB.toResult(testID, "B")

	significant := accA.samples >= test.MinSamples && accB.samples >= test.MinSamples

	var shouldPromote bool
	if significant {
		// B must beat A on confidence (higher is better) AND cost (lower is better).
		shouldPromote = resB.AvgConfidence > resA.AvgConfidence && resB.AvgCost < resA.AvgCost
	}

	return &ABTestComparison{
		TestID:          testID,
		Status:          test.Status,
		VariantA:        resA,
		VariantB:        resB,
		ConfidenceDelta: resB.AvgConfidence - resA.AvgConfidence,
		LatencyDelta:    resB.AvgLatencyMs - resA.AvgLatencyMs,
		CostDelta:       resB.AvgCost - resA.AvgCost,
		Significant:     significant,
		ShouldPromote:   shouldPromote,
	}, nil
}

// PromoteVariant marks the test as "promoted" and, when a PromptRegistry was
// provided at construction time, sets a version override so that the registry
// serves variant B's version for the operation going forward.
func (m *ABTestManager) PromoteVariant(testID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	test, ok := m.tests[testID]
	if !ok {
		return fmt.Errorf("ab_test: test %q not found", testID)
	}

	test.Status = "promoted"

	if m.registry != nil {
		// SetVersionOverride acquires its own lock; releasing m.mu first would
		// be cleaner but risks a TOCTOU race on test.Status.  Since registry.mu
		// and m.mu are always acquired in this fixed order (m.mu → registry.mu)
		// and never in the reverse order, there is no deadlock.
		m.registry.SetVersionOverride(test.OperationID, test.VariantB)
	}

	return nil
}

// ListTests returns a snapshot of all registered tests (in arbitrary order).
func (m *ABTestManager) ListTests() []ABTest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]ABTest, 0, len(m.tests))
	for _, t := range m.tests {
		out = append(out, *t)
	}
	return out
}
