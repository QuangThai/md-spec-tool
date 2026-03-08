# Social Post Draft

[agent: content]

## Twitter / X (short)

Shipped 4 improvements to the md-spec-tool AI pipeline:
- Parse-first: Google Sheets paste now consistently converts to spec format
- AI + fallback merge: high-confidence AI mappings kept even when overall confidence is low
- Prompt injection defense wired for headers & sample cells
- Stream API supports include_metadata & number_rows

---

## LinkedIn (medium)

**AI Pipeline updates in md-spec-tool**

We audited the full flow from input detection → column mapping → rendering and shipped four improvements:

1. **Parse-first unification** — Preview and Convert now agree: paste multi-column data from Google Sheets and it reliably takes the table path.

2. **AI + fallback merge** — Instead of discarding all AI mappings when confidence is low, we keep mappings with confidence ≥ 0.75 and fill the rest with rule-based fallback.

3. **Prompt injection defense** — Headers and sample cells are sanitized before reaching the LLM, closing an OWASP gap.

4. **Stream API options** — The streaming convert endpoint now respects `include_metadata` and `number_rows`, matching the non-stream API.

All backward-compatible. No migration required.

---

## Thread-style (3 tweets)

Tweet 1: Shipped 4 AI pipeline improvements for md-spec-tool. Thread 🧵

Tweet 2: 1) Parse-first — paste from Google Sheets now reliably converts to spec format. No more preview-as-table → convert-as-markdown mismatch.

Tweet 3: 2) AI merge — we keep high-confidence AI mappings (≥0.75) and fill gaps with rules. 3) Prompt injection defense wired. 4) Stream API has include_metadata & number_rows.
