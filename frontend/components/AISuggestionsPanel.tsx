"use client";

import { AISuggestion, AISuggestionType } from "@/lib/mdflowApi";
import { AnimatePresence, motion } from "framer-motion";
import {
  AlertCircle,
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  FileQuestion,
  Info,
  Lightbulb,
  ListChecks,
  Sparkles,
  TextCursorInput,
  Wand2,
} from "lucide-react";
import { useState, useMemo } from "react";

interface AISuggestionsPanelProps {
  suggestions: AISuggestion[];
  loading?: boolean;
  error?: string | null;
  configured?: boolean | null;
}

// Suggestion type icons and labels
const typeConfig: Record<
  AISuggestionType,
  { icon: React.ComponentType<{ className?: string }>; label: string; color: string }
> = {
  missing_field: {
    icon: FileQuestion,
    label: "Missing Field",
    color: "text-red-400",
  },
  vague_description: {
    icon: TextCursorInput,
    label: "Vague Description",
    color: "text-amber-400",
  },
  incomplete_steps: {
    icon: ListChecks,
    label: "Incomplete Steps",
    color: "text-orange-400",
  },
  formatting: {
    icon: Wand2,
    label: "Formatting",
    color: "text-blue-400",
  },
  coverage: {
    icon: AlertCircle,
    label: "Coverage Gap",
    color: "text-purple-400",
  },
};

// Severity config for badges
const severityConfig = {
  info: {
    icon: Info,
    color: "text-blue-400",
    bg: "bg-blue-500/10",
    border: "border-blue-500/20",
    badge: "bg-blue-500/20 text-blue-400",
  },
  warn: {
    icon: AlertTriangle,
    color: "text-amber-400",
    bg: "bg-amber-500/10",
    border: "border-amber-500/20",
    badge: "bg-amber-500/20 text-amber-400",
  },
  error: {
    icon: AlertCircle,
    color: "text-red-400",
    bg: "bg-red-500/10",
    border: "border-red-500/20",
    badge: "bg-red-500/20 text-red-400",
  },
};

function SuggestionItem({ suggestion }: { suggestion: AISuggestion }) {
  const [expanded, setExpanded] = useState(false);
  const severity = severityConfig[suggestion.severity] || severityConfig.info;
  const typeInfo = typeConfig[suggestion.type] || typeConfig.vague_description;
  const TypeIcon = typeInfo.icon;

  // Truncate long field names
  const displayField = suggestion.field && suggestion.field.length > 20 
    ? suggestion.field.substring(0, 20) + "..." 
    : suggestion.field;

  return (
    <motion.div
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, x: -8, height: 0 }}
      layout
      className={`rounded-xl border ${severity.border} ${severity.bg} overflow-hidden`}
    >
      <div
        className="flex items-start gap-3 p-3 cursor-pointer min-w-0"
        onClick={() => setExpanded(!expanded)}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            setExpanded(!expanded);
          }
        }}
      >
        <TypeIcon className={`w-4 h-4 mt-0.5 shrink-0 ${typeInfo.color}`} />

        <div className="flex-1 min-w-0 overflow-hidden">
          <div className="flex items-center gap-2 flex-wrap">
            <span
              className={`text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded shrink-0 ${severity.badge}`}
            >
              {suggestion.severity}
            </span>
            <span className="text-[9px] text-white/40 uppercase font-medium tracking-wider shrink-0">
              {typeInfo.label}
            </span>
            {suggestion.row_ref && (
              <span className="text-[9px] text-white/30 font-mono shrink-0">
                Row {suggestion.row_ref}
              </span>
            )}
            {displayField && (
              <span className="text-[9px] text-white/30 font-mono truncate max-w-[120px]" title={suggestion.field}>
                [{displayField}]
              </span>
            )}
          </div>

          <p className="text-[11px] text-white/80 mt-1.5 leading-relaxed wrap-break-word">
            {suggestion.message}
          </p>
        </div>

        <div className="flex items-center gap-1 shrink-0">
          <div
            className={`p-1 ${expanded ? "rotate-90" : ""} transition-transform`}
          >
            <ChevronRight className="w-3.5 h-3.5 text-white/40" />
          </div>
        </div>
      </div>

      <AnimatePresence>
        {expanded && suggestion.suggestion && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="border-t border-white/5 overflow-hidden"
          >
            <div className="p-3 min-w-0">
              <div className="flex items-center gap-1.5 mb-2">
                <Lightbulb className="w-3 h-3 text-accent-gold/80 shrink-0" />
                <span className="text-[9px] font-bold uppercase tracking-wider text-accent-gold/80">
                  Suggested Improvement
                </span>
              </div>
              <p className="text-[10px] text-white/70 pl-4 border-l-2 border-accent-gold/30 leading-relaxed whitespace-pre-wrap wrap-break-word">
                {suggestion.suggestion}
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

export function AISuggestionsPanel({
  suggestions,
  loading,
  error,
  configured,
}: AISuggestionsPanelProps) {
  const [collapsed, setCollapsed] = useState(false);

  // Group by severity for summary
  const severityCounts = useMemo(() => {
    const counts = { error: 0, warn: 0, info: 0 };
    suggestions.forEach((s) => {
      if (s.severity in counts) {
        counts[s.severity as keyof typeof counts]++;
      }
    });
    return counts;
  }, [suggestions]);

  // Don't render if not configured
  if (configured === false) {
    return null;
  }

  // Show error state
  if (error) {
    // Extract a short error message
    const shortError = error.length > 100 
      ? error.includes("status 401") 
        ? "Invalid OpenAI API key. Please check your API key configuration."
        : error.includes("status 429")
        ? "Rate limit exceeded. Please try again later."
        : error.includes("status 500")
        ? "OpenAI service error. Please try again."
        : error.substring(0, 100) + "..."
      : error;
    
    return (
      <div className="space-y-3" data-tour="ai-suggestions">
        <div className="flex items-center gap-2">
          <Sparkles className="w-3.5 h-3.5 text-purple-400/90" />
          <span className="text-[9px] font-black uppercase tracking-[0.25em] text-purple-400/80">
            AI Suggestions
          </span>
        </div>
        <div className="rounded-xl border border-red-500/20 bg-red-500/10 p-3 overflow-hidden">
          <p className="text-[11px] text-red-400 wrap-break-word">{shortError}</p>
        </div>
      </div>
    );
  }

  // Show loading state
  if (loading) {
    return (
      <div className="space-y-3" data-tour="ai-suggestions">
        <div className="flex items-center gap-2">
          <Sparkles className="w-3.5 h-3.5 text-purple-400/90 animate-pulse" />
          <span className="text-[9px] font-black uppercase tracking-[0.25em] text-purple-400/80">
            AI Suggestions
          </span>
          <span className="text-[9px] text-white/40">Analyzing...</span>
        </div>
        <div className="rounded-xl border border-purple-500/20 bg-purple-500/5 p-4">
          <div className="flex items-center gap-3">
            <div className="w-4 h-4 border-2 border-purple-400/40 border-t-purple-400 rounded-full animate-spin" />
            <p className="text-[11px] text-white/60">
              AI is analyzing your specification for quality improvements...
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Don't render if no suggestions
  if (suggestions.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3" data-tour="ai-suggestions">
      {/* Header */}
      <div
        className="flex items-center justify-between cursor-pointer"
        onClick={() => setCollapsed(!collapsed)}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            setCollapsed(!collapsed);
          }
        }}
      >
        <div className="flex items-center gap-2">
          <Sparkles className="w-3.5 h-3.5 text-purple-400/90" />
          <span className="text-[9px] font-black uppercase tracking-[0.25em] text-purple-400/80">
            AI Suggestions
          </span>
          <span className="text-[9px] text-white/40 font-mono">
            ({suggestions.length})
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* Severity badges */}
          <div className="flex items-center gap-1.5">
            {severityCounts.error > 0 && (
              <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-red-500/20 text-red-400">
                {severityCounts.error} critical
              </span>
            )}
            {severityCounts.warn > 0 && (
              <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-400">
                {severityCounts.warn} important
              </span>
            )}
            {severityCounts.info > 0 && (
              <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400">
                {severityCounts.info} minor
              </span>
            )}
          </div>

          <ChevronDown
            className={`w-4 h-4 text-white/40 transition-transform ${
              collapsed ? "-rotate-90" : ""
            }`}
          />
        </div>
      </div>

      {/* Suggestions list */}
      <AnimatePresence>
        {!collapsed && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="space-y-2 max-h-[400px] overflow-y-auto custom-scrollbar"
          >
            <AnimatePresence mode="popLayout">
              {suggestions.map((suggestion, index) => (
                <SuggestionItem
                  key={`${suggestion.type}-${suggestion.row_ref || index}-${index}`}
                  suggestion={suggestion}
                />
              ))}
            </AnimatePresence>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
