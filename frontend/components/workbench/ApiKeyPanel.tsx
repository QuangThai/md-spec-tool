"use client";

import React, { memo } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { KeyRound } from "lucide-react";

export interface ApiKeyPanelProps {
  show: boolean;
  openaiKey: string;
  apiKeyDraft: string;
  onDraftChange: (v: string) => void;
  onSave: () => void;
  onClear: () => void;
}

export const ApiKeyPanel = memo(function ApiKeyPanel({
  show,
  openaiKey,
  apiKeyDraft,
  onDraftChange,
  onSave,
  onClear,
}: ApiKeyPanelProps) {
  return (
    <AnimatePresence>
      {show ? (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: "auto" }}
          exit={{ opacity: 0, height: 0 }}
          transition={{ duration: 0.2 }}
          className="px-3 sm:px-4 py-2.5 border-b border-white/5 bg-black/20 shrink-0"
        >
          <div className="flex items-center gap-2">
            <KeyRound className="w-3.5 h-3.5 text-white/40 shrink-0" />
            {openaiKey ? (
              <>
                <div className="flex-1 bg-black/30 border border-green-500/20 rounded-lg px-3 py-1.5 text-xs text-green-400/80 font-mono">
                  sk-...{openaiKey.slice(-4)}
                </div>
                <button
                  type="button"
                  onClick={onClear}
                  className="shrink-0 px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider rounded-lg border border-red-500/30 bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-all cursor-pointer"
                >
                  Clear
                </button>
              </>
            ) : (
              <>
                <input
                  type="password"
                  value={apiKeyDraft}
                  onChange={(e) => onDraftChange(e.target.value)}
                  placeholder="sk-..."
                  className="flex-1 min-w-0 bg-black/30 border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white/90 placeholder-white/30 focus:border-accent-orange/30 focus:outline-none font-mono"
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && apiKeyDraft.trim().length >= 10) {
                      onSave();
                    }
                  }}
                />
                <button
                  type="button"
                  onClick={onSave}
                  disabled={apiKeyDraft.trim().length < 10}
                  className={`shrink-0 px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider rounded-lg border transition-all ${
                    apiKeyDraft.trim().length >= 10
                      ? "border-accent-orange/30 bg-accent-orange/20 text-white hover:bg-accent-orange/30 cursor-pointer"
                      : "border-white/10 bg-white/5 text-white/30 cursor-not-allowed"
                  }`}
                >
                  Save
                </button>
              </>
            )}
          </div>
          <p className="text-[9px] text-white/35 mt-1.5 pl-5.5">
            Your key is stored locally in your browser and sent to the backend
            as a per-request header. Never stored on our server.
          </p>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
});

ApiKeyPanel.displayName = "ApiKeyPanel";
