# Phase 1g: Extract `useFileHandling`

> **Prerequisite**: Phase 1f complete  
> **Deps on other hooks**: None (receives `setLastFailedAction` from orchestrator)  
> **File**: `frontend/hooks/useFileHandling.ts` (refactor existing file)  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## ⚠️ Existing file

`frontend/hooks/useFileHandling.ts` already exists but is **not imported** by `MDFlowWorkbench.tsx`. Audit it first:
- If compatible with the spec below → refactor in place
- If incompatible → replace entirely

## What moves OUT of `MDFlowWorkbench.tsx`

| Item | Type | Original line |
|------|------|---------------|
| `dragOver` | useState | L186 |
| `getSheetsMutation` | mutation | L234 |
| `handleFileChange` | useCallback | L484–L523 |
| `onDrop` | useCallback | L1054–L1105 |
| `onDragOver` (inline) | event handler | L1574 |
| `onDragLeave` (inline) | event handler | L1578 |

## Input params (from orchestrator)

```tsx
interface UseFileHandlingParams {
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
}
```

## What the hook returns

```tsx
interface UseFileHandlingReturn {
  dragOver: boolean;
  handleFileChange: (e: React.ChangeEvent<HTMLInputElement>) => Promise<void>;
  onDrop: (e: React.DragEvent) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragLeave: () => void;
}
```

## Implementation notes

The hook internally reads/writes these **Zustand store values**:
- `mode` — read to determine file type validation in `onDrop`
- `setFile`, `setLoading`, `setError`, `setSheets`, `setSelectedSheet`, `setPreview` — write in both handlers

**Key behaviors preserved**:
1. `handleFileChange` (L484–L523): sets file, gets sheets for XLSX, skips for TSV
2. `onDrop` (L1054–L1105): validates file extension matches mode, then same logic as handleFileChange
3. Both handlers clear `lastFailedAction` on start, set `"other"` on sheet-read failure
4. `onDragOver`: `e.preventDefault()` + `setDragOver(true)`
5. `onDragLeave`: `setDragOver(false)`

## Wire in orchestrator

```tsx
const fileHandling = useFileHandling({ setLastFailedAction });

// Replace:
//   dragOver         → fileHandling.dragOver
//   handleFileChange → fileHandling.handleFileChange
//   onDrop           → fileHandling.onDrop
```

In JSX (FileUploadInput area, ~L1574):
```tsx
// Before:
onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
onDragLeave={() => setDragOver(false)}
onDrop={onDrop}

// After:
onDragOver={fileHandling.onDragOver}
onDragLeave={fileHandling.onDragLeave}
onDrop={fileHandling.onDrop}
```

**Remove from orchestrator**: `dragOver` state, `getSheetsMutation`, `handleFileChange` callback, `onDrop` callback.

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: XLSX mode → drag .xlsx file → drop → sheets load → sheet selector appears
- [ ] Manual: TSV mode → drag .tsv file → drop → file set, no sheet loading
- [ ] Manual: XLSX mode → drag .tsv file → nothing happens (wrong type)
- [ ] Manual: Click to upload → file picker → select file → works
- [ ] Manual: Drag over → visual feedback (border color change) → drag leave → reverts

## Commit

```
git add frontend/hooks/useFileHandling.ts frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): extract useFileHandling hook"
```
