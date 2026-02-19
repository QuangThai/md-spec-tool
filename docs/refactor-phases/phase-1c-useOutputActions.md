# Phase 1c: Extract `useOutputActions`

> **Prerequisite**: Phase 1b complete  
> **Deps on other hooks**: None (reads `mdflowOutput` from Zustand store directly)  
> **File**: `frontend/hooks/useOutputActions.ts`  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## What moves OUT of `MDFlowWorkbench.tsx`

| Item | Type | Original line |
|------|------|---------------|
| `copied` | useState | L185 |
| `handleCopy` | useCallback | L861–L865 |
| `handleDownload` | useCallback | L867–L877 |

> **Note**: Dead share code (L199–L208, L879–L958) was already deleted in Phase 1a.

## What the hook returns

```tsx
interface UseOutputActionsReturn {
  copied: boolean;
  handleCopy: () => void;
  handleDownload: () => void;
}
```

## Implementation

Create `frontend/hooks/useOutputActions.ts`:

```tsx
import { useState, useCallback } from "react";
import { useMDFlowStore } from "@/lib/mdflowStore";

export function useOutputActions() {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    const output = useMDFlowStore.getState().mdflowOutput;
    navigator.clipboard.writeText(output);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, []);

  const handleDownload = useCallback(() => {
    const output = useMDFlowStore.getState().mdflowOutput;
    const blob = new Blob([output], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "spec.mdflow.md";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, []);

  return { copied, handleCopy, handleDownload };
}
```

> **Performance win**: `handleCopy` and `handleDownload` now read `mdflowOutput` via `getState()` at call time instead of subscribing via `useShallow`. This follows `rerender-defer-reads` — the orchestrator no longer re-renders on every output change just because these handlers exist.

## Wire in orchestrator

```tsx
const output = useOutputActions();

// Replace:
//   copied         → output.copied
//   handleCopy     → output.handleCopy
//   handleDownload → output.handleDownload
```

## ⚠️ Check: `mdflowOutput` subscription in orchestrator

After this extraction, audit whether `mdflowOutput` is still needed in the orchestrator's `useShallow` selector. It may still be needed for:
- OutputContent rendering (`mdflowOutput ? <pre>...</pre> : <empty state>`)
- Snapshot save (`diff.saveSnapshot(mdflowOutput)`)
- Disable states (`disabled={!mdflowOutput}`)

If so, keep it in the selector. If ALL reads can be deferred to `getState()`, remove from selector for perf.

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: Convert → Copy → "Copied!" feedback → clipboard has output
- [ ] Manual: Convert → Download → `spec.mdflow.md` file downloads
- [ ] Keyboard: `Cmd+Shift+C` copies, `Cmd+Shift+E` downloads

## Commit

```
git add frontend/hooks/useOutputActions.ts frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): extract useOutputActions hook"
```
