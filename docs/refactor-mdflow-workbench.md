# Refactor Plan: MDFlowWorkbench.tsx

> **Status**: 🔴 Not Started  
> **Target**: 2160 lines → ~250 lines (orchestrator) + focused sub-components  
> **File**: `frontend/components/MDFlowWorkbench.tsx`

---

## Current Problems

| Problem | Detail |
|---------|--------|
| 32 `useState` in 1 component | Re-render cascade on every state change |
| ~15 `useEffect` | Preview sync, telemetry, gsheet tabs — mixed concerns |
| ~10 `useCallback` handlers | `handleConvert` alone is 212 lines |
| ~1050 lines JSX | Left panel, right panel, 5 modals inline |
| Existing hooks unused | `useConversionFlow`, `useFileHandling`, `usePreviewManagement`, `useExportFunctionality`, `useUIState` exist in `hooks/` but are not imported |

---

## ⚠️ Cross-Dependency Map (MUST READ BEFORE IMPLEMENTING)

After deep audit, these are the critical data flows that cross hook boundaries.
Ignoring any of these **will break behavior**.

### Issue A: `lastFailedAction` is written by 5 different concerns

| Writer | Value set | Location |
|--------|-----------|----------|
| Preview (activePreviewError effect) | `"preview"` | L310–L314 |
| Google Sheet tab loading | `null` or `"preview"` | L410–L431 |
| Google auth connect | `null` | L547–L555 |
| File handling (handleFileChange, onDrop) | `null` or `"other"` | L484–L523, L1054–L1105 |
| Conversion (handleConvert) | `null` or `"convert"` | L648–L826 |

**Decision**: Hoist `lastFailedAction` + `setLastFailedAction` to **orchestrator level** and pass down to hooks that need it.

### Issue B: `debouncedPasteText` drives both Preview AND Google Sheet hooks

- `debouncedPasteText` is set by debounce effect (L334–L345)
- `isGsheetUrl = isGoogleSheetsURL(debouncedPasteText.trim())` (L244) — used by both preview and gsheet
- Google Sheet tab loading effect depends on `debouncedPasteText` (L392–L451)
- 4 preview queries depend on `debouncedPasteText` (L278–L296)

**Decision**: Hoist `debouncedPasteText` + its debounce effect to **orchestrator level**. Compute `isGsheetUrl` at orchestrator and pass to both preview and gsheet hooks.

### Issue C: `previewGoogleSheetQuery.data` is needed by Conversion

`handleConvert` reads these slices (L671–L699):
- `.selected_block_id` → sent as `selectedBlockId`
- `.column_mapping` → used to build `trustedPreviewMapping`
- `.confidence` → threshold gate for trusted mapping
- `.mapping_quality.column_confidence` → per-column confidence filter

**Decision**: Preview hook returns `gsheetPreviewSlice` = `{ selectedBlockId, trustedMapping }` (memoized). Conversion hook accepts it as input.

### Issue D: Conversion writes Review Gate state

`handleConvert` (L769–L790) calls:
- `setRequiresReviewApproval(true/false)`
- `setReviewApproved(false)`
- `setReviewRequiredColumns(columns)`
- `setReviewedColumns({})`

**Decision**: Review Gate hook exposes API: `review.open(columns)` / `review.clear()`. Conversion hook calls this API instead of raw setters.

### Issue E: `preview` store value used for review gating in Conversion

`handleConvert` (L770):
```ts
const uniqueColumns = buildReviewRequiredColumns(preview ?? null);
```

**Decision**: Conversion reads `preview` from Zustand store directly (it's already a store value), no hook ownership conflict.

### Issue F: `gsheetRangeValue` needed by both Preview query AND Conversion

- Preview: `usePreviewGoogleSheetQuery(..., gsheetRangeValue)` (L295)
- Conversion: `range: gsheetRangeValue || undefined` (L696)

**Decision**: `gsheetRangeValue` computed in Google Sheet hook, returned and passed to both preview hook and conversion hook by orchestrator.

### Issue G: Keyboard shortcuts reference state from multiple hooks

`useKeyboardShortcuts` (L1028–L1052) depends on:
- `handleConvert` (conversion), `handleCopy`/`handleDownload` (output actions)
- `showDiff` + `setShowDiff` (diff snapshots)
- `showPreview` (store), `showHistory`, `showCommandPalette`, etc. (UI toggles)

**Decision**: Keyboard shortcuts stay in orchestrator. All hooks return their handlers/state needed for wiring.

### Issue H: Missing refs in plan

- `previewStartedAtRef` (L223) — used by preview telemetry effect (L557–L646)
- `previewAttemptRef` (L224) — used by preview telemetry effect (L557–L646)

**Decision**: These refs must move INTO useWorkbenchPreview hook alongside the telemetry effect.

### Issue I: `isGsheetUrl` vs `isInputGsheetUrl` are NOT interchangeable

- `isGsheetUrl` = `isGoogleSheetsURL(debouncedPasteText.trim())` → controls query enablement, tab loading
- `isInputGsheetUrl` = `isGoogleSheetsURL(pasteText.trim())` → controls `inputSource` derivation, JSX rendering

**Decision**: Both computed at orchestrator. `isGsheetUrl` uses hoisted `debouncedPasteText`. `isInputGsheetUrl` uses store `pasteText`. Never unify them.

---

## Phase 1: Extract Custom Hooks (Logic Layer)

> **Goal**: Move all business logic out of the component.  
> **Vercel rule**: `rerender-defer-reads` — only subscribe to state actually needed for render.

### 1.0 Hoisted shared state (stays in orchestrator, passed as params to hooks)

These values are used across multiple hooks and MUST live at orchestrator level:

```tsx
// Orchestrator computes these, passes to hooks
const [lastFailedAction, setLastFailedAction] = useState<"preview" | "convert" | "other" | null>(null);  // L216
const [debouncedPasteText, setDebouncedPasteText] = useState("");  // L209

// Debounce effect (L334–L345) stays in orchestrator
useEffect(() => {
  if (mode !== "paste") { setDebouncedPasteText(""); return; }
  const timer = setTimeout(() => setDebouncedPasteText(pasteText), 500);
  return () => clearTimeout(timer);
}, [pasteText, mode]);

// Derived values computed from hoisted state
const isGsheetUrl = isGoogleSheetsURL(debouncedPasteText.trim());        // L244
const isInputGsheetUrl = isGoogleSheetsURL(pasteText.trim());            // L245
const inputSource = mode === "paste" ? (isInputGsheetUrl ? "gsheet" : "paste") : mode;  // L246–L247
```

### 1.1 `useWorkbenchConversion`

**Inputs** (params from orchestrator):
- `setLastFailedAction` (hoisted)
- `gsheetPreviewSlice` (from preview hook: `{ selectedBlockId, trustedMapping }`)
- `gsheetRangeValue` (from gsheet hook)
- `reviewApi` (from review hook: `{ open, clear }`)
- `includeMetadata`, `numberRows` (from UI state)
- `inputSource` (hoisted derived)

**Owns**:
- [ ] `handleConvert` callback (L648–L859)
- [ ] `handleGetAISuggestions` callback (L960–L992)
- [ ] Own state: `showFeedback` (L197)
- [ ] Own telemetry: `convert_started` (L653), `convert_succeeded` (L792), `convert_failed` (L817)

**Returns**: `{ handleConvert, handleGetAISuggestions, showFeedback, setShowFeedback }`

**Source lines**: L197, L237–L240, L242, L648–L859, L960–L992

### 1.2 `useWorkbenchPreview`

**Inputs** (params from orchestrator):
- `debouncedPasteText` (hoisted)
- `isGsheetUrl` (hoisted derived)
- `gsheetRangeValue` (from gsheet hook)
- `setLastFailedAction` (hoisted)
- `inputSource`, `format` (for telemetry)

**Owns**:
- [ ] 4 preview queries (L278–L296)
- [ ] `activePreviewError` derived memo (L297–L309)
- [ ] `activePreviewError` → `setLastFailedAction("preview")` effect (L310–L314)
- [ ] Paste preview sync effect (L453–L482)
- [ ] XLSX preview sync effect (L525–L530)
- [ ] TSV preview sync effect (L532–L537)
- [ ] Preview loading + telemetry effect (L557–L646)
- [ ] Own refs: `previewStartedAtRef` (L223), `previewAttemptRef` (L224)
- [ ] `handleRetryPreview` callback (L994–L1015)

**Returns**: `{ previewQueries, activePreviewError, handleRetryPreview, gsheetPreviewSlice }`

Where `gsheetPreviewSlice` is a memoized object:
```tsx
const gsheetPreviewSlice = useMemo(() => ({
  selectedBlockId: previewGoogleSheetQuery.data?.selected_block_id,
  trustedMapping: /* computed from confidence + column_mapping + column_confidence */,
}), [previewGoogleSheetQuery.data?.selected_block_id, ...]);
```

**Source lines**: L223–L224, L278–L314, L453–L482, L525–L537, L557–L646, L994–L1015

### 1.3 `useGoogleSheetInput`

**Inputs** (params from orchestrator):
- `debouncedPasteText` (hoisted)
- `setLastFailedAction` (hoisted)

**Owns**:
- [ ] Google Sheet tab loading effect (L392–L451)
- [ ] Own state: `gsheetLoading` (L198), `gsheetRange` (L210)
- [ ] Derived: `gsheetRangeValue` memo (L262–L270)
- [ ] Google auth error toast effect (L539–L543)
- [ ] Google auth connection tracking effect (L547–L555, includes `prevConnectedRef` declared at L546)
- [ ] `useGoogleAuth()` hook call (L228)
- [ ] `useGetGoogleSheetSheetsMutation()` hook call (L235–L236)

**Returns**: `{ gsheetLoading, gsheetRange, setGsheetRange, gsheetRangeValue, googleAuth }`

**Source lines**: L198, L210, L228, L235–L236, L262–L270, L392–L451, L539–L555

### 1.4 `useReviewGate`

**Inputs** (params from orchestrator):
- `inputSource`, `format` (for telemetry)
- `mode`, `pasteText`, `file`, `isInputGsheetUrl` (for input-change reset effect deps)

**Owns**:
- [ ] Own state: `requiresReviewApproval` (L217), `reviewApproved` (L218), `reviewRequiredColumns` (L219), `reviewedColumns` (L220)
- [ ] Derived: `reviewGateReason` (L249–L252), `reviewRemainingCount` memo (L253–L256)
- [ ] Handler: `completeReview` (L371–L380)
- [ ] Handler: `handleColumnOverride` (L382–L390)
- [ ] Input change reset effect (L347–L369, includes `latestInputSignatureRef` at L222)

**Returns**: `{ state, reviewGateReason, reviewRemainingCount, completeReview, handleColumnOverride, open(columns), clear(), reviewedColumns, setReviewedColumns }`

Where `open(columns)` and `clear()` encapsulate the 4-setter patterns used by handleConvert:
```tsx
// Called by handleConvert when result.needs_review === true
open: (columns: string[]) => {
  setRequiresReviewApproval(true);
  setReviewApproved(false);
  setReviewRequiredColumns(columns);
  setReviewedColumns({});
}
// Called by handleConvert when result.needs_review === false
clear: () => {
  setRequiresReviewApproval(false);
  setReviewApproved(false);
  setReviewRequiredColumns([]);
  setReviewedColumns({});
}
```

**Source lines**: L217–L222, L249–L256, L347–L390

### 1.5 `useDiffSnapshots`

**Inputs**: none (self-contained)

**Owns**:
- [ ] Own state: `snapshotA` (L189), `snapshotB` (L190), `currentDiff` (L191), `showDiff` (L187)
- [ ] `useBodyScrollLock(showDiff)` (L230)
- [ ] `diffMDFlowMutation` (L241)

**Returns**: `{ showDiff, setShowDiff, snapshotA, snapshotB, currentDiff, saveSnapshot, compareSnapshots, clearSnapshots }`

Where handlers encapsulate the inline JSX logic:
```tsx
saveSnapshot: (output: string) => { /* L1882–L1894 logic */ }
compareSnapshots: async () => { /* L1910–L1928 logic */ }
clearSnapshots: () => { setSnapshotA(""); setSnapshotB(""); }
```

**Source lines**: L187, L189–L191, L230, L241
**JSX source**: L1855–L1947 (snapshot badges + save/compare/clear buttons)

### 1.6 `useOutputActions`

**Inputs**: none (reads `mdflowOutput` from store directly)

**Owns**:
- [ ] `handleCopy` + `copied` state (L185, L861–L865)
- [ ] `handleDownload` (L867–L877)

**⚠️ DEAD CODE — DELETE, do NOT extract**:
The following share state and logic exist in the source but are **never wired to any JSX**.
The actual UI uses `<ShareButton />` (L1974–L1978), a self-contained component that manages its own share logic.
These must be **deleted** during refactor, not extracted into a hook:
- `creatingShare` (L199), `shareTitle` (L200), `shareSlug` (L201), `shareVisibility` (L202–L204), `shareAllowComments` (L205), `showShareOptions` (L206), `shareSlugError` (L207), `shareOptionsRef` (L208)
- `handleCreateShare` callback (L879–L946)
- Click-outside effect for share options (L948–L958)

**Returns**: `{ copied, handleCopy, handleDownload }`

**Source lines**: L185, L861–L877  
**Dead code to delete**: L199–L208, L879–L958

### 1.7 `useFileHandling` (refactor existing `hooks/useFileHandling.ts`)

**Inputs** (params from orchestrator):
- `setLastFailedAction` (hoisted)

**Owns**:
- [ ] `handleFileChange` callback (L484–L523)
- [ ] `onDrop` callback (L1054–L1105)
- [ ] Own state: `dragOver` (L186)
- [ ] `getSheetsMutation` (L234)
- [ ] Drag event handlers: `onDragOver` (sets dragOver true), `onDragLeave` (sets dragOver false) — used inline in JSX at L1574–L1578

**Returns**: `{ dragOver, handleFileChange, onDrop, onDragOver, onDragLeave }`

**Source lines**: L186, L234, L484–L523, L1054–L1105

### 1.8 Remaining state & effects (stays in orchestrator)

These items are UI-only toggles or cross-cutting wiring:

- [ ] `showHistory` (L192), `showValidationConfigurator` (L193–L194), `showTemplateEditor` (L195), `showCommandPalette` (L196)
- [ ] `showApiKeyInput` (L211), `apiKeyDraft` (L212)
- [ ] `showAdvancedOptions` (L213), `includeMetadata` (L214), `numberRows` (L215)
- [ ] `numberRows` auto-reset effect when `isTableFormat` changes (L272–L276) — needs `setNumberRows` accessible at orchestrator
- [ ] `studioOpenedTrackedRef` telemetry-once effect (L221, L321–L331)
- [ ] `reset()` on unmount effect (L317–L319)
- [ ] `useKeyboardShortcuts` wiring (L1028–L1052) — depends on handlers from hooks 1.1, 1.5, 1.6 + UI toggles
- [ ] `isTableFormat` derived value (L248)
- [ ] `changedOutputOptionsCount` derived value (L261)
- [ ] `mappedAppError` memo (L257–L260)
- [ ] `templates` / `useMDFlowTemplatesQuery` (L232–L233)
- [ ] `addToHistory` / `history` from `useHistoryStore` (L182–L183)
- [ ] `openaiKey` / `setOpenaiKey` / `clearOpenaiKey` from `useOpenAIKeyStore` (L225–L227)
- [ ] `handleRetryFailedAction` callback (L1017–L1025) — depends on `handleConvert` (hook 1.1) + `handleRetryPreview` (hook 1.2) + `lastFailedAction` (hoisted)

**Source lines**: L182–L183, L192–L196, L211–L215, L221, L225–L227, L232–L233, L248, L257–L261, L272–L276, L317–L331, L1017–L1025, L1028–L1052

### Phase 1 Hook Wiring Diagram

```
MDFlowWorkbench (orchestrator)
│
├── Hoisted state: lastFailedAction, debouncedPasteText
├── Hoisted derived: isGsheetUrl, isInputGsheetUrl, inputSource
│
├── useGoogleSheetInput(debouncedPasteText, setLastFailedAction)
│   └── returns: gsheetRangeValue, gsheetRange, setGsheetRange, gsheetLoading, googleAuth
│
├── useWorkbenchPreview(debouncedPasteText, isGsheetUrl, gsheetRangeValue, setLastFailedAction, inputSource, format)
│   └── returns: handleRetryPreview, activePreviewError, gsheetPreviewSlice
│
├── useReviewGate(inputSource, format, mode, pasteText, file, isInputGsheetUrl)
│   └── returns: reviewGateReason, reviewRemainingCount, completeReview, handleColumnOverride, open(), clear(), state...
│
├── useWorkbenchConversion(setLastFailedAction, gsheetPreviewSlice, gsheetRangeValue, reviewApi, includeMetadata, numberRows, inputSource)
│   └── returns: handleConvert, handleGetAISuggestions, showFeedback
│
├── useFileHandling(setLastFailedAction)
│   └── returns: dragOver, handleFileChange, onDrop, onDragOver, onDragLeave
│
├── useDiffSnapshots()
│   └── returns: showDiff, setShowDiff, snapshotA, snapshotB, currentDiff, saveSnapshot, compareSnapshots, clearSnapshots
│
├── useOutputActions()
│   └── returns: copied, handleCopy, handleDownload
│
└── handleRetryFailedAction = f(handleConvert, handleRetryPreview, lastFailedAction)
```

### Phase 1 Execution Order (hooks must be extracted in this sequence)

1. **1.5 useDiffSnapshots** — self-contained, zero deps on others
2. **1.6 useOutputActions** — self-contained, reads store directly
3. **1.4 useReviewGate** — self-contained after hoisting inputSource
4. **1.3 useGoogleSheetInput** — depends on hoisted debouncedPasteText
5. **1.2 useWorkbenchPreview** — depends on gsheetRangeValue from 1.3
6. **1.7 useFileHandling** — depends on hoisted setLastFailedAction
7. **1.1 useWorkbenchConversion** — depends on preview slice from 1.2, review API from 1.4, gsheet range from 1.3
8. **1.8 Orchestrator cleanup** — wire keyboard shortcuts, handleRetryFailedAction

### Phase 1 Checklist

- [ ] Hoist `lastFailedAction` + `debouncedPasteText` + derived values to orchestrator FIRST
- [ ] Each hook has its own file in `frontend/hooks/`
- [ ] Verify `isGsheetUrl` always uses `debouncedPasteText` (NOT `pasteText`)
- [ ] Verify `isInputGsheetUrl` always uses `pasteText` (NOT `debouncedPasteText`)
- [ ] Verify `previewStartedAtRef` (L223) and `previewAttemptRef` (L224) moved into useWorkbenchPreview
- [ ] Verify `useBodyScrollLock` called exactly once (in useDiffSnapshots only)
- [ ] `npm run build` passes after each hook extraction
- [ ] `npm test` passes
- [ ] No behavior change (pure refactor)
- [ ] Manual test after EACH hook: paste gsheet URL → tabs load → preview → convert → review gate → share/copy

---

## Phase 2: Extract UI Components (Presentation Layer)

> **Goal**: Split 1050-line JSX into composable, memoizable components.  
> **Vercel rules**: `rerender-memo`, `bundle-dynamic-imports`.

### 2.1 `SourcePanel` → `components/workbench/SourcePanel.tsx`

Contains the entire left column (L1123–L1796). Composes:

- [ ] **`SourcePanelHeader`** (~70 lines, L1131–L1197) — Mode toggle (Paste/Excel/TSV), quick action buttons (API key, Template Editor, Validation), `QuotaStatus`
- [ ] **`ApiKeyPanel`** (~70 lines, L1199–L1267) — Collapsible OpenAI API key input with save/clear
- [ ] **`ErrorBanner`** (~40 lines, L1271–L1307) — Error display with mapped error, request_id, retry button
- [ ] **`ReviewGateBanner`** (~75 lines, L1309–L1383) — Review required UI with column checkboxes, Mark All, Confirm Review
- [ ] **`PasteInput`** (~170 lines, L1385–L1561) — Google Sheet status bar, tab selector, range input, auth banner, preview table, textarea
- [ ] **`FileUploadInput`** (~160 lines, L1562–L1726) — Drag & drop zone, file info, sheet selector (xlsx), file preview table
- [ ] **`WorkbenchFooter`** (~65 lines, L1730–L1794) — Template selector + Run button with disabled logic

### 2.2 `OutputPanel` → `components/workbench/OutputPanel.tsx`

Contains the entire right column (L1799–L2040). Composes:

- [ ] **`OutputToolbar`** (~155 lines, L1808–L1991) — Copy, snapshot badges (A✓/B✓), Save snapshot, Compare, Clear snapshots, Export, ShareButton, History buttons + review status badges
- [ ] **`OutputContent`** (~25 lines, L1993–L2019) — Loading skeleton / output pre / empty state
- [ ] **Stats footer** (~15 lines, L2021–L2037) — `TechnicalAnalysis` component (already extracted)

### 2.3 `DiffModal` → `components/workbench/DiffModal.tsx`

- [ ] Wrap existing `DiffViewer` in modal chrome (L2043–L2080)

### File Structure

```
frontend/components/workbench/
├── SourcePanel.tsx
├── SourcePanelHeader.tsx
├── ApiKeyPanel.tsx
├── PasteInput.tsx
├── FileUploadInput.tsx
├── ErrorBanner.tsx
├── ReviewGateBanner.tsx
├── WorkbenchFooter.tsx
├── OutputPanel.tsx
├── OutputToolbar.tsx
├── OutputContent.tsx
└── DiffModal.tsx
```

### Phase 2 Checklist

- [ ] Each component in its own file under `components/workbench/`
- [ ] Props are typed with explicit interfaces
- [ ] No business logic in components — only presentation + event forwarding
- [ ] `npm run build` passes after each component extraction
- [ ] `npm test` passes

---

## Phase 3: Performance Optimization

> **Goal**: Apply Vercel React Best Practices after structure is clean.

| Rule | Action | Status |
|------|--------|--------|
| `rerender-defer-reads` | Handlers (`handleCopy`, `handleDownload`) read from store directly in callback instead of subscribing to `mdflowOutput` | ⬜ |
| `rerender-derived-state` | Subscribe to derived booleans (`isGsheetUrl`, `inputSource`) instead of raw `pasteText` | ⬜ |
| `rerender-memo` | Wrap `OutputPanel`, `SourcePanel`, `OutputToolbar` with `React.memo()` | ⬜ |
| `bundle-dynamic-imports` | `ApiKeyPanel`, `ReviewGateBanner`, `DiffModal` → `next/dynamic` (only rendered when toggled) | ⬜ |
| `rerender-functional-setstate` | Audit all `setState` calls — use functional form where prev-dependent | ⬜ |
| `rendering-conditional-render` | Convert `&&` → ternary `condition ? <X /> : null` for conditional JSX | ⬜ |
| `js-hoist-regexp` | ~~Hoist `/^[a-z0-9]+(?:-[a-z0-9]+)*$/`~~ — dead code in `handleCreateShare`, will be deleted | ✅ N/A |
| `rerender-memo-with-default-value` | Hoist default non-primitive props (empty arrays, objects) to module scope | ⬜ |

### Phase 3 Checklist

- [ ] No unnecessary re-renders confirmed via React DevTools Profiler
- [ ] Bundle size delta checked (no regression)
- [ ] `npm run build` passes
- [ ] `npm test` passes

---

## Phase 4: Integration & Verification

### 4.1 Final `MDFlowWorkbench.tsx` (~250 lines)

```tsx
export default function MDFlowWorkbench() {
  // ── Zustand store ──
  const { mode, pasteText, file, format, ... } = useMDFlowStore(useShallow(...));
  const { setMode, setPasteText, ... } = useMDFlowActions();
  const addToHistory = useHistoryStore((s) => s.addToHistory);
  const history = useHistoryStore((s) => s.history);

  // ── Hoisted shared state (cross-hook) ──
  const [lastFailedAction, setLastFailedAction] = useState(null);
  const [debouncedPasteText, setDebouncedPasteText] = useState("");
  useEffect(() => { /* debounce effect L334–L345 */ }, [pasteText, mode]);
  const isGsheetUrl = isGoogleSheetsURL(debouncedPasteText.trim());
  const isInputGsheetUrl = isGoogleSheetsURL(pasteText.trim());
  const inputSource = mode === "paste" ? (isInputGsheetUrl ? "gsheet" : "paste") : mode;

  // ── UI toggles (orchestrator-owned) ──
  const [showHistory, setShowHistory] = useState(false);
  const [showTemplateEditor, setShowTemplateEditor] = useState(false);
  // ... etc

  // ── Hooks (in dependency order) ──
  const diff = useDiffSnapshots();
  const output = useOutputActions(); // { copied, handleCopy, handleDownload } — share handled by <ShareButton />
  const review = useReviewGate({ inputSource, format, mode, pasteText, file, isInputGsheetUrl });
  const gsheet = useGoogleSheetInput({ debouncedPasteText, setLastFailedAction });
  const preview = useWorkbenchPreview({ debouncedPasteText, isGsheetUrl, gsheetRangeValue: gsheet.gsheetRangeValue, setLastFailedAction, inputSource, format });
  const fileHandling = useFileHandling({ setLastFailedAction });
  const conversion = useWorkbenchConversion({ setLastFailedAction, gsheetPreviewSlice: preview.gsheetPreviewSlice, gsheetRangeValue: gsheet.gsheetRangeValue, reviewApi: review, includeMetadata, numberRows, inputSource });

  // ── Keyboard shortcuts ──
  useKeyboardShortcuts({ ... });

  // ── handleRetryFailedAction (orchestrator-level) ──
  const handleRetryFailedAction = useCallback(async () => {
    if (lastFailedAction === "convert") await conversion.handleConvert();
    else if (lastFailedAction === "preview") await preview.handleRetryPreview();
  }, [conversion.handleConvert, preview.handleRetryPreview, lastFailedAction]);

  return (
    <motion.div ...>
      <OnboardingTour />
      <div className="grid grid-cols-1 lg:grid-cols-2 ...">
        <SourcePanel ... />
        <OutputPanel ... />
      </div>
      <DiffModal ... />
      <HistoryModal ... />
      <ValidationConfigurator ... />
      <TemplateEditor ... />
      <CommandPalette ... />
      <ConversionFeedback ... />
      <ToastContainer />
    </motion.div>
  );
}
```

### 4.2 Manual Testing Checklist

- [ ] Paste → Preview → Convert → Copy/Export/Share
- [ ] XLSX upload → Sheet select → Convert
- [ ] TSV upload → Convert
- [ ] Google Sheets URL → Tab select → Range → Convert
- [ ] Google Sheets URL → change tab → preview refreshes → convert uses new tab
- [ ] Google Auth connect → error clears → toast shows → convert works with private sheet
- [ ] Google Auth disconnect → re-paste URL → public access only
- [ ] Diff snapshots: Save A → Save B → Compare → Clear
- [ ] Review gate flow (low confidence mapping → review columns → confirm → share unlocks)
- [ ] Review gate: change input → review resets → new convert → new review if needed
- [ ] Command palette (`Cmd+K`) — all actions work
- [ ] Keyboard shortcuts (`Cmd+Enter`, `Cmd+Shift+C`, `Cmd+Shift+E`, `Cmd+Shift+P`)
- [ ] Keyboard `Escape` closes: command palette → history → diff → template editor → validation (in priority order)
- [ ] History modal — select previous conversion
- [ ] Template editor — open/close
- [ ] Validation configurator — open/close
- [ ] API key — set/clear
- [ ] Error → Retry preview (banner button)
- [ ] Error → Retry convert (banner button)
- [ ] Onboarding tour triggers correctly
- [ ] Responsive layout (mobile, tablet, desktop)
- [ ] File drag & drop: XLSX file on xlsx mode, TSV file on tsv mode
- [ ] File drag & drop: wrong file type → no action

---

## Execution Order

```
Phase 1 (hooks)  ──commit──▶  Phase 2 (components)  ──commit──▶  Phase 3 (perf)  ──commit──▶  Phase 4 (verify)
     │                              │                                  │
     └─ test after each hook        └─ test after each component       └─ profiler check
```

Within Phase 1, extract in this order:
```
1. Hoist shared state (lastFailedAction, debouncedPasteText, derived values)
2. useDiffSnapshots (zero deps)
3. useOutputActions (zero deps)
4. useReviewGate (self-contained + inputSource param)
5. useGoogleSheetInput (needs debouncedPasteText)
6. useWorkbenchPreview (needs gsheetRangeValue from 5)
7. useFileHandling (needs setLastFailedAction)
8. useWorkbenchConversion (needs slices from 4, 5, 6)
9. Orchestrator cleanup (keyboard shortcuts, handleRetryFailedAction)
```

Each phase should be a **separate commit** for easy rollback.

---

## Notes

- Existing hooks in `frontend/hooks/` (`useConversionFlow`, `useFileHandling`, `usePreviewManagement`, `useExportFunctionality`, `useUIState`) should be audited — reuse if compatible, replace if not.
- The `stagger` animation config (lines 97–106) stays in `MDFlowWorkbench.tsx` as module-level constant.
- Dynamic imports already exist for `DiffViewer`, `TemplateEditor`, `ValidationConfigurator` — keep them.
- `preview` (store value) is read by conversion for `buildReviewRequiredColumns` — no ownership conflict since it's a Zustand store value, not hook-local state.
- `gsheetTabs`, `selectedGid`, `columnOverrides` are Zustand store values — readable by any hook via `useMDFlowStore`, no ownership conflict.
- **⚠️ Dead code discovered**: `handleCreateShare` + 8 share-related state variables (L199–L208, L879–L958) are **never referenced in JSX**. The actual UI uses `<ShareButton />` component (L1974–L1978) which self-manages share logic. Delete this dead code in Phase 1 step 3 (useOutputActions extraction).
