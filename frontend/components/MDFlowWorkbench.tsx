"use client";

import { useMediaQuery } from "@/hooks/useMediaQuery";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import {
  useAISuggestionsMutation,
  useConvertGoogleSheetMutation,
  useConvertPasteMutation,
  useConvertTSVMutation,
  useConvertXLSXMutation,
  useDiffMDFlowMutation,
  useGetXLSXSheetsMutation,
  useMDFlowTemplatesQuery,
  usePreviewPasteQuery,
  usePreviewTSVQuery,
  usePreviewXLSXQuery,
} from "@/lib/mdflowQueries";
import {
  ConversionRecord,
  useHistoryStore,
  useMDFlowStore,
} from "@/lib/mdflowStore";
import { AnimatePresence, motion } from "framer-motion";
import {
  AlertCircle,
  Boxes,
  Check,
  Copy,
  Database,
  Eye,
  EyeOff,
  FileCode,
  FileSpreadsheet,
  FileText,
  GitCompare,
  History,
  Link2,
  RefreshCcw,
  Save,
  ShieldCheck,
  Sparkles,
  Terminal,
  Zap,
} from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { DiffViewer } from "./DiffViewer";
import HistoryModal, { KeyboardShortcutsTooltip } from "./HistoryModal";
import { OnboardingTour } from "./OnboardingTour";
import { TemplateCards } from "./TemplateCards";
import { TemplateEditor } from "./TemplateEditor";
import { ValidationConfigurator } from "./ValidationConfigurator";
import {
  ExportDropdown,
  PreviewTable,
  ShareButton,
  TechnicalAnalysis,
} from "./index";
import { Select } from "./ui/Select";

const stagger = {
  container: {
    animate: { transition: { staggerChildren: 0.08, delayChildren: 0.12 } },
  },
  item: {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.5, ease: [0.16, 1, 0.3, 1] },
  },
};

const steps = [
  { title: "Import", label: "Paste or upload", icon: Database },
  { title: "Configure", label: "Sheet & template", icon: Terminal },
  { title: "Generate", label: "Run engine", icon: Zap },
];

export default function MDFlowWorkbench() {
  const {
    mode,
    pasteText,
    file,
    sheets,
    selectedSheet,
    template,
    mdflowOutput,
    warnings,
    meta,
    loading,
    error,
    preview,
    previewLoading,
    showPreview,
    columnOverrides,
    setMode,
    setPasteText,
    setFile,
    setSheets,
    setSelectedSheet,
    setTemplate,
    setResult,
    setLoading,
    setError,
    setPreview,
    setPreviewLoading,
    setShowPreview,
    setColumnOverride,
    aiSuggestions,
    aiSuggestionsLoading,
    aiSuggestionsError,
    aiConfigured,
    setAISuggestions,
    setAISuggestionsLoading,
    setAISuggestionsError,
    clearAISuggestions,
    reset,
  } = useMDFlowStore();

  const { addToHistory, history } = useHistoryStore();

  const [templates, setTemplates] = useState<string[]>(["default"]);
  const [copied, setCopied] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [showDiff, setShowDiff] = useState(false);
  const [previousOutput, setPreviousOutput] = useState<string>("");
  const [currentDiff, setCurrentDiff] = useState<any>(null);
  const [showHistory, setShowHistory] = useState(false);
  const [showValidationConfigurator, setShowValidationConfigurator] =
    useState(false);
  const [showTemplateEditor, setShowTemplateEditor] = useState(false);
  const [debouncedPasteText, setDebouncedPasteText] = useState("");
  const isNarrow = useMediaQuery("(max-width: 480px)");

  const { data: templateList = ["default"] } = useMDFlowTemplatesQuery();
  const getSheetsMutation = useGetXLSXSheetsMutation();
  const convertPasteMutation = useConvertPasteMutation();
  const convertXLSXMutation = useConvertXLSXMutation();
  const convertTSVMutation = useConvertTSVMutation();
  const convertGoogleSheetMutation = useConvertGoogleSheetMutation();
  const diffMDFlowMutation = useDiffMDFlowMutation();
  const aiSuggestionsMutation = useAISuggestionsMutation();

  const previewPasteQuery = usePreviewPasteQuery(
    debouncedPasteText,
    mode === "paste"
  );
  const previewTSVQuery = usePreviewTSVQuery(file, mode === "tsv");
  const previewXLSXQuery = usePreviewXLSXQuery(
    file,
    selectedSheet,
    mode === "xlsx"
  );

  useEffect(() => {
    setTemplates(templateList);
  }, [templateList]);

  // Reset store when leaving Studio so data is not shown when user comes back
  useEffect(() => {
    return () => reset();
  }, [reset]);

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

  useEffect(() => {
    if (mode !== "paste") return;
    if (!debouncedPasteText.trim()) {
      setPreview(null);
      setShowPreview(false);
      return;
    }
    if (previewPasteQuery.data) {
      setPreview(previewPasteQuery.data);
      setShowPreview(true);
    }
  }, [
    debouncedPasteText,
    mode,
    previewPasteQuery.data,
    setPreview,
    setShowPreview,
  ]);

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (!selectedFile) return;

      setFile(selectedFile);
      setLoading(true);
      setError(null);
      setPreview(null);

      if (/\.tsv$/i.test(selectedFile.name)) {
        setLoading(false);
        return;
      }

      try {
        const result = await getSheetsMutation.mutateAsync(selectedFile);
        setSheets(result.sheets);
        setSelectedSheet(result.active_sheet);
      } catch (error) {
        setError(error instanceof Error ? error.message : "Failed to read sheets");
      } finally {
        setLoading(false);
      }
    },
    [
      setFile,
      setLoading,
      setError,
      setSheets,
      setSelectedSheet,
      setPreview,
      getSheetsMutation,
    ]
  );

  useEffect(() => {
    if (mode === "xlsx" && previewXLSXQuery.data) {
      setPreview(previewXLSXQuery.data);
      setShowPreview(true);
    }
  }, [mode, previewXLSXQuery.data, setPreview, setShowPreview]);

  useEffect(() => {
    if (mode === "tsv" && previewTSVQuery.data) {
      setPreview(previewTSVQuery.data);
      setShowPreview(true);
    }
  }, [mode, previewTSVQuery.data, setPreview, setShowPreview]);

  useEffect(() => {
    const isLoading =
      (mode === "paste" && previewPasteQuery.isFetching) ||
      (mode === "xlsx" && previewXLSXQuery.isFetching) ||
      (mode === "tsv" && previewTSVQuery.isFetching);
    setPreviewLoading(isLoading);
  }, [
    mode,
    previewPasteQuery.isFetching,
    previewTSVQuery.isFetching,
    previewXLSXQuery.isFetching,
    setPreviewLoading,
  ]);

  const handleConvert = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      let result;
      let inputPreview = "";
      if (mode === "paste") {
        if (!pasteText.trim()) {
          setError("Missing source data");
          return;
        }

        // Check if it's a Google Sheets URL
        if (isGoogleSheetsURL(pasteText.trim())) {
          result = await convertGoogleSheetMutation.mutateAsync({
            url: pasteText.trim(),
            template,
          });
          inputPreview = `Google Sheet: ${pasteText.trim().slice(0, 60)}...`;
        } else {
          result = await convertPasteMutation.mutateAsync({
            pasteText,
            template,
          });
          inputPreview =
            pasteText.slice(0, 200) + (pasteText.length > 200 ? "..." : "");
        }
      } else if (mode === "xlsx") {
        if (!file) {
          setError("No file uploaded");
          return;
        }
        result = await convertXLSXMutation.mutateAsync({
          file,
          sheetName: selectedSheet,
          template,
        });
        inputPreview = `${file.name}${
          selectedSheet ? ` (${selectedSheet})` : ""
        }`;
      } else {
        if (!file) {
          setError("No file uploaded");
          return;
        }
        result = await convertTSVMutation.mutateAsync({
          file,
          template,
        });
        inputPreview = file.name;
      }

      if (result) {
        setResult(result.mdflow, result.warnings, result.meta);
        // Add to history
        addToHistory({
          mode,
          template,
          inputPreview,
          output: result.mdflow,
          meta: result.meta,
        });
      }
    } catch (error) {
      setError(error instanceof Error ? error.message : "Conversion failed");
    } finally {
      setLoading(false);
    }
  }, [
    mode,
    pasteText,
    file,
    selectedSheet,
    template,
    setLoading,
    setError,
    setResult,
    addToHistory,
    convertGoogleSheetMutation,
    convertPasteMutation,
    convertTSVMutation,
    convertXLSXMutation,
  ]);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(mdflowOutput);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [mdflowOutput]);

  const handleDownload = useCallback(() => {
    const blob = new Blob([mdflowOutput], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "spec.mdflow.md";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [mdflowOutput]);

  const handleGetAISuggestions = useCallback(async () => {
    if (!pasteText.trim() || aiSuggestionsLoading) return;

    setAISuggestionsLoading(true);
    setAISuggestionsError(null);
    clearAISuggestions();

    try {
      const result = await aiSuggestionsMutation.mutateAsync({
        pasteText,
        template,
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
    template,
    aiSuggestionsLoading,
    setAISuggestionsLoading,
    setAISuggestionsError,
    setAISuggestions,
    clearAISuggestions,
    aiSuggestionsMutation,
  ]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const isMac = navigator.platform.toUpperCase().indexOf("MAC") >= 0;
      const mod = isMac ? e.metaKey : e.ctrlKey;

      if (mod && e.key === "Enter") {
        e.preventDefault();
        handleConvert();
      } else if (mod && e.shiftKey && e.key.toLowerCase() === "c") {
        e.preventDefault();
        if (mdflowOutput) handleCopy();
      } else if (mod && e.key.toLowerCase() === "s" && mdflowOutput) {
        e.preventDefault();
        handleDownload();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [mdflowOutput, handleConvert, handleCopy, handleDownload]);

  const currentStep = mdflowOutput ? 3 : file || pasteText.trim() ? 2 : 1;

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const f = e.dataTransfer.files?.[0];
      if (!f) return;

      if (mode === "tsv" && /\.tsv$/i.test(f.name)) {
        setFile(f);
        setError(null);
        setPreview(null);
        setLoading(true);
        setLoading(false);
        return;
      }

      if (mode === "xlsx" && /\.(xlsx|xls)$/i.test(f.name)) {
        setFile(f);
        setLoading(true);
        setError(null);
        setPreview(null);
        getSheetsMutation
          .mutateAsync(f)
          .then((result) => {
            setSheets(result.sheets);
            setSelectedSheet(result.active_sheet);
          })
          .catch((error) => {
            setError(error instanceof Error ? error.message : "Failed to read sheets");
          })
          .finally(() => {
            setLoading(false);
          });
      }
    },
    [
      mode,
      setFile,
      setLoading,
      setError,
      setSheets,
      setSelectedSheet,
      setPreview,
      getSheetsMutation,
    ]
  );

  return (
    <motion.div
      variants={stagger.container}
      initial="initial"
      animate="animate"
      className="flex flex-col gap-6 sm:gap-8 lg:gap-12 relative"
    >
      {/* Onboarding Tour */}
      <OnboardingTour />

      {/* Step flow strip */}
      <motion.div
        variants={stagger.item}
        className="flex justify-center"
        aria-label="Workflow steps"
        data-tour="welcome"
      >
        <div className="inline-flex items-center rounded-xl sm:rounded-2xl border border-white/10 bg-white/2 p-1 sm:p-1.5 shadow-lg">
          {steps.map((s, i) => {
            const StepIcon = s.icon;
            const active = currentStep >= i + 1;
            return (
              <div key={s.title} className="flex items-center">
                <div
                  className={`
                    flex items-center gap-1.5 sm:gap-2.5 px-3 sm:px-5 py-2 sm:py-2.5 rounded-lg sm:rounded-xl transition-all duration-300
                    ${
                      active
                        ? "bg-accent-orange/15 text-accent-orange border border-accent-orange/30 shadow-[0_0_20px_rgba(242,123,47,0.15)]"
                        : "text-muted/70 border border-transparent"
                    }
                  `}
                >
                  <div
                    className={`
                      flex h-6 w-6 sm:h-8 sm:w-8 items-center justify-center rounded-md sm:rounded-lg
                      ${active ? "bg-accent-orange/20" : "bg-white/5"}
                    `}
                  >
                    <StepIcon className="h-3.5 w-3.5 sm:h-4 sm:w-4" />
                  </div>
                  <div className="hidden sm:block">
                    <span className="text-[10px] font-black uppercase tracking-widest block">
                      {s.title}
                    </span>
                    <span className="text-[9px] text-muted/80 font-medium uppercase tracking-wider">
                      {s.label}
                    </span>
                  </div>
                </div>
                {i < steps.length - 1 && (
                  <div
                    className={`
                      mx-0.5 h-px w-6 sm:w-8 rounded-full transition-colors
                      ${active ? "bg-accent-orange/40" : "bg-white/10"}
                    `}
                  />
                )}
              </div>
            );
          })}
        </div>
      </motion.div>

      {/* Main workspace: fixed height on lg so only the two content areas scroll, footer stays at bottom */}
      <div className="grid grid-cols-1 lg:grid-cols-[1.15fr_1fr] gap-5 sm:gap-6 lg:gap-10 items-stretch min-h-[360px] sm:min-h-[420px] h-full sm:h-[70vh] md:h-full lg:h-screen lg:min-h-[80vh]">
        {/* Left: Source & config — scrollable body, footer (Template + Run) stays at bottom */}
        <motion.div
          variants={stagger.item}
          className="flex flex-col min-h-0 h-full overflow-hidden"
        >
          <section className="surface p-0 flex flex-col h-full min-h-0 border-white/10 bg-white/2 relative overflow-hidden rounded-2xl">
            <div className="studio-grain" aria-hidden />
            <div className="relative z-10 flex flex-col h-full min-h-0">
              <div className="flex items-center justify-between gap-4 p-5 sm:p-6 lg:p-8 border-b border-white/5 bg-white/2 shrink-0">
                <div className="flex items-center gap-2 sm:gap-3 min-w-0">
                  <div className="h-7 w-7 sm:h-8 sm:w-8 lg:h-9 lg:w-9 rounded-lg sm:rounded-xl bg-accent-orange/10 flex items-center justify-center border border-accent-orange/20 shadow-[0_0_24px_rgba(242,123,47,0.08)] shrink-0">
                    <Database className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-accent-orange" />
                  </div>
                  <div className="min-w-0">
                    <h2 className="text-xs sm:text-sm font-black uppercase tracking-widest text-white">
                      Source
                    </h2>
                    <p className="text-[8px] sm:text-[9px] text-muted/80 uppercase tracking-wider mt-0.5">
                      Paste or upload
                    </p>
                  </div>
                </div>
                <div
                  className="flex bg-black/50 p-1.5 rounded-lg sm:rounded-xl border border-white/5 shadow-inner shrink-0"
                  data-tour="input-mode"
                >
                  <button
                    type="button"
                    onClick={() => {
                      setMode("paste");
                      setFile(null);
                    }}
                    className={`
                      px-4 sm:px-5 py-2 sm:py-2.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md sm:rounded-lg transition-all duration-200
                      ${
                        mode === "paste"
                          ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                          : "text-muted hover:text-white hover:bg-white/5"
                      }
                    `}
                  >
                    Paste
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setMode("xlsx");
                      setFile(null);
                    }}
                    className={`
                      px-4 sm:px-5 py-2 sm:py-2.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md sm:rounded-lg transition-all duration-200
                      ${
                        mode === "xlsx"
                          ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                          : "text-muted hover:text-white hover:bg-white/5"
                      }
                    `}
                  >
                    Excel
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setMode("tsv");
                      setFile(null);
                    }}
                    className={`
                      px-4 sm:px-5 py-2 sm:py-2.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md sm:rounded-lg transition-all duration-200
                      ${
                        mode === "tsv"
                          ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                          : "text-muted hover:text-white hover:bg-white/5"
                      }
                    `}
                  >
                    TSV
                  </button>
                </div>
              </div>

              <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-5 sm:px-6 lg:px-8 py-4 sm:py-5 lg:py-6 custom-scrollbar bg-black/3">
                <AnimatePresence mode="wait" initial={false}>
                  {error && (
                    <motion.div
                      initial={{ opacity: 0, y: -8 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -8 }}
                      className="mb-6 p-4 bg-accent-red/10 border border-accent-red/25 rounded-xl flex items-center gap-3 text-accent-red text-[10px] font-black uppercase tracking-widest shrink-0"
                    >
                      <AlertCircle className="w-4 h-4 shrink-0" /> {error}
                    </motion.div>
                  )}
                </AnimatePresence>

                <AnimatePresence mode="wait">
                  {mode === "paste" ? (
                    <motion.div
                      key="paste"
                      initial={{ opacity: 0, x: -8 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0, x: 8 }}
                      transition={{ duration: 0.25 }}
                      className="h-full flex flex-col min-h-0"
                    >
                      <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1 text-[10px] uppercase font-bold text-muted/60 mb-3 shrink-0">
                        <span>Paste TSV / CSV / URL</span>
                        <div className="flex items-center gap-2">
                          {isGoogleSheetsURL(pasteText.trim()) && (
                            <span className="flex items-center gap-1.5 text-green-400/80 font-medium">
                              <Link2 className="w-3 h-3" />
                              Google Sheet detected
                            </span>
                          )}
                          {preview &&
                            preview.input_type === "table" &&
                            preview.headers.length > 0 &&
                            !isGoogleSheetsURL(pasteText.trim()) && (
                              <button
                                type="button"
                                onClick={() => setShowPreview(!showPreview)}
                                className="flex items-center gap-1 text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer"
                              >
                                {showPreview ? (
                                  <EyeOff className="w-3 h-3" />
                                ) : (
                                  <Eye className="w-3 h-3" />
                                )}
                                {showPreview ? "Hide" : "Show"} Preview
                              </button>
                            )}
                          {!isGoogleSheetsURL(pasteText.trim()) && (
                            <span className="text-accent-orange/50 font-medium sm:text-right">
                              Tab-separated, comma, or Google Sheets URL
                            </span>
                          )}
                        </div>
                      </div>

                      {/* Preview Table */}
                      <AnimatePresence>
                        {showPreview &&
                          preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 && (
                            <motion.div
                              initial={{ opacity: 0, height: 0 }}
                              animate={{ opacity: 1, height: "auto" }}
                              exit={{ opacity: 0, height: 0 }}
                              className="mb-4 shrink-0"
                              data-tour="preview-table"
                            >
                              <PreviewTable
                                preview={preview}
                                columnOverrides={columnOverrides}
                                onColumnOverride={setColumnOverride}
                              />
                            </motion.div>
                          )}
                        {preview && preview.input_type === "markdown" && (
                          <motion.div
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: "auto" }}
                            exit={{ opacity: 0, height: 0 }}
                            className="mb-4 shrink-0"
                          >
                            <div className="rounded-xl border border-blue-500/20 bg-blue-500/5 p-3 flex items-center gap-3">
                              <FileText className="w-4 h-4 text-blue-400/80 shrink-0" />
                              <div>
                                <span className="text-[10px] font-bold text-blue-400/90 uppercase tracking-wider">
                                  Markdown/Prose detected
                                </span>
                                <p className="text-[9px] text-blue-400/60 mt-0.5">
                                  Content will be passed through with minimal
                                  formatting. No table preview available.
                                </p>
                              </div>
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>

                      <textarea
                        value={pasteText}
                        onChange={(e) => setPasteText(e.target.value)}
                        placeholder="Paste your table data here…"
                        className="input flex-1 font-mono text-[13px] leading-relaxed resize-none border-white/5 bg-black/20 focus:bg-black/30 focus:border-accent-orange/30 custom-scrollbar min-h-[200px] rounded-xl"
                        aria-label="Paste TSV or CSV data"
                        data-tour="paste-area"
                      />

                      {previewLoading && (
                        <div className="absolute bottom-4 right-4 flex items-center gap-2 text-[10px] text-accent-orange/60">
                          <RefreshCcw className="w-3 h-3 animate-spin" />
                          Analyzing...
                        </div>
                      )}
                    </motion.div>
                  ) : (
                    <motion.div
                      key={mode}
                      initial={{ opacity: 0, x: 8 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0, x: -8 }}
                      transition={{ duration: 0.25 }}
                      className="h-full flex flex-col justify-center gap-6 min-h-0"
                    >
                      <div
                        onDragOver={(e) => {
                          e.preventDefault();
                          setDragOver(true);
                        }}
                        onDragLeave={() => setDragOver(false)}
                        onDrop={onDrop}
                        className={`
                          relative rounded-2xl border-2 border-dashed p-6 sm:p-8 lg:p-12 text-center transition-all duration-300 cursor-pointer
                          overflow-hidden
                          ${
                            dragOver
                              ? "border-accent-orange/50 bg-accent-orange/10 scale-[1.02]"
                              : "border-white/10 hover:border-white/20 hover:bg-white/3"
                          }
                        `}
                      >
                        <input
                          type="file"
                          accept={mode === "tsv" ? ".tsv" : ".xlsx,.xls"}
                          onChange={handleFileChange}
                          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                          aria-label={
                            mode === "tsv"
                              ? "Upload TSV file"
                              : "Upload Excel file"
                          }
                        />
                        <motion.div
                          animate={{ scale: dragOver ? 1.05 : 1 }}
                          className="flex flex-col items-center gap-4"
                        >
                          <div
                            className={`
                              h-16 w-16 rounded-2xl flex items-center justify-center transition-all duration-300
                              ${
                                dragOver
                                  ? "bg-accent-orange/20"
                                  : "bg-white/5 group-hover:bg-white/10"
                              }
                            `}
                          >
                            <FileSpreadsheet
                              className={`w-8 h-8 transition-colors ${
                                dragOver
                                  ? "text-accent-orange"
                                  : "text-muted/50"
                              }`}
                            />
                          </div>
                          <div>
                            <p className="text-sm font-black text-white uppercase tracking-widest">
                              {dragOver
                                ? "Drop file here"
                                : mode === "tsv"
                                ? "Upload .tsv"
                                : "Upload .xlsx or .xls"}
                            </p>
                            <p className="text-[10px] text-muted mt-1.5 uppercase font-medium">
                              Click or drag & drop
                            </p>
                          </div>
                        </motion.div>
                      </div>

                      {file && (
                        <motion.div
                          initial={{ opacity: 0, y: 8 }}
                          animate={{ opacity: 1, y: 0 }}
                          className="flex items-center gap-4 p-4 rounded-xl bg-accent-orange/5 border border-accent-orange/20 shrink-0"
                        >
                          <Check className="w-4 h-4 text-accent-orange shrink-0" />
                          <span className="text-[11px] font-bold text-white uppercase truncate flex-1">
                            {file.name}
                          </span>
                          <span className="text-[10px] text-muted font-mono">
                            {(file.size / 1024).toFixed(1)} KB
                          </span>
                        </motion.div>
                      )}

                      {mode === "xlsx" && sheets.length > 0 && (
                        <div className="space-y-3 shrink-0">
                          <label className="label mb-2">Sheet</label>
                          <Select
                            value={selectedSheet}
                            onValueChange={setSelectedSheet}
                            options={sheets.map((s) => ({
                              label: s,
                              value: s,
                            }))}
                            placeholder="Choose sheet"
                            className="h-12"
                          />
                        </div>
                      )}

                      {/* File Preview Table */}
                      <AnimatePresence>
                        {file &&
                          preview &&
                          preview.input_type === "table" &&
                          preview.headers.length > 0 && (
                            <motion.div
                              initial={{ opacity: 0, height: 0 }}
                              animate={{ opacity: 1, height: "auto" }}
                              exit={{ opacity: 0, height: 0 }}
                              className="shrink-0"
                            >
                              <div className="flex items-center justify-between mb-2">
                                <span className="text-[10px] text-white/50 uppercase font-bold tracking-wider">
                                  Data Preview
                                </span>
                                <button
                                  type="button"
                                  onClick={() => setShowPreview(!showPreview)}
                                  className="flex items-center gap-1 text-[9px] text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer font-bold uppercase"
                                >
                                  {showPreview ? (
                                    <EyeOff className="w-3 h-3" />
                                  ) : (
                                    <Eye className="w-3 h-3" />
                                  )}
                                  {showPreview ? "Hide" : "Show"}
                                </button>
                              </div>
                              {showPreview && (
                                <PreviewTable
                                  preview={preview}
                                  columnOverrides={columnOverrides}
                                  onColumnOverride={setColumnOverride}
                                />
                              )}
                              {previewLoading && (
                                <div className="flex items-center gap-2 text-[10px] text-accent-orange/60 mt-2">
                                  <RefreshCcw className="w-3 h-3 animate-spin" />
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

              <div className="px-5 sm:px-6 lg:px-8 py-5 sm:py-6 lg:py-8 border-t border-white/5 bg-white/2 flex flex-col gap-4 shrink-0">
                <div className="w-full space-y-2" data-tour="template-selector">
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between mb-2">
                    <label className="label mb-0">Template</label>
                    <div className="flex flex-col xs:flex-row gap-2 xs:gap-2">
                      <button
                        type="button"
                        onClick={() => setShowTemplateEditor(true)}
                        className="flex items-center justify-center gap-1.5 px-3 py-2 sm:px-2.5 sm:py-1.5 rounded-lg text-[9px] font-bold uppercase tracking-wider bg-white/5 hover:bg-accent-orange/20 border border-white/10 hover:border-accent-orange/30 text-white/70 hover:text-accent-orange transition-all touch-manipulation min-w-0"
                        title="Create custom templates"
                      >
                        <FileCode className="w-3 h-3 shrink-0" />
                        <span className="truncate">
                          {isNarrow ? "Editor" : "Template Editor"}
                        </span>
                      </button>
                      <button
                        type="button"
                        onClick={() => setShowValidationConfigurator(true)}
                        className="flex items-center justify-center gap-1.5 px-3 py-2 sm:px-2.5 sm:py-1.5 rounded-lg text-[9px] font-bold uppercase tracking-wider bg-white/5 hover:bg-accent-orange/20 border border-white/10 hover:border-accent-orange/30 text-white/70 hover:text-accent-orange transition-all touch-manipulation min-w-0"
                        title="Configure validation rules"
                      >
                        <ShieldCheck className="w-3 h-3 shrink-0" />
                        <span className="truncate">
                          {isNarrow ? "Rules" : "Validation rules"}
                        </span>
                      </button>
                    </div>
                  </div>
                  <TemplateCards
                    templates={templates}
                    selected={template}
                    onSelect={setTemplate}
                    compact
                  />
                </div>
                <div className="flex justify-end">
                  <div className="relative group/button" data-tour="run-button">
                    {(() => {
                      const isDisabled =
                        loading ||
                        (mode === "paste" && !pasteText.trim()) ||
                        ((mode === "xlsx" || mode === "tsv") && !file);
                      const isMac =
                        typeof navigator !== "undefined" &&
                        navigator.platform.toUpperCase().indexOf("MAC") >= 0;
                      const modKey = isMac ? "⌘" : "Ctrl";

                      return (
                        <>
                          <motion.button
                            type="button"
                            whileHover={!isDisabled ? { scale: 1.02 } : {}}
                            whileTap={!isDisabled ? { scale: 0.98 } : {}}
                            onClick={handleConvert}
                            disabled={isDisabled || loading}
                            className={`
                              h-12 w-full sm:w-auto min-w-[140px] sm:min-w-[160px] px-6 sm:px-8 
                              uppercase tracking-[0.18em] rounded-xl shrink-0
                              flex items-center justify-center gap-2.5
                              transition-all duration-200
                              ${
                                isDisabled
                                  ? "bg-white/5 border border-white/10 text-white/30 cursor-not-allowed shadow-none"
                                  : "btn-primary shadow-xl shadow-accent-orange/25 group cursor-pointer hover:shadow-2xl hover:shadow-accent-orange/30"
                              }
                            `}
                            title={
                              isDisabled
                                ? mode === "paste"
                                  ? "Paste data to continue"
                                  : "Upload a file to continue"
                                : `Run conversion (${modKey}+Enter)`
                            }
                          >
                            {loading ? (
                              <>
                                <RefreshCcw
                                  className="w-4 h-4 animate-spin"
                                  aria-hidden
                                />
                                <span>Running</span>
                              </>
                            ) : (
                              <>
                                <Zap
                                  className={`w-4 h-4 transition-transform ${
                                    isDisabled
                                      ? ""
                                      : "group-hover:scale-110 group-hover:rotate-12"
                                  }`}
                                />
                                <span>Run</span>
                              </>
                            )}
                          </motion.button>
                          {!isDisabled && (
                            <div className="absolute -top-8 right-0 opacity-0 group-hover/button:opacity-100 transition-opacity duration-200 pointer-events-none">
                              <div className="bg-black/90 backdrop-blur-sm border border-white/20 rounded-lg px-2 py-1 text-[9px] font-mono text-white/70 whitespace-nowrap">
                                {modKey}+Enter
                              </div>
                            </div>
                          )}
                        </>
                      );
                    })()}
                  </div>
                </div>
              </div>
            </div>
          </section>
        </motion.div>

        {/* Right: Output — scrollable pre, footer (Stats) stays at bottom */}
        <motion.div
          variants={stagger.item}
          className="flex flex-col min-h-0 h-full overflow-hidden"
          data-tour="output-panel"
        >
          <div className="code-editor p-0 border-white/10 h-full min-h-0 flex flex-col shadow-2xl bg-black/40 backdrop-blur-2xl relative overflow-hidden rounded-2xl">
            <div className="studio-grain" aria-hidden />
            <div className="relative z-10 flex flex-col h-full min-h-0">
              <div className="flex items-center justify-between gap-3 p-3 sm:p-4 border-b border-white/10 bg-white/6 shrink-0">
                <div className="flex items-center gap-2 min-w-0">
                  <div className="flex gap-1 shrink-0">
                    <span className="w-1.5 h-1.5 rounded-full bg-accent-red/80 shadow-[0_0_4px_rgba(239,68,68,0.4)]" />
                    <span className="w-1.5 h-1.5 rounded-full bg-accent-gold/80 shadow-[0_0_4px_rgba(251,191,36,0.4)]" />
                    <span className="w-1.5 h-1.5 rounded-full bg-accent-orange/80 shadow-[0_0_4px_rgba(242,123,47,0.4)]" />
                  </div>
                  <span className="text-[9px] font-black uppercase tracking-[0.25em] text-white/80 font-mono">
                    Output
                  </span>
                </div>
                {mdflowOutput && (
                  <div className="flex items-center gap-3 shrink-0">
                    <button
                      type="button"
                      onClick={handleCopy}
                      className="flex items-center justify-center w-8 h-8 shrink-0 rounded-lg bg-white/10 hover:bg-white/20 text-white/80 hover:text-white transition-all border border-white/10 hover:border-white/20 cursor-pointer"
                      title={copied ? "Copied" : "Copy (⌘+Shift+C)"}
                      aria-label={copied ? "Copied" : "Copy to clipboard"}
                    >
                      {copied ? (
                        <Check className="w-3 h-3 text-accent-orange" />
                      ) : (
                        <Copy className="w-3 h-3" />
                      )}
                    </button>
                    {previousOutput && (
                      <button
                        type="button"
                        onClick={async () => {
                          const diff = await diffMDFlowMutation.mutateAsync({
                            before: previousOutput,
                            after: mdflowOutput,
                          });
                          setCurrentDiff(diff);
                          setShowDiff(true);
                        }}
                        className="flex items-center justify-center h-8 px-3 rounded-lg bg-white/10 hover:bg-white/20 text-white/80 hover:text-white transition-all border border-white/10 hover:border-white/20 text-[9px] font-bold uppercase tracking-wider cursor-pointer shrink-0"
                        title="Compare with saved output"
                      >
                        <GitCompare className="w-3 h-3 shrink-0" />
                      </button>
                    )}
                    <button
                      type="button"
                      onClick={() => {
                        setPreviousOutput(mdflowOutput);
                        setCopied(true);
                        setTimeout(() => setCopied(false), 1500);
                      }}
                      className="flex items-center justify-center h-8 px-3 rounded-lg bg-white/10 hover:bg-white/20 text-white/80 hover:text-white transition-all border border-white/10 hover:border-white/20 text-[9px] font-bold uppercase tracking-wider cursor-pointer shrink-0"
                      title="Save current output for comparison"
                    >
                      <Save className="w-3 h-3 shrink-0" />
                    </button>
                    <button
                      type="button"
                      onClick={handleGetAISuggestions}
                      disabled={!pasteText.trim() || aiSuggestionsLoading}
                      className={`flex items-center justify-center h-8 px-3 rounded-lg text-[9px] font-bold uppercase tracking-wider transition-all cursor-pointer shrink-0 gap-1.5 ${
                        !pasteText.trim() || aiSuggestionsLoading
                          ? "bg-white/5 text-white/30 cursor-not-allowed border border-white/10"
                          : "bg-purple-500/20 hover:bg-purple-500/30 text-purple-300 hover:text-purple-200 border border-purple-500/30 hover:border-purple-500/40"
                      }`}
                      title="Get AI suggestions for quality improvements"
                    >
                      <Sparkles
                        className={`w-3 h-3 shrink-0 ${
                          aiSuggestionsLoading ? "animate-pulse" : ""
                        }`}
                      />
                      <span className="hidden sm:inline">AI</span>
                    </button>
                    <ShareButton
                      mdflowOutput={mdflowOutput}
                      template={template}
                    />
                    <ExportDropdown
                      mdflowOutput={mdflowOutput}
                    />
                    {history.length > 0 && (
                      <button
                        type="button"
                        onClick={() => setShowHistory(true)}
                        className="flex items-center justify-center w-8 h-8 shrink-0 rounded-lg bg-white/5 hover:bg-white/10 text-white/50 hover:text-white/80 transition-all border border-white/5 hover:border-white/10 cursor-pointer"
                        title="View history"
                      >
                        <History className="w-3 h-3" />
                      </button>
                    )}
                  </div>
                )}
              </div>

              <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-5 sm:px-6 lg:px-8 py-4 sm:py-5 lg:py-6 custom-scrollbar">
                {mdflowOutput ? (
                  <pre className="whitespace-pre-wrap wrap-break-word font-mono text-xs sm:text-[13px] leading-relaxed text-white/95 selection:bg-accent-orange/30 selection:text-white">
                    {mdflowOutput}
                  </pre>
                ) : (
                  <div className="h-full min-h-[240px] sm:min-h-[280px] flex flex-col items-center justify-center text-center py-8">
                    <motion.div
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      transition={{ delay: 0.2 }}
                      className="rounded-2xl bg-white/5 border border-white/5 p-6 sm:p-8 mb-4 sm:mb-6"
                    >
                      <Terminal className="w-10 h-10 text-white/25" />
                    </motion.div>
                    <p className="text-[11px] font-black uppercase tracking-[0.25em] text-white/50">
                      Output will appear here
                    </p>
                    <p className="text-[10px] text-muted/70 mt-1.5 uppercase tracking-wider">
                      Run the engine to generate MDFlow
                    </p>
                  </div>
                )}
              </div>

              <div className="border-t border-white/10 bg-white/2 px-4 sm:px-5 lg:px-6 py-4 sm:py-5 shrink-0 min-h-[100px] sm:min-h-[120px] flex flex-col justify-center">
                <TechnicalAnalysis
                  meta={meta}
                  warnings={warnings}
                  mdflowOutput={mdflowOutput}
                  aiSuggestions={aiSuggestions}
                  aiSuggestionsLoading={aiSuggestionsLoading}
                  aiSuggestionsError={aiSuggestionsError}
                  aiConfigured={aiConfigured}
                />
              </div>
            </div>
          </div>
        </motion.div>
      </div>

      <motion.div variants={stagger.item} className="flex justify-center gap-3">
        <Link
          href="/batch"
          className="inline-flex items-center gap-2 text-[8px] sm:text-[9px] font-bold text-muted/50 hover:text-accent-orange uppercase tracking-[0.25em] sm:tracking-[0.3em] bg-white/3 hover:bg-white/5 border border-white/5 hover:border-accent-orange/20 px-3 sm:px-4 py-1.5 sm:py-2 rounded-full transition-all"
        >
          <Boxes className="w-3 h-3" />
          Batch Mode
        </Link>
        <div className="inline-flex items-center gap-2 text-[8px] sm:text-[9px] font-bold text-muted/50 uppercase tracking-[0.3em] sm:tracking-[0.35em] bg-white/3 border border-white/5 px-3 sm:px-5 py-1.5 sm:py-2 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-accent-orange/80 animate-pulse" />
          MDFlow Studio
        </div>
      </motion.div>

      {/* Diff Viewer Modal */}
      <AnimatePresence>
        {showDiff && currentDiff && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setShowDiff(false)}
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
                  onClick={() => setShowDiff(false)}
                  className="p-2 h-auto -mr-2 rounded-md hover:bg-white/20 transition-colors cursor-pointer text-white/60 hover:text-white"
                  aria-label="Close"
                >
                  ✕
                </button>
              </div>
              <div className="flex-1 min-h-0 overflow-auto custom-scrollbar">
                <DiffViewer
                  diff={currentDiff}
                />
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
    </motion.div>
  );
}
