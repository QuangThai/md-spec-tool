# Phase 1d: Extract `useReviewGate`

> **Prerequisite**: Phase 1c complete  
> **Deps on other hooks**: None (self-contained after hoisting `inputSource`)  
> **File**: `frontend/hooks/useReviewGate.ts`  
> **Commit after**: Yes  
> **Build check**: `cd frontend && npm run build && npm test`

---

## What moves OUT of `MDFlowWorkbench.tsx`

| Item | Type | Original line |
|------|------|---------------|
| `requiresReviewApproval` | useState | L217 |
| `reviewApproved` | useState | L218 |
| `reviewRequiredColumns` | useState | L219 |
| `reviewedColumns` | useState | L220 |
| `latestInputSignatureRef` | useRef | L222 |
| `reviewGateReason` | derived | L249–L252 |
| `reviewRemainingCount` | useMemo | L253–L256 |
| `completeReview` | useCallback | L371–L380 |
| `handleColumnOverride` | useCallback | L382–L390 |
| input-change reset effect | useEffect | L347–L369 |

## Input params (from orchestrator)

```tsx
interface UseReviewGateParams {
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  format: string;
  mode: string;
  pasteText: string;
  file: File | null;
  isInputGsheetUrl: boolean;
}
```

## What the hook returns

```tsx
interface UseReviewGateReturn {
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
  reviewRequiredColumns: string[];
  reviewedColumns: Record<string, boolean>;
  setReviewedColumns: React.Dispatch<React.SetStateAction<Record<string, boolean>>>;
  reviewGateReason: string | undefined;
  reviewRemainingCount: number;
  completeReview: () => void;
  handleColumnOverride: (column: string, field: string) => void;
  open: (columns: string[]) => void;   // called by conversion when needs_review === true
  clear: () => void;                    // called by conversion when needs_review === false
}
```

## Implementation

Create `frontend/hooks/useReviewGate.ts`:

```tsx
import { useState, useCallback, useMemo, useRef } from "react";
import { useMDFlowActions } from "@/lib/mdflowStore";
import { canConfirmReview, countRemainingReviews } from "@/lib/reviewGate";
import { emitTelemetryEvent } from "@/lib/telemetry";
import { toast } from "@/components/ui/Toast";
import { useEffect } from "react";

interface UseReviewGateParams {
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  format: string;
  mode: string;
  pasteText: string;
  file: File | null;
  isInputGsheetUrl: boolean;
}

export function useReviewGate({ inputSource, format, mode, pasteText, file, isInputGsheetUrl }: UseReviewGateParams) {
  const { setColumnOverride } = useMDFlowActions();

  const [requiresReviewApproval, setRequiresReviewApproval] = useState(false);
  const [reviewApproved, setReviewApproved] = useState(false);
  const [reviewRequiredColumns, setReviewRequiredColumns] = useState<string[]>([]);
  const [reviewedColumns, setReviewedColumns] = useState<Record<string, boolean>>({});
  const latestInputSignatureRef = useRef("");

  const reviewGateReason =
    requiresReviewApproval && !reviewApproved
      ? "Review mapping first"
      : undefined;

  const reviewRemainingCount = useMemo(
    () => countRemainingReviews(reviewRequiredColumns, reviewedColumns),
    [reviewRequiredColumns, reviewedColumns]
  );

  // Reset review state when input changes
  useEffect(() => {
    let signature = "";
    if (mode === "paste" && pasteText.trim()) {
      signature = `${mode}:${isInputGsheetUrl ? "gsheet" : "paste"}:${pasteText.trim().slice(0, 50)}`;
    } else if ((mode === "xlsx" || mode === "tsv") && file) {
      signature = `${mode}:${file.name}:${file.size}`;
    }

    if (!signature || signature === latestInputSignatureRef.current) {
      return;
    }
    latestInputSignatureRef.current = signature;
    setRequiresReviewApproval(false);
    setReviewApproved(false);
    setReviewRequiredColumns([]);
    setReviewedColumns({});

    emitTelemetryEvent("input_provided", {
      status: "success",
      input_source: inputSource,
      template_type: format,
    });
  }, [mode, pasteText, file, isInputGsheetUrl, format, inputSource]);

  const completeReview = useCallback(() => {
    setReviewApproved(true);
    emitTelemetryEvent("review_mapping_completed", {
      status: "success",
      input_source: inputSource,
      template_type: format,
      reviewed_columns: reviewRequiredColumns.length,
    });
    toast.success("Review confirmed", "Sharing, export, and copy are now enabled");
  }, [format, inputSource, reviewRequiredColumns.length]);

  const handleColumnOverride = useCallback(
    (column: string, field: string) => {
      setColumnOverride(column, field);
      if (!requiresReviewApproval) return;
      if (!reviewRequiredColumns.includes(column)) return;
      setReviewedColumns((prev) => ({ ...prev, [column]: true }));
    },
    [setColumnOverride, requiresReviewApproval, reviewRequiredColumns]
  );

  // API for conversion hook to call
  const open = useCallback((columns: string[]) => {
    setRequiresReviewApproval(true);
    setReviewApproved(false);
    setReviewRequiredColumns(columns);
    setReviewedColumns({});
  }, []);

  const clear = useCallback(() => {
    setRequiresReviewApproval(false);
    setReviewApproved(false);
    setReviewRequiredColumns([]);
    setReviewedColumns({});
  }, []);

  return {
    requiresReviewApproval, reviewApproved,
    reviewRequiredColumns, reviewedColumns, setReviewedColumns,
    reviewGateReason, reviewRemainingCount,
    completeReview, handleColumnOverride,
    open, clear,
  };
}
```

## Wire in orchestrator

```tsx
const review = useReviewGate({ inputSource, format, mode, pasteText, file, isInputGsheetUrl });

// Replace all usages:
//   requiresReviewApproval  → review.requiresReviewApproval
//   reviewApproved          → review.reviewApproved
//   reviewRequiredColumns   → review.reviewRequiredColumns
//   reviewedColumns         → review.reviewedColumns
//   setReviewedColumns      → review.setReviewedColumns
//   reviewGateReason        → review.reviewGateReason
//   reviewRemainingCount    → review.reviewRemainingCount
//   completeReview          → review.completeReview
//   handleColumnOverride    → review.handleColumnOverride
```

## ⚠️ Cross-dependency: `handleConvert` writes review state

In `handleConvert` (L769–L790), replace the 4 raw setters with:

```tsx
// Before (L769–L774):
if (result.needs_review) {
  const uniqueColumns = buildReviewRequiredColumns(preview ?? null);
  setRequiresReviewApproval(true);
  setReviewApproved(false);
  setReviewRequiredColumns(uniqueColumns);
  setReviewedColumns({});
  // ... rest stays

// After:
if (result.needs_review) {
  const uniqueColumns = buildReviewRequiredColumns(preview ?? null);
  review.open(uniqueColumns);
  // ... rest stays

// Before (L786–L790):
} else {
  setRequiresReviewApproval(false);
  setReviewApproved(false);
  setReviewRequiredColumns([]);
  setReviewedColumns({});
}

// After:
} else {
  review.clear();
}
```

This is a **temporary** wiring — it will be cleaned up when `useWorkbenchConversion` is extracted in Phase 1h.

## Verify

- [ ] `cd frontend && npm run build` — passes
- [ ] `cd frontend && npm test` — passes
- [ ] Manual: Convert with low-confidence → review banner appears → check columns → Confirm Review → "Reviewed" badge shows
- [ ] Manual: Change input → review resets → new convert → new review if needed
- [ ] Manual: Copy/Export/Share disabled during review gate, enabled after confirm
- [ ] `canConfirmReview` logic: button disabled until all columns checked or "Mark All" used

## Commit

```
git add frontend/hooks/useReviewGate.ts frontend/components/MDFlowWorkbench.tsx
git commit -m "refactor(workbench): extract useReviewGate hook"
```
