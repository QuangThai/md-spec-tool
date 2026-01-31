"use client";

import {
  convertTSV,
  convertXLSX,
  getXLSXSheets,
  MDFlowConvertResponse,
} from "@/lib/mdflowApi";
import { AnimatePresence, motion } from "framer-motion";
import JSZip from "jszip";
import {
  AlertCircle,
  AlertTriangle,
  Check,
  ChevronRight,
  Download,
  File,
  FileSpreadsheet,
  FileText,
  Folder,
  Loader2,
  Package,
  Upload,
  X,
  Zap,
} from "lucide-react";
import { useCallback, useState } from "react";

interface BatchFile {
  id: string;
  file: File;
  status: "pending" | "processing" | "success" | "error";
  progress: number;
  result?: MDFlowConvertResponse;
  error?: string;
  sheets?: string[];
  selectedSheet?: string;
}

interface BatchProcessorProps {
  template: string;
}

export function BatchProcessor({ template }: BatchProcessorProps) {
  const [files, setFiles] = useState<BatchFile[]>([]);
  const [isProcessing, setIsProcessing] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [processAllSheets, setProcessAllSheets] = useState(true);

  // Add files to the batch
  const addFiles = useCallback(async (newFiles: FileList | File[]) => {
    const fileArray = Array.from(newFiles);
    const validFiles = fileArray.filter((f) =>
      /\.(xlsx|xls|tsv|csv)$/i.test(f.name)
    );

    const batchFiles: BatchFile[] = [];

    for (const file of validFiles) {
      const id = `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      const batchFile: BatchFile = {
        id,
        file,
        status: "pending",
        progress: 0,
      };

      // Get sheets for XLSX files
      if (/\.xlsx?$/i.test(file.name)) {
        try {
          const sheetsResult = await getXLSXSheets(file);
          if (sheetsResult.data) {
            batchFile.sheets = sheetsResult.data.sheets;
            batchFile.selectedSheet = sheetsResult.data.active_sheet;
          }
        } catch (e) {
          // Ignore sheet detection errors
        }
      }

      batchFiles.push(batchFile);
    }

    setFiles((prev) => [...prev, ...batchFiles]);
  }, []);

  // Remove a file from the batch
  const removeFile = useCallback((id: string) => {
    setFiles((prev) => prev.filter((f) => f.id !== id));
  }, []);

  // Clear all files
  const clearAll = useCallback(() => {
    setFiles([]);
  }, []);

  // Process all files
  const processAll = useCallback(async () => {
    if (files.length === 0 || isProcessing) return;

    setIsProcessing(true);

    const updatedFiles = [...files];

    for (let i = 0; i < updatedFiles.length; i++) {
      const batchFile = updatedFiles[i];
      if (batchFile.status === "success") continue;

      // Update status to processing
      updatedFiles[i] = { ...batchFile, status: "processing", progress: 50 };
      setFiles([...updatedFiles]);

      try {
        let result;
        const isExcel = /\.xlsx?$/i.test(batchFile.file.name);

        if (isExcel && processAllSheets && batchFile.sheets && batchFile.sheets.length > 1) {
          // Process all sheets - combine results
          const allResults: string[] = [];
          const allWarnings: any[] = [];

          for (const sheet of batchFile.sheets) {
            const sheetResult = await convertXLSX(batchFile.file, sheet, template);
            if (sheetResult.data) {
              allResults.push(`\n\n<!-- Sheet: ${sheet} -->\n\n${sheetResult.data.mdflow}`);
              allWarnings.push(...(sheetResult.data.warnings || []));
            }
          }

          result = {
            data: {
              mdflow: allResults.join("\n---\n"),
              warnings: allWarnings,
              meta: { total_rows: 0, header_row: 0, column_map: {} },
            },
          };
        } else if (isExcel) {
          result = await convertXLSX(
            batchFile.file,
            batchFile.selectedSheet,
            template
          );
        } else {
          result = await convertTSV(batchFile.file, template);
        }

        if (result.error) {
          updatedFiles[i] = {
            ...batchFile,
            status: "error",
            error: result.error,
            progress: 100,
          };
        } else if (result.data) {
          updatedFiles[i] = {
            ...batchFile,
            status: "success",
            result: result.data,
            progress: 100,
          };
        }
      } catch (error) {
        updatedFiles[i] = {
          ...batchFile,
          status: "error",
          error: error instanceof Error ? error.message : "Unknown error",
          progress: 100,
        };
      }

      setFiles([...updatedFiles]);
    }

    setIsProcessing(false);
  }, [files, isProcessing, template, processAllSheets]);

  // Download all results as ZIP
  const downloadAsZip = useCallback(async () => {
    const successFiles = files.filter(
      (f) => f.status === "success" && f.result
    );
    if (successFiles.length === 0) return;

    const zip = new JSZip();

    for (const batchFile of successFiles) {
      const baseName = batchFile.file.name.replace(/\.[^.]+$/, "");
      const fileName = `${baseName}.mdflow.md`;
      zip.file(fileName, batchFile.result!.mdflow);
    }

    const blob = await zip.generateAsync({ type: "blob" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `mdflow-batch-${Date.now()}.zip`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [files]);

  // Download single file
  const downloadSingle = useCallback((batchFile: BatchFile) => {
    if (!batchFile.result) return;
    const baseName = batchFile.file.name.replace(/\.[^.]+$/, "");
    const blob = new Blob([batchFile.result.mdflow], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${baseName}.mdflow.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, []);

  // Handle file drop
  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      if (e.dataTransfer.files.length > 0) {
        addFiles(e.dataTransfer.files);
      }
    },
    [addFiles]
  );

  // Handle file input change
  const onFileChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (e.target.files && e.target.files.length > 0) {
        addFiles(e.target.files);
        e.target.value = "";
      }
    },
    [addFiles]
  );

  const successCount = files.filter((f) => f.status === "success").length;
  const errorCount = files.filter((f) => f.status === "error").length;
  const pendingCount = files.filter(
    (f) => f.status === "pending" || f.status === "processing"
  ).length;

  return (
    <div className="space-y-6">
      {/* Drop zone */}
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={onDrop}
        className={`
          relative rounded-2xl border-2 border-dashed p-8 text-center transition-all duration-300 cursor-pointer
          ${
            dragOver
              ? "border-accent-orange/50 bg-accent-orange/10 scale-[1.01]"
              : "border-white/10 hover:border-white/20 hover:bg-white/3"
          }
        `}
      >
        <input
          type="file"
          accept=".xlsx,.xls,.tsv,.csv"
          multiple
          onChange={onFileChange}
          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
          aria-label="Upload files"
        />

        <div className="flex flex-col items-center gap-4">
          <div
            className={`
              h-16 w-16 rounded-2xl flex items-center justify-center transition-all duration-300
              ${dragOver ? "bg-accent-orange/20" : "bg-white/5"}
            `}
          >
            <Upload
              className={`w-8 h-8 ${
                dragOver ? "text-accent-orange" : "text-muted/50"
              }`}
            />
          </div>
          <div>
            <p className="text-sm font-black text-white uppercase tracking-widest">
              {dragOver ? "Drop files here" : "Drop files or click to upload"}
            </p>
            <p className="text-[10px] text-muted mt-1.5 uppercase font-medium">
              .xlsx, .xls, .tsv, .csv • Multiple files supported
            </p>
          </div>
        </div>
      </div>

      {/* File list */}
      {files.length > 0 && (
        <div className="space-y-4">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Folder className="w-4 h-4 text-accent-orange" />
              <span className="text-xs font-black uppercase tracking-widest text-white">
                Files ({files.length})
              </span>
              {successCount > 0 && (
                <span className="text-[9px] font-bold px-2 py-0.5 rounded-full bg-green-500/20 text-green-400">
                  {successCount} done
                </span>
              )}
              {errorCount > 0 && (
                <span className="text-[9px] font-bold px-2 py-0.5 rounded-full bg-red-500/20 text-red-400">
                  {errorCount} failed
                </span>
              )}
            </div>
            <button
              onClick={clearAll}
              className="text-[9px] text-white/40 hover:text-red-400 transition-colors cursor-pointer font-bold uppercase"
            >
              Clear all
            </button>
          </div>

          {/* Options */}
          <div className="flex items-center gap-4 p-3 rounded-xl bg-white/5 border border-white/10">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={processAllSheets}
                onChange={(e) => setProcessAllSheets(e.target.checked)}
                className="w-4 h-4 rounded border-white/20 bg-white/5 text-accent-orange focus:ring-accent-orange/50"
              />
              <span className="text-[11px] text-white/70">
                Process all sheets in Excel files
              </span>
            </label>
          </div>

          {/* File items */}
          <div className="space-y-2 max-h-[400px] overflow-y-auto custom-scrollbar">
            <AnimatePresence mode="popLayout">
              {files.map((batchFile) => (
                <BatchFileItem
                  key={batchFile.id}
                  batchFile={batchFile}
                  onRemove={() => removeFile(batchFile.id)}
                  onDownload={() => downloadSingle(batchFile)}
                  onSheetChange={(sheet) => {
                    setFiles((prev) =>
                      prev.map((f) =>
                        f.id === batchFile.id
                          ? { ...f, selectedSheet: sheet }
                          : f
                      )
                    );
                  }}
                />
              ))}
            </AnimatePresence>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-between gap-4 pt-4 border-t border-white/10">
            <div className="text-[10px] text-white/40">
              {pendingCount > 0 && `${pendingCount} pending • `}
              Template: <span className="text-accent-orange">{template}</span>
            </div>

            <div className="flex items-center gap-3">
              {successCount > 0 && (
                <motion.button
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  onClick={downloadAsZip}
                  className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-green-500/20 hover:bg-green-500/30 border border-green-500/30 text-green-400 text-[10px] font-bold uppercase tracking-wider transition-all cursor-pointer"
                >
                  <Package className="w-4 h-4" />
                  Download ZIP ({successCount})
                </motion.button>
              )}

              <button
                onClick={processAll}
                disabled={isProcessing || pendingCount === 0}
                className={`
                  flex items-center gap-2 px-6 py-2.5 rounded-xl text-[10px] font-bold uppercase tracking-wider transition-all
                  ${
                    isProcessing || pendingCount === 0
                      ? "bg-white/5 text-white/30 cursor-not-allowed"
                      : "bg-accent-orange hover:bg-accent-orange/90 text-white shadow-lg shadow-accent-orange/25 cursor-pointer"
                  }
                `}
              >
                {isProcessing ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Processing...
                  </>
                ) : (
                  <>
                    <Zap className="w-4 h-4" />
                    Convert All
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Empty state */}
      {files.length === 0 && (
        <div className="text-center py-8">
          <p className="text-[11px] text-white/40 uppercase tracking-wider">
            No files added yet
          </p>
        </div>
      )}
    </div>
  );
}

/* Individual file item */
function BatchFileItem({
  batchFile,
  onRemove,
  onDownload,
  onSheetChange,
}: {
  batchFile: BatchFile;
  onRemove: () => void;
  onDownload: () => void;
  onSheetChange: (sheet: string) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const isExcel = /\.xlsx?$/i.test(batchFile.file.name);
  const hasMultipleSheets = batchFile.sheets && batchFile.sheets.length > 1;
  const warningCount = batchFile.result?.warnings?.length || 0;

  const statusIcon = {
    pending: <File className="w-4 h-4 text-white/40" />,
    processing: <Loader2 className="w-4 h-4 text-accent-orange animate-spin" />,
    success: <Check className="w-4 h-4 text-green-400" />,
    error: <AlertCircle className="w-4 h-4 text-red-400" />,
  };

  const statusColor = {
    pending: "border-white/10 bg-white/2",
    processing: "border-accent-orange/30 bg-accent-orange/5",
    success: "border-green-500/30 bg-green-500/5",
    error: "border-red-500/30 bg-red-500/5",
  };

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, x: -20, height: 0 }}
      className={`rounded-xl border ${statusColor[batchFile.status]} overflow-hidden`}
    >
      <div className="flex items-center gap-3 p-3">
        {/* Status icon */}
        <div className="shrink-0">{statusIcon[batchFile.status]}</div>

        {/* File info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            {isExcel ? (
              <FileSpreadsheet className="w-3.5 h-3.5 text-green-400 shrink-0" />
            ) : (
              <FileText className="w-3.5 h-3.5 text-blue-400 shrink-0" />
            )}
            <span className="text-[11px] font-bold text-white truncate">
              {batchFile.file.name}
            </span>
            <span className="text-[9px] text-white/40 font-mono shrink-0">
              {(batchFile.file.size / 1024).toFixed(1)} KB
            </span>
          </div>

          {/* Sheet selector for Excel */}
          {isExcel && hasMultipleSheets && batchFile.status === "pending" && (
            <div className="flex items-center gap-2 mt-1.5">
              <span className="text-[9px] text-white/40">Sheet:</span>
              <select
                value={batchFile.selectedSheet}
                onChange={(e) => onSheetChange(e.target.value)}
                className="text-[9px] px-2 py-0.5 rounded bg-black/40 border border-white/10 text-white/80 cursor-pointer"
                onClick={(e) => e.stopPropagation()}
              >
                {batchFile.sheets?.map((sheet) => (
                  <option key={sheet} value={sheet}>
                    {sheet}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Error message */}
          {batchFile.status === "error" && batchFile.error && (
            <p className="text-[9px] text-red-400 mt-1 truncate">
              {batchFile.error}
            </p>
          )}

          {/* Warnings */}
          {batchFile.status === "success" && warningCount > 0 && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="flex items-center gap-1 text-[9px] text-amber-400 mt-1 cursor-pointer hover:text-amber-300"
            >
              <AlertTriangle className="w-3 h-3" />
              {warningCount} warning{warningCount > 1 ? "s" : ""}
              <ChevronRight
                className={`w-3 h-3 transition-transform ${
                  expanded ? "rotate-90" : ""
                }`}
              />
            </button>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1.5 shrink-0">
          {batchFile.status === "success" && (
            <button
              onClick={onDownload}
              className="p-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-white/60 hover:text-white transition-all cursor-pointer"
              title="Download"
            >
              <Download className="w-3.5 h-3.5" />
            </button>
          )}
          {batchFile.status !== "processing" && (
            <button
              onClick={onRemove}
              className="p-1.5 rounded-lg hover:bg-red-500/20 text-white/40 hover:text-red-400 transition-all cursor-pointer"
              title="Remove"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      </div>

      {/* Expanded warnings */}
      <AnimatePresence>
        {expanded && batchFile.result?.warnings && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="border-t border-white/5 bg-black/20"
          >
            <div className="p-3 space-y-1">
              {batchFile.result.warnings.map((w, i) => (
                <p
                  key={i}
                  className="text-[9px] text-amber-400/80 pl-3 border-l-2 border-amber-400/30"
                >
                  {w.message}
                </p>
              ))}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}
