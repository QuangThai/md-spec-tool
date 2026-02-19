"use client";

import React, { memo } from "react";
import { RefreshCcw, Zap } from "lucide-react";
import { motion } from "framer-motion";
import { TemplateCards } from "@/components/TemplateCards";
import { isMac } from "@/lib/useKeyboardShortcuts";
import type { OutputFormat } from "@/lib/types";

export interface WorkbenchFooterProps {
  format: OutputFormat;
  onFormatChange: (v: OutputFormat) => void;
  onConvert: () => void;
  loading: boolean;
  disabled: boolean;
  mode: "paste" | "xlsx" | "tsv";
  pasteText: string;
  file: File | null;
}

export const WorkbenchFooter = memo(function WorkbenchFooter({
  format,
  onFormatChange,
  onConvert,
  loading,
  disabled,
  mode,
  pasteText,
  file,
}: WorkbenchFooterProps) {
  const modKey = isMac() ? "⌘" : "Ctrl";

  const buttonDisabled =
    disabled ||
    (mode === "paste" && !pasteText.trim()) ||
    ((mode === "xlsx" || mode === "tsv") && !file);

  return (
    <div className="px-3 sm:px-4 py-2.5 sm:py-3 border-t border-white/5 bg-white/2 shrink-0">
      <div className="flex items-center gap-2 sm:gap-3" data-tour="template-selector">
        {/* Template dropdown - collapsible on mobile */}
         <div className="flex-1 min-w-48">
           <TemplateCards selected={format} onSelect={(f) => onFormatChange(f as any)} compact />
         </div>

        {/* Run button */}
        <div className="shrink-0" data-tour="run-button">
          <motion.button
            type="button"
            whileHover={!buttonDisabled ? { scale: 1.02 } : {}}
            whileTap={!buttonDisabled ? { scale: 0.98 } : {}}
            onClick={onConvert}
            disabled={buttonDisabled || loading}
            className={`
              h-9 sm:h-10 px-4 sm:px-6
              uppercase tracking-wider text-[10px] sm:text-xs font-bold rounded-lg
              flex items-center justify-center gap-2
              transition-all duration-200
              ${
                buttonDisabled
                  ? "bg-white/5 border border-white/10 text-white/30 cursor-not-allowed"
                  : "btn-primary shadow-lg shadow-accent-orange/20 cursor-pointer hover:shadow-xl hover:shadow-accent-orange/30"
              }
            `}
            title={
              buttonDisabled
                ? mode === "paste"
                  ? "Paste data"
                  : "Upload file"
                : `${modKey}+Enter`
            }
          >
            {loading ? (
              <RefreshCcw className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Zap className="w-3.5 h-3.5" />
            )}
            <span className="hidden xs:inline">
              {loading ? "Running" : "Run"}
            </span>
          </motion.button>
        </div>
      </div>
    </div>
  );
});

WorkbenchFooter.displayName = "WorkbenchFooter";
