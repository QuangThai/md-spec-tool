# Changelog

[agent: content]

## [Unreleased] - 2026-03-08

### Added
- **Parse-first unification**: Convert paste now uses the same logic as Preview — when you paste multi-column data (e.g. from Google Sheets CSV), it consistently takes the table conversion path instead of sometimes treating it as markdown.
- **AI + fallback merge**: When AI returns partial column mappings, high-confidence mappings (≥0.75) are now merged with rule-based fallback instead of discarded. You get the best of both worlds when overall AI confidence is low.
- **Stream API options**: The streaming conversion endpoint (`POST /api/v1/mdflow/convert/stream`) now accepts `include_metadata` and `number_rows` in the request body, matching the non-stream paste convert API.

### Changed
- **Prompt injection defense**: Headers and sample cells are sanitized before being sent to the AI mapping service. This closes an OWASP gap where user-supplied content could influence the LLM prompt.

### Fixed
- **Preview vs Convert consistency**: Preview and Convert no longer disagree on input type for multi-column paste; both use parse-first detection.
- **aiExtraFieldNames cleanup**: Removed redundant entries (assignee, component, category) that were already covered by aiCanonicalAliases.
