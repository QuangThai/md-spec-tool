"use client";

import {
  convertGoogleSheet,
  convertPaste,
  convertTSV,
  convertXLSX,
  diffMDFlow,
  getMDFlowTemplates,
  getXLSXSheets,
  isGoogleSheetsURL,
  previewPaste,
  previewTSV,
  previewXLSX,
} from "@/lib/mdflowApi";
import { useHistoryStore, useMDFlowStore } from "@/lib/mdflowStore";
import { AnimatePresence, motion } from "framer-motion";
import {
  Activity,
  AlertCircle,
  ArrowRight,
  BookOpen,
  Boxes,
  Check,
  Clock,
  Copy,
  Database,
  Download,
  Eye,
  EyeOff,
  ExternalLink,
  FileCode,
  FileJson,
  FileSpreadsheet,
  FileText,
  GitCompare,
  History,
  Keyboard,
  Link2,
  RefreshCcw,
  Save,
  Share2,
  ShieldCheck,
  Table,
  Terminal,
  X,
  Zap,
} from "lucide-react";
import Link from "next/link";
import { generateShareURL, isShareDataTooLong, ShareData } from "@/lib/shareUtils";
import { useCallback, useEffect, useState } from "react";
import { DiffViewer } from "./DiffViewer";
import { OnboardingTour, RestartTourButton } from "./OnboardingTour";
import { TemplateCards } from "./TemplateCards";
import { TemplateEditor } from "./TemplateEditor";
import { ValidationConfigurator } from "./ValidationConfigurator";
import { WarningPanel } from "./WarningPanel";
import { ResizablePanels } from "./ui/ResizablePanels";
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
    clearColumnOverrides,
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
  const [showValidationConfigurator, setShowValidationConfigurator] = useState(false);
  const [showTemplateEditor, setShowTemplateEditor] = useState(false);

  useEffect(() => {
    getMDFlowTemplates().then((res) => {
      if (res.data?.templates) {
        // Ensure "default" is always first
        const sorted = [...res.data.templates].sort((a, b) => {
          if (a === "default") return -1;
          if (b === "default") return 1;
          return 0;
        });
        setTemplates(sorted);
      }
    });
  }, []);

  // Reset store when leaving Studio so data is not shown when user comes back
  useEffect(() => {
    return () => reset();
  }, [reset]);

  // Auto-preview with debounce when paste text changes
  useEffect(() => {
    if (mode !== "paste" || !pasteText.trim()) {
      setPreview(null);
      setShowPreview(false);
      return;
    }

    const timer = setTimeout(async () => {
      setPreviewLoading(true);
      const result = await previewPaste(pasteText);
      setPreviewLoading(false);
      if (result.data) {
        setPreview(result.data);
        setShowPreview(true);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [pasteText, mode, setPreview, setPreviewLoading, setShowPreview]);

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (!selectedFile) return;

      setFile(selectedFile);
      setLoading(true);
      setError(null);
      setPreview(null);

      if (/\.tsv$/i.test(selectedFile.name)) {
        // Fetch TSV preview
        const previewResult = await previewTSV(selectedFile);
        setLoading(false);
        if (previewResult.data) {
          setPreview(previewResult.data);
          setShowPreview(true);
        }
        return;
      }

      const result = await getXLSXSheets(selectedFile);

      if (result.error) {
        setLoading(false);
        setError(result.error);
      } else if (result.data) {
        setSheets(result.data.sheets);
        setSelectedSheet(result.data.active_sheet);

        // Fetch XLSX preview for the active sheet
        const previewResult = await previewXLSX(
          selectedFile,
          result.data.active_sheet
        );
        setLoading(false);
        if (previewResult.data) {
          setPreview(previewResult.data);
          setShowPreview(true);
        }
      } else {
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
      setShowPreview,
    ]
  );

  // Update preview when sheet selection changes for XLSX
  useEffect(() => {
    if (mode !== "xlsx" || !file || !selectedSheet) return;

    const fetchPreview = async () => {
      setPreviewLoading(true);
      const result = await previewXLSX(file, selectedSheet);
      setPreviewLoading(false);
      if (result.data) {
        setPreview(result.data);
        setShowPreview(true);
      }
    };

    fetchPreview();
  }, [
    selectedSheet,
    file,
    mode,
    setPreview,
    setPreviewLoading,
    setShowPreview,
  ]);

  const handleConvert = useCallback(async () => {
    setLoading(true);
    setError(null);

    let result;
    let inputPreview = "";
    if (mode === "paste") {
      if (!pasteText.trim()) {
        setError("Missing source data");
        setLoading(false);
        return;
      }

      // Check if it's a Google Sheets URL
      if (isGoogleSheetsURL(pasteText.trim())) {
        result = await convertGoogleSheet(pasteText.trim(), template);
        inputPreview = `Google Sheet: ${pasteText.trim().slice(0, 60)}...`;
      } else {
        result = await convertPaste(pasteText, template);
        inputPreview =
          pasteText.slice(0, 200) + (pasteText.length > 200 ? "..." : "");
      }
    } else if (mode === "xlsx") {
      if (!file) {
        setError("No file uploaded");
        setLoading(false);
        return;
      }
      result = await convertXLSX(file, selectedSheet, template);
      inputPreview = `${file.name}${
        selectedSheet ? ` (${selectedSheet})` : ""
      }`;
    } else {
      if (!file) {
        setError("No file uploaded");
        setLoading(false);
        return;
      }
      result = await convertTSV(file, template);
      inputPreview = file.name;
    }

    setLoading(false);

    if (result.error) {
      setError(result.error);
    } else if (result.data) {
      setResult(result.data.mdflow, result.data.warnings, result.data.meta);
      // Add to history
      addToHistory({
        mode,
        template,
        inputPreview,
        output: result.data.mdflow,
        meta: result.data.meta,
      });
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
        previewTSV(f).then((result) => {
          setLoading(false);
          if (result.data) {
            setPreview(result.data);
            setShowPreview(true);
          }
        });
        return;
      }

      if (mode === "xlsx" && /\.(xlsx|xls)$/i.test(f.name)) {
        setFile(f);
        setLoading(true);
        setError(null);
        setPreview(null);
        getXLSXSheets(f).then(async (result) => {
          if (result.error) {
            setLoading(false);
            setError(result.error);
          } else if (result.data) {
            setSheets(result.data.sheets);
            setSelectedSheet(result.data.active_sheet);
            // Fetch preview
            const previewResult = await previewXLSX(
              f,
              result.data.active_sheet
            );
            setLoading(false);
            if (previewResult.data) {
              setPreview(previewResult.data);
              setShowPreview(true);
            }
          }
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
      setShowPreview,
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
                <div className="flex bg-black/50 p-1.5 rounded-lg sm:rounded-xl border border-white/5 shadow-inner shrink-0" data-tour="input-mode">
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
                  <div className="flex items-center justify-between gap-2 mb-2">
                    <label className="label">Template</label>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => setShowTemplateEditor(true)}
                        className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-[9px] font-bold uppercase tracking-wider bg-white/5 hover:bg-accent-orange/20 border border-white/10 hover:border-accent-orange/30 text-white/70 hover:text-accent-orange transition-all"
                        title="Create custom templates"
                      >
                        <FileCode className="w-3 h-3" />
                        Template Editor
                      </button>
                      <button
                        type="button"
                        onClick={() => setShowValidationConfigurator(true)}
                        className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-[9px] font-bold uppercase tracking-wider bg-white/5 hover:bg-accent-orange/20 border border-white/10 hover:border-accent-orange/30 text-white/70 hover:text-accent-orange transition-all"
                        title="Configure validation rules"
                      >
                        <ShieldCheck className="w-3 h-3" />
                        Validation rules
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
                            disabled={isDisabled}
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
                                  className="w-4 h-4 animate-spin text-accent-orange"
                                  aria-hidden
                                />
                                <span className="text-accent-orange">
                                  Processing...
                                </span>
                              </>
                            ) : (
                              <span className="flex items-center gap-2.5">
                                <Zap
                                  className={`w-4 h-4 transition-transform ${
                                    isDisabled
                                      ? ""
                                      : "group-hover:scale-110 group-hover:rotate-12"
                                  }`}
                                />
                                Run
                                {!isDisabled && (
                                  <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-0.5" />
                                )}
                              </span>
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
                          const diff = await diffMDFlow(
                            previousOutput,
                            mdflowOutput
                          );
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
                    <ShareButton
                      mdflowOutput={mdflowOutput}
                      template={template}
                    />
                    <ExportDropdown
                      mdflowOutput={mdflowOutput}
                      onDownloadMD={handleDownload}
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
                  onClose={() => setShowDiff(false)}
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
            onSelect={(record) => {
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

/* Footer analytics panel */
import { MDFlowWarning } from "@/lib/mdflowApi";

function TechnicalAnalysis({
  meta,
  warnings,
  mdflowOutput,
}: {
  meta: {
    nodeCount?: number;
    total_rows?: number;
    headerCount?: number;
    header_row?: number;
    source_type?: string;
    parser?: string;
    source_url?: string;
  } | null;
  warnings: MDFlowWarning[];
  mdflowOutput: string | null;
}) {
  return (
    <AnimatePresence mode="wait">
      {!mdflowOutput ? (
        <motion.div
          key="idle"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="flex items-center justify-between gap-4"
        >
          <div className="flex items-center gap-3">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full rounded-full bg-accent-orange/40 animate-ping opacity-75" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-white/30" />
            </span>
            <span className="text-[10px] font-black uppercase tracking-[0.25em] text-white/40">
              Standby — run engine to see stats
            </span>
          </div>
          <div className="flex gap-1.5">
            <span className="w-10 h-1 rounded-full bg-white/10" />
            <span className="w-6 h-1 rounded-full bg-white/10" />
          </div>
        </motion.div>
      ) : (
        <motion.div
          key="active"
          initial={{ opacity: 0, y: 4 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.3 }}
          className="space-y-4"
        >
          {/* Stats row */}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Activity className="w-3.5 h-3.5 text-accent-orange/80" />
              <span className="text-[9px] font-black uppercase tracking-[0.25em] text-white/50">
                Stats
              </span>
            </div>
            <div className="flex gap-3">
              <div className="px-3 py-2 rounded-lg bg-white/5 border border-white/10 flex items-center gap-2">
                <span className="text-[8px] text-white/40 font-bold uppercase tracking-wider">
                  Rows
                </span>
                <span className="text-sm font-black font-mono text-white leading-none">
                  {meta?.nodeCount ?? meta?.total_rows ?? 0}
                </span>
              </div>
              <div className="px-3 py-2 rounded-lg bg-white/5 border border-white/10 flex items-center gap-2">
                <span className="text-[8px] text-white/40 font-bold uppercase tracking-wider">
                  Header
                </span>
                <span className="text-sm font-black font-mono text-white leading-none">
                  {meta?.headerCount ?? meta?.header_row ?? 0}
                </span>
              </div>
            </div>
          </div>

          {/* Enhanced Warning Panel */}
          {warnings && warnings.length > 0 && (
            <WarningPanel warnings={warnings} />
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
}

/* Preview Table Component */
import { PreviewResponse } from "@/lib/mdflowApi";

const CANONICAL_FIELDS = [
  "id",
  "feature",
  "scenario",
  "instructions",
  "inputs",
  "expected",
  "precondition",
  "priority",
  "type",
  "status",
  "endpoint",
  "notes",
  "no",
  "item_name",
  "item_type",
  "required_optional",
  "input_restrictions",
  "display_conditions",
  "action",
  "navigation_destination",
];

function PreviewTable({
  preview,
  columnOverrides,
  onColumnOverride,
}: {
  preview: PreviewResponse;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const maxCollapsedRows = 4;
  const hasMoreRows = preview.rows.length > maxCollapsedRows;
  const displayRows = expanded
    ? preview.rows
    : preview.rows.slice(0, maxCollapsedRows);

  return (
    <div className="rounded-xl border border-white/10 bg-black/30 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2.5 bg-white/5 border-b border-white/10">
        <div className="flex items-center gap-2">
          <Table className="w-3.5 h-3.5 text-accent-orange/80" />
          <span className="text-[10px] font-black uppercase tracking-widest text-white/70">
            Preview
          </span>
          <span className="text-[9px] text-white/40 font-mono">
            {preview.total_rows} rows • {preview.headers.length} cols
          </span>
          {preview.confidence < 70 && (
            <span className="text-[9px] text-accent-gold/80 font-medium">
              (low confidence: {preview.confidence}%)
            </span>
          )}
        </div>
        {hasMoreRows && (
          <button
            type="button"
            onClick={() => setExpanded(!expanded)}
            className="text-[9px] text-accent-orange/70 hover:text-accent-orange cursor-pointer font-bold uppercase"
          >
            {expanded ? "Show less" : `Show all ${preview.rows.length}`}
          </button>
        )}
      </div>

      <div className="overflow-x-auto custom-scrollbar">
        <table className="w-full text-[11px]">
          <thead>
            <tr className="border-b border-white/10 bg-white/3">
              {preview.headers.map((header, i) => {
                const mappedField =
                  columnOverrides[header] ||
                  preview.column_mapping[header] ||
                  "";
                const isUnmapped =
                  !mappedField && preview.unmapped_columns.includes(header);

                return (
                  <th key={i} className="px-3 py-2 text-left">
                    <div className="space-y-1">
                      <span
                        className="font-bold text-white/90 block truncate max-w-[150px]"
                        title={header}
                      >
                        {header}
                      </span>
                      <select
                        value={mappedField}
                        onChange={(e) =>
                          onColumnOverride(header, e.target.value)
                        }
                        className={`
                          text-[9px] px-1.5 py-0.5 rounded bg-black/40 border cursor-pointer
                          ${
                            isUnmapped
                              ? "border-accent-gold/40 text-accent-gold/80"
                              : "border-white/10 text-accent-orange/80"
                          }
                        `}
                      >
                        <option value="">— unmapped —</option>
                        {CANONICAL_FIELDS.map((field) => (
                          <option key={field} value={field}>
                            {field.replace(/_/g, " ")}
                          </option>
                        ))}
                      </select>
                    </div>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {displayRows.map((row, rowIdx) => (
              <tr
                key={rowIdx}
                className="border-b border-white/5 hover:bg-white/3"
              >
                {row.map((cell, cellIdx) => (
                  <td
                    key={cellIdx}
                    className="px-3 py-2 text-white/70 font-mono"
                  >
                    <span className="block truncate max-w-[200px]" title={cell}>
                      {cell || <span className="text-white/30">—</span>}
                    </span>
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {preview.total_rows > preview.preview_rows && (
        <div className="px-4 py-2 text-[9px] text-white/40 bg-white/3 border-t border-white/5">
          Showing {displayRows.length} of {preview.total_rows} rows
        </div>
      )}
    </div>
  );
}

/* Share Button Component */
function ShareButton({
  mdflowOutput,
  template,
}: {
  mdflowOutput: string;
  template: string;
}) {
  const [showTooltip, setShowTooltip] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const shareData: ShareData = {
    mdflow: mdflowOutput,
    template,
    createdAt: Date.now(),
  };

  const isTooLong = isShareDataTooLong(shareData);

  const handleShare = useCallback(() => {
    if (isTooLong) {
      setError("Content too large for URL sharing. Try exporting instead.");
      setShowTooltip(true);
      return;
    }

    try {
      const url = generateShareURL(shareData);
      navigator.clipboard.writeText(url);
      setCopied(true);
      setError(null);
      setShowTooltip(true);
      setTimeout(() => {
        setCopied(false);
        setShowTooltip(false);
      }, 3000);
    } catch (err) {
      setError("Failed to generate share link");
      setShowTooltip(true);
    }
  }, [shareData, isTooLong]);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={handleShare}
        onMouseEnter={() => !copied && setShowTooltip(true)}
        onMouseLeave={() => !copied && setShowTooltip(false)}
        disabled={isTooLong}
        className={`
          flex items-center justify-center h-8 px-3 rounded-lg text-[9px] font-bold uppercase tracking-wider transition-all cursor-pointer shrink-0
          ${isTooLong
            ? "bg-white/5 text-white/30 cursor-not-allowed border border-white/10"
            : "bg-white/10 hover:bg-white/20 text-white/80 hover:text-white border border-white/10 hover:border-white/20"
          }
        `}
        title={isTooLong ? "Content too large to share via URL" : "Share link"}
      >
        <Share2 className="w-3 h-3 shrink-0" />
      </button>

      <AnimatePresence>
        {showTooltip && (
          <motion.div
            initial={{ opacity: 0, y: 4, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 4, scale: 0.95 }}
            className="absolute right-0 top-full mt-2 z-50 w-56 rounded-xl bg-black/95 backdrop-blur-xl border border-white/20 shadow-2xl overflow-hidden"
          >
            <div className="p-3">
              {error ? (
                <div className="flex items-start gap-2">
                  <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                  <p className="text-[11px] text-red-400">{error}</p>
                </div>
              ) : copied ? (
                <div className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-green-400" />
                  <div>
                    <p className="text-[11px] font-bold text-white">Link copied!</p>
                    <p className="text-[9px] text-white/50 mt-0.5">
                      Share this URL with anyone
                    </p>
                  </div>
                </div>
              ) : (
                <div className="flex items-start gap-2">
                  <Share2 className="w-4 h-4 text-accent-orange shrink-0 mt-0.5" />
                  <div>
                    <p className="text-[11px] font-bold text-white">Share Link</p>
                    <p className="text-[9px] text-white/50 mt-0.5">
                      Click to copy shareable URL
                    </p>
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

/* Export Dropdown Component */
function ExportDropdown({
  mdflowOutput,
  onDownloadMD,
}: {
  mdflowOutput: string;
  onDownloadMD: () => void;
}) {
  const [open, setOpen] = useState(false);

  const downloadAs = (format: "md" | "json" | "yaml") => {
    let content = mdflowOutput;
    let filename = "spec";
    let mimeType = "text/plain";

    if (format === "json") {
      // Extract YAML frontmatter and convert to JSON
      const yamlMatch = mdflowOutput.match(/^---\n([\s\S]*?)\n---/);
      const frontmatter = yamlMatch ? yamlMatch[1] : "";
      const body = mdflowOutput.replace(/^---\n[\s\S]*?\n---\n?/, "");

      content = JSON.stringify(
        {
          frontmatter: frontmatter.split("\n").reduce((acc, line) => {
            const [key, ...vals] = line.split(":");
            if (key && vals.length) acc[key.trim()] = vals.join(":").trim();
            return acc;
          }, {} as Record<string, string>),
          content: body,
        },
        null,
        2
      );
      filename = "spec.json";
      mimeType = "application/json";
    } else if (format === "yaml") {
      filename = "spec.yaml";
      mimeType = "text/yaml";
    } else {
      filename = "spec.mdflow.md";
      mimeType = "text/markdown";
    }

    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    setOpen(false);
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex items-center justify-center h-8 px-3 rounded-lg bg-accent-orange text-[9px] font-bold uppercase tracking-wider text-white shadow-md shadow-accent-orange/25 hover:bg-accent-orange/90 active:scale-95 transition-all cursor-pointer shrink-0"
      >
        <span className="inline-flex items-center gap-1.5 leading-none">
          <Download className="block w-3 h-3 shrink-0" />
          <span className="leading-none">Export</span>
        </span>
      </button>

      <AnimatePresence>
        {open && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setOpen(false)}
            />
            <motion.div
              initial={{ opacity: 0, y: -8, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -8, scale: 0.95 }}
              className="absolute right-0 top-full mt-2 z-50 w-48 rounded-xl bg-black/90 backdrop-blur-xl border border-white/20 shadow-2xl overflow-hidden"
            >
              <div className="p-1">
                <button
                  onClick={() => downloadAs("md")}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-white/10 transition-colors cursor-pointer"
                >
                  <FileText className="w-4 h-4 text-accent-orange" />
                  <div className="text-left">
                    <p className="text-[11px] font-bold text-white">Markdown</p>
                    <p className="text-[9px] text-white/50">.mdflow.md</p>
                  </div>
                </button>
                <button
                  onClick={() => downloadAs("json")}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-white/10 transition-colors cursor-pointer"
                >
                  <FileJson className="w-4 h-4 text-blue-400" />
                  <div className="text-left">
                    <p className="text-[11px] font-bold text-white">JSON</p>
                    <p className="text-[9px] text-white/50">.json</p>
                  </div>
                </button>
                <button
                  onClick={() => downloadAs("yaml")}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-white/10 transition-colors cursor-pointer"
                >
                  <FileText className="w-4 h-4 text-green-400" />
                  <div className="text-left">
                    <p className="text-[11px] font-bold text-white">YAML</p>
                    <p className="text-[9px] text-white/50">.yaml</p>
                  </div>
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}

/* History Modal Component */
import { ConversionRecord } from "@/lib/mdflowStore";

function HistoryModal({
  history,
  onClose,
  onSelect,
}: {
  history: ConversionRecord[];
  onClose: () => void;
  onSelect: (record: ConversionRecord) => void;
}) {
  const { clearHistory } = useHistoryStore();
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = useCallback(
    (record: ConversionRecord, e: React.MouseEvent) => {
      e.stopPropagation();
      navigator.clipboard.writeText(record.output);
      setCopiedId(record.id);
      setTimeout(() => setCopiedId(null), 2000);
    },
    []
  );

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      onClick={onClose}
      className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <motion.div
        initial={{ scale: 0.95, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.95, opacity: 0 }}
        onClick={(e) => e.stopPropagation()}
        className="bg-black/80 backdrop-blur-xl border border-white/20 rounded-2xl shadow-2xl max-w-2xl w-full max-h-[70vh] flex flex-col overflow-hidden"
      >
        <div className="flex items-center justify-between gap-4 px-6 py-4 border-b border-white/10 bg-white/3 shrink-0">
          <div className="flex items-center gap-3">
            <History className="w-4 h-4 text-accent-orange" />
            <span className="text-[11px] font-black uppercase tracking-[0.2em] text-white/80">
              Conversion History
            </span>
            <span className="text-[10px] text-white/40 font-mono">
              {history.length} records
            </span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => {
                clearHistory();
              }}
              className="text-[9px] text-white/40 hover:text-accent-red/80 transition-colors cursor-pointer font-bold uppercase"
            >
              Clear all
            </button>
            <button
              onClick={onClose}
              className="p-2 rounded-md hover:bg-white/10 transition-colors cursor-pointer text-white/60 hover:text-white"
              aria-label="Close"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-2">
          {history.length === 0 ? (
            <div className="text-center py-12 text-white/40">
              <Clock className="w-8 h-8 mx-auto mb-3 opacity-50" />
              <p className="text-[11px] font-bold uppercase tracking-wider">
                No history yet
              </p>
              <p className="text-[10px] mt-1">Conversions will appear here</p>
            </div>
          ) : (
            history.map((record, idx) => (
              <motion.div
                key={record.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.05 }}
                className="group relative"
              >
                <div
                  onClick={() => onSelect(record)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      onSelect(record);
                    }
                  }}
                  className="w-full text-left p-4 rounded-xl bg-linear-to-br from-white/5 to-white/2 hover:from-white/10 hover:to-white/5 border border-white/5 hover:border-white/15 transition-all duration-300 cursor-pointer relative overflow-hidden"
                >
                  {/* Subtle gradient overlay on hover */}
                  <div className="absolute inset-0 bg-linear-to-r from-accent-orange/0 via-accent-orange/0 to-accent-orange/0 group-hover:from-accent-orange/5 group-hover:via-accent-orange/0 group-hover:to-transparent transition-all duration-500 pointer-events-none" />

                  <div className="relative z-10">
                    <div className="flex items-center justify-between gap-4 mb-3">
                      <div className="flex items-center gap-2.5 flex-wrap">
                        <motion.span
                          whileHover={{ scale: 1.05 }}
                          className="text-[9px] font-black uppercase tracking-[0.15em] px-2.5 py-1.5 h-6 rounded-lg bg-linear-to-r from-accent-orange/20 to-accent-orange/10 border border-accent-orange/30 text-accent-orange shadow-[0_0_8px_rgba(242,123,47,0.15)] flex items-center leading-none whitespace-nowrap"
                        >
                          {record.mode}
                        </motion.span>

                        {/* Premium Template Badge */}
                        <div className="relative">
                          <div className="absolute inset-0 bg-linear-to-r from-white/10 via-white/5 to-transparent rounded-lg blur-sm opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                          <span className="relative text-[9px] font-semibold uppercase tracking-[0.2em] px-3 py-1.5 h-6 rounded-lg bg-white/5 border border-white/10 text-white/70 group-hover:text-white/90 group-hover:border-white/20 transition-all duration-300 backdrop-blur-sm flex items-center leading-none whitespace-nowrap">
                            {record.template.replace(/-/g, " ")}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <span className="text-[9px] text-white/30 font-mono">
                          {new Date(record.timestamp).toLocaleString()}
                        </span>
                        <motion.button
                          whileHover={{ scale: 1.1 }}
                          whileTap={{ scale: 0.95 }}
                          onClick={(e) => handleCopy(record, e)}
                          className="p-1.5 rounded-lg bg-white/5 hover:bg-accent-orange/20 border border-white/10 hover:border-accent-orange/30 text-white/50 hover:text-accent-orange transition-all duration-200 cursor-pointer"
                          title="Copy output"
                        >
                          {copiedId === record.id ? (
                            <Check className="w-3.5 h-3.5 text-accent-orange" />
                          ) : (
                            <Copy className="w-3.5 h-3.5" />
                          )}
                        </motion.button>
                      </div>
                    </div>

                    <p className="text-[10px] text-white/60 font-mono truncate mb-1.5">
                      {record.inputPreview}
                    </p>
                    <div className="flex items-center gap-2">
                      <div className="h-px flex-1 bg-linear-to-r from-white/10 to-transparent" />
                      <p className="text-[9px] text-white/30 font-medium">
                        {record.meta?.total_rows ?? 0} rows
                      </p>
                    </div>
                  </div>
                </div>
              </motion.div>
            ))
          )}
        </div>
      </motion.div>
    </motion.div>
  );
}

/* Keyboard Shortcuts Tooltip */
import { useOnboardingStore } from "@/lib/onboardingStore";

function KeyboardShortcutsTooltip() {
  const [show, setShow] = useState(false);
  const { resetTour, startTour } = useOnboardingStore();
  const isMac =
    typeof navigator !== "undefined" &&
    navigator.platform.toUpperCase().indexOf("MAC") >= 0;
  const mod = isMac ? "⌘" : "Ctrl";

  const shortcuts = [
    { keys: `${mod}+Enter`, action: "Convert" },
    { keys: `${mod}+Shift+C`, action: "Copy output" },
    { keys: `${mod}+S`, action: "Download" },
  ];

  const handleRestartTour = () => {
    setShow(false);
    resetTour();
    setTimeout(startTour, 100);
  };

  return (
    <div
      className="relative"
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
    >
      <button
        type="button"
        onClick={() => setShow(!show)}
        className="p-2.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white/40 hover:text-white/60 transition-all cursor-pointer"
        aria-label="Keyboard shortcuts"
      >
        <Keyboard className="w-4 h-4" />
      </button>

      <AnimatePresence>
        {show && (
          <motion.div
            initial={{ opacity: 0, y: 8, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 8, scale: 0.95 }}
            className="absolute bottom-full right-0 mb-2 w-52 rounded-xl bg-black/90 backdrop-blur-xl border border-white/20 shadow-2xl overflow-hidden"
          >
            <div className="px-3 py-2 border-b border-white/10 bg-white/5">
              <span className="text-[9px] font-black uppercase tracking-widest text-white/60">
                Shortcuts
              </span>
            </div>
            <div className="p-2 space-y-1">
              {shortcuts.map((s) => (
                <div
                  key={s.keys}
                  className="flex items-center justify-between px-2 py-1.5 rounded-lg hover:bg-white/5"
                >
                  <span className="text-[10px] text-white/70">{s.action}</span>
                  <kbd className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-white/10 text-white/60">
                    {s.keys}
                  </kbd>
                </div>
              ))}
            </div>
            <div className="p-2 border-t border-white/10">
              <button
                onClick={handleRestartTour}
                className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-accent-orange/10 hover:bg-accent-orange/20 border border-accent-orange/20 text-[10px] font-bold uppercase tracking-wider text-accent-orange hover:text-accent-orange transition-all cursor-pointer"
              >
                <BookOpen className="w-3.5 h-3.5" />
                Restart Tour
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
