"use client";

import React, { memo } from "react";
import {
  FileCode,
  KeyRound,
  ShieldCheck,
} from "lucide-react";
import { QuotaStatus } from "@/components/QuotaStatus";
import { Tooltip } from "@/components/ui/Tooltip";
import { toast } from "@/components/ui/Toast";

export interface SourcePanelHeaderProps {
  mode: "paste" | "xlsx" | "tsv";
  onModeChange: (mode: "paste" | "xlsx" | "tsv") => void;
  openaiKey: string;
  showApiKeyInput: boolean;
  onToggleApiKey: () => void;
  onOpenTemplateEditor: () => void;
  onOpenValidation: () => void;
}

export const SourcePanelHeader = memo(function SourcePanelHeader({
  mode,
  onModeChange,
  openaiKey,
  showApiKeyInput,
  onToggleApiKey,
  onOpenTemplateEditor,
  onOpenValidation,
}: SourcePanelHeaderProps) {
  return (
    <div className="flex items-center justify-between gap-2 px-3 sm:px-4 py-2.5 sm:py-3 border-b border-white/5 bg-white/2 shrink-0">
      <div
        className="flex bg-black/40 rounded-lg border border-white/5 shrink-0"
        data-tour="input-mode"
      >
        {[
          { key: "paste", label: "Paste" },
          { key: "xlsx", label: "Excel" },
          { key: "tsv", label: "TSV" },
        ].map((m) => (
          <button
            key={m.key}
            type="button"
            onClick={() => {
              onModeChange(m.key as "paste" | "xlsx" | "tsv");
            }}
            className={`
              px-3 sm:px-4 py-1.5 text-[9px] sm:text-[10px] font-bold uppercase cursor-pointer tracking-wider rounded-md transition-all duration-200
              ${
                mode === m.key
                  ? "bg-accent-orange text-white shadow-lg shadow-accent-orange/25"
                  : "text-muted hover:text-white hover:bg-white/5"
              }
            `}
          >
            {m.label}
          </button>
        ))}
      </div>

      {/* Quick actions */}
      <div className="flex items-center gap-1.5">
        <QuotaStatus
          compact
          onQuotaExceeded={() => {
            toast.error(
              "Daily quota exceeded",
              "Your token limit has been reached. Try again tomorrow."
            );
          }}
        />
        <Tooltip content={openaiKey ? "OpenAI Key Active" : "Set OpenAI API Key"}>
          <button
            type="button"
            onClick={onToggleApiKey}
            className={`p-1.5 sm:p-2 rounded-lg border transition-all cursor-pointer ${
              openaiKey
                ? "bg-green-500/15 hover:bg-green-500/25 border-green-500/30 hover:border-green-500/40 text-green-400"
                : "bg-white/5 hover:bg-white/10 border-white/10 hover:border-white/20 text-white/60 hover:text-white"
            }`}
          >
            <KeyRound className="w-3.5 h-3.5" />
          </button>
        </Tooltip>
        <Tooltip content="Template Editor">
          <button
            type="button"
            onClick={onOpenTemplateEditor}
            className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
          >
            <FileCode className="w-3.5 h-3.5" />
          </button>
        </Tooltip>
        <Tooltip content="Validation Rules">
          <button
            type="button"
            onClick={onOpenValidation}
            className="p-1.5 sm:p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white transition-all"
          >
            <ShieldCheck className="w-3.5 h-3.5" />
          </button>
        </Tooltip>
      </div>
    </div>
  );
});

SourcePanelHeader.displayName = "SourcePanelHeader";
