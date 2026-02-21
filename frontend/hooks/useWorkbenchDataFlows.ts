import { useWorkbenchStoreSelection } from "@/hooks/useWorkbenchStoreSelection";
import { useWorkbenchUIState } from "@/hooks/useWorkbenchUIState";
import { useWorkbenchInputPipelines } from "@/hooks/useWorkbenchInputPipelines";
import { useWorkbenchConversionPipelines } from "@/hooks/useWorkbenchConversionPipelines";
import { useDiffSnapshots } from "@/hooks/useDiffSnapshots";
import { emitTelemetryEvent } from "@/lib/telemetry";
import { useEffect, useRef } from "react";
import type { WorkbenchDataFlows } from "@/hooks/useWorkbenchContracts";

export function useWorkbenchDataFlows(): WorkbenchDataFlows {
  const { state, actions, history, openaiKey, setOpenaiKey, clearOpenaiKey, templates } =
    useWorkbenchStoreSelection();

  const {
    showHistory,
    setShowHistory,
    showValidationConfigurator,
    setShowValidationConfigurator,
    showTemplateEditor,
    setShowTemplateEditor,
    showCommandPalette,
    setShowCommandPalette,
    showApiKeyInput,
    toggleApiKeyInput,
    apiKeyDraft,
    setApiKeyDraft,
    includeMetadata,
    numberRows,
    lastFailedAction,
    setLastFailedAction,
    debouncedPasteText,
    openTemplateEditor,
    openValidationConfigurator,
    openHistory,
    openCommandPalette,
  } = useWorkbenchUIState({
    mode: state.mode,
    pasteText: state.pasteText,
    format: state.format,
  });

  const input = useWorkbenchInputPipelines({
    mode: state.mode,
    pasteText: state.pasteText,
    file: state.file,
    format: state.format,
    debouncedPasteText,
    error: state.error,
    setLastFailedAction,
  });

  const conversion = useWorkbenchConversionPipelines({
    setLastFailedAction,
    gsheetPreviewSlice: input.gsheetPreviewSlice,
    selectedGoogleSheetRange: input.selectedGoogleSheetRange,
    reviewApi: input.review,
    includeMetadata,
    numberRows,
    inputSource: input.inputSource,
  });

  const diff = useDiffSnapshots();

  const studioOpenedTrackedRef = useRef(false);

  useEffect(() => {
    return () => actions.reset();
  }, [actions]);

  useEffect(() => {
    if (studioOpenedTrackedRef.current) {
      return;
    }
    studioOpenedTrackedRef.current = true;
    emitTelemetryEvent("studio_opened", {
      status: "success",
      input_source: input.inputSource,
      template_type: state.format,
    });
  }, [state.format, input.inputSource]);

  return {
    mode: state.mode,
    pasteText: state.pasteText,
    file: state.file,
    sheets: state.sheets,
    selectedSheet: state.selectedSheet,
    gsheetTabs: state.gsheetTabs,
    selectedGid: state.selectedGid,
    format: state.format,
    mdflowOutput: state.mdflowOutput,
    warnings: state.warnings,
    meta: state.meta,
    loading: state.loading,
    error: state.error,
    preview: state.preview,
    previewLoading: state.previewLoading,
    showPreview: state.showPreview,
    columnOverrides: state.columnOverrides,
    aiSuggestions: state.aiSuggestions,
    aiSuggestionsLoading: state.aiSuggestionsLoading,
    aiSuggestionsError: state.aiSuggestionsError,
    aiConfigured: state.aiConfigured ?? false,
    setMode: actions.setMode,
    setPasteText: actions.setPasteText,
    setFile: actions.setFile,
    setSelectedSheet: actions.setSelectedSheet,
    setSelectedGid: actions.setSelectedGid,
    setFormat: actions.setFormat,
    setResult: actions.setResult,
    setShowPreview: actions.setShowPreview,
    history,
    openaiKey,
    setOpenaiKey,
    clearOpenaiKey,
    templates,
    showHistory,
    setShowHistory,
    showValidationConfigurator,
    setShowValidationConfigurator,
    showTemplateEditor,
    setShowTemplateEditor,
    showCommandPalette,
    setShowCommandPalette,
    showApiKeyInput,
    toggleApiKeyInput,
    apiKeyDraft,
    setApiKeyDraft,
    lastFailedAction,
    openTemplateEditor,
    openValidationConfigurator,
    openHistory,
    openCommandPalette,
    diff,
    gsheetLoading: input.gsheetLoading,
    gsheetRange: input.gsheetRange,
    setGsheetRange: input.setGsheetRange,
    googleAuth: input.googleAuth,
    dragOver: input.dragOver,
    handleFileChange: input.handleFileChange,
    onDrop: input.onDrop,
    onDragOver: input.onDragOver,
    onDragLeave: input.onDragLeave,
    isInputGsheetUrl: input.isInputGsheetUrl,
    inputSource: input.inputSource,
    review: input.review,
    canConfirmReviewGate: input.canConfirmReviewGate,
    mappedAppError: input.mappedAppError,
    handleRetryPreview: input.handleRetryPreview,
    refetchGoogleSheetPreview: input.refetchGoogleSheetPreview,
    output: conversion.output,
    handleConvert: conversion.handleConvert,
    handleGetAISuggestions: conversion.handleGetAISuggestions,
    showFeedback: conversion.showFeedback,
    setShowFeedback: conversion.setShowFeedback,
  };
}
