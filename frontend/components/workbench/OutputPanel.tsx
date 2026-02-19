"use client";

import React, { memo } from "react";
import { motion } from "framer-motion";
import { OutputToolbar } from "./OutputToolbar";
import { OutputContent } from "./OutputContent";
import { TechnicalAnalysis } from "@/components/TechnicalAnalysis";
import type { MDFlowMeta, MDFlowWarning, AISuggestion, OutputFormat } from "@/lib/types";

const stagger = {
  item: {
    initial: { opacity: 0, y: 12 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.35, ease: [0.16, 1, 0.3, 1] },
  },
};

export interface OutputPanelProps {
  // Output
  mdflowOutput: string;
  loading: boolean;
  meta: MDFlowMeta | null;
  warnings: MDFlowWarning[];
  aiSuggestions: AISuggestion[];
  aiSuggestionsLoading: boolean;
  aiSuggestionsError: string | null;
  aiConfigured: boolean;
  onRetryAISuggestions: () => void;

  // Review state
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
  reviewGateReason: string | undefined;

  // Format
  format: OutputFormat;

  // Copy/download
  copied: boolean;
  onCopy: () => void;
  onDownload: () => void;

  // Snapshots
  snapshotA: string;
  snapshotB: string;
  compareLoading: boolean;
  onSaveSnapshot: () => void;
  onCompareSnapshots: () => void;
  onClearSnapshots: () => void;

  // History
  historyCount: number;
  onShowHistory: () => void;
}

export const OutputPanel = memo(function OutputPanel({
  mdflowOutput,
  loading,
  meta,
  warnings,
  aiSuggestions,
  aiSuggestionsLoading,
  aiSuggestionsError,
  aiConfigured,
  onRetryAISuggestions,
  requiresReviewApproval,
  reviewApproved,
  reviewGateReason,
  format,
  copied,
  onCopy,
  onDownload,
  snapshotA,
  snapshotB,
  compareLoading,
  onSaveSnapshot,
  onCompareSnapshots,
  onClearSnapshots,
  historyCount,
  onShowHistory,
}: OutputPanelProps) {
  return (
    <motion.div
      variants={stagger.item}
      className="flex flex-col min-h-0 h-full overflow-hidden"
      data-tour="output-panel"
    >
      <div className="p-0 flex flex-col h-full min-h-0 border border-white/10 bg-black/30 backdrop-blur-xl relative overflow-hidden rounded-xl sm:rounded-2xl">
        <div className="studio-grain" aria-hidden />
        <div className="relative z-10 flex flex-col h-full min-h-0">
          {/* Output Toolbar */}
          <OutputToolbar
            mdflowOutput={mdflowOutput}
            meta={meta}
            requiresReviewApproval={requiresReviewApproval}
            reviewApproved={reviewApproved}
            reviewGateReason={reviewGateReason}
            copied={copied}
            onCopy={onCopy}
            snapshotA={snapshotA}
            snapshotB={snapshotB}
            compareLoading={compareLoading}
            onSaveSnapshot={onSaveSnapshot}
            onCompareSnapshots={onCompareSnapshots}
            onClearSnapshots={onClearSnapshots}
            onDownload={onDownload}
            format={format}
            historyCount={historyCount}
            onShowHistory={onShowHistory}
          />

          {/* Output Content */}
          <OutputContent loading={loading} mdflowOutput={mdflowOutput} />

          {/* Technical Analysis Footer */}
          {mdflowOutput || warnings.length > 0 || aiSuggestions.length > 0 ? (
            <div className="border-t border-white/10 bg-white/2 px-3 sm:px-4 py-2 sm:py-2.5 shrink-0">
              <TechnicalAnalysis
                meta={meta}
                warnings={warnings}
                mdflowOutput={mdflowOutput}
                aiSuggestions={aiSuggestions}
                aiSuggestionsLoading={aiSuggestionsLoading}
                aiSuggestionsError={aiSuggestionsError}
                aiConfigured={aiConfigured}
                onRetryAISuggestions={onRetryAISuggestions}
              />
            </div>
          ) : null}
        </div>
      </div>
    </motion.div>
  );
});

OutputPanel.displayName = "OutputPanel";
