"use client";

import { motion } from "framer-motion";
import {
  ChevronLeft,
  ChevronRight,
  Cpu,
  Database,
  Download,
  Eye,
  FileCode,
  GitCompare,
  History,
  Keyboard,
  Link2,
  Package,
  Search,
  Shield,
  Sparkles,
  Table,
  Terminal,
  Zap,
} from "lucide-react";
import { useSearchParams } from "next/navigation";
import {
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

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
      {
        id: "templates",
        title: "Templates & Schemas",
        icon: <FileCode className="w-4 h-4" />,
      },
    ],
  },
  {
    title: "Studio Features",
    items: [
      {
        id: "preview",
        title: "Data Preview",
        icon: <Eye className="w-4 h-4" />,
      },
      {
        id: "column-mapping",
        title: "Column Mapping",
        icon: <Table className="w-4 h-4" />,
      },
      {
        id: "integrations",
        title: "Integrations",
        icon: <Link2 className="w-4 h-4" />,
      },
      {
        id: "workflow",
        title: "Workflow Tools",
        icon: <Zap className="w-4 h-4" />,
      },
    ],
  },
  {
    title: "Automation & Collaboration",
    items: [
      {
        id: "ai-suggestions",
        title: "AI Suggestions",
        icon: <Sparkles className="w-4 h-4" />,
      },
      {
        id: "template-editor",
        title: "Template Editor",
        icon: <FileCode className="w-4 h-4" />,
      },
      {
        id: "batch-processing",
        title: "Batch Processing",
        icon: <Package className="w-4 h-4" />,
      },
      {
        id: "share-links",
        title: "Share Links",
        icon: <Link2 className="w-4 h-4" />,
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
          <p className="text-base sm:text-lg lg:text-xl text-muted leading-relaxed font-light">
            MDFlow Studio is a high-performance{" "}
            <strong>Engineering Studio</strong> that bridges the gap between raw
            technical data (Excel/CSV/TSV) and standardized documentation. It
            runs as a web app with a Next.js UI and a Go API, supporting
            paste-based input, file uploads, Google Sheets integration, batch
            processing, shareable links, and real-time preview with intelligent
            column mapping.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6">
            <div className="p-4 sm:p-6 lg:p-8 rounded-2xl sm:rounded-3xl bg-white/5 border border-white/10 hover:border-accent-orange/30 transition-colors group">
              <div className="mb-4 sm:mb-6 p-3 sm:p-4 rounded-xl sm:rounded-2xl bg-black/40 w-fit text-accent-orange group-hover:scale-110 transition-transform">
                <Database className="w-5 h-5 sm:w-6 sm:h-6" />
              </div>
              <h4 className="text-sm sm:text-base font-black uppercase tracking-wider mb-2 text-white">
                Frontend Studio
              </h4>
              <p className="text-xs sm:text-sm text-muted">
                Built with Next.js 16 and React 19. Handles the UI, local state
                (Zustand), real-time preview, conversion history, diff viewer,
                template editor, batch processing, share links, and the request
                flow for paste/XLSX/TSV/Google Sheets conversions.
              </p>
            </div>

            <div className="p-4 sm:p-6 lg:p-8 rounded-2xl sm:rounded-3xl bg-white/5 border border-white/10 hover:border-accent-gold/30 transition-colors group">
              <div className="mb-4 sm:mb-6 p-3 sm:p-4 rounded-xl sm:rounded-2xl bg-black/40 w-fit text-accent-gold group-hover:scale-110 transition-transform">
                <Cpu className="w-5 h-5 sm:w-6 sm:h-6" />
              </div>
              <h4 className="text-sm sm:text-base font-black uppercase tracking-wider mb-2 text-white">
                Backend Engine
              </h4>
              <p className="text-xs sm:text-sm text-muted">
                Powered by Go (Golang) with Gin. Provides the MDFlow API for
                template-based conversion, sheet discovery, preview endpoints,
                and CLI tool support for automation workflows.
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
                  Gin handlers `/api/mdflow/paste`, `/api/mdflow/preview`,
                  `/api/mdflow/tsv`, `/api/mdflow/xlsx`, `/api/mdflow/gsheet`,
                  `/api/mdflow/ai/suggest`, and `/api/mdflow/templates` power
                  conversion, previews, integrations, and automation.
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
                  App router with client components for the interactive Studio
                  experience (TypeScript + Tailwind CSS).
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
                  Client-side state for file selection, sheet selection,
                  template choice, parsed results, preview data, column
                  overrides, and conversion history (MDFlowStore +
                  HistoryStore).
                </span>
              </div>
            </li>
            <li className="flex items-start gap-4 p-4 rounded-xl bg-white/5 border border-white/5">
              <span className="text-accent-orange font-black text-xs uppercase tracking-widest mt-1">
                UI
              </span>
              <div>
                <strong className="text-white block mb-1">
                  Framer Motion + Tailwind CSS
                </strong>
                <span className="text-sm text-white/40">
                  Smooth animations, transitions, and modern UI components with
                  dark theme and responsive design.
                </span>
              </div>
            </li>
            <li className="flex items-start gap-4 p-4 rounded-xl bg-white/5 border border-white/5">
              <span className="text-accent-orange font-black text-xs uppercase tracking-widest mt-1">
                Templates
              </span>
              <div>
                <strong className="text-white block mb-1">
                  MDFlow Renderer
                </strong>
                <span className="text-sm text-white/40">
                  Built-in templates render test cases, feature specs, API docs,
                  and spec tables, with a default fallback if the template name
                  is missing or unknown.
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
                or TSV sources (1MB limit). Supports Google Sheets URLs. Use{" "}
                <code>detect_only=true</code> to return input type analysis
                without conversion.
              </p>
              <div className="bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/paste <br />
                {`{ "paste_text": "...", "template": "default" }`}
              </div>
              <div className="mt-3 bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/preview <br />
                {`{ "paste_text": "..." }`}
              </div>
              <div className="mt-3 bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/gsheet/convert <br />
                {`{ "url": "https://docs.google.com/...", "template": "default" }`}
              </div>
            </div>

            <div className="relative pl-8 border-l-2 border-accent-gold/20">
              <h4 className="text-white font-bold mb-2">
                2. TSV Mode (Binary)
              </h4>
              <p className="text-sm text-muted mb-4">
                Designed for exported sheet data. Accepts `multipart/form-data`
                text streams (10MB limit, .tsv).
              </p>
              <ul className="list-disc list-inside text-sm text-white/60 space-y-2 mb-4">
                <li>No sheet selection required</li>
                <li>Optional `template` parameter for output format</li>
              </ul>
              <div className="bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/tsv (form-data: file, template?)
              </div>
              <div className="mt-3 bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/tsv/preview (form-data: file)
              </div>
            </div>

            <div className="relative pl-8 border-l-2 border-accent-gold/20">
              <h4 className="text-white font-bold mb-2">
                3. XLSX Mode (Binary)
              </h4>
                <p className="text-sm text-muted mb-4">
                  Designed for production specifications. Accepts
                  `multipart/form-data` binary streams (10MB limit, .xlsx only).
                </p>
              <ul className="list-disc list-inside text-sm text-white/60 space-y-2 mb-4">
                <li>
                  Supports multi-sheet discovery (`/api/mdflow/xlsx/sheets`)
                </li>
                <li>Sheet-specific targeting via `sheet_name` parameter</li>
                <li>Optional `template` parameter for output format</li>
              </ul>
              <div className="bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/xlsx (form-data: file, sheet_name?, template?)
              </div>
              <div className="mt-3 bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/xlsx/preview (form-data: file, sheet_name?)
              </div>
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
            The Go backend performs format detection, header mapping, and column
            normalization:
          </p>
          <div className="grid gap-3">
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                1. Input Detection
              </span>
              <span className="text-xs text-white/40 font-mono">
                Table vs Markdown
              </span>
            </div>
            <div className="flex justify-center">
              <ChevronRight className="rotate-90 text-white/20" />
            </div>
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                2. Header Detection
              </span>
              <span className="text-xs text-white/40 font-mono">
                Confidence Scoring (warn if &lt; 50)
              </span>
            </div>
            <div className="flex justify-center">
              <ChevronRight className="rotate-90 text-white/20" />
            </div>
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                3. Column Mapping
              </span>
              <span className="text-xs text-white/40 font-mono">
                Column → Canonical Field
              </span>
            </div>
            <div className="flex justify-center">
              <ChevronRight className="rotate-90 text-white/20" />
            </div>
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                4. Spec Assembly
              </span>
              <span className="text-xs text-white/40 font-mono">
                SpecDoc + Metadata
              </span>
            </div>
            <div className="flex justify-center">
              <ChevronRight className="rotate-90 text-white/20" />
            </div>
            <div className="p-4 rounded-xl bg-white/5 flex items-center justify-between">
              <span className="text-white text-sm font-medium">
                5. Template Rendering
              </span>
              <span className="text-xs text-white/40 font-mono">
                Spec → MDFlow
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
          <div className="bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
            POST /api/mdflow/validate <br />
            {`{ "paste_text": "...", "validation_rules": { "required": ["feature", "scenario"] } }`}
          </div>
          <div className="p-6 bg-black/40 border border-white/10 rounded-xl space-y-4 font-mono text-sm">
            <div>
              <span className="text-accent-purple">meta.column_map</span>
              <p className="text-white/40 pl-4">
                Logs exactly which columns were mapped to canonical fields.
              </p>
            </div>
            <div>
              <span className="text-accent-purple">meta.header_row</span>
              <p className="text-white/40 pl-4">
                The detected header row index used for column mapping.
              </p>
            </div>
            <div>
              <span className="text-accent-purple">meta.sheet_name</span>
              <p className="text-white/40 pl-4">
                The active sheet used for XLSX conversions when applicable.
              </p>
            </div>
            <div>
              <span className="text-accent-gold">meta.total_rows</span>
              <p className="text-white/40 pl-4">
                Counts total parsed rows for the selected sheet.
              </p>
            </div>
            <div>
              <span className="text-accent-gold">meta.rows_by_feature</span>
              <p className="text-white/40 pl-4">
                Aggregated row counts grouped by feature name.
              </p>
            </div>
            <div>
              <span className="text-accent-red">warnings[]</span>
              <p className="text-white/40 pl-4">
                Returns issues like unmapped columns, low header confidence, or
                empty input.
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
    templates: {
      title: "Templates & Schemas",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            MDFlow renders outputs using named templates. The API defaults to{" "}
            <code className="text-accent-orange">default</code> when no template
            is provided, and falls back to <code>default</code> when the
            template name is unknown. Markdown/prose inputs always render with
            the built-in markdown template.
          </p>
          <div className="grid gap-4">
            <div className="p-5 rounded-xl bg-white/5 border border-white/5">
              <div className="text-xs font-black uppercase tracking-widest text-accent-orange mb-2">
                Available Templates
              </div>
              <ul className="text-sm text-white/60 list-disc list-inside space-y-1">
                <li className="capitalize">default</li>
                <li className="capitalize">feature-spec</li>
                <li className="capitalize">test-plan</li>
                <li className="capitalize">api-endpoint</li>
                <li className="capitalize">spec-table</li>
              </ul>
            </div>
            <div className="p-5 rounded-xl bg-black/40 border border-white/10 font-mono text-xs text-accent-green">
              GET /api/mdflow/templates
            </div>
            <div className="p-5 rounded-xl bg-black/40 border border-white/10 font-mono text-xs text-accent-green">
              GET /api/mdflow/templates/info <br />
              GET /api/mdflow/templates/:name
            </div>
            <div className="p-5 rounded-xl bg-black/40 border border-white/10 font-mono text-xs text-accent-green">
              POST /api/mdflow/templates/preview <br />
              {`{ "template_content": "...", "sample_data": "..." }`}
            </div>
            <div className="p-5 rounded-xl bg-white/5 border border-white/5">
              <div className="text-xs font-black uppercase tracking-widest text-accent-orange mb-2">
                Template Usage
              </div>
              <div className="text-sm text-white/50">
                Include <code>template</code> in paste/XLSX requests. If
                omitted, the backend falls back to <code>default</code>. The
                <code>spec-table</code> template highlights Phase 3 fields like
                Item Name, Display Conditions, and Navigation Destination.
              </div>
            </div>
          </div>
        </div>
      ),
    },
    preview: {
      title: "Data Preview",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            MDFlow Studio provides real-time preview functionality to help you
            verify data parsing and column mapping before conversion. Preview is
            automatically triggered with a 500ms debounce when pasting data.
          </p>

          <div className="space-y-6">
            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <Eye className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">Preview Features</h4>
              </div>
              <ul className="space-y-3 text-sm text-white/70">
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Real-time Table Preview</strong> - Automatically
                    shows parsed table structure with headers and sample rows
                    (default: 4 rows)
                  </span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Column Mapping Display</strong> - Shows detected
                    column mappings with visual indicators for unmapped columns
                  </span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Confidence Scoring</strong> - Displays header
                    detection confidence (UI warns if &lt; 70%, backend warnings
                    start at &lt; 50%)
                  </span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Markdown Detection</strong> - Automatically detects
                    markdown/prose content and shows appropriate message
                  </span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Multi-format Support</strong> - Works with paste,
                    TSV, and XLSX inputs (including sheet selection changes)
                  </span>
                </li>
              </ul>
            </div>

            <div className="bg-black/40 p-5 rounded-xl border border-white/10 font-mono text-xs space-y-3">
              <div>
                <span className="text-accent-green">
                  Preview Response Structure:
                </span>
                <pre className="mt-2 text-white/60 overflow-x-auto">
                  {`{
  headers: string[]
  rows: string[][]
  total_rows: number
  preview_rows: number
  header_row: number
  confidence: number
  column_mapping: Record<string, string>
  unmapped_columns: string[]
  input_type: 'table' | 'markdown'
}`}
                </pre>
              </div>
            </div>
          </div>
        </div>
      ),
    },
    "column-mapping": {
      title: "Column Mapping Override",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            The Studio provides manual column mapping override UI to correct
            auto-detection errors. You can reassign columns to canonical fields
            directly in the preview table.
          </p>

          <div className="space-y-6">
            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <Table className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">Manual Column Mapping</h4>
              </div>
              <ul className="space-y-3 text-sm text-white/70">
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Dropdown Selectors</strong> - Each column header has
                    a dropdown to select the canonical field it should map to
                  </span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Visual Indicators</strong> - Unmapped columns are
                    highlighted in gold/orange to draw attention
                  </span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>Canonical Fields</strong> - Supports all standard
                    fields including navigation_destination, item_name,
                    display_conditions, etc.
                  </span>
                </li>
                <li className="flex items-start gap-3">
                  <span className="text-accent-orange mt-1">•</span>
                  <span>
                    <strong>State Persistence</strong> - Column overrides are
                    stored in Zustand store and persist during the session
                  </span>
                </li>
              </ul>
            </div>

            <div className="bg-black/40 p-5 rounded-xl border border-white/10">
              <div className="text-xs font-black uppercase tracking-widest text-accent-orange mb-3">
                Supported Canonical Fields
              </div>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-2 text-xs text-white/60 font-mono">
                <div>id</div>
                <div>feature</div>
                <div>scenario</div>
                <div>instructions</div>
                <div>inputs</div>
                <div>expected</div>
                <div>precondition</div>
                <div>priority</div>
                <div>type</div>
                <div>status</div>
                <div>endpoint</div>
                <div>notes</div>
                <div>no</div>
                <div>item_name</div>
                <div>item_type</div>
                <div>required_optional</div>
                <div>input_restrictions</div>
                <div>display_conditions</div>
                <div>action</div>
                <div>navigation_destination</div>
              </div>
            </div>
          </div>
        </div>
      ),
    },
    integrations: {
      title: "Integrations",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            MDFlow Studio integrates with external services and provides
            multiple export formats for seamless workflow integration.
          </p>

          <div className="space-y-6">
            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <Link2 className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">
                  Google Sheets Integration
                </h4>
              </div>
              <p className="text-sm text-white/70 mb-4">
                Automatically detect and fetch data from public Google Sheets
                URLs. Supports both conversion and preview modes.
              </p>
              <ul className="space-y-2 text-sm text-white/60">
                <li>• Paste Google Sheets URL directly into paste mode</li>
                <li>• Auto-detection of sheet ID and optional gid parameter</li>
                <li>• Fetches CSV export format from public sheets</li>
                <li>• Supports preview and full conversion</li>
              </ul>
              <div className="mt-4 bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/gsheet <br />
                {`{ "url": "https://docs.google.com/spreadsheets/d/..." }`}
              </div>
              <div className="mt-4 bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/gsheet/convert
                <br />
                {`{ "url": "https://docs.google.com/spreadsheets/d/...", "template": "default" }`}
              </div>
            </div>

            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <Download className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">Export Formats</h4>
              </div>
              <p className="text-sm text-white/70 mb-4">
                Export conversion results in multiple formats for different use
                cases.
              </p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                <div className="p-3 rounded-lg bg-black/40 border border-white/5">
                  <div className="text-xs font-bold text-white mb-1">
                    Markdown
                  </div>
                  <div className="text-xs text-white/50">.mdflow.md</div>
                </div>
                <div className="p-3 rounded-lg bg-black/40 border border-white/5">
                  <div className="text-xs font-bold text-white mb-1">JSON</div>
                  <div className="text-xs text-white/50">.json</div>
                </div>
                <div className="p-3 rounded-lg bg-black/40 border border-white/5">
                  <div className="text-xs font-bold text-white mb-1">YAML</div>
                  <div className="text-xs text-white/50">.yaml</div>
                </div>
              </div>
            </div>

            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <Terminal className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">CLI Tool</h4>
              </div>
              <p className="text-sm text-white/70 mb-4">
                Command-line interface for automation and CI/CD workflows.
              </p>
              <div className="bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green space-y-2">
                <div>
                  <span className="text-white/50"># Convert file</span>
                  <br />
                  mdflow convert --input spec.xlsx --output spec.mdflow.md
                  --template test-plan
                </div>
                <div>
                  <span className="text-white/50"># Diff comparison</span>
                  <br />
                  mdflow diff before.md after.md --output diff.json
                </div>
              </div>
            </div>
          </div>
        </div>
      ),
    },
    workflow: {
      title: "Workflow Tools",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            MDFlow Studio includes powerful workflow tools to enhance
            productivity and maintain conversion history.
          </p>

          <div className="space-y-6">
            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <History className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">Conversion History</h4>
              </div>
              <p className="text-sm text-white/70 mb-4">
                Automatically saves conversion history in localStorage. Each
                record includes mode, template, input preview, output, and
                metadata.
              </p>
              <ul className="space-y-2 text-sm text-white/60">
                <li>• View all past conversions with timestamps</li>
                <li>• Restore previous outputs with one click</li>
                <li>• Copy output directly from history</li>
                <li>• Clear history when needed</li>
              </ul>
            </div>

            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <GitCompare className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">Diff Viewer</h4>
              </div>
              <p className="text-sm text-white/70 mb-4">
                Compare different conversion outputs side-by-side or inline to
                track changes.
              </p>
              <ul className="space-y-2 text-sm text-white/60">
                <li>• Save current output for comparison</li>
                <li>• Side-by-side and inline diff views</li>
                <li>• Visual highlighting of additions and deletions</li>
                <li>• Copy diff results</li>
              </ul>
              <div className="mt-4 bg-black/40 p-4 rounded-lg border border-white/5 font-mono text-xs text-accent-green">
                POST /api/mdflow/diff
                <br />
                {`{ "before": "...", "after": "..." }`}
              </div>
            </div>

            <div className="p-5 rounded-xl bg-white/5 border border-white/10">
              <div className="flex items-center gap-3 mb-4">
                <Keyboard className="w-5 h-5 text-accent-orange" />
                <h4 className="text-white font-bold">Keyboard Shortcuts</h4>
              </div>
              <div className="space-y-3 text-sm">
                <div className="flex items-center justify-between p-3 rounded-lg bg-black/40 border border-white/5">
                  <span className="text-white">Convert</span>
                  <kbd className="px-2 py-1 rounded bg-white/10 text-white/70 font-mono text-xs">
                    ⌘/Ctrl + Enter
                  </kbd>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-black/40 border border-white/5">
                  <span className="text-white">Copy Output</span>
                  <kbd className="px-2 py-1 rounded bg-white/10 text-white/70 font-mono text-xs">
                    ⌘/Ctrl + Shift + C
                  </kbd>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-black/40 border border-white/5">
                  <span className="text-white">Download</span>
                  <kbd className="px-2 py-1 rounded bg-white/10 text-white/70 font-mono text-xs">
                    ⌘/Ctrl + S
                  </kbd>
                </div>
              </div>
            </div>
          </div>
        </div>
      ),
    },
    "ai-suggestions": {
      title: "AI Suggestions",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            AI suggestions are optional and only activate when the backend is
            configured with <code>OPENAI_API_KEY</code>. The Studio calls the
            AI endpoint to return actionable improvements for clarity,
            completeness, and formatting.
          </p>
          <div className="bg-black/40 p-5 rounded-xl border border-white/10 font-mono text-xs text-accent-green space-y-2">
            <div>POST /api/mdflow/ai/suggest</div>
            <div>{`{ "paste_text": "...", "template": "default" }`}</div>
          </div>
          <ul className="space-y-2 text-sm text-white/60">
            <li>
              • Returns <code>{`{ suggestions: [], configured: boolean }`}</code>
            </li>
            <li>• Safe to call even when AI is disabled (configured=false)</li>
            <li>• Suggestions are not stored server-side</li>
          </ul>
        </div>
      ),
    },
    "template-editor": {
      title: "Template Editor",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            The Template Editor lets you load built-in templates, edit custom
            templates, and preview output against your current sample data.
            Custom templates are stored locally in the browser.
          </p>
          <div className="grid gap-4">
            <div className="p-5 rounded-xl bg-black/40 border border-white/10 font-mono text-xs text-accent-green">
              GET /api/mdflow/templates <br />
              GET /api/mdflow/templates/:name
            </div>
            <div className="p-5 rounded-xl bg-black/40 border border-white/10 font-mono text-xs text-accent-green">
              POST /api/mdflow/templates/preview <br />
              {`{ "template_content": "...", "sample_data": "..." }`}
            </div>
          </div>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Preview renders server-side using the same converter</li>
            <li>• Built-in templates remain available even after customization</li>
            <li>• Export or copy rendered output for quick iteration</li>
          </ul>
        </div>
      ),
    },
    "batch-processing": {
      title: "Batch Processing",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            Convert many spreadsheets at once from the Batch page. Each file is
            processed independently with progress tracking, and successful
            outputs can be downloaded as a ZIP archive.
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Supports .xlsx and .tsv uploads</li>
            <li>• Optional “process all sheets” for multi-sheet workbooks</li>
            <li>• ZIP export names files as <code>*.mdflow.md</code></li>
          </ul>
          <div className="bg-black/40 p-5 rounded-xl border border-white/10 font-mono text-xs text-accent-green">
            Uses /api/mdflow/xlsx and /api/mdflow/tsv under the hood
          </div>
        </div>
      ),
    },
    "share-links": {
      title: "Share Links",
      content: (
        <div className="space-y-6">
          <p className="text-muted">
            Share links are generated client-side by compressing the MDFlow
            output into a URL-safe payload. This keeps sharing fast and avoids
            storing data on the server.
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Links open in the read-only Share page</li>
            <li>• The payload lives entirely in the URL</li>
            <li>• Large specs may exceed browser URL limits</li>
          </ul>
          <div className="bg-black/40 p-5 rounded-xl border border-white/10 font-mono text-xs text-accent-green">
            /share?d=&lt;compressed-payload&gt;
          </div>
        </div>
      ),
    },
  };

// Mobile Navigation Component with Scroll Indicators
const MobileDocsNav: React.FC<{
  filteredSections: typeof docsSections;
  activeSection: string;
  setActiveSection: (id: string) => void;
  searchQuery: string;
  setSearchQuery: (query: string) => void;
}> = ({
  filteredSections,
  activeSection,
  setActiveSection,
  searchQuery,
  setSearchQuery,
}) => {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(true);

  const checkScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    setCanScrollLeft(el.scrollLeft > 8);
    setCanScrollRight(el.scrollLeft < el.scrollWidth - el.clientWidth - 8);
  }, []);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    checkScroll();
    el.addEventListener("scroll", checkScroll, { passive: true });
    window.addEventListener("resize", checkScroll);
    return () => {
      el.removeEventListener("scroll", checkScroll);
      window.removeEventListener("resize", checkScroll);
    };
  }, [checkScroll, filteredSections]);

  const scroll = (direction: "left" | "right") => {
    const el = scrollRef.current;
    if (!el) return;
    const scrollAmount = el.clientWidth * 0.6;
    el.scrollBy({
      left: direction === "left" ? -scrollAmount : scrollAmount,
      behavior: "smooth",
    });
  };

  return (
    <div className="lg:hidden sticky top-14 sm:top-16 z-30 overflow-hidden">
      {/* Background with blur */}
      <div className="absolute inset-0 bg-background/95 backdrop-blur-xl border-b border-white/5" />
      <div className="absolute inset-0 bg-linear-to-b from-background via-background/80 to-background/60 pointer-events-none" />

      {/* Search Bar */}
      <div className="relative mb-3 pt-3">
        <div className="relative">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40 pointer-events-none" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search documentation..."
            className="w-full h-11 pl-10 pr-4 rounded-2xl bg-white/5 border border-white/10 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-accent-orange/40 focus:bg-white/8 transition-all"
          />
        </div>
      </div>

      {/* Horizontal Scrollable Navigation with Arrows */}
      <div className="relative pb-3">
        {/* Left Scroll Button */}
        {/* <motion.button
          initial={false}
          animate={{ 
            opacity: canScrollLeft ? 1 : 0,
            scale: canScrollLeft ? 1 : 0.8,
          }}
          transition={{ duration: 0.15 }}
          onClick={(e) => {
            e.stopPropagation();
            e.preventDefault();
            scroll("left");
          }}
          onPointerDown={(e) => e.stopPropagation()}
          className="absolute left-2 top-1/2 -translate-y-1/2 z-20 w-9 h-9 flex items-center justify-center rounded-full bg-white/10 hover:bg-white/15 active:bg-white/20 border border-white/20 shadow-xl cursor-pointer transition-colors"
          style={{ pointerEvents: canScrollLeft ? "auto" : "none" }}
        >
          <ChevronLeft className="w-5 h-5 text-white" strokeWidth={2.5} />
        </motion.button> */}

        {/* Right Scroll Button */}
        {/* <motion.button
          initial={false}
          animate={{ 
            opacity: canScrollRight ? 1 : 0,
            scale: canScrollRight ? 1 : 0.8,
          }}
          transition={{ duration: 0.15 }}
          onClick={(e) => {
            e.stopPropagation();
            e.preventDefault();
            scroll("right");
          }}
          onPointerDown={(e) => e.stopPropagation()}
          className="absolute right-2 top-1/2 -translate-y-1/2 z-20 w-9 h-9 flex items-center justify-center rounded-full bg-white/10 hover:bg-white/15 active:bg-white/20 border border-white/20 shadow-xl cursor-pointer transition-colors"
          style={{ pointerEvents: canScrollRight ? "auto" : "none" }}
        >
          <ChevronRight className="w-5 h-5 text-white" strokeWidth={2.5} />
        </motion.button> */}

        {/* Left Fade */}
        <div
          className="pointer-events-none absolute left-0 top-0 bottom-0 w-12 bg-linear-to-r from-background to-transparent z-5 transition-opacity duration-200"
          style={{ opacity: canScrollLeft ? 1 : 0 }}
        />

        {/* Right Fade */}
        <div
          className="pointer-events-none absolute right-0 top-0 bottom-0 w-12 bg-linear-to-l from-background to-transparent z-5 transition-opacity duration-200"
          style={{ opacity: canScrollRight ? 1 : 0 }}
        />

        {/* Scrollable Tabs */}
        <div
          ref={scrollRef}
          className="overflow-x-auto scrollbar-hide scroll-smooth w-full"
        >
          <div className="inline-flex gap-2 px-4 sm:px-6 py-1">
            {filteredSections.map((section) =>
              section.items.map((item) => {
                const isActive = activeSection === item.id;
                return (
                  <motion.button
                    key={item.id}
                    onClick={() => setActiveSection(item.id)}
                    whileTap={{ scale: 0.96 }}
                    className={`relative flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium whitespace-nowrap transition-all cursor-pointer ${
                      isActive
                        ? "text-white"
                        : "text-white/50 hover:text-white/70 active:bg-white/5"
                    }`}
                  >
                    {isActive && (
                      <motion.div
                        layoutId="mobile-active-tab"
                        className="absolute inset-0 rounded-xl bg-accent-orange/15 border border-accent-orange/25"
                        transition={{
                          type: "spring",
                          damping: 30,
                          stiffness: 400,
                        }}
                      />
                    )}
                    <span
                      className={`relative z-10 transition-colors ${
                        isActive ? "text-accent-orange" : "text-white/40"
                      }`}
                    >
                      {item.icon}
                    </span>
                    <span className="relative z-10">{item.title}</span>
                  </motion.button>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

const DocsContentBody: React.FC = () => {
  const [activeSection, setActiveSection] = useState("intro");
  const [searchQuery, setSearchQuery] = useState("");
  const searchParams = useSearchParams();
  const validSections = useMemo(
    () =>
      docsSections.flatMap((section) => section.items.map((item) => item.id)),
    []
  );

  const filteredSections = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return docsSections;
    return docsSections
      .map((section) => ({
        ...section,
        items: section.items.filter(
          (item) =>
            item.title.toLowerCase().includes(q) ||
            item.id.toLowerCase().includes(q) ||
            section.title.toLowerCase().includes(q)
        ),
      }))
      .filter((section) => section.items.length > 0);
  }, [searchQuery]);

  useEffect(() => {
    const section = searchParams.get("section");
    if (section && validSections.includes(section)) {
      setActiveSection(section);
    }
  }, [searchParams, validSections]);

  useEffect(() => {
    const visibleIds = filteredSections.flatMap((s) =>
      s.items.map((i) => i.id)
    );
    if (visibleIds.length > 0 && !visibleIds.includes(activeSection)) {
      setActiveSection(visibleIds[0]);
    }
  }, [filteredSections, activeSection]);

  return (
    <div className="min-h-screen pb-0">
      <div className="">
        <div className="grid lg:grid-cols-[260px_1fr] xl:grid-cols-[280px_1fr] gap-6 sm:gap-8 lg:gap-12">
          <motion.aside
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="hidden lg:block space-y-6 lg:space-y-8 sticky top-0 max-h-[calc(100vh-64px)] overflow-y-auto custom-scrollbar pr-2"
          >
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search docs..."
                className="w-full h-9 sm:h-10 pl-9 sm:pl-10 pr-3 sm:pr-4 rounded-lg sm:rounded-xl bg-white/5 border border-white/10 text-xs sm:text-sm text-white placeholder:text-white/20 focus:outline-hidden focus:border-accent-orange/50 transition-all"
              />
            </div>

            <nav className="space-y-8">
              {filteredSections.map((section, idx) => (
                <div key={idx} className="space-y-3">
                  <h3 className="text-[10px] font-black uppercase tracking-[0.2em] text-white/40 ml-3">
                    {section.title}
                  </h3>
                  <div className="space-y-1">
                    {section.items.map((item) => (
                      <button
                        key={item.id}
                        onClick={() => setActiveSection(item.id)}
                        className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all cursor-pointer group ${
                          activeSection === item.id
                            ? "bg-accent-orange/10 text-accent-orange"
                            : "text-muted hover:text-white hover:bg-white/5"
                        }`}
                      >
                        <span
                          className={`transition-colors ${
                            activeSection === item.id
                              ? "text-accent-orange"
                              : "text-white/40 group-hover:text-white"
                          }`}
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

          {/* Mobile Navigation - Horizontal Scrollable Tabs */}
          <MobileDocsNav
            filteredSections={filteredSections}
            activeSection={activeSection}
            setActiveSection={setActiveSection}
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
          />

          <main className="min-w-0 lg:col-start-2">
            <motion.div
              key={activeSection}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ type: "spring", damping: 28, stiffness: 300 }}
              className="space-y-6 sm:space-y-8"
            >
              <div className="flex items-center gap-2 text-[9px] sm:text-[10px] font-bold uppercase tracking-[0.2em] text-accent-orange/90 mb-1">
                <span className="text-white/50">Docs</span>
                <ChevronRight className="w-3 h-3 text-white/30" />
                {
                  docsSections.find((s) =>
                    s.items.find((i) => i.id === activeSection)
                  )?.title
                }
              </div>

              <h1 className="text-xl sm:text-3xl lg:text-4xl xl:text-5xl font-black text-white tracking-tighter uppercase leading-tight">
                {docContent[activeSection]?.title || "Documentation"}
              </h1>

              <div className="prose prose-invert prose-sm sm:prose-base lg:prose-lg max-w-none">
                {docContent[activeSection]?.content}
              </div>

              <div className="pt-12 sm:pt-16 lg:pt-20 mt-12 sm:mt-16 lg:mt-20 border-t border-white/5 flex flex-col sm:flex-row items-start sm:items-center justify-end gap-4">
                <div className="text-xs sm:text-sm text-muted">
                  Last updated:{" "}
                  <span className="text-white">January 31, 2026</span>
                </div>
              </div>
            </motion.div>
          </main>
        </div>
      </div>
    </div>
  );
};

export default function DocsContent() {
  return (
    <Suspense>
      <DocsContentBody />
    </Suspense>
  );
}
