# Phase 1b: Extract `useDiffSnapshots`

> **Prerequisite**: Phase 1a complete  
> **Deps on other hooks**: None (self-contained)  
> **File**: `frontend/hooks/useDiffSnapshots.ts`  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## What moves OUT of `MDFlowWorkbench.tsx`

| Item | Type | Original line |
|------|------|---------------|
| `showDiff` | useState | L187 |
| `snapshotA` | useState | L189 |
| `snapshotB` | useState | L190 |
| `currentDiff` | useState | L191 |
| `useBodyScrollLock(showDiff)` | hook call | L230 |
| `diffMDFlowMutation` | mutation | L241 |

## What the hook returns

```tsx
interface UseDiffSnapshotsReturn {
  showDiff: boolean;
  setShowDiff: (v: boolean) => void;
  snapshotA: string;
  snapshotB: string;
  currentDiff: any;
  saveSnapshot: (output: string) => void;
  compareSnapshots: () => Promise<void>;
  clearSnapshots: () => void;
}
```

## Implementation

Create `frontend/hooks/useDiffSnapshots.ts`:

```tsx
import { useState, useCallback } from "react";
import { useDiffMDFlowMutation } from "@/lib/mdflowQueries";
import { useBodyScrollLock } from "@/lib/useBodyScrollLock";
import { toast } from "@/components/ui/Toast";

export function useDiffSnapshots() {
  const [showDiff, setShowDiff] = useState(false);
  const [snapshotA, setSnapshotA] = useState("");
  const [snapshotB, setSnapshotB] = useState("");
  const [currentDiff, setCurrentDiff] = useState<any>(null);

  useBodyScrollLock(showDiff);

  const diffMDFlowMutation = useDiffMDFlowMutation();

  const saveSnapshot = useCallback((output: string) => {
    if (!output) return;
    if (!snapshotA) {
      setSnapshotA(output);
      toast.success("Version A saved", "Run with new input, save again");
    } else if (!snapshotB) {
      setSnapshotB(output);
      toast.success("Version B saved", "Ready to compare");
    } else {
      setSnapshotB(output);
      toast.success("Version B updated");
    }
  }, [snapshotA, snapshotB]);

  const compareSnapshots = useCallback(async () => {
    const diff = await diffMDFlowMutation.mutateAsync({
      before: snapshotA,
      after: snapshotB,
    });
    setCurrentDiff(diff);
    setShowDiff(true);
    if (diff && !diff.hunks?.length && diff.added_lines === 0 && diff.removed_lines === 0) {
      toast.info("No changes detected", "Versions may be identical");
    }
  }, [snapshotA, snapshotB, diffMDFlowMutation]);

  const clearSnapshots = useCallback(() => {
    setSnapshotA("");
    setSnapshotB("");
    toast.success("Snapshots cleared");
  }, []);

  return {
    showDiff, setShowDiff,
    snapshotA, snapshotB, currentDiff,
    saveSnapshot, compareSnapshots, clearSnapshots,
  };
}
```

## Wire in orchestrator

```tsx
// Replace the 4 useState + useBodyScrollLock + diffMDFlowMutation with:
const diff = useDiffSnapshots();

// Then replace all usages:
//   showDiff        → diff.showDiff
//   setShowDiff     → diff.setShowDiff
//   snapshotA       → diff.snapshotA
//   snapshotB       → diff.snapshotB
//   currentDiff     → diff.currentDiff
//   inline save     → diff.saveSnapshot(mdflowOutput)
//   inline compare  → diff.compareSnapshots()
//   inline clear    → diff.clearSnapshots()
```

## JSX changes (OutputToolbar area, ~L1855–L1947)

Replace inline snapshot logic with hook calls:

- Save button `onClick` → `diff.saveSnapshot(mdflowOutput)`
- Compare button `onClick` → `diff.compareSnapshots()`
- Clear button `onClick` → `diff.clearSnapshots()`

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: Save A → convert with different input → Save B → Compare → diff modal shows → Clear → badges gone
- [ ] Escape closes diff modal
- [ ] Body scroll locked when diff modal open

## Commit

```
git add frontend/hooks/useDiffSnapshots.ts frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): extract useDiffSnapshots hook"
```
