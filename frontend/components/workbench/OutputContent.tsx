"use client";

import React, { memo } from "react";
import { motion } from "framer-motion";
import { Terminal } from "lucide-react";
import { OutputSkeleton } from "@/components/ui/Skeleton";

export interface OutputContentProps {
  loading: boolean;
  mdflowOutput: string;
}

export const OutputContent = memo(function OutputContent({
  loading,
  mdflowOutput,
}: OutputContentProps) {
  return (
    <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden px-3 sm:px-4 py-3 custom-scrollbar">
      {loading ? (
        <OutputSkeleton />
      ) : mdflowOutput ? (
        <motion.pre
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.2 }}
          className="whitespace-pre-wrap wrap-break-word font-mono text-[11px] sm:text-[12px] leading-relaxed text-white/90 selection:bg-accent-orange/30"
        >
          {mdflowOutput}
        </motion.pre>
      ) : (
        <div className="h-full flex flex-col items-center justify-center text-center py-6">
          <div className="rounded-xl bg-white/5 border border-white/5 p-4 mb-3">
            <Terminal className="w-8 h-8 text-white/20" />
          </div>
          <p className="text-[10px] font-bold uppercase tracking-widest text-white/40">
            Output will appear here
          </p>
          <p className="text-[9px] text-muted/50 mt-1">
            Paste data and run to generate
          </p>
        </div>
      )}
    </div>
  );
});

OutputContent.displayName = "OutputContent";
