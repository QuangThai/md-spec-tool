# MDFlowWorkbench Refactor — Phase Breakdown

> Master plan: [`../refactor-mdflow-workbench.md`](../refactor-mdflow-workbench.md)

## Execution order

```
Phase 1a ─► 1b ─► 1c ─► 1d ─► 1e ─► 1f ─► 1g ─► 1h ─► 1i ─► Phase 2 ─► Phase 3 ─► Phase 4
 hoist    diff  output review gsheet preview file  convert orch   UI        perf       verify
```

Each step = 1 commit. Build + test after every step.

---

## Phase 1: Extract Custom Hooks

| Step | File | Hook | Deps | Lines moved |
|------|------|------|------|-------------|
| [1a](phase-1a-hoist-and-cleanup.md) | — | Delete dead code + hoist shared state | None | ~80 deleted |
| [1b](phase-1b-useDiffSnapshots.md) | `hooks/useDiffSnapshots.ts` | `useDiffSnapshots` | None | ~60 |
| [1c](phase-1c-useOutputActions.md) | `hooks/useOutputActions.ts` | `useOutputActions` | None | ~30 |
| [1d](phase-1d-useReviewGate.md) | `hooks/useReviewGate.ts` | `useReviewGate` | `inputSource` | ~120 |
| [1e](phase-1e-useGoogleSheetInput.md) | `hooks/useGoogleSheetInput.ts` | `useGoogleSheetInput` | `debouncedPasteText` | ~100 |
| [1f](phase-1f-useWorkbenchPreview.md) | `hooks/useWorkbenchPreview.ts` | `useWorkbenchPreview` | `gsheetRangeValue` from 1e | ~180 |
| [1g](phase-1g-useFileHandling.md) | `hooks/useFileHandling.ts` | `useFileHandling` | `setLastFailedAction` | ~80 |
| [1h](phase-1h-useWorkbenchConversion.md) | `hooks/useWorkbenchConversion.ts` | `useWorkbenchConversion` | slices from 1d, 1e, 1f | ~200 |
| [1i](phase-1i-orchestrator-cleanup.md) | `MDFlowWorkbench.tsx` | Final wiring | All hooks | ~1900 removed |

## Phase 2–4

| Phase | File | What |
|-------|------|------|
| [2](phase-2-ui-components.md) | `components/workbench/*.tsx` | Extract 12 UI components |
| [3](phase-3-performance.md) | Various | `React.memo`, `next/dynamic`, defer reads |
| [4](phase-4-verify.md) | — | Full regression test |

## Result

```
Before: MDFlowWorkbench.tsx = 2160 lines (1 file)
After:  MDFlowWorkbench.tsx = ~250 lines + 7 hooks + 12 components
```
