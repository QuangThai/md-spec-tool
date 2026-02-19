# Phase 1i: Orchestrator Cleanup

> **Prerequisite**: Phase 1h complete (all hooks extracted)  
> **File**: `frontend/components/MDFlowWorkbench.tsx`  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## Goal

After all 7 hooks are extracted, the orchestrator should contain ONLY:

1. Zustand store subscriptions
2. Hoisted shared state (`lastFailedAction`, `debouncedPasteText`, derived values)
3. UI toggle state (modals, panels)
4. Hook calls (in dependency order)
5. `useKeyboardShortcuts` wiring
6. `handleRetryFailedAction` callback
7. JSX return

## What remains in orchestrator

### Zustand store
```tsx
const { mode, pasteText, file, format, mdflowOutput, warnings, meta, loading, error,
        preview, previewLoading, showPreview, columnOverrides,
        aiSuggestions, aiSuggestionsLoading, aiSuggestionsError, aiConfigured,
} = useMDFlowStore(useShallow(...));
const { setMode, setPasteText, setFile, setFormat, setShowPreview, reset } = useMDFlowActions();
const addToHistory = useHistoryStore((s) => s.addToHistory);
const history = useHistoryStore((s) => s.history);
```

### Hoisted shared state (§1.0)
```tsx
const [lastFailedAction, setLastFailedAction] = useState(null);
const [debouncedPasteText, setDebouncedPasteText] = useState("");
// + debounce effect + isGsheetUrl + isInputGsheetUrl + inputSource
```

### UI toggle state
```tsx
const [showHistory, setShowHistory] = useState(false);
const [showValidationConfigurator, setShowValidationConfigurator] = useState(false);
const [showTemplateEditor, setShowTemplateEditor] = useState(false);
const [showCommandPalette, setShowCommandPalette] = useState(false);
const [showApiKeyInput, setShowApiKeyInput] = useState(false);
const [apiKeyDraft, setApiKeyDraft] = useState("");
const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);
const [includeMetadata, setIncludeMetadata] = useState(true);
const [numberRows, setNumberRows] = useState(false);
```

### Remaining effects
```tsx
// numberRows auto-reset when format changes (L272–L276)
useEffect(() => {
  if (!isTableFormat && numberRows) setNumberRows(false);
}, [isTableFormat, numberRows]);

// studio_opened telemetry-once (L321–L331)
useEffect(() => { /* studioOpenedTrackedRef */ }, [format, inputSource]);

// reset() on unmount (L317–L319)
useEffect(() => { return () => reset(); }, [reset]);
```

### Remaining derived values
```tsx
const isTableFormat = format === "table";
const changedOutputOptionsCount = (includeMetadata ? 0 : 1) + (numberRows ? 1 : 0);
const mappedAppError = useMemo(() => (error ? mapErrorToUserFacing(error) : null), [error]);
const { data: templateList = [] } = useMDFlowTemplatesQuery();
const openaiKey = useOpenAIKeyStore((s) => s.apiKey);
const setOpenaiKey = useOpenAIKeyStore((s) => s.setApiKey);
const clearOpenaiKey = useOpenAIKeyStore((s) => s.clearApiKey);
```

### Keyboard shortcuts (L1028–L1052)
```tsx
useKeyboardShortcuts({
  commandPalette: () => setShowCommandPalette(true),
  convert: conversion.handleConvert,
  copy: () => {
    if (mdflowOutput) { output.handleCopy(); toast.success("Copied to clipboard"); }
  },
  export: () => {
    if (mdflowOutput) { output.handleDownload(); toast.success("Downloaded spec.mdflow.md"); }
  },
  togglePreview: () => setShowPreview(!showPreview),
  showShortcuts: () => {},
  escape: () => {
    if (showCommandPalette) setShowCommandPalette(false);
    else if (showHistory) setShowHistory(false);
    else if (diff.showDiff) diff.setShowDiff(false);
    else if (showTemplateEditor) setShowTemplateEditor(false);
    else if (showValidationConfigurator) setShowValidationConfigurator(false);
  },
});
```

### handleRetryFailedAction (L1017–L1025)
```tsx
const handleRetryFailedAction = useCallback(async () => {
  if (lastFailedAction === "convert") await conversion.handleConvert();
  else if (lastFailedAction === "preview") await preview.handleRetryPreview();
}, [conversion.handleConvert, preview.handleRetryPreview, lastFailedAction]);
```

## Cleanup tasks

- [ ] Remove all `useState`/`useEffect`/`useCallback`/`useMemo`/`useRef` that were extracted to hooks
- [ ] Remove all mutation imports that moved to hooks (`useConvertPasteMutation`, etc.)
- [ ] Remove unused icon imports from lucide-react
- [ ] Remove unused utility imports (`buildReviewRequiredColumns`, `canConfirmReview`, `countRemainingReviews` if fully encapsulated)
- [ ] Update `useMDFlowActions` destructuring — remove actions that hooks now call internally
- [ ] Update `useMDFlowStore` `useShallow` selector — remove state values only used by hooks (if any)
- [ ] Verify hook call order matches dependency chain:
  ```tsx
  const diff = useDiffSnapshots();
  const output = useOutputActions();
  const review = useReviewGate({ ... });
  const gsheet = useGoogleSheetInput({ ... });
  const preview = useWorkbenchPreview({ ..., gsheetRangeValue: gsheet.gsheetRangeValue, ... });
  const fileHandling = useFileHandling({ ... });
  const conversion = useWorkbenchConversion({ ..., gsheetPreviewSlice: preview.gsheetPreviewSlice, reviewApi: review, gsheetRangeValue: gsheet.gsheetRangeValue, ... });
  ```

## Expected result

`MDFlowWorkbench.tsx` should be **~250 lines** (down from 2160):
- ~30 lines: imports
- ~40 lines: store subscriptions + hoisted state
- ~15 lines: UI toggle state + derived values
- ~20 lines: hook calls
- ~30 lines: keyboard shortcuts + handleRetryFailedAction + remaining effects
- ~115 lines: JSX return (will be further reduced in Phase 2)

## Verify (full regression)

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] **Full manual test** (all items from §4.2 in master plan):
  - [ ] Paste → Preview → Convert → Copy/Export/Share
  - [ ] XLSX upload → Sheet select → Convert
  - [ ] TSV upload → Convert
  - [ ] Google Sheets URL → Tab select → Range → Convert
  - [ ] Google Sheets URL → change tab → preview refreshes → convert uses new tab
  - [ ] Google Auth connect/disconnect
  - [ ] Diff snapshots: Save A → Save B → Compare → Clear
  - [ ] Review gate full flow
  - [ ] Command palette (`Cmd+K`)
  - [ ] All keyboard shortcuts
  - [ ] Escape priority chain
  - [ ] History modal, Template editor, Validation configurator
  - [ ] API key set/clear
  - [ ] Error retry (preview + convert)
  - [ ] File drag & drop

## Commit

```
git add frontend/components/MDFlowWorkbench.tsx frontend/hooks/
git commit -m "refactor(workbench): Phase 1 complete — orchestrator cleanup, all hooks extracted"
```
