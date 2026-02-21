import { useCallback, useMemo } from "react";
import type { ConversionRecord, OutputFormat } from "@/lib/types";
import type { SourcePanelProps } from "@/components/workbench/SourcePanel";
import type { OutputPanelProps } from "@/components/workbench/OutputPanel";
import type { WorkbenchOverlaysProps } from "@/components/workbench/WorkbenchOverlays";

interface UseWorkbenchViewPropsParams {
  sourcePanelBase: Omit<
    SourcePanelProps,
    | "onToggleReviewColumn"
    | "onMarkAllReviewed"
    | "onModeChange"
    | "onRetryFailedAction"
    | "onConvert"
    | "onOpenTemplateEditor"
    | "onOpenValidation"
  >;
  outputPanelBase: Omit<
    OutputPanelProps,
    | "onRetryAISuggestions"
    | "onSaveSnapshot"
    | "onCompareSnapshots"
    | "onClearSnapshots"
    | "onShowHistory"
  >;
  overlaysBase: Omit<
    WorkbenchOverlaysProps,
    | "onSelectHistory"
    | "onDismissFeedback"
    | "onConvert"
    | "onCopy"
    | "onExport"
    | "onShowHistory"
    | "onOpenTemplateEditor"
    | "onOpenValidation"
  >;
  setMode: (mode: "paste" | "xlsx" | "tsv") => void;
  setFile: (file: File | null) => void;
  reviewRequiredColumns: string[];
  setReviewedColumns: (
    updater:
      | Record<string, boolean>
      | ((prev: Record<string, boolean>) => Record<string, boolean>)
  ) => void;
  onRetryFailedAction: () => Promise<void>;
  onConvert: () => Promise<void>;
  onOpenTemplateEditor: () => void;
  onOpenValidation: () => void;
  onRetryAISuggestions: () => void;
  onSaveSnapshot: () => void;
  onCompareSnapshots: () => void;
  onClearSnapshots: () => void;
  onShowHistory: () => void;
  onSelectHistory: (record: ConversionRecord) => void;
  onDismissFeedback: () => void;
  onCopyCommand: () => void;
  onExportCommand: () => void;
  format: OutputFormat;
}

export function useWorkbenchViewProps({
  sourcePanelBase,
  outputPanelBase,
  overlaysBase,
  setMode,
  setFile,
  reviewRequiredColumns,
  setReviewedColumns,
  onRetryFailedAction,
  onConvert,
  onOpenTemplateEditor,
  onOpenValidation,
  onRetryAISuggestions,
  onSaveSnapshot,
  onCompareSnapshots,
  onClearSnapshots,
  onShowHistory,
  onSelectHistory,
  onDismissFeedback,
  onCopyCommand,
  onExportCommand,
  format,
}: UseWorkbenchViewPropsParams) {
  const handleModeChange = useCallback(
    (newMode: "paste" | "xlsx" | "tsv") => {
      setMode(newMode);
      setFile(null);
    },
    [setMode, setFile]
  );

  const handleToggleReviewColumn = useCallback(
    (column: string) => {
      setReviewedColumns((prev) => ({
        ...prev,
        [column]: !prev[column],
      }));
    },
    [setReviewedColumns]
  );

  const handleMarkAllReviewed = useCallback(() => {
    setReviewedColumns(
      Object.fromEntries(reviewRequiredColumns.map((column) => [column, true]))
    );
  }, [reviewRequiredColumns, setReviewedColumns]);

  const sourcePanelProps = useMemo<SourcePanelProps>(
    () => ({
      ...sourcePanelBase,
      onModeChange: handleModeChange,
      onToggleReviewColumn: handleToggleReviewColumn,
      onMarkAllReviewed: handleMarkAllReviewed,
      onRetryFailedAction: () => {
        void onRetryFailedAction();
      },
      onConvert: () => {
        void onConvert();
      },
      onOpenTemplateEditor,
      onOpenValidation,
    }),
    [
      sourcePanelBase,
      handleModeChange,
      handleToggleReviewColumn,
      handleMarkAllReviewed,
      onRetryFailedAction,
      onConvert,
      onOpenTemplateEditor,
      onOpenValidation,
    ]
  );

  const outputPanelProps = useMemo<OutputPanelProps>(
    () => ({
      ...outputPanelBase,
      onRetryAISuggestions,
      onSaveSnapshot,
      onCompareSnapshots,
      onClearSnapshots,
      onShowHistory,
    }),
    [
      outputPanelBase,
      onRetryAISuggestions,
      onSaveSnapshot,
      onCompareSnapshots,
      onClearSnapshots,
      onShowHistory,
    ]
  );

  const overlaysProps = useMemo<WorkbenchOverlaysProps>(
    () => ({
      ...overlaysBase,
      onConvert,
      onCopy: onCopyCommand,
      onExport: onExportCommand,
      onShowHistory,
      onOpenTemplateEditor,
      onOpenValidation,
      onSelectHistory,
      currentTemplate: format,
      onSelectTemplate: overlaysBase.onSelectTemplate,
      onDismissFeedback,
    }),
    [
      overlaysBase,
      onConvert,
      onCopyCommand,
      onExportCommand,
      onShowHistory,
      onOpenTemplateEditor,
      onOpenValidation,
      onSelectHistory,
      format,
      onDismissFeedback,
    ]
  );

  return {
    sourcePanelProps,
    outputPanelProps,
    overlaysProps,
  };
}
