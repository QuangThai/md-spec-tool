package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourorg/md-spec-tool/internal/ai"
)

type MappingEvalCase struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Headers            []string          `json:"headers"`
	SampleRows         [][]string        `json:"sample_rows"`
	Format             string            `json:"format"`
	FileType           string            `json:"file_type"`
	SourceLang         string            `json:"source_lang"`
	SchemaHint         string            `json:"schema_hint"`
	Expected           map[string]int    `json:"expected"`
	RequiredFields     []string          `json:"required_fields"`
	ForbiddenFields    []string          `json:"forbidden_fields"`
	MinRequiredRecall  float64           `json:"min_required_recall"`
	MinExactMatchRatio float64           `json:"min_exact_match_ratio"`
	MinAvgConfidence   float64           `json:"min_avg_confidence"`
	MaxUnmappedColumns int               `json:"max_unmapped_columns"`
	Tags               map[string]string `json:"tags"`
}

type SuggestionEvalCase struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Template            string            `json:"template"`
	RowCount            int               `json:"row_count"`
	SpecContent         string            `json:"spec_content"`
	ExpectedTypes       []string          `json:"expected_types"`
	MinSuggestions      int               `json:"min_suggestions"`
	MinExpectedTypeHits int               `json:"min_expected_type_hits"`
	AllowedUnknownTypes bool              `json:"allowed_unknown_types"`
	Tags                map[string]string `json:"tags"`
}

type MappingEvalResult struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Passed               bool              `json:"passed"`
	Error                string            `json:"error,omitempty"`
	AvgConfidence        float64           `json:"avg_confidence,omitempty"`
	MappedColumns        int               `json:"mapped_columns,omitempty"`
	UnmappedColumns      int               `json:"unmapped_columns,omitempty"`
	RequiredHitCount     int               `json:"required_hit_count,omitempty"`
	RequiredTotal        int               `json:"required_total,omitempty"`
	RequiredRecall       float64           `json:"required_recall,omitempty"`
	ExactMatchCount      int               `json:"exact_match_count,omitempty"`
	ExpectedTotal        int               `json:"expected_total,omitempty"`
	ExactMatchRatio      float64           `json:"exact_match_ratio,omitempty"`
	ForbiddenHitFields   []string          `json:"forbidden_hit_fields,omitempty"`
	FailedChecks         []string          `json:"failed_checks,omitempty"`
	DetectedType         string            `json:"detected_type,omitempty"`
	SourceLanguage       string            `json:"source_language,omitempty"`
	ActualMappedFields   map[string]int    `json:"actual_mapped_fields,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	EvaluatedAtUnixMilli int64             `json:"evaluated_at_unix_milli"`
}

type SuggestionEvalResult struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Passed               bool              `json:"passed"`
	Error                string            `json:"error,omitempty"`
	SuggestionsCount     int               `json:"suggestions_count,omitempty"`
	ExpectedTypeHits     int               `json:"expected_type_hits,omitempty"`
	ExpectedTypeTotal    int               `json:"expected_type_total,omitempty"`
	PresentTypes         []string          `json:"present_types,omitempty"`
	UnknownTypes         []string          `json:"unknown_types,omitempty"`
	FailedChecks         []string          `json:"failed_checks,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	EvaluatedAtUnixMilli int64             `json:"evaluated_at_unix_milli"`
}

type EvalSummary struct {
	TotalCases         int     `json:"total_cases"`
	PassedCases        int     `json:"passed_cases"`
	FailedCases        int     `json:"failed_cases"`
	PassRate           float64 `json:"pass_rate"`
	MappingCases       int     `json:"mapping_cases"`
	MappingPassed      int     `json:"mapping_passed"`
	SuggestionCases    int     `json:"suggestion_cases"`
	SuggestionPassed   int     `json:"suggestion_passed"`
	AverageMapScore    float64 `json:"average_map_score"`
	AverageMapRecall   float64 `json:"average_map_recall"`
	AverageSuggestions float64 `json:"average_suggestions"`
}

type EvalReport struct {
	SchemaVersion  string                 `json:"schema_version"`
	GeneratedAt    string                 `json:"generated_at"`
	RunnerVersion  string                 `json:"runner_version"`
	Model          string                 `json:"model"`
	PromptProfile  string                 `json:"prompt_profile"`
	PromptVersion  map[string]string      `json:"prompt_version"`
	Thresholds     map[string]float64     `json:"thresholds"`
	Summary        EvalSummary            `json:"summary"`
	Mapping        []MappingEvalResult    `json:"mapping"`
	Suggestions    []SuggestionEvalResult `json:"suggestions"`
	Baseline       *EvalBaseline          `json:"baseline,omitempty"`
	ComparisonNote string                 `json:"comparison_note,omitempty"`
}

const (
	runnerVersion              = "v2"
	defaultMinPassRate         = 0.70
	defaultMinRequiredRecall   = 0.80
	defaultMinExactMatchRatio  = 0.70
	defaultMinAvgConfidence    = 0.60
	defaultMinSuggestionsCount = 1
)

type EvalBaseline struct {
	PromptProfile       string            `json:"prompt_profile"`
	PromptVersion       map[string]string `json:"prompt_version"`
	Summary             EvalSummary       `json:"summary"`
	DeltaPassRate       float64           `json:"delta_pass_rate"`
	DeltaMapScore       float64           `json:"delta_map_score"`
	DeltaMapRecall      float64           `json:"delta_map_recall"`
	DeltaAvgSuggestions float64           `json:"delta_avg_suggestions"`
}

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	var (
		model             = flag.String("model", envOr("OPENAI_MODEL_CONVERT", envOr("OPENAI_MODEL", "gpt-4o-mini")), "OpenAI model for eval")
		promptProfile     = flag.String("prompt-profile", envOr("AI_PROMPT_PROFILE", ai.PromptProfileStaticV3), "Prompt profile: static_v3 or legacy_v2")
		baselineProfile   = flag.String("baseline-prompt-profile", "", "Optional baseline prompt profile to compare against")
		timeout           = flag.Duration("timeout", 45*time.Second, "AI request timeout")
		maxCompletion     = flag.Int("max-completion-tokens", 1200, "Max completion tokens per request")
		minPassRate       = flag.Float64("min-pass-rate", defaultMinPassRate, "Minimum global pass rate to succeed")
		mappingDataset    = flag.String("mapping-dataset", "./testdata/evals/mapping_cases.json", "Path to mapping eval dataset")
		suggestionDataset = flag.String("suggestion-dataset", "./testdata/evals/suggestion_cases.json", "Path to suggestion eval dataset")
		outputPath        = flag.String("output", "./artifacts/ai-eval-report.json", "Output report path")
		runMapping        = flag.Bool("run-mapping", true, "Run mapping eval cases")
		runSuggestions    = flag.Bool("run-suggestions", true, "Run suggestion eval cases")
	)
	flag.Parse()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		slog.Error("OPENAI_API_KEY is required for eval runner")
		os.Exit(2)
	}

	cfg := ai.DefaultConfig()
	cfg.APIKey = apiKey
	cfg.Model = *model
	cfg.PromptProfile = ai.NormalizePromptProfile(*promptProfile)
	cfg.RequestTimeout = *timeout
	cfg.MaxCompletionTokens = *maxCompletion
	cfg.DisableCache = true

	service, err := ai.NewService(cfg)
	if err != nil {
		slog.Error("failed to initialize AI service", "error", err)
		os.Exit(2)
	}

	report := EvalReport{
		SchemaVersion: "v2",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		RunnerVersion: runnerVersion,
		Model:         *model,
		PromptProfile: cfg.PromptProfile,
		PromptVersion: map[string]string{
			"mapping":     ai.ColumnMappingPromptVersion(cfg.PromptProfile),
			"suggestions": ai.SuggestionsPromptVersion(cfg.PromptProfile),
		},
		Thresholds: map[string]float64{
			"global_min_pass_rate": *minPassRate,
		},
	}

	ctx := context.Background()

	mappingCases := []MappingEvalCase{}
	suggestionCases := []SuggestionEvalCase{}
	if *runMapping {
		cases, loadErr := loadMappingCases(*mappingDataset)
		if loadErr != nil {
			slog.Error("failed to load mapping dataset", "path", *mappingDataset, "error", loadErr)
			os.Exit(2)
		}
		mappingCases = cases
	}

	if *runSuggestions {
		cases, loadErr := loadSuggestionCases(*suggestionDataset)
		if loadErr != nil {
			slog.Error("failed to load suggestion dataset", "path", *suggestionDataset, "error", loadErr)
			os.Exit(2)
		}
		suggestionCases = cases
	}

	for _, tc := range mappingCases {
		report.Mapping = append(report.Mapping, evaluateMappingCase(ctx, service, tc))
	}
	for _, tc := range suggestionCases {
		report.Suggestions = append(report.Suggestions, evaluateSuggestionCase(ctx, service, tc))
	}

	report.Summary = buildSummary(report.Mapping, report.Suggestions)

	if *baselineProfile != "" {
		bp := ai.NormalizePromptProfile(*baselineProfile)
		if bp != cfg.PromptProfile {
			baseCfg := cfg
			baseCfg.PromptProfile = bp
			baseService, baseErr := ai.NewService(baseCfg)
			if baseErr != nil {
				slog.Error("failed to initialize baseline AI service", "error", baseErr)
				os.Exit(2)
			}
			baseMapping := make([]MappingEvalResult, 0, len(mappingCases))
			baseSuggestions := make([]SuggestionEvalResult, 0, len(suggestionCases))
			for _, tc := range mappingCases {
				baseMapping = append(baseMapping, evaluateMappingCase(ctx, baseService, tc))
			}
			for _, tc := range suggestionCases {
				baseSuggestions = append(baseSuggestions, evaluateSuggestionCase(ctx, baseService, tc))
			}
			baseSummary := buildSummary(baseMapping, baseSuggestions)
			report.Baseline = &EvalBaseline{
				PromptProfile: bp,
				PromptVersion: map[string]string{
					"mapping":     ai.ColumnMappingPromptVersion(bp),
					"suggestions": ai.SuggestionsPromptVersion(bp),
				},
				Summary:             baseSummary,
				DeltaPassRate:       report.Summary.PassRate - baseSummary.PassRate,
				DeltaMapScore:       report.Summary.AverageMapScore - baseSummary.AverageMapScore,
				DeltaMapRecall:      report.Summary.AverageMapRecall - baseSummary.AverageMapRecall,
				DeltaAvgSuggestions: report.Summary.AverageSuggestions - baseSummary.AverageSuggestions,
			}
			report.ComparisonNote = "delta = primary - baseline"
		}
	}

	if err := writeJSON(*outputPath, report); err != nil {
		slog.Error("failed to write eval report", "path", *outputPath, "error", err)
		os.Exit(2)
	}

	fmt.Printf("AI Eval Report: %s\n", *outputPath)
	fmt.Printf("Pass rate: %.2f%% (%d/%d)\n",
		report.Summary.PassRate*100,
		report.Summary.PassedCases,
		report.Summary.TotalCases,
	)
	if report.Baseline != nil {
		fmt.Printf("Baseline (%s) pass rate: %.2f%%\n", report.Baseline.PromptProfile, report.Baseline.Summary.PassRate*100)
		fmt.Printf("Delta pass rate: %+0.2f%%\n", report.Baseline.DeltaPassRate*100)
	}

	if report.Summary.PassRate < *minPassRate {
		fmt.Printf("Gate failed: pass_rate %.4f < min_pass_rate %.4f\n", report.Summary.PassRate, *minPassRate)
		os.Exit(1)
	}
}

func evaluateMappingCase(ctx context.Context, service ai.Service, tc MappingEvalCase) MappingEvalResult {
	result := MappingEvalResult{
		ID:                   tc.ID,
		Name:                 tc.Name,
		Passed:               true,
		Metadata:             tc.Tags,
		EvaluatedAtUnixMilli: time.Now().UnixMilli(),
	}

	req := ai.MapColumnsRequest{
		Headers:    tc.Headers,
		SampleRows: tc.SampleRows,
		Format:     fallbackString(tc.Format, "spec"),
		FileType:   fallbackString(tc.FileType, "table"),
		SourceLang: fallbackString(tc.SourceLang, "auto"),
		SchemaHint: fallbackString(tc.SchemaHint, "auto"),
	}

	mapped, err := service.MapColumns(ctx, req)
	if err != nil {
		result.Passed = false
		result.Error = err.Error()
		result.FailedChecks = append(result.FailedChecks, "request_error")
		return result
	}

	actual := make(map[string]int, len(mapped.CanonicalFields))
	for _, m := range mapped.CanonicalFields {
		actual[m.CanonicalName] = m.ColumnIndex
	}

	result.ActualMappedFields = actual
	result.AvgConfidence = mapped.Meta.AvgConfidence
	result.MappedColumns = mapped.Meta.MappedColumns
	result.UnmappedColumns = mapped.Meta.UnmappedColumns
	result.DetectedType = mapped.Meta.DetectedType
	result.SourceLanguage = mapped.Meta.SourceLanguage

	requiredRecall, requiredHit := computeRequiredRecall(actual, tc.RequiredFields)
	result.RequiredRecall = requiredRecall
	result.RequiredHitCount = requiredHit
	result.RequiredTotal = len(tc.RequiredFields)

	exactMatchRatio, exactMatches := computeExactMatchRatio(actual, tc.Expected)
	result.ExactMatchRatio = exactMatchRatio
	result.ExactMatchCount = exactMatches
	result.ExpectedTotal = len(tc.Expected)

	forbiddenHits := findForbiddenHits(actual, tc.ForbiddenFields)
	result.ForbiddenHitFields = forbiddenHits

	minRequiredRecall := tc.MinRequiredRecall
	if minRequiredRecall == 0 {
		minRequiredRecall = defaultMinRequiredRecall
	}
	if len(tc.RequiredFields) > 0 && requiredRecall < minRequiredRecall {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "required_recall")
	}

	minExactRatio := tc.MinExactMatchRatio
	if minExactRatio == 0 {
		minExactRatio = defaultMinExactMatchRatio
	}
	if len(tc.Expected) > 0 && exactMatchRatio < minExactRatio {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "exact_match_ratio")
	}

	minAvgConfidence := tc.MinAvgConfidence
	if minAvgConfidence == 0 {
		minAvgConfidence = defaultMinAvgConfidence
	}
	if result.AvgConfidence < minAvgConfidence {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "avg_confidence")
	}

	if tc.MaxUnmappedColumns >= 0 && result.UnmappedColumns > tc.MaxUnmappedColumns {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "max_unmapped_columns")
	}

	if len(forbiddenHits) > 0 {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "forbidden_fields")
	}

	return result
}

func evaluateSuggestionCase(ctx context.Context, service ai.Service, tc SuggestionEvalCase) SuggestionEvalResult {
	result := SuggestionEvalResult{
		ID:                   tc.ID,
		Name:                 tc.Name,
		Passed:               true,
		Metadata:             tc.Tags,
		EvaluatedAtUnixMilli: time.Now().UnixMilli(),
	}

	req := ai.SuggestionsRequest{
		SpecContent: tc.SpecContent,
		Template:    fallbackString(tc.Template, "spec"),
		RowCount:    max(tc.RowCount, 1),
	}

	resp, err := service.GetSuggestions(ctx, req)
	if err != nil {
		result.Passed = false
		result.Error = err.Error()
		result.FailedChecks = append(result.FailedChecks, "request_error")
		return result
	}

	result.SuggestionsCount = len(resp.Suggestions)

	presentTypeSet := make(map[string]bool)
	unknownSet := make(map[string]bool)
	validTypes := make(map[string]bool)
	for _, t := range ai.ValidSuggestionTypes() {
		validTypes[t] = true
	}

	for _, s := range resp.Suggestions {
		t := string(s.Type)
		presentTypeSet[t] = true
		if !validTypes[t] {
			unknownSet[t] = true
		}
	}

	result.PresentTypes = mapKeysSorted(presentTypeSet)
	result.UnknownTypes = mapKeysSorted(unknownSet)

	typeHits := 0
	for _, t := range tc.ExpectedTypes {
		if presentTypeSet[t] {
			typeHits++
		}
	}
	result.ExpectedTypeHits = typeHits
	result.ExpectedTypeTotal = len(tc.ExpectedTypes)

	minSuggestions := tc.MinSuggestions
	if minSuggestions == 0 {
		minSuggestions = defaultMinSuggestionsCount
	}
	if result.SuggestionsCount < minSuggestions {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "min_suggestions")
	}

	minTypeHits := tc.MinExpectedTypeHits
	if minTypeHits == 0 && len(tc.ExpectedTypes) > 0 {
		minTypeHits = 1
	}
	if result.ExpectedTypeHits < minTypeHits {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "expected_type_hits")
	}

	if !tc.AllowedUnknownTypes && len(result.UnknownTypes) > 0 {
		result.Passed = false
		result.FailedChecks = append(result.FailedChecks, "unknown_types")
	}

	return result
}

func buildSummary(mapping []MappingEvalResult, suggestions []SuggestionEvalResult) EvalSummary {
	summary := EvalSummary{
		MappingCases:    len(mapping),
		SuggestionCases: len(suggestions),
	}

	totalMapScore := 0.0
	totalMapRecall := 0.0
	totalSuggestions := 0.0

	for _, r := range mapping {
		summary.TotalCases++
		if r.Passed {
			summary.PassedCases++
			summary.MappingPassed++
		}
		totalMapScore += r.ExactMatchRatio
		totalMapRecall += r.RequiredRecall
	}
	for _, r := range suggestions {
		summary.TotalCases++
		if r.Passed {
			summary.PassedCases++
			summary.SuggestionPassed++
		}
		totalSuggestions += float64(r.SuggestionsCount)
	}

	summary.FailedCases = summary.TotalCases - summary.PassedCases
	if summary.TotalCases > 0 {
		summary.PassRate = float64(summary.PassedCases) / float64(summary.TotalCases)
	}
	if len(mapping) > 0 {
		summary.AverageMapScore = totalMapScore / float64(len(mapping))
		summary.AverageMapRecall = totalMapRecall / float64(len(mapping))
	}
	if len(suggestions) > 0 {
		summary.AverageSuggestions = totalSuggestions / float64(len(suggestions))
	}

	return summary
}

func loadMappingCases(path string) ([]MappingEvalCase, error) {
	var out []MappingEvalCase
	if err := readJSON(path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func loadSuggestionCases(path string) ([]SuggestionEvalCase, error) {
	var out []SuggestionEvalCase
	if err := readJSON(path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func readJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func writeJSON(path string, in any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func computeRequiredRecall(actual map[string]int, required []string) (float64, int) {
	if len(required) == 0 {
		return 1, 0
	}
	hit := 0
	for _, field := range required {
		if _, ok := actual[field]; ok {
			hit++
		}
	}
	return float64(hit) / float64(len(required)), hit
}

func computeExactMatchRatio(actual map[string]int, expected map[string]int) (float64, int) {
	if len(expected) == 0 {
		return 1, 0
	}
	match := 0
	for field, idx := range expected {
		if got, ok := actual[field]; ok && got == idx {
			match++
		}
	}
	return float64(match) / float64(len(expected)), match
}

func findForbiddenHits(actual map[string]int, forbidden []string) []string {
	hits := make([]string, 0)
	for _, field := range forbidden {
		if _, ok := actual[field]; ok {
			hits = append(hits, field)
		}
	}
	return hits
}

func mapKeysSorted(in map[string]bool) []string {
	out := make([]string, 0, len(in))
	for k := range in {
		out = append(out, k)
	}
	// Keep deterministic output without importing sort package.
	if len(out) < 2 {
		return out
	}
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func fallbackString(in, fallback string) string {
	if in == "" {
		return fallback
	}
	return in
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
