import { useGoogleSheetInput } from "@/hooks/useGoogleSheetInput";
import { useWorkbenchPreview } from "@/hooks/useWorkbenchPreview";
import { useFileHandling } from "@/hooks/useFileHandling";
import { useReviewGate } from "@/hooks/useReviewGate";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import { mapErrorToUserFacing } from "@/lib/errorUtils";
import { canConfirmReview } from "@/lib/reviewGate";
import type { FailedAction } from "@/hooks/useWorkbenchContracts";
import type { InputMode, OutputFormat } from "@/lib/types";
import { useMemo } from "react";

interface UseWorkbenchInputPipelinesParams {
  mode: InputMode;
  pasteText: string;
  file: File | null;
  format: OutputFormat;
  debouncedPasteText: string;
  error: string | null;
  setLastFailedAction: (action: FailedAction) => void;
}

export function useWorkbenchInputPipelines({
  mode,
  pasteText,
  file,
  format,
  debouncedPasteText,
  error,
  setLastFailedAction,
}: UseWorkbenchInputPipelinesParams) {
  const {
    gsheetLoading,
    gsheetRange,
    setGsheetRange,
    gsheetRangeValue,
    googleAuth,
  } = useGoogleSheetInput({
    debouncedPasteText,
    setLastFailedAction,
  });

  const { dragOver, handleFileChange, onDrop, onDragOver, onDragLeave } =
    useFileHandling({
      setLastFailedAction,
    });

  const isGsheetUrl = isGoogleSheetsURL(debouncedPasteText.trim());
  const isInputGsheetUrl = isGoogleSheetsURL(pasteText.trim());
  const inputSource: "paste" | "xlsx" | "gsheet" | "tsv" =
    mode === "paste" ? (isInputGsheetUrl ? "gsheet" : "paste") : mode;

  const review = useReviewGate({
    inputSource,
    format,
    mode,
    pasteText,
    file,
    isInputGsheetUrl,
  });

  const canConfirmReviewGate = useMemo(
    () => canConfirmReview(review.reviewRequiredColumns, review.reviewedColumns),
    [review.reviewRequiredColumns, review.reviewedColumns]
  );

  const mappedAppError = useMemo(
    () => (error ? mapErrorToUserFacing(error) : null),
    [error]
  );

  const { handleRetryPreview, refetchGoogleSheetPreview, gsheetPreviewSlice } =
    useWorkbenchPreview({
      debouncedPasteText,
      isGsheetUrl,
      gsheetRangeValue,
      setLastFailedAction,
      inputSource,
      format,
    });

  return {
    gsheetLoading,
    gsheetRange,
    setGsheetRange,
    googleAuth,
    dragOver,
    handleFileChange,
    onDrop,
    onDragOver,
    onDragLeave,
    isInputGsheetUrl,
    inputSource,
    review,
    canConfirmReviewGate,
    mappedAppError,
    handleRetryPreview,
    refetchGoogleSheetPreview,
    gsheetPreviewSlice,
    selectedGoogleSheetRange: gsheetRangeValue,
  };
}
