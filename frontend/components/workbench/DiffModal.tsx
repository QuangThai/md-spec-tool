"use client";

import React, { memo } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { X } from "lucide-react";
import { DiffViewer } from "@/components/DiffViewer";
import type { DiffResponse } from "@/lib/diffTypes";

export interface DiffModalProps {
  showDiff: boolean;
  currentDiff: DiffResponse | null;
  onClose: () => void;
}

export const DiffModal = memo(function DiffModal({
  showDiff,
  currentDiff,
  onClose,
}: DiffModalProps) {
  return (
    <AnimatePresence>
      {showDiff && currentDiff ? (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        >
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0.95, opacity: 0 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-black/60 backdrop-blur-xl border border-white/20 rounded-2xl shadow-2xl max-w-4xl w-full max-h-[80vh] flex flex-col overflow-hidden"
          >
            <div className="flex items-center justify-between gap-4 px-6 py-3 border-b border-white/10 bg-white/3 shrink-0">
              <div className="flex items-center gap-3">
                <span className="text-[10px] font-black uppercase tracking-[0.25em] text-white/80">
                  MDFlow Diff Viewer
                </span>
              </div>
              <button
                onClick={onClose}
                className="p-2 rounded-md hover:bg-white/10 transition-colors cursor-pointer text-white/60 hover:text-white"
                aria-label="Close"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            <div className="flex-1 min-h-0 overflow-auto custom-scrollbar">
              <DiffViewer diff={currentDiff} />
            </div>
          </motion.div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
});

DiffModal.displayName = "DiffModal";
