# OpenAI Integration Plan for MD-Spec-Tool

> **Version**: 1.0  
> **Created**: 2026-02-05  
> **Status**: Planning

---

## Table of Contents

1. [Overview](#overview)
2. [Goals & Non-Goals](#goals--non-goals)
3. [Key Decisions](#key-decisions)
4. [Architecture](#architecture)
5. [Phase 1: Foundation - AI Service Layer](#phase-1-foundation---ai-service-layer)
6. [Phase 2: Intelligent Column Mapping](#phase-2-intelligent-column-mapping)
7. [Phase 3: Simplify Renderers](#phase-3-simplify-renderers)
8. [Phase 4: Smart Paste Processing](#phase-4-smart-paste-processing)
9. [Phase 5: Frontend Updates](#phase-5-frontend-updates)
10. [Phase 6: Testing, Documentation & Cleanup](#phase-6-testing-documentation--cleanup)
11. [Migration Strategy](#migration-strategy)
12. [Risks & Mitigations](#risks--mitigations)
13. [Appendix: Output Format Specifications](#appendix-output-format-specifications)

---

## Overview

### Current Problems

1. **Hardcoded column mapping**: `column_map.go` contains ~180+ hardcoded mappings in `HeaderSynonyms`
2. **Inflexible**: Adding new columns requires code changes
3. **Too many templates**: Current system has many complex formats; only 2 are needed: `spec` and `table`
4. **Basic paste handling**: Needs smarter processing with LLM

### Solution

Integrate OpenAI GPT-4o-mini with **Structured Outputs** to:
- Automatically detect and map columns based on context
- Convert flexibly without hardcoding
- Support new columns automatically
- Process paste content intelligently

---

## Goals & Non-Goals

### Goals

- Replace hardcoded `HeaderSynonyms` with LLM-based column mapping
- Simplify output to only 2 formats: `spec` and `table`
- Smart paste processing that understands content context
- Cache responses to minimize API costs
- Maintain backward compatibility during transition (1-2 weeks)

### Non-Goals

- Support for multiple LLM providers (future consideration)
- Real-time streaming responses (not needed for this use case)
- User-configurable AI toggle (always use AI)
- Rate limiting (not needed currently)

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| OpenAI Model | GPT-4o-mini | Cost-effective, good structured output support |
| Output Formats | `spec`, `table` | Simplified, covers all use cases |
| API Key Management | Environment variable `OPENAI_API_KEY` | Server-side, secure |
| Caching | Yes - SHA256 hash based | Reduce API costs for identical inputs |
| AI Folder | `backend/internal/ai/` | Provider-agnostic naming for future flexibility |
| Fallback Logic | None - Trust OpenAI 100% | Structured Outputs guarantee schema compliance |
| Backward Compatibility | 1-2 weeks transition | Deprecation warnings, then remove old endpoints |
| AI Toggle | Always use AI | No user option to disable |

---

## Architecture

### Current Flow

```
Input (Paste/XLSX/TSV)
    ↓
[Parser] → CellMatrix
    ↓
[HeaderSynonyms] ← Hardcoded mapping (180+ entries)
    ↓
[Multiple Renderers] → Various formats
    ↓
Markdown Output
```

### New Flow

```
Input (Paste/XLSX/TSV)
    ↓
[Parser] → CellMatrix (headers + rows)
    ↓
[AI Service] ← OpenAI Structured Outputs
    ├─ Column Mapping (replaces HeaderSynonyms)
    ├─ Content Understanding
    └─ Format Optimization
    ↓
[Cache Layer] ← SHA256 hash lookup
    ↓
[Renderer] → spec | table
    ↓
Markdown Output
```

### New Folder Structure

```
backend/internal/ai/
├── client.go           # OpenAI client wrapper
├── schemas.go          # Structured output schemas (Go structs)
├── cache.go            # Response caching (SHA256 hash)
├── column_mapper.go    # LLM column mapping
├── paste_processor.go  # Smart paste handling
└── prompts.go          # System prompts & few-shot examples

backend/internal/converter/
├── spec_renderer.go    # New: Spec format renderer
├── table_renderer.go   # New: Table format renderer
├── converter.go        # Modified: Integrate AI service
└── renderer_factory.go # Modified: Only 2 formats
```

---

## Phase 1: Foundation - AI Service Layer

**Goal**: Create the base AI service layer for OpenAI integration

### Tasks

#### 1.1 Create `backend/internal/ai/client.go`

```go
package ai

import (
    "context"
    "github.com/openai/openai-go/v3"
)

type Client struct {
    client *openai.Client
    model  string
}

type Config struct {
    APIKey string
    Model  string // default: gpt-4o-mini
}

func NewClient(cfg Config) (*Client, error)
func (c *Client) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
```

**Features**:
- Wrapper for OpenAI Go SDK (`github.com/openai/openai-go/v3`)
- Config from environment variable `OPENAI_API_KEY`
- Retry logic with exponential backoff (3 attempts)
- Request timeout (30 seconds default)

#### 1.2 Create `backend/internal/ai/schemas.go`

```go
package ai

// ColumnMappingResult is the structured output for column mapping
type ColumnMappingResult struct {
    Mappings []ColumnMapping `json:"mappings" jsonschema_description:"List of column mappings"`
}

type ColumnMapping struct {
    OriginalHeader string `json:"original_header" jsonschema_description:"Original column header from input"`
    CanonicalField string `json:"canonical_field" jsonschema_description:"Mapped canonical field name or empty if unmapped"`
    Reasoning      string `json:"reasoning" jsonschema_description:"Brief explanation for the mapping decision"`
}

// PasteAnalysis is the structured output for paste processing
type PasteAnalysis struct {
    InputType           string   `json:"input_type" jsonschema:"enum=table,enum=markdown,enum=mixed,enum=unknown"`
    DetectedFormat      string   `json:"detected_format" jsonschema:"enum=csv,enum=tsv,enum=markdown_table,enum=free_text"`
    Headers             []string `json:"headers,omitempty"`
    SuggestedOutputFormat string `json:"suggested_output_format" jsonschema:"enum=spec,enum=table"`
}

// ConversionResult is the structured output for full conversion
type ConversionResult struct {
    Title       string              `json:"title"`
    Categories  []CategoryResult    `json:"categories"`
    Summary     SummaryResult       `json:"summary"`
}
```

#### 1.3 Create `backend/internal/ai/cache.go`

```go
package ai

import (
    "crypto/sha256"
    "sync"
    "time"
)

type Cache struct {
    mu      sync.RWMutex
    entries map[string]cacheEntry
    ttl     time.Duration
}

type cacheEntry struct {
    value     interface{}
    expiresAt time.Time
}

func NewCache(ttl time.Duration) *Cache
func (c *Cache) Get(key string) (interface{}, bool)
func (c *Cache) Set(key string, value interface{})
func (c *Cache) HashKey(input string) string // SHA256
```

**Features**:
- In-memory cache with configurable TTL (default: 1 hour)
- Thread-safe with RWMutex
- SHA256 hash for cache keys
- Automatic expiration cleanup

#### 1.4 Create `backend/internal/ai/prompts.go`

```go
package ai

const SystemPromptColumnMapping = `You are an expert at analyzing spreadsheet data...`
const SystemPromptPasteAnalysis = `You are an expert at understanding pasted content...`

// Few-shot examples for better accuracy
var ColumnMappingExamples = []Example{...}
```

### Dependencies

```go
// Add to go.mod
require (
    github.com/openai/openai-go/v3 latest
    github.com/invopop/jsonschema v0.2.0
)
```

### Deliverables

- [ ] OpenAI client with Structured Outputs support
- [ ] Schema definitions with JSON schema tags
- [ ] In-memory caching layer
- [ ] Prompt templates
- [ ] Unit tests for all components

### Estimated Effort: 2-3 days

---

## Phase 2: Intelligent Column Mapping

**Goal**: Replace hardcoded `HeaderSynonyms` with LLM-based mapping

### Tasks

#### 2.1 Create `backend/internal/ai/column_mapper.go`

```go
package ai

type ColumnMapperService struct {
    client *Client
    cache  *Cache
}

type MapColumnsRequest struct {
    Headers    []string   `json:"headers"`
    SampleRows [][]string `json:"sample_rows"` // 3-5 rows for context
    Format     string     `json:"format"`      // "spec" | "table"
}

func (s *ColumnMapperService) MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error)
```

**Canonical Fields** (simplified from current 20+ to essential ones):

For `spec` format:
- `id`, `title`, `description`, `category`, `type`, `priority`, `status`
- `precondition`, `steps`, `expected_result`, `notes`

For `table` format:
- All columns preserved as-is (no mapping needed)

#### 2.2 Prompt Engineering

System prompt structure:
```
You are an expert at analyzing spreadsheet headers and mapping them to canonical fields.

CANONICAL FIELDS:
- id: Unique identifier (TC-001, #1, ID, etc.)
- title: Main title or name of the item
- description: Detailed description or instructions
- category: Group/feature/module name
- type: Type classification (functional, UI, API, etc.)
- priority: Priority level (high, medium, low, P1, P2, etc.)
- status: Current status (active, draft, deprecated, etc.)
- precondition: Prerequisites or setup requirements
- steps: Step-by-step instructions
- expected_result: Expected outcome or acceptance criteria
- notes: Additional notes or comments

RULES:
1. Map each header to the most appropriate canonical field
2. If a header doesn't match any canonical field, set canonical_field to empty string
3. Consider the sample data to understand context
4. Support multiple languages (English, Japanese, Vietnamese, etc.)

OUTPUT FORMAT:
Return a JSON object matching the ColumnMappingResult schema.
```

#### 2.3 Integration with Converter

Modify `backend/internal/converter/converter.go`:

```go
type Converter struct {
    aiService *ai.ColumnMapperService
    // ... other fields
}

func (c *Converter) Convert(ctx context.Context, input ConvertInput) (*ConvertOutput, error) {
    // 1. Parse input to CellMatrix
    matrix := c.parse(input)
    
    // 2. Use AI for column mapping
    mappingResult, err := c.aiService.MapColumns(ctx, ai.MapColumnsRequest{
        Headers:    matrix.Headers(),
        SampleRows: matrix.SampleRows(5),
        Format:     input.Format,
    })
    
    // 3. Apply mapping and render
    // ...
}
```

### Deliverables

- [ ] Column mapper service with OpenAI integration
- [ ] Optimized prompts with few-shot examples
- [ ] Integration with converter
- [ ] Tests with various header formats (EN, JP, VN)

### Estimated Effort: 2-3 days

---

## Phase 3: Simplify Renderers

**Goal**: Reduce to only 2 output formats: `spec` and `table`

### Tasks

#### 3.1 Create `backend/internal/converter/spec_renderer.go`

Output format (see [Appendix](#appendix-output-format-specifications) for full spec):

```markdown
---
title: User Authentication Module
generated: 2026-02-05T10:30:00Z
source: xlsx
total_items: 15
categories: 3
---

# User Authentication Module

## Summary

| Metric | Value |
|--------|-------|
| Total Items | 15 |
| Categories | 3 |
| High Priority | 5 |

## Specifications

### Login Feature

#### TC-001: Valid Login

| Field | Value |
|-------|-------|
| ID | TC-001 |
| Type | Functional |
| Priority | High |
| Status | Active |

**Description:**
User should be able to login with valid credentials.

**Precondition:**
User account exists in the system.

**Steps:**
1. Navigate to login page
2. Enter valid username
3. Enter valid password
4. Click login button

**Expected Result:**
User is redirected to dashboard.

---

#### TC-002: Invalid Login
...
```

#### 3.2 Create `backend/internal/converter/table_renderer.go`

Simple markdown table output (preserves original structure):

```markdown
| ID | Feature | Description | Priority | Status |
|----|---------|-------------|----------|--------|
| TC-001 | Login | Valid login test | High | Active |
| TC-002 | Login | Invalid login test | Medium | Active |
```

#### 3.3 Update `backend/internal/converter/renderer_factory.go`

```go
func NewRenderer(format string) (Renderer, error) {
    switch format {
    case "spec", "":
        return NewSpecRenderer(), nil
    case "table":
        return NewTableRenderer(), nil
    default:
        return nil, fmt.Errorf("unknown format: %s (supported: spec, table)", format)
    }
}
```

#### 3.4 Deprecate Old Renderers

Files to remove after transition:
- `row_cards_renderer.go`
- `testspec_renderer.go`
- `generic_renderer.go`
- Complex templates in `renderer.go`
- YAML templates in `backend/templates/`

### Deliverables

- [ ] New `spec_renderer.go`
- [ ] New `table_renderer.go`
- [ ] Updated `renderer_factory.go`
- [ ] Deprecation warnings for old formats
- [ ] Tests for both renderers

### Estimated Effort: 1-2 days

---

## Phase 4: Smart Paste Processing

**Goal**: Intelligent paste content handling with LLM

### Tasks

#### 4.1 Create `backend/internal/ai/paste_processor.go`

```go
package ai

type PasteProcessorService struct {
    client *Client
    cache  *Cache
}

type ProcessPasteRequest struct {
    Content string `json:"content"`
}

type ProcessPasteResponse struct {
    Analysis    PasteAnalysis `json:"analysis"`
    CleanedData [][]string    `json:"cleaned_data,omitempty"`
    Headers     []string      `json:"headers,omitempty"`
}

func (s *PasteProcessorService) Process(ctx context.Context, req ProcessPasteRequest) (*ProcessPasteResponse, error)
```

#### 4.2 Detection Pipeline

1. **Format Detection**: CSV, TSV, markdown table, free text
2. **Header Detection**: Identify header row automatically
3. **Data Cleaning**: Normalize whitespace, handle empty cells
4. **Format Suggestion**: Recommend `spec` or `table` based on content

#### 4.3 Integration

Update `backend/internal/http/handlers/mdflow_handler.go`:

```go
func (h *Handler) ConvertPaste(c *gin.Context) {
    // 1. Get paste content
    content := req.PasteText
    
    // 2. Process with AI
    processed, err := h.aiService.ProcessPaste(ctx, ai.ProcessPasteRequest{
        Content: content,
    })
    
    // 3. Convert using detected format
    result, err := h.converter.Convert(ctx, converter.ConvertInput{
        Matrix: processed.CleanedData,
        Headers: processed.Headers,
        Format: processed.Analysis.SuggestedOutputFormat,
    })
    
    // 4. Return response
}
```

### Deliverables

- [ ] Paste processor service
- [ ] Format detection logic
- [ ] Integration with handlers
- [ ] Tests for various paste formats

### Estimated Effort: 2-3 days

---

## Phase 5: Frontend Updates

**Goal**: Update UI to reflect backend changes

### Tasks

#### 5.1 Update `frontend/lib/mdflowApi.ts`

```typescript
export interface ConversionRequest {
  paste_text?: string;
  file?: File;
  format: 'spec' | 'table';
  sheet_name?: string;
}

export interface ConversionResponse {
  markdown: string;
  warnings: Warning[];
  meta: {
    title: string;
    total_items: number;
    categories: number;
    column_mappings: ColumnMapping[];
  };
}

export interface ColumnMapping {
  original_header: string;
  canonical_field: string;
  reasoning: string;
}
```

#### 5.2 Update `frontend/lib/types.ts`

```typescript
export type OutputFormat = 'spec' | 'table';

// Remove old format types
// - 'test_spec'
// - 'generic_table'
// - 'row_cards'
// - etc.
```

#### 5.3 Update UI Components

1. **Format Selector**: Only 2 options
   ```tsx
   <Select value={format} onChange={setFormat}>
     <Option value="spec">Spec Document</Option>
     <Option value="table">Simple Table</Option>
   </Select>
   ```

2. **Loading States**: Add AI processing indicator
   ```tsx
   {isProcessing && <Spinner label="AI processing..." />}
   ```

3. **Column Mapping Display**: Show AI decisions
   ```tsx
   <Collapsible title="Column Mappings">
     {mappings.map(m => (
       <div key={m.original_header}>
         {m.original_header} → {m.canonical_field}
         <small>{m.reasoning}</small>
       </div>
     ))}
   </Collapsible>
   ```

### Deliverables

- [ ] Updated API types
- [ ] Simplified format selector
- [ ] AI processing indicators
- [ ] Column mapping display
- [ ] Error handling for AI failures

### Estimated Effort: 1-2 days

---

## Phase 6: Testing, Documentation & Cleanup

**Goal**: Ensure quality and maintainability

### Tasks

#### 6.1 Testing

1. **Unit Tests**
   - AI client (with mocked OpenAI)
   - Cache layer
   - Column mapper
   - Paste processor
   - Both renderers

2. **Integration Tests**
   - Full conversion flow
   - API endpoints
   - Cache hit/miss scenarios

3. **E2E Tests**
   - Paste conversion
   - XLSX upload
   - Various input formats

4. **Edge Cases**
   - Empty input
   - Malformed data
   - Large files (1000+ rows)
   - Unicode/multi-language headers

#### 6.2 Documentation

1. Update `README.md`
2. Update `AGENTS.md`
3. API documentation
4. Architecture decision record (this document)

#### 6.3 Cleanup

1. Remove deprecated files:
   - `backend/internal/converter/row_cards_renderer.go`
   - `backend/internal/converter/testspec_renderer.go`
   - `backend/internal/converter/generic_renderer.go`
   - `backend/templates/*.yaml`

2. Remove `HeaderSynonyms` from `column_map.go` (keep file for utilities)

3. Remove old format handling from:
   - `renderer.go`
   - `renderer_factory.go`
   - `template_registry.go`

### Deliverables

- [ ] Comprehensive test suite (>80% coverage for new code)
- [ ] Updated documentation
- [ ] Clean codebase without deprecated code

### Estimated Effort: 2-3 days

---

## Migration Strategy

### Timeline

```
Week 1:
├── Phase 1: Foundation (Days 1-3)
├── Phase 2: Column Mapping (Days 3-5)
└── Phase 3: Renderers (Days 5-6)

Week 2:
├── Phase 4: Paste Processing (Days 1-3)
├── Phase 5: Frontend (Days 3-4)
└── Phase 6: Testing & Cleanup (Days 4-7)
```

### Backward Compatibility

1. **During transition (1-2 weeks)**:
   - Old API endpoints continue to work
   - Old formats return deprecation warning in response
   - New `format` parameter accepts old values but maps to new ones:
     ```
     "test_spec" → "spec"
     "generic_table" → "table"
     "row_cards" → "spec"
     ```

2. **After transition**:
   - Remove old format support
   - Return error for deprecated formats
   - Remove deprecated code

### API Changes

| Old | New | Notes |
|-----|-----|-------|
| `format=test_spec` | `format=spec` | Deprecated, auto-mapped |
| `format=generic_table` | `format=table` | Deprecated, auto-mapped |
| `format=row_cards` | `format=spec` | Deprecated, auto-mapped |
| `format=test_spec_v1` | `format=spec` | Deprecated, auto-mapped |

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| OpenAI API costs | Medium | Medium | Caching, GPT-4o-mini (cheaper) |
| API latency | Medium | Low | Caching, loading states |
| LLM inaccuracy | Low | Medium | Good prompts, few-shot examples |
| Breaking changes | High | Medium | Transition period, deprecation warnings |
| OpenAI downtime | Low | High | Graceful error handling, retry logic |

---

## Appendix: Output Format Specifications

### Spec Format

```markdown
---
title: [Auto-detected or sheet name]
generated: [ISO 8601 timestamp]
source: [paste | xlsx | tsv | gsheet]
total_items: [number]
categories: [number]
---

# [Title]

## Summary

| Metric | Value |
|--------|-------|
| Total Items | [number] |
| Categories | [number] |
| High Priority | [number] |
| Medium Priority | [number] |
| Low Priority | [number] |

## Specifications

### [Category 1]

#### [Item Title]

| Field | Value |
|-------|-------|
| ID | [value] |
| Type | [value] |
| Priority | [value] |
| Status | [value] |

**Description:**
[Long text content, can be multi-line]

**Precondition:**
[Prerequisites if any]

**Steps:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Expected Result:**
[Expected outcome]

**Notes:**
[Additional notes if any]

---

### [Category 2]
...
```

### Table Format

```markdown
| [Header 1] | [Header 2] | [Header 3] | ... |
|------------|------------|------------|-----|
| [Value] | [Value] | [Value] | ... |
| [Value] | [Value] | [Value] | ... |
```

**Notes for Table Format:**
- Preserves original column order
- No column mapping applied
- Headers used as-is
- Empty cells rendered as empty
- Long content truncated with `...` (configurable)

---

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-05 | Initial plan |
