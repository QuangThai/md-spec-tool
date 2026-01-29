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
        setError("Please paste some data first");
        setLoading(false);
        return;
      }
      result = await convertPaste(pasteText, template);
    } else {
      if (!file) {
        setError("Please upload a file first");
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
      document.getElementById("output")?.scrollIntoView({ behavior: "smooth" });
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
      className="flex flex-col gap-16 pb-32"
    >
      <div id="workspace" className="relative group">
        <div className="absolute -inset-10 -z-10 opacity-20 bg-size-[60px_60px] bg-[radial-gradient(#f27b2f15_1px,transparent_1px)]" />

        <section className="grid grid-cols-1 gap-12 lg:grid-cols-[1fr_0.42fr]">
          <div className="flex flex-col gap-12">
            <motion.div
              variants={fadeInUp}
              className="surface overflow-hidden relative p-6! sm:p-10! focus-within:border-accent-orange/20"
            >
              <div className="flex flex-wrap items-center justify-between gap-10 mb-10 pb-10 border-b border-white/5">
                <div className="flex flex-col gap-3">
                  <div className="flex items-center gap-4">
                    <div className="p-2.5 rounded-xl bg-white/5 border border-white/10">
                      <Cpu className="w-5 h-5 text-accent-orange" />
                    </div>
                    <div className="flex flex-col">
                      <h2 className="text-2xl font-bold text-white uppercase tracking-tight">
                        Workbench
                      </h2>
                      <div className="flex items-center gap-2 mt-0.5">
                        <span className="h-1.5 w-1.5 rounded-full bg-accent-orange animate-pulse" />
                        <span className="text-[9px] font-bold text-accent-orange/80 uppercase tracking-widest">
                          Interface_Online
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="flex items-center rounded-xl bg-black/40 p-1 border border-white/5">
                  <button
                    onClick={() => setMode("paste")}
                    className={`flex items-center gap-3 rounded-lg px-6 py-2.5 text-[10px] font-bold uppercase tracking-widest transition-all ${
                      mode === "paste"
                        ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/20"
                        : "text-muted hover:text-white"
                    }`}
                  >
                    <Copy className="w-3 h-3" /> Ingest.txt
                  </button>
                  <button
                    onClick={() => setMode("xlsx")}
                    className={`flex items-center gap-3 rounded-lg px-6 py-2.5 text-[10px] font-bold uppercase tracking-widest transition-all ${
                      mode === "xlsx"
                        ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/20"
                        : "text-muted hover:text-white"
                    }`}
                  >
                    <FileSpreadsheet className="w-3 h-3" /> Stream.io
                  </button>
                </div>
              </div>

              <AnimatePresence mode="wait">
                {error && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0.98 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.98 }}
                    className="flex md:items-center gap-4 rounded-xl bg-accent-red/10 p-5 text-[10px] font-bold text-accent-red border border-accent-red/20 mb-10"
                  >
                    <AlertCircle className="w-4 h-4 shrink-0" />
                    <span className="uppercase tracking-widest leading-relaxed">
                      {error}
                    </span>
                  </motion.div>
                )}
              </AnimatePresence>

              <div className="min-h-[400px]">
                <AnimatePresence mode="wait">
                  {mode === "paste" ? (
                    <motion.div
                      key="paste"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      className="space-y-6"
                    >
                      <div className="flex items-center justify-between">
                        <label className="label">Cluster_Ingest_01</label>
                        <span className="text-[8px] font-bold text-muted uppercase tracking-widest border border-white/10 px-3 py-1 rounded-full">
                          TSV_CSV_COMPLIANT
                        </span>
                      </div>
                      <textarea
                        value={pasteText}
                        onChange={(e) => setPasteText(e.target.value)}
                        placeholder="Paste technical data matrix here... (CMD+V)"
                        className="input min-h-[320px] font-mono leading-relaxed resize-none"
                      />
                    </motion.div>
                  ) : (
                    <motion.div
                      key="xlsx"
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      className="grid gap-10"
                    >
                      <div className="space-y-6">
                        <label className="label">Binary_Stream_XLSX</label>
                        <div className="group/file relative">
                          <input
                            type="file"
                            accept=".xlsx,.xls"
                            onChange={handleFileChange}
                            className="input cursor-pointer file:mr-6 file:rounded-lg file:border-0 file:bg-accent-orange file:px-6 file:py-1.5 file:text-[9px] file:font-bold file:uppercase file:text-white hover:bg-white/5 transition-all border-dashed border-2"
                          />
                          {file && (
                            <motion.div
                              initial={{ opacity: 0, x: -10 }}
                              animate={{ opacity: 1, x: 0 }}
                              className="mt-4 flex items-center gap-4 text-[9px] font-bold text-white bg-white/5 p-4 rounded-xl border border-white/10"
                            >
                              <FileSpreadsheet className="w-3.5 h-3.5 text-accent-orange" />
                              <span className="uppercase tracking-widest">
                                {file.name}
                              </span>
                              <span className="ml-auto text-muted">
                                {(file.size / 1024).toFixed(1)} KB
                              </span>
                            </motion.div>
                          )}
                        </div>
                      </div>

                      {sheets.length > 0 && (
                        <motion.div
                          initial={{ opacity: 0, scale: 0.98 }}
                          animate={{ opacity: 1, scale: 1 }}
                          className="space-y-6"
                        >
                          <label className="label">
                            Target_Plane_Selection
                          </label>
                          <select
                            value={selectedSheet}
                            onChange={(e) => setSelectedSheet(e.target.value)}
                            className="input appearance-none bg-no-repeat bg-position-[right_1.5rem_center] cursor-pointer"
                            style={{
                              backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23f27b2f' stroke-width='2.5'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' d='M19.5 8.25l-7.5 7.5-7.5-7.5' /%3E%3C/svg%3E")`,
                              backgroundSize: "1rem",
                            }}
                          >
                            {sheets.map((sheet) => (
                              <option
                                key={sheet}
                                value={sheet}
                                className="bg-surface text-white"
                              >
                                {sheet}
                              </option>
                            ))}
                          </select>
                        </motion.div>
                      )}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              <div className="mt-10 pt-10 border-t border-white/5">
                <div className="grid gap-10 lg:grid-cols-[1fr_auto]">
                  <div className="space-y-6">
                    <label className="label">Schema_Model_Logic</label>
                    <div className="relative">
                      <select
                        value={template}
                        onChange={(e) => setTemplate(e.target.value)}
                        className="input appearance-none bg-no-repeat bg-position-[right_1.5rem_center] h-14 font-bold uppercase tracking-widest"
                        style={{
                          backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23f27b2f' stroke-width='2.5'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' d='M19.5 8.25l-7.5 7.5-7.5-7.5' /%3E%3C/svg%3E")`,
                          backgroundSize: "1rem",
                        }}
                      >
                        {templates.map((t) => (
                          <option
                            key={t}
                            value={t}
                            className="bg-surface text-white"
                          >
                            {t.toUpperCase().replace(/-/g, "_")}_MODEL
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>

                  <div className="flex gap-4 items-end">
                    <motion.button
                      whileHover={{ scale: 1.01 }}
                      whileTap={{ scale: 0.99 }}
                      onClick={handleConvert}
                      disabled={loading}
                      className="btn-primary flex-1 lg:w-[280px] h-14 uppercase tracking-[0.2em]"
                    >
                      {loading ? (
                        <span className="flex items-center gap-3">
                          <RefreshCcw className="w-4 h-4 animate-spin" />
                          Processing...
                        </span>
                      ) : (
                        <span className="flex items-center gap-2">
                          Commit Conversion <ArrowRight className="w-4 h-4" />
                        </span>
                      )}
                    </motion.button>
                    <motion.button
                      whileHover={{
                        backgroundColor: "rgba(212,86,86,0.1)",
                        borderColor: "rgba(212,86,86,0.2)",
                        color: "#d45656",
                      }}
                      whileTap={{ scale: 0.95 }}
                      onClick={reset}
                      className="w-14 h-14 rounded-xl flex items-center justify-center border border-white/10 bg-white/5 text-muted transition-all"
                      title="Reset Cluster"
                    >
                      <RefreshCcw className="w-4 h-4" />
                    </motion.button>
                  </div>
                </div>
              </div>
            </motion.div>

            <motion.div
              id="output"
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              className="flex flex-col gap-8"
            >
              {mdflowOutput ? (
                <div className="code-editor relative group border-white/10">
                  <div className="absolute right-8 top-8 flex items-center gap-3 z-10 opacity-0 group-hover:opacity-100 transition-all translate-y-1 group-hover:translate-y-0">
                    <button
                      onClick={handleCopy}
                      className="flex items-center gap-2.5 h-10 px-6 rounded-lg bg-white/5 text-[10px] font-bold uppercase tracking-widest text-white hover:bg-white/10 backdrop-blur-md border border-white/10"
                    >
                      {copied ? (
                        <Check className="w-3.5 h-3.5 text-accent-orange" />
                      ) : (
                        <Copy className="w-3.5 h-3.5" />
                      )}
                      {copied ? "Captured" : "Copy_Raw"}
                    </button>
                    <button
                      onClick={handleDownload}
                      className="flex items-center gap-2.5 h-10 px-6 rounded-lg bg-accent-orange text-[10px] font-bold uppercase tracking-widest text-white hover:bg-accent-orange/90"
                    >
                      <Download className="w-3.5 h-3.5" /> Export_MD
                    </button>
                  </div>

                  <div className="flex items-center justify-between mb-8 border-b border-white/5 pb-8">
                    <div className="flex items-center gap-4">
                      <div className="flex gap-2">
                        <span className="h-2.5 w-2.5 rounded-full bg-accent-red/30" />
                        <span className="h-2.5 w-2.5 rounded-full bg-accent-gold/30" />
                        <span className="h-2.5 w-2.5 rounded-full bg-accent-orange/30" />
                      </div>
                      <div className="h-4 w-px bg-white/10 mx-1" />
                      <span className="text-[10px] font-bold uppercase tracking-[0.3em] text-muted font-mono">
                        Pipeline_Out.mdflow
                      </span>
                    </div>
                  </div>

                  <div className="max-h-[800px] overflow-y-auto custom-scrollbar pr-4">
                    <pre className="whitespace-pre-wrap leading-relaxed text-white/80 font-mono text-[13px]">
                      {mdflowOutput}
                    </pre>
                  </div>
                </div>
              ) : (
                <div className="surface flex h-[400px] flex-col items-center justify-center p-12 text-center border-dashed border-2 border-white/5 bg-black/5 rounded-2xl">
                  <div className="mb-8 p-5 rounded-full bg-white/5 border border-white/5">
                    <Terminal className="w-10 h-10 text-white/20" />
                  </div>
                  <h3 className="text-xl font-bold text-muted uppercase tracking-[0.2em]">
                    Terminal Idle
                  </h3>
                  <p className="mt-4 max-w-[280px] text-[10px] font-medium text-muted/80 leading-relaxed uppercase tracking-wider">
                    Waiting for technical assets. Ingest data to generate
                    specification stream.
                  </p>
                </div>
              )}
            </motion.div>
          </div>

          <aside className="space-y-10">
            <AnimatePresence>
              {meta && mdflowOutput && (
                <motion.div
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  className="surface p-8 relative overflow-hidden"
                >
                  <h4 className="label mb-8 border-b border-white/5 pb-4">
                    Engine_Telemetry
                  </h4>

                  <div className="space-y-6">
                    <div className="flex justify-between items-center group">
                      <span className="text-[9px] font-bold text-muted uppercase tracking-widest group-hover:text-white/60">
                        Nodes Processed
                      </span>
                      <span className="text-sm font-bold text-white font-mono bg-white/5 px-3 py-1 rounded-lg border border-white/5">
                        {meta.total_rows}
                      </span>
                    </div>
                    <div className="flex justify-between items-center group">
                      <span className="text-[9px] font-bold text-muted uppercase tracking-widest group-hover:text-white/60">
                        Header Row
                      </span>
                      <span className="text-sm font-bold text-white font-mono bg-white/5 px-3 py-1 rounded-lg border border-white/5">
                        {meta.header_row}
                      </span>
                    </div>
                    {meta.sheet_name && (
                      <div className="pt-2 space-y-2">
                        <span className="text-[8px] font-bold text-muted uppercase tracking-[0.2em] block">
                          Data Plane Source
                        </span>
                        <div className="text-[10px] font-bold text-accent-orange bg-accent-orange/5 p-3 rounded-lg border border-accent-orange/10 break-all font-mono">
                          {meta.sheet_name}
                        </div>
                      </div>
                    )}

                    {meta.rows_by_feature &&
                      Object.keys(meta.rows_by_feature).length > 0 && (
                        <div className="pt-6 space-y-4">
                          <span className="text-[8px] font-bold text-muted uppercase tracking-[0.2em] block opacity-50">
                            Object Distribution
                          </span>
                          <div className="grid gap-3">
                            {Object.entries(meta.rows_by_feature).map(
                              ([feature, count]) => (
                                <div
                                  key={feature}
                                  className="flex flex-col gap-2 p-3 rounded-lg bg-black/20 border border-white/5 hover:border-accent-orange/20 transition-all"
                                >
                                  <div className="flex justify-between items-center">
                                    <span className="text-[9px] font-bold text-muted uppercase tracking-wider">
                                      {feature}
                                    </span>
                                    <span className="text-[10px] font-bold text-white font-mono">
                                      {count}
                                    </span>
                                  </div>
                                  <div className="h-1 w-full bg-white/5 rounded-full overflow-hidden">
                                    <div
                                      className="h-full bg-accent-orange rounded-full"
                                      style={{
                                        width: `${Math.min(100, (Number(count) / meta.total_rows) * 100)}%`,
                                      }}
                                    />
                                  </div>
                                </div>
                              ),
                            )}
                          </div>
                        </div>
                      )}
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Validation Monitor */}
            <AnimatePresence>
              {warnings && warnings.length > 0 && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  className="rounded-2xl border border-accent-gold/10 bg-accent-gold/5 p-8 backdrop-blur-md shadow-premium"
                >
                  <h4 className="label text-accent-gold mb-6 border-b border-accent-gold/10 pb-4 flex items-center gap-2">
                    <Activity className="w-3.5 h-3.5" />
                    Validation_Warnings
                  </h4>
                  <ul className="space-y-4">
                    {warnings.map((warning, i) => (
                      <li
                        key={i}
                        className="flex items-start gap-3 text-[10px] font-bold text-accent-gold/80 leading-relaxed uppercase"
                      >
                        <div className="mt-1 h-1 w-1 shrink-0 rounded-full bg-accent-gold shadow-[0_0_8px_rgba(195,125,13,0.5)]" />
                        {warning}
                      </li>
                    ))}
                  </ul>
                </motion.div>
              )}
            </AnimatePresence>

            <motion.div
              variants={fadeInUp}
              className="surface p-8 relative overflow-hidden"
            >
              <h4 className="label border-b border-white/5 pb-4 mb-8">
                Pipeline_Blueprint
              </h4>
              <div className="space-y-8 relative z-10">
                {steps.map((step, index) => (
                  <div key={step.title} className="relative group">
                    {index < steps.length - 1 && (
                      <div className="absolute left-6 top-10 h-10 w-px bg-white/5 group-hover:bg-accent-orange/20 transition-colors" />
                    )}
                    <div className="flex items-start gap-5">
                      <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-white/5 text-muted border border-white/10 shadow-inner transition-all group-hover:bg-accent-orange group-hover:text-white">
                        {index === 0 ? (
                          <Database className="w-4 h-4" />
                        ) : index === 1 ? (
                          <Terminal className="w-4 h-4" />
                        ) : (
                          <Zap className="w-4 h-4" />
                        )}
                      </div>
                      <div>
                        <p className="text-[11px] font-bold text-white uppercase tracking-wider group-hover:text-accent-orange">
                          {step.title}
                        </p>
                        <p className="text-[9px] text-muted mt-1 uppercase font-medium">
                          {step.desc}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              <div className="mt-12 rounded-xl bg-black/40 p-6 border border-white/5">
                <div className="flex items-center gap-2 mb-3">
                  <div className="h-1.5 w-1.5 rounded-full bg-accent-orange animate-pulse" />
                  <p className="text-[9px] font-bold uppercase tracking-widest text-accent-orange">
                    Core_Active
                  </p>
                </div>
                <p className="text-[9px] font-bold text-muted/90 leading-relaxed uppercase">
                  v1.2.0 Automated Parity Mapping Engine
                </p>
              </div>
            </motion.div>

            <div className="flex justify-center">
              <div className="inline-flex items-center gap-2 text-[8px] font-bold text-muted/80 uppercase tracking-[0.4em] bg-white/5 border border-white/5 px-5 py-2 rounded-full">
                <span className="w-2 h-2 rounded-full bg-accent-orange animate-pulse" />
                Technical Studio v1.2
              </div>
            </div>
          </aside>
        </section>
      </div>
    </motion.div>
  );
}
