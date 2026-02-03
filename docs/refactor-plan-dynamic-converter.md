# Refactor Plan: Dynamic Markdown Converter

## Problem Statement

The current implementation in `column_map.go` uses **hardcoded header synonyms** (180+ mappings) to convert Google Sheets to Markdown. This approach has critical limitations:

- **Not scalable**: Requires code changes for every new column type
- **Data loss**: Columns not in the mapping are marked as "unmapped" and potentially lost
- **Inflexible**: Cannot handle arbitrary spreadsheet structures

## Target Solution

A **schema-agnostic, config-driven** conversion pipeline:

```
Google Sheets API
       │
       ▼
┌──────────────────┐
│   Sheet Parser   │  ← Detects headers dynamically
│  (no mapping)    │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Normalized Model│  ← Table{Headers, Rows}
│  (schema-agnostic)│
└────────┬─────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐ ┌────────────┐
│Generic │ │Template    │  ← Config-driven (YAML)
│Renderer│ │Registry    │
└────┬───┘ └─────┬──────┘
     │           │
     ▼           ▼
┌──────────────────┐
│  Markdown Output │
└──────────────────┘
```

---

## Phase 0: Baseline & Safety Net

**Duration**: < 1 day  
**Goal**: Make refactoring safe and measurable

### Tasks

- [ ] **0.1** Create golden test fixtures
  - Typical "test spec" sheet
  - Sheet with extra/unmapped columns
  - Missing canonical columns
  - Duplicate/blank headers

- [ ] **0.2** Add snapshot tests for current Markdown output

- [ ] **0.3** Document current conversion contract
  - How "row", "header", "empty row" are handled
  - Merged cells behavior

- [ ] **0.4** (Optional) Add conversion trace/warnings collector

### Exit Criteria

- [ ] Tests validate current output is unchanged
- [ ] Safe to refactor internals

---

## Phase 1: Schema-Agnostic Table Model

**Duration**: 1-2 days  
**Goal**: Stop thinking in canonical fields during parsing

### New Core Types

```go
// Normalized table structure - no field mapping
type Table struct {
    SheetName string
    Headers   []string      // Original headers, normalized
    Rows      []Row
    Meta      TableMeta
}

type Row struct {
    Cells []string          // Aligned to Headers length
}

type TableMeta struct {
    HeaderRowIndex int
    SourceURL      string
    Warnings       []string
}
```

### Tasks

- [ ] **1.1** Create `internal/converter/table.go` with new types

- [ ] **1.2** Create `internal/converter/parser.go`
  - Parse Google Sheets values to `Table`
  - Detect header row (first non-empty row or configurable)
  - Normalize headers (trim, collapse whitespace)

- [ ] **1.3** Handle header edge cases
  - Duplicates: `Status`, `Status (2)`, `Status (3)`
  - Blank: `Column 7`, `Unnamed 7`

- [ ] **1.4** Guarantee row alignment
  - Pad short rows with empty strings
  - Truncate or warn for long rows

- [ ] **1.5** Wire existing converter to use `Table` internally
  - Adapter to maintain backward compatibility

### Exit Criteria

- [ ] All existing tests pass
- [ ] Any sheet can be parsed without losing columns

---

## Phase 2: Generic Markdown Renderer

**Duration**: 1-2 days  
**Goal**: Convert ANY spreadsheet to Markdown without restrictions

### Renderer Interface

```go
type Renderer interface {
    Render(table Table) (string, []string, error)
    // Returns: markdown, warnings, error
}
```

### Tasks

- [ ] **2.1** Create `internal/converter/renderer.go` with interface

- [ ] **2.2** Implement `GenericTableRenderer`
  - Header row from `Table.Headers`
  - Body rows from `Table.Rows`
  - Calculate column widths dynamically
  - Escape pipes `|`, normalize newlines to `<br>`

- [ ] **2.3** Add format option to API
  - `format=generic_table` parameter
  - Default to generic for new usage

- [ ] **2.4** Add unit tests for generic renderer

### Example Output

```markdown
| ID | Feature | Custom Column | Notes |
|----|---------|---------------|-------|
| 1  | Login   | Extra Data    | ...   |
| 2  | Signup  | More Data     | ...   |
```

### Exit Criteria

- [ ] Arbitrary spreadsheet → complete Markdown table
- [ ] No columns lost

---

## Phase 3: Config-Driven Legacy Renderer

**Duration**: 2-4 days  
**Goal**: Move hardcoded mappings to configuration

### Configuration Structure

```yaml
# templates/test_spec_v1.yaml
name: test_spec_v1
description: Legacy test specification format

header_synonyms:
  FieldID:
    - id
    - tc_id
    - test id
    - case_id
    - ref
    - reference
  FieldScenario:
    - scenario
    - test case
    - tc
    - case name
  FieldExpected:
    - expected
    - expected result
    - acceptance criteria
  # ... other fields

required_fields:
  - FieldScenario

output:
  type: test_spec_markdown
  unmapped_columns: append_section  # or: ignore, generic_table
```

### Tasks

- [ ] **3.1** Define template config schema (`internal/converter/template.go`)

- [ ] **3.2** Create `TemplateRegistry`
  - Load from embedded default config
  - Support external config path (env var)

- [ ] **3.3** Implement `TestSpecRenderer`
  - Uses template config for field mapping
  - Unmapped columns handled per config

- [ ] **3.4** Create `HeaderResolver` (replaces `ColumnMapper`)
  - Fed by template config, not hardcoded map

- [ ] **3.5** Embed default `test_spec_v1.yaml` for OOTB behavior

- [ ] **3.6** Deprecate `column_map.go` hardcoding

### Exit Criteria

- [ ] Default template reproduces current output exactly
- [ ] Users can customize mapping via config
- [ ] Extra columns are NOT lost (configurable behavior)

---

## Phase 4: Template-Driven Formatting

**Duration**: 3-6 days  
**Goal**: Support multiple Markdown shapes via config

### Supported Output Types

| Type | Description |
|------|-------------|
| `generic_table` | Standard Markdown table |
| `test_spec_markdown` | Legacy grouped format |
| `row_cards` | Per-row card/section format |

### Row Cards Template Example

```yaml
name: row_cards_v1
output:
  type: row_cards
  title_from: ID
  sections:
    - label: Scenario
      from: [Scenario, Test Case]
    - label: Steps
      from: [Steps, Instructions]
    - label: Expected
      from: [Expected, Outcome]
  extras:
    mode: append_section
    label: Other Columns
```

### Tasks

- [ ] **4.1** Add header reference abstraction
  - Select by exact name or first match from list

- [ ] **4.2** Implement `RowCardsRenderer`

- [ ] **4.3** Handle multi-sheet documents
  - Concatenate with `## Sheet: {name}` headings
  - Per-sheet template selection (optional)

- [ ] **4.4** Add template validation on load

- [ ] **4.5** Create cookbook templates in `templates/`

### Exit Criteria

- [ ] New formats via config, no code changes
- [ ] Multi-sheet support

---

## Phase 5: Migration & Hardening

**Duration**: 2-5 days  
**Goal**: Production readiness

### Tasks

- [ ] **5.1** Deprecation plan
  - Mark `HeaderSynonyms` as deprecated
  - Remove after N releases

- [ ] **5.2** Observability
  - Structured warnings in API response
  - Include count + samples of issues

- [ ] **5.3** Validation & UX
  - Missing header references → warning, not error
  - Best-effort output always

- [ ] **5.4** Performance optimization
  - Use `strings.Builder` for large sheets
  - Benchmark with 10k+ rows

- [ ] **5.5** Documentation
  - API docs for format options
  - Template authoring guide
  - Migration guide from hardcoded approach

- [ ] **5.6** Update frontend to support format selection

### Exit Criteria

- [ ] Stable API
- [ ] Users can self-serve new formats via config
- [ ] Complete documentation

---

## Timeline Summary

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 0 | < 1 day | 1 day |
| Phase 1 | 1-2 days | 3 days |
| Phase 2 | 1-2 days | 5 days |
| Phase 3 | 2-4 days | 9 days |
| Phase 4 | 3-6 days | 15 days |
| Phase 5 | 2-5 days | 20 days |

**Total estimated: 2-4 weeks**

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Header duplicates/ambiguity | Deterministic uniquing + warnings |
| Messy header rows | Configurable `header_row_index` |
| Data loss via normalization | Keep original in metadata |
| Backward-compat regression | Golden tests + default template |

---

## Files to Create/Modify

### New Files

```
backend/internal/converter/
├── table.go           # Table, Row, TableMeta types
├── parser.go          # Sheet → Table parsing
├── renderer.go        # Renderer interface
├── generic_renderer.go
├── testspec_renderer.go
├── template.go        # Template config types
└── template_registry.go

backend/templates/
├── test_spec_v1.yaml  # Default legacy template
└── generic.yaml       # Generic table template
```

### Modified Files

```
backend/internal/converter/
├── column_map.go      # Deprecate, then remove
├── converter.go       # Use new pipeline
└── excel_to_markdown.go
```

---

## Success Metrics

1. **Zero data loss**: All columns appear in output
2. **No code changes for new columns**: Config-only additions
3. **Backward compatible**: Existing users see same output
4. **Extensible**: New formats via YAML templates

---

## Use Cases Compatibility Matrix

The solution MUST support all existing use cases in `use-cases/`:

### Input Sources

| Source | Current | New Solution | Notes |
|--------|---------|--------------|-------|
| **Google Sheets API** | ✅ | ✅ | Primary source, returns 2D array |
| **XLSX files** | ✅ | ✅ | Via `excelize` library (`xlsx_parser.go`) |
| **TSV paste/file** | ✅ | ✅ | Via `paste_parser.go` with delimiter detection |
| **CSV paste/file** | ✅ | ✅ | Via `paste_parser.go` with delimiter detection |

### Use Case Examples Analysis

| File | Type | Columns | Compatible |
|------|------|---------|------------|
| `example-1.md` | Free-form text (not table) | N/A | ✅ Pass-through |
| `example-2.md` | TSV table | No, Item Name, Item Type, Required/Optional, Input Restrictions, Display Conditions, Action, Navigation Destination | ✅ All 8 columns preserved |
| `example-3.md` | TSV table with Notes | No, Item Name, Item Type, Required/Optional, Input Restrictions, Display Conditions, Action, Navigation Destination, **Notes** | ✅ All 9 columns preserved |
| `example-4.md` | Free-form text (not table) | N/A | ✅ Pass-through |
| `example-5.md` | Free-form text (not table) | N/A | ✅ Pass-through |
| `example-6.md` | TSV table with Notes | Same as example-3 | ✅ All 9 columns preserved |
| `table.tsv` | Raw TSV file | Same structure as examples | ✅ All columns preserved |

### Input Format Detection (Already Implemented)

The existing parsers handle format detection:

```
backend/internal/converter/
├── xlsx_parser.go      # Excel files via excelize
├── paste_parser.go     # TSV/CSV with auto delimiter detection
├── input_detect.go     # Markdown vs Table scoring
└── converter.go        # Orchestration
```

### Parser → Table Integration

The new `Table` model will receive data from existing parsers:

```go
// xlsx_parser.go already returns [][]string
func (p *XLSXParser) Parse(file) ([][]string, error)

// paste_parser.go already returns [][]string  
func (p *PasteParser) Parse(content) ([][]string, error)

// NEW: Convert to normalized Table
func ToTable(headers []string, rows [][]string) *Table {
    return &Table{
        Headers: headers,           // First row
        Rows:    normalizeRows(rows[1:], len(headers)),
    }
}
```

### Key Columns in Use Cases

These columns appear in the examples and MUST be supported:

| Column Name | Japanese Equivalent | Status |
|-------------|---------------------|--------|
| No | 番号 | ✅ In current mapping |
| Item Name | 項目名 | ✅ In current mapping |
| Item Type | 項目種別 | ✅ In current mapping |
| Required/Optional | 必須/任意 | ✅ In current mapping |
| Input Restrictions | 入力制限 | ✅ In current mapping |
| Input Constraints | - | ⚠️ Synonym of Input Restrictions |
| Display Conditions | 表示条件 | ✅ In current mapping |
| Display Condition | - | ⚠️ Needs synonym |
| Action | アクション | ✅ In current mapping |
| Navigation Destination | 遷移先 | ✅ In current mapping |
| Destination | - | ⚠️ Synonym needed |
| Notes | - | ❌ NOT in current mapping! |

**Problem identified**: `Notes` column in example-3, example-6, table.tsv is currently UNMAPPED and would be lost!

### Solution Guarantees

With the new dynamic approach:

1. **Generic mode**: ALL columns preserved regardless of mapping
2. **Legacy mode**: Unmapped columns added to "Additional Fields" section
3. **No data loss**: Even unknown columns like `Notes` are rendered

---

## Affected Parsers (No Breaking Changes)

The refactor focuses on the **output/rendering layer**, not input parsing:

| Parser | Change Required |
|--------|-----------------|
| `xlsx_parser.go` | None - already returns `[][]string` |
| `paste_parser.go` | None - already returns `[][]string` |
| `input_detect.go` | None - scoring logic unchanged |
| `google_sheet_fetcher.go` | None - already returns sheet values |

Only the **converter/renderer** layer changes:

| File | Change |
|------|--------|
| `column_map.go` | Deprecate hardcoded map → move to config |
| `converter.go` | Use new `Table` model + `Renderer` interface |
| `excel_to_markdown.go` | Refactor to use `GenericRenderer` |

---

## Test Data for Golden Tests (Phase 0)

Use these files as test fixtures:

```
use-cases/
├── example-2.md    # 8 columns, standard structure
├── example-3.md    # 9 columns (includes Notes!)
├── example-6.md    # 9 columns, Japanese content
└── table.tsv       # Raw TSV, 33 rows
```

Golden test should verify:
- [ ] All columns appear in output
- [ ] Column order preserved
- [ ] Multi-line cell content handled (e.g., row 3 in example-2)
- [ ] Special characters escaped (pipes, Japanese text)
- [ ] Empty cells handled correctly
