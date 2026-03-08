# Research: AI Pipeline Logic Audit

> Date: 2026-03-08  
> Agent: brainstorm  
> Topic: AI Pipeline Logic Audit — Quality and Accuracy Improvements

## Executive Summary

Full audit of the AI pipeline flow from input detection → column mapping → quality fallback → rendering. Identified **8 logic issues** (2 potential bugs, 6 improvements) with a proposed spec structure for remediation.

## Top Sources Consulted

**Local codebase:**
- `backend/internal/converter/ai_mapping.go` — AI column mapping logic
- `backend/internal/converter/converter.go` — Main convert flow
- `backend/internal/converter/input_detect.go` — Input type detection
- `backend/internal/converter/mapping_quality.go` — Quality evaluation
- `backend/internal/converter/dynamic_mapping.go` — Heuristic enhancement
- `backend/internal/ai/service.go`, `client.go` — AI service layer
- `backend/internal/http/handlers/preview_handler.go`, `convert_handler.go`, `stream_handler.go`

**External (subagent research):**
- [OpenAI Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)
- [Schema Validation for Real Pipelines](https://collinwilkins.com/articles/structured-output)
- [Self-Recovering Structured Outputs](https://usebrainbits.com/blog/self-recovering-structured-outputs)
- [Measuring LLM Confidence](https://kmad.ai/Measuring-LLM-Confidence)
- [CleverCSV / DuckDB auto-detection](https://clevercsv.readthedocs.io)
- OWASP API Security Top 10, OWASP LLM Prompt Injection Prevention

## AI Pipeline Flow (High-Level)

```
Input (paste/file)
    │
    ▼
DetectInputType(text) ──► Markdown? ──► convertMarkdown() [no AI]
    │                         │
    ▼ Table                   └─────────────────────────────────► Response
Parse (PasteParser/XLSX)
    │
    ▼
DetectHeaderRow(matrix)
    │
    ▼
resolveColumnMapping(ctx, headers, dataRows, format)
    │
    ├─ format=="table" ──► fallback only [no AI]
    ├─ skipAI==true ────► fallback only [preview fast path]
    ├─ aiService==nil ──► fallback only
    │
    └─ AI path:
         ├─ MapColumns(cleanHeaders, sampleRows, schemaHint) 
         ├─ On error: fallback + warning
         └─ On low confidence (<0.75 or mappedRatio<0.6): fallback + warning
    │
    ▼
enhanceColumnMapping(headers, dataRows, colMap)  [heuristic fill gaps]
    │
    ▼
evaluateMappingQuality(headerConfidence, headers, colMap)
    │
    ▼
shouldFallbackToTable(format, quality)?
    │  Yes: format = "table", add warning
    │
    ▼
tableParser.MatrixToTable → Renderer.Render → Response
```

## Logic Issues & Improvement Opportunities

### 1. **Input Detection: Ambiguous Default** (Minor)

**File**: `input_detect.go:60-65`

When `mdScore == tableScore` or both fall below the threshold, the pipeline **defaults to Markdown**:

```go
// Default to markdown when ambiguous
return InputAnalysis{Type: InputTypeMarkdown, Confidence: 50, ...}
```

**Issue**: CSV/TSV with weak signals (e.g., 2 columns, inconsistent tab/comma usage) may be misclassified as markdown → incorrect output.

**Recommendation**: Consider additional heuristics (e.g., multiple rows with similar length → favor table) or expose a user override when confidence is low.

---

### 2. **Preview Handler: Parse-First vs DetectInputType Mismatch** (Design)

**File**: `preview_handler.go:136-158`

Preview uses **parse-first**: if `parser.Parse()` succeeds and yields ≥2 columns → treat as table. `DetectInputType` is only invoked when parse fails or produces a single column.

Convert (ConvertPaste) uses **DetectInputType first**: if markdown is detected, it converts as markdown immediately and does not parse.

**Consequence**: Preview may show a table preview for input that Convert will process as markdown. Inconsistent UX.

**Recommendation**: Unify logic: Convert should use the same parse-first strategy for paste (as already fixed in preview: "Google Sheet CSV misclassified as markdown").

---

### 3. **AI Mapping: Fallback All-or-Nothing** (Improvement)

**File**: `ai_mapping.go:136-149`

When AI returns results but `!aiMappingMeetsThreshold(meta, len(headers))` or `len(colMap)==0`, the pipeline **discards the entire AI result** and falls back to rule-based mapping.

**Issue**: AI may correctly map a subset (e.g., 3/5 columns) but because avg confidence is 0.72 < 0.75, all mappings are rejected.

**Recommendation**: Merge strategy: retain high-confidence AI mappings and fill gaps with fallback instead of replacing everything.

---

### 4. **Schema Hint: String Containment Edge Cases** (Low)

**File**: `ai_mapping.go:164-193` — `inferSchemaHint`

Uses `strings.Contains(joined, " endpoint")` to detect api_spec. Strings such as `"endpoint_description"` also match.

**Impact**: Low — schema hint is only a prompt suggestion, not core logic.

---

### 5. **aiExtraFieldNames vs aiCanonicalAliases Overlap** (Code hygiene)

**File**: `ai_mapping.go:324-334`, `329-333`

`aiExtraFieldNames` = {"assignee", "component", "category"} — "valid metadata but don't map to CanonicalField".

Meanwhile `aiCanonicalAliases` includes:
- "assignee" → FieldAssignee  
- "component" → FieldComponent  
- "category" → FieldCategory  

Because `aiCanonicalAliases` is checked first, entries in `aiExtraFieldNames` for these three fields **are never used**. Dead code / confusing.

**Recommendation**: Remove assignee, component, and category from aiExtraFieldNames; or clarify intent in comments.

---

### 6. **Stream Converter: Hardcoded Options** (Minor)

**File**: `stream_converter.go:165`

```go
table.Meta.IncludeMetadata = true  // hardcoded
// No NumberRows option
```

ConvertPasteWithOverridesAndOptions accepts overrides; stream conversion always sets IncludeMetadata=true and does not support NumberRows.

**Recommendation**: Pass ConvertOptions into ConvertPasteStreaming to align with the non-stream API.

---

### 7. **Mapping Quality: coreCoverage Max Logic** (Review)

**File**: `mapping_quality.go:73-88`

```go
coreCoverage := float64(testCaseMapped) / float64(len(coreFieldsTestCase))
// ... take max across all schema types
if specTableCoverage > coreCoverage { coreCoverage = specTableCoverage }
```

A table may match both test-case and spec-table schemas (e.g., has No, ItemName, Scenario). Taking max coverage can inflate the score when the true schema is mixed or unclear.

**Impact**: May delay fallback-to-table when fallback is appropriate. Requires testing with mixed-schema inputs.

---

### 8. **Preview skip_ai Default** (Design — OK)

**File**: `preview_handler.go:56`

```go
skipAI := c.Query("skip_ai") != "false"  // default: skip AI
```

Preview defaults to **no AI** (rule-based only). Likely intentional to reduce latency. Not a bug, but should be documented for the frontend (when to send `skip_ai=false`).

---

## Recommendation Summary

| # | Issue | Severity | Action |
|---|-------|----------|--------|
| 1 | Input detection ambiguous default | Minor | Add heuristics or user override |
| 2 | Preview vs Convert parse order | Medium | Unify parse-first for Convert paste |
| 3 | AI fallback all-or-nothing | Medium | Merge AI + fallback strategy |
| 4 | Schema hint edge cases | Low | Optional: word-boundary matching |
| 5 | aiExtraFieldNames overlap | Low | Clean up dead / clarify |
| 6 | Stream options hardcoded | Minor | Pass ConvertOptions |
| 7 | coreCoverage max logic | Low | Add tests, consider weighting |
| 8 | Preview skip_ai default | Info | Document only |

---

## Subagent Research Synthesis (2026-03-08)

> 3 parallel subagents (Track A: Technical, Track B: Best Practices, Track C: Security) executed per kd-brainstorm workflow.

### Track A — Technical Feasibility (Agent 9db6f49)
- **Parse-first**: Preview pattern at `preview_handler.go:135-155`; Convert uses detect-first. Adding a shared helper is feasible.
- **detect_only** semantics: Parse-first changes behavior — decide: (a) keep heuristic, or (b) return type based on parse when hasTable.
- **AI merge**: Per-mapping `Confidence` exists in `ai/schemas.go`. Implementing `mergeAIMappingWithFallback` is feasible; define precedence when AI and fallback map the same field.
- **Stream options**: `PasteConvertRequest` already has `IncludeMetadata`, `NumberRows`; only requires plumbing in `stream_handler.go` and `stream_converter.go`.

### Track B — External Best Practices (Agent 03344129)
- **Sources**: [OpenAI Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs), [Schema Validation for Real Pipelines](https://collinwilkins.com/articles/structured-output), [Measuring LLM Confidence](https://kmad.ai/Measuring-LLM-Confidence).
- **Best practices**: (1) Per-field confidence instead of global; (2) partial merge instead of all-or-nothing; (3) validate-repair-retry for structured output; (4) native JSON schema with gpt-4o/mini.
- **Rejected**: All-or-nothing AI confidence; blind retries without feedback; schema-only validation without verification.
- **Note**: Track B suggests "detect-then-parse" for delimiter; our parse-first = parse-as-detection for the table path (PasteParser handles delimiter). Both approaches are compatible: try parse → if table OK then use; else heuristic.

### Track C — Security (Agent 8a2c2c8)
- **Critical gap**: `SanitizeForPrompt` and `SanitizeHeadersForPrompt` exist in `ai/injection_guard.go` but **are not used** in the AI mapping path. Headers and sample rows sent to the LLM are not sanitized for prompt injection.
- **OWASP**: API4 (Unrestricted Resource Consumption), API8 (Security Misconfiguration), OWASP LLM Prompt Injection Prevention.
- **Mitigations required**: Wire `SanitizeHeadersForPrompt` in `ai_mapping.go`; sanitize sample rows before inclusion in the prompt.
- **Existing**: MaxPasteBytes, MaxUploadBytes, BYOK cache isolation, file validation — adequate.

### Updated Fact Ledger
- **Promoted**: Per-field merge aligns with external best practices.
- **Promoted**: Prompt injection defenses exist in code but are not wired — requires an additional phase.
- **Open**: `detect_only` semantics; whether `BuildSpecDocFromPaste` uses parse-first; default for IncludeMetadata/NumberRows when nil in stream.

---

## Next Steps

1. Update SPEC-001: add Phase 2.5 (Prompt Injection Defense) or merge into Phase 2.
2. Priorities: #2 (unify parse), #3 (merge AI+fallback), **prompt injection** (security).
