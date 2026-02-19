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
import {
  AlertCircle,
  Check,
  ChevronDown,
  Copy,
  Download,
  Eye,
  EyeOff,
  FileCode,
  FileSpreadsheet,
  FileText,
  GitCompare,
  History,
  KeyRound,
  Link2,
  RefreshCcw,
  RotateCcw,
  Save,
  Share2,
  ShieldCheck,
  Terminal,
  Zap,
} from "lucide-react";
import dynamic from "next/dynamic";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useShallow } from "zustand/react/shallow";
import { CommandPalette } from "./CommandPalette";
import { ConversionFeedback } from "@/components/ConversionFeedback";
import HistoryModal, { KeyboardShortcutsTooltip } from "./HistoryModal";
import { OnboardingTour } from "./OnboardingTour";
import { PreviewTable } from "./PreviewTable";
import { QuotaStatus } from "./QuotaStatus";
import { ShareButton } from "./ShareButton";
import { TechnicalAnalysis } from "./TechnicalAnalysis";
import { TemplateCards } from "./TemplateCards";
import { Select } from "./ui/Select";
import { OutputSkeleton } from "./ui/Skeleton";
import { toast, ToastContainer } from "./ui/Toast";
import { Tooltip } from "./ui/Tooltip";

const DiffViewer = dynamic(
  () => import("./DiffViewer").then((mod) => mod.DiffViewer),
  { ssr: false }
);

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

  const { data: templateList = [] } = useMDFlowTemplatesQuery();
  const templates = templateList;

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
    togglePreview: () => setShowPreview(!showPreview),
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
        {/* Left: Source & config — compact header with integrated controls */}
        <motion.div
          variants={stagger.item}
          className="flex flex-col min-h-0 h-full overflow-hidden"
        >
          <section className="p-0 flex flex-col h-full min-h-0 border border-white/10 bg-black/30 backdrop-blur-xl relative overflow-hidden rounded-xl sm:rounded-2xl">
            <div className="studio-grain" aria-hidden />
            <div className="relative z-10 flex flex-col h-full min-h-0">
              {/* Compact header with mode toggle */}
              <div className="flex items-center justify-between gap-2 px-3 sm:px-4 py-2.5 sm:py-3 border-b border-white/5 bg-white/2 shrink-0">
                <div
                  className="flex bg-black/40 rounded-lg border border-white/5 shrink-0"
                  data-tour="input-mode"
                >
                  {[
                    { key: "paste", label: "Paste" },
                    { key: "xlsx", label: "Excel" },
                    { key: "tsv", label: "TSV" },
                  ].map((m) => (
                    <button
                      key={m.key}
                      type="button"
                      onClick={() => {
                        setMode(m.key as "paste" | "xlsx" | "tsv");
                        setFile(null);
                      }}
                      className={`
                        px-3 sm:px-4 py-1.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md transition-all duration-200
                        ${mode === m.key
                          ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                          : "text-muted hover:text-white hover:bg-white/5"
                        }
                      `}
                    >
                      {m.label}
                    </button>
                  ))}
                </div>

                {/* Quick actions */}
                <div className="flex items-center gap-1.5">
                  <QuotaStatus compact onQuotaExceeded={() => {
                    toast.error("Daily quota exceeded", "Your token limit has been reached. Try again tomorrow.");
                  }} />
                  <Tooltip content={openaiKey ? "OpenAI Key Active" : "Set OpenAI API Key"}>
                    <button
                      type="button"
                      onClick={() => setShowApiKeyInput(!showApiKeyInput)}
                      className={`p-1.5 sm:p-2 rounded-lg border transition-all cursor-pointer ${openaiKey
                        ? "bg-green-500/15 hover:bg-green-500/25 border-green-500/30 hover:border-green-500/40 text-green-400"
                        : "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
                        }`}
                    >
                      <KeyRound className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                  <Tooltip content="Template Editor">
                    <button
                      type="button"
                      onClick={() => setShowTemplateEditor(true)}
                      className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                    >
                      <FileCode className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                  <Tooltip content="Validation Rules">
                    <button
                      type="button"
                      onClick={() => setShowValidationConfigurator(true)}
                      className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                    >
                      <ShieldCheck className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                </div>
              </div>

              {/* BYOK: OpenAI API Key input panel */}
              <AnimatePresence>
                {showApiKeyInput && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: "auto" }}
                    exit={{ opacity: 0, height: 0 }}
                    transition={{ duration: 0.2 }}
                    className="px-3 sm:px-4 py-2.5 border-b border-white/5 bg-black/20 shrink-0"
                  >
                    <div className="flex items-center gap-2">
                      <KeyRound className="w-3.5 h-3.5 text-white/40 shrink-0" />
                      {openaiKey ? (
                        <>
                          <div className="flex-1 bg-black/30 border border-green-500/20 rounded-lg px-3 py-1.5 text-xs text-green-400/80 font-mono">
                            sk-...{openaiKey.slice(-4)}
                          </div>
                          <button
                            type="button"
                            onClick={() => {
                              clearOpenaiKey();
                              setApiKeyDraft("");
                              toast.success("API key cleared");
                            }}
                            className="shrink-0 px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider rounded-lg border border-red-500/30 bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-all cursor-pointer"
                          >
                            Clear
                          </button>
                        </>
                      ) : (
                        <>
                          <input
                            type="password"
                            value={apiKeyDraft}
                            onChange={(e) => setApiKeyDraft(e.target.value)}
                            placeholder="sk-..."
                            className="flex-1 min-w-0 bg-black/30 border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white/90 placeholder-white/30 focus:border-accent-orange/30 focus:outline-none font-mono"
                            onKeyDown={(e) => {
                              if (e.key === "Enter" && apiKeyDraft.trim().length >= 10) {
                                setOpenaiKey(apiKeyDraft.trim());
                                toast.success("API key saved", "AI features are now enabled");
                              }
                            }}
                          />
                          <button
                            type="button"
                            onClick={() => {
                              if (apiKeyDraft.trim().length >= 10) {
                                setOpenaiKey(apiKeyDraft.trim());
                                toast.success("API key saved", "AI features are now enabled");
                              }
                            }}
                            disabled={apiKeyDraft.trim().length < 10}
                            className={`shrink-0 px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider rounded-lg border transition-all ${apiKeyDraft.trim().length >= 10
                              ? "border-accent-orange/30 bg-accent-orange/20 text-white hover:bg-accent-orange/30 cursor-pointer"
                              : "border-white/10 bg-white/5 text-white/30 cursor-not-allowed"
                              }`}
                          >
                            Save
                          </button>
                        </>
                      )}
                    </div>
                    <p className="text-[9px] text-white/35 mt-1.5 pl-5.5">
                      Your key is stored locally in your browser and sent to the backend as a per-request header. Never stored on our server.
                    </p>
                  </motion.div>
                )}
              </AnimatePresence>

              <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-3 sm:px-4 py-3 custom-scrollbar bg-black/3">
                <AnimatePresence mode="wait" initial={false}>
                  {error && (
                    <motion.div
                      initial={{ opacity: 0, y: -8 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -8 }}
                      className="mb-3 rounded-lg border border-accent-red/25 bg-accent-red/10 p-2.5 shrink-0"
                    >
                      <div className="flex items-start gap-2">
                        <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-accent-red" />
                        <div className="min-w-0 flex-1">
                          <p className="text-[9px] font-bold uppercase tracking-wider text-accent-red">
                            {mappedAppError?.title || "Request Failed"}
                          </p>
                          <p className="mt-1 text-[11px] text-accent-red/90">
                            {mappedAppError?.message || error}
                          </p>
                          {mappedAppError?.requestId ? (
                            <p className="mt-1 text-[10px] font-mono text-accent-red/70">
                              request_id={mappedAppError.requestId}
                            </p>
                          ) : null}
                          {mappedAppError?.retryable &&
                          (lastFailedAction === "preview" || lastFailedAction === "convert") ? (
                            <button
                              type="button"
                              onClick={() => void handleRetryFailedAction()}
                              disabled={loading || previewLoading}
                              className="mt-2 inline-flex h-7 items-center rounded-md border border-accent-red/35 bg-accent-red/15 px-2.5 text-[10px] font-bold uppercase tracking-wider text-accent-red hover:bg-accent-red/25"
                            >
                              {lastFailedAction === "convert" ? "Retry Convert" : "Retry Preview"}
                            </button>
                          ) : null}
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>

                {review.state.requiresReviewApproval && !review.state.reviewApproved && (
                  <div className="mb-3 flex items-center justify-between gap-3 rounded-lg border border-accent-gold/30 bg-accent-gold/10 px-3 py-2">
                    <div className="min-w-0 flex-1">
                      <p className="text-[10px] font-bold uppercase tracking-wider text-accent-gold/90">
                        Review Required
                      </p>
                      <p className="text-[10px] text-white/70">
                        Low-confidence mapping detected. Confirm review before sharing, exporting, or copying output.
                      </p>
                      {review.state.reviewRequiredColumns.length > 0 && (
                        <div className="mt-2 space-y-1">
                          <p className="text-[9px] uppercase tracking-wider text-white/60">
                            Review Columns ({review.state.reviewRequiredColumns.length})
                          </p>
                          <div className="flex flex-wrap gap-1.5">
                            {review.state.reviewRequiredColumns.map((column) => {
                              const checked = Boolean(review.state.reviewedColumns[column]);
                              return (
                                <button
                                  key={column}
                                  type="button"
                                  onClick={() =>
                                    review.setReviewedColumns({
                                      ...review.state.reviewedColumns,
                                      [column]: !checked,
                                    })
                                  }
                                  className={`rounded-md border px-2 py-1 text-[9px] font-bold uppercase tracking-wider transition-colors ${
                                    checked
                                      ? "border-green-500/40 bg-green-500/20 text-green-300"
                                      : "border-white/15 bg-black/20 text-white/70 hover:bg-black/35"
                                  }`}
                                >
                                  {checked ? "✓ " : ""}{column}
                                </button>
                              );
                            })}
                          </div>
                          <p className="text-[9px] text-white/55">
                           {review.reviewRemainingCount} column(s) remaining
                          </p>
                        </div>
                      )}
                    </div>
                    <div className="shrink-0 flex flex-col gap-1.5">
                      {review.state.reviewRequiredColumns.length > 0 && (
                        <button
                          type="button"
                          onClick={() =>
                            review.setReviewedColumns(
                              Object.fromEntries(
                                review.state.reviewRequiredColumns.map((column) => [column, true])
                              )
                            )
                          }
                          className="rounded-lg border border-white/20 bg-black/20 px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider text-white/85 hover:bg-black/35 transition-colors"
                        >
                          Mark All
                        </button>
                      )}
                      <button
                        type="button"
                        onClick={review.completeReview}
                        disabled={!canConfirmReview(review.state.reviewRequiredColumns, review.state.reviewedColumns)}
                        className={`rounded-lg border px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider transition-colors ${
                          !canConfirmReview(review.state.reviewRequiredColumns, review.state.reviewedColumns)
                            ? "border-white/15 bg-white/5 text-white/35 cursor-not-allowed"
                            : "border-accent-gold/40 bg-accent-gold/20 text-white hover:bg-accent-gold/30"
                        }`}
                      >
                        Confirm Review
                      </button>
                    </div>
                  </div>
                )}

                <AnimatePresence mode="wait">
                  {mode === "paste" ? (
                    <motion.div
                      key="paste"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="h-full flex flex-col min-h-0"
                    >
                      {/* Compact status bar */}
                      <div className="flex flex-wrap items-center gap-2 text-[9px] uppercase font-bold text-muted/50 mb-2 shrink-0">
                        {isGoogleSheetsURL(pasteText.trim()) && (
                          <span className="flex items-center gap-1 text-green-400/80 bg-green-400/10 px-2 py-0.5 rounded">
                            <Link2 className="w-3 h-3" />
                            Google Sheet
                          </span>
                        )}
                        {preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 &&
                          !isGoogleSheetsURL(pasteText.trim()) && (
                            <button
                              type="button"
                              onClick={() => setShowPreview(!showPreview)}
                              className="flex items-center gap-1 text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer bg-accent-orange/10 px-2 py-0.5 rounded"
                            >
                              {showPreview ? (
                                <EyeOff className="w-3 h-3" />
                              ) : (
                                <Eye className="w-3 h-3" />
                              )}
                              {showPreview ? "Hide" : "Show"} Preview
                            </button>
                          )}
                        {previewLoading && (
                          <span className="flex items-center gap-1 text-accent-orange/60">
                            <RefreshCcw className="w-3 h-3 animate-spin" />
                            Analyzing...
                          </span>
                        )}
                        {isGoogleSheetsURL(pasteText.trim()) && gsheetTabs.length > 0 && (
                          <button
                            type="button"
                            onClick={() => void previewGoogleSheetQuery.refetch()}
                            className="flex items-center gap-1 text-blue-400/70 hover:text-blue-400 transition-colors cursor-pointer bg-blue-400/10 px-2 py-0.5 rounded"
                          >
                            <RefreshCcw className="w-3 h-3" />
                            Refresh
                          </button>
                        )}
                        {isGoogleSheetsURL(pasteText.trim()) && gsheetLoading && (
                          <span className="flex items-center gap-1 text-blue-400/70">
                            <RefreshCcw className="w-3 h-3 animate-spin" />
                            Loading sheets...
                          </span>
                        )}
                      </div>

                      {isGoogleSheetsURL(pasteText.trim()) && gsheetTabs.length > 0 && (
                        <div className="mb-3 shrink-0">
                          <Select
                            value={selectedGid}
                            onValueChange={setSelectedGid}
                            options={gsheetTabs.map((tab) => ({
                              label: tab.title,
                              value: tab.gid,
                            }))}
                            placeholder="Choose sheet"
                            size="compact"
                            className="w-auto min-w-40"
                          />
                        </div>
                      )}
                      {isGoogleSheetsURL(pasteText.trim()) && (
                        <div className="mb-3 flex flex-wrap items-center gap-2 text-[9px] text-white/60 uppercase font-bold tracking-wider">
                          <label className="text-white/50">Range</label>
                          <input
                            type="text"
                            value={gsheetRange}
                            onChange={(e) => setGsheetRange(e.target.value)}
                            placeholder="A1:F200 or Sheet1!A1:F200"
                            className="h-7 px-2 rounded-md bg-black/40 border border-white/10 text-[10px] text-white/80 placeholder-white/30 focus:outline-none focus:border-accent-orange/40"
                          />
                        </div>
                      )}
                      {isInputGsheetUrl && (
                        <div className={`mb-3 flex flex-wrap items-center gap-2 rounded-lg px-3 py-2 text-[10px] text-white/70 border ${googleAuth.connected
                          ? "border-green-500/20 bg-green-500/10"
                          : "border-accent-orange/20 bg-accent-orange/10"
                          }`}>
                          {googleAuth.connected ? (
                            <Check className="h-3 w-3 shrink-0 text-green-400" />
                          ) : (
                            <AlertCircle className="h-3 w-3 shrink-0 text-accent-orange/80" />
                          )}
                          <span className="flex-1 min-w-40">
                            {googleAuth.connected
                              ? "Google connected. You can access private sheets without sharing."
                              : "Private sheet? Connect Google to access without sharing."}
                          </span>
                          {googleAuth.connected ? (
                            <button
                              type="button"
                              onClick={googleAuth.logout}
                              className="inline-flex items-center gap-1 rounded-md border border-white/10 bg-white/5 px-2 py-1 text-[9px] font-bold uppercase tracking-wider text-white/70 hover:text-white transition-colors"
                            >
                              Disconnect
                            </button>
                          ) : (
                            <button
                              type="button"
                              onClick={googleAuth.login}
                              disabled={googleAuth.loading}
                              className="inline-flex items-center gap-1 rounded-md border border-accent-orange/30 bg-accent-orange/20 px-2 py-1 text-[9px] font-bold uppercase tracking-wider text-white/90 hover:bg-accent-orange/30 transition-colors"
                            >
                              <KeyRound className="h-3 w-3" />
                              Connect Google
                            </button>
                          )}
                        </div>
                      )}

                      {/* Preview Table - Collapsible */}
                      <AnimatePresence>
                        {showPreview &&
                          preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 && (
                            <motion.div
                              initial={{ opacity: 0, height: 0 }}
                              animate={{ opacity: 1, height: "auto" }}
                              exit={{ opacity: 0, height: 0 }}
                              className="mb-3 shrink-0 max-h-[30vh] overflow-auto custom-scrollbar"
                              data-tour="preview-table"
                            >
                              <PreviewTable
                                preview={preview}
                                columnOverrides={columnOverrides}
                                onColumnOverride={review.handleColumnOverride}
                                needsReview={review.state.requiresReviewApproval}
                                reviewApproved={review.state.reviewApproved}
                                sourceUrl={isGoogleSheetsURL(pasteText.trim()) ? pasteText.trim() : undefined}
                                onSelectBlockRange={
                                  isGoogleSheetsURL(pasteText.trim())
                                    ? (range) => setGsheetRange(range)
                                    : undefined
                                }
                              />
                            </motion.div>
                          )}
                        {preview && preview.input_type === "markdown" && (
                          <motion.div
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: "auto" }}
                            exit={{ opacity: 0, height: 0 }}
                            className="mb-3 shrink-0"
                          >
                            <div className="rounded-lg border border-blue-500/20 bg-blue-500/5 px-3 py-2 flex items-center gap-2">
                              <FileText className="w-3.5 h-3.5 text-blue-400/80 shrink-0" />
                              <span className="text-[9px] font-bold text-blue-400/90 uppercase tracking-wider">
                                Markdown detected - passthrough mode
                              </span>
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>

                      <textarea
                        value={pasteText}
                        onChange={(e) => setPasteText(e.target.value)}
                        placeholder="Paste your table data here (TSV, CSV, or Google Sheets URL)…"
                        className="input flex-1 font-mono text-[12px] leading-relaxed resize-none border-white/5 bg-black/30 focus:bg-black/40 focus:border-accent-orange/30 custom-scrollbar min-h-30 rounded-lg"
                        aria-label="Paste TSV or CSV data"
                        data-tour="paste-area"
                      />
                    </motion.div>
                  ) : (
                    <motion.div
                      key={mode}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className={`h-full flex flex-col gap-4 min-h-0 ${!file ? "justify-center items-center" : "justify-start"
                        }`}
                    >
                      {/* File drop zone - centered when no file, shrink when file uploaded */}
                      <div
                        onDragOver={onDragOver}
                        onDragLeave={onDragLeave}
                        onDrop={onDrop}
                        className={`
                          relative rounded-2xl border-2 border-dashed transition-all duration-300 cursor-pointer w-full shrink-0
                          ${file ? "p-4" : "p-8 sm:p-12 max-w-lg"}
                          ${dragOver
                            ? "border-accent-orange/50 bg-accent-orange/10 scale-[1.02]"
                            : file
                              ? "border-accent-orange/30 bg-accent-orange/5"
                              : "border-white/20 hover:border-accent-orange/40 hover:bg-white/5"
                          }
                        `}
                      >
                        <input
                          type="file"
                          accept={mode === "tsv" ? ".tsv" : ".xlsx"}
                          onChange={handleFileChange}
                          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                          aria-label={
                            mode === "tsv"
                              ? "Upload TSV file"
                              : "Upload Excel file"
                          }
                        />
                        <div
                          className={`flex items-center gap-4 ${file ? "justify-start" : "justify-center flex-col"
                            }`}
                        >
                          <div
                            className={`
                              rounded-2xl flex items-center justify-center transition-all
                              ${file
                                ? "h-12 w-12 bg-accent-orange/20"
                                : "h-16 w-16 bg-white/10"
                              }
                            `}
                          >
                            {file ? (
                              <Check className="w-6 h-6 text-accent-orange" />
                            ) : (
                              <FileSpreadsheet
                                className={`w-8 h-8 ${dragOver
                                  ? "text-accent-orange"
                                  : "text-white/40"
                                  }`}
                              />
                            )}
                          </div>
                          <div className={file ? "text-left" : "text-center"}>
                            {file ? (
                              <>
                                <p className="text-sm font-bold text-white truncate max-w-62.5">
                                  {file.name}
                                </p>
                                <p className="text-xs text-white/50 font-mono">
                                  {(file.size / 1024).toFixed(1)} KB
                                </p>
                              </>
                            ) : (
                              <>
                                <p className="text-sm font-black text-white uppercase tracking-widest">
                                  {dragOver
                                    ? "Drop file here"
                                    : mode === "tsv"
                                      ? "Upload .TSV"
                                      : "Upload .XLSX"}
                                </p>
                                <p className="text-xs text-white/50 mt-1">
                                  Click or drag & drop
                                </p>
                              </>
                            )}
                          </div>
                        </div>
                      </div>

                      {/* Sheet selector */}
                      {mode === "xlsx" && sheets.length > 0 && (
                        <div className="shrink-0">
                          <Select
                            value={selectedSheet}
                            onValueChange={setSelectedSheet}
                            options={sheets.map((s) => ({
                              label: s,
                              value: s,
                            }))}
                            placeholder="Choose sheet"
                            size="compact"
                            className="w-auto min-w-30"
                          />
                        </div>
                      )}

                      {/* File Preview Table - takes remaining space */}
                      <AnimatePresence>
                        {file &&
                          preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 && (
                            <motion.div
                              initial={{ opacity: 0 }}
                              animate={{ opacity: 1 }}
                              exit={{ opacity: 0 }}
                              className="flex-1 min-h-0 flex flex-col"
                            >
                              <div className="flex items-center justify-between mb-2 shrink-0">
                                <span className="text-[10px] text-white/50 uppercase font-bold tracking-wider">
                                  Data Preview
                                </span>
                                <button
                                  type="button"
                                  onClick={() => setShowPreview(!showPreview)}
                                  className="flex items-center gap-1.5 text-[10px] text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer font-bold uppercase"
                                >
                                  {showPreview ? (
                                    <EyeOff className="w-3.5 h-3.5" />
                                  ) : (
                                    <Eye className="w-3.5 h-3.5" />
                                  )}
                                  {showPreview ? "Hide" : "Show"}
                                </button>
                              </div>
                              {showPreview && (
                                <div className="flex-1 min-h-0 overflow-auto custom-scrollbar rounded-lg border border-white/10">
                                  <PreviewTable
                                    preview={preview}
                                    columnOverrides={columnOverrides}
                                    onColumnOverride={review.handleColumnOverride}
                                    needsReview={review.state.requiresReviewApproval}
                                    reviewApproved={review.state.reviewApproved}
                                    onSelectBlockRange={
                                      isGoogleSheetsURL(pasteText.trim())
                                        ? (range) => setGsheetRange(range)
                                        : undefined
                                    }
                                  />
                                </div>
                              )}
                              {previewLoading && (
                                <div className="flex items-center gap-2 text-[10px] text-accent-orange/60 mt-2 shrink-0">
                                  <RefreshCcw className="w-3.5 h-3.5 animate-spin" />
                                  Loading preview...
                                </div>
                              )}
                            </motion.div>
                          )}
                      </AnimatePresence>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              {/* Compact footer with template, format & run */}
              <div className="px-3 sm:px-4 py-2.5 sm:py-3 border-t border-white/5 bg-white/2 shrink-0">
                <div
                  className="flex items-center gap-2 sm:gap-3"
                  data-tour="template-selector"
                >
                  {/* Template dropdown - collapsible on mobile */}
                  <div className="flex-1 min-w-48">
                    <TemplateCards
                      selected={format}
                      onSelect={setFormat}
                      compact
                    />
                  </div>

                  {/* Run button */}
                  <div className="shrink-0" data-tour="run-button">
                    {(() => {
                      const isDisabled =
                        loading ||
                        (mode === "paste" && !pasteText.trim()) ||
                        ((mode === "xlsx" || mode === "tsv") && !file);
                      const modKey = isMac() ? "⌘" : "Ctrl";

                      return (
                        <motion.button
                          type="button"
                          whileHover={!isDisabled ? { scale: 1.02 } : {}}
                          whileTap={!isDisabled ? { scale: 0.98 } : {}}
                          onClick={handleConvert}
                          disabled={isDisabled || loading}
                          className={`
                            h-9 sm:h-10 px-4 sm:px-6
                            uppercase tracking-wider text-[10px] sm:text-xs font-bold rounded-lg
                            flex items-center justify-center gap-2
                            transition-all duration-200
                            ${isDisabled
                              ? "bg-white/5 border border-white/10 text-white/30 cursor-not-allowed"
                              : "btn-primary shadow-lg shadow-accent-orange/20 cursor-pointer hover:shadow-xl hover:shadow-accent-orange/30"
                            }
                          `}
                          title={
                            isDisabled
                              ? mode === "paste"
                                ? "Paste data"
                                : "Upload file"
                              : `${modKey}+Enter`
                          }
                        >
                          {loading ? (
                            <RefreshCcw className="w-3.5 h-3.5 animate-spin" />
                          ) : (
                            <Zap className="w-3.5 h-3.5" />
                          )}
                          <span className="hidden xs:inline">
                            {loading ? "Running" : "Run"}
                          </span>
                        </motion.button>
                      );
                    })()}
                  </div>
                </div>

              </div>
            </div>
          </section>
        </motion.div>

        {/* Right: Output — compact and efficient */}
        <motion.div
          variants={stagger.item}
          className="flex flex-col min-h-0 h-full overflow-hidden"
          data-tour="output-panel"
        >
          <div className="p-0 flex flex-col h-full min-h-0 border border-white/10 bg-black/30 backdrop-blur-xl relative overflow-hidden rounded-xl sm:rounded-2xl">
            <div className="studio-grain" aria-hidden />
            <div className="relative z-10 flex flex-col h-full min-h-0">
              {/* Header - synced with Source section */}
              <div className="flex items-center justify-between gap-2 px-3 sm:px-4 py-2.5 sm:py-3 border-b border-white/5 bg-white/2 shrink-0">
                <div className="flex items-center gap-2 min-w-0">
                  <div className="flex gap-0.5 shrink-0">
                    <span
                      className={`w-1.5 h-1.5 rounded-full bg-red-400/80`}
                    />
                    <span className="w-1.5 h-1.5 rounded-full bg-yellow-400/80" />
                    <span className="w-1.5 h-1.5 rounded-full bg-green-400/80" />
                  </div>
                  <span className="text-[9px] sm:text-[10px] font-bold uppercase tracking-wider text-white/70">
                    Output
                  </span>
                  {review.state.requiresReviewApproval && !review.state.reviewApproved && (
                    <span className="text-[8px] uppercase tracking-wider font-bold px-2 py-0.5 rounded bg-accent-gold/20 border border-accent-gold/30 text-accent-gold/80">
                      Needs Review
                    </span>
                  )}
                  {review.state.requiresReviewApproval && review.state.reviewApproved && (
                    <span className="text-[8px] uppercase tracking-wider font-bold px-2 py-0.5 rounded bg-green-500/20 border border-green-500/30 text-green-300">
                      Reviewed
                    </span>
                  )}
                  {mdflowOutput && meta && (
                    <span className="text-[8px] sm:text-[9px] hidden sm:inline text-muted/50 font-mono">
                      {meta.total_rows || 0} rows
                    </span>
                  )}
                </div>
                {/* Action buttons - always visible, disabled when no output */}
                <div className="flex items-center gap-1.5 shrink-0">
                  <Tooltip content={review.reviewGateReason ?? (output.copied ? "Copied!" : "Copy")}>
                    <button
                      type="button"
                      onClick={output.handleCopy}
                      disabled={!mdflowOutput || Boolean(review.reviewGateReason)}
                      className={`p-1.5 sm:p-2 rounded-lg border transition-all ${mdflowOutput
                        ? "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
                        : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
                        }`}
                    >
                      {output.copied ? (
                        <Check className="w-3.5 h-3.5 text-accent-orange" />
                      ) : (
                        <Copy className="w-3.5 h-3.5" />
                      )}
                    </button>
                  </Tooltip>
                  {diff.snapshotA && (
                    <span
                      className="hidden sm:inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[8px] font-bold bg-rose-500/20 text-rose-300/90 border border-rose-500/30"
                      title="Version A saved"
                    >
                      A ✓
                    </span>
                  )}
                  {diff.snapshotB && (
                    <span
                      className="hidden sm:inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[8px] font-bold bg-emerald-500/20 text-emerald-300/90 border border-emerald-500/30"
                      title="Version B saved"
                    >
                      B ✓
                    </span>
                  )}
                  <Tooltip
                    content={
                      !diff.snapshotA
                        ? "Save as Version A (before)"
                        : !diff.snapshotB
                          ? "Save as Version B (after)"
                          : "Overwrite Version B"
                    }
                  >
                    <button
                      type="button"
                      onClick={() => diff.saveSnapshot(mdflowOutput)}
                      disabled={!mdflowOutput}
                      className={`p-1.5 sm:p-2 rounded-lg border transition-all ${mdflowOutput
                        ? "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
                        : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
                        }`}
                    >
                      <Save className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                  {diff.snapshotA && diff.snapshotB && (
                    <>
                      <Tooltip content="Compare (2 versions)">
                        <button
                          type="button"
                          onClick={() => diff.compareSnapshots()}
                          className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                        >
                          <GitCompare className="w-3.5 h-3.5" />
                        </button>
                      </Tooltip>
                      <Tooltip content="Clear snapshots">
                        <button
                          type="button"
                          onClick={() => diff.clearSnapshots()}
                          className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                        >
                          <RotateCcw className="w-3.5 h-3.5" />
                        </button>
                      </Tooltip>
                    </>
                  )}
                  <Tooltip content={review.reviewGateReason ?? "Export"}>
                    <button
                      type="button"
                      disabled={!mdflowOutput || Boolean(review.reviewGateReason)}
                      className={`p-1.5 sm:p-2 rounded-lg border transition-all ${mdflowOutput
                        ? "bg-accent-orange/90 hover:bg-accent-orange border-accent-orange/50 text-white"
                        : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
                        }`}
                      onClick={() => {
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
                    >
                      <Download className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                  <ShareButton
                    mdflowOutput={mdflowOutput}
                    template={format}
                    disabledReason={review.reviewGateReason}
                  />
                  {history.length > 0 && (
                    <Tooltip content="History">
                      <button
                        type="button"
                        onClick={() => setShowHistory(true)}
                        className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
                      >
                        <History className="w-3.5 h-3.5" />
                      </button>
                    </Tooltip>
                  )}
                </div>
              </div>

              {/* Output content */}
              <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-3 sm:px-4 py-3 custom-scrollbar">
                {loading ? (
                  <OutputSkeleton />
                ) : mdflowOutput ? (
                  <motion.pre
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 0.2 }}
                    className="whitespace-pre-wrap wrap-break-word font-mono text-[11px] sm:text-[12px] leading-relaxed text-white/90 selection:bg-accent-orange/30"
                  >
                    {mdflowOutput}
                  </motion.pre>
                ) : (
                  <div className="h-full flex flex-col items-center justify-center text-center py-6">
                    <div className="rounded-xl bg-white/5 border border-white/5 p-4 mb-3">
                      <Terminal className="w-8 h-8 text-white/20" />
                    </div>
                    <p className="text-[10px] font-bold uppercase tracking-widest text-white/40">
                      Output will appear here
                    </p>
                    <p className="text-[9px] text-muted/50 mt-1">
                      Paste data and run to generate
                    </p>
                  </div>
                )}
              </div>

              {/* Compact stats footer - only show when there's output */}
              {(mdflowOutput ||
                warnings.length > 0 ||
                aiSuggestions.length > 0) && (
                  <div className="border-t border-white/10 bg-white/2 px-3 sm:px-4 py-2 sm:py-2.5 shrink-0">
                    <TechnicalAnalysis
                      meta={meta}
                      warnings={warnings}
                      mdflowOutput={mdflowOutput}
                      aiSuggestions={aiSuggestions}
                      aiSuggestionsLoading={aiSuggestionsLoading}
                      aiSuggestionsError={aiSuggestionsError}
                      aiConfigured={aiConfigured}
                      onRetryAISuggestions={() => void handleGetAISuggestions()}
                    />
                  </div>
                )}
            </div>
          </div>
        </motion.div>
      </div>

      {/* Diff Viewer Modal */}
      <AnimatePresence>
        {diff.showDiff && diff.currentDiff && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => diff.setShowDiff(false)}
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
          >
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              onClick={(e) => e.stopPropagation()}
              className="bg-black/60 backdrop-blur-xl border border-white/20 rounded-2xl shadow-2xl max-w-4xl w-full max-h-[80vh] flex flex-col overflow-hidden"
            >
              <div className="flex items-center justify-between gap-4 px-6 py-3 border-b border-white/10 bg-white/3 shrink-0">
                <div className="flex items-center gap-3">
                  <span className="text-[10px] font-black uppercase tracking-[0.25em] text-white/80">
                    MDFlow Diff Viewer
                  </span>
                </div>
                <button
                  onClick={() => diff.setShowDiff(false)}
                  className="p-2 h-auto -mr-2 rounded-md hover:bg-white/20 transition-colors cursor-pointer text-white/60 hover:text-white"
                  aria-label="Close"
                >
                  ✕
                </button>
              </div>
              <div className="flex-1 min-h-0 overflow-auto custom-scrollbar">
                <DiffViewer diff={diff.currentDiff} />
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* History Modal */}
      <AnimatePresence>
        {showHistory && (
          <HistoryModal
            history={history}
            onClose={() => setShowHistory(false)}
            onSelect={(record: ConversionRecord) => {
              setResult(record.output, [], record.meta!);
              setShowHistory(false);
            }}
          />
        )}
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
        onTogglePreview={() => setShowPreview(!showPreview)}
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
