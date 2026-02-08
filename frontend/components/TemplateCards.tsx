"use client";

import { memo } from "react";
import { motion } from "framer-motion";
import { Check, FileText, Table2 } from "lucide-react";
import { OutputFormat } from "@/lib/types";

interface TemplateCardProps {
  selected: OutputFormat;
  onSelect: (format: OutputFormat) => void;
  compact?: boolean;
}

// ONLY 2 formats - spec & table
const FORMAT_META: Record<OutputFormat, {
  icon: typeof FileText;
  label: string;
  description: string;
  preview: string;
}> = {
  spec: {
    icon: FileText,
    label: "Spec Document",
    description: "AGENTS.md compatible specification",
    preview: `# Feature Specification
## Column Mappings
| Original | Mapped To | Confidence |
|----------|-----------|------------|
| Name | title | 95% |

## Specifications
### REQ-001: User Login
...`,
  },
  table: {
    icon: Table2,
    label: "Simple Table",
    description: "Clean markdown table format",
    preview: `| ID | Title | Status |
|---|---|---|
| REQ-001 | User Login | Done |
| REQ-002 | Profile | In Progress |`,
  },
};

const FORMATS: OutputFormat[] = ["spec", "table"];

export const TemplateCards = memo(function TemplateCards({ selected, onSelect, compact = false }: TemplateCardProps) {
  if (compact) {
    return (
      <div className="flex flex-nowrap gap-3 overflow-x-auto custom-scrollbar pb-1 -mx-1 px-1">
        {FORMATS.map((format) => {
          const meta = FORMAT_META[format];
          const Icon = meta.icon;
          const isSelected = selected === format;

          return (
            <motion.button
              key={format}
              type="button"
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => onSelect(format)}
              className={`
                flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg border transition-all cursor-pointer shrink-0 whitespace-nowrap
                ${isSelected
                  ? "bg-orange-500/20 border-orange-500/50 text-white"
                  : "bg-white/5 border-white/10 text-white/70 hover:bg-white/10 hover:border-white/20"
                }
              `}
            >
              <Icon className={`w-3 h-3 shrink-0 ${isSelected ? "text-orange-400" : ""}`} />
              <span className="text-[9px] font-bold uppercase tracking-wider">{meta.label}</span>
              {isSelected && <Check className="w-2.5 h-2.5 shrink-0 text-accent-orange" />}
            </motion.button>
          );
        })}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
      {FORMATS.map((format) => {
        const meta = FORMAT_META[format];
        const Icon = meta.icon;
        const isSelected = selected === format;

        return (
          <motion.button
            key={format}
            type="button"
            whileHover={{ scale: 1.02, y: -2 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => onSelect(format)}
            className={`
              relative group text-left p-4 rounded-xl border transition-all cursor-pointer overflow-hidden
              ${isSelected
                ? "bg-accent-orange/10 border-accent-orange/40 shadow-lg shadow-accent-orange/10"
                : "bg-white/5 border-white/10 hover:bg-white/8 hover:border-white/20"
              }
            `}
          >
            {/* Selection indicator */}
            {isSelected && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className="absolute top-2 right-2 w-5 h-5 rounded-full bg-accent-orange flex items-center justify-center"
              >
                <Check className="w-3 h-3 text-white" />
              </motion.div>
            )}

            {/* Header */}
            <div className="flex items-center gap-2 mb-2">
              <div className={`
                w-7 h-7 rounded-lg flex items-center justify-center
                ${isSelected ? "bg-accent-orange/20" : "bg-white/10"}
              `}>
                <Icon className={`w-4 h-4 ${isSelected ? "text-accent-orange" : "text-white/60"}`} />
              </div>
              <div>
                <h3 className={`text-[11px] font-black uppercase tracking-wider ${isSelected ? "text-white" : "text-white/80"}`}>
                  {meta.label}
                </h3>
                <p className="text-[9px] text-white/40">{meta.description}</p>
              </div>
            </div>

            {/* Preview */}
            <div className={`
              mt-3 p-2 rounded-lg bg-black/30 border border-white/5
              font-mono text-[8px] leading-relaxed overflow-hidden
              ${isSelected ? "text-white/70" : "text-white/50"}
            `}>
              <pre className="whitespace-pre-wrap line-clamp-6">
                {meta.preview}
              </pre>
            </div>

            {/* Hover gradient */}
            <div className={`
              absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none
              bg-linear-to-t from-accent-orange/5 to-transparent
            `} />
          </motion.button>
        );
      })}
    </div>
  );
});
