import { useCallback, useMemo, useRef, useEffect } from "react";
import {
  usePreviewPasteQuery,
  usePreviewTSVQuery,
  usePreviewXLSXQuery,
  usePreviewGoogleSheetQuery,
} from "@/lib/mdflowQueries";
import { emitTelemetryEvent } from "@/lib/telemetry";
import type { PreviewResponse } from "@/lib/types";

export interface UseWorkbenchPreviewProps {
  debouncedPasteText: string;
  isGsheetUrl: boolean;
  gsheetRangeValue: string;
  setLastFailedAction: (action: "preview" | "convert" | "other" | null) => void;
  inputSource: "paste" | "xlsx" | "gsheet" | "tsv";
  format: string;
  mode: "paste" | "xlsx" | "tsv";
  file: File | null;
  selectedSheet: string;
  selectedGid: string;
  gsheetTabs: Array<{ gid: string; title: string }>;
  setPreview: (preview: PreviewResponse | null) => void;
  setPreviewLoading: (loading: boolean) => void;
  setShowPreview: (show: boolean) => void;
}

export interface GsheetPreviewSlice {
  selectedBlockId?: string;
  trustedMapping: Record<string, string>;
}

export interface UseWorkbenchPreviewReturn {
  previewQueries: {
    pasteQuery: ReturnType<typeof usePreviewPasteQuery>;
    tsvQuery: ReturnType<typeof usePreviewTSVQuery>;
    xlsxQuery: ReturnType<typeof usePreviewXLSXQuery>;
    googleSheetQuery: ReturnType<typeof usePreviewGoogleSheetQuery>;
  };
  activePreviewError: Error | null;
  handleRetryPreview: () => Promise<void>;
  gsheetPreviewSlice: GsheetPreviewSlice;
}

export const useWorkbenchPreview = ({
  debouncedPasteText,
  isGsheetUrl,
  gsheetRangeValue,
  setLastFailedAction,
  inputSource,
  format,
  mode,
  file,
  selectedSheet,
  selectedGid,
  gsheetTabs,
  setPreview,
  setPreviewLoading,
  setShowPreview,
}: UseWorkbenchPreviewProps): UseWorkbenchPreviewReturn => {
  // Refs for telemetry tracking
  const previewStartedAtRef = useRef<number | null>(null);
  const previewAttemptRef = useRef(false);

  // Query hooks
  const pasteQuery = usePreviewPasteQuery(
    debouncedPasteText,
    mode === "paste" && debouncedPasteText.trim().length > 0 && !isGsheetUrl,
    format
  );

  const tsvQuery = usePreviewTSVQuery(file, mode === "tsv", format);

  const xlsxQuery = usePreviewXLSXQuery(
    file,
    selectedSheet,
    mode === "xlsx",
    format
  );

  const googleSheetQuery = usePreviewGoogleSheetQuery(
    debouncedPasteText.trim(),
    selectedGid,
    mode === "paste" && isGsheetUrl && gsheetTabs.length > 0 && Boolean(selectedGid),
    format,
    gsheetRangeValue
  );

  // Determine active preview error based on mode
  const activePreviewError = useMemo(() => {
    if (mode === "paste" && isGsheetUrl) return googleSheetQuery.error;
    if (mode === "paste") return pasteQuery.error;
    if (mode === "xlsx") return xlsxQuery.error;
    return tsvQuery.error;
  }, [
    isGsheetUrl,
    mode,
    googleSheetQuery.error,
    pasteQuery.error,
    tsvQuery.error,
    xlsxQuery.error,
  ]);

  // Effect: Track preview errors
  useEffect(() => {
    if (activePreviewError) {
      setLastFailedAction("preview");
    }
  }, [activePreviewError, setLastFailedAction]);

  // Effect: Sync preview data to store (paste mode)
  useEffect(() => {
    if (mode !== "paste") return;
    if (!debouncedPasteText.trim()) {
      setPreview(null);
      setShowPreview(false);
      return;
    }
    if (isGsheetUrl) {
      if (googleSheetQuery.data) {
        setPreview(googleSheetQuery.data);
        setShowPreview(true);
      } else {
        setPreview(null);
        setShowPreview(false);
      }
      return;
    }
    if (pasteQuery.data) {
      setPreview(pasteQuery.data);
      setShowPreview(true);
    }
  }, [
    debouncedPasteText,
    mode,
    isGsheetUrl,
    pasteQuery.data,
    googleSheetQuery.data,
    setPreview,
    setShowPreview,
  ]);

  // Effect: Sync preview data to store (xlsx mode)
  useEffect(() => {
    if (mode === "xlsx" && xlsxQuery.data) {
      setPreview(xlsxQuery.data);
      setShowPreview(true);
    }
  }, [mode, xlsxQuery.data, setPreview, setShowPreview]);

  // Effect: Sync preview data to store (tsv mode)
  useEffect(() => {
    if (mode === "tsv" && tsvQuery.data) {
      setPreview(tsvQuery.data);
      setShowPreview(true);
    }
  }, [mode, tsvQuery.data, setPreview, setShowPreview]);

  // Effect: Preview loading & telemetry
  useEffect(() => {
    const isLoading =
      (mode === "paste" && pasteQuery.isFetching) ||
      (mode === "paste" && isGsheetUrl && googleSheetQuery.isFetching) ||
      (mode === "xlsx" && xlsxQuery.isFetching) ||
      (mode === "tsv" && tsvQuery.isFetching);

    setPreviewLoading(isLoading);

    if (isLoading && !previewAttemptRef.current) {
      previewAttemptRef.current = true;
      previewStartedAtRef.current = Date.now();
      emitTelemetryEvent("preview_started", {
        status: "success",
        input_source: inputSource,
        template_type: format,
      });
      return;
    }

    if (!isLoading && previewAttemptRef.current) {
      previewAttemptRef.current = false;
      const durationMs = previewStartedAtRef.current
        ? Date.now() - previewStartedAtRef.current
        : undefined;
      previewStartedAtRef.current = null;

      const activePreviewData =
        mode === "paste" && isGsheetUrl
          ? googleSheetQuery.data
          : mode === "paste"
            ? pasteQuery.data
            : mode === "xlsx"
              ? xlsxQuery.data
              : tsvQuery.data;

      const activePreviewErrorForTelemetry =
        mode === "paste" && isGsheetUrl
          ? googleSheetQuery.error
          : mode === "paste"
            ? pasteQuery.error
            : mode === "xlsx"
              ? xlsxQuery.error
              : tsvQuery.error;

      if (activePreviewErrorForTelemetry) {
        const errorMessage =
          activePreviewErrorForTelemetry instanceof Error
            ? activePreviewErrorForTelemetry.message
            : "preview_failed";
        emitTelemetryEvent("preview_failed", {
          status: "error",
          input_source: inputSource,
          template_type: format,
          duration_ms: durationMs,
          error_code: errorMessage,
        });
      } else if (activePreviewData) {
        const confidence = activePreviewData.confidence ?? 0;
        const lowConfidenceColumns =
          activePreviewData.mapping_quality?.low_confidence_columns?.length ?? 0;
        emitTelemetryEvent("preview_succeeded", {
          status: "success",
          input_source: inputSource,
          template_type: format,
          duration_ms: durationMs,
          confidence_score: confidence,
          warning_count: activePreviewData.unmapped_columns?.length ?? 0,
          needs_review: confidence < 50 || lowConfidenceColumns > 0,
        });
      }
    }
  }, [
    mode,
    isGsheetUrl,
    inputSource,
    format,
    pasteQuery.isFetching,
    googleSheetQuery.isFetching,
    tsvQuery.isFetching,
    xlsxQuery.isFetching,
    pasteQuery.data,
    googleSheetQuery.data,
    xlsxQuery.data,
    tsvQuery.data,
    pasteQuery.error,
    googleSheetQuery.error,
    xlsxQuery.error,
    tsvQuery.error,
    setPreviewLoading,
  ]);

  // Compute gsheet preview slice (memoized for conversion hook)
  const gsheetPreviewSlice = useMemo(() => {
    if (!googleSheetQuery.data) {
      return {
        selectedBlockId: undefined,
        trustedMapping: {},
      };
    }

    const previewSelectedBlockId = googleSheetQuery.data.selected_block_id;
    const previewColumnMapping = googleSheetQuery.data.column_mapping || {};
    const previewColumnConfidence =
      googleSheetQuery.data.mapping_quality?.column_confidence || {};

    const trustedMapping =
      (googleSheetQuery.data.confidence ?? 0) >= 50
        ? Object.fromEntries(
            Object.entries(previewColumnMapping).filter(([header, mappedField]) => {
              if (!header || !mappedField) return false;
              const score = previewColumnConfidence[header];
              return typeof score !== "number" || score >= 0.7;
            })
          )
        : {};

    return {
      selectedBlockId: previewSelectedBlockId,
      trustedMapping,
    };
  }, [googleSheetQuery.data]);

  // Retry handler
  const handleRetryPreview = useCallback(async () => {
    // Trigger refetch based on current mode
    if (mode === "paste" && isGsheetUrl) {
      googleSheetQuery.refetch?.();
    } else if (mode === "paste") {
      pasteQuery.refetch?.();
    } else if (mode === "xlsx") {
      xlsxQuery.refetch?.();
    } else if (mode === "tsv") {
      tsvQuery.refetch?.();
    }
  }, [mode, isGsheetUrl, googleSheetQuery, pasteQuery, xlsxQuery, tsvQuery]);

  return {
    previewQueries: {
      pasteQuery,
      tsvQuery,
      xlsxQuery,
      googleSheetQuery,
    },
    activePreviewError,
    handleRetryPreview,
    gsheetPreviewSlice,
  };
};
