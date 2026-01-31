"use client";

import { MDFlowWarning } from "@/lib/types";
import { useMDFlowStore } from "@/lib/mdflowStore";
import { AnimatePresence, motion } from "framer-motion";
import {
  AlertCircle,
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  HelpCircle,
  Info,
  Lightbulb,
  X,
  Zap,
} from "lucide-react";
import { useState, useMemo } from "react";

interface WarningPanelProps {
  warnings: MDFlowWarning[];
  onApplyFix?: (fixAction: FixAction) => void;
}

export interface FixAction {
  type: "map_column" | "ignore_warning" | "auto_fix";
  warningCode: string;
  payload?: Record<string, unknown>;
}

// Warning severity icons and colors
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

// Category labels
const categoryLabels: Record<string, string> = {
  input: "Input",
  detect: "Detection",
  header: "Header",
  mapping: "Mapping",
  rows: "Rows",
  render: "Render",
};

// Smart suggestions based on warning codes
function getSuggestions(warning: MDFlowWarning): string[] {
  const suggestions: string[] = [];

  if (warning.hint) {
    suggestions.push(warning.hint);
  }

  // Add smart suggestions based on warning code patterns
  switch (warning.code) {
    case "UNMAPPED_COLUMNS":
      suggestions.push(
        "Consider mapping these columns manually using the dropdown selectors in the preview table."
      );
      break;
    case "LOW_HEADER_CONFIDENCE":
      suggestions.push(
        "Try ensuring your first row contains clear column headers like 'Feature', 'Scenario', 'Instructions', etc."
      );
      break;
    case "EMPTY_CELLS":
      suggestions.push(
        "Fill in the empty cells in your source data, or the tool will use placeholder values."
      );
      break;
    case "DUPLICATE_HEADERS":
      suggestions.push(
        "Rename duplicate column headers in your source data to avoid confusion."
      );
      break;
    case "NO_FEATURE_COLUMN":
      suggestions.push(
        "Add a column named 'Feature', 'Title', 'Name', or similar to identify your items."
      );
      break;
    case "NO_INSTRUCTIONS_COLUMN":
      suggestions.push(
        "Add a column named 'Instructions', 'Steps', 'Description', or similar for the main content."
      );
      break;
    default:
      if (warning.category === "mapping") {
        suggestions.push(
          "You can manually override column mappings using the preview table."
        );
      }
  }

  return suggestions;
}

// Get quick fix actions based on warning
function getQuickFixes(warning: MDFlowWarning): { label: string; action: FixAction }[] {
  const fixes: { label: string; action: FixAction }[] = [];

  if (warning.code === "UNMAPPED_COLUMNS" && warning.details?.columns) {
    fixes.push({
      label: "Ignore unmapped columns",
      action: {
        type: "ignore_warning",
        warningCode: warning.code,
      },
    });
  }

  // Add dismiss option for all warnings
  fixes.push({
    label: "Dismiss this warning",
    action: {
      type: "ignore_warning",
      warningCode: warning.code,
    },
  });

  return fixes;
}

function WarningItem({
  warning,
  onDismiss,
  onApplyFix,
}: {
  warning: MDFlowWarning;
  onDismiss: () => void;
  onApplyFix?: (action: FixAction) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const config = severityConfig[warning.severity] || severityConfig.warn;
  const Icon = config.icon;
  const suggestions = useMemo(() => getSuggestions(warning), [warning]);
  const quickFixes = useMemo(() => getQuickFixes(warning), [warning]);

  return (
    <motion.div
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, x: -8, height: 0 }}
      layout
      className={`rounded-xl border ${config.border} ${config.bg} overflow-hidden`}
    >
      <div
        className="flex items-start gap-3 p-3 cursor-pointer"
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
        <Icon className={`w-4 h-4 mt-0.5 shrink-0 ${config.color}`} />
        
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className={`text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded ${config.badge}`}>
              {warning.severity}
            </span>
            <span className="text-[9px] text-white/40 uppercase font-medium tracking-wider">
              {categoryLabels[warning.category] || warning.category}
            </span>
            {warning.code && (
              <span className="text-[9px] text-white/30 font-mono">
                {warning.code}
              </span>
            )}
          </div>
          
          <p className="text-[11px] text-white/80 mt-1.5 leading-relaxed">
            {warning.message}
          </p>
        </div>

        <div className="flex items-center gap-1 shrink-0">
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDismiss();
            }}
            className="p-1 rounded hover:bg-white/10 text-white/30 hover:text-white/60 transition-colors cursor-pointer"
            title="Dismiss"
          >
            <X className="w-3.5 h-3.5" />
          </button>
          <div className={`p-1 ${expanded ? "rotate-90" : ""} transition-transform`}>
            <ChevronRight className="w-3.5 h-3.5 text-white/40" />
          </div>
        </div>
      </div>

      <AnimatePresence>
        {expanded && (suggestions.length > 0 || quickFixes.length > 0) && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="border-t border-white/5 overflow-hidden"
          >
            <div className="p-3 space-y-3">
              {/* Suggestions */}
              {suggestions.length > 0 && (
                <div>
                  <div className="flex items-center gap-1.5 mb-2">
                    <Lightbulb className="w-3 h-3 text-accent-gold/80" />
                    <span className="text-[9px] font-bold uppercase tracking-wider text-accent-gold/80">
                      Suggestions
                    </span>
                  </div>
                  <ul className="space-y-1.5">
                    {suggestions.map((suggestion, i) => (
                      <li
                        key={i}
                        className="text-[10px] text-white/60 pl-4 border-l-2 border-white/10 leading-relaxed"
                      >
                        {suggestion}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {/* Quick fixes */}
              {quickFixes.length > 0 && (
                <div>
                  <div className="flex items-center gap-1.5 mb-2">
                    <Zap className="w-3 h-3 text-accent-orange/80" />
                    <span className="text-[9px] font-bold uppercase tracking-wider text-accent-orange/80">
                      Quick Actions
                    </span>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {quickFixes.map((fix, i) => (
                      <button
                        key={i}
                        onClick={(e) => {
                          e.stopPropagation();
                          onApplyFix?.(fix.action);
                          if (fix.action.type === "ignore_warning") {
                            onDismiss();
                          }
                        }}
                        className="text-[9px] font-bold uppercase tracking-wider px-2.5 py-1.5 rounded-lg bg-white/5 hover:bg-white/10 border border-white/10 text-white/70 hover:text-white transition-all cursor-pointer"
                      >
                        {fix.label}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Details (if available) */}
              {warning.details && Object.keys(warning.details).length > 0 && (
                <div>
                  <div className="flex items-center gap-1.5 mb-2">
                    <HelpCircle className="w-3 h-3 text-white/40" />
                    <span className="text-[9px] font-bold uppercase tracking-wider text-white/40">
                      Details
                    </span>
                  </div>
                  <pre className="text-[9px] text-white/40 font-mono bg-black/30 rounded-lg p-2 overflow-x-auto">
                    {JSON.stringify(warning.details, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

export function WarningPanel({ warnings, onApplyFix }: WarningPanelProps) {
  const { dismissWarning, dismissedWarningCodes } = useMDFlowStore();
  const [collapsed, setCollapsed] = useState(false);

  // Filter out dismissed warnings
  const visibleWarnings = useMemo(
    () => warnings.filter((w) => !dismissedWarningCodes[w.code]),
    [warnings, dismissedWarningCodes]
  );

  // Group by severity for summary
  const severityCounts = useMemo(() => {
    const counts = { error: 0, warn: 0, info: 0 };
    visibleWarnings.forEach((w) => {
      if (w.severity in counts) {
        counts[w.severity as keyof typeof counts]++;
      }
    });
    return counts;
  }, [visibleWarnings]);

  if (visibleWarnings.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3" data-tour="warnings-panel">
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
          <AlertCircle className="w-3.5 h-3.5 text-accent-gold/90" />
          <span className="text-[9px] font-black uppercase tracking-[0.25em] text-accent-gold/80">
            Warnings
          </span>
          <span className="text-[9px] text-white/40 font-mono">
            ({visibleWarnings.length})
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* Severity badges */}
          <div className="flex items-center gap-1.5">
            {severityCounts.error > 0 && (
              <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-red-500/20 text-red-400">
                {severityCounts.error} error{severityCounts.error > 1 ? "s" : ""}
              </span>
            )}
            {severityCounts.warn > 0 && (
              <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-400">
                {severityCounts.warn} warn{severityCounts.warn > 1 ? "s" : ""}
              </span>
            )}
            {severityCounts.info > 0 && (
              <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400">
                {severityCounts.info} info
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

      {/* Warning list */}
      <AnimatePresence>
        {!collapsed && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="space-y-2 max-h-[300px] overflow-y-auto custom-scrollbar"
          >
            <AnimatePresence mode="popLayout">
              {visibleWarnings.map((warning, index) => (
                <WarningItem
                  key={`${warning.code}-${index}`}
                  warning={warning}
                  onDismiss={() => dismissWarning(warning.code)}
                  onApplyFix={onApplyFix}
                />
              ))}
            </AnimatePresence>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

// Compact inline warning for preview table
export function InlineWarningBadge({
  warning,
  className = "",
}: {
  warning: MDFlowWarning;
  className?: string;
}) {
  const config = severityConfig[warning.severity] || severityConfig.warn;
  const Icon = config.icon;

  return (
    <div
      className={`inline-flex items-center gap-1 text-[9px] ${config.color} ${className}`}
      title={warning.message}
    >
      <Icon className="w-3 h-3" />
      <span className="font-medium">{warning.message}</span>
    </div>
  );
}
