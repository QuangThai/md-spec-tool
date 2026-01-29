"use client";

import {
  convertPaste,
  convertXLSX,
  getMDFlowTemplates,
  getXLSXSheets,
} from "@/lib/mdflowApi";
import { useMDFlowStore } from "@/lib/mdflowStore";
import { AnimatePresence, motion } from "framer-motion";
import {
  Activity,
  AlertCircle,
  ArrowRight,
  Check,
  Copy,
  Database,
  Download,
  FileSpreadsheet,
  RefreshCcw,
  Terminal,
  Zap,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
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
    setMode,
    setPasteText,
    setFile,
    setSheets,
    setSelectedSheet,
    setTemplate,
    setResult,
    setLoading,
    setError,
    reset,
  } = useMDFlowStore();

  const [templates, setTemplates] = useState<string[]>(["default"]);
  const [copied, setCopied] = useState(false);
  const [dragOver, setDragOver] = useState(false);

  useEffect(() => {
    getMDFlowTemplates().then((res) => {
      if (res.data?.templates) {
        setTemplates(res.data.templates);
      }
    });
  }, []);

  // Reset store when leaving Studio so data is not shown when user comes back
  useEffect(() => {
    return () => reset();
  }, [reset]);

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0];
      if (!selectedFile) return;

      setFile(selectedFile);
      setLoading(true);
      setError(null);

      const result = await getXLSXSheets(selectedFile);
      setLoading(false);

      if (result.error) {
        setError(result.error);
      } else if (result.data) {
        setSheets(result.data.sheets);
        setSelectedSheet(result.data.active_sheet);
      }
    },
    [setFile, setLoading, setError, setSheets, setSelectedSheet],
  );

  const handleConvert = useCallback(async () => {
    setLoading(true);
    setError(null);

    let result;
    if (mode === "paste") {
      if (!pasteText.trim()) {
        setError("Missing source data");
        setLoading(false);
        return;
      }
      result = await convertPaste(pasteText, template);
    } else {
      if (!file) {
        setError("No file uploaded");
        setLoading(false);
        return;
      }
      result = await convertXLSX(file, selectedSheet, template);
    }

    setLoading(false);

    if (result.error) {
      setError(result.error);
    } else if (result.data) {
      setResult(result.data.mdflow, result.data.warnings, result.data.meta);
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

  const currentStep = mdflowOutput ? 3 : file || pasteText.trim() ? 2 : 1;

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const f = e.dataTransfer.files?.[0];
      if (f && /\.(xlsx|xls)$/i.test(f.name)) {
        setFile(f);
        setLoading(true);
        setError(null);
        getXLSXSheets(f).then((result) => {
          setLoading(false);
          if (result.error) setError(result.error);
          else if (result.data) {
            setSheets(result.data.sheets);
            setSelectedSheet(result.data.active_sheet);
          }
        });
      }
    },
    [setFile, setLoading, setError, setSheets, setSelectedSheet],
  );

  return (
    <motion.div
      variants={stagger.container}
      initial="initial"
      animate="animate"
      className="flex flex-col gap-6 sm:gap-8 lg:gap-12 relative"
    >
      {/* Step flow strip */}
      <motion.div
        variants={stagger.item}
        className="flex justify-center"
        aria-label="Workflow steps"
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
                    ${active
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
                <div className="flex bg-black/50 p-1.5 rounded-lg sm:rounded-xl border border-white/5 shadow-inner shrink-0">
                  <button
                    type="button"
                    onClick={() => setMode("paste")}
                    className={`
                      px-4 sm:px-5 py-2 sm:py-2.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md sm:rounded-lg transition-all duration-200
                      ${mode === "paste"
                        ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                        : "text-muted hover:text-white hover:bg-white/5"
                      }
                    `}
                  >
                    Paste
                  </button>
                  <button
                    type="button"
                    onClick={() => setMode("xlsx")}
                    className={`
                      px-4 sm:px-5 py-2 sm:py-2.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md sm:rounded-lg transition-all duration-200
                      ${mode === "xlsx"
                        ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                        : "text-muted hover:text-white hover:bg-white/5"
                      }
                    `}
                  >
                    Excel
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
                        <span>Paste TSV / CSV</span>
                        <span className="text-accent-orange/50 font-medium sm:text-right">
                          Tab-separated or comma
                        </span>
                      </div>
                      <textarea
                        value={pasteText}
                        onChange={(e) => setPasteText(e.target.value)}
                        placeholder="Paste your table data here…"
                        className="input flex-1 font-mono text-[13px] leading-relaxed resize-none border-white/5 bg-black/20 focus:bg-black/30 focus:border-accent-orange/30 custom-scrollbar min-h-[200px] rounded-xl"
                        aria-label="Paste TSV or CSV data"
                      />
                    </motion.div>
                  ) : (
                    <motion.div
                      key="xlsx"
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
                          ${dragOver
                            ? "border-accent-orange/50 bg-accent-orange/10 scale-[1.02]"
                            : "border-white/10 hover:border-white/20 hover:bg-white/3"
                          }
                        `}
                      >
                        <input
                          type="file"
                          accept=".xlsx,.xls"
                          onChange={handleFileChange}
                          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                          aria-label="Upload Excel file"
                        />
                        <motion.div
                          animate={{ scale: dragOver ? 1.05 : 1 }}
                          className="flex flex-col items-center gap-4"
                        >
                          <div
                            className={`
                              h-16 w-16 rounded-2xl flex items-center justify-center transition-all duration-300
                              ${dragOver ? "bg-accent-orange/20" : "bg-white/5 group-hover:bg-white/10"}
                            `}
                          >
                            <FileSpreadsheet
                              className={`w-8 h-8 transition-colors ${dragOver ? "text-accent-orange" : "text-muted/50"}`}
                            />
                          </div>
                          <div>
                            <p className="text-sm font-black text-white uppercase tracking-widest">
                              {dragOver
                                ? "Drop file here"
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

                      {sheets.length > 0 && (
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
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              <div className="px-5 sm:px-6 lg:px-8 py-5 sm:py-6 lg:py-8 border-t border-white/5 bg-white/2 flex flex-col sm:flex-row gap-4 sm:gap-5 items-end shrink-0">
                <div className="flex-1 w-full min-w-0 space-y-2">
                  <label className="label mb-2">Template</label>
                  <Select
                    value={template}
                    onValueChange={setTemplate}
                    options={templates.map((t) => ({
                      label:
                        t.charAt(0).toUpperCase() + t.slice(1).toLowerCase(),
                      value: t,
                    }))}
                    placeholder="Template"
                    side="top"
                    className="h-12"
                  />
                </div>
                <motion.button
                  type="button"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={handleConvert}
                  disabled={loading}
                  className="btn-primary h-12 w-full sm:w-auto min-w-[140px] sm:min-w-[160px] px-6 sm:px-8 uppercase tracking-[0.18em] shadow-xl shadow-accent-orange/25 group cursor-pointer rounded-xl shrink-0"
                >
                  {loading ? (
                    <RefreshCcw className="w-4 h-4 animate-spin" aria-hidden />
                  ) : (
                    <span className="flex items-center gap-2.5">
                      Run{" "}
                      <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-0.5" />
                    </span>
                  )}
                </motion.button>
              </div>
            </div>
          </section>
        </motion.div>

        {/* Right: Output — scrollable pre, footer (Stats) stays at bottom */}
        <motion.div
          variants={stagger.item}
          className="flex flex-col min-h-0 h-full overflow-hidden"
        >
          <div className="code-editor p-0 border-white/10 h-full min-h-0 flex flex-col shadow-2xl bg-black/40 backdrop-blur-2xl relative overflow-hidden rounded-2xl">
            <div className="studio-grain" aria-hidden />
            <div className="relative z-10 flex flex-col h-full min-h-0">
              <div className="flex items-center justify-between gap-4 p-4 sm:p-5 lg:p-6 border-b border-white/10 bg-white/6 shrink-0">
                <div className="flex items-center gap-2 sm:gap-3 min-w-0">
                  <div className="flex gap-1 sm:gap-1.5 shrink-0">
                    <span className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-accent-red/80 shadow-[0_0_6px_rgba(239,68,68,0.5)]" />
                    <span className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-accent-gold/80 shadow-[0_0_6px_rgba(251,191,36,0.5)]" />
                    <span className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-accent-orange/80 shadow-[0_0_6px_rgba(242,123,47,0.5)]" />
                  </div>
                  <span className="text-[9px] sm:text-[10px] font-black uppercase tracking-[0.3em] sm:tracking-[0.35em] text-white/80 font-mono">
                    Output
                  </span>
                </div>
                {mdflowOutput && (
                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      type="button"
                      onClick={handleCopy}
                      className="p-2.5 rounded-xl bg-white/10 hover:bg-white/20 text-white/80 hover:text-white transition-all border border-white/10 hover:border-white/20 cursor-pointer"
                      title={copied ? "Copied" : "Copy"}
                      aria-label={copied ? "Copied" : "Copy to clipboard"}
                    >
                      {copied ? (
                        <Check className="w-3.5 h-3.5 text-accent-orange" />
                      ) : (
                        <Copy className="w-3.5 h-3.5" />
                      )}
                    </button>
                    <button
                      type="button"
                      onClick={handleDownload}
                      className="flex items-center gap-2 h-9 px-4 rounded-xl bg-accent-orange text-[10px] font-bold uppercase tracking-widest text-white shadow-lg shadow-accent-orange/30 hover:bg-accent-orange/90 active:scale-95 transition-all cursor-pointer"
                    >
                      <Download className="w-3.5 h-3.5" /> Export
                    </button>
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

      <motion.div variants={stagger.item} className="flex justify-center">
        <div className="inline-flex items-center gap-2 text-[8px] sm:text-[9px] font-bold text-muted/50 uppercase tracking-[0.3em] sm:tracking-[0.35em] bg-white/3 border border-white/5 px-3 sm:px-5 py-1.5 sm:py-2 rounded-full">
          <span className="w-1.5 h-1.5 rounded-full bg-accent-orange/80 animate-pulse" />
          MDFlow Studio
        </div>
      </motion.div>
    </motion.div>
  );
}

/* Footer analytics panel */
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
  } | null;
  warnings: string[];
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
          className="grid grid-cols-1 md:grid-cols-2 gap-6 items-start"
        >
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Activity className="w-3.5 h-3.5 text-accent-orange/80" />
              <span className="text-[9px] font-black uppercase tracking-[0.25em] text-white/50">
                Stats
              </span>
            </div>
            <div className="flex gap-3">
              <div className="flex-1 p-3 rounded-xl bg-white/5 border border-white/10 text-center">
                <p className="text-[8px] text-white/40 mb-0.5 font-bold uppercase tracking-wider">
                  Nodes
                </p>
                <p className="text-sm font-black font-mono text-white leading-none">
                  {meta?.nodeCount ?? meta?.total_rows ?? 0}
                </p>
              </div>
              <div className="flex-1 p-3 rounded-xl bg-white/5 border border-white/10 text-center">
                <p className="text-[8px] text-white/40 mb-0.5 font-bold uppercase tracking-wider">
                  Header
                </p>
                <p className="text-sm font-black font-mono text-white leading-none">
                  {meta?.headerCount ?? meta?.header_row ?? 0}
                </p>
              </div>
            </div>
          </div>
          {warnings && warnings.length > 0 && (
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <AlertCircle className="w-3.5 h-3.5 text-accent-gold/90" />
                <span className="text-[9px] font-black uppercase tracking-[0.25em] text-accent-gold/80">
                  Warnings
                </span>
              </div>
              <div className="max-h-[52px] overflow-auto custom-scrollbar space-y-1.5">
                {warnings.map((w: string, i: number) => (
                  <p
                    key={`${i}-${w.slice(0, 20)}`}
                    className="text-[9px] font-medium text-accent-gold/90 leading-snug pl-2.5 border-l-2 border-accent-gold/40"
                  >
                    {w}
                  </p>
                ))}
              </div>
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
}
