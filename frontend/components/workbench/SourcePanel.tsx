"use client";

import React, { memo } from "react";
import { AnimatePresence, motion } from "framer-motion";
import dynamic from "next/dynamic";
import { SourcePanelHeader } from "./SourcePanelHeader";
import { ErrorBanner } from "./ErrorBanner";
import { PasteInput } from "./PasteInput";
import { FileUploadInput } from "./FileUploadInput";
import { WorkbenchFooter } from "./WorkbenchFooter";
import { canConfirmReview } from "@/lib/reviewGate";
import type { PreviewResponse, OutputFormat } from "@/lib/types";

const ApiKeyPanel = dynamic(
  () => import("./ApiKeyPanel").then((mod) => mod.ApiKeyPanel),
  { ssr: false }
);

const ReviewGateBanner = dynamic(
  () => import("./ReviewGateBanner").then((mod) => mod.ReviewGateBanner),
  { ssr: false }
);

const EMPTY_GSHEET_TABS: Array<{ title: string; gid: string }> = [];

const stagger = {
  container: {
    animate: { transition: { staggerChildren: 0.05, delayChildren: 0.08 } },
  },
  item: {
    initial: { opacity: 0, y: 12 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.35, ease: [0.16, 1, 0.3, 1] },
  },
};

export interface SourcePanelProps {
  // Mode and file
  mode: "paste" | "xlsx" | "tsv";
  onModeChange: (mode: "paste" | "xlsx" | "tsv") => void;
  pasteText: string;
  onPasteTextChange: (v: string) => void;
  file: File | null;
  onFileChange: (file: File | null) => void;
  sheets: string[];
  selectedSheet: string;
  onSelectSheet: (sheet: string) => void;

  // Format/template
  format: OutputFormat;
  onFormatChange: (v: OutputFormat) => void;

  // Preview and column overrides
  preview: PreviewResponse | null;
  showPreview: boolean;
  onTogglePreview: () => void;
  previewLoading: boolean;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;

  // API Key
  openaiKey: string;
  setOpenaiKey: (key: string) => void;
  clearOpenaiKey: () => void;
  showApiKeyInput: boolean;
  onToggleApiKeyInput: () => void;
  apiKeyDraft: string;
  onApiKeyDraftChange: (v: string) => void;

  // Error
  error: string | null;
  mappedAppError: {
    title?: string;
    message?: string;
    requestId?: string;
    retryable?: boolean;
  } | null;
  lastFailedAction: "preview" | "convert" | "other" | null;
  onRetryFailedAction: () => void;

  // Review gate
  reviewRequiredColumns: string[];
  reviewedColumns: Record<string, boolean>;
  reviewRemainingCount: number;
  onToggleReviewColumn: (column: string) => void;
  onMarkAllReviewed: () => void;
  onCompleteReview: () => void;
  requiresReviewApproval: boolean;
  reviewApproved: boolean;

  // Google Sheets
   googleSheetInput: any;
   isInputGsheetUrl: boolean;

   // File handling
   fileHandling: any;

  // Loading
  loading: boolean;

  // Conversion
  onConvert: () => void;

  // Modals
  onOpenTemplateEditor: () => void;
  onOpenValidation: () => void;
}

export const SourcePanel = memo(function SourcePanel({
  mode,
  onModeChange,
  pasteText,
  onPasteTextChange,
  file,
  onFileChange,
  sheets,
  selectedSheet,
  onSelectSheet,
  format,
  onFormatChange,
  preview,
  showPreview,
  onTogglePreview,
  previewLoading,
  columnOverrides,
  onColumnOverride,
  openaiKey,
  setOpenaiKey,
  clearOpenaiKey,
  showApiKeyInput,
  onToggleApiKeyInput,
  apiKeyDraft,
  onApiKeyDraftChange,
  error,
  mappedAppError,
  lastFailedAction,
  onRetryFailedAction,
  reviewRequiredColumns,
  reviewedColumns,
  reviewRemainingCount,
  onToggleReviewColumn,
  onMarkAllReviewed,
  onCompleteReview,
  requiresReviewApproval,
  reviewApproved,
  googleSheetInput,
  isInputGsheetUrl,
  fileHandling,
  loading,
  onConvert,
  onOpenTemplateEditor,
  onOpenValidation,
}: SourcePanelProps) {
  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    void fileHandling.handleFileChange(e);
  };

  return (
    <motion.div variants={stagger.item} className="flex flex-col min-h-0 h-full overflow-hidden">
      <section className="p-0 flex flex-col h-full min-h-0 border border-white/10 bg-black/30 backdrop-blur-xl relative overflow-hidden rounded-xl sm:rounded-2xl">
        <div className="studio-grain" aria-hidden />
        <div className="relative z-10 flex flex-col h-full min-h-0">
          {/* Header */}
          <SourcePanelHeader
            mode={mode}
            onModeChange={(newMode) => {
              onModeChange(newMode);
              onFileChange(null);
            }}
            openaiKey={openaiKey}
            showApiKeyInput={showApiKeyInput}
            onToggleApiKey={onToggleApiKeyInput}
            onOpenTemplateEditor={onOpenTemplateEditor}
            onOpenValidation={onOpenValidation}
          />

          {/* API Key Panel */}
          <ApiKeyPanel
            show={showApiKeyInput}
            openaiKey={openaiKey}
            apiKeyDraft={apiKeyDraft}
            onDraftChange={onApiKeyDraftChange}
            onSave={() => {
              setOpenaiKey(apiKeyDraft.trim());
            }}
            onClear={() => {
              clearOpenaiKey();
              onApiKeyDraftChange("");
            }}
          />

          {/* Main content area */}
          <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-3 sm:px-4 py-3 custom-scrollbar bg-black/3">
            <AnimatePresence mode="wait" initial={false}>
              {/* Error Banner */}
              {error ? (
                <ErrorBanner
                  key="error-banner"
                  error={error}
                  mappedAppError={mappedAppError}
                  lastFailedAction={lastFailedAction}
                  loading={loading}
                  previewLoading={previewLoading}
                  onRetry={onRetryFailedAction}
                />
              ) : null}

              {/* Review Gate Banner */}
              {requiresReviewApproval && !reviewApproved ? (
                <ReviewGateBanner
                  key="review-gate-banner"
                  reviewRequiredColumns={reviewRequiredColumns}
                  reviewedColumns={reviewedColumns}
                  reviewRemainingCount={reviewRemainingCount}
                  onToggleColumn={(col) => {
                    onToggleReviewColumn(col);
                  }}
                  onMarkAll={onMarkAllReviewed}
                  onConfirmReview={onCompleteReview}
                  canConfirm={canConfirmReview(reviewRequiredColumns, reviewedColumns)}
                />
              ) : null}

              {/* Input content - paste or file */}
              {mode === "paste" ? (
                <PasteInput
                  key="paste-input"
                  pasteText={pasteText}
                  onPasteTextChange={onPasteTextChange}
                  isInputGsheetUrl={isInputGsheetUrl}
                  gsheetTabs={googleSheetInput.googleSheetInput?.gsheetTabs ?? EMPTY_GSHEET_TABS}
                  selectedGid={googleSheetInput.googleSheetInput?.selectedGid || ""}
                  onSelectGid={googleSheetInput.setSelectedGid}
                  gsheetLoading={googleSheetInput.gsheetLoading}
                  gsheetRange={googleSheetInput.gsheetRange}
                  onGsheetRangeChange={googleSheetInput.setGsheetRange}
                  googleAuth={googleSheetInput.googleAuth}
                  preview={preview}
                  showPreview={showPreview}
                  onTogglePreview={onTogglePreview}
                  previewLoading={previewLoading}
                  columnOverrides={columnOverrides}
                  onColumnOverride={onColumnOverride}
                  requiresReviewApproval={requiresReviewApproval}
                  reviewApproved={reviewApproved}
                  onRefetchGsheetPreview={() => {
                    // Refetch will be triggered by hook
                  }}
                />
              ) : (
                <FileUploadInput
                  key="file-upload-input"
                  mode={mode as "xlsx" | "tsv"}
                  file={file}
                  sheets={sheets}
                  selectedSheet={selectedSheet}
                  onSelectSheet={onSelectSheet}
                  dragOver={fileHandling.dragOver}
                  onDragOver={fileHandling.onDragOver}
                  onDragLeave={fileHandling.onDragLeave}
                  onDrop={fileHandling.onDrop}
                  onFileChange={handleFileInputChange}
                  preview={preview}
                  showPreview={showPreview}
                  onTogglePreview={onTogglePreview}
                  previewLoading={previewLoading}
                  columnOverrides={columnOverrides}
                  onColumnOverride={onColumnOverride}
                  requiresReviewApproval={requiresReviewApproval}
                  reviewApproved={reviewApproved}
                />
              )}
            </AnimatePresence>
          </div>

          {/* Footer with format and run button */}
          <WorkbenchFooter
            format={format}
            onFormatChange={(v) => onFormatChange(v as any)}
            onConvert={onConvert}
            loading={loading}
            disabled={false}
            mode={mode}
            pasteText={pasteText}
            file={file}
          />
        </div>
      </section>
    </motion.div>
  );
});

SourcePanel.displayName = "SourcePanel";
