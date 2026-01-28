# Google Sheet to MDFlow Converter - Implementation Plan

## Overview

Tool to convert Google Sheet specs (table format) to MDFlow format, helping AI coding agents understand and implement specs accurately.

**Key Features:**
- No login required
- Two input methods: Paste table / Upload .xlsx
- Output: MDFlow executable markdown

---

## MDFlow Format Reference

MDFlow is a markdown-based CLI framework that turns `.md` files into executable AI agent scripts.

### MDFlow Structure

```yaml
---
# Frontmatter (YAML config)
model: sonnet
_feature_name: "Authentication"
_inputs:
  _service_name:
    type: text
    description: "Service name"
---
# Body (Markdown prompt)

## Requirements
- Feature: {{ _feature_name }}
- Instructions: ...

@./src/api.ts          # File imports
!`git branch`          # Command inlines
```

### Conversion Mapping

| Google Sheet Column | MDFlow Element |
|---------------------|----------------|
| Feature/Task | `_feature_name` variable |
| Instructions | Markdown body |
| Inputs | `_inputs` object |
| Expected Output | Prompt structure |
| Code Context | `@` imports |
| Parameters | Template variables |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Paste Tab   │  │ Upload Tab  │  │ Output Panel        │  │
│  │ (textarea)  │  │ (.xlsx)     │  │ (preview/download)  │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────▲──────────┘  │
│         │                │                     │             │
└─────────┼────────────────┼─────────────────────┼─────────────┘
          │                │                     │
          ▼                ▼                     │
┌─────────────────────────────────────────────────────────────┐
│                        Backend                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                    Converter Service                 │    │
│  │  ┌───────────┐  ┌───────────┐  ┌─────────────────┐  │    │
│  │  │ Paste     │  │ XLSX      │  │ Header          │  │    │
│  │  │ Parser    │  │ Parser    │  │ Detection       │  │    │
│  │  └─────┬─────┘  └─────┬─────┘  └────────┬────────┘  │    │
│  │        │              │                  │           │    │
│  │        ▼              ▼                  ▼           │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │              CellMatrix                      │    │    │
│  │  │         (normalized rows/cols)               │    │    │
│  │  └─────────────────────┬───────────────────────┘    │    │
│  │                        │                            │    │
│  │                        ▼                            │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │           Column Mapper                      │    │    │
│  │  │    (fuzzy matching + synonyms)               │    │    │
│  │  └─────────────────────┬───────────────────────┘    │    │
│  │                        │                            │    │
│  │                        ▼                            │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │           SpecDoc Model                      │    │    │
│  │  │    []SpecRow + Metadata + Warnings           │    │    │
│  │  └─────────────────────┬───────────────────────┘    │    │
│  │                        │                            │    │
│  │                        ▼                            │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │         MDFlow Renderer                      │    │    │
│  │  │   (template engine → markdown output)        │    │    │
│  │  └─────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 0: Scope & Fixtures (0.5 day)

### Objective
Lock exact MDFlow output format and create test fixtures.

### Tasks

- [ ] **0.1** Analyze `authentication_requirements_testcases.xlsx` structure
- [ ] **0.2** Create 3-5 sample paste inputs (TSV format)
- [ ] **0.3** Define golden MDFlow output for each fixture
- [ ] **0.4** Document accepted table formats

### Deliverables

```
docs/
├── fixtures/
│   ├── input/
│   │   ├── auth_spec.tsv
│   │   ├── api_endpoints.tsv
│   │   └── test_cases.xlsx
│   └── output/
│       ├── auth_spec.mdflow.md
│       ├── api_endpoints.mdflow.md
│       └── test_cases.mdflow.md
└── TABLE_FORMATS.md
```

### Acceptance Criteria
- [ ] Golden outputs defined for at least 3 fixtures
- [ ] Table format variations documented

---

## Phase 1: Backend Ingestion (1 day)

### Objective
Parse paste text and .xlsx into normalized matrix, detect headers, map columns.

### Tasks

- [ ] **1.1** Create converter package structure
- [ ] **1.2** Implement CellMatrix type and normalization
- [ ] **1.3** Implement paste parser (TSV/CSV detection)
- [ ] **1.4** Implement XLSX parser (reuse excelize)
- [ ] **1.5** Implement header detection algorithm
- [ ] **1.6** Implement fuzzy column mapping
- [ ] **1.7** Create SpecRow/SpecDoc models
- [ ] **1.8** Write unit tests

### File Structure

```
backend/internal/converter/
├── matrix.go           # CellMatrix type, normalization
├── paste_parser.go     # TSV/CSV parsing
├── xlsx_parser.go      # excelize → matrix
├── header_detect.go    # Header row scoring
├── column_map.go       # Canonical fields + fuzzy matching
├── model.go            # SpecRow, SpecDoc
├── warnings.go         # Warning types
└── converter_test.go   # Unit tests
```

### Data Models

```go
// CellMatrix represents normalized spreadsheet data
type CellMatrix [][]string

// CanonicalField represents mapped column types
type CanonicalField string

const (
    FieldFeature      CanonicalField = "feature"
    FieldInstructions CanonicalField = "instructions"
    FieldInputs       CanonicalField = "inputs"
    FieldExpected     CanonicalField = "expected_output"
    FieldPrecondition CanonicalField = "precondition"
    FieldPriority     CanonicalField = "priority"
    FieldNotes        CanonicalField = "notes"
)

// ColumnMap maps canonical fields to column indices
type ColumnMap map[CanonicalField]int

// SpecRow represents a single spec requirement
type SpecRow struct {
    Feature       string
    Instructions  string
    Inputs        string
    Expected      string
    Precondition  string
    Priority      string
    Notes         string
    Metadata      map[string]string
}

// SpecDoc represents the complete parsed document
type SpecDoc struct {
    Title      string
    Rows       []SpecRow
    Warnings   []string
    Meta       struct {
        SheetName  string
        HeaderRow  int
        ColumnMap  ColumnMap
    }
}
```

### Header Synonyms Map

```go
var headerSynonyms = map[string]CanonicalField{
    // Feature
    "feature":      FieldFeature,
    "req":          FieldFeature,
    "requirement":  FieldFeature,
    "story":        FieldFeature,
    "user story":   FieldFeature,
    "task":         FieldFeature,
    "title":        FieldFeature,
    "name":         FieldFeature,
    "test case":    FieldFeature,
    "tc":           FieldFeature,
    
    // Instructions
    "instructions":  FieldInstructions,
    "description":   FieldInstructions,
    "steps":         FieldInstructions,
    "test steps":    FieldInstructions,
    "action":        FieldInstructions,
    
    // Inputs
    "inputs":       FieldInputs,
    "input":        FieldInputs,
    "test data":    FieldInputs,
    "data":         FieldInputs,
    "parameters":   FieldInputs,
    
    // Expected
    "expected":           FieldExpected,
    "expected output":    FieldExpected,
    "expected result":    FieldExpected,
    "acceptance":         FieldExpected,
    "acceptance criteria": FieldExpected,
    "result":             FieldExpected,
    
    // Precondition
    "precondition":   FieldPrecondition,
    "preconditions":  FieldPrecondition,
    "prerequisite":   FieldPrecondition,
    "setup":          FieldPrecondition,
    
    // Priority
    "priority":   FieldPriority,
    "severity":   FieldPriority,
    "importance": FieldPriority,
    
    // Notes
    "notes":    FieldNotes,
    "comments": FieldNotes,
    "remarks":  FieldNotes,
}
```

### Acceptance Criteria
- [ ] Paste TSV/CSV parsed correctly
- [ ] XLSX single sheet parsed correctly
- [ ] Header row detected with accuracy > 80%
- [ ] Column mapping works with synonyms
- [ ] Unit tests pass

---

## Phase 2: MDFlow Renderer (0.5 day)

### Objective
Convert SpecDoc → MDFlow markdown with proper frontmatter.

### Tasks

- [ ] **2.1** Create MDFlow renderer
- [ ] **2.2** Implement frontmatter generation
- [ ] **2.3** Implement body template
- [ ] **2.4** Handle edge cases (escaping, formatting)
- [ ] **2.5** Write renderer tests

### File Structure

```
backend/internal/converter/
├── ... (existing)
├── renderer.go         # MDFlow rendering
├── templates/
│   └── default.tmpl    # Default MDFlow template
└── renderer_test.go
```

### MDFlow Output Template

```markdown
---
name: "{{ .Title }}"
version: "1.0"
generated_at: "{{ .GeneratedAt }}"
source: "{{ .Source }}"
---

# {{ .Title }}

## Overview

This specification was converted from a spreadsheet.
Total requirements: {{ len .Rows }}

---

## Requirements

{{ range $i, $row := .Rows }}
### {{ inc $i }}. {{ $row.Feature }}

{{ if $row.Precondition }}
**Precondition:** {{ $row.Precondition }}
{{ end }}

**Instructions:**
{{ $row.Instructions }}

{{ if $row.Inputs }}
**Inputs:**
{{ $row.Inputs }}
{{ end }}

**Expected Output:**
{{ $row.Expected }}

{{ if $row.Priority }}
**Priority:** {{ $row.Priority }}
{{ end }}

{{ if $row.Notes }}
**Notes:** {{ $row.Notes }}
{{ end }}

---

{{ end }}

## Metadata

- Source: Spreadsheet Import
- Generated: {{ .GeneratedAt }}
- Warnings: {{ len .Warnings }}

{{ if .Warnings }}
### Warnings
{{ range .Warnings }}
- {{ . }}
{{ end }}
{{ end }}
```

### Acceptance Criteria
- [ ] Valid YAML frontmatter generated
- [ ] Markdown body properly formatted
- [ ] Special characters escaped
- [ ] Golden output tests pass

---

## Phase 3: API Endpoints (0.5 day)

### Objective
Expose converter functionality via REST API.

### Tasks

- [ ] **3.1** Create convert handlers
- [ ] **3.2** Implement `/api/convert/paste` endpoint
- [ ] **3.3** Implement `/api/convert/xlsx` endpoint
- [ ] **3.4** Add request validation
- [ ] **3.5** Write integration tests

### File Structure

```
backend/internal/http/
├── ... (existing)
└── convert_handlers.go
```

### API Endpoints

#### POST /api/convert/paste

```json
// Request
{
  "text": "Feature\tDescription\tExpected\nLogin\tUser enters credentials\tSuccess message"
}

// Response
{
  "success": true,
  "mdflow": "---\nname: \"Converted Spec\"\n...",
  "warnings": [],
  "meta": {
    "rows_count": 1,
    "header_row": 0,
    "detected_columns": {
      "feature": 0,
      "instructions": 1,
      "expected_output": 2
    }
  }
}
```

#### POST /api/convert/xlsx

```
// Request (multipart/form-data)
file: <xlsx file>
sheet_name: "Sheet1" (optional)

// Response
{
  "success": true,
  "mdflow": "---\nname: \"Converted Spec\"\n...",
  "warnings": [],
  "sheets": ["Sheet1", "Sheet2"],
  "selected_sheet": "Sheet1",
  "meta": { ... }
}
```

### Acceptance Criteria
- [ ] Both endpoints return valid MDFlow
- [ ] Error handling for invalid input
- [ ] Sheet selection works for multi-sheet xlsx

---

## Phase 4: Frontend UI (1 day)

### Objective
Build converter page với Paste/Upload tabs và output preview.

### Tasks

- [ ] **4.1** Create `/converter` route
- [ ] **4.2** Build ConverterTabs component
- [ ] **4.3** Build PasteInput component
- [ ] **4.4** Build UploadInput component
- [ ] **4.5** Build OutputPanel component
- [ ] **4.6** Add Zustand store for converter state
- [ ] **4.7** Implement copy/download functionality
- [ ] **4.8** Add loading states và error handling

### File Structure

```
frontend/
├── app/
│   └── converter/
│       ├── page.tsx
│       └── components/
│           ├── ConverterTabs.tsx
│           ├── PasteInput.tsx
│           ├── UploadInput.tsx
│           ├── SheetPicker.tsx
│           └── OutputPanel.tsx
├── lib/
│   └── api/
│       └── converter.ts
└── stores/
    └── converterStore.ts
```

### Zustand Store

```typescript
interface ConverterState {
  // Input
  mode: 'paste' | 'xlsx'
  pasteText: string
  file: File | null
  sheets: string[]
  selectedSheet: string
  
  // Output
  mdflowOutput: string
  warnings: string[]
  meta: ConvertMeta | null
  
  // UI
  loading: boolean
  error: string | null
  
  // Actions
  setMode: (mode: 'paste' | 'xlsx') => void
  setPasteText: (text: string) => void
  setFile: (file: File | null) => void
  setSelectedSheet: (sheet: string) => void
  convert: () => Promise<void>
  copyToClipboard: () => void
  downloadMdflow: () => void
  reset: () => void
}
```

### UI Wireframe

```
┌─────────────────────────────────────────────────────────────┐
│  🔄 Google Sheet to MDFlow Converter                        │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┬──────────────┐                            │
│  │  📋 Paste    │  📁 Upload   │  (tabs)                    │
│  └──────────────┴──────────────┘                            │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                       │   │
│  │  Paste your table from Google Sheet here...          │   │
│  │                                                       │   │
│  │  (textarea)                                           │   │
│  │                                                       │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  [🔄 Convert]                                               │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  📄 Output                                     [📋] [⬇️]    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ---                                                   │   │
│  │ name: "Converted Spec"                               │   │
│  │ version: "1.0"                                       │   │
│  │ ---                                                   │   │
│  │                                                       │   │
│  │ # Converted Spec                                     │   │
│  │ ...                                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ⚠️ Warnings (if any)                                       │
│  - Column "status" was not mapped                          │
└─────────────────────────────────────────────────────────────┘
```

### Acceptance Criteria
- [ ] Paste tab works end-to-end
- [ ] Upload tab works with sheet selection
- [ ] Copy to clipboard works
- [ ] Download as .md works
- [ ] Loading and error states displayed
- [ ] Responsive design

---

## Phase 5: Format Detection Enhancement (1 day)

### Objective
Support multiple table formats and add mapping editor.

### Tasks

- [ ] **5.1** Detect row-per-requirement format
- [ ] **5.2** Detect key-value format
- [ ] **5.3** Detect sectioned tables
- [ ] **5.4** Add confidence scoring
- [ ] **5.5** Build MappingPreview component
- [ ] **5.6** Allow manual column override

### Supported Formats

#### Format 1: Row-per-Requirement (Default)

```
| Feature | Description | Expected |
|---------|-------------|----------|
| Login   | Enter creds | Success  |
| Logout  | Click btn   | Redirect |
```

#### Format 2: Key-Value

```
| Field       | Value                    |
|-------------|--------------------------|
| Feature     | User Authentication      |
| Description | Login flow               |
| Expected    | User is authenticated    |
```

#### Format 3: Sectioned

```
| Authentication Module           |
|---------------------------------|
| Feature | Description | Expected |
| Login   | ...         | ...      |
|                                 |
| User Management Module          |
|---------------------------------|
| Feature | Description | Expected |
| Create  | ...         | ...      |
```

### Acceptance Criteria
- [ ] All 3 formats detected correctly
- [ ] Confidence score displayed
- [ ] Manual override works

---

## Phase 6: Templates System (0.5 day)

### Objective
Add multiple built-in templates for different spec types.

### Tasks

- [ ] **6.1** Create template registry
- [ ] **6.2** Add `feature-spec` template
- [ ] **6.3** Add `test-plan` template
- [ ] **6.4** Add `api-endpoint` template
- [ ] **6.5** Add template selector to UI

### Built-in Templates

| Template | Use Case | Output Style |
|----------|----------|--------------|
| `feature-spec` | Product requirements | User stories format |
| `test-plan` | QA test cases | Test steps + expected |
| `api-endpoint` | API documentation | Endpoint + request/response |

### Acceptance Criteria
- [ ] 3 templates available
- [ ] Template selector in UI
- [ ] Output changes based on template

---

## Phase 7: Production Hardening (0.5 day)

### Objective
Make tool safe for public deployment.

### Tasks

- [ ] **7.1** Add rate limiting (IP-based)
- [ ] **7.2** Add request size limits
- [ ] **7.3** Add timeouts
- [ ] **7.4** Sanitize output (escape YAML delimiters)
- [ ] **7.5** Add structured logging
- [ ] **7.6** Add health check endpoint

### Limits

| Resource | Limit |
|----------|-------|
| Paste text size | 1MB |
| XLSX file size | 10MB |
| Max rows | 5,000 |
| Max columns | 100 |
| Request timeout | 5s |
| Rate limit | 60 req/min |

### Acceptance Criteria
- [ ] Limits enforced
- [ ] Abuse prevented
- [ ] Logs structured

---

## Timeline Summary

| Phase | Description | Effort | Cumulative |
|-------|-------------|--------|------------|
| **Phase 0** | Scope & Fixtures | 0.5 day | 0.5 day |
| **Phase 1** | Backend Ingestion | 1 day | 1.5 days |
| **Phase 2** | MDFlow Renderer | 0.5 day | 2 days |
| **Phase 3** | API Endpoints | 0.5 day | 2.5 days |
| **Phase 4** | Frontend UI | 1 day | 3.5 days |
| **--- MVP Complete ---** | | | **3.5 days** |
| **Phase 5** | Format Detection | 1 day | 4.5 days |
| **Phase 6** | Templates | 0.5 day | 5 days |
| **Phase 7** | Production | 0.5 day | 5.5 days |
| **--- Full Complete ---** | | | **5.5 days** |

---

## MVP Success Criteria

- [ ] 70-80% of typical spec sheets convert cleanly
- [ ] Paste TSV from Google Sheet works
- [ ] Upload .xlsx works
- [ ] Output is valid MDFlow format
- [ ] Copy/download works
- [ ] No login required

---

## Edge Cases & Considerations

### Clipboard Paste
- Google Sheets paste = TSV format
- Handle trailing tabs, empty columns
- Multiline cells may break row splitting → warn user

### XLSX Parsing
- Multiple sheets → require selection
- Merged cells → forward-fill or warn
- Hidden rows/columns → ignore
- Large sheets → enforce limits

### Header Detection
- Non-tabular title rows above header
- Duplicate column names
- Missing required columns → warn

### Output Safety
- Escape `---` in cell content (YAML delimiter)
- Normalize bullet formatting
- Preserve row order

---

## Future Enhancements (Post-MVP)

- [ ] Share link (would need DB)
- [ ] Saved conversion history
- [ ] Custom template editor
- [ ] Real-time preview while typing
- [ ] AI-assisted column mapping
- [ ] Export to multiple formats (not just MDFlow)
