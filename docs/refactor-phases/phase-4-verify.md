# Phase 4: Integration & Verification

> **Prerequisite**: Phase 3 complete  
> **Commit after**: Yes (final commit)  
> **Build check**: `cd frontend && npm run build && npm test`

---

## Goal

Full regression test. Confirm zero behavior change across the entire refactor.

---

## 4.1 Build & Test

- [ ] `cd frontend && npm run build` — passes with zero warnings
- [ ] `cd frontend && npm test` — all tests pass
- [ ] No TypeScript errors (`npx tsc --noEmit`)

## 4.2 File size audit

```
MDFlowWorkbench.tsx: 2160 lines → ~250 lines ✅ or ✗
hooks/useDiffSnapshots.ts:    ~60 lines
hooks/useOutputActions.ts:    ~30 lines
hooks/useReviewGate.ts:       ~120 lines
hooks/useGoogleSheetInput.ts: ~100 lines
hooks/useWorkbenchPreview.ts: ~180 lines
hooks/useFileHandling.ts:     ~80 lines
hooks/useWorkbenchConversion.ts: ~200 lines
components/workbench/:        ~12 files
```

## 4.3 Manual Testing Checklist

### Core flows
- [ ] Paste → Preview → Convert → Copy/Export/Share
- [ ] XLSX upload → Sheet select → Convert
- [ ] TSV upload → Convert

### Google Sheets
- [ ] Google Sheets URL → Tab select → Range → Convert
- [ ] Google Sheets URL → change tab → preview refreshes → convert uses new tab
- [ ] Google Auth connect → error clears → toast shows → convert works with private sheet
- [ ] Google Auth disconnect → re-paste URL → public access only

### Diff snapshots
- [ ] Save A → Save B → Compare → diff modal shows
- [ ] Clear snapshots → badges gone
- [ ] Body scroll locked when diff modal open

### Review gate
- [ ] Low confidence mapping → review columns → confirm → share unlocks
- [ ] Review gate: change input → review resets → new convert → new review if needed
- [ ] Copy/Export/Share disabled during review gate, enabled after confirm

### Keyboard & command palette
- [ ] Command palette (`Cmd+K`) — all actions work
- [ ] `Cmd+Enter` → convert
- [ ] `Cmd+Shift+C` → copy
- [ ] `Cmd+Shift+E` → download
- [ ] `Cmd+Shift+P` → toggle preview
- [ ] `Escape` closes: command palette → history → diff → template editor → validation (in priority order)

### Modals & panels
- [ ] History modal — select previous conversion → output restores
- [ ] Template editor — open/close
- [ ] Validation configurator — open/close
- [ ] API key — set → "AI features enabled" → clear

### Error handling
- [ ] Preview error → error banner → Retry Preview button works
- [ ] Convert error → error banner → Retry Convert button works
- [ ] Error shows `request_id` when available

### File handling
- [ ] XLSX mode: drag .xlsx → drop → sheets load
- [ ] TSV mode: drag .tsv → drop → file set
- [ ] Wrong file type → no action
- [ ] Click upload → file picker → works
- [ ] Drag over → visual feedback → drag leave → reverts

### Misc
- [ ] Onboarding tour triggers correctly
- [ ] Responsive layout (mobile, tablet, desktop)
- [ ] `<ShareButton />` still works independently
- [ ] AI suggestions load in TechnicalAnalysis
- [ ] Conversion feedback modal appears 2s after convert

## 4.4 Performance validation

- [ ] React DevTools Profiler: typing in paste area doesn't re-render OutputPanel
- [ ] React DevTools Profiler: snapshot save doesn't re-render SourcePanel
- [ ] Bundle size comparison: no significant regression from pre-refactor

## Commit

```
git add -A
git commit -m "refactor(workbench): Phase 4 complete — full verification passed"
```
