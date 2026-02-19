"use client";

import { useDiffSnapshots } from "@/hooks/useDiffSnapshots";
import { useOutputActions } from "@/hooks/useOutputActions";
import { useReviewGate } from "@/hooks/useReviewGate";
import { useGoogleSheetInput } from "@/hooks/useGoogleSheetInput";
import { useWorkbenchPreview } from "@/hooks/useWorkbenchPreview";
import { useFileHandling } from "@/hooks/useFileHandling";
import { useWorkbenchConversion } from "@/hooks/useWorkbenchConversion";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import { useMDFlowTemplatesQuery } from "@/lib/mdflowQueries";
import {
  useHistoryStore,
  useMDFlowActions,
  useMDFlowStore,
  useOpenAIKeyStore,
  type MDFlowState,
} from "@/lib/mdflowStore";
import { emitTelemetryEvent } from "@/lib/telemetry";
import { mapErrorToUserFacing } from "@/lib/errorUtils";
import { buildReviewRequiredColumns, canConfirmReview } from "@/lib/reviewGate";
import { ConversionRecord } from "@/lib/types";
import { isMac, useKeyboardShortcuts } from "@/lib/useKeyboardShortcuts";
import { AnimatePresence, motion } from "framer-motion";
import { RefreshCcw, Zap } from "lucide-react";
import dynamic from "next/dynamic";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useShallow } from "zustand/react/shallow";
import { CommandPalette } from "./CommandPalette";
import { ConversionFeedback } from "@/components/ConversionFeedback";
import HistoryModal, { KeyboardShortcutsTooltip } from "./HistoryModal";
import { OnboardingTour } from "./OnboardingTour";
import { QuotaStatus } from "./QuotaStatus";
import { toast, ToastContainer } from "./ui/Toast";
import { SourcePanel, OutputPanel } from "./workbench";

const EMPTY_TEMPLATES: string[] = [];

const TemplateEditor = dynamic(
  () => import("./TemplateEditor").then((mod) => mod.TemplateEditor),
  { ssr: false }
);

const ValidationConfigurator = dynamic(
  () =>
    import("./ValidationConfigurator").then(
      (mod) => mod.ValidationConfigurator
    ),
  { ssr: false }
);

const DiffModal = dynamic(
  () => import("./workbench/DiffModal").then((mod) => mod.DiffModal),
  { ssr: false }
);

const stagger = {
  container: {
    animate: { transition: { staggerChildren: 0.05, delayChildren: 0.08 } },
  },
  item: {
    initial: { opacity: 0, y: 12 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.35, ease: [0.16, 1, 0.3, 1] },
  },
};

export default function MDFlowWorkbench() {
  // Subscribe to state values with shallow comparison for performance
  const {
    mode,
    pasteText,
    file,
    sheets,
    selectedSheet,
    gsheetTabs,
    selectedGid,
    format,
    mdflowOutput,
    warnings,
    meta,
    loading,
    error,
    preview,
    previewLoading,
    showPreview,
    columnOverrides,
    aiSuggestions,
    aiSuggestionsLoading,
    aiSuggestionsError,
    aiConfigured,
  } = useMDFlowStore(
    useShallow((state): Omit<MDFlowState, 'validationRules' | 'dismissedWarningCodes' | 'template'> => ({
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
      aiConfigured: state.aiConfigured,
    }))
  );

  // Get all actions with single selector - no performance impact since actions never change
  const {
    setMode,
    setPasteText,
    setFile,
    setSheets,
    setSelectedSheet,
    setGsheetTabs,
    setSelectedGid,
    setFormat,
    setResult,
    setLoading,
    setError,
    setPreview,
    setPreviewLoading,
    setShowPreview,
    setColumnOverride,
    setAISuggestions,
    setAISuggestionsLoading,
    setAISuggestionsError,
    clearAISuggestions,
    reset,
  } = useMDFlowActions();

  const addToHistory = useHistoryStore((state) => state.addToHistory);
  const history = useHistoryStore((state) => state.history);

  const [showHistory, setShowHistory] = useState(false);
  const [showValidationConfigurator, setShowValidationConfigurator] =
    useState(false);
  const [showTemplateEditor, setShowTemplateEditor] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [debouncedPasteText, setDebouncedPasteText] = useState("");
  const [showApiKeyInput, setShowApiKeyInput] = useState(false);
  const [apiKeyDraft, setApiKeyDraft] = useState("");
  const [showAdvancedOptions, setShowAdvancedOptions] = useState(false);
  const [includeMetadata, setIncludeMetadata] = useState(true);
  const [numberRows, setNumberRows] = useState(false);
  const [lastFailedAction, setLastFailedAction] = useState<"preview" | "convert" | "other" | null>(null);
  const studioOpenedTrackedRef = useRef(false);
  const openaiKey = useOpenAIKeyStore((s) => s.apiKey);
  const setOpenaiKey = useOpenAIKeyStore((s) => s.setApiKey);
  const clearOpenaiKey = useOpenAIKeyStore((s) => s.clearApiKey);

  const { data: templateList } = useMDFlowTemplatesQuery();
  const templates = templateList ?? EMPTY_TEMPLATES;

  // ── Hooks ──
  const diff = useDiffSnapshots();
  const {
    gsheetLoading,
    gsheetRange,
    setGsheetRange,
    gsheetRangeValue,
    googleAuth,
  } = useGoogleSheetInput({
    debouncedPasteText,
    setLastFailedAction,
    mode,
    gsheetTabs,
    selectedGid,
    setGsheetTabs,
    setSelectedGid,
    setError,
  });

  const {
    dragOver,
    setDragOver,
    handleFileChange,
    onDrop,
    onDragOver,
    onDragLeave,
  } = useFileHandling({
    setLastFailedAction,
    mode,
    file,
    setFile,
    setLoading,
    setError,
    setPreview,
    setSheets,
    setSelectedSheet,
  });

  const isGsheetUrl = isGoogleSheetsURL(debouncedPasteText.trim());
  const isInputGsheetUrl = isGoogleSheetsURL(pasteText.trim());
  const inputSource: "paste" | "xlsx" | "gsheet" | "tsv" =
    mode === "paste" ? (isInputGsheetUrl ? "gsheet" : "paste") : mode;

  const review = useReviewGate({
    inputSource,
    format,
    mode: mode as "paste" | "xlsx" | "tsv",
    pasteText,
    file,
    isInputGsheetUrl,
    setColumnOverride,
  });

  const isTableFormat = format === "table";
  const previewQueryErrorRef = useRef(false);
  const mappedAppError = useMemo(
    () => (error ? mapErrorToUserFacing(error) : null),
    [error]
  );
  const changedOutputOptionsCount = (includeMetadata ? 0 : 1) + (numberRows ? 1 : 0);

  useEffect(() => {
    if (!isTableFormat && numberRows) {
      setNumberRows(false);
    }
  }, [isTableFormat, numberRows]);

  const {
    previewQueries,
    activePreviewError,
    handleRetryPreview,
    gsheetPreviewSlice,
  } = useWorkbenchPreview({
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
  });

  const previewPasteQuery = previewQueries.pasteQuery;
  const previewTSVQuery = previewQueries.tsvQuery;
  const previewXLSXQuery = previewQueries.xlsxQuery;
  const previewGoogleSheetQuery = previewQueries.googleSheetQuery;

  useEffect(() => {
    if (activePreviewError) {
      const message =
        activePreviewError instanceof Error
          ? activePreviewError.message
          : "Preview failed";
      previewQueryErrorRef.current = true;
      setError(message);
      setLastFailedAction("preview");
      return;
    }

    if (previewQueryErrorRef.current && lastFailedAction === "preview" && error) {
      previewQueryErrorRef.current = false;
      setError(null);
      setLastFailedAction(null);
    }
  }, [activePreviewError, error, lastFailedAction, setError]);

  const { handleConvert, handleGetAISuggestions, showFeedback, setShowFeedback } =
    useWorkbenchConversion({
      setLastFailedAction,
      gsheetPreviewSlice,
      gsheetRangeValue,
      reviewApi: review,
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
    });

  // Reset store when leaving Studio so data is not shown when user comes back
  useEffect(() => {
    return () => reset();
  }, [reset]);

  useEffect(() => {
    if (studioOpenedTrackedRef.current) {
      return;
    }
    studioOpenedTrackedRef.current = true;
    emitTelemetryEvent("studio_opened", {
      status: "success",
      input_source: inputSource,
      template_type: format,
    });
  }, [format, inputSource]);

  // Debounce paste text for preview queries
  useEffect(() => {
    if (mode !== "paste") {
      setDebouncedPasteText("");
      return;
    }

    const timer = setTimeout(() => {
      setDebouncedPasteText(pasteText);
    }, 500);

    return () => clearTimeout(timer);
  }, [pasteText, mode]);

  const output = useOutputActions(mdflowOutput);

  const handleRetryFailedAction = useCallback(async () => {
    if (lastFailedAction === "convert") {
      await handleConvert();
      return;
    }
    if (lastFailedAction === "preview") {
      await handleRetryPreview();
    }
  }, [handleConvert, handleRetryPreview, lastFailedAction]);

  const togglePreview = useCallback(() => {
    setShowPreview(!useMDFlowStore.getState().showPreview);
  }, [setShowPreview]);

  // Keyboard shortcuts via hook
  useKeyboardShortcuts({
    commandPalette: () => setShowCommandPalette(true),
    convert: handleConvert,
    copy: () => {
      if (mdflowOutput) {
        output.handleCopy();
        toast.success("Copied to clipboard");
      }
    },
    export: () => {
      if (mdflowOutput) {
        output.handleDownload();
        toast.success("Downloaded spec.mdflow.md");
      }
    },
    togglePreview,
    showShortcuts: () => { }, // Handled by KeyboardShortcutsTooltip
    escape: () => {
      if (showCommandPalette) setShowCommandPalette(false);
      else if (showHistory) setShowHistory(false);
      else if (diff.showDiff) diff.setShowDiff(false);
      else if (showTemplateEditor) setShowTemplateEditor(false);
      else if (showValidationConfigurator) setShowValidationConfigurator(false);
    },
  });

  return (
    <motion.div
      variants={stagger.container}
      initial="initial"
      animate="animate"
      className="flex flex-col gap-3 sm:gap-4 relative h-[calc(100vh-6rem)] sm:h-[calc(100vh-7rem)] lg:h-[calc(100vh-8rem)]"
    >
      {/* Onboarding Tour */}
      <OnboardingTour />

      {/* Main workspace: optimized for immediate visibility */}
      <div
        className="grid grid-cols-1 lg:grid-cols-2 gap-3 sm:gap-4 lg:gap-5 items-stretch flex-1 min-h-0"
        data-tour="welcome"
      >
        {/* Left: Source Panel */}
        <SourcePanel
          mode={mode}
          onModeChange={(newMode) => {
            setMode(newMode);
            setFile(null);
          }}
          pasteText={pasteText}
          onPasteTextChange={setPasteText}
          file={file}
          onFileChange={setFile}
          sheets={sheets}
          selectedSheet={selectedSheet}
          onSelectSheet={setSelectedSheet}
          format={format}
          onFormatChange={(v) => setFormat(v as any)}
          preview={preview}
          showPreview={showPreview}
          onTogglePreview={togglePreview}
          previewLoading={previewLoading}
          columnOverrides={columnOverrides}
          onColumnOverride={setColumnOverride}
          openaiKey={openaiKey}
          setOpenaiKey={setOpenaiKey}
          clearOpenaiKey={clearOpenaiKey}
          showApiKeyInput={showApiKeyInput}
          onToggleApiKeyInput={() => setShowApiKeyInput((prev) => !prev)}
          apiKeyDraft={apiKeyDraft}
          onApiKeyDraftChange={setApiKeyDraft}
          error={error}
          mappedAppError={mappedAppError}
          lastFailedAction={lastFailedAction}
          onRetryFailedAction={handleRetryFailedAction}
          reviewRequiredColumns={review.state.reviewRequiredColumns}
          reviewedColumns={review.state.reviewedColumns}
          reviewRemainingCount={review.reviewRemainingCount}
          onToggleReviewColumn={(col) =>
            review.setReviewedColumns((prev) => ({
              ...prev,
              [col]: !prev[col],
            }))
          }
          onMarkAllReviewed={() =>
            review.setReviewedColumns(
              Object.fromEntries(
                review.state.reviewRequiredColumns.map((column) => [column, true])
              )
            )
          }
          onCompleteReview={review.completeReview}
          requiresReviewApproval={review.state.requiresReviewApproval}
          reviewApproved={review.state.reviewApproved}
          googleSheetInput={{
            gsheetLoading,
            gsheetRange,
            setGsheetRange,
            gsheetRangeValue,
            setSelectedGid,
            googleAuth,
            googleSheetInput: {
              gsheetTabs,
              selectedGid,
            }
          } as any}
          isInputGsheetUrl={isInputGsheetUrl}
          fileHandling={{
            dragOver,
            setDragOver,
            handleFileChange,
            onDrop,
            onDragOver,
            onDragLeave,
          }}
          loading={loading}
          onConvert={handleConvert}
          onOpenTemplateEditor={() => setShowTemplateEditor(true)}
          onOpenValidation={() => setShowValidationConfigurator(true)}
        />

        {/* Right: Output Panel */}
        <OutputPanel
          mdflowOutput={mdflowOutput}
          loading={loading}
          meta={meta}
          warnings={warnings}
          aiSuggestions={aiSuggestions}
          aiSuggestionsLoading={aiSuggestionsLoading}
          aiSuggestionsError={aiSuggestionsError}
          aiConfigured={aiConfigured ?? false}
          onRetryAISuggestions={() => void handleGetAISuggestions()}
          requiresReviewApproval={review.state.requiresReviewApproval}
          reviewApproved={review.state.reviewApproved}
          reviewGateReason={review.reviewGateReason}
          format={format}
          copied={output.copied}
          onCopy={output.handleCopy}
          onDownload={() => {
            if (mdflowOutput) {
              const blob = new Blob([mdflowOutput], {
                type: "text/markdown",
              });
              const url = URL.createObjectURL(blob);
              const a = document.createElement("a");
              a.href = url;
              a.download = "spec.mdflow.md";
              a.click();
              URL.revokeObjectURL(url);
            }
          }}
          snapshotA={diff.snapshotA}
          snapshotB={diff.snapshotB}
          compareLoading={diff.compareLoading}
          onSaveSnapshot={() => diff.saveSnapshot(mdflowOutput)}
          onCompareSnapshots={() => diff.compareSnapshots()}
          onClearSnapshots={() => diff.clearSnapshots()}
          historyCount={history.length}
          onShowHistory={() => setShowHistory(true)}
        />
      </div>

      {/* Diff Viewer Modal */}
      <DiffModal
        showDiff={diff.showDiff}
        currentDiff={diff.currentDiff}
        onClose={() => diff.setShowDiff(false)}
      />

      {/* History Modal */}
      <AnimatePresence>
        {showHistory ? (
          <HistoryModal
            history={history}
            onClose={() => setShowHistory(false)}
            onSelect={(record: ConversionRecord) => {
              setResult(record.output, [], record.meta!);
              setShowHistory(false);
            }}
          />
        ) : null}
      </AnimatePresence>

      {/* Validation Rules Configurator */}
      <ValidationConfigurator
        open={showValidationConfigurator}
        onClose={() => setShowValidationConfigurator(false)}
        showValidateAction={true}
      />

      {/* Template Editor */}
      <TemplateEditor
        isOpen={showTemplateEditor}
        onClose={() => setShowTemplateEditor(false)}
        currentSampleData={pasteText || undefined}
      />

      {/* Keyboard shortcuts tooltip */}
      <div className="fixed bottom-4 right-4 z-40">
        <KeyboardShortcutsTooltip />
      </div>

      {/* Command Palette */}
      <CommandPalette
        open={showCommandPalette}
        onOpenChange={setShowCommandPalette}
        onConvert={handleConvert}
        onCopy={() => {
          if (review.reviewGateReason) {
            toast.error("Review required", review.reviewGateReason);
            return;
          }
          if (mdflowOutput) {
            output.handleCopy();
            toast.success("Copied to clipboard");
          }
        }}
        onExport={() => {
          if (review.reviewGateReason) {
            toast.error("Review required", review.reviewGateReason);
            return;
          }
          if (mdflowOutput) {
            output.handleDownload();
            toast.success("Downloaded spec.mdflow.md");
          }
        }}
        onTogglePreview={togglePreview}
        onShowHistory={() => setShowHistory(true)}
        onOpenTemplateEditor={() => setShowTemplateEditor(true)}
        onOpenValidation={() => setShowValidationConfigurator(true)}
        templates={templates}
        currentTemplate={format}
        onSelectTemplate={setFormat}
        hasOutput={Boolean(mdflowOutput)}
      />

      <ConversionFeedback
        visible={showFeedback}
        inputSource={inputSource}
        onDismiss={() => setShowFeedback(false)}
      />

      {/* Toast notifications */}
      <ToastContainer />
    </motion.div>
  );
}
