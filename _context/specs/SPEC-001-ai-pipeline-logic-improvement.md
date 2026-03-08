---
id: SPEC-001
title: AI Pipeline Logic Improvement
status: archived
priority: P1
effort: M
created: 2026-03-08
author: agent:brainstorm
---

# SPEC-001: AI Pipeline Logic Improvement

## Problem

Request to improve the AI pipeline and audit its logic. After auditing the full flow (input detection → column mapping → quality fallback → rendering), the following issues were identified:

1. **Preview vs Convert inconsistency** — Preview uses parse-first (fix for Google Sheet CSV); Convert uses DetectInputType first → table may preview as table but convert to markdown.
2. **AI fallback all-or-nothing** — When AI confidence is below threshold, the entire AI result is discarded; valid mappings from AI may be lost.
3. **Code hygiene** — aiExtraFieldNames overlaps with aiCanonicalAliases (dead code).
4. **Stream API missing options** — IncludeMetadata/NumberRows hardcoded; not aligned with non-stream API.

Audit details: `_context/research/2026-03-08-ai-pipeline-logic-audit.md`

**Research (subagents)**: 3 tracks executed — Technical, Best Practices, Security. Additional finding: **prompt injection defense** (SanitizeForPrompt exists but is not yet wired).

## Fact Ledger

### Verified Facts
- Convert flow: `ConvertPaste` → `DetectInputType` → if Markdown then `convertMarkdown`, if Table then parse → `convertMatrixWithFormatAndOptions` → `resolveColumnMapping` (AI) → `enhanceColumnMapping` → `shouldFallbackToTable`.
- Preview flow: parse-first; `skip_ai` defaults to true (rule-based only) for latency.
- AI thresholds: `aiMinAvgConfidence=0.75`, `aiMinMappedRatio=0.60` in `ai_mapping.go`.
- Format "table" always skips AI (rule-based column map).
- Per-mapping `Confidence` exists in `ai/schemas.go` — supports per-field merge.
- `SanitizeForPrompt` and `SanitizeHeadersForPrompt` exist in `ai/injection_guard.go` but **are not used** in the AI mapping path (OWASP prompt injection gap).

### Assumptions
- Unifying parse-first for Convert paste will improve UX (Google Sheet CSV case). Confidence: High.
- Merge AI + fallback (retain high-confidence AI mappings) aligns with best practice (external validation). Confidence: High.
- Stream options sync is low-effort, high-consistency. Confidence: High.

## Proposed Solution

### Phase 1: Unify Input Detection (Preview ↔ Convert)
- Apply parse-first to `ConvertPaste` as in Preview: if paste parses successfully as a multi-column table, treat as table; use `DetectInputType` only when parse fails or yields a single column.
- Files: `convert_handler.go`, `converter.go` (ConvertPasteWithOverridesAndOptions path).

### Phase 2: AI + Fallback Merge Strategy
- When `aiMappingMeetsThreshold` fails: instead of 100% fallback, merge AI mappings with `confidence >= 0.75` with fallback for remaining columns.
- Use pure fallback only when: AI error, or no AI mappings meet threshold.
- File: `ai_mapping.go` — `resolveColumnMappingWithFallback`.

### Phase 3: Prompt Injection Defense (Security)
- Wire `SanitizeHeadersForPrompt` before sending headers to the AI mapper.
- Sanitize sample rows (each cell) via `SanitizeForPrompt` before inclusion in the prompt.
- File: `ai_mapping.go` — before calling `c.aiMapper.MapColumns`.

### Phase 4: Code Cleanup & Stream Options
- Remove or clarify `aiExtraFieldNames` overlap with `aiCanonicalAliases`.
- Add `ConvertOptions` to `ConvertPasteStreaming` and wire from `StreamHandler`.

## Technical Approach

### Backend Changes
- `internal/converter/converter.go`: Add shared parse-first helper; refactor `ConvertPasteWithOverridesAndOptions`.
- `internal/converter/ai_mapping.go`: (1) Implement `mergeAIMappingWithFallback`; (2) Wire `SanitizeHeadersForPrompt` and sanitize sample rows before AI call.
- `internal/converter/stream_converter.go`: Accept `ConvertOptions`; propagate `IncludeMetadata`, `NumberRows`.
- `internal/http/handlers/stream_handler.go`: Parse options from request; pass to `ConvertPasteStreaming`.
- `internal/http/handlers/convert_handler.go`: Parse-first for paste (align with Preview).

### Frontend Changes
- May need to update stream request body schema if options are added (include_metadata, number_rows). Verify current API contract.

## Non-Goals
- Change AI model, prompt, or cache behavior.
- Change thresholds 0.75 / 0.60 (may be a later phase if tuning is needed).
- Preview `skip_ai` default — keep unchanged for latency.

## Success Metrics
- Convert paste with Google Sheet CSV (multi-column) always outputs table format, not markdown.
- When AI returns partial mapping (≥1 column with confidence ≥0.75), output uses AI for those columns instead of discarding all.
- Stream API supports `include_metadata`, `number_rows` same as non-stream.

## Rollout & Observability
- Feature flag: not required — logic changes are backward-compatible.
- Migration: none.
- Metrics: log `ai.mapping_merge_used` when merge strategy is used.

## Phases (effort M)

### Phase 1: Unify Parse-First for Convert Paste
- Scope: `convert_handler.go`, `converter.go`
- Acceptance Criteria:
  - [ ] Paste with multi-column table (parse OK) → always takes table path; no DetectInputType markdown short-circuit.
  - [ ] Paste single-column or parse fail → use DetectInputType as today.
  - [ ] Existing tests pass.

### Phase 2: AI + Fallback Merge
- Scope: `ai_mapping.go`
- Acceptance Criteria:
  - [ ] When AI result has ≥1 mapping with confidence≥0.75 but overall fails threshold → merge with fallback for unmapped.
  - [ ] Warning "MAPPING_AI_PARTIAL_MERGE" when merge is used.
  - [ ] Pure fallback when AI error or 0 mappings meet threshold.
- Depends on: Phase 1 (optional — can be done independently).

### Phase 3: Prompt Injection Defense
- Scope: `ai_mapping.go`, `ai/injection_guard.go`
- Acceptance Criteria:
  - [ ] Headers pass through `SanitizeHeadersForPrompt` before being sent to `MapColumns`.
  - [ ] Sample rows (each cell) pass through `SanitizeForPrompt` before inclusion in the prompt.
  - [ ] Existing injection_guard tests pass.
- Depends on: Phase 2 (can be done in parallel).

### Phase 4: Stream Options & Cleanup
- Scope: `stream_converter.go`, `stream_handler.go`, `ai_mapping.go`
- Acceptance Criteria:
  - [ ] ConvertPasteStreaming accepts ConvertOptions.
  - [ ] Stream request can send include_metadata, number_rows.
  - [ ] Clean up aiExtraFieldNames / clarify comments.

## Acceptance Criteria (Overall)
- [ ] Phase 1, 2, 3, 4 meet the acceptance criteria above.
- [ ] `go test ./...` passes.
- [ ] Manual test: paste Google Sheet CSV → convert → spec format.
- [ ] Manual test: paste ambiguous input → no regression.

## Open Questions
- [ ] Should unit tests be added for `mergeAIMappingWithFallback` with mock AI result? (Recommended: yes)
- [ ] Is a config flag needed to disable merge strategy (fallback to all-or-nothing)? (Nice-to-have)
- [x] `detect_only` query param: **Resolved** — When parse succeeds (hasTable), return "table"; else use DetectInputType. Align with conversion.
- [ ] `BuildSpecDocFromPaste`: does it use parse-first for consistency with validation?
- [ ] Stream request: default `IncludeMetadata`/`NumberRows` when nil? (e.g. nil → true/false per non-stream)
