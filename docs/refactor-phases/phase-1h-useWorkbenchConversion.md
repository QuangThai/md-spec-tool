# Phase 1h: Extract `useWorkbenchConversion`

> **Prerequisite**: Phase 1g complete  
> **Deps on other hooks**: `gsheetPreviewSlice` from preview (1f), `reviewApi` from review (1d), `gsheetRangeValue` from gsheet (1e)  
> **File**: `frontend/hooks/useWorkbenchConversion.ts`  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## What moves OUT of `MDFlowWorkbench.tsx`

| Item | Type | Original line |
|------|------|---------------|
| `showFeedback` | useState | L197 |
| `convertPasteMutation` | mutation | L237 |
| `convertXLSXMutation` | mutation | L238 |
| `convertTSVMutation` | mutation | L239 |
| `convertGoogleSheetMutation` | mutation | L240 |
| `aiSuggestionsMutation` | mutation | L242 |
| `handleConvert` | useCallback | L648–L859 |
| `handleGetAISuggestions` | useCallback | L960–L992 |

## Input params (from orchestrator)

```tsx
interface UseWorkbenchConversionParams {
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
  gsheetPreviewSlice: {
    selectedBlockId: string | undefined;
    trustedMapping: Record<string, string>;
  };
  gsheetRangeValue: string;
  reviewApi: { open: (columns: string[]) => void; clear: () => void };
  includeMetadata: boolean;
  numberRows: boolean;
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
}
```

## What the hook returns

```tsx
interface UseWorkbenchConversionReturn {
  handleConvert: () => Promise<void>;
  handleGetAISuggestions: () => Promise<void>;
  showFeedback: boolean;
  setShowFeedback: (v: boolean) => void;
}
```

## ⚠️ Critical: Cross-dependency rewiring inside `handleConvert`

### Before (reads preview query data directly):
```tsx
const previewSelectedBlockId = previewGoogleSheetQuery.data?.selected_block_id;
const previewColumnMapping = previewGoogleSheetQuery.data?.column_mapping || {};
const previewColumnConfidence = previewGoogleSheetQuery.data?.mapping_quality?.column_confidence || {};
const trustedPreviewMapping = (previewGoogleSheetQuery.data?.confidence ?? 0) >= 50
  ? Object.fromEntries(/* filter logic */)
  : {};
```

### After (reads memoized slice from preview hook):
```tsx
const { selectedBlockId, trustedMapping } = gsheetPreviewSlice;
const effectiveColumnOverrides = { ...trustedMapping, ...columnOverrides };
```

### Before (writes review state with raw setters):
```tsx
if (result.needs_review) {
  setRequiresReviewApproval(true);
  setReviewApproved(false);
  setReviewRequiredColumns(uniqueColumns);
  setReviewedColumns({});
}
```

### After (uses review API):
```tsx
if (result.needs_review) {
  const uniqueColumns = buildReviewRequiredColumns(preview ?? null);
  reviewApi.open(uniqueColumns);
} else {
  reviewApi.clear();
}
```

## Implementation notes

The hook internally reads these **Zustand store values** (no ownership conflict):
- `mode`, `pasteText`, `file`, `selectedSheet`, `selectedGid`, `format`, `columnOverrides`, `gsheetTabs` — read in `handleConvert`
- `preview` — read for `buildReviewRequiredColumns` (Issue E)
- `aiSuggestionsLoading` — guard in `handleGetAISuggestions`
- `setLoading`, `setError`, `setResult`, `setShowPreview` — write in `handleConvert`
- `setAISuggestions`, `setAISuggestionsLoading`, `setAISuggestionsError`, `clearAISuggestions` — write in AI suggestions
- `addToHistory` from `useHistoryStore` — write in `handleConvert`

**Key behaviors preserved**:
1. `handleConvert`: mode-based dispatch (paste/gsheet/xlsx/tsv), column override merging, history recording, review gate opening, telemetry
2. `handleGetAISuggestions`: calls AI mutation with current paste text + format
3. `showFeedback`: shown 2 seconds after successful conversion

## Wire in orchestrator

```tsx
const conversion = useWorkbenchConversion({
  setLastFailedAction,
  gsheetPreviewSlice: preview.gsheetPreviewSlice,
  gsheetRangeValue: gsheet.gsheetRangeValue,
  reviewApi: review,
  includeMetadata, numberRows, inputSource,
});

// Replace:
//   handleConvert        → conversion.handleConvert
//   handleGetAISuggestions → conversion.handleGetAISuggestions
//   showFeedback         → conversion.showFeedback
//   setShowFeedback      → conversion.setShowFeedback
```

**Remove from orchestrator**: `showFeedback` state, all 5 mutation calls, both callbacks.

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: Paste data → Run → output appears → feedback modal after 2s
- [ ] Manual: Paste Google Sheet URL → select tab → Run → correct tab used
- [ ] Manual: Google Sheet with range → range sent to backend
- [ ] Manual: Google Sheet with trusted mapping → column overrides merged correctly
- [ ] Manual: XLSX upload → select sheet → Run → output
- [ ] Manual: TSV upload → Run → output
- [ ] Manual: Convert error → error banner → Retry Convert works
- [ ] Manual: Low confidence → review gate opens → review → share/copy unlocks
- [ ] Manual: AI suggestions → loads → shows in TechnicalAnalysis
- [ ] Keyboard: `Cmd+Enter` triggers convert

## Commit

```
git add frontend/hooks/useWorkbenchConversion.ts frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): extract useWorkbenchConversion hook"
```
