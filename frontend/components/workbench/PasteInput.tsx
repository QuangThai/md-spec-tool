"use client";

import React, { memo } from "react";
import { AnimatePresence, LazyMotion, domAnimation, m } from "framer-motion";
import {
  AlertCircle,
  Check,
  Eye,
  EyeOff,
  FileText,
  KeyRound,
  Link2,
  RefreshCcw,
} from "lucide-react";
import { isGoogleSheetsURL } from "@/lib/mdflowApi";
import { PreviewTable } from "@/components/PreviewTable";
import { Select } from "@/components/ui/Select";
import type { PreviewResponse } from "@/lib/types";

export interface PasteInputProps {
  pasteText: string;
  onPasteTextChange: (v: string) => void;
  isInputGsheetUrl: boolean;
  gsheetTabs: Array<{ title: string; gid: string }>;
  selectedGid: string;
  onSelectGid: (v: string) => void;
  gsheetLoading: boolean;
  gsheetRange: string;
  onGsheetRangeChange: (v: string) => void;
  googleAuth: {
    connected: boolean;
    loading: boolean;
    login: () => void;
    logout: () => void;
  };
  preview: PreviewResponse | null;
  showPreview: boolean;
  onTogglePreview: () => void;
  previewLoading: boolean;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
  onRefetchGsheetPreview: () => void;
}

export const PasteInput = memo(function PasteInput({
  pasteText,
  onPasteTextChange,
  isInputGsheetUrl,
  gsheetTabs,
  selectedGid,
  onSelectGid,
  gsheetLoading,
  gsheetRange,
  onGsheetRangeChange,
  googleAuth,
  preview,
  showPreview,
  onTogglePreview,
  previewLoading,
  columnOverrides,
  onColumnOverride,
  requiresReviewApproval,
  reviewApproved,
  onRefetchGsheetPreview,
}: PasteInputProps) {
  const trimmedPasteText = pasteText.trim();
  const isGoogleSheetUrl = isGoogleSheetsURL(trimmedPasteText);
  const tablePreview =
    preview !== null && preview.input_type === "table" && preview.headers.length > 0
      ? preview
      : null;
  const markdownPreview =
    preview !== null && preview.input_type === "markdown" ? preview : null;

  return (
    <LazyMotion features={domAnimation}>
      <m.div
        key="paste"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.2 }}
        className="h-full flex flex-col min-h-0"
      >
      {/* Compact status bar */}
      <div className="flex flex-wrap items-center gap-2 text-[9px] uppercase font-bold text-muted/50 mb-2 shrink-0">
        {isGoogleSheetUrl ? (
          <span className="flex items-center gap-1 text-green-400/80 bg-green-400/10 px-2 py-0.5 rounded">
            <Link2 className="w-3 h-3" />
            Google Sheet
          </span>
        ) : null}
        {tablePreview && !isGoogleSheetUrl ? (
          <button
            type="button"
            onClick={onTogglePreview}
            className="flex items-center gap-1 text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer bg-accent-orange/10 px-2 py-0.5 rounded"
          >
            {showPreview ? (
              <EyeOff className="w-3 h-3" />
            ) : (
              <Eye className="w-3 h-3" />
            )}
            {showPreview ? "Hide" : "Show"} Preview
          </button>
        ) : null}
        {previewLoading ? (
          <span className="flex items-center gap-1 text-accent-orange/60">
            <RefreshCcw className="w-3 h-3 animate-spin" />
            Analyzing…
          </span>
        ) : null}
        {isGoogleSheetUrl && gsheetTabs.length > 0 ? (
          <button
            type="button"
            onClick={onRefetchGsheetPreview}
            className="flex items-center gap-1 text-blue-400/70 hover:text-blue-400 transition-colors cursor-pointer bg-blue-400/10 px-2 py-0.5 rounded"
          >
            <RefreshCcw className="w-3 h-3" />
            Refresh
          </button>
        ) : null}
        {isGoogleSheetUrl && gsheetLoading ? (
          <span className="flex items-center gap-1 text-blue-400/70">
            <RefreshCcw className="w-3 h-3 animate-spin" />
            Loading sheets…
          </span>
        ) : null}
      </div>

      {isGoogleSheetUrl && gsheetTabs.length > 0 ? (
        <div className="mb-3 shrink-0">
          <Select
            value={selectedGid}
            onValueChange={onSelectGid}
            options={gsheetTabs.map((tab) => ({
              label: tab.title,
              value: tab.gid,
            }))}
            placeholder="Choose sheet"
            size="compact"
            className="w-auto min-w-40"
          />
        </div>
      ) : null}
      {isGoogleSheetUrl ? (
        <div className="mb-3 flex flex-wrap items-center gap-2 text-[9px] text-white/60 uppercase font-bold tracking-wider">
          <label htmlFor="gsheet-range-input" className="text-white/50">Range</label>
          <input
            id="gsheet-range-input"
            name="gsheetRange"
            autoComplete="off"
            type="text"
            value={gsheetRange}
            onChange={(e) => onGsheetRangeChange(e.target.value)}
            placeholder="A1:F200 or Sheet1!A1:F200…"
            className="h-7 px-2 rounded-md bg-black/40 border border-white/10 text-[10px] text-white/80 placeholder-white/30 focus:outline-none focus-visible:border-accent-orange/40 focus-visible:ring-2 focus-visible:ring-accent-orange/20"
          />
        </div>
      ) : null}
      {isInputGsheetUrl ? (
        <div
          className={`mb-3 flex flex-wrap items-center gap-2 rounded-lg px-3 py-2 text-[10px] text-white/70 border ${
            googleAuth.connected
              ? "border-green-500/20 bg-green-500/10"
              : "border-accent-orange/20 bg-accent-orange/10"
          }`}
        >
          {googleAuth.connected ? (
            <Check className="h-3 w-3 shrink-0 text-green-400" />
          ) : (
            <AlertCircle className="h-3 w-3 shrink-0 text-accent-orange/80" />
          )}
          <span className="flex-1 min-w-40">
            {googleAuth.connected
              ? "Google connected. You can access private sheets without sharing."
              : "Private sheet? Connect Google to access without sharing."}
          </span>
          {googleAuth.connected ? (
            <button
              type="button"
              onClick={googleAuth.logout}
              className="inline-flex items-center gap-1 rounded-md border border-white/10 bg-white/5 px-2 py-1 text-[9px] font-bold uppercase tracking-wider text-white/70 hover:text-white transition-colors"
            >
              Disconnect
            </button>
          ) : (
            <button
              type="button"
              onClick={googleAuth.login}
              disabled={googleAuth.loading}
              className="inline-flex items-center gap-1 rounded-md border border-accent-orange/30 bg-accent-orange/20 px-2 py-1 text-[9px] font-bold uppercase tracking-wider text-white/90 hover:bg-accent-orange/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <KeyRound className="h-3 w-3" />
              Connect Google
            </button>
          )}
        </div>
      ) : null}

      {/* Preview Table - Collapsible */}
      <div data-tour="preview-table">
        <AnimatePresence>
          {showPreview && tablePreview ? (
            <m.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              exit={{ opacity: 0, height: 0 }}
              className="mb-3 shrink-0 max-h-[30vh] overflow-auto custom-scrollbar"
            >
              <PreviewTable
                preview={tablePreview}
                columnOverrides={columnOverrides}
                onColumnOverride={onColumnOverride}
                needsReview={requiresReviewApproval}
                reviewApproved={reviewApproved}
                sourceUrl={isGoogleSheetUrl ? trimmedPasteText : undefined}
                onSelectBlockRange={
                  isGoogleSheetUrl ? (range) => onGsheetRangeChange(range) : undefined
                }
              />
            </m.div>
          ) : null}
          {markdownPreview ? (
            <m.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              exit={{ opacity: 0, height: 0 }}
              className="mb-3 shrink-0"
            >
              <div className="rounded-lg border border-blue-500/20 bg-blue-500/5 px-3 py-2 flex items-center gap-2">
                <FileText className="w-3.5 h-3.5 text-blue-400/80 shrink-0" />
                <span className="text-[9px] font-bold text-blue-400/90 uppercase tracking-wider">
                  Markdown detected - passthrough mode
                </span>
              </div>
            </m.div>
          ) : null}
        </AnimatePresence>
      </div>

      <textarea
        data-tour="paste-area"
        value={pasteText}
        onChange={(e) => onPasteTextChange(e.target.value)}
        placeholder="Paste your table data here (TSV, CSV, or Google Sheets URL)…"
        className="input flex-1 font-mono text-[12px] leading-relaxed resize-none border-white/5 bg-black/30 focus:bg-black/40 focus:border-accent-orange/30 custom-scrollbar min-h-30 rounded-lg"
        aria-label="Paste TSV or CSV data"
      />
      </m.div>
    </LazyMotion>
  );
});

PasteInput.displayName = "PasteInput";
