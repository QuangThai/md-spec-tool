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
  Cpu,
  Database,
  Download,
  FileSpreadsheet,
  RefreshCcw,
  Terminal,
  Zap,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Select } from "./ui/Select";

const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.8, ease: [0.16, 1, 0.3, 1] },
};

const steps = [
  {
    title: "Import Data",
    desc: "Paste TSV/CSV or upload Excel.",
    icon: <Database />,
  },
  {
    title: "Configure",
    desc: "Select sheet and template logic.",
    icon: <Terminal />,
  },
  { title: "Generate", desc: "Instant MDFlow specification.", icon: <Zap /> },
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

  useEffect(() => {
    getMDFlowTemplates().then((res) => {
      if (res.data?.templates) {
        setTemplates(res.data.templates);
      }
    });
  }, []);

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

  return (
    <motion.div
      initial="initial"
      animate="animate"
      className="flex flex-col gap-10"
    >
      {/* 🚀 Integrated Studio Workspace */}
      <div className="grid grid-cols-1 lg:grid-cols-[1.2fr_1fr] gap-8 items-stretch h-auto lg:h-[800px] lg:min-h-[800px] lg:max-h-[800px]">
        {/* 🛠 Left Column: Configuration & Ingestion */}
        <div className="flex flex-col h-full overflow-hidden">
          <motion.section
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5 }}
            className="surface p-0! flex flex-col h-full border-white/10 bg-white/2 relative"
          >
            {/* Header: Source Toggle */}
            <div className="flex items-center justify-between p-6 border-b border-white/5 bg-white/2 shrink-0">
              <div className="flex items-center gap-3">
                <div className="h-8 w-8 rounded-lg bg-accent-orange/10 flex items-center justify-center border border-accent-orange/20">
                  <Database className="w-4 h-4 text-accent-orange" />
                </div>
                <h2 className="text-sm font-black uppercase tracking-widest text-white">
                  Source Channel
                </h2>
              </div>

              <div className="flex bg-black/40 p-1 rounded-lg border border-white/5">
                <button
                  onClick={() => setMode("paste")}
                  className={`px-4 py-1.5 text-[9px] font-bold uppercase cursor-pointer tracking-wider rounded-md transition-all ${mode === "paste"
                      ? "bg-accent-orange text-white shadow-md shadow-accent-orange/20"
                      : "text-muted hover:text-white"
                    }`}
                >
                  Raw Pasted
                </button>
                <button
                  onClick={() => setMode("xlsx")}
                  className={`px-4 py-1.5 text-[9px] font-bold uppercase cursor-pointer tracking-wider rounded-md transition-all ${mode === "xlsx"
                      ? "bg-accent-orange text-white shadow-md shadow-accent-orange/20"
                      : "text-muted hover:text-white"
                    }`}
                >
                  Excel Stream
                </button>
              </div>
            </div>

            {/* Body: Input Area (Scrollable) */}
            <div className="flex-1 overflow-auto p-8 custom-scrollbar bg-black/5 min-h-0">
              <AnimatePresence mode="wait" initial={false}>
                {error && (
                  <motion.div
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    className="mb-8 p-4 bg-accent-red/10 border border-accent-red/20 rounded-xl flex items-center gap-3 text-accent-red text-[10px] font-black uppercase tracking-widest shrink-0"
                  >
                    <AlertCircle className="w-4 h-4" /> {error}
                  </motion.div>
                )}
              </AnimatePresence>

              {mode === "paste" ? (
                <motion.div
                  key="p"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="h-full flex flex-col min-h-0"
                >
                  <div className="flex justify-between items-center text-[10px] uppercase font-bold text-muted/60 mb-4 shrink-0">
                    <span>Dataset Entry</span>
                    <span className="text-accent-orange/40">
                      TSV / CSV Supported
                    </span>
                  </div>
                  <textarea
                    value={pasteText}
                    onChange={(e) => setPasteText(e.target.value)}
                    placeholder="Input structured technical data..."
                    className="input flex-1 font-mono text-[13px] leading-relaxed resize-none border-white/5 bg-black/20 focus:bg-black/40 custom-scrollbar min-h-0"
                  />
                </motion.div>
              ) : (
                <motion.div
                  key="x"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="h-full flex flex-col justify-center space-y-8 min-h-0"
                >
                  <div className="border-2 border-dashed border-white/10 rounded-2xl p-12 text-center hover:bg-white/5 transition-all group cursor-pointer relative shrink-0">
                    <input
                      type="file"
                      accept=".xlsx,.xls"
                      onChange={handleFileChange}
                      className="absolute inset-0 opacity-0 cursor-pointer"
                    />
                    <div className="flex flex-col items-center gap-4">
                      <div className="h-16 w-16 rounded-full bg-white/5 border border-white/5 flex items-center justify-center group-hover:scale-110 transition-transform">
                        <FileSpreadsheet className="w-8 h-8 text-muted/40 group-hover:text-accent-orange" />
                      </div>
                      <div>
                        <p className="text-sm font-bold text-white uppercase tracking-widest">
                          Connect Binary Asset
                        </p>
                        <p className="text-[10px] text-muted mt-2 uppercase font-medium">
                          Click to upload spreadsheet stream
                        </p>
                      </div>
                    </div>
                  </div>

                  {file && (
                    <div className="flex items-center gap-4 p-4 rounded-xl bg-accent-orange/5 border border-accent-orange/20 shrink-0">
                      <Check className="w-4 h-4 text-accent-orange" />
                      <span className="text-[11px] font-bold text-white uppercase truncate flex-1">
                        {file.name}
                      </span>
                      <span className="text-[10px] text-muted font-mono">
                        {(file.size / 1024).toFixed(1)}KB
                      </span>
                    </div>
                  )}

                  {sheets.length > 0 && (
                    <div className="space-y-4 shrink-0">
                      <label className="label">Active Data Plane</label>
                      <Select
                        value={selectedSheet}
                        onValueChange={setSelectedSheet}
                        options={sheets.map((s) => ({
                          label: s.toUpperCase(),
                          value: s,
                        }))}
                        placeholder="Select target sheet"
                        className="h-14"
                      />
                    </div>
                  )}
                </motion.div>
              )}
            </div>

            {/* Footer: Global Config & Action - Stays at bottom */}
            <div className="p-8 border-t border-white/5 bg-white/2 flex flex-col sm:flex-row gap-6 items-end shrink-0 mt-auto relative z-20">
              <div className="flex-1 w-full space-y-4">
                <label className="label">Blueprint Schema</label>
                <Select
                  value={template}
                  onValueChange={setTemplate}
                  options={templates.map((t) => ({
                    label: `${t.charAt(0).toUpperCase() + t.slice(1).toLowerCase()} Model`,
                    value: t,
                  }))}
                  placeholder="Select technical model"
                  side="top"
                  className="h-12"
                />
              </div>
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={handleConvert}
                disabled={loading}
                className="btn-primary h-12 w-full sm:w-auto px-10 uppercase tracking-[0.2em] shadow-2xl shadow-accent-orange/30 group cursor-pointer"
              >
                {loading ? (
                  <RefreshCcw className="w-4 h-4 animate-spin" />
                ) : (
                  <span className="flex items-center gap-3">
                    Run Engine{" "}
                    <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-1" />
                  </span>
                )}
              </motion.button>
            </div>
          </motion.section>
        </div>

        {/* 💻 Right Column: Real-time Output Engine */}
        <div className="flex flex-col h-full overflow-hidden">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5 }}
            className="code-editor p-0! border-white/10 h-full flex flex-col shadow-3xl bg-white/5 backdrop-blur-3xl relative"
          >
            {/* Editor Header - Brighten for visibility */}
            <div className="flex items-center justify-between p-6 border-b border-white/10 bg-white/15 shrink-0">
              <div className="flex items-center gap-4">
                <div className="flex gap-2">
                  <span className="w-2.5 h-2.5 rounded-full bg-accent-red/70 shadow-[0_0_8px_rgba(239,68,68,0.4)]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-accent-gold/70 shadow-[0_0_8px_rgba(251,191,36,0.4)]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-accent-orange/70 shadow-[0_0_8px_rgba(242,123,47,0.4)]" />
                </div>
                <div className="h-4 w-px bg-white/30 mx-2" />
                <span className="text-[10px] font-black uppercase tracking-[0.4em] text-white/80 font-mono">
                  Terminal:127.0.0.1
                </span>
              </div>

              {mdflowOutput && (
                <div className="flex items-center gap-3">
                  <button
                    onClick={handleCopy}
                    className="p-2 rounded-lg bg-white/10 hover:bg-white/20 text-white/80 hover:text-white transition-all border border-white/20"
                  >
                    {copied ? (
                      <Check className="w-3.5 h-3.5 text-accent-orange" />
                    ) : (
                      <Copy className="w-3.5 h-3.5" />
                    )}
                  </button>
                  <button
                    onClick={handleDownload}
                    className="flex items-center gap-1.5 h-9 px-4 rounded-lg bg-accent-orange text-[9px] font-bold uppercase tracking-widest text-white shadow-lg shadow-accent-orange/30 transition-transform active:scale-95"
                  >
                    <Download className="w-3 h-3" /> Export
                  </button>
                </div>
              )}
            </div>

            {/* Editor Body */}
            <div className="flex-1 overflow-auto p-8 custom-scrollbar min-h-0">
              {mdflowOutput ? (
                <pre className="whitespace-pre-wrap leading-relaxed text-white font-mono text-[13px] selection:bg-accent-orange/40">
                  {mdflowOutput}
                </pre>
              ) : (
                <div className="h-full flex flex-col items-center justify-center text-center opacity-60">
                  <Terminal className="w-12 h-12 mb-6 text-white/20 animate-pulse" />
                  <p className="text-[10px] font-black uppercase tracking-[0.3em] text-white/60">
                    Awaiting Data Stream
                  </p>
                </div>
              )}
            </div>

            {/* Integrated Technical Analytics (Always visible for stability) */}
            <div className="border-t border-white/10 bg-white/2 p-6 shrink-0 min-h-[140px] flex flex-col justify-center">
              <TechnicalAnalysis
                meta={meta}
                warnings={warnings}
                mdflowOutput={mdflowOutput}
              />
            </div>
          </motion.div>
        </div>
      </div>

      <div className="flex justify-center mt-4">
        <div className="inline-flex items-center gap-2 text-[8px] font-bold text-muted/60 uppercase tracking-[0.4em] bg-white/5 border border-white/5 px-5 py-2 rounded-full">
          <span className="w-2 h-2 rounded-full bg-accent-orange animate-pulse" />
          Technical Studio v1.2
        </div>
      </div>
    </motion.div>
  );
}

/* 🕵️ Internal Component for Sidebar/Footer Analytics */
const TechnicalAnalysis = ({
  meta,
  warnings,
  mdflowOutput,
}: {
  meta: any;
  warnings: string[];
  mdflowOutput: string | null;
}) => {
  return (
    <AnimatePresence mode="wait">
      {!mdflowOutput ? (
        <motion.div
          key="idle"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="flex items-center justify-between"
        >
          <div className="flex items-center gap-4">
            <div className="h-2 w-2 rounded-full bg-white/40 animate-pulse" />
            <span className="text-[10px] font-black uppercase tracking-[0.3em] text-white/40">
              Engine Status: Standby
            </span>
          </div>
          <div className="flex gap-2">
            <div className="w-12 h-1 bg-white/10 rounded-full" />
            <div className="w-8 h-1 bg-white/10 rounded-full" />
          </div>
        </motion.div>
      ) : (
        <motion.div
          key="active"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="grid grid-cols-1 md:grid-cols-2 gap-8 items-center"
        >
          {/* Telemetry Data */}
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <Activity className="w-3.5 h-3.5 text-accent-orange opacity-60" />
              <h3 className="text-[9px] font-black uppercase tracking-[0.3em] text-white/60">
                Telemetry Data
              </h3>
            </div>
            <div className="flex gap-4">
              <div className="flex-1 p-3 rounded-xl bg-white/5 border border-white/10">
                <p className="text-[7px] text-white/40 mb-1 font-bold uppercase tracking-widest text-center">
                  Nodes
                </p>
                <p className="text-sm font-black font-mono text-white text-center leading-none">
                  {meta.nodeCount || meta.total_rows || 0}
                </p>
              </div>
              <div className="flex-1 p-3 rounded-xl bg-white/5 border border-white/10">
                <p className="text-[7px] text-white/40 mb-1 font-bold uppercase tracking-widest text-center">
                  Header
                </p>
                <p className="text-sm font-black font-mono text-white text-center leading-none">
                  {meta.headerCount || meta.header_row || 0}
                </p>
              </div>
            </div>
          </div>

          {/* Warnings Monitor */}
          {warnings && warnings.length > 0 && (
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <AlertCircle className="w-3.5 h-3.5 text-accent-gold" />
                <h3 className="text-[9px] font-black uppercase tracking-[0.3em] text-accent-gold/80">
                  Validation Alerts
                </h3>
              </div>
              <div className="max-h-[60px] overflow-auto custom-scrollbar space-y-2">
                {warnings.map((w: string, i: number) => (
                  <div
                    key={i}
                    className="text-[8px] font-bold text-accent-gold/80 uppercase leading-snug pl-3 border-l border-accent-gold/50"
                  >
                    {w}
                  </div>
                ))}
              </div>
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
};
