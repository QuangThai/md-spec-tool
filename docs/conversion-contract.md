# Current Conversion Contract Documentation

This document describes the current behavior of the converter **before refactoring**.  
It serves as a reference to ensure backward compatibility during the migration to the dynamic converter.

---

## Overview

The converter transforms spreadsheet data (Google Sheets, TSV, CSV, XLSX) into Markdown format.

### Pipeline Flow

```
Input (TSV/CSV/XLSX/Paste)
        │
        ▼
┌──────────────────┐
│  Input Detection │ ← DetectInputType (Markdown vs Table)
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Parser          │ ← PasteParser / XLSXParser
│  (returns [][]string)
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Header Detection│ ← HeaderDetector.DetectHeaderRow
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Column Mapping  │ ← ColumnMapper.MapColumns (HeaderSynonyms)
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  SpecDoc Build   │ ← buildSpecDoc
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Markdown Render │ ← MDFlowRenderer.Render
└──────────────────┘
```

---

## 1. Header Detection

### Current Behavior

- **Location**: `header_detect.go` → `HeaderDetector.DetectHeaderRow`
- **Algorithm**:
  1. Checks first 5 rows
  2. Scores each row based on:
     - Matches against `HeaderSynonyms` map (+25 points per match)
     - Typical header characteristics (+5 points)
     - Penalties: markdown markers, too few non-empty cells
  3. Returns row index and confidence score (0-100)

### Edge Cases

| Case | Current Behavior |
|------|------------------|
| No recognizable headers | Returns row 0, low confidence (<50) |
| Duplicate headers (e.g., "Status", "Status") | First occurrence mapped, second ignored |
| Blank headers (e.g., "", "Item Name", "") | Blank headers are unmapped |
| Multi-line headers | Treated as single string with newlines |
| Merged cells | Not detected (depends on parser output) |

---

## 2. Column Mapping

### Current Behavior

- **Location**: `column_map.go` → `HeaderSynonyms` map (180+ entries)
- **Process**:
  1. Normalize header: lowercase, trim, collapse whitespace
  2. Lookup in `HeaderSynonyms` map
  3. Map to `CanonicalField` (e.g., `FieldItemName`, `FieldScenario`)
  4. Track unmapped columns

### Canonical Fields

| CanonicalField | Examples | Notes |
|----------------|----------|-------|
| `FieldNo` | "No", "No.", "#", "番号" | Row number |
| `FieldItemName` | "Item Name", "項目名" | Primary item identifier |
| `FieldItemType` | "Item Type", "種別" | Type/category |
| `FieldScenario` | "Scenario", "Test Case" | Test scenario |
| `FieldExpected` | "Expected", "Expected Result" | Expected outcome |
| `FieldNotes` | "Notes", "Note", "Comments" | ⚠️ **IS mapped** (since recent update) |

### Unmapped Columns

- **Behavior**: Stored in `SpecRow.Metadata` map
- **Warning**: `MAPPING_UNMAPPED_COLUMNS` warning with severity `warn`
- **Impact**: Data NOT lost, but not rendered in standard templates

---

## 3. Row Processing

### Current Behavior

- **Location**: `converter.go` → `buildSpecDoc`

### Row Skipping Rules

| Condition | Action |
|-----------|--------|
| All core fields empty (`Feature`, `Scenario`, `Instructions`, `ItemName`, `No`, `Notes`) | Skip row |
| Row has only `No` field and no meaningful content | Append to previous row's `Notes` (continuation) |
| Empty cells | Normalized to empty string |
| Cells with just "-" | Normalized to empty string |

### Multi-line Cell Handling

- **Behavior**: Preserved as-is (newlines kept)
- **Rendering**: Newlines may be converted to `<br>` in Markdown templates

### Metadata Preservation

```go
// Unmapped columns stored here:
specRow.Metadata[headerName] = cellValue
```

**Important**: Metadata is preserved but NOT rendered by default templates!

---

## 4. Empty Row Handling

### Current Contract

| Input | Current Output |
|-------|----------------|
| Completely empty row (`\n\n`) | Skipped (not in SpecDoc.Rows) |
| Row with only whitespace (`\t\t\t`) | Skipped |
| Row with only "-" markers (`-\t-\t-`) | Skipped (normalized to empty) |

---

## 5. Spec-Table Field Mapping

### Special Behavior for Spec-Table Format

When `ItemName` is present but `Feature`/`Scenario` are empty:

```go
if specRow.Feature == "" && specRow.ItemName != "" {
    specRow.Feature = specRow.ItemName
    if specRow.Scenario == "" {
        specRow.Scenario = specRow.ItemName
    }
}
```

### Field Synthesis

`Instructions` field is synthesized from spec-table fields:

```go
if specRow.Instructions == "" {
    var parts []string
    if specRow.DisplayConditions != "" {
        parts = append(parts, "Display Conditions: "+specRow.DisplayConditions)
    }
    if specRow.InputRestrictions != "" {
        parts = append(parts, "Input Restrictions: "+specRow.InputRestrictions)
    }
    if specRow.Action != "" {
        parts = append(parts, "Action: "+specRow.Action)
    }
    specRow.Instructions = strings.Join(parts, "\n")
}
```

`Expected` field synthesis:

```go
if specRow.Expected == "" && specRow.NavigationDest != "" {
    specRow.Expected = "Navigation: " + specRow.NavigationDest
}
```

---

## 6. Merged Cells Behavior

### Current Behavior

- **XLSX Parser**: Depends on `excelize` library behavior
- **TSV/CSV**: No merged cell concept (each cell is independent)
- **Contract**: Not explicitly handled; assumes parsers flatten merged cells

---

## 7. Warnings Generated

| Warning Code | Severity | Condition |
|--------------|----------|-----------|
| `INPUT_EMPTY` | `warn` | Empty input (no rows) |
| `HEADER_LOW_CONFIDENCE` | `warn` | Header detection confidence < 50 |
| `MAPPING_UNMAPPED_COLUMNS` | `warn` | Columns not in `HeaderSynonyms` |

### Warning Structure

```go
type Warning struct {
    Code     string          // e.g., "MAPPING_UNMAPPED_COLUMNS"
    Message  string          // Human-readable message
    Severity WarningSeverity // info | warn | error
    Category WarningCategory // input | header | mapping | rows | render
    Hint     string          // User-facing suggestion
    Details  map[string]any  // Structured metadata
}
```

---

## 8. Data Loss Prevention

### Current Guarantees

✅ **Preserved**:
- Unmapped columns → `SpecRow.Metadata`
- Multi-line cell content
- Special characters (Japanese, pipes, etc.)

❌ **NOT Preserved in Output** (but stored in metadata):
- Unmapped columns (not rendered by templates)
- Column order (templates decide field order)

### Critical Test Case: "Notes" Column

**Problem Identified**: In `use-cases/example-3.md` and `table.tsv`, the "Notes" column appears.

**Current Status**:
- `FieldNotes` **IS** in `HeaderSynonyms` (line 110-115 in column_map.go)
- Synonyms: "notes", "note", "comments", "comment", "remarks", "remark"
- **Should be mapped correctly**

**Verification**: See `TestSnapshot_Example3_UnmappedColumns`

---

## 9. Continuation Rows

### Special Logic

If a row:
- Has `No` field populated
- Has NO meaningful fields (Feature, Scenario, Instructions, etc.)
- Is not the first row

**Then**: Content is appended to previous row's Notes (or Expected/Instructions)

```go
func shouldAppendContinuation(rows []SpecRow, row SpecRow) bool {
    if row.No == "" { return false }
    if hasMeaningfulFields(row) { return false }
    if len(rows) == 0 { return false }
    appendContinuation(&rows[len(rows)-1], row.No)
    return true
}
```

---

## 10. Normalization Rules

### Cell Normalization

```go
func normalizeCell(value string) string {
    trimmed := strings.TrimSpace(value)
    if trimmed == "-" {
        return ""
    }
    return trimmed
}
```

### Header Normalization

```go
func (m *ColumnMapper) normalizeHeader(header string) string {
    h := strings.ToLower(header)
    h = strings.TrimSpace(h)
    h = strings.Join(strings.Fields(h), " ") // Collapse multiple spaces
    return h
}
```

---

## 11. Template Rendering

### Current Templates

- `default`: Standard test case format
- `spec-table`: Spec table format (Item Name, Item Type, etc.)
- `test-plan`: Test plan format

### Rendering Location

- **File**: `renderer.go` → `MDFlowRenderer`
- **Method**: `Render(specDoc *SpecDoc, template string) (string, error)`

---

## 12. Known Limitations

### Critical Issues to Address in Refactor

1. **Hardcoded Synonyms**: All 180+ mappings in code
2. **Data Loss Risk**: Unmapped columns not rendered (though stored)
3. **No Column Reordering**: Templates control order, not user
4. **Duplicate Headers**: Second occurrence ignored silently
5. **Blank Headers**: Completely lost (not even in metadata)
6. **No Configuration**: Adding new column types requires code change

---

## 13. Test Coverage

### Golden Tests

- `TestGoldenOutput_Example2`: 8-column table
- `TestGoldenOutput_Example3`: 9-column table (with Notes)
- `TestGoldenOutput_Example6`: Japanese content
- `TestGoldenOutput_TableTSV`: Large TSV file (71 rows)
- `TestGoldenOutput_EdgeCases`: Duplicate/blank headers, multiline cells

### Snapshot Tests

- `TestSnapshot_Example2_Metadata`: Metadata validation
- `TestSnapshot_Example3_UnmappedColumns`: Unmapped column tracking
- `TestSnapshot_TableTSV_Headers`: Header detection accuracy
- `TestSnapshot_EdgeCase_DuplicateHeaders`: Duplicate header behavior
- `TestSnapshot_EdgeCase_BlankHeaders`: Blank header behavior

---

## 14. Backward Compatibility Requirements

### Phase 1 Refactor MUST:

1. ✅ Preserve all current test outputs (golden files)
2. ✅ Generate same warnings for same inputs
3. ✅ Handle all edge cases identically
4. ✅ Maintain same normalization rules
5. ✅ Keep same continuation row logic

### Phase 1 CAN:

- Change internal data structures
- Add new fields/metadata
- Improve warnings (add details)
- Add more comprehensive logging

---

## 15. Migration Path

### Step 1: Add New Table Model (Phase 1)

- Create `Table`, `Row`, `TableMeta` types
- Parse to `Table` instead of `SpecDoc` directly
- Adapter layer to convert `Table` → `SpecDoc`

### Step 2: Maintain Compatibility

- All existing tests pass
- Golden files unchanged
- Warnings identical (or improved with more details)

### Step 3: Gradual Migration (Phase 2+)

- Introduce generic renderer
- Add config-driven mapping
- Deprecate hardcoded synonyms

---

## Appendix: Example Conversion

### Input (TSV)

```
No	Item Name	Item Type	Notes
1	Login Button	button	Primary action
2	Email Field	text	
```

### Processing

1. **Header Detection**: Row 0 detected (confidence: 100)
2. **Column Mapping**:
   - "No" → `FieldNo`
   - "Item Name" → `FieldItemName`
   - "Item Type" → `FieldItemType`
   - "Notes" → `FieldNotes`
   - Unmapped: `[]` (all mapped)
3. **Row Build**:
   ```go
   Row 1: {No: "1", ItemName: "Login Button", ItemType: "button", Notes: "Primary action"}
   Row 2: {No: "2", ItemName: "Email Field", ItemType: "text", Notes: ""}
   ```
4. **SpecDoc**:
   ```go
   {
     Title: "Converted Spec",
     Rows: [2 rows],
     Meta: {
       HeaderRow: 0,
       TotalRows: 2,
       UnmappedColumns: [],
     }
   }
   ```

---

## References

- `backend/internal/converter/converter.go`
- `backend/internal/converter/column_map.go`
- `backend/internal/converter/header_detect.go`
- `backend/internal/converter/model.go`
- `use-cases/example-*.md`
- `backend/internal/converter/golden_test.go`
