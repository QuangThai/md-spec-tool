import { useWorkbenchInteractions } from "@/hooks/useWorkbenchInteractions";
import { useMDFlowStore } from "@/lib/mdflowStore";
import { useCallback, useMemo } from "react";
import type {
  WorkbenchActionBundle,
  WorkbenchDataFlows,
} from "@/hooks/useWorkbenchContracts";

export function useWorkbenchActions(
  data: WorkbenchDataFlows
): WorkbenchActionBundle {
  const retryAISuggestions = useCallback(() => {
    void data.handleGetAISuggestions();
  }, [data.handleGetAISuggestions]);

  const saveSnapshot = useCallback(() => {
    data.diff.saveSnapshot(data.mdflowOutput);
  }, [data.diff.saveSnapshot, data.mdflowOutput]);

  const compareSnapshots = useCallback(() => {
    void data.diff.compareSnapshots();
  }, [data.diff.compareSnapshots]);

  const clearSnapshots = useCallback(() => {
    data.diff.clearSnapshots();
  }, [data.diff.clearSnapshots]);

  const handleRetryFailedAction = useCallback(async () => {
    if (data.lastFailedAction === "convert") {
      await data.handleConvert();
      return;
    }
    if (data.lastFailedAction === "preview") {
      await data.handleRetryPreview();
    }
  }, [data.handleConvert, data.handleRetryPreview, data.lastFailedAction]);

  const togglePreview = useCallback(() => {
    data.setShowPreview(!useMDFlowStore.getState().showPreview);
  }, [data.setShowPreview]);

  const { handleCommandPaletteCopy, handleCommandPaletteExport } =
    useWorkbenchInteractions({
      showCommandPalette: data.showCommandPalette,
      setShowCommandPalette: data.setShowCommandPalette,
      showHistory: data.showHistory,
      setShowHistory: data.setShowHistory,
      showDiff: data.diff.showDiff,
      setShowDiff: data.diff.setShowDiff,
      showTemplateEditor: data.showTemplateEditor,
      setShowTemplateEditor: data.setShowTemplateEditor,
      showValidationConfigurator: data.showValidationConfigurator,
      setShowValidationConfigurator: data.setShowValidationConfigurator,
      openCommandPalette: data.openCommandPalette,
      handleConvert: data.handleConvert,
      togglePreview,
      mdflowOutput: data.mdflowOutput,
      reviewGateReason: data.review.reviewGateReason,
      copyOutput: data.output.handleCopy,
      exportOutput: data.output.handleDownload,
    });

  const googleSheetInputProps = useMemo(
    () => ({
      gsheetLoading: data.gsheetLoading,
      gsheetRange: data.gsheetRange,
      setGsheetRange: data.setGsheetRange,
      setSelectedGid: data.setSelectedGid,
      googleAuth: data.googleAuth,
      gsheetTabs: data.gsheetTabs,
      selectedGid: data.selectedGid,
      onRefetchGsheetPreview: data.refetchGoogleSheetPreview,
    }),
    [
      data.gsheetLoading,
      data.gsheetRange,
      data.setGsheetRange,
      data.setSelectedGid,
      data.googleAuth,
      data.gsheetTabs,
      data.selectedGid,
      data.refetchGoogleSheetPreview,
    ]
  );

  const fileHandlingProps = useMemo(
    () => ({
      dragOver: data.dragOver,
      handleFileChange: data.handleFileChange,
      onDrop: data.onDrop,
      onDragOver: data.onDragOver,
      onDragLeave: data.onDragLeave,
    }),
    [
      data.dragOver,
      data.handleFileChange,
      data.onDrop,
      data.onDragOver,
      data.onDragLeave,
    ]
  );

  return {
    retryAISuggestions,
    saveSnapshot,
    compareSnapshots,
    clearSnapshots,
    handleRetryFailedAction,
    togglePreview,
    handleCommandPaletteCopy,
    handleCommandPaletteExport,
    googleSheetInputProps,
    fileHandlingProps,
  };
}
