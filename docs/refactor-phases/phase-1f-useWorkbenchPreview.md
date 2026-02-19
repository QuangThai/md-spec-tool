# Phase 1f: Extract `useWorkbenchPreview`

> **Prerequisite**: Phase 1e complete  
> **Deps on other hooks**: `gsheetRangeValue` from `useGoogleSheetInput` (via orchestrator)  
> **File**: `frontend/hooks/useWorkbenchPreview.ts`  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## What moves OUT of `MDFlowWorkbench.tsx`

| Item | Type | Original line |
|------|------|---------------|
| `previewStartedAtRef` | useRef | L223 |
| `previewAttemptRef` | useRef | L224 |
| `previewPasteQuery` | query | L278–L282 |
| `previewTSVQuery` | query | L283 |
| `previewXLSXQuery` | query | L284–L289 |
| `previewGoogleSheetQuery` | query | L290–L296 |
| `activePreviewError` | useMemo | L297–L309 |
| error → setLastFailedAction("preview") | useEffect | L310–L314 |
| Paste preview sync effect | useEffect | L453–L482 |
| XLSX preview sync effect | useEffect | L525–L530 |
| TSV preview sync effect | useEffect | L532–L537 |
| Preview loading + telemetry effect | useEffect | L557–L646 |
| `handleRetryPreview` | useCallback | L994–L1015 |

## Input params (from orchestrator)

```tsx
interface UseWorkbenchPreviewParams {
  debouncedPasteText: string;
  isGsheetUrl: boolean;
  gsheetRangeValue: string;
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  format: string;
}
```

## What the hook returns

```tsx
interface UseWorkbenchPreviewReturn {
  activePreviewError: Error | null;
  handleRetryPreview: () => Promise<void>;
  gsheetPreviewSlice: {
    selectedBlockId: string | undefined;
    trustedMapping: Record<string, string>;
  };
}
```

## ⚠️ Critical: `gsheetPreviewSlice` computation

This memoized slice is consumed by `useWorkbenchConversion` (Phase 1h). It encapsulates the preview data that `handleConvert` currently reads at L671–L699:

```tsx
const gsheetPreviewSlice = useMemo(() => {
  const data = previewGoogleSheetQuery.data;
  if (!data) return { selectedBlockId: undefined, trustedMapping: {} };

  const previewColumnMapping = data.column_mapping || {};
  const previewColumnConfidence = data.mapping_quality?.column_confidence || {};
  const trustedMapping =
    (data.confidence ?? 0) >= 50
      ? Object.fromEntries(
          Object.entries(previewColumnMapping).filter(([header, mappedField]) => {
            if (!header || !mappedField) return false;
            const score = previewColumnConfidence[header];
            return typeof score !== "number" || score >= 0.7;
          })
        )
      : {};

  return {
    selectedBlockId: data.selected_block_id,
    trustedMapping,
  };
}, [
  previewGoogleSheetQuery.data?.selected_block_id,
  previewGoogleSheetQuery.data?.column_mapping,
  previewGoogleSheetQuery.data?.confidence,
  previewGoogleSheetQuery.data?.mapping_quality?.column_confidence,
]);
```

## Implementation notes

The hook internally reads/writes these **Zustand store values** (no ownership conflict):
- `mode`, `file`, `selectedSheet` — read for query enablement + preview sync
- `setPreview`, `setShowPreview`, `setPreviewLoading` — write in sync effects

**Key behaviors preserved**:
1. 4 preview queries use `debouncedPasteText` (NOT raw `pasteText`) for paste/gsheet
2. `activePreviewError` → `setLastFailedAction("preview")` effect
3. Paste sync effect (L453–L482): sets preview from query data, clears when empty
4. XLSX/TSV sync effects: simple data → store sync
5. Preview loading + telemetry effect (L557–L646): tracks `isFetching` across all 4 queries, emits start/success/fail telemetry with duration

## Wire in orchestrator

```tsx
const preview = useWorkbenchPreview({
  debouncedPasteText, isGsheetUrl,
  gsheetRangeValue: gsheet.gsheetRangeValue,
  setLastFailedAction, inputSource, format,
});

// Replace:
//   activePreviewError   → preview.activePreviewError
//   handleRetryPreview   → preview.handleRetryPreview
```

**Remove from orchestrator**: all 4 query calls, `activePreviewError` memo, the 5 effects, the 2 refs, `handleRetryPreview` callback.

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: Paste tabular data → preview table appears
- [ ] Manual: Paste Google Sheets URL → gsheet preview loads with correct tab
- [ ] Manual: Upload XLSX → preview table appears
- [ ] Manual: Upload TSV → preview table appears
- [ ] Manual: Preview error → retry button works
- [ ] Manual: "Analyzing..." spinner shows during preview load
- [ ] Manual: Change input → old preview clears, new preview loads

## Commit

```
git add frontend/hooks/useWorkbenchPreview.ts frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): extract useWorkbenchPreview hook"
```
