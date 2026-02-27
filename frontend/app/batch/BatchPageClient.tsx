"use client";

import { TemplateCards } from "@/components/TemplateCards";
import { LazyMotion, domAnimation, m } from "framer-motion";
import { ArrowLeft, Boxes, FileSpreadsheet, Layers, Zap } from "lucide-react";
import dynamic from "next/dynamic";
import Link from "next/link";
import { memo, useState } from "react";

// Dynamic import for heavy BatchProcessor component
const BatchProcessor = dynamic(
  () => import("@/components/BatchProcessor").then((m) => m.BatchProcessor),
  {
    ssr: false,
    loading: () => (
      <div className="h-48 flex items-center justify-center">
        <div className="animate-pulse text-white/40 text-sm">Loading…</div>
      </div>
    ),
  }
);

// Static animation configuration - hoisted outside component
const STAGGER_CONFIG = {
  container: {
    animate: { transition: { staggerChildren: 0.08, delayChildren: 0.12 } },
  },
  item: {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.5, ease: [0.16, 1, 0.3, 1] },
  },
} as const;

// Feature card data - static configuration
const FEATURES = [
  {
    icon: FileSpreadsheet,
    iconColor: "text-green-400",
    iconBg: "bg-green-500/10",
    title: "Multi-file Upload",
    description: "Add multiple files together",
  },
  {
    icon: Layers,
    iconColor: "text-blue-400",
    iconBg: "bg-blue-500/10",
    title: "All Sheets",
    description: "Convert every worksheet in Excel",
  },
  {
    icon: Zap,
    iconColor: "text-accent-orange",
    iconBg: "bg-accent-orange/10",
    title: "ZIP Download",
    description: "Download results as one ZIP",
  },
] as const;

// Tips data - static configuration
const TIPS = [
  "Drag and drop multiple files to speed up batch setup",
  'Enable "Process all sheets" to convert every worksheet in Excel files',
  "Download results per file or bundled as a ZIP",
  "Larger batches may take longer depending on file size",
] as const;

// Memoized FeatureCard component - prevents re-renders
const FeatureCard = memo(function FeatureCard({
  icon: Icon,
  iconColor,
  iconBg,
  title,
  description,
}: {
  icon: typeof FileSpreadsheet;
  iconColor: string;
  iconBg: string;
  title: string;
  description: string;
}) {
  return (
    <div className="p-4 rounded-xl bg-white/5 border border-white/10 flex items-start gap-3">
      <div className={`p-2 rounded-lg ${iconBg} ${iconColor}`}>
        <Icon className="w-4 h-4" />
      </div>
      <div>
        <p className="text-xs font-bold text-white mb-0.5">{title}</p>
        <p className="text-[10px] text-white/50">{description}</p>
      </div>
    </div>
  );
});

// Memoized TipsSection component
const TipsSection = memo(function TipsSection() {
  return (
    <div className="rounded-xl border border-blue-500/20 bg-blue-500/5 p-5">
      <h3 className="text-xs font-bold text-blue-400 mb-3 uppercase tracking-wider">
        Tips
      </h3>
      <ul className="space-y-2 text-[11px] text-blue-400/70">
        {TIPS.map((tip) => (
          <li key={tip} className="flex items-start gap-2">
            <span className="text-blue-400">•</span>
            {tip}
          </li>
        ))}
      </ul>
    </div>
  );
});

// Memoized Header component
const PageHeader = memo(function PageHeader() {
  return (
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
  );
});

// Memoized Hero component
const HeroSection = memo(function HeroSection() {
  return (
    <div className="text-center">
      <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-accent-orange/10 border border-accent-orange/20 mb-6">
        <Layers className="w-8 h-8 text-accent-orange" />
      </div>
      <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight mb-3">
        Batch Processing
      </h1>
      <p className="text-sm text-white/60 max-w-lg mx-auto">
        Convert multiple files at once. Upload Excel (.xlsx) or TSV and
        download all outputs as a ZIP.
      </p>
    </div>
  );
});

export default function BatchPageClient() {
  const [format, setFormat] = useState<'spec' | 'table'>('spec');

  return (
    <div className="min-h-screen bg-white/2 rounded-2xl overflow-hidden">
      <PageHeader />

      <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">
        <LazyMotion features={domAnimation}>
          <m.div
            variants={STAGGER_CONFIG.container}
            initial="initial"
            animate="animate"
            className="space-y-8"
          >
            <m.div variants={STAGGER_CONFIG.item}>
              <HeroSection />
            </m.div>

            <m.div variants={STAGGER_CONFIG.item}>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                {FEATURES.map((feature) => (
                  <FeatureCard key={feature.title} {...feature} />
                ))}
              </div>
            </m.div>

            <m.div variants={STAGGER_CONFIG.item}>
              <div
                className="rounded-2xl border border-white/10 bg-white/2 p-6"
                role="group"
                aria-labelledby="batch-output-format-label"
              >
                <span
                  id="batch-output-format-label"
                  className="text-[10px] font-black uppercase tracking-widest text-white/50 block mb-4"
                >
                  Output Format
                </span>
                <TemplateCards
                  selected={format}
                  onSelect={setFormat}
                  compact
                />
              </div>
            </m.div>

            <m.div variants={STAGGER_CONFIG.item}>
              <div className="rounded-2xl border border-white/10 bg-white/2 p-6">
                <BatchProcessor format={format} />
              </div>
            </m.div>

            <m.div variants={STAGGER_CONFIG.item}>
              <TipsSection />
            </m.div>
          </m.div>
        </LazyMotion>
      </main>
    </div>
  );
}
