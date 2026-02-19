import { buildReviewRequiredColumns, countRemainingReviews } from "@/lib/reviewGate";
import { emitTelemetryEvent } from "@/lib/telemetry";
import { toast } from "@/components/ui/Toast";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";

interface UseReviewGateParams {
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  format: string;
  mode: "paste" | "xlsx" | "tsv";
  pasteText: string;
  file: File | null;
  isInputGsheetUrl: boolean;
  setColumnOverride: (column: string, field: string) => void;
}

interface ReviewGateState {
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
  reviewRequiredColumns: string[];
  reviewedColumns: Record<string, boolean>;
}

interface UseReviewGateReturn {
  state: ReviewGateState;
  reviewGateReason: string | undefined;
  reviewRemainingCount: number;
  completeReview: () => void;
  handleColumnOverride: (column: string, field: string) => void;
  open: (columns: string[]) => void;
  clear: () => void;
  setReviewedColumns: Dispatch<SetStateAction<Record<string, boolean>>>;
}

export function useReviewGate({
  inputSource,
  format,
  mode,
  pasteText,
  file,
  isInputGsheetUrl,
  setColumnOverride,
}: UseReviewGateParams): UseReviewGateReturn {
  const [requiresReviewApproval, setRequiresReviewApproval] = useState(false);
  const [reviewApproved, setReviewApproved] = useState(false);
  const [reviewRequiredColumns, setReviewRequiredColumns] = useState<string[]>([]);
  const [reviewedColumns, setReviewedColumns] = useState<Record<string, boolean>>({});
  const latestInputSignatureRef = useRef("");

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

  // Complete review confirmation
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

  // Handle column override
  const handleColumnOverride = useCallback(
    (column: string, field: string) => {
      setColumnOverride(column, field);
      if (!requiresReviewApproval) return;
      if (!reviewRequiredColumns.includes(column)) return;
      setReviewedColumns((prev) => ({ ...prev, [column]: true }));
    },
    [setColumnOverride, requiresReviewApproval, reviewRequiredColumns]
  );

  // Open review gate with columns
  const open = useCallback((columns: string[]) => {
    setRequiresReviewApproval(true);
    setReviewApproved(false);
    setReviewRequiredColumns(columns);
    setReviewedColumns({});
  }, []);

  // Clear review state
  const clear = useCallback(() => {
    setRequiresReviewApproval(false);
    setReviewApproved(false);
    setReviewRequiredColumns([]);
    setReviewedColumns({});
  }, []);

  // Calculate reason and remaining count
  const reviewGateReason = useMemo(
    () => (requiresReviewApproval && !reviewApproved ? "Review mapping first" : undefined),
    [requiresReviewApproval, reviewApproved]
  );

  const reviewRemainingCount = useMemo(
    () => countRemainingReviews(reviewRequiredColumns, reviewedColumns),
    [reviewRequiredColumns, reviewedColumns]
  );

  return {
    state: {
      requiresReviewApproval,
      reviewApproved,
      reviewRequiredColumns,
      reviewedColumns,
    },
    reviewGateReason,
    reviewRemainingCount,
    completeReview,
    handleColumnOverride,
    open,
    clear,
    setReviewedColumns,
  };
}
