"use client";

import { motion } from "framer-motion";
import {
  ChevronRight,
  Cpu,
  Database,
  FileCode,
  Search,
  Shield,
  Terminal,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";

// Documentation Data - Auto-Generated from System Truth
// Documentation Data - Auto-Generated from System Truth
const docsSections = [
  {
    title: "System Architecture",
    items: [
      {
        id: "intro",
        title: "Overview",
        icon: <FileCode className="w-4 h-4" />,
      },
      { id: "stack", title: "Tech Stack", icon: <Cpu className="w-4 h-4" /> },
    ],
  },
  {
    title: "Engine Mechanics",
    items: [
      {
        id: "ingestion",
        title: "Ingestion Layers",
        icon: <Database className="w-4 h-4" />,
      },
      {
        id: "parsing",
        title: "Parsing Logic",
        icon: <Terminal className="w-4 h-4" />,
      },
      {
        id: "validation",
        title: "Validation Rules",
        icon: <Shield className="w-4 h-4" />,
      },
    ],
  },
];

const docContent: Record<string, { title: string; content: React.ReactNode }> =
  {
    intro: {
      title: "System Overview",
      content: (
        <div className="space-y-8">
          <p className="text-xl text-muted leading-relaxed font-light">
            MDFlow is a high-performance <strong>Engineering Studio</strong>{" "}
            that bridges the gap between raw technical data (Excel/CSV) and
            standardized documentation. It operates as a hybrid{" "}
            <strong>Local-Hosted</strong> system, ensuring maximum speed and
            privacy.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="p-8 rounded-3xl bg-white/5 border border-white/10 hover:border-accent-orange/30 transition-colors group">
              <div className="mb-6 p-4 rounded-2xl bg-black/40 w-fit text-accent-orange group-hover:scale-110 transition-transform">
                <Database className="w-6 h-6" />
              </div>
              <h4 className="text-white font-black uppercase tracking-wider mb-2">
                Frontend Studio
              </h4>
              <p className="text-sm text-muted">
                built with Next.js 16 & Turbopack. Handles the visual interface,
                state management (Zustand), and real-time user feedback.
              </p>
            </div>

            <div className="p-8 rounded-3xl bg-white/5 border border-white/10 hover:border-accent-gold/30 transition-colors group">
              <div className="mb-6 p-4 rounded-2xl bg-black/40 w-fit text-accent-gold group-hover:scale-110 transition-transform">
                <Cpu className="w-6 h-6" />
              </div>
              <h4 className="text-white font-black uppercase tracking-wider mb-2">
                Backend Engine
              </h4>
              <p className="text-sm text-muted">
                Powered by Go (Golang). A dedicated parsing server (port 8080)
                that executes complex AST mappings and huge-file processing with
                zero latency.
              </p>
            </div>
          </div>
        </div>
      ),
    },
    stack: {
      title: "Technology Stack",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            The MDFlow architecture is designed for "Engineering Grade"
            reliability.
          </p>
          <ul className="space-y-4">
            <li className="flex items-start gap-4 p-4 rounded-xl bg-white/5 border border-white/5">
              <span className="text-accent-orange font-black text-xs uppercase tracking-widest mt-1">
                Core
              </span>
              <div>
                <strong className="text-white block mb-1">
                  Go (Golang) Backend
                </strong>
                <span className="text-sm text-white/40">
                  Handlers `api/mdflow/xlsx` and `api/mdflow/paste` provide
                  robust, type-safe parsing streams.
                </span>
              </div>
            </li>
            <li className="flex items-start gap-4 p-4 rounded-xl bg-white/5 border border-white/5">
              <span className="text-accent-orange font-black text-xs uppercase tracking-widest mt-1">
                Web
              </span>
              <div>
                <strong className="text-white block mb-1">
                  Next.js 16 + React 19
                </strong>
                <span className="text-sm text-white/40">
                  Utilizing Server Components for layout and Client Components
                  for the interactive Workbench.
                </span>
              </div>
            </li>
            <li className="flex items-start gap-4 p-4 rounded-xl bg-white/5 border border-white/5">
              <span className="text-accent-orange font-black text-xs uppercase tracking-widest mt-1">
                State
              </span>
              <div>
                <strong className="text-white block mb-1">Zustand Store</strong>
                <span className="text-sm text-white/40">
                  Client-side state manager handling file blobs, sheet
                  selection, and temporary parsed results (MDFlowStore).
                </span>
              </div>
            </li>
          </ul>
        </div>
      ),
    },
    ingestion: {
      title: "Ingestion Layers",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            The engine supports two distinct modes of data ingestion, handled by
            the `mdflowApi` service.
          </p>

          <div className="space-y-8">
            <div className="relative pl-8 border-l-2 border-accent-orange/20">
              <h4 className="text-white font-bold mb-2">
                1. Paste Mode (Clipboard)
              </h4>
              <p className="text-sm text-muted mb-4">
                Designed for quick prototyping. Accepts raw text buffer from CSV
                or TSV sources.
              </p>
              <div className="bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/paste <br />
                {`{ "paste_text": "...", "template": "default" }`}
              </div>
            </div>

            <div className="relative pl-8 border-l-2 border-accent-gold/20">
              <h4 className="text-white font-bold mb-2">
                2. XLSX Mode (Binary)
              </h4>
              <p className="text-sm text-muted mb-4">
                Designed for production specifications. Accepts
                `multipart/form-data` binary streams.
              </p>
              <ul className="list-disc list-inside text-sm text-white/60 space-y-2 mb-4">
                <li>
                  Supports multi-sheet discovery (`/api/mdflow/xlsx/sheets`)
                </li>
                <li>Sheet-specific targeting via `sheet_name` parameter</li>
                <li>Preserves cell formatting and relationships</li>
              </ul>
            </div>
          </div>
        </div>
      ),
    },
    parsing: {
      title: "Parsing Logic",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            The Go backend employs a highly efficient parsing strategy:
          </p>
          <div className="grid gap-3">
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                1. Stream Decode
              </span>
              <span className="text-xs text-white/40 font-mono">io.Reader</span>
            </div>
            <div className="flex justify-center">
              <ChevronRight className="rotate-90 text-white/20" />
            </div>
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                2. Header Recognition
              </span>
              <span className="text-xs text-white/40 font-mono">
                Row 1 Detection
              </span>
            </div>
            <div className="flex justify-center">
              <ChevronRight className="rotate-90 text-white/20" />
            </div>
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                3. Map Generation
              </span>
              <span className="text-xs text-white/40 font-mono">
                Column - Key
              </span>
            </div>
          </div>
        </div>
      ),
    },
    validation: {
      title: "Validation Rules",
      content: (
        <div className="space-y-4">
          <p className="text-muted">
            The engine returns a structured{" "}
            <code className="text-accent-orange">MDFlowConvertResponse</code>{" "}
            containing metadata and warnings.
          </p>
          <div className="p-6 bg-black/40 border border-white/10 rounded-xl space-y-4 font-mono text-sm">
            <div>
              <span className="text-accent-purple">meta.column_map</span>
              <p className="text-white/40 pl-4">
                Logs exactly which Excel columns were mapped to internal nodes.
              </p>
            </div>
            <div>
              <span className="text-accent-red">warnings[]</span>
              <p className="text-white/40 pl-4">
                Returns an array of specific issues (e.g., "Row 42: Missing
                required ID").
              </p>
            </div>
            <div>
              <span className="text-accent-green">mdflow</span>
              <p className="text-white/40 pl-4">
                The final compiled Markdown output string.
              </p>
            </div>
          </div>
        </div>
      ),
    },
  };

export default function DocsPage() {
  const [activeSection, setActiveSection] = useState("intro");

  return (
    <div className="min-h-screen pt-12 pb-24">
      <div className="app-container">
        <div className="grid lg:grid-cols-[280px_1fr] gap-12">
          {/* Sidebar Navigation */}
          <motion.aside
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="hidden lg:block space-y-8 sticky top-32 h-fit"
          >
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
              <input
                type="text"
                placeholder="Search docs..."
                className="w-full h-10 pl-10 pr-4 rounded-xl bg-white/5 border border-white/10 text-sm text-white placeholder:text-white/20 focus:outline-hidden focus:border-accent-orange/50 transition-all"
              />
            </div>

            <nav className="space-y-8">
              {docsSections.map((section, idx) => (
                <div key={idx} className="space-y-3">
                  <h3 className="text-[10px] font-black uppercase tracking-[0.2em] text-white/40 ml-3">
                    {section.title}
                  </h3>
                  <div className="space-y-1">
                    {section.items.map((item) => (
                      <button
                        key={item.id}
                        onClick={() => setActiveSection(item.id)}
                        className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all group ${
                          activeSection === item.id
                            ? "bg-accent-orange/10 text-accent-orange"
                            : "text-muted hover:text-white hover:bg-white/5"
                        }`}
                      >
                        <span
                          className={`transition-colors ${activeSection === item.id ? "text-accent-orange" : "text-white/40 group-hover:text-white"}`}
                        >
                          {item.icon}
                        </span>
                        {item.title}
                        {activeSection === item.id && (
                          <motion.div
                            layoutId="active-pill"
                            className="ml-auto w-1.5 h-1.5 rounded-full bg-accent-orange"
                          />
                        )}
                      </button>
                    ))}
                  </div>
                </div>
              ))}
            </nav>
          </motion.aside>

          {/* Main Content Area */}
          <main className="min-w-0">
            <motion.div
              key={activeSection}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3 }}
              className="space-y-8"
            >
              <div className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest text-accent-orange mb-2">
                Docs <ChevronRight className="w-3 h-3" />{" "}
                {
                  docsSections.find((s) =>
                    s.items.find((i) => i.id === activeSection),
                  )?.title
                }
              </div>

              <h1 className="text-4xl sm:text-5xl font-black text-white tracking-tighter uppercase">
                {docContent[activeSection]?.title || "Documentation"}
              </h1>

              <div className="prose prose-invert prose-lg max-w-none">
                {docContent[activeSection]?.content}
              </div>

              {/* Navigation Footer for Docs */}
              <div className="pt-20 mt-20 border-t border-white/5 flex items-center justify-between">
                <div className="text-sm text-muted">
                  Last updated: <span className="text-white">Jan 29, 2026</span>
                </div>
                <div className="flex gap-4">
                  <button className="text-xs font-bold uppercase tracking-wider text-muted hover:text-accent-orange transition-colors">
                    Edit on GitHub
                  </button>
                </div>
              </div>
            </motion.div>
          </main>
        </div>
      </div>
    </div>
  );
}
