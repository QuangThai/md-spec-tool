"use client";

import React, { memo } from "react";
import {
  Copy,
  Download,
  GitCompare,
  History,
  Loader2,
  RotateCcw,
  Save,
  Check,
} from "lucide-react";
import { Tooltip } from "@/components/ui/Tooltip";
import { ShareButton } from "@/components/ShareButton";
import type { MDFlowMeta } from "@/lib/types";

export interface OutputToolbarProps {
  // Output state
  mdflowOutput: string;
  meta: MDFlowMeta | null;
  // Review state
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
  reviewGateReason: string | undefined;
  // Copy
  copied: boolean;
  onCopy: () => void;
  // Snapshots
  snapshotA: string;
  snapshotB: string;
  compareLoading: boolean;
  onSaveSnapshot: () => void;
  onCompareSnapshots: () => void;
  onClearSnapshots: () => void;
  // Export
  onDownload: () => void;
  // Share
  format: string;
  // History
  historyCount: number;
  onShowHistory: () => void;
}

export const OutputToolbar = memo(function OutputToolbar({
  mdflowOutput,
  meta,
  requiresReviewApproval,
  reviewApproved,
  reviewGateReason,
  copied,
  onCopy,
  snapshotA,
  snapshotB,
  compareLoading,
  onSaveSnapshot,
  onCompareSnapshots,
  onClearSnapshots,
  onDownload,
  format,
  historyCount,
  onShowHistory,
}: OutputToolbarProps) {
  return (
    <div className="flex items-center justify-between gap-2 px-3 sm:px-4 py-2.5 sm:py-3 border-b border-white/5 bg-white/2 shrink-0">
      <div className="flex items-center gap-2 min-w-0">
        <div className="flex gap-0.5 shrink-0">
          <span className={`w-1.5 h-1.5 rounded-full bg-red-400/80`} />
          <span className="w-1.5 h-1.5 rounded-full bg-yellow-400/80" />
          <span className="w-1.5 h-1.5 rounded-full bg-green-400/80" />
        </div>
        <span className="text-[9px] sm:text-[10px] font-bold uppercase tracking-wider text-white/70">
          Output
        </span>
        {requiresReviewApproval && !reviewApproved ? (
          <span className="text-[8px] uppercase tracking-wider font-bold px-2 py-0.5 rounded bg-accent-gold/20 border border-accent-gold/30 text-accent-gold/80">
            Needs Review
          </span>
        ) : null}
        {requiresReviewApproval && reviewApproved ? (
          <span className="text-[8px] uppercase tracking-wider font-bold px-2 py-0.5 rounded bg-green-500/20 border border-green-500/30 text-green-300">
            Reviewed
          </span>
        ) : null}
        {mdflowOutput && meta ? (
          <span className="text-[8px] sm:text-[9px] hidden sm:inline text-muted/50 font-mono">
            {meta.total_rows || 0} rows
          </span>
        ) : null}
      </div>
      {/* Action buttons - always visible, disabled when no output */}
      <div className="flex items-center gap-1.5 shrink-0">
        <Tooltip
          content={reviewGateReason ?? (copied ? "Copied!" : "Copy")}
        >
          <button
            type="button"
            aria-label="Copy output"
            onClick={onCopy}
            disabled={!mdflowOutput || Boolean(reviewGateReason)}
            className={`p-1.5 sm:p-2 rounded-lg border transition-[background-color,border-color,color] duration-200 ${
              mdflowOutput
                ? "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
                : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
            }`}
          >
            {copied ? (
              <Check className="w-3.5 h-3.5 text-accent-orange" />
            ) : (
              <Copy className="w-3.5 h-3.5" />
            )}
          </button>
        </Tooltip>
        {snapshotA ? (
          <span
            className="hidden sm:inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[8px] font-bold bg-rose-500/20 text-rose-300/90 border border-rose-500/30"
            title="Version A saved"
          >
            A ✓
          </span>
        ) : null}
        {snapshotB ? (
          <span
            className="hidden sm:inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[8px] font-bold bg-emerald-500/20 text-emerald-300/90 border border-emerald-500/30"
            title="Version B saved"
          >
            B ✓
          </span>
        ) : null}
        <Tooltip
          content={
            !snapshotA
              ? "Save as Version A (before)"
              : !snapshotB
                ? "Save as Version B (after)"
                : "Overwrite Version B"
          }
        >
          <button
            type="button"
            aria-label="Save snapshot"
            onClick={onSaveSnapshot}
            disabled={!mdflowOutput}
            className={`p-1.5 sm:p-2 rounded-lg border transition-[background-color,border-color,color] duration-200 ${
              mdflowOutput
                ? "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
                : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
            }`}
          >
            <Save className="w-3.5 h-3.5" />
          </button>
        </Tooltip>
        {snapshotA && snapshotB ? (
          <>
            <Tooltip content={compareLoading ? "Comparing…" : "Compare (2 versions)"}>
              <button
                type="button"
                aria-label="Compare snapshots"
                onClick={onCompareSnapshots}
                disabled={compareLoading}
                className={`p-1.5 sm:p-2 rounded-lg border text-white/60 transition-[background-color,border-color,color] duration-200 ${
                  compareLoading
                    ? "bg-white/10 border-white/20 cursor-wait"
                    : "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 hover:text-white"
                }`}
              >
                {compareLoading ? (
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <GitCompare className="w-3.5 h-3.5" />
                )}
              </button>
            </Tooltip>
            <Tooltip content="Clear snapshots">
              <button
                type="button"
                aria-label="Clear snapshots"
                onClick={onClearSnapshots}
                className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-[background-color,border-color,color] duration-200"
              >
                <RotateCcw className="w-3.5 h-3.5" />
              </button>
            </Tooltip>
          </>
        ) : null}
        <Tooltip content={reviewGateReason ?? "Export"}>
          <button
            type="button"
            aria-label="Export output"
            disabled={!mdflowOutput || Boolean(reviewGateReason)}
            className={`p-1.5 sm:p-2 rounded-lg border transition-[background-color,border-color,color] duration-200 ${
              mdflowOutput
                ? "bg-accent-orange/90 hover:bg-accent-orange border-accent-orange/50 text-white"
                : "bg-white/5 border-white/5 text-white/20 cursor-not-allowed"
            }`}
            onClick={onDownload}
          >
            <Download className="w-3.5 h-3.5" />
          </button>
        </Tooltip>
        <ShareButton
          mdflowOutput={mdflowOutput}
          template={format}
          disabledReason={reviewGateReason}
        />
        {historyCount > 0 ? (
          <Tooltip content="History">
            <button
              type="button"
              aria-label="Show history"
              onClick={onShowHistory}
              className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-[background-color,border-color,color] duration-200"
            >
              <History className="w-3.5 h-3.5" />
            </button>
          </Tooltip>
        ) : null}
      </div>
    </div>
  );
});

OutputToolbar.displayName = "OutputToolbar";
