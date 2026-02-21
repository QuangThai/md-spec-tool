import { useOutputActions } from "@/hooks/useOutputActions";
import { useWorkbenchConversion } from "@/hooks/useWorkbenchConversion";
import type { FailedAction } from "@/hooks/useWorkbenchContracts";

interface UseWorkbenchConversionPipelinesParams {
  setLastFailedAction: (action: FailedAction) => void;
  gsheetPreviewSlice: {
    selectedBlockId?: string;
    trustedMapping: Record<string, string>;
  };
  selectedGoogleSheetRange: string;
  reviewApi: {
    open(columns: string[]): void;
    clear(): void;
  };
  includeMetadata: boolean;
  numberRows: boolean;
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
}

export function useWorkbenchConversionPipelines({
  setLastFailedAction,
  gsheetPreviewSlice,
  selectedGoogleSheetRange,
  reviewApi,
  includeMetadata,
  numberRows,
  inputSource,
}: UseWorkbenchConversionPipelinesParams) {
  const { handleConvert, handleGetAISuggestions, showFeedback, setShowFeedback } =
    useWorkbenchConversion({
      setLastFailedAction,
      gsheetPreviewSlice,
      gsheetRangeValue: selectedGoogleSheetRange,
      reviewApi,
      includeMetadata,
      numberRows,
      inputSource,
    });

  const output = useOutputActions();

  return {
    handleConvert,
    handleGetAISuggestions,
    showFeedback,
    setShowFeedback,
    output,
  };
}
