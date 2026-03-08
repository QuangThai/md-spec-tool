---
id: HO-003
from: release
to: content
priority: P1
status: done
created: 2026-03-08T14:35:00Z
spec: SPEC-001
total_phases: 1
current_phase: 1
loop_count: 0
output_mode: last_message
---

# Content: AI Pipeline Logic Improvement

## Contract
- **task_description**: Generate content artifacts (changelog, blog, social, docs) for the shipped feature. Write for the target audience.
- **acceptance_criteria**: At least changelog entry and one blog/social draft produced. Content is accurate and references actual implementation.
- **context_keys**: _context/specs/SPEC-001-ai-pipeline-logic-improvement.md, _context/product-state.md, _context/research/2026-03-08-ai-pipeline-logic-audit.md
- **output_mode**: last_message

## What Was Shipped
AI Pipeline Logic Improvement: four backend changes that make conversion more reliable and secure. (1) Paste from Google Sheets CSV now consistently takes the table path instead of sometimes being treated as markdown. (2) When AI returns partial mappings, high-confidence columns (≥0.75) are now merged with rule-based fallback instead of discarded. (3) Headers and sample cells are sanitized before AI calls to prevent prompt injection. (4) Stream conversion API now respects `include_metadata` and `number_rows` options.

## Key Features
- Parse-first unification: Convert and Preview behave consistently for multi-column paste
- AI + fallback merge: Retains good AI mappings when overall confidence is low
- Prompt injection defense: SanitizeHeadersForPrompt, SanitizeForPrompt wired before MapColumns
- Stream options: include_metadata, number_rows supported in stream request

## Screenshots / Demo Points
- Paste Google Sheet CSV (multi-column) → convert → spec format (no markdown fallback)
- Stream API request with `include_metadata: false` → output without front matter

## Technical Highlights
- mergeAIMappingWithFallback with 0.75 threshold; MAPPING_AI_PARTIAL_MERGE warning when used
- aiExtraFieldNames cleanup (assignee/component/category in aiCanonicalAliases)

## Target Audience
- Users pasting from Google Sheets, Excel
- Integrators using the stream conversion API
- Security-conscious deployers

## Content Suggestions
- [ ] Changelog entry
- [ ] Blog post / social post
- [ ] Documentation update (API docs for stream options)
- [ ] Demo video script (optional)
