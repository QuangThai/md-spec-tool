"use client";

import { BatchProcessor } from "@/components/BatchProcessor";
import { TemplateCards } from "@/components/TemplateCards";
import { getMDFlowTemplates } from "@/lib/mdflowApi";
import { motion } from "framer-motion";
import {
  ArrowLeft,
  Boxes,
  FileSpreadsheet,
  Layers,
  Zap,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

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

export default function BatchPage() {
  const [template, setTemplate] = useState("default");
  const [templates, setTemplates] = useState<string[]>(["default"]);

  useEffect(() => {
    getMDFlowTemplates().then((res) => {
      if (res.data?.templates) {
        const sorted = [...res.data.templates].sort((a, b) => {
          if (a === "default") return -1;
          if (b === "default") return 1;
          return 0;
        });
        setTemplates(sorted);
      }
    });
  }, []);

  return (
    <div className="min-h-screen bg-white/2 rounded-2xl overflow-hidden">
      {/* Header */}
      <header className="border-b border-white/10 bg-white/2 backdrop-blur-xl sticky top-0 z-49">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              href="/studio"
              className="flex items-center gap-2 text-white/60 hover:text-white transition-colors"
            >
              <ArrowLeft className="w-4 h-4" />
              <span className="text-sm font-medium hidden sm:inline">
                Back to Studio
              </span>
            </Link>
            <div className="h-6 w-px bg-white/10" />
            <div className="flex items-center gap-2">
              <Boxes className="w-4 h-4 text-accent-orange" />
              <span className="text-sm font-black text-white uppercase tracking-wider">
                Batch Processing
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">
        <motion.div
          variants={stagger.container}
          initial="initial"
          animate="animate"
          className="space-y-8"
        >
          {/* Hero */}
          <motion.div variants={stagger.item} className="text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-accent-orange/10 border border-accent-orange/20 mb-6">
              <Layers className="w-8 h-8 text-accent-orange" />
            </div>
            <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight mb-3">
              Batch Processing
            </h1>
            <p className="text-sm text-white/60 max-w-lg mx-auto">
              Convert multiple files at once. Upload Excel, TSV, or CSV files and download all outputs as a ZIP.
            </p>
          </motion.div>

          {/* Features */}
          <motion.div
            variants={stagger.item}
            className="grid grid-cols-1 sm:grid-cols-3 gap-4"
          >
            <div className="p-4 rounded-xl bg-white/5 border border-white/10 flex items-start gap-3">
              <div className="p-2 rounded-lg bg-green-500/10 text-green-400">
                <FileSpreadsheet className="w-4 h-4" />
              </div>
              <div>
                <p className="text-xs font-bold text-white mb-0.5">Multi-file Upload</p>
                <p className="text-[10px] text-white/50">
                  Drop multiple files at once
                </p>
              </div>
            </div>
            <div className="p-4 rounded-xl bg-white/5 border border-white/10 flex items-start gap-3">
              <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400">
                <Layers className="w-4 h-4" />
              </div>
              <div>
                <p className="text-xs font-bold text-white mb-0.5">All Sheets</p>
                <p className="text-[10px] text-white/50">
                  Process all sheets in Excel
                </p>
              </div>
            </div>
            <div className="p-4 rounded-xl bg-white/5 border border-white/10 flex items-start gap-3">
              <div className="p-2 rounded-lg bg-accent-orange/10 text-accent-orange">
                <Zap className="w-4 h-4" />
              </div>
              <div>
                <p className="text-xs font-bold text-white mb-0.5">ZIP Download</p>
                <p className="text-[10px] text-white/50">
                  Download all outputs at once
                </p>
              </div>
            </div>
          </motion.div>

          {/* Template selector */}
          <motion.div variants={stagger.item}>
            <div className="rounded-2xl border border-white/10 bg-white/2 p-6">
              <label className="text-[10px] font-black uppercase tracking-widest text-white/50 block mb-4">
                Output Template
              </label>
              <TemplateCards
                templates={templates}
                selected={template}
                onSelect={setTemplate}
                compact
              />
            </div>
          </motion.div>

          {/* Batch processor */}
          <motion.div variants={stagger.item}>
            <div className="rounded-2xl border border-white/10 bg-white/2 p-6">
              <BatchProcessor template={template} />
            </div>
          </motion.div>

          {/* Tips */}
          <motion.div variants={stagger.item}>
            <div className="rounded-xl border border-blue-500/20 bg-blue-500/5 p-5">
              <h3 className="text-xs font-bold text-blue-400 mb-3 uppercase tracking-wider">
                Tips
              </h3>
              <ul className="space-y-2 text-[11px] text-blue-400/70">
                <li className="flex items-start gap-2">
                  <span className="text-blue-400">•</span>
                  Drag and drop multiple files at once for faster processing
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-blue-400">•</span>
                  Enable "Process all sheets" to convert every sheet in Excel workbooks
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-blue-400">•</span>
                  Individual files can be downloaded separately or all at once as ZIP
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-blue-400">•</span>
                  Files are processed in your browser - nothing is stored on the server
                </li>
              </ul>
            </div>
          </motion.div>
        </motion.div>
      </main>
    </div>
  );
}
