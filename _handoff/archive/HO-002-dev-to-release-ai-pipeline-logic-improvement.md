---
id: HO-002
from: dev
to: release
priority: P1
status: done
created: 2026-03-08T14:30:00Z
spec: SPEC-001
total_phases: 4
current_phase: 4
loop_count: 0
origin_handoff_id: HO-001
output_mode: last_message
---

# Release: AI Pipeline Logic Improvement

## Contract
- **task_description**: Deploy AI Pipeline Logic Improvement (SPEC-001) to production. Run pre-deploy checks, present deploy command, verify post-deploy health.
- **acceptance_criteria**: All health checks pass, no errors in docker logs, smoke tests green.
- **context_keys**: `_context/specs/SPEC-001-ai-pipeline-logic-improvement.md`, `_handoff/archive/HO-001-brainstorm-to-dev-ai-pipeline-logic-improvement.md`
- **output_mode**: last_message

## Summary
SPEC-001 delivered four improvements: (1) parse-first for Convert paste so Google Sheet CSV and multi-column paste always take the table path; (2) AI + fallback merge when ≥1 mapping has confidence ≥0.75 instead of all-or-nothing discard; (3) prompt injection defense by sanitizing headers and sample cells before AI calls; (4) stream API supports include_metadata and number_rows, with aiExtraFieldNames cleanup. No migration, no breaking changes.

## Changes
### Backend
- `internal/converter/converter.go` — `resolvePasteParseFirst`, `AnalyzePasteForConvert`; parse-first in `ConvertPasteWithOverridesAndOptions`
- `internal/http/handlers/convert_handler.go` — parse-first for detect_only and conversion; uses `AnalyzePasteForConvert`
- `internal/converter/ai_mapping.go` — `mergeAIMappingWithFallback`, `countHighConfidenceAIMappings`; prompt-safe headers via `ai.SanitizeHeadersForPrompt`; prompt-safe sample rows via `ai.SanitizeForPrompt`; aiExtraFieldNames emptied
- `internal/converter/stream_converter.go` — `ConvertOptions` param; `table.Meta.IncludeMetadata/NumberRows` from options
- `internal/http/handlers/stream_handler.go` — `resolveConvertOptions(req.IncludeMetadata, req.NumberRows)`; pass options to `ConvertPasteStreaming`
- `internal/converter/converter_fallback_test.go`, `stream_converter_test.go`, `ai_mapping_test.go` — new/updated tests

### Frontend
- No changes. PasteConvertRequest already has include_metadata, number_rows for stream schema.

### Database
- Migration: no
- Migration name: N/A

## QA Status
- All automated checks: ✅ (1 pre-existing env failure: gsheet fixture missing file)
- Acceptance criteria: ✅
- QA report: HO-001 (Phases 3, 4 — Full Verification) — PASS

## Deploy Notes
- Environment variables: none new
- Migration steps: none
- Breaking changes: none
- Rollback plan: Revert commit(s) containing SPEC-001 changes; backend/frontend APIs remain backward-compatible.

## Post-Deploy Verification
- [ ] Health check passes
- [ ] Smoke test: POST /api/v1/mdflow/paste with multi-column table → spec format
- [ ] Monitor: conversion logs for MAPPING_AI_PARTIAL_MERGE (expected when merge path used)
