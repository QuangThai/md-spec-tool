import { useCallback, useState } from "react";
import {
  useConvertPasteMutation,
  useConvertXLSXMutation,
  useConvertTSVMutation,
  useConvertGoogleSheetMutation,
  useAISuggestionsMutation,
} from "@/lib/mdflowQueries";
import { emitTelemetryEvent } from "@/lib/telemetry";
import { buildReviewRequiredColumns } from "@/lib/reviewGate";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import { toast } from "@/components/ui/Toast";
import type { MDFlowWarning, MDFlowMeta, ConversionRecord, AISuggestion, PreviewResponse } from "@/lib/types";

interface UseWorkbenchConversionProps {
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
  gsheetPreviewSlice: {
    selectedBlockId?: string;
    trustedMapping: Record<string, string>;
  };
  gsheetRangeValue: string;
  reviewApi: {
    open(columns: string[]): void;
    clear(): void;
  };
  includeMetadata: boolean;
  numberRows: boolean;
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  mode: "paste" | "xlsx" | "tsv";
  pasteText: string;
  file: File | null;
  selectedSheet: string;
  selectedGid: string;
  format: string;
  columnOverrides: Record<string, string>;
  preview: PreviewResponse | null;
  gsheetTabs: Array<{ gid: string; title: string }>;
  mdflowOutput: string;
  setResult: (output: string, warnings: MDFlowWarning[], meta: MDFlowMeta) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setShowPreview: (show: boolean) => void;
  addToHistory: (record: Omit<ConversionRecord, "id" | "timestamp">) => void;
  aiSuggestionsLoading: boolean;
  setAISuggestionsLoading: (loading: boolean) => void;
  setAISuggestionsError: (error: string | null) => void;
  setAISuggestions: (suggestions: AISuggestion[], configured: boolean) => void;
  clearAISuggestions: () => void;
}

interface UseWorkbenchConversionReturn {
  handleConvert: () => Promise<void>;
  handleGetAISuggestions: () => Promise<void>;
  showFeedback: boolean;
  setShowFeedback: (show: boolean) => void;
}

export function useWorkbenchConversion(
  props: UseWorkbenchConversionProps
): UseWorkbenchConversionReturn {
  const {
    setLastFailedAction,
    gsheetPreviewSlice,
    gsheetRangeValue,
    reviewApi,
    includeMetadata,
    numberRows,
    inputSource,
    mode,
    pasteText,
    file,
    selectedSheet,
    selectedGid,
    format,
    columnOverrides,
    preview,
    gsheetTabs,
    mdflowOutput,
    setResult,
    setLoading,
    setError,
    setShowPreview,
    addToHistory,
    aiSuggestionsLoading,
    setAISuggestionsLoading,
    setAISuggestionsError,
    setAISuggestions,
    clearAISuggestions,
  } = props;

  const [showFeedback, setShowFeedback] = useState(false);

  const convertPasteMutation = useConvertPasteMutation();
  const convertXLSXMutation = useConvertXLSXMutation();
  const convertTSVMutation = useConvertTSVMutation();
  const convertGoogleSheetMutation = useConvertGoogleSheetMutation();
  const aiSuggestionsMutation = useAISuggestionsMutation();

  const handleConvert = useCallback(async () => {
    setLoading(true);
    setError(null);
    setLastFailedAction(null);
    const startedAt = Date.now();
    emitTelemetryEvent("convert_started", {
      status: "success",
      input_source: inputSource,
      template_type: format,
    });

    try {
      let result;
      let inputPreview = "";
      if (mode === "paste") {
        if (!pasteText.trim()) {
          setError("Missing source data");
          setLastFailedAction("convert");
          return;
        }

        // Check if it's a Google Sheets URL
        if (isGoogleSheetsURL(pasteText.trim())) {
          const effectiveColumnOverrides = {
            ...gsheetPreviewSlice.trustedMapping,
            ...columnOverrides,
          };
          result = await convertGoogleSheetMutation.mutateAsync({
            url: pasteText.trim(),
            template: format,
            gid: selectedGid,
            format,
            range: gsheetRangeValue || undefined,
            selectedBlockId: gsheetPreviewSlice.selectedBlockId,
            columnOverrides: effectiveColumnOverrides,
            includeMetadata,
            numberRows,
          });
          const selectedTab = gsheetTabs.find((tab) => tab.gid === selectedGid);
          const tabLabel = selectedTab?.title || selectedGid;
          inputPreview = tabLabel
            ? `Google Sheet: ${pasteText.trim().slice(0, 60)}... (${tabLabel})`
            : `Google Sheet: ${pasteText.trim().slice(0, 60)}...`;
        } else {
          result = await convertPasteMutation.mutateAsync({
            pasteText,
            template: format,
            format,
            columnOverrides,
            includeMetadata,
            numberRows,
          });
          inputPreview =
            pasteText.slice(0, 200) + (pasteText.length > 200 ? "..." : "");
        }
      } else if (mode === "xlsx") {
        if (!file) {
          setError("No file uploaded");
          setLastFailedAction("convert");
          return;
        }
        result = await convertXLSXMutation.mutateAsync({
          file,
          sheetName: selectedSheet,
          template: format,
          format,
          columnOverrides,
          includeMetadata,
          numberRows,
        });
        inputPreview = `${file.name}${selectedSheet ? ` (${selectedSheet})` : ""
          }`;
      } else {
        if (!file) {
          setError("No file uploaded");
          setLastFailedAction("convert");
          return;
        }
        result = await convertTSVMutation.mutateAsync({
          file,
          template: format,
          format,
          columnOverrides,
          includeMetadata,
          numberRows,
        });
        inputPreview = file.name;
      }

      if (result) {
        setResult(result.mdflow, result.warnings, result.meta);
        // Add to history (use format as display template name)
        addToHistory({
          mode,
          template: format,
          inputPreview,
          output: result.mdflow,
          meta: result.meta,
        });
        toast.success(
          "Conversion complete",
          `${result.meta?.total_rows || 0} rows processed`
        );
        if (result.needs_review) {
          const uniqueColumns = buildReviewRequiredColumns(preview ?? null);
          reviewApi.open(uniqueColumns);
          setShowPreview(true);
          emitTelemetryEvent("review_mapping_opened", {
            status: "success",
            input_source: inputSource,
            template_type: format,
            pending_columns: uniqueColumns.length,
          });
          toast.error(
            "Review recommended",
            "Low-confidence mapping detected. Please review preview before sharing."
          );
        } else {
          reviewApi.clear();
        }
        emitTelemetryEvent("convert_succeeded", {
          status: "success",
          input_source: inputSource,
          template_type: format,
          duration_ms: Date.now() - startedAt,
          warning_count: result.warnings?.length ?? 0,
          total_rows: result.meta?.total_rows ?? 0,
          confidence_score: Math.round(
            result.meta?.quality_report?.header_confidence ?? 0
          ),
          needs_review: Boolean(result.needs_review),
          ai_mode: result.meta?.ai_mode ?? "off",
          ai_used: result.meta?.ai_used ?? false,
          ai_model: result.meta?.ai_model ?? "",
          ai_prompt_version: result.meta?.ai_prompt_version ?? "",
          ai_estimated_cost_usd: result.meta?.ai_estimated_cost_usd ?? 0,
          ai_input_tokens: result.meta?.ai_estimated_input_tokens ?? 0,
          ai_output_tokens: result.meta?.ai_estimated_output_tokens ?? 0,
        });
        setTimeout(() => setShowFeedback(true), 2000);
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Conversion failed";
      setError(errorMessage);
      setLastFailedAction("convert");
      toast.error("Conversion failed", errorMessage);
      emitTelemetryEvent("convert_failed", {
        status: "error",
        input_source: inputSource,
        template_type: format,
        duration_ms: Date.now() - startedAt,
        error_code: errorMessage,
      });
    } finally {
      setLoading(false);
    }
  }, [
    mode,
    pasteText,
    file,
    selectedSheet,
    selectedGid,
    format,
    columnOverrides,
    includeMetadata,
    numberRows,
    gsheetRangeValue,
    gsheetPreviewSlice.selectedBlockId,
    gsheetPreviewSlice.trustedMapping,
    preview,
    setLoading,
    setError,
    setLastFailedAction,
    setResult,
    setShowPreview,
    reviewApi,
    inputSource,
    addToHistory,
    gsheetTabs,
    convertGoogleSheetMutation,
    convertPasteMutation,
    convertTSVMutation,
    convertXLSXMutation,
  ]);

  const handleGetAISuggestions = useCallback(async () => {
    if (!pasteText.trim() || aiSuggestionsLoading) return;

    setAISuggestionsLoading(true);
    setAISuggestionsError(null);
    clearAISuggestions();

    try {
      const result = await aiSuggestionsMutation.mutateAsync({
        pasteText,
        template: format,
      });
      setAISuggestions(result.suggestions, result.configured);
      if (result.error) {
        setAISuggestionsError(result.error);
      }
    } catch (error) {
      setAISuggestionsError(
        error instanceof Error ? error.message : "Failed to get suggestions"
      );
    } finally {
      setAISuggestionsLoading(false);
    }
  }, [
    pasteText,
    format,
    aiSuggestionsLoading,
    setAISuggestionsLoading,
    setAISuggestionsError,
    setAISuggestions,
    clearAISuggestions,
    aiSuggestionsMutation,
  ]);

  return {
    handleConvert,
    handleGetAISuggestions,
    showFeedback,
    setShowFeedback,
  };
}
