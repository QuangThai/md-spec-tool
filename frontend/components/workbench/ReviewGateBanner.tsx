"use client";

import React, { memo } from "react";

export interface ReviewGateBannerProps {
  reviewRequiredColumns: string[];
  reviewedColumns: Record<string, boolean>;
  reviewRemainingCount: number;
  onToggleColumn: (column: string) => void;
  onMarkAll: () => void;
  onConfirmReview: () => void;
  canConfirm: boolean;
}

export const ReviewGateBanner = memo(function ReviewGateBanner({
  reviewRequiredColumns,
  reviewedColumns,
  reviewRemainingCount,
  onToggleColumn,
  onMarkAll,
  onConfirmReview,
  canConfirm,
}: ReviewGateBannerProps) {
  return (
    <div className="mb-3 flex items-center justify-between gap-3 rounded-lg border border-accent-gold/30 bg-accent-gold/10 px-3 py-2">
      <div className="min-w-0 flex-1">
        <p className="text-[10px] font-bold uppercase tracking-wider text-accent-gold/90">
          Review Required
        </p>
        <p className="text-[10px] text-white/70">
          Low-confidence mapping detected. Confirm review before sharing,
          exporting, or copying output.
        </p>
        {reviewRequiredColumns.length > 0 ? (
          <div className="mt-2 space-y-1">
            <p className="text-[9px] uppercase tracking-wider text-white/60">
              Review Columns ({reviewRequiredColumns.length})
            </p>
            <div className="flex flex-wrap gap-1.5">
              {reviewRequiredColumns.map((column) => {
                const checked = Boolean(reviewedColumns[column]);
                return (
                  <button
                    key={column}
                    type="button"
                    onClick={() => onToggleColumn(column)}
                    className={`rounded-md border px-2 py-1 text-[9px] font-bold uppercase tracking-wider transition-colors ${
                      checked
                        ? "border-green-500/40 bg-green-500/20 text-green-300"
                        : "border-white/15 bg-black/20 text-white/70 hover:bg-black/35"
                    }`}
                  >
                    {checked ? "✓ " : ""}
                    {column}
                  </button>
                );
              })}
            </div>
            <p className="text-[9px] text-white/55">
              {reviewRemainingCount} column(s) remaining
            </p>
          </div>
        ) : null}
      </div>
      <div className="shrink-0 flex flex-col gap-1.5">
        {reviewRequiredColumns.length > 0 ? (
          <button
            type="button"
            onClick={onMarkAll}
            className="rounded-lg border border-white/20 bg-black/20 px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider text-white/85 hover:bg-black/35 transition-colors"
          >
            Mark All
          </button>
        ) : null}
        <button
          type="button"
          onClick={onConfirmReview}
          disabled={!canConfirm}
          className={`rounded-lg border px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider transition-colors ${
            !canConfirm
              ? "border-white/15 bg-white/5 text-white/35 cursor-not-allowed"
              : "border-accent-gold/40 bg-accent-gold/20 text-white hover:bg-accent-gold/30"
          }`}
        >
          Confirm Review
        </button>
      </div>
    </div>
  );
});

ReviewGateBanner.displayName = "ReviewGateBanner";
