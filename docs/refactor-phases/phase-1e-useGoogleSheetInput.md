# Phase 1e: Extract `useGoogleSheetInput`

> **Prerequisite**: Phase 1d complete  
> **Deps on other hooks**: None (receives `debouncedPasteText` and `setLastFailedAction` from orchestrator)  
> **File**: `frontend/hooks/useGoogleSheetInput.ts`  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## What moves OUT of `MDFlowWorkbench.tsx`

| Item | Type | Original line |
|------|------|---------------|
| `gsheetLoading` | useState | L198 |
| `gsheetRange` | useState | L210 |
| `gsheetRangeValue` | useMemo | L262–L270 |
| `useGoogleAuth()` | hook call | L228 |
| `useGetGoogleSheetSheetsMutation()` | hook call | L235–L236 |
| Google Sheet tab loading effect | useEffect | L392–L451 |
| Google auth error toast effect | useEffect | L539–L543 |
| Google auth "just connected" effect | useEffect + `prevConnectedRef` | L546–L555 |

## Input params (from orchestrator)

```tsx
interface UseGoogleSheetInputParams {
  debouncedPasteText: string;
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
}
```

## What the hook returns

```tsx
interface UseGoogleSheetInputReturn {
  gsheetLoading: boolean;
  gsheetRange: string;
  setGsheetRange: (v: string) => void;
  gsheetRangeValue: string;              // computed: "SheetTitle!A1:F200" or ""
  googleAuth: ReturnType<typeof useGoogleAuth>;
}
```

## Implementation notes

The hook internally reads/writes these **Zustand store values** (no ownership conflict):
- `gsheetTabs`, `selectedGid` — read for `gsheetRangeValue` memo + tab loading
- `setGsheetTabs`, `setSelectedGid` — write in tab loading effect
- `setError` — write in tab loading effect

**Key behavior preserved**:
1. Tab loading effect (L392–L451): triggers on `debouncedPasteText` change, `googleAuth.connected` change, or `mode` change. Calls `fetchGoogleSheetTabs`, sets tabs/gid, clears on error.
2. `gsheetRangeValue` memo (L262–L270): prepends sheet title if range doesn't include `!`. Falls back to `""`.
3. Auth error toast (L539–L543): fires `toast.error` when `googleAuth.error` changes.
4. Auth connection tracking (L546–L555): detects connection event → clears error + lastFailedAction → shows success toast.

## ⚠️ Cross-dependency: `gsheetRangeValue` is consumed by

- `useWorkbenchPreview` → `usePreviewGoogleSheetQuery(..., gsheetRangeValue)` (extracted in Phase 1f)
- `useWorkbenchConversion` → `range: gsheetRangeValue || undefined` (extracted in Phase 1h)

Both receive it via orchestrator pass-through.

## Wire in orchestrator

```tsx
const gsheet = useGoogleSheetInput({ debouncedPasteText, setLastFailedAction });

// Replace all usages:
//   gsheetLoading    → gsheet.gsheetLoading
//   gsheetRange      → gsheet.gsheetRange
//   setGsheetRange   → gsheet.setGsheetRange
//   gsheetRangeValue → gsheet.gsheetRangeValue
//   googleAuth       → gsheet.googleAuth
```

**Remove from orchestrator**: `useGoogleAuth()`, `useGetGoogleSheetSheetsMutation()`, the 3 effects, `gsheetLoading` state, `gsheetRange` state, `gsheetRangeValue` memo.

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: Paste Google Sheets URL → tabs load → tab selector appears
- [ ] Manual: Select different tab → preview refreshes
- [ ] Manual: Enter range "A1:F200" → `gsheetRangeValue` includes sheet title prefix
- [ ] Manual: Google Auth connect → error clears → success toast → can access private sheets
- [ ] Manual: Google Auth disconnect → reconnect flow works
- [ ] Manual: Paste non-gsheet URL → tabs clear, gsheet loading doesn't fire

## Commit

```
git add frontend/hooks/useGoogleSheetInput.ts frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): extract useGoogleSheetInput hook"
```
