# Smarter AI Conversion: What We Shipped

[agent: content]

Ever pasted a table from Google Sheets and gotten markdown instead of a proper spec? Or wondered why the preview looked right but the final convert didn’t? We fixed that — and a few things under the hood.

## The Problem

We audited the AI pipeline and found four issues:

1. **Preview and Convert disagreed** — Preview used parse-first (try parsing as table first), Convert used DetectInputType first. Paste the same multi-column data and you could get a table preview but a markdown output.

2. **AI was all-or-nothing** — If AI’s overall confidence dipped below threshold, we threw away *all* AI mappings, including the ones it got right.

3. **User input reached the AI unsanitized** — Headers and sample cells went straight to the LLM. We had sanitization functions, but they weren’t wired in.

4. **Stream API was hardcoded** — The streaming convert endpoint always used `include_metadata: true` and didn’t support `number_rows` at all.

## The Solution

### Parse-first unification
Convert paste now matches Preview: if the input parses as a multi-column table, we treat it as a table. Single-column or parse failure still falls back to DetectInputType.

### AI + fallback merge
When AI returns partial mappings, we keep anything with confidence ≥ 0.75 and fill the rest with rule-based fallback. You get AI where it’s confident, rules where it isn’t.

### Prompt injection defense
Headers and sample rows are sanitized with `SanitizeHeadersForPrompt` and `SanitizeForPrompt` before being sent to the AI. We close the gap OWASP flagged.

### Stream options
The stream endpoint now accepts `include_metadata` and `number_rows`, same as the non-stream paste API.

## How It Works

- **Parse-first**: We try `PasteParser.Parse()`; if it succeeds and we have ≥2 columns, we treat it as a table.
- **AI merge**: We call `mergeAIMappingWithFallback` with a 0.75 threshold. AI mappings override fallback for the same field.
- **Sanitization**: Before `MapColumns`, we run headers and sample cells through the injection guard.
- **Stream options**: `resolveConvertOptions` in the stream handler passes options through to `ConvertPasteStreaming`.

## Try It Out

1. Copy a multi-column table from Google Sheets (TSV or CSV) and paste into md-spec-tool.
2. Watch it consistently take the table path and produce spec format.
3. For stream integrators: send `include_metadata: false` or `number_rows: true` in your stream request body.
