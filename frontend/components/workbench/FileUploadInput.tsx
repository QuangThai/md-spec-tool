"use client";

import React, { memo } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
  Check,
  Eye,
  EyeOff,
  FileSpreadsheet,
  RefreshCcw,
} from "lucide-react";
import { Select } from "@/components/ui/Select";
import { PreviewTable } from "@/components/PreviewTable";
import type { PreviewResponse } from "@/lib/types";

export interface FileUploadInputProps {
  mode: "xlsx" | "tsv";
  file: File | null;
  sheets: string[];
  selectedSheet: string;
  onSelectSheet: (v: string) => void;
  dragOver: boolean;
  onDragOver: (e: React.DragEvent) => void;
  onDragLeave: () => void;
  onDrop: (e: React.DragEvent) => void;
  onFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  preview: PreviewResponse | null;
  showPreview: boolean;
  onTogglePreview: () => void;
  previewLoading: boolean;
  columnOverrides: Record<string, string>;
  onColumnOverride: (column: string, field: string) => void;
  requiresReviewApproval: boolean;
  reviewApproved: boolean;
}

export const FileUploadInput = memo(function FileUploadInput({
  mode,
  file,
  sheets,
  selectedSheet,
  onSelectSheet,
  dragOver,
  onDragOver,
  onDragLeave,
  onDrop,
  onFileChange,
  preview,
  showPreview,
  onTogglePreview,
  previewLoading,
  columnOverrides,
  onColumnOverride,
  requiresReviewApproval,
  reviewApproved,
}: FileUploadInputProps) {
  const tablePreview =
    file !== null &&
    preview !== null &&
    preview.input_type === "table" &&
    preview.headers.length > 0
      ? preview
      : null;

  return (
    <motion.div
      key={mode}
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.2 }}
      className={`h-full flex flex-col gap-4 min-h-0 ${
        !file ? "justify-center items-center" : "justify-start"
      }`}
    >
      {/* File drop zone - centered when no file, shrink when file uploaded */}
      <div
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
        onDrop={onDrop}
        className={`
          relative rounded-2xl border-2 border-dashed transition-all duration-300 cursor-pointer w-full shrink-0
          ${file ? "p-4" : "p-8 sm:p-12 max-w-lg"}
          ${
            dragOver
              ? "border-accent-orange/50 bg-accent-orange/10 scale-[1.02]"
              : file
                ? "border-accent-orange/30 bg-accent-orange/5"
                : "border-white/20 hover:border-accent-orange/40 hover:bg-white/5"
          }
        `}
      >
        <input
          type="file"
          accept={mode === "tsv" ? ".tsv" : ".xlsx"}
          onChange={onFileChange}
          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
          aria-label={mode === "tsv" ? "Upload TSV file" : "Upload Excel file"}
        />
        <div
          className={`flex items-center gap-4 ${
            file ? "justify-start" : "justify-center flex-col"
          }`}
        >
          <div
            className={`
              rounded-2xl flex items-center justify-center transition-all
              ${file ? "h-12 w-12 bg-accent-orange/20" : "h-16 w-16 bg-white/10"}
            `}
          >
            {file ? (
              <Check className="w-6 h-6 text-accent-orange" />
            ) : (
              <FileSpreadsheet
                className={`w-8 h-8 ${
                  dragOver ? "text-accent-orange" : "text-white/40"
                }`}
              />
            )}
          </div>
          <div className={file ? "text-left" : "text-center"}>
            {file ? (
              <>
                <p className="text-sm font-bold text-white truncate max-w-62.5">
                  {file.name}
                </p>
                <p className="text-xs text-white/50 font-mono">
                  {(file.size / 1024).toFixed(1)} KB
                </p>
              </>
            ) : (
              <>
                <p className="text-sm font-black text-white uppercase tracking-widest">
                  {dragOver
                    ? "Drop file here"
                    : mode === "tsv"
                      ? "Upload .TSV"
                      : "Upload .XLSX"}
                </p>
                <p className="text-xs text-white/50 mt-1">
                  Click or drag & drop
                </p>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Sheet selector */}
      {mode === "xlsx" && sheets.length > 0 ? (
        <div className="shrink-0">
          <Select
            value={selectedSheet}
            onValueChange={onSelectSheet}
            options={sheets.map((s) => ({
              label: s,
              value: s,
            }))}
            placeholder="Choose sheet"
            size="compact"
            className="w-auto min-w-30"
          />
        </div>
      ) : null}

      {/* File Preview Table - takes remaining space */}
      <div data-tour="preview-table">
        <AnimatePresence>
          {tablePreview ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="flex-1 min-h-0 flex flex-col"
            >
              <div className="flex items-center justify-between mb-2 shrink-0">
                <span className="text-[10px] text-white/50 uppercase font-bold tracking-wider">
                  Data Preview
                </span>
                <button
                  type="button"
                  onClick={onTogglePreview}
                  className="flex items-center gap-1.5 text-[10px] text-accent-orange/70 hover:text-accent-orange transition-colors cursor-pointer font-bold uppercase"
                >
                  {showPreview ? (
                    <EyeOff className="w-3.5 h-3.5" />
                  ) : (
                    <Eye className="w-3.5 h-3.5" />
                  )}
                  {showPreview ? "Hide" : "Show"}
                </button>
              </div>
              {showPreview ? (
                <div className="flex-1 min-h-0 overflow-auto custom-scrollbar rounded-lg border border-white/10">
                  <PreviewTable
                    preview={tablePreview}
                    columnOverrides={columnOverrides}
                    onColumnOverride={onColumnOverride}
                    needsReview={requiresReviewApproval}
                    reviewApproved={reviewApproved}
                  />
                </div>
              ) : null}
              {previewLoading ? (
                <div className="flex items-center gap-2 text-[10px] text-accent-orange/60 mt-2 shrink-0">
                  <RefreshCcw className="w-3.5 h-3.5 animate-spin" />
                  Loading preview...
                </div>
              ) : null}
            </motion.div>
          ) : null}
        </AnimatePresence>
      </div>
    </motion.div>
  );
});

FileUploadInput.displayName = "FileUploadInput";
